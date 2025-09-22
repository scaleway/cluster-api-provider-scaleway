package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
)

// ScalewayManagedControlPlaneReconciler reconciles a ScalewayManagedControlPlane object
type ScalewayManagedControlPlaneReconciler struct {
	client.Client
	createScalewayManagedControlPlaneService scalewayManagedControlPlaneServiceCreator
}

// scalewayManagedControlPlaneServiceCreator is a function that creates a new scalewayManagedControlPlaneService reconciler.
type scalewayManagedControlPlaneServiceCreator func(*scope.ManagedControlPlane) *scalewayManagedControlPlaneService

func NewScalewayManagedControlPlaneReconciler(c client.Client) *ScalewayManagedControlPlaneReconciler {
	return &ScalewayManagedControlPlaneReconciler{
		Client:                                   c,
		createScalewayManagedControlPlaneService: newScalewayManagedControlPlaneService,
	}
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedcontrolplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedcontrolplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedcontrolplanes/finalizers,verbs=update
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedmachinepools,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ScalewayManagedControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := logf.FromContext(ctx)

	// Fetch the ScalewayManagedControlPlane instance
	managedControlPlane := &infrav1.ScalewayManagedControlPlane{}
	if err := r.Get(ctx, req.NamespacedName, managedControlPlane); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, managedControlPlane.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	if annotations.IsPaused(cluster, managedControlPlane) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Get the managed cluster
	managedCluster := &infrav1.ScalewayManagedCluster{}
	key := client.ObjectKey{
		Namespace: managedControlPlane.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}

	if err := r.Get(ctx, key, managedCluster); err != nil {
		log.Error(err, "Failed to retrieve ScalewayManagedCluster from the API Server")
		return ctrl.Result{}, err
	}

	managedControlPlaneScope, err := scope.NewManagedControlPlane(ctx, &scope.ManagedControlPlaneParams{
		Client:              r.Client,
		Cluster:             cluster,
		ManagedCluster:      managedCluster,
		ManagedControlPlane: managedControlPlane,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create scope: %w", err)
	}

	// Always close the scope when exiting this function so we can persist any ScalewayManagedControlPlane changes.
	defer func() {
		if err := managedControlPlaneScope.Close(ctx); err != nil && retErr == nil {
			retErr = err
		}
	}()

	// Handle deleted clusters
	if !managedControlPlane.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, managedControlPlaneScope)
	}

	return r.reconcileNormal(ctx, managedControlPlaneScope)
}

func (r *ScalewayManagedControlPlaneReconciler) reconcileNormal(ctx context.Context, s *scope.ManagedControlPlane) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayManagedControlPlane")
	managedControlPlane := s.ManagedControlPlane

	// Register our finalizer immediately to avoid orphaning Scaleway resources on delete
	if controllerutil.AddFinalizer(managedControlPlane, infrav1.ManagedControlPlaneFinalizer) {
		if err := s.PatchObject(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	if !s.ManagedCluster.Status.Ready {
		log.Info("ScalewayManagedCluster not ready yet, retry later")
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	if err := r.createScalewayManagedControlPlaneService(s).Reconcile(ctx); err != nil {
		// Handle terminal & transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) {
			if reconcileError.IsTerminal() {
				log.Error(err, "Failed to reconcile ScalewayManagedControlPlane")
				return ctrl.Result{}, nil
			} else if reconcileError.IsTransient() {
				log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayManagedControlPlane, retrying: %s", reconcileError.Error()))
				return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
			}
		}

		return ctrl.Result{}, fmt.Errorf("failed to reconcile cluster services: %w", err)
	}

	s.ManagedControlPlane.Status.Initialized = true
	s.ManagedControlPlane.Status.Ready = true
	s.ManagedControlPlane.Status.ExternalManagedControlPlane = true
	s.ManagedControlPlane.Spec.Version = s.FixedVersion()

	return ctrl.Result{}, nil
}

func (r *ScalewayManagedControlPlaneReconciler) reconcileDelete(ctx context.Context, s *scope.ManagedControlPlane) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayManagedControlPlane delete")

	managedControlPlane := s.ManagedControlPlane

	if err := r.createScalewayManagedControlPlaneService(s).Delete(ctx); err != nil {
		// Handle transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) {
			if reconcileError.IsTransient() {
				log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayManagedControlPlane, retrying: %s", reconcileError.Error()))
				return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
			}
		}

		return ctrl.Result{}, fmt.Errorf("failed to delete cluster services: %w", err)
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(managedControlPlane, infrav1.ManagedControlPlaneFinalizer)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalewayManagedControlPlaneReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.ScalewayManagedControlPlane{}).
		Named("scalewaymanagedcontrolplane").
		WithEventFilter(predicates.ResourceNotPaused(mgr.GetScheme(), mgr.GetLogger())).
		// Add a watch on clusterv1.Cluster object for unpause and infra ready notifications.
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(util.ClusterToInfrastructureMapFunc(ctx, infrav1.GroupVersion.WithKind("ScalewayManagedControlPlane"), mgr.GetClient(), &infrav1.ScalewayManagedControlPlane{})),
			builder.WithPredicates(predicates.ClusterPausedTransitionsOrInfrastructureReady(mgr.GetScheme(), mgr.GetLogger())),
		).
		Complete(r)
}
