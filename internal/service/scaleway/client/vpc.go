package client

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// FindPrivateNetwork finds an existing Private Network by name.
// It returns ErrNoItemFound if no matching Private Network is found.
func (c *Client) FindPrivateNetwork(ctx context.Context, name string, vpcID *string) (*vpc.PrivateNetwork, error) {
	resp, err := c.vpc.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
		Name:      scw.StringPtr(name),
		ProjectID: &c.projectID,
		VpcID:     vpcID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListPrivateNetworks", err)
	}

	// Filter out all Private Networks that have the wrong name.
	pns := slices.DeleteFunc(resp.PrivateNetworks, func(pn *vpc.PrivateNetwork) bool {
		return pn.Name != name
	})

	switch len(pns) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return pns[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d PrivateNetworks with name %s", ErrTooManyItemsFound, len(pns), name)
	}
}

func (c *Client) DeletePrivateNetwork(ctx context.Context, id string) error {
	if err := c.vpc.DeletePrivateNetwork(&vpc.DeletePrivateNetworkRequest{
		PrivateNetworkID: id,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeletePrivateNetwork", err)
	}

	return nil
}

func (c *Client) CreatePrivateNetwork(
	ctx context.Context,
	name string,
	vpcID, subnet *string,
	tags []string,
) (*vpc.PrivateNetwork, error) {
	params := &vpc.CreatePrivateNetworkRequest{
		Name:  name,
		VpcID: vpcID,
		Tags:  tags,
	}

	if subnet != nil {
		_, ipNet, err := net.ParseCIDR(*subnet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PrivateNetwork subnet: %w", err)
		}

		params.Subnets = append(params.Subnets, scw.IPNet{IPNet: *ipNet})
	}

	pn, err := c.vpc.CreatePrivateNetwork(params, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreatePrivateNetwork", err)
	}

	return pn, nil
}
