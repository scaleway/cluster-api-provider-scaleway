package scope

import (
	"context"
	"errors"
	"fmt"
	"strings"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ManagedCluster struct {
	patchHelper *patch.Helper

	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane // ManagedControlPlane may be nil, on Cluster deletion.
	ScalewayClient      scwClient.Interface
}

// ClusterParams contains mandatory params for creating the Cluster scope.
type ManagedClusterParams struct {
	Client              client.Client
	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane
}

// NewManagedCluster creates a new Cluster scope.
func NewManagedCluster(ctx context.Context, params *ManagedClusterParams) (*ManagedCluster, error) {
	c, err := newScalewayClientForScalewayManagedCluster(ctx, params.Client, params.ManagedCluster)
	if err != nil {
		return nil, err
	}

	helper, err := patch.NewHelper(params.ManagedCluster, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayCluster: %w", err)
	}

	return &ManagedCluster{
		patchHelper:         helper,
		ScalewayClient:      c,
		ManagedCluster:      params.ManagedCluster,
		ManagedControlPlane: params.ManagedControlPlane,
	}, nil
}

// PatchObject patches the ScalewayManagedCluster object.
func (m *ManagedCluster) PatchObject(ctx context.Context) error {
	return m.patchHelper.Patch(ctx, m.ManagedCluster)
}

// Close closes the Machine scope by patching the ScalewayManagedCluster object.
func (m *ManagedCluster) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// ResourceName returns the name/prefix that resources created for the cluster should have.
// It is possible to provide additional suffixes that will be appended to the name with a leading "-".
func (c *ManagedCluster) ResourceName(suffixes ...string) string {
	return nameWithSuffixes(c.ManagedCluster.Name, suffixes...)
}

// ResourceTags returns the tags that resources created for the cluster should have.
// It is possible to provide additional tags that will be added to the default tags.
func (c *ManagedCluster) ResourceTags(additional ...string) []string {
	return append(
		[]string{
			fmt.Sprintf("caps-namespace=%s", c.ManagedCluster.Namespace),
			fmt.Sprintf("caps-scalewaymanagedcluster=%s", c.ManagedCluster.Name),
		}, additional...)
}

// SetCloud sets the Scaleway client object.
func (c *ManagedCluster) SetCloud(sc scwClient.Interface) {
	c.ScalewayClient = sc
}

// Cloud returns the initialized Scaleway client object.
func (c *ManagedCluster) Cloud() scwClient.Interface {
	return c.ScalewayClient
}

// HasPrivateNetwork returns true if the cluster should have a Private Network.
// It's only false if the multicloud cluster type is used.
func (c *ManagedCluster) HasPrivateNetwork() bool {
	// On Cluster deletion, we no longer have the info, we have to return true
	// to force private network cleanup.
	if c.ManagedControlPlane == nil {
		return true
	}

	return !strings.HasPrefix(c.ManagedControlPlane.Spec.Type, "multicloud")
}

// IsVPCStatusSet if the VPC fields are set in the status.
func (c *ManagedCluster) IsVPCStatusSet() bool {
	return c.ManagedCluster.Status.Network != nil &&
		c.ManagedCluster.Status.Network.PrivateNetworkID != nil
}

// PrivateNetworkParams returns the private network parameters.
func (c *ManagedCluster) PrivateNetworkParams() infrav1.PrivateNetworkParams {
	if c.ManagedCluster.Spec.Network == nil || c.ManagedCluster.Spec.Network.PrivateNetwork == nil {
		return infrav1.PrivateNetworkParams{}
	}

	return *c.ManagedCluster.Spec.Network.PrivateNetwork
}

// SetVPCStatus sets the VPC fields in the status.
func (c *ManagedCluster) SetVPCStatus(pnID, _ string) {
	if c.ManagedCluster.Status.Network == nil {
		c.ManagedCluster.Status.Network = &infrav1.ManagedNetworkStatus{}
	}

	c.ManagedCluster.Status.Network.PrivateNetworkID = &pnID
}

// PrivateNetworkID returns the PrivateNetwork ID of the managed cluster, obtained from
// the status of the ScalewayManagedCluster resource.
func (c *ManagedCluster) PrivateNetworkID() (string, error) {
	if !c.HasPrivateNetwork() {
		return "", errors.New("cluster has no Private Network")
	}

	if c.ManagedCluster.Status.Network == nil || c.ManagedCluster.Status.Network.PrivateNetworkID == nil {
		return "", errors.New("PrivateNetworkID not found in ScalewayManagedCluster status")
	}

	return *c.ManagedCluster.Status.Network.PrivateNetworkID, nil
}

// PublicGateways returns the desired Public Gateways.
func (c *ManagedCluster) PublicGateways() []infrav1.PublicGatewaySpec {
	if c.ManagedCluster.Spec.Network == nil {
		return nil
	}

	return c.ManagedCluster.Spec.Network.PublicGateways
}
