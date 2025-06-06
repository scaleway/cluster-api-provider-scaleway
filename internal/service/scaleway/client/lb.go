package client

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func (c *Client) FindLB(ctx context.Context, zone scw.Zone, tags []string) (*lb.LB, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.lb.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone:      zone,
		Tags:      tags,
		ProjectID: &c.projectID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListLBs", err)
	}

	// Filter out all LBs that have the wrong tags.
	lbs := slices.DeleteFunc(resp.LBs, func(lb *lb.LB) bool {
		return !matchTags(lb.Tags, tags)
	})

	switch len(lbs) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return lbs[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d LBs with tags %s", ErrTooManyItemsFound, len(lbs), tags)
	}
}

func (c *Client) MigrateLB(ctx context.Context, zone scw.Zone, id string, newType string) (*lb.LB, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	loadbalancer, err := c.lb.MigrateLB(&lb.ZonedAPIMigrateLBRequest{
		Zone: zone,
		LBID: id,
		Type: strings.ToLower(newType),
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("MigrateLB", err)
	}

	return loadbalancer, nil
}

func (c *Client) FindLBIP(ctx context.Context, zone scw.Zone, ip string) (*lb.IP, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	ips, err := c.lb.ListIPs(&lb.ZonedAPIListIPsRequest{
		Zone:      zone,
		IPAddress: &ip,
		ProjectID: &c.projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("ListIPs", err)
	}

	for _, lbIP := range ips.IPs {
		if lbIP.IPAddress == ip {
			return lbIP, nil
		}
	}

	return nil, ErrNoItemFound
}

func (c *Client) CreateLB(
	ctx context.Context,
	zone scw.Zone,
	name, lbType string,
	ipID *string,
	tags []string,
) (*lb.LB, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	loadbalancer, err := c.lb.CreateLB(&lb.ZonedAPICreateLBRequest{
		Zone:               zone,
		Name:               name,
		Type:               lbType,
		IPID:               ipID,
		Tags:               append(tags, createdByTag),
		Description:        createdByDescription,
		AssignFlexibleIP:   scw.BoolPtr(ipID == nil),
		AssignFlexibleIPv6: scw.BoolPtr(false),
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateLB", err)
	}

	return loadbalancer, nil
}

func (c *Client) DeleteLB(ctx context.Context, zone scw.Zone, id string, releaseIP bool) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if err := c.lb.DeleteLB(&lb.ZonedAPIDeleteLBRequest{
		Zone:      zone,
		LBID:      id,
		ReleaseIP: releaseIP,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteLB", err)
	}

	return nil
}

func (c *Client) FindLBs(ctx context.Context, tags []string) ([]*lb.LB, error) {
	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.lb.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone:      scw.ZoneFrPar1, // Dummy value, refer to the scw.WithZones option.
		ProjectID: &c.projectID,
		Tags:      tags,
	}, scw.WithContext(ctx), scw.WithAllPages(), scw.WithZones(c.productZones(c.lb)...))
	if err != nil {
		return nil, newCallError("ListLBs", err)
	}

	// Filter out LBs that don't have the right prefix or tags.
	lbs := slices.DeleteFunc(resp.LBs, func(lb *lb.LB) bool {
		return !matchTags(lb.Tags, tags)
	})

	return lbs, nil
}

func (c *Client) FindBackend(ctx context.Context, zone scw.Zone, lbID, name string) (*lb.Backend, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	resp, err := c.lb.ListBackends(&lb.ZonedAPIListBackendsRequest{
		Zone: zone,
		LBID: lbID,
		Name: &name,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListBackends", err)
	}

	// Filter out all Backends that have the wrong name.
	backends := slices.DeleteFunc(resp.Backends, func(backend *lb.Backend) bool {
		return backend.Name != name
	})

	switch len(backends) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return backends[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d Backends with name %s", ErrTooManyItemsFound, len(backends), name)
	}
}

func (c *Client) CreateBackend(
	ctx context.Context,
	zone scw.Zone,
	lbID,
	name string,
	servers []string,
	port int32,
) (*lb.Backend, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	backend, err := c.lb.CreateBackend(&lb.ZonedAPICreateBackendRequest{
		Zone:            zone,
		LBID:            lbID,
		Name:            name,
		ForwardProtocol: lb.ProtocolTCP,
		ForwardPort:     port,
		HealthCheck: &lb.HealthCheck{
			Port:            port,
			CheckMaxRetries: 5,
			TCPConfig:       &lb.HealthCheckTCPConfig{},
		},
		ServerIP: servers,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateBackend", err)
	}

	return backend, nil
}

func (c *Client) SetBackendServers(
	ctx context.Context,
	zone scw.Zone,
	backendID string,
	servers []string,
) (*lb.Backend, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	backend, err := c.lb.SetBackendServers(&lb.ZonedAPISetBackendServersRequest{
		Zone:      zone,
		BackendID: backendID,
		ServerIP:  servers,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("SetBackendServers", err)
	}

	return backend, nil
}

func (c *Client) FindFrontend(ctx context.Context, zone scw.Zone, lbID, name string) (*lb.Frontend, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	resp, err := c.lb.ListFrontends(&lb.ZonedAPIListFrontendsRequest{
		Zone: zone,
		LBID: lbID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListFrontends", err)
	}

	// Filter out all Frontends that have the wrong name.
	frontends := slices.DeleteFunc(resp.Frontends, func(frontend *lb.Frontend) bool {
		return frontend.Name != name
	})

	switch len(frontends) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return frontends[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d Frontends with name %s", ErrTooManyItemsFound, len(frontends), name)
	}
}

func (c *Client) CreateFrontend(
	ctx context.Context,
	zone scw.Zone,
	lbID, name, backendID string,
	port int32,
) (*lb.Frontend, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	frontend, err := c.lb.CreateFrontend(&lb.ZonedAPICreateFrontendRequest{
		Zone:        zone,
		LBID:        lbID,
		Name:        name,
		InboundPort: port,
		BackendID:   backendID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateFrontend", err)
	}

	return frontend, nil
}

func (c *Client) FindLBPrivateNetwork(
	ctx context.Context,
	zone scw.Zone,
	lbID, privateNetworkID string,
) (*lb.PrivateNetwork, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	resp, err := c.lb.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
		Zone: zone,
		LBID: lbID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListLBPrivateNetworks", err)
	}

	// Filter out all Private Networks that have the wrong ID.
	privateNetworks := slices.DeleteFunc(resp.PrivateNetwork, func(privateNetwork *lb.PrivateNetwork) bool {
		return privateNetwork.PrivateNetworkID != privateNetworkID
	})

	switch len(privateNetworks) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return privateNetworks[0], nil
	default:
		// This should not be possible.
		return nil, fmt.Errorf(
			"%w: found %d attached Private Networks with id %s",
			ErrTooManyItemsFound,
			len(privateNetworkID),
			privateNetworkID,
		)
	}
}

func (c *Client) AttachLBPrivateNetwork(ctx context.Context, zone scw.Zone, lbID, privateNetworkID string) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if _, err := c.lb.AttachPrivateNetwork(&lb.ZonedAPIAttachPrivateNetworkRequest{
		Zone:             zone,
		LBID:             lbID,
		PrivateNetworkID: privateNetworkID,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("AttachPrivateNetwork", err)
	}

	return nil
}

func (c *Client) ListLBACLs(ctx context.Context, zone scw.Zone, frontendID string) ([]*lb.ACL, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	resp, err := c.lb.ListACLs(&lb.ZonedAPIListACLsRequest{
		Zone:       zone,
		FrontendID: frontendID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListACLs", err)
	}

	return resp.ACLs, nil
}

func (c *Client) SetLBACLs(ctx context.Context, zone scw.Zone, frontendID string, acls []*lb.ACLSpec) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if _, err := c.lb.SetACLs(&lb.ZonedAPISetACLsRequest{
		Zone:       zone,
		FrontendID: frontendID,
		ACLs:       acls,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("SetACLs", err)
	}

	return nil
}

func (c *Client) FindLBACLByName(ctx context.Context, zone scw.Zone, frontendID, name string) (*lb.ACL, error) {
	if err := c.validateZone(c.lb, zone); err != nil {
		return nil, err
	}

	resp, err := c.lb.ListACLs(&lb.ZonedAPIListACLsRequest{
		Zone:       zone,
		FrontendID: frontendID,
		Name:       &name,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListACLs", err)
	}

	// Filter out all Private Networks that have the wrong ID.
	acls := slices.DeleteFunc(resp.ACLs, func(acl *lb.ACL) bool {
		return acl.Name != name
	})

	switch len(acls) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return acls[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d ACLs with name %s", ErrTooManyItemsFound, len(acls), name)
	}
}

func (c *Client) DeleteLBACL(ctx context.Context, zone scw.Zone, aclID string) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if err := c.lb.DeleteACL(&lb.ZonedAPIDeleteACLRequest{
		Zone:  zone,
		ACLID: aclID,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteACL", err)
	}

	return nil
}

func (c *Client) CreateLBACL(
	ctx context.Context,
	zone scw.Zone,
	frontendID, name string,
	index int32,
	action lb.ACLActionType,
	matchedSubnets []string,
) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if _, err := c.lb.CreateACL(&lb.ZonedAPICreateACLRequest{
		Zone:       zone,
		FrontendID: frontendID,
		Name:       name,
		Index:      index,
		Action:     &lb.ACLAction{Type: action},
		Match:      &lb.ACLMatch{IPSubnet: scw.StringSlicePtr(matchedSubnets)},
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("CreateACL", err)
	}

	return nil
}

func (c *Client) UpdateLBACL(
	ctx context.Context,
	zone scw.Zone,
	aclID, name string,
	index int32,
	action lb.ACLActionType,
	matchedSubnets []string,
) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if _, err := c.lb.UpdateACL(&lb.ZonedAPIUpdateACLRequest{
		ACLID:  aclID,
		Zone:   zone,
		Name:   name,
		Index:  index,
		Action: &lb.ACLAction{Type: action},
		Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr(matchedSubnets)},
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdateACL", err)
	}

	return nil
}

func (c *Client) RemoveBackendServer(ctx context.Context, zone scw.Zone, backendID, ip string) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if _, err := c.lb.RemoveBackendServers(&lb.ZonedAPIRemoveBackendServersRequest{
		Zone:      zone,
		BackendID: backendID,
		ServerIP:  []string{ip},
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("RemoveBackendServers", err)
	}

	return nil
}

func (c *Client) AddBackendServer(ctx context.Context, zone scw.Zone, backendID, ip string) error {
	if err := c.validateZone(c.lb, zone); err != nil {
		return err
	}

	if _, err := c.lb.AddBackendServers(&lb.ZonedAPIAddBackendServersRequest{
		Zone:      zone,
		BackendID: backendID,
		ServerIP:  []string{ip},
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("AddBackendServers", err)
	}

	return nil
}
