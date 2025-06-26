package client

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type IPAMAPI interface {
	ListIPs(req *ipam.ListIPsRequest, opts ...scw.RequestOption) (*ipam.ListIPsResponse, error)
	ReleaseIPSet(req *ipam.ReleaseIPSetRequest, opts ...scw.RequestOption) error
}

type IPAM interface {
	FindPrivateNICIPs(ctx context.Context, privateNICID string) ([]*ipam.IP, error)
	FindLBServersIPs(ctx context.Context, privateNetworkID string, lbIDs []string) ([]*ipam.IP, error)
	FindAvailableIPs(ctx context.Context, privateNetworkID string) ([]*ipam.IP, error)
	CleanAvailableIPs(ctx context.Context, privateNetworkID string) error
}

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

func (c *Client) FindLBServersIPs(ctx context.Context, privateNetworkID string, lbIDs []string) ([]*ipam.IP, error) {
	ips, err := c.ipam.ListIPs(&ipam.ListIPsRequest{
		ProjectID:        &c.projectID,
		ResourceType:     ipam.ResourceTypeLBServer,
		ResourceIDs:      lbIDs,
		PrivateNetworkID: &privateNetworkID,
		IsIPv6:           scw.BoolPtr(false),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListIPs", err)
	}

	return ips.IPs, nil
}

func (c *Client) FindAvailableIPs(ctx context.Context, privateNetworkID string) ([]*ipam.IP, error) {
	ips, err := c.ipam.ListIPs(&ipam.ListIPsRequest{
		ProjectID:        &c.projectID,
		PrivateNetworkID: &privateNetworkID,
		IsIPv6:           scw.BoolPtr(false),
		Attached:         scw.BoolPtr(false),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListIPs", err)
	}

	return ips.IPs, nil
}

func (c *Client) CleanAvailableIPs(ctx context.Context, privateNetworkID string) error {
	resp, err := c.ipam.ListIPs(&ipam.ListIPsRequest{
		ProjectID:        &c.projectID,
		PrivateNetworkID: &privateNetworkID,
		Attached:         scw.BoolPtr(false),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return newCallError("ListIPs", err)
	}

	ipIDs := make([]string, 0, len(resp.IPs))

	for _, ip := range resp.IPs {
		ipIDs = append(ipIDs, ip.ID)
	}

	if err := c.ipam.ReleaseIPSet(&ipam.ReleaseIPSetRequest{
		IPIDs: ipIDs,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("ReleaseIPSet", err)
	}

	return nil
}
