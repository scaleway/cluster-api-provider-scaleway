package scope

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/scaleway-sdk-go/scw"

	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// defaultFrontendControlPlanePort is the default port for the control plane
// loadbalancer frontend.
const defaultFrontendControlPlanePort = 6443

// Cluster is a Cluster scope.
type Cluster struct {
	patchHelper *patch.Helper

	Cluster         *clusterv1.Cluster
	ScalewayCluster *infrav1.ScalewayCluster
	ScalewayClient  scwClient.Interface
}

// ClusterParams contains mandatory params for creating the Cluster scope.
type ClusterParams struct {
	Client          client.Client
	Cluster         *clusterv1.Cluster
	ScalewayCluster *infrav1.ScalewayCluster
}

// NewCluster creates a new Cluster scope.
func NewCluster(ctx context.Context, params *ClusterParams) (*Cluster, error) {
	region, err := scw.ParseRegion(params.ScalewayCluster.Spec.Region)
	if err != nil {
		return nil, fmt.Errorf("unable to parse region %q: %w", params.ScalewayCluster.Spec.Region, err)
	}

	secret := &corev1.Secret{}
	if err := params.Client.Get(ctx, client.ObjectKey{
		Name:      params.ScalewayCluster.Spec.ScalewaySecretName,
		Namespace: params.ScalewayCluster.Namespace,
	}, secret); err != nil {
		return nil, fmt.Errorf("failed to get ScalewaySecret: %w", err)
	}

	c, err := scwClient.New(region, params.ScalewayCluster.Spec.ProjectID, secret.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to create Scaleway client from ScalewaySecret: %w", err)
	}

	helper, err := patch.NewHelper(params.ScalewayCluster, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayCluster: %w", err)
	}

	return &Cluster{
		patchHelper:     helper,
		ScalewayCluster: params.ScalewayCluster,
		ScalewayClient:  c,
		Cluster:         params.Cluster,
	}, nil
}

// PatchObject patches the ScalewayCluster object.
func (c *Cluster) PatchObject(ctx context.Context) error {
	return c.patchHelper.Patch(ctx, c.ScalewayCluster)
}

// Close closes the Cluster scope by patching the ScalewayCluster object.
func (c *Cluster) Close(ctx context.Context) error {
	return c.PatchObject(ctx)
}

// ResourceNameName returns the name/prefix that resources created for the cluster should have.
// It is possible to provide additional suffixes that will be appended to the name with a leading "-".
func (c *Cluster) ResourceName(suffixes ...string) string {
	return strings.Join(append([]string{c.ScalewayCluster.Name}, suffixes...), "-")
}

// ResourceTags returns the tags that resources created for the cluster should have.
// It is possible to provide additional tags that will be added to the default tags.
func (c *Cluster) ResourceTags(additional ...string) []string {
	return append(
		[]string{
			fmt.Sprintf("caps-namespace=%s", c.ScalewayCluster.Namespace),
			fmt.Sprintf("caps-scalewaycluster=%s", c.ScalewayCluster.Name),
		}, additional...)
}

// HasPrivateNetwork returns true if the cluster has a Private Network.
func (c *Cluster) HasPrivateNetwork() bool {
	return c.ScalewayCluster.Spec.Network != nil &&
		c.ScalewayCluster.Spec.Network.PrivateNetwork.Enabled
}

// ShouldManagePrivateNetwork returns true if the provider should manage the
// Private Network of the cluster.
func (c *Cluster) ShouldManagePrivateNetwork() bool {
	return c.HasPrivateNetwork() &&
		c.ScalewayCluster.Spec.Network.PrivateNetwork != nil &&
		c.ScalewayCluster.Spec.Network.PrivateNetwork.ID == nil
}

