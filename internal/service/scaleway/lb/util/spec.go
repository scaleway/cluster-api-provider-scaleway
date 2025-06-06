package util

import (
	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// lbDefaultType is the default type of LB to be created if no type is provided by user.
const lbDefaultType = "LB-S"

// LBSpec returns the zone and type of the LoadBalancer based on the provided LoadBalancerSpec.
// If the zone is not specified in the spec, it defaults to the zone of the Scaleway client.
// If the type is not specified, it defaults to lbDefaultType.
func LBSpec(c *client.Client, spec infrav1.LoadBalancerSpec) (zone scw.Zone, lbType string, err error) {
	zone, err = c.GetZoneOrDefault(spec.Zone)
	if err != nil {
		return
	}

	lbType = lbDefaultType
	if spec.Type != nil {
		lbType = *spec.Type
	}

	return
}

// MainLBSpec returns the zone and type of the main LoadBalancer for the ScalewayCluster.
func MainLBSpec(c *client.Client, scalewayCluster *infrav1.ScalewayCluster) (zone scw.Zone, lbType string, err error) {
	var spec infrav1.LoadBalancerSpec
	if scalewayCluster.Spec.Network != nil && scalewayCluster.Spec.Network.ControlPlaneLoadBalancer != nil {
		spec = scalewayCluster.Spec.Network.ControlPlaneLoadBalancer.LoadBalancerSpec
	}

	return LBSpec(c, spec)
}
