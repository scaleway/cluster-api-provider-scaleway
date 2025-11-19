package client

import (
	"context"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/utils/ptr"
)

type VPCGWAPI interface {
	zonesGetter

	ListGateways(req *vpcgw.ListGatewaysRequest, opts ...scw.RequestOption) (*vpcgw.ListGatewaysResponse, error)
	DeleteGateway(req *vpcgw.DeleteGatewayRequest, opts ...scw.RequestOption) (*vpcgw.Gateway, error)
	ListIPs(req *vpcgw.ListIPsRequest, opts ...scw.RequestOption) (*vpcgw.ListIPsResponse, error)
	CreateGateway(req *vpcgw.CreateGatewayRequest, opts ...scw.RequestOption) (*vpcgw.Gateway, error)
	CreateGatewayNetwork(req *vpcgw.CreateGatewayNetworkRequest, opts ...scw.RequestOption) (*vpcgw.GatewayNetwork, error)
	ListGatewayTypes(req *vpcgw.ListGatewayTypesRequest, opts ...scw.RequestOption) (*vpcgw.ListGatewayTypesResponse, error)
	UpgradeGateway(req *vpcgw.UpgradeGatewayRequest, opts ...scw.RequestOption) (*vpcgw.Gateway, error)
}

type VPCGW interface {
	FindGateways(ctx context.Context, tags []string) ([]*vpcgw.Gateway, error)
	DeleteGateway(ctx context.Context, zone scw.Zone, id string, deleteIP bool) error
	FindGatewayIP(ctx context.Context, zone scw.Zone, ip string) (*vpcgw.IP, error)
	CreateGateway(
		ctx context.Context,
		zone scw.Zone,
		name, gwType string,
		tags []string,
		ipID *string,
	) (*vpcgw.Gateway, error)
	CreateGatewayNetwork(ctx context.Context, zone scw.Zone, gatewayID, privateNetworkID string) error
	ListGatewayTypes(ctx context.Context, zone scw.Zone) ([]string, error)
	UpgradeGateway(ctx context.Context, zone scw.Zone, gatewayID, newType string) (*vpcgw.Gateway, error)
}

func (c *Client) FindGateways(ctx context.Context, tags []string) ([]*vpcgw.Gateway, error) {
	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.vpcgw.ListGateways(&vpcgw.ListGatewaysRequest{
		Zone:      scw.ZoneFrPar1, // Dummy value, refer to the scw.WithZones option.
		ProjectID: &c.projectID,
		Tags:      tags,
	}, scw.WithContext(ctx), scw.WithAllPages(), scw.WithZones(c.productZones(c.vpcgw)...))
	if err != nil {
		return nil, newCallError("ListGateways", err)
	}

	// Filter out Gateways that don't have the right tags.
	gws := slices.DeleteFunc(resp.Gateways, func(gw *vpcgw.Gateway) bool {
		return !matchTags(gw.Tags, tags)
	})

	return gws, nil
}

func (c *Client) DeleteGateway(ctx context.Context, zone scw.Zone, id string, deleteIP bool) error {
	if err := c.validateZone(c.vpcgw, zone); err != nil {
		return err
	}

	if _, err := c.vpcgw.DeleteGateway(&vpcgw.DeleteGatewayRequest{
		Zone:      zone,
		GatewayID: id,
		DeleteIP:  deleteIP,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteGateway", err)
	}

	return nil
}

func (c *Client) FindGatewayIP(ctx context.Context, zone scw.Zone, ip string) (*vpcgw.IP, error) {
	if err := c.validateZone(c.vpcgw, zone); err != nil {
		return nil, err
	}

	ips, err := c.vpcgw.ListIPs(&vpcgw.ListIPsRequest{
		Zone:      zone,
		IsFree:    ptr.To(true),
		ProjectID: &c.projectID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListIPs", err)
	}

	for _, vpcgwIP := range ips.IPs {
		if vpcgwIP.Address.String() == ip {
			return vpcgwIP, nil
		}
	}

	return nil, ErrNoItemFound
}

func (c *Client) CreateGateway(
	ctx context.Context,
	zone scw.Zone,
	name, gwType string,
	tags []string,
	ipID *string,
) (*vpcgw.Gateway, error) {
	if err := c.validateZone(c.vpcgw, zone); err != nil {
		return nil, err
	}

	gateway, err := c.vpcgw.CreateGateway(&vpcgw.CreateGatewayRequest{
		Zone: zone,
		Name: name,
		Tags: append(tags, createdByTag),
		Type: gwType,
		IPID: ipID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateGateway", err)
	}

	return gateway, nil
}

func (c *Client) CreateGatewayNetwork(ctx context.Context, zone scw.Zone, gatewayID, privateNetworkID string) error {
	if err := c.validateZone(c.vpcgw, zone); err != nil {
		return err
	}

	if _, err := c.vpcgw.CreateGatewayNetwork(&vpcgw.CreateGatewayNetworkRequest{
		Zone:             zone,
		GatewayID:        gatewayID,
		PrivateNetworkID: privateNetworkID,
		EnableMasquerade: true,
		PushDefaultRoute: true,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("CreateGatewayNetwork", err)
	}

	return nil
}

func (c *Client) ListGatewayTypes(ctx context.Context, zone scw.Zone) ([]string, error) {
	if err := c.validateZone(c.vpcgw, zone); err != nil {
		return nil, err
	}

	resp, err := c.vpcgw.ListGatewayTypes(&vpcgw.ListGatewayTypesRequest{
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("ListGatewayTypes", err)
	}

	// We assume the API returns the gateway types in the correct order (S -> M -> L, etc.).
	types := make([]string, 0, len(resp.Types))
	for _, t := range resp.Types {
		types = append(types, t.Name)
	}

	return types, nil
}

func (c *Client) UpgradeGateway(ctx context.Context, zone scw.Zone, gatewayID, newType string) (*vpcgw.Gateway, error) {
	if err := c.validateZone(c.vpcgw, zone); err != nil {
		return nil, err
	}

	gateway, err := c.vpcgw.UpgradeGateway(&vpcgw.UpgradeGatewayRequest{
		Zone:      zone,
		GatewayID: gatewayID,
		Type:      &newType,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("UpgradeGateway", err)
	}

	return gateway, nil
}
