package scope

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/scw"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
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
	helper, err := patch.NewHelper(params.ScalewayCluster, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayCluster: %w", err)
	}

	scope := &Cluster{
		patchHelper:     helper,
		ScalewayCluster: params.ScalewayCluster,
		Cluster:         params.Cluster,
	}

	scope.ScalewayClient, err = newScalewayClientForScalewayCluster(ctx, params.Client, params.ScalewayCluster)
	if err != nil {
		return nil, errors.Join(err, scope.Close(ctx))
	}

	return scope, nil
}

// PatchObject patches the ScalewayCluster object.
func (c *Cluster) PatchObject(ctx context.Context) error {
	summaryConditions := []string{
		infrav1.PrivateNetworkReadyCondition,
		infrav1.PublicGatewaysReadyCondition,
		infrav1.ScalewayClusterLoadBalancersReadyCondition,
		infrav1.ScalewayClusterDomainReadyCondition,
	}

	if err := conditions.SetSummaryCondition(c.ScalewayCluster, c.ScalewayCluster, infrav1.ScalewayClusterReadyCondition, conditions.ForConditionTypes(summaryConditions)); err != nil {
		return err
	}

	return c.patchHelper.Patch(ctx, c.ScalewayCluster, patch.WithOwnedConditions{
		Conditions: append(summaryConditions, infrav1.ScalewayClusterReadyCondition),
	})
}

// Close closes the Cluster scope by patching the ScalewayCluster object.
func (c *Cluster) Close(ctx context.Context) error {
	return c.PatchObject(ctx)
}