// PrivateNetworkID returns the PrivateNetwork ID of the cluster, obtained from
// the status of the ScalewayCluster resource.
func (c *Cluster) PrivateNetworkID() (string, error) {
	if !c.HasPrivateNetwork() {
		return "", errors.New("cluster has no Private Network")
	}

	if c.ScalewayCluster.Status.Network == nil || c.ScalewayCluster.Status.Network.PrivateNetworkID == nil {
		return "", errors.New("PrivateNetworkID not found in ScalewayCluster status")
	}

	return *c.ScalewayCluster.Status.Network.PrivateNetworkID, nil
}

// ControlPlaneLoadBalancerPort returns the port to use for the control plane
// loadbalancer frontend.
func (c *Cluster) ControlPlaneLoadBalancerPort() int32 {
	var port int32 = defaultFrontendControlPlanePort

	if c.ScalewayCluster.Spec.Network != nil &&
		c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer != nil &&
		c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.Port != nil {
		port = *c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.Port
	}

	return port
}

// ControlPlaneLoadBalancerAllowedRanges returns the control plane loadbalancer
// allowed ranges.
func (c *Cluster) ControlPlaneLoadBalancerAllowedRanges() []string {
	var result []string
	if c.ScalewayCluster.Spec.Network != nil &&
		c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer != nil {
		for _, cidr := range c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.AllowedRanges {
			result = append(result, string(cidr))
		}
	}

	return result
}

// HasControlPlaneDNS returns true if the cluster has an associated domain (public or private).
func (c *Cluster) HasControlPlaneDNS() bool {
	return c.hasControlPlaneDNS() || c.hasControlPlanePrivateDNS()
}

// hasControlPlaneDNS returns true if the cluster has an associated domain that is public.
func (c *Cluster) hasControlPlaneDNS() bool {
	return c.ScalewayCluster.Spec.Network != nil &&
		c.ScalewayCluster.Spec.Network.ControlPlaneDNS != nil
}

// hasControlPlanePrivateDNS returns true if the cluster has an associated domain that is private.
func (c *Cluster) hasControlPlanePrivateDNS() bool {
	return c.ControlPlaneLoadBalancerPrivate() &&
		c.ScalewayCluster.Spec.Network.ControlPlanePrivateDNS != nil
}

// ControlPlaneDNSZoneAndName returns the DNS zone and the name of the records
// that should be updated.
func (c *Cluster) ControlPlaneDNSZoneAndName() (string, string, error) {
	if c.hasControlPlanePrivateDNS() {
		if c.ScalewayCluster.Status.Network == nil {
			return "", "", errors.New("missing network field in status")
		}

		if c.ScalewayCluster.Status.Network.VPCID == nil {
			return "", "", errors.New("missing vpcID in status")
		}

		if c.ScalewayCluster.Status.Network.PrivateNetworkID == nil {
			return "", "", errors.New("missing privateNetworkID in status")
		}

		zone := fmt.Sprintf(
			"%s.%s.privatedns",
			*c.ScalewayCluster.Status.Network.PrivateNetworkID,
			*c.ScalewayCluster.Status.Network.VPCID,
		)

		return zone, c.ScalewayCluster.Spec.Network.ControlPlanePrivateDNS.Name, nil
	}

	if c.hasControlPlaneDNS() {
		return c.ScalewayCluster.Spec.Network.ControlPlaneDNS.Domain,
			c.ScalewayCluster.Spec.Network.ControlPlaneDNS.Name, nil
	}

	return "", "", errors.New("control plane has no zone or domain")
}

