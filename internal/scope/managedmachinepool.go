package scope

import (
	"context"
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

type ManagedMachinePool struct {
	patchHelper                 *patch.Helper
	Client                      client.Client
	Cluster                     *clusterv1.Cluster
	MachinePool                 *clusterv1.MachinePool
	ScalewayManagedCluster      *infrav1.ScalewayManagedCluster
	ScalewayManagedControlPlane *infrav1.ScalewayManagedControlPlane
	ScalewayManagedMachinePool  *infrav1.ScalewayManagedMachinePool
	ScalewayClient              scwClient.Interface
}

// ClusterParams contains mandatory params for creating the Cluster scope.
type ManagedMachinePoolParams struct {
	Client              client.Client
	Cluster             *clusterv1.Cluster
	MachinePool         *clusterv1.MachinePool
	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane
	ManagedMachinePool  *infrav1.ScalewayManagedMachinePool
}

// NewCluster creates a new Cluster scope.
func NewManagedMachinePool(ctx context.Context, params *ManagedMachinePoolParams) (*ManagedMachinePool, error) {
	helper, err := patch.NewHelper(params.ManagedMachinePool, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayManagedMachinePool: %w", err)
	}

	mmp := &ManagedMachinePool{
		patchHelper:                 helper,
		Client:                      params.Client,
		Cluster:                     params.Cluster,
		MachinePool:                 params.MachinePool,
		ScalewayManagedCluster:      params.ManagedCluster,
		ScalewayManagedControlPlane: params.ManagedControlPlane,
		ScalewayManagedMachinePool:  params.ManagedMachinePool,
	}

	mmp.ScalewayClient, err = newScalewayClientForScalewayManagedCluster(ctx, params.Client, params.ManagedCluster)
	if err != nil {
		return nil, err
	}

	return mmp, nil
}

// PatchObject patches the ScalewayManagedControlPlane object.
func (m *ManagedMachinePool) PatchObject(ctx context.Context) error {
	summaryConditions := []string{
		infrav1.ScalewayManagedMachinePoolPoolReadyCondition,
	}

	if err := conditions.SetSummaryCondition(
		m.ScalewayManagedMachinePool, m.ScalewayManagedMachinePool,
		infrav1.ScalewayManagedMachinePoolReadyCondition,
		conditions.ForConditionTypes(summaryConditions),
	); err != nil {
		return err
	}

	return m.patchHelper.Patch(ctx, m.ScalewayManagedMachinePool, patch.WithOwnedConditions{
		Conditions: append(summaryConditions, infrav1.ScalewayManagedMachinePoolReadyCondition),
	})
}

// Close closes the Machine scope by patching the ScalewayManagedControlPlane object.
func (m *ManagedMachinePool) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// ResourceName returns the name/prefix that resources created for the cluster should have.
// It is possible to provide additional suffixes that will be appended to the name with a leading "-".
func (m *ManagedMachinePool) ResourceName(suffixes ...string) string {
	return strings.Join(append([]string{m.ScalewayManagedMachinePool.Name}, suffixes...), "-")
}

// ResourceTags returns the tags that resources created for the cluster should have.
// It is possible to provide additional tags that will be added to the default tags.
func (c *ManagedMachinePool) ResourceTags(additional ...string) []string {
	return append(
		[]string{
			fmt.Sprintf("caps-namespace=%s", c.ScalewayManagedMachinePool.Namespace),
			fmt.Sprintf("caps-scalewaymanagedmachinepool=%s", c.ScalewayManagedMachinePool.Name),
		}, additional...)
}

func (c *ManagedMachinePool) ClusterName() string {
	return c.ScalewayManagedControlPlane.Spec.ClusterName
}

func (c *ManagedMachinePool) Scaling() (autoscaling bool, size, minSize, maxSize uint32) {
	// Completely ignore scaling parameters for external node pools.
	if c.ScalewayManagedMachinePool.Spec.NodeType == "external" {
		return
	}

	scaling := c.ScalewayManagedMachinePool.Spec.Scaling

	size = c.replicas()
	autoscaling = ptr.Deref(scaling.Autoscaling, false)
	minSize = min(uint32(ptr.Deref(scaling.MinSize, 0)), size)
	maxSize = max(uint32(ptr.Deref(scaling.MaxSize, 0)), size)

	return
}

func (c *ManagedMachinePool) Autohealing() bool {
	if c.ScalewayManagedMachinePool.Spec.Autohealing == nil {
		return false
	}

	return *c.ScalewayManagedMachinePool.Spec.Autohealing
}

func (c *ManagedMachinePool) PublicIPDisabled() bool {
	if c.ScalewayManagedMachinePool.Spec.PublicIPDisabled == nil {
		return false
	}

	return *c.ScalewayManagedMachinePool.Spec.PublicIPDisabled
}

func (c *ManagedMachinePool) replicas() uint32 {
	return uint32(ptr.Deref(c.MachinePool.Spec.Replicas, 3))
}

func (c *ManagedMachinePool) RootVolumeSizeGB() *uint64 {
	if c.ScalewayManagedMachinePool.Spec.RootVolumeSizeGB == 0 {
		return nil
	}

	return ptr.To(uint64(c.ScalewayManagedMachinePool.Spec.RootVolumeSizeGB))
}

func (c *ManagedMachinePool) SetProviderIDs(nodes []*k8s.Node) {
	providerIDs := make([]string, 0, len(nodes))

	for _, node := range nodes {
		if node.ProviderID == "" {
			continue
		}

		providerIDs = append(providerIDs, node.ProviderID)
	}

	c.ScalewayManagedMachinePool.Spec.ProviderIDList = providerIDs
}

func (c *ManagedMachinePool) SetStatusReplicas(replicas uint32) {
	c.ScalewayManagedMachinePool.Status.Replicas = ptr.To(int32(replicas))
}

func (c *ManagedMachinePool) RootVolumeType() k8s.PoolVolumeType {
	if c.ScalewayManagedMachinePool.Spec.RootVolumeType == "" {
		return k8s.PoolVolumeTypeDefaultVolumeType
	}

	return k8s.PoolVolumeType(c.ScalewayManagedMachinePool.Spec.RootVolumeType)
}

func (c *ManagedMachinePool) DesiredPoolUpgradePolicy() *k8s.PoolUpgradePolicy {
	return &k8s.PoolUpgradePolicy{
		MaxSurge:       uint32(ptr.Deref(c.ScalewayManagedMachinePool.Spec.UpgradePolicy.MaxSurge, 0)),
		MaxUnavailable: uint32(ptr.Deref(c.ScalewayManagedMachinePool.Spec.UpgradePolicy.MaxUnavailable, 1)),
	}
}

func (m *ManagedMachinePool) DesiredTags() []string {
	return m.ResourceTags(m.ScalewayManagedMachinePool.Spec.AdditionalTags...)
}

func (m *ManagedMachinePool) DesiredVersion() *string {
	version := m.MachinePool.Spec.Template.Spec.Version
	if version == "" {
		return nil
	}

	version, _ = strings.CutPrefix(version, "v")
	return &version
}

func (m *ManagedMachinePool) PlacementGroupID() *string {
	if m.ScalewayManagedMachinePool.Spec.PlacementGroupID == "" {
		return nil
	}

	return ptr.To(string(m.ScalewayManagedMachinePool.Spec.PlacementGroupID))
}

func (m *ManagedMachinePool) SecurityGroupID() *string {
	if m.ScalewayManagedMachinePool.Spec.SecurityGroupID == "" {
		return nil
	}

	return ptr.To(string(m.ScalewayManagedMachinePool.Spec.SecurityGroupID))
}
