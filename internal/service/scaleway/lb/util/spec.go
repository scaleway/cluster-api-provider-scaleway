package util

import (
	"github.com/scaleway/scaleway-sdk-go/scw"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

// lbDefaultType is the default type of LB to be created if no type is provided by user.
const lbDefaultType = "LB-S"

// LBSpec returns the zone and type of the LoadBalancer based on the provided LoadBalancerSpec.
// If the zone is not specified in the spec, it defaults to the zone of the Scaleway client.
// If the type is not specified, it defaults to lbDefaultType.
func LBSpec(c client.Zones, spec infrav1.LoadBalancer) (zone scw.Zone, lbType string, err error) {
	zone, err = c.GetZoneOrDefault(string(spec.Zone))
	if err != nil {
		return
	}

	lbType = lbDefaultType
	if spec.Type != "" {
		lbType = spec.Type
	}

	return
}
