package client

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// FindPrivateNetwork finds an existing Private Network by tags.
// It returns ErrNoItemFound if no matching Private Network is found.
func (c *Client) FindPrivateNetwork(ctx context.Context, tags []string, vpcID *string) (*vpc.PrivateNetwork, error) {
	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.vpc.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
		Tags:      tags,
		ProjectID: &c.projectID,
		VpcID:     vpcID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListPrivateNetworks", err)
	}

	// Filter out all Private Networks that have the wrong tags.
	pns := slices.DeleteFunc(resp.PrivateNetworks, func(pn *vpc.PrivateNetwork) bool {
		return !matchTags(pn.Tags, tags)
	})

	switch len(pns) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return pns[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d PrivateNetworks with tags %s", ErrTooManyItemsFound, len(pns), tags)
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
		Tags:  append(tags, createdByTag),
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
