package scope

import (
	"context"
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

type ManagedCluster struct {
	patchHelper *patch.Helper

	ScalewayManagedCluster      *infrav1.ScalewayManagedCluster
	ScalewayManagedControlPlane *infrav1.ScalewayManagedControlPlane // ManagedControlPlane may be nil, on Cluster deletion.
	ScalewayClient              scwClient.Interface
}

// ClusterParams contains mandatory params for creating the Cluster scope.
type ManagedClusterParams struct {
	Client              client.Client
	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane
}

// NewManagedCluster creates a new Cluster scope.
func NewManagedCluster(ctx context.Context, params *ManagedClusterParams) (*ManagedCluster, error) {
	helper, err := patch.NewHelper(params.ManagedCluster, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayCluster: %w", err)
	}

	mc := &ManagedCluster{
		patchHelper:                 helper,
		ScalewayManagedCluster:      params.ManagedCluster,
		ScalewayManagedControlPlane: params.ManagedControlPlane,
	}

	mc.ScalewayClient, err = newScalewayClientForScalewayManagedCluster(ctx, params.Client, params.ManagedCluster)
	if err != nil {
		return nil, errors.Join(err, mc.Close(ctx))
	}

	return mc, nil
}

// PatchObject patches the ScalewayManagedCluster object.
func (m *ManagedCluster) PatchObject(ctx context.Context) error {
	summaryConditions := []string{
		infrav1.PrivateNetworkReadyCondition,
		infrav1.PublicGatewaysReadyCondition,
	}

	if err := conditions.SetSummaryCondition(m.ScalewayManagedCluster, m.ScalewayManagedCluster, infrav1.ScalewayManagedClusterReadyCondition, conditions.ForConditionTypes(summaryConditions)); err != nil {
		return err
	}

	return m.patchHelper.Patch(ctx, m.ScalewayManagedCluster, patch.WithOwnedConditions{
		Conditions: append(summaryConditions, infrav1.ScalewayManagedClusterReadyCondition),
	})
}

// Close closes the Machine scope by patching the ScalewayManagedCluster object.
func (m *ManagedCluster) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// ResourceName returns the name/prefix that resources created for the cluster should have.
// It is possible to provide additional suffixes that will be appended to the name with a leading "-".
func (c *ManagedCluster) ResourceName(suffixes ...string) string {
	return nameWithSuffixes(c.ScalewayManagedCluster.Name, suffixes...)
}

// ResourceTags returns the tags that resources created for the cluster should have.
// It is possible to provide additional tags that will be added to the default tags.
func (c *ManagedCluster) ResourceTags(additional ...string) []string {
	return append(
		[]string{
			fmt.Sprintf("caps-namespace=%s", c.ScalewayManagedCluster.Namespace),
			fmt.Sprintf("caps-scalewaymanagedcluster=%s", c.ScalewayManagedCluster.Name),
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
	if c.ScalewayManagedControlPlane == nil {
		return true
	}

	return !strings.HasPrefix(c.ScalewayManagedControlPlane.Spec.Type, "multicloud")
}

// IsVPCStatusSet returns true if the VPC fields are set in the status.
func (c *ManagedCluster) IsVPCStatusSet() bool {
	return c.ScalewayManagedCluster.Status.Network.PrivateNetworkID != ""
}

// PrivateNetwork returns the private network parameters.
func (c *ManagedCluster) PrivateNetwork() infrav1.PrivateNetwork {
	return c.ScalewayManagedCluster.Spec.Network.PrivateNetwork
}

// SetVPCStatus sets the VPC fields in the status.
func (c *ManagedCluster) SetVPCStatus(pnID, _ string) {
	c.ScalewayManagedCluster.Status.Network.PrivateNetworkID = infrav1.UUID(pnID)
}

// PrivateNetworkID returns the PrivateNetwork ID of the managed cluster, obtained from
// the status of the ScalewayManagedCluster resource.
func (c *ManagedCluster) PrivateNetworkID() (string, error) {
	if !c.HasPrivateNetwork() {
		return "", errors.New("cluster has no Private Network")
	}

	if c.ScalewayManagedCluster.Status.Network.PrivateNetworkID == "" {
		return "", errors.New("PrivateNetworkID not found in ScalewayManagedCluster status")
	}

	return string(c.ScalewayManagedCluster.Status.Network.PrivateNetworkID), nil
}

// PublicGateways returns the desired Public Gateways.
func (c *ManagedCluster) PublicGateways() []infrav1.PublicGateway {
	return c.ScalewayManagedCluster.Spec.Network.PublicGateways
}

func (c *ManagedCluster) SetConditions(cond []metav1.Condition) {
	c.ScalewayManagedCluster.SetConditions(cond)
}

func (c *ManagedCluster) GetConditions() []metav1.Condition {
	return c.ScalewayManagedCluster.GetConditions()
}
