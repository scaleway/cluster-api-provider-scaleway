package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
)

// ScalewayMachineReconciler reconciles a ScalewayMachine object
type ScalewayMachineReconciler struct {
	client.Client

	createScalewayMachineService scalewayMachineServiceCreator
}

// scalewayMachineServiceCreator is a function that creates a new scalewayMachineService reconciler.
type scalewayMachineServiceCreator func(machineScope *scope.Machine) *scalewayMachineService

// NewScalewayClusterReconciler returns a new ScalewayClusterReconciler.
func NewScalewayMachineReconciler(c client.Client) *ScalewayMachineReconciler {
	return &ScalewayMachineReconciler{
		Client:                       c,
		createScalewayMachineService: newScalewayMachineService,
	}
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymachines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ScalewayMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := logf.FromContext(ctx)

	scalewayMachine := &infrav1.ScalewayMachine{}
	err := r.Get(ctx, req.NamespacedName, scalewayMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the Machine.
	machine, err := util.GetOwnerMachine(ctx, r.Client, scalewayMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}

	if machine == nil {
		log.Info("Machine Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("machine", machine.Name)

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		log.Info("Machine is missing cluster label or cluster does not exist")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	log = log.WithValues("ScalewayCluster", cluster.Spec.InfrastructureRef.Name)
	scalewayCluster := &infrav1.ScalewayCluster{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: scalewayMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}, scalewayCluster); err != nil {
		log.Info("ScalewayCluster is not available yet")
		return ctrl.Result{}, nil
	}

	// Create the cluster scope
	clusterScope, err := scope.NewCluster(ctx, &scope.ClusterParams{
		Client:          r.Client,
		Cluster:         cluster,
		ScalewayCluster: scalewayCluster,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create the machine scope
	machineScope, err := scope.NewMachine(&scope.MachineParams{
		Client:          r.Client,
		ClusterScope:    clusterScope,
		Machine:         machine,
		ScalewayMachine: scalewayMachine,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create scope: %w", err)
	}

	// Always close the scope when exiting this function so we can persist any ScalewayMachine changes.
	defer func() {
		if err := machineScope.Close(ctx); err != nil && retErr == nil {
			retErr = err
		}
	}()

	// Handle deleted machines
	if !scalewayMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machineScope)
	}

	// Handle non-deleted machines
	return r.reconcileNormal(ctx, machineScope, clusterScope)
}

func (r *ScalewayMachineReconciler) reconcileNormal(ctx context.Context, machineScope *scope.Machine, clusterScope *scope.Cluster) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayMachine")

	scalewayMachine := machineScope.ScalewayMachine

	// Register our finalizer immediately to avoid orphaning Scaleway resources on delete
	if controllerutil.AddFinalizer(scalewayMachine, infrav1.MachineFinalizer) {
		if err := machineScope.PatchObject(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Make sure the Cluster Infrastructure is ready.
	if !clusterScope.Cluster.Status.InfrastructureReady {
		log.Info("Cluster infrastructure is not ready yet")
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	// Make sure bootstrap data is available and populated.
	if machineScope.Machine.Spec.Bootstrap.DataSecretName == nil {
		log.Info("Bootstrap data secret reference is not yet available")
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	if err := r.createScalewayMachineService(machineScope).Reconcile(ctx); err != nil {
		// Handle terminal & transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) {
			if reconcileError.IsTerminal() {
				log.Error(err, "Failed to reconcile ScalewayMachine")
				return ctrl.Result{}, nil
			} else if reconcileError.IsTransient() {
				log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayMachine, retrying: %s", reconcileError.Error()))
				return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
			}
		}

		return ctrl.Result{}, fmt.Errorf("failed to reconcile machine services: %w", err)
	}

	scalewayMachine.Status.Ready = true

	return ctrl.Result{}, nil
}

func (r *ScalewayMachineReconciler) reconcileDelete(ctx context.Context, machineScope *scope.Machine) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayMachine delete")

	if err := r.createScalewayMachineService(machineScope).Delete(ctx); err != nil {
		// Handle transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) {
			if reconcileError.IsTransient() {
				log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayMachine, retrying: %s", reconcileError.Error()))
				return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
			}
		}

		return ctrl.Result{}, fmt.Errorf("failed to delete machine services: %w", err)
	}

	// Machine is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(machineScope.ScalewayMachine, infrav1.MachineFinalizer)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalewayMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.ScalewayMachine{}).
		// Watch for changes to Machine and enqueue requests for ScalewayMachine
		// when the NodeRef becomes available in order to remove cloud-init user data.
		Watches(
			&clusterv1.Machine{},
			handler.EnqueueRequestsFromMapFunc(util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("ScalewayMachine"))),
			builder.WithPredicates(MachineUpdateNodeRefAvailable()),
		).
		Named("scalewaymachine").
		Complete(r)
}

// MachineUpdateNodeRefAvailable is a predicate that checks if the Machine's NodeRef has become available.
func MachineUpdateNodeRefAvailable() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldCluster, ok := e.ObjectOld.(*clusterv1.Machine)
			if !ok {
				return false
			}

			newCluster, ok := e.ObjectNew.(*clusterv1.Machine)
			if !ok {
				return false
			}

			return oldCluster.Status.NodeRef == nil && newCluster.Status.NodeRef != nil
		},
		CreateFunc:  func(event.CreateEvent) bool { return false },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}
