package controller

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1" //nolint:staticcheck
	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
)

// ScalewayManagedClusterReconciler reconciles a ScalewayManagedCluster object
type ScalewayManagedClusterReconciler struct {
	client.Client

	createScalewayManagedClusterService scalewayManagedClusterServiceCreator
}

// scalewayManagedClusterServiceCreator is a function that creates a new scalewayManagedClusterService reconciler.
type scalewayManagedClusterServiceCreator func(*scope.ManagedCluster) *scalewayManagedClusterService

func NewScalewayManagedClusterReconciler(c client.Client) *ScalewayManagedClusterReconciler {
	return &ScalewayManagedClusterReconciler{
		Client:                              c,
		createScalewayManagedClusterService: newScalewayManagedClusterService,
	}
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedcontrolplanes,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ScalewayManagedClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := logf.FromContext(ctx)

	managedCluster := &infrav1.ScalewayManagedCluster{}
	if err := r.Get(ctx, req.NamespacedName, managedCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, managedCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, managedCluster) {
		log.Info("ScalewayManagedCluster or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	if !cluster.Spec.ControlPlaneRef.IsDefined() {
		return ctrl.Result{}, errors.New("missing controlPlaneRef in cluster spec")
	}
	controlPlane := &infrav1.ScalewayManagedControlPlane{}
	controlPlaneRef := types.NamespacedName{
		Name:      cluster.Spec.ControlPlaneRef.Name,
		Namespace: cluster.Namespace,
	}

	if err := r.Get(ctx, controlPlaneRef, controlPlane); err != nil {
		if !apierrors.IsNotFound(err) || managedCluster.DeletionTimestamp.IsZero() {
			return ctrl.Result{}, fmt.Errorf("failed to get control plane ref: %w", err)
		}
		controlPlane = nil
	}

	log = log.WithValues("controlPlane", controlPlaneRef.Name)
	ctx = logf.IntoContext(ctx, log)

	managedClusterScope, err := scope.NewManagedCluster(ctx, &scope.ManagedClusterParams{
		Client:              r.Client,
		ManagedCluster:      managedCluster,
		ManagedControlPlane: controlPlane,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := managedClusterScope.Close(ctx); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if err := claimScalewaySecret(ctx, r, managedCluster, managedCluster.Spec.ScalewaySecretName); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to claim ScalewaySecret: %w", err)
	}

	// Replace legacy finalizer with the up-to-date one.
	if migrateFinalizer(managedCluster, infrav1alpha1.ManagedClusterFinalizer, infrav1.ScalewayManagedClusterFinalizer) {
		if err := managedClusterScope.PatchObject(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	if !managedCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, managedClusterScope)
	}

	return r.reconcileNormal(ctx, managedClusterScope)
}

func (r *ScalewayManagedClusterReconciler) reconcileNormal(ctx context.Context, s *scope.ManagedCluster) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayManagedCluster")
	managedCluster := s.ScalewayManagedCluster

	// Register our finalizer immediately to avoid orphaning Scaleway resources on delete
	if controllerutil.AddFinalizer(managedCluster, infrav1.ScalewayManagedClusterFinalizer) {
		if err := s.PatchObject(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.createScalewayManagedClusterService(s).Reconcile(ctx); err != nil {
		// Handle terminal & transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) && reconcileError.RequeueAfter() != 0 {
			log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayManagedCluster, retrying: %s", reconcileError.Error()))
			return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to reconcile cluster services: %w", err)
	}

	// Infrastructure must be ready before control plane. We should also enqueue
	// requests from control plane to infra cluster to keep control plane endpoint accurate.
	s.ScalewayManagedCluster.Status.Initialization.Provisioned = ptr.To(true)
	s.ScalewayManagedCluster.Spec.ControlPlaneEndpoint = s.ScalewayManagedControlPlane.Spec.ControlPlaneEndpoint

	return ctrl.Result{}, nil
}

func (r *ScalewayManagedClusterReconciler) reconcileDelete(ctx context.Context, s *scope.ManagedCluster) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayManagedCluster delete")

	numDependencies, err := r.dependencyCount(ctx, s)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get cluster dependencies: %w", err)
	}
	if numDependencies > 0 {
		log.V(4).Info("Scaleway managed cluster still has dependencies - requeue needed", "dependencyCount", numDependencies)
		return ctrl.Result{RequeueAfter: DefaultRetryTime}, nil
	}

	if s.ScalewayManagedControlPlane != nil {
		log.Info("ScalewayManagedControlPlane not deleted yet, retry later")
		return ctrl.Result{RequeueAfter: DefaultRetryTime}, nil
	}

	managedCluster := s.ScalewayManagedCluster

	if err := r.createScalewayManagedClusterService(s).Delete(ctx); err != nil {
		// Handle transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) && reconcileError.RequeueAfter() != 0 {
			log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayManagedCluster, retrying: %s", reconcileError.Error()))
			return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to delete cluster services: %w", err)
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(managedCluster, infrav1.ScalewayManagedClusterFinalizer)

	if err := releaseScalewaySecret(ctx, r, managedCluster, managedCluster.Spec.ScalewaySecretName); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalewayManagedClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.ScalewayManagedCluster{}).
		WithEventFilter(predicates.ResourceNotPaused(mgr.GetScheme(), mgr.GetLogger())).
		// watch ScalewayManagedControlPlane resources
		Watches(
			&infrav1.ScalewayManagedControlPlane{},
			handler.EnqueueRequestsFromMapFunc(r.managedControlPlaneMapper()),
		).
		// Add a watch on clusterv1.Cluster object for unpause notifications.
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(util.ClusterToInfrastructureMapFunc(ctx, infrav1.GroupVersion.WithKind("ScalewayManagedCluster"), mgr.GetClient(), &infrav1.ScalewayManagedCluster{})),
			builder.WithPredicates(predicates.ClusterUnpaused(mgr.GetScheme(), mgr.GetLogger())),
		).
		Named("scalewaymanagedcluster").
		Complete(r)
}