// ControlPlaneHost returns the control plane host.
func (c *Cluster) ControlPlaneHost() (string, error) {
	if c.hasControlPlanePrivateDNS() {
		if c.ScalewayCluster.Status.Network == nil {
			return "", errors.New("missing network field in status")
		}

		if c.ScalewayCluster.Status.Network.PrivateNetworkID == nil {
			return "", errors.New("missing privateNetworkID in status")
		}

		return fmt.Sprintf(
			"%s.%s.internal",
			c.ScalewayCluster.Spec.Network.ControlPlanePrivateDNS.Name,
			*c.ScalewayCluster.Status.Network.PrivateNetworkID,
		), nil
	}

	if c.hasControlPlaneDNS() {
		return fmt.Sprintf(
			"%s.%s",
			c.ScalewayCluster.Spec.Network.ControlPlaneDNS.Name,
			c.ScalewayCluster.Spec.Network.ControlPlaneDNS.Domain,
		), nil
	}

	if ips := c.ControlPlaneLoadBalancerIPs(); len(ips) != 0 {
		return ips[0], nil
	}

	return "", errors.New("unable to determine control plane host")
}

// ControlPlaneLoadBalancerIPs returns the IPs of the control plane loadbalancers.
func (c *Cluster) ControlPlaneLoadBalancerIPs() []string {
	ips := make([]string, 0)

	if network := c.ScalewayCluster.Status.Network; network != nil {
		if network.LoadBalancerIP != nil {
			ips = append(ips, *network.LoadBalancerIP)
		}

		ips = append(ips, network.ExtraLoadBalancerIPs...)
	}

	return slices.Sorted(slices.Values(ips))
}

// ControlPlaneLoadBalancerPrivate returns true if the control plane should only
// be accessible through a private endpoint.
func (c *Cluster) ControlPlaneLoadBalancerPrivate() bool {
	return c.ScalewayCluster.Spec.Network != nil &&
		c.ScalewayCluster.Spec.Network.PrivateNetwork != nil &&
		c.ScalewayCluster.Spec.Network.PrivateNetwork.Enabled && // Private Network must be enabled.
		c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer != nil &&
		c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.Private != nil &&
		*c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.Private
}

// SetStatusPrivateNetworkID sets the Private Network ID in the status of the
// ScalewayCluster object.
func (c *Cluster) SetStatusPrivateNetworkID(pnID string) {
	if c.ScalewayCluster.Status.Network == nil {
		c.ScalewayCluster.Status.Network = &infrav1.NetworkStatus{
			PrivateNetworkID: &pnID,
		}
	} else {
		c.ScalewayCluster.Status.Network.PrivateNetworkID = &pnID
	}
}

// SetStatusVPCID sets the VPC ID in the status of the ScalewayCluster object.
func (c *Cluster) SetStatusVPCID(vpcID string) {
	if c.ScalewayCluster.Status.Network == nil {
		c.ScalewayCluster.Status.Network = &infrav1.NetworkStatus{
			VPCID: &vpcID,
		}
	} else {
		c.ScalewayCluster.Status.Network.VPCID = &vpcID
	}
}

// SetStatusLoadBalancerIP sets the loadbalancer IP in the status.
func (c *Cluster) SetStatusLoadBalancerIP(ip string) {
	if c.ScalewayCluster.Status.Network == nil {
		c.ScalewayCluster.Status.Network = &infrav1.NetworkStatus{
			LoadBalancerIP: &ip,
		}
	} else {
		c.ScalewayCluster.Status.Network.LoadBalancerIP = &ip
	}
}

// SetStatusExtraLoadBalancerIPs sets the extra loadbalancer IPs in the status.
func (c *Cluster) SetStatusExtraLoadBalancerIPs(ips []string) {
	if c.ScalewayCluster.Status.Network == nil {
		c.ScalewayCluster.Status.Network = &infrav1.NetworkStatus{
			ExtraLoadBalancerIPs: ips,
		}
	} else {
		c.ScalewayCluster.Status.Network.ExtraLoadBalancerIPs = ips
	}
}

// SetFailureDomains sets the failure domains of the cluster.
func (c *Cluster) SetFailureDomains(zones []scw.Zone) {
	c.ScalewayCluster.Status.FailureDomains = make(clusterv1.FailureDomains)

	for _, zone := range zones {
		c.ScalewayCluster.Status.FailureDomains[string(zone)] = clusterv1.FailureDomainSpec{
			ControlPlane: true,
		}
	}
}
