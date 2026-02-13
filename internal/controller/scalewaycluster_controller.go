package controller

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
)

// SecretFinalizer is the finalizer for secrets.
const SecretFinalizer = "infrastructure.cluster.x-k8s.io/caps-secret"

// ScalewayClusterReconciler reconciles a ScalewayCluster object
type ScalewayClusterReconciler struct {
	client.Client

	createScalewayClusterService scalewayClusterServiceCreator
}

// scalewayClusterServiceCreator is a function that creates a new scalewayClusterService reconciler.
type scalewayClusterServiceCreator func(clusterScope *scope.Cluster) *scalewayClusterService

// NewScalewayClusterReconciler returns a new ScalewayClusterReconciler.
func NewScalewayClusterReconciler(c client.Client) *ScalewayClusterReconciler {
	return &ScalewayClusterReconciler{
		Client:                       c,
		createScalewayClusterService: newScalewayClusterService,
	}
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewayclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewayclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewayclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
<<<<<<< HEAD
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScalewayCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *ScalewayClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)
=======
func (r *ScalewayClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := logf.FromContext(ctx)
>>>>>>> tmp-original-13-02-26-16-17

	scalewayCluster := &infrav1.ScalewayCluster{}
	if err := r.Get(ctx, req.NamespacedName, scalewayCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, scalewayCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)
	ctx = logf.IntoContext(ctx, log)

	clusterScope, err := scope.NewCluster(ctx, &scope.ClusterParams{
		Client:          r.Client,
		Cluster:         cluster,
		ScalewayCluster: scalewayCluster,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := clusterScope.Close(ctx); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if annotations.IsPaused(cluster, scalewayCluster) {
		log.Info("ScalewayCluster or linked Cluster is marked as paused. Won't reconcile normally")
		return ctrl.Result{}, nil
	}

	if err := claimScalewaySecret(ctx, r, scalewayCluster, scalewayCluster.Spec.ScalewaySecretName); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to claim ScalewaySecret: %w", err)
	}

	if !scalewayCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, clusterScope)
	}

	return r.reconcileNormal(ctx, clusterScope)
}

func (r *ScalewayClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.Cluster) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayCluster delete")

	scalewayCluster := clusterScope.ScalewayCluster

	if err := r.createScalewayClusterService(clusterScope).Delete(ctx); err != nil {
		// Handle transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) && reconcileError.RequeueAfter() != 0 {
			log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayCluster, retrying: %s", reconcileError.Error()))
			return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to delete cluster services: %w", err)
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(scalewayCluster, infrav1.ScalewayClusterFinalizer)

	if err := releaseScalewaySecret(ctx, r, scalewayCluster, scalewayCluster.Spec.ScalewaySecretName); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ScalewayClusterReconciler) reconcileNormal(ctx context.Context, clusterScope *scope.Cluster) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayCluster")
	scalewayCluster := clusterScope.ScalewayCluster

	// Register our finalizer immediately to avoid orphaning Scaleway resources on delete
	if controllerutil.AddFinalizer(scalewayCluster, infrav1.ScalewayClusterFinalizer) {
		if err := clusterScope.PatchObject(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.createScalewayClusterService(clusterScope).Reconcile(ctx); err != nil {
		// Handle terminal & transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) && reconcileError.RequeueAfter() != 0 {
			log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayCluster, retrying: %s", reconcileError.Error()))
			return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to reconcile cluster services: %w", err)
	}

	// Set APIEndpoints so the Cluster API Cluster Controller can pull them
	if scalewayCluster.Spec.ControlPlaneEndpoint.Host == "" {
		host, err := clusterScope.ControlPlaneHost()
		if err != nil {
			return ctrl.Result{}, err
		}

		scalewayCluster.Spec.ControlPlaneEndpoint.Host = host
	}
	if scalewayCluster.Spec.ControlPlaneEndpoint.Port == 0 {
		scalewayCluster.Spec.ControlPlaneEndpoint.Port = clusterScope.ControlPlaneLoadBalancerPort()
	}

	// No errors, so mark us ready so the Cluster API Cluster Controller can pull it
	scalewayCluster.Status.Initialization.Provisioned = ptr.To(true)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalewayClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.ScalewayCluster{}).
		WithEventFilter(predicates.ResourceIsNotExternallyManaged(mgr.GetScheme(), mgr.GetLogger())).
		// Add a watch on clusterv1.Cluster object for pause/unpause notifications.
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(util.ClusterToInfrastructureMapFunc(ctx, infrav1.GroupVersion.WithKind("ScalewayCluster"), mgr.GetClient(), &infrav1.ScalewayCluster{})),
			builder.WithPredicates(predicates.ClusterPausedTransitions(mgr.GetScheme(), mgr.GetLogger())),
		).
		Named("scalewaycluster").
		Complete(r)
}
