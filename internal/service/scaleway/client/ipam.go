package client

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func (c *Client) FindPrivateNICIPs(ctx context.Context, privateNICID string) ([]*ipam.IP, error) {
	ips, err := c.ipam.ListIPs(&ipam.ListIPsRequest{
		ProjectID:    &c.projectID,
		ResourceType: ipam.ResourceTypeInstancePrivateNic,
		ResourceID:   &privateNICID,
		IsIPv6:       scw.BoolPtr(false),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListIPs", err)
	}

	return ips.IPs, nil
}
