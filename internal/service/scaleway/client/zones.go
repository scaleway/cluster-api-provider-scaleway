package client

import (
	"fmt"
	"slices"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type Zones interface {
	GetZoneOrDefault(zone *string) (scw.Zone, error)
	DefaultZone() scw.Zone
	GetControlPlaneZones() []scw.Zone
}

// GetZoneOrDefault dereferences and parses the provided zone, or returns the default zone.
func (c *Client) GetZoneOrDefault(zone *string) (scw.Zone, error) {
	if zone == nil {
		return c.DefaultZone(), nil
	}

	providedZone, err := scw.ParseZone(*zone)
	if err != nil {
		return "", scaleway.WithTerminalError(fmt.Errorf("zone %s is not valid: %w", *zone, err))
	}

	return providedZone, nil
}

// DefaultZone returns the first zone of the region.
func (c *Client) DefaultZone() scw.Zone {
	return scw.Zone(fmt.Sprintf("%s-1", c.region))
}

// GetControlPlaneZones returns the availables zone for the control plane machines.
func (c *Client) GetControlPlaneZones() []scw.Zone {
	return c.productZones(c.instance)
}

type zonesGetter interface {
	Zones() []scw.Zone
}

func (c *Client) productZones(productAPI zonesGetter) []scw.Zone {
	zones := make([]scw.Zone, 0)

	for _, zone := range productAPI.Zones() {
		if r, _ := zone.Region(); r == c.region {
			zones = append(zones, zone)
		}
	}

	if len(zones) == 0 {
		zones = append(zones, c.DefaultZone())
	}

	return zones
}

func (c *Client) validateZone(productAPI zonesGetter, zone scw.Zone) error {
	zones := c.productZones(productAPI)
	if !slices.Contains(zones, zone) {
		return scaleway.WithTerminalError(fmt.Errorf("zone %s must be one of the following zones (%s)", zone, zones))
	}

	return nil
}