// ResourceName returns the name/prefix that resources created for the cluster should have.
// It is possible to provide additional suffixes that will be appended to the name with a leading "-".
func (c *Cluster) ResourceName(suffixes ...string) string {
	return nameWithSuffixes(c.ScalewayCluster.Name, suffixes...)
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

// SetCloud sets the Scaleway client object.
func (c *Cluster) SetCloud(sc scwClient.Interface) {
	c.ScalewayClient = sc
}

// Cloud returns the initialized Scaleway client object.
func (c *Cluster) Cloud() scwClient.Interface {
	return c.ScalewayClient
}

// HasPrivateNetwork returns true if the cluster has a Private Network.
func (c *Cluster) HasPrivateNetwork() bool {
	return ptr.Deref(c.ScalewayCluster.Spec.Network.PrivateNetwork.Enabled, false)
}

// PrivateNetworkParams returns the private network parameters.
func (c *Cluster) PrivateNetwork() infrav1.PrivateNetwork {
	return c.ScalewayCluster.Spec.Network.PrivateNetwork.PrivateNetwork
}

// PrivateNetworkID returns the PrivateNetwork ID of the cluster, obtained from
// the status of the ScalewayCluster resource.
func (c *Cluster) PrivateNetworkID() (string, error) {
	if !c.HasPrivateNetwork() {
		return "", errors.New("cluster has no Private Network")
	}

	if c.ScalewayCluster.Status.Network.PrivateNetworkID == "" {
		return "", errors.New("PrivateNetworkID not found in ScalewayCluster status")
	}

	return string(c.ScalewayCluster.Status.Network.PrivateNetworkID), nil
}

// ControlPlaneLoadBalancerPort returns the port to use for the control plane
// loadbalancer frontend.
func (c *Cluster) ControlPlaneLoadBalancerPort() int32 {
	var port int32 = defaultFrontendControlPlanePort

	if c.Cluster.Spec.ClusterNetwork.APIServerPort != 0 {
		port = c.Cluster.Spec.ClusterNetwork.APIServerPort
	}

	return port
}

// ControlPlaneLoadBalancerAllowedRanges returns the control plane loadbalancer
// allowed ranges.
func (c *Cluster) ControlPlaneLoadBalancerAllowedRanges() []string {
	result := make([]string, 0, len(c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.AllowedRanges))

	for _, cidr := range c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.AllowedRanges {
		result = append(result, string(cidr))
	}

	return result
}

// ControlPlaneDNSZoneAndName returns the DNS zone and the name of the records
// that should be updated.
func (c *Cluster) ControlPlaneDNSZoneAndName() (string, string, error) {
	cpDNS := c.ScalewayCluster.Spec.Network.ControlPlaneDNS
	if !cpDNS.IsDefined() {
		return "", "", errors.New("control plane has no zone or domain")
	}

	if c.ControlPlaneLoadBalancerPrivate() {
		if c.ScalewayCluster.Status.Network.VPCID == "" {
			return "", "", errors.New("missing vpcID in status")
		}

		if c.ScalewayCluster.Status.Network.PrivateNetworkID == "" {
			return "", "", errors.New("missing privateNetworkID in status")
		}

		// The domain field does not need to be set for the configuration of the
		// private zone. As a special case, we use this field to override the private
		// zone suffix.
		zoneSuffix := "privatedns"
		if cpDNS.Domain != "" && !strings.Contains(cpDNS.Domain, ".") {
			zoneSuffix = cpDNS.Domain
		}

		zone := fmt.Sprintf(
			"%s.%s.%s",
			c.ScalewayCluster.Status.Network.PrivateNetworkID,
			c.ScalewayCluster.Status.Network.VPCID,
			zoneSuffix,
		)

		return zone, cpDNS.Name, nil
	}

	return c.ScalewayCluster.Spec.Network.ControlPlaneDNS.Domain, c.ScalewayCluster.Spec.Network.ControlPlaneDNS.Name, nil
}

// ControlPlaneHost returns the control plane host.
func (c *Cluster) ControlPlaneHost() (string, error) {
	if cpDNS := c.ScalewayCluster.Spec.Network.ControlPlaneDNS; cpDNS.IsDefined() {
		if c.ControlPlaneLoadBalancerPrivate() {
			if c.ScalewayCluster.Status.Network.PrivateNetworkID == "" {
				return "", errors.New("missing privateNetworkID in status")
			}

			return fmt.Sprintf("%s.%s.internal", cpDNS.Name, c.ScalewayCluster.Status.Network.PrivateNetworkID), nil
		}

		return fmt.Sprintf("%s.%s", cpDNS.Name, cpDNS.Domain), nil
	}

	if ips := c.ControlPlaneLoadBalancerIPs(); len(ips) != 0 {
		return ips[0], nil
	}

	return "", errors.New("unable to determine control plane host")
}

// ControlPlaneLoadBalancerIPs returns the IPs of the control plane loadbalancers.
func (c *Cluster) ControlPlaneLoadBalancerIPs() []string {
	ips := make([]string, 0)

	if c.ScalewayCluster.Status.Network.LoadBalancerIP != "" {
		ips = append(ips, string(c.ScalewayCluster.Status.Network.LoadBalancerIP))
	}

	for _, ip := range c.ScalewayCluster.Status.Network.ExtraLoadBalancerIPs {
		ips = append(ips, string(ip))
	}

	return slices.Sorted(slices.Values(ips))
}

// ControlPlaneLoadBalancerPrivate returns true if the control plane should only
// be accessible through a private endpoint.
func (c *Cluster) ControlPlaneLoadBalancerPrivate() bool {
	return c.HasPrivateNetwork() && ptr.Deref(c.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.Private, false)
}

// IsVPCStatusSet if the VPC fields are set in the status.
func (c *Cluster) IsVPCStatusSet() bool {
	return c.ScalewayCluster.Status.Network.PrivateNetworkID != "" &&
		c.ScalewayCluster.Status.Network.VPCID != ""
}

// SetVPCStatus sets the VPC fields in the status.
func (c *Cluster) SetVPCStatus(pnID, vpcID string) {
	c.ScalewayCluster.Status.Network.PrivateNetworkID = infrav1.UUID(pnID)
	c.ScalewayCluster.Status.Network.VPCID = infrav1.UUID(vpcID)
}

// SetStatusLoadBalancerIP sets the loadbalancer IP in the status.
func (c *Cluster) SetStatusLoadBalancerIP(ip string) {
	c.ScalewayCluster.Status.Network.LoadBalancerIP = infrav1.IPv4(ip)
}

// SetStatusExtraLoadBalancerIPs sets the extra loadbalancer IPs in the status.
func (c *Cluster) SetStatusExtraLoadBalancerIPs(ips []string) {
	extraIPs := make([]infrav1.IPv4, 0, len(ips))

	for _, ip := range ips {
		extraIPs = append(extraIPs, infrav1.IPv4(ip))
	}

	c.ScalewayCluster.Status.Network.ExtraLoadBalancerIPs = extraIPs
}

// SetFailureDomains sets the failure domains of the cluster.
func (c *Cluster) SetFailureDomains(zones []scw.Zone) {
	failureDomains := make([]clusterv1.FailureDomain, 0, len(zones))

	for _, zone := range zones {
		failureDomains = append(failureDomains, clusterv1.FailureDomain{
			Name:         string(zone),
			ControlPlane: ptr.To(true),
		})
	}

	c.ScalewayCluster.Status.FailureDomains = failureDomains
}

// PublicGateways returns the desired Public Gateways.
func (c *Cluster) PublicGateways() []infrav1.PublicGateway {
	return c.ScalewayCluster.Spec.Network.PublicGateways
}

func (c *Cluster) SetConditions(cond []metav1.Condition) {
	c.ScalewayCluster.SetConditions(cond)
}

func (c *Cluster) GetConditions() []metav1.Condition {
	return c.ScalewayCluster.GetConditions()
}