func (r *ScalewayManagedClusterReconciler) dependencyCount(ctx context.Context, clusterScope *scope.ManagedCluster) (int, error) {
	clusterName, clusterNamespace := clusterScope.ScalewayManagedCluster.Name, clusterScope.ScalewayManagedCluster.Namespace

	listOptions := []client.ListOption{
		client.InNamespace(clusterNamespace),
		client.MatchingLabels(map[string]string{clusterv1.ClusterNameLabel: clusterName}),
	}

	managedMachinePools := &infrav1.ScalewayManagedMachinePoolList{}
	if err := r.List(ctx, managedMachinePools, listOptions...); err != nil {
		return 0, fmt.Errorf("failed to list managed machine pools for cluster %s/%s: %w", clusterNamespace, clusterName, err)
	}

	return len(managedMachinePools.Items), nil
}

func (r *ScalewayManagedClusterReconciler) managedControlPlaneMapper() handler.MapFunc {
	return func(ctx context.Context, o client.Object) []ctrl.Request {
		log := logf.FromContext(ctx)

		scalewayManagedControlPlane, ok := o.(*infrav1.ScalewayManagedControlPlane)
		if !ok {
			log.Error(fmt.Errorf("expected a ScalewayManagedControlPlane, got %T instead", o), "failed to map ScalewayManagedControlPlane")
			return nil
		}

		// Don't handle deleted ScalewayManagedControlPlane
		if !scalewayManagedControlPlane.DeletionTimestamp.IsZero() {
			return nil
		}

		cluster, err := util.GetOwnerCluster(ctx, r.Client, scalewayManagedControlPlane.ObjectMeta)
		if err != nil {
			log.Error(err, "failed to get owning cluster")
			return nil
		}
		if cluster == nil {
			return nil
		}

		managedClusterRef := cluster.Spec.InfrastructureRef
		if managedClusterRef.Kind != "ScalewayManagedCluster" {
			return nil
		}

		return []ctrl.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      managedClusterRef.Name,
					Namespace: cluster.Namespace,
				},
			},
		}
	}
}
