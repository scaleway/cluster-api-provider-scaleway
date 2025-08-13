package scope

import (
	"context"
	"fmt"
	"math"
	"strings"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	expclusterv1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ManagedMachinePool struct {
	patchHelper         *patch.Helper
	Client              client.Client
	Cluster             *clusterv1.Cluster
	MachinePool         *expclusterv1.MachinePool
	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane
	ManagedMachinePool  *infrav1.ScalewayManagedMachinePool
	ScalewayClient      scwClient.Interface
}

// ClusterParams contains mandatory params for creating the Cluster scope.
type ManagedMachinePoolParams struct {
	Client              client.Client
	Cluster             *clusterv1.Cluster
	MachinePool         *expclusterv1.MachinePool
	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane
	ManagedMachinePool  *infrav1.ScalewayManagedMachinePool
}

// NewCluster creates a new Cluster scope.
func NewManagedMachinePool(ctx context.Context, params *ManagedMachinePoolParams) (*ManagedMachinePool, error) {
	c, err := newScalewayClientForScalewayManagedCluster(ctx, params.Client, params.ManagedCluster)
	if err != nil {
		return nil, err
	}

	helper, err := patch.NewHelper(params.ManagedMachinePool, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayManagedMachinePool: %w", err)
	}

	return &ManagedMachinePool{
		patchHelper:         helper,
		Client:              params.Client,
		ScalewayClient:      c,
		Cluster:             params.Cluster,
		MachinePool:         params.MachinePool,
		ManagedCluster:      params.ManagedCluster,
		ManagedControlPlane: params.ManagedControlPlane,
		ManagedMachinePool:  params.ManagedMachinePool,
	}, nil
}

// PatchObject patches the ScalewayManagedControlPlane object.
func (m *ManagedMachinePool) PatchObject(ctx context.Context) error {
	return m.patchHelper.Patch(ctx, m.ManagedMachinePool)
}

// Close closes the Machine scope by patching the ScalewayManagedControlPlane object.
func (m *ManagedMachinePool) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// ResourceName returns the name/prefix that resources created for the cluster should have.
// It is possible to provide additional suffixes that will be appended to the name with a leading "-".
func (m *ManagedMachinePool) ResourceName(suffixes ...string) string {
	return strings.Join(append([]string{m.ManagedMachinePool.Name}, suffixes...), "-")
}

// ResourceTags returns the tags that resources created for the cluster should have.
// It is possible to provide additional tags that will be added to the default tags.
func (c *ManagedMachinePool) ResourceTags(additional ...string) []string {
	return append(
		[]string{
			fmt.Sprintf("caps-namespace=%s", c.ManagedMachinePool.Namespace),
			fmt.Sprintf("caps-scalewaymanagedmachinepool=%s", c.ManagedMachinePool.Name),
		}, additional...)
}

func (c *ManagedMachinePool) ClusterName() (string, bool) {
	if c.ManagedControlPlane.Spec.ClusterName == nil {
		return "", false
	}

	return *c.ManagedControlPlane.Spec.ClusterName, true
}

func (c *ManagedMachinePool) Scaling() (autoscaling bool, size, min, max uint32) {
	// Completely ignore scaling parameters for external node pools.
	if c.ManagedMachinePool.Spec.NodeType == "external" {
		return
	}

	size = c.replicas()

	if c.ManagedMachinePool.Spec.Scaling != nil {
		if c.ManagedMachinePool.Spec.Scaling.Autoscaling != nil {
			autoscaling = *c.ManagedMachinePool.Spec.Scaling.Autoscaling
		}

		if c.ManagedMachinePool.Spec.Scaling.MinSize != nil {
			min = uint32(*c.ManagedMachinePool.Spec.Scaling.MinSize)
		}

		if c.ManagedMachinePool.Spec.Scaling.MaxSize != nil {
			max = uint32(*c.ManagedMachinePool.Spec.Scaling.MaxSize)
		}
	}

	min = uint32(math.Min(float64(min), float64(size)))
	max = uint32(math.Max(float64(max), float64(size)))

	return
}

func (c *ManagedMachinePool) Autohealing() bool {
	if c.ManagedMachinePool.Spec.Autohealing == nil {
		return false
	}

	return *c.ManagedMachinePool.Spec.Autohealing
}

func (c *ManagedMachinePool) PublicIPDisabled() bool {
	if c.ManagedMachinePool.Spec.PublicIPDisabled == nil {
		return false
	}

	return *c.ManagedMachinePool.Spec.PublicIPDisabled
}

func (c *ManagedMachinePool) replicas() uint32 {
	if c.MachinePool.Spec.Replicas == nil {
		return 3
	}

	return uint32(*c.MachinePool.Spec.Replicas)
}

func (c *ManagedMachinePool) RootVolumeSizeGB() *uint64 {
	if c.ManagedMachinePool.Spec.RootVolumeSizeGB == nil {
		return nil
	}

	return scw.Uint64Ptr(uint64(*c.ManagedMachinePool.Spec.RootVolumeSizeGB))
}

func (c *ManagedMachinePool) SetProviderIDs(nodes []*k8s.Node) {
	providerIDs := make([]string, 0, len(nodes))

	for _, node := range nodes {
		if node.ProviderID == "" {
			continue
		}

		providerIDs = append(providerIDs, node.ProviderID)
	}

	c.ManagedMachinePool.Spec.ProviderIDList = providerIDs
}

func (c *ManagedMachinePool) SetStatusReplicas(replicas uint32) {
	c.ManagedMachinePool.Status.Replicas = int32(replicas)
}

func (c *ManagedMachinePool) RootVolumeType() k8s.PoolVolumeType {
	if c.ManagedMachinePool.Spec.RootVolumeType == nil {
		return k8s.PoolVolumeTypeDefaultVolumeType
	}

	return k8s.PoolVolumeType(*c.ManagedMachinePool.Spec.RootVolumeType)
}

func (c *ManagedMachinePool) DesiredPoolUpgradePolicy() *k8s.PoolUpgradePolicy {
	policy := &k8s.PoolUpgradePolicy{
		MaxSurge:       0,
		MaxUnavailable: 1,
	}

	if c.ManagedMachinePool.Spec.UpgradePolicy == nil {
		return policy
	}

	if c.ManagedMachinePool.Spec.UpgradePolicy.MaxSurge != nil {
		policy.MaxSurge = uint32(*c.ManagedMachinePool.Spec.UpgradePolicy.MaxSurge)
	}

	if c.ManagedMachinePool.Spec.UpgradePolicy.MaxUnavailable != nil {
		policy.MaxUnavailable = uint32(*c.ManagedMachinePool.Spec.UpgradePolicy.MaxUnavailable)
	}

	return policy
}

func (m *ManagedMachinePool) DesiredTags() []string {
	return m.ResourceTags(m.ManagedMachinePool.Spec.AdditionalTags...)
}

func (m *ManagedMachinePool) DesiredVersion() *string {
	version := m.MachinePool.Spec.Template.Spec.Version
	if version == nil {
		return nil
	}

	*version, _ = strings.CutPrefix(*version, "v")
	return version
}
