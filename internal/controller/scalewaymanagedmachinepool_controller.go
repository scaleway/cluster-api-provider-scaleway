package controller

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// ScalewayManagedMachinePoolReconciler reconciles a ScalewayManagedMachinePool object
type ScalewayManagedMachinePoolReconciler struct {
	client.Client
	createScalewayManagedMachinePoolService scalewayManagedMachinePoolServiceCreator
}

// scalewayManagedControlPlaneServiceCreator is a function that creates a new scalewayManagedControlPlaneService reconciler.
type scalewayManagedMachinePoolServiceCreator func(*scope.ManagedMachinePool) *scalewayManagedMachinePoolService

func NewScalewayManagedMachinePoolReconciler(c client.Client) *ScalewayManagedMachinePoolReconciler {
	return &ScalewayManagedMachinePoolReconciler{
		Client:                                  c,
		createScalewayManagedMachinePoolService: newScalewayManagedMachinePoolService,
	}
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedmachinepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedmachinepools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedmachinepools/finalizers,verbs=update
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools;machinepools/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
<<<<<<< HEAD
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScalewayManagedMachinePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *ScalewayManagedMachinePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)
=======
func (r *ScalewayManagedMachinePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := logf.FromContext(ctx)
>>>>>>> tmp-original-13-02-26-16-17

	// Get the managed machine pool
	managedMachinePool := &infrav1.ScalewayManagedMachinePool{}
	if err := r.Get(ctx, req.NamespacedName, managedMachinePool); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the machine pool
	machinePool, err := getOwnerMachinePool(ctx, r.Client, managedMachinePool.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machinePool == nil {
		return ctrl.Result{}, nil
	}

	// Get the cluster
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machinePool.ObjectMeta)
	if err != nil {
		log.Info("Failed to retrieve Cluster from MachinePool")
		return ctrl.Result{}, err
	}
	if annotations.IsPaused(cluster, managedMachinePool) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Get the managed cluster
	managedClusterKey := client.ObjectKey{
		Namespace: managedMachinePool.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	managedCluster := &infrav1.ScalewayManagedCluster{}
	if err := r.Get(ctx, managedClusterKey, managedCluster); err != nil {
		return ctrl.Result{}, err
	}

	// Get the managed control plane
	managedControlPlaneKey := client.ObjectKey{
		Namespace: managedMachinePool.Namespace,
		Name:      cluster.Spec.ControlPlaneRef.Name,
	}
	managedControlPlane := &infrav1.ScalewayManagedControlPlane{}
	if err := r.Get(ctx, managedControlPlaneKey, managedControlPlane); err != nil {
		log.Info("Failed to retrieve ManagedControlPlane from ManagedMachinePool")
		return ctrl.Result{}, nil
	}

	managedMachinePoolScope, err := scope.NewManagedMachinePool(ctx, &scope.ManagedMachinePoolParams{
		Client:              r.Client,
		Cluster:             cluster,
		MachinePool:         machinePool,
		ManagedCluster:      managedCluster,
		ManagedControlPlane: managedControlPlane,
		ManagedMachinePool:  managedMachinePool,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create scope: %w", err)
	}

	// Always close the scope when exiting this function so we can persist any ScalewayManagedMachinePool changes.
	defer func() {
		if err := managedMachinePoolScope.Close(ctx); err != nil && retErr == nil {
			retErr = err
		}
	}()

	// Replace legacy finalizer with the up-to-date one.
	if migrateFinalizer(managedMachinePool, infrav1alpha1.ManagedMachinePoolFinalizer, infrav1.ScalewayManagedMachinePoolFinalizer) {
		if err := managedMachinePoolScope.PatchObject(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deleted machine pool
	if !managedMachinePool.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, managedMachinePoolScope)
	}

	// Handle non-deleted machine pool
	return r.reconcileNormal(ctx, managedMachinePoolScope)
}

func (r *ScalewayManagedMachinePoolReconciler) reconcileNormal(ctx context.Context, s *scope.ManagedMachinePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayManagedMachinePool")
	managedMachinePool := s.ScalewayManagedMachinePool

	// Register our finalizer immediately to avoid orphaning Scaleway resources on delete
	if controllerutil.AddFinalizer(managedMachinePool, infrav1.ScalewayManagedMachinePoolFinalizer) {
		if err := s.PatchObject(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.createScalewayManagedMachinePoolService(s).Reconcile(ctx); err != nil {
		// Handle terminal & transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) && reconcileError.RequeueAfter() != 0 {
			log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayManagedMachinePool, retrying: %s", reconcileError.Error()))
			return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to reconcile cluster services: %w", err)
	}

	s.ScalewayManagedMachinePool.Status.Initialization.Provisioned = ptr.To(true)
	s.ScalewayManagedMachinePool.Status.Ready = ptr.To(true)

	return ctrl.Result{}, nil
}

func (r *ScalewayManagedMachinePoolReconciler) reconcileDelete(ctx context.Context, s *scope.ManagedMachinePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling ScalewayManagedMachinePool delete")

	managedMachinePool := s.ScalewayManagedMachinePool

	if err := r.createScalewayManagedMachinePoolService(s).Delete(ctx); err != nil {
		// Handle transient errors
		var reconcileError *scaleway.ReconcileError
		if errors.As(err, &reconcileError) && reconcileError.RequeueAfter() != 0 {
			log.Info(fmt.Sprintf("Transient failure to reconcile ScalewayManagedMachinePool, retrying: %s", reconcileError.Error()))
			return ctrl.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to delete services: %w", err)
	}

	// Pool is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(managedMachinePool, infrav1.ScalewayManagedMachinePoolFinalizer)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalewayManagedMachinePoolReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	scalewayManagedMachinePoolMapper, err := util.ClusterToTypedObjectsMapper(r.Client, &infrav1.ScalewayManagedMachinePoolList{}, mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("failed to create mapper for Cluster to ScalewayManagedMachinePools: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.ScalewayManagedMachinePool{}).
		Named("scalewaymanagedmachinepool").
		WithEventFilter(predicates.ResourceNotPaused(mgr.GetScheme(), mgr.GetLogger())).
		// watch for changes in CAPI MachinePool resources
		Watches(
			&clusterv1.MachinePool{},
			handler.EnqueueRequestsFromMapFunc(machinePoolToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("ScalewayManagedMachinePool"))),
		).
		// watch for changes in ScalewayManagedControlPlanes
		Watches(
			&infrav1.ScalewayManagedControlPlane{},
			handler.EnqueueRequestsFromMapFunc(managedControlPlaneToManagedMachinePoolMapFunc(ctx, mgr.GetClient(), infrav1.GroupVersion.WithKind("ScalewayManagedMachinePool"))),
		).
		// Add a watch on clusterv1.Cluster object for pause/unpause & ready notifications.
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(scalewayManagedMachinePoolMapper),
			builder.WithPredicates(predicates.ClusterPausedTransitionsOrInfrastructureProvisioned(mgr.GetScheme(), mgr.GetLogger())),
		).
		Complete(r)
}

// getOwnerMachinePool returns the MachinePool object owning the current resource.
func getOwnerMachinePool(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*clusterv1.MachinePool, error) {
	for _, ref := range obj.OwnerReferences {
		if ref.Kind != "MachinePool" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse group version: %w", err)
		}
		if gv.Group == clusterv1.GroupVersion.Group {
			return getMachinePoolByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// getMachinePoolByName finds and return a Machine object using the specified params.
func getMachinePoolByName(ctx context.Context, c client.Client, namespace, name string) (*clusterv1.MachinePool, error) {
	m := &clusterv1.MachinePool{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := c.Get(ctx, key, m); err != nil {
		return nil, err
	}
	return m, nil
}

// machinePoolToInfrastructureMapFunc returns a handler.MapFunc that watches for
// MachinePool events and returns reconciliation requests for an infrastructure provider object.
func machinePoolToInfrastructureMapFunc(gvk schema.GroupVersionKind) handler.MapFunc {
	return func(_ context.Context, o client.Object) []ctrl.Request {
		m, ok := o.(*clusterv1.MachinePool)
		if !ok {
			return nil
		}

		gk := gvk.GroupKind()
		ref := m.Spec.Template.Spec.InfrastructureRef
		// Return early if the GroupKind doesn't match what we expect.
		infraGK := ref.GroupKind()
		if gk != infraGK {
			return nil
		}

		return []ctrl.Request{
			{
				NamespacedName: client.ObjectKey{
					Namespace: m.Namespace,
					Name:      ref.Name,
				},
			},
		}
	}
}

// getOwnerClusterKey returns only the Cluster name and namespace.
func getOwnerClusterKey(obj metav1.ObjectMeta) (*client.ObjectKey, error) {
	for _, ref := range obj.OwnerReferences {
		if ref.Kind != "Cluster" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse group version: %w", err)
		}
		if gv.Group == clusterv1.GroupVersion.Group {
			return &client.ObjectKey{
				Namespace: obj.Namespace,
				Name:      ref.Name,
			}, nil
		}
	}
	return nil, nil
}

func managedControlPlaneToManagedMachinePoolMapFunc(ctx context.Context, c client.Client, gvk schema.GroupVersionKind) handler.MapFunc {
	log := logf.FromContext(ctx)

	return func(ctx context.Context, o client.Object) []ctrl.Request {
		scalewayManagedControlPlane, ok := o.(*infrav1.ScalewayManagedControlPlane)
		if !ok {
			panic(fmt.Sprintf("Expected a ScalewayManagedControlPlane but got a %T", o))
		}

		if !scalewayManagedControlPlane.DeletionTimestamp.IsZero() {
			return nil
		}

		clusterKey, err := getOwnerClusterKey(scalewayManagedControlPlane.ObjectMeta)
		if err != nil {
			log.Error(err, "couldn't get ScalewayManagedControlPlane owner ObjectKey")
			return nil
		}
		if clusterKey == nil {
			return nil
		}

		managedPoolForClusterList := clusterv1.MachinePoolList{}
		if err := c.List(
			ctx, &managedPoolForClusterList, client.InNamespace(clusterKey.Namespace), client.MatchingLabels{clusterv1.ClusterNameLabel: clusterKey.Name},
		); err != nil {
			log.Error(err, "couldn't list pools for cluster")
			return nil
		}

		mapFunc := machinePoolToInfrastructureMapFunc(gvk)

		var results []ctrl.Request
		for i := range managedPoolForClusterList.Items {
			managedPool := mapFunc(ctx, &managedPoolForClusterList.Items[i])
			results = append(results, managedPool...)
		}

		return results
	}
}
