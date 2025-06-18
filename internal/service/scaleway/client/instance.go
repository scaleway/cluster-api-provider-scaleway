package client

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// FindServer finds an existing Instance server by tags.
// It returns ErrNoItemFound if no matching server is found.
func (c *Client) FindServer(ctx context.Context, zone scw.Zone, tags []string) (*instance.Server, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.instance.ListServers(&instance.ListServersRequest{
		Tags:    tags,
		Project: &c.projectID,
		Zone:    zone,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListServers", err)
	}

	// Filter out all servers that have the wrong tags.
	servers := slices.DeleteFunc(resp.Servers, func(server *instance.Server) bool {
		return !matchTags(server.Tags, tags)
	})

	switch len(servers) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return servers[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d servers with tags %s", ErrTooManyItemsFound, len(servers), tags)
	}
}

func (c *Client) CreateServer(
	ctx context.Context,
	zone scw.Zone,
	name, commercialType, imageID string,
	placementGroupID *string,
	rootVolumeSize scw.Size,
	rootVolumeType instance.VolumeVolumeType,
	publicIPs, tags []string,
) (*instance.Server, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	serverType, err := c.instance.GetServerType(&instance.GetServerTypeRequest{
		Zone: zone,
		Name: commercialType,
	})
	if err != nil {
		return nil, newCallError("GetServerType", err)
	}

	req := &instance.CreateServerRequest{
		Zone:              zone,
		Name:              name,
		CommercialType:    commercialType,
		DynamicIPRequired: scw.BoolPtr(false),
		Image:             &imageID,
		PlacementGroup:    placementGroupID,
		Volumes: map[string]*instance.VolumeServerTemplate{
			"0": {
				Size:       &rootVolumeSize,
				VolumeType: rootVolumeType,
				Boot:       scw.BoolPtr(true),
			},
		},
		Tags: append(tags, createdByTag),
	}

	if len(publicIPs) > 0 {
		req.PublicIPs = &publicIPs
	}

	// Automatically attach scratch volume if server supports it.
	if serverType.ScratchStorageMaxSize != nil && *serverType.ScratchStorageMaxSize > 0 {
		req.Volumes["1"] = &instance.VolumeServerTemplate{
			Name:       scw.StringPtr(fmt.Sprintf("%s-scratch", name)),
			Size:       serverType.ScratchStorageMaxSize,
			VolumeType: instance.VolumeVolumeTypeScratch,
		}
	}

	server, err := c.instance.CreateServer(req, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateServer", err)
	}

	return server.Server, nil
}

// FindImage finds an existing Instance image by name.
// It returns ErrNoItemFound if no matching image is found.
func (c *Client) FindImage(ctx context.Context, zone scw.Zone, name string) (*instance.Image, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	resp, err := c.instance.ListImages(&instance.ListImagesRequest{
		Zone:    zone,
		Project: &c.projectID,
		Name:    scw.StringPtr(name),
		Public:  scw.BoolPtr(false),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListImages", err)
	}

	// Filter out all images that have the wrong name.
	images := slices.DeleteFunc(resp.Images, func(image *instance.Image) bool {
		return image.Name != name
	})

	switch len(images) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return images[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d images with name %s", ErrTooManyItemsFound, len(images), name)
	}
}

func (c *Client) FindIPs(ctx context.Context, zone scw.Zone, tags []string) ([]*instance.IP, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.instance.ListIPs(&instance.ListIPsRequest{
		Zone:    zone,
		Project: &c.projectID,
		Tags:    tags,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListIPs", err)
	}

	// Filter out all images that have the wrong tags.
	ips := slices.DeleteFunc(resp.IPs, func(ip *instance.IP) bool {
		return !matchTags(ip.Tags, tags)
	})

	return ips, nil
}

func (c *Client) CreateIP(ctx context.Context, zone scw.Zone, ipType instance.IPType, tags []string) (*instance.IP, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	ip, err := c.instance.CreateIP(&instance.CreateIPRequest{
		Zone: zone,
		Tags: append(tags, createdByTag),
		Type: ipType,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateIP", err)
	}

	return ip.IP, nil
}

func (c *Client) DeleteIP(ctx context.Context, zone scw.Zone, ipID string) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if err := c.instance.DeleteIP(&instance.DeleteIPRequest{
		Zone: zone,
		IP:   ipID,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteIP", err)
	}

	return nil
}

func (c *Client) CreatePrivateNIC(ctx context.Context, zone scw.Zone, serverID, privateNetworkID string) (*instance.PrivateNIC, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	privateNIC, err := c.instance.CreatePrivateNIC(&instance.CreatePrivateNICRequest{
		Zone:             zone,
		ServerID:         serverID,
		PrivateNetworkID: privateNetworkID,
		Tags:             []string{createdByTag},
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreatePrivateNIC", err)
	}

	return privateNIC.PrivateNic, nil
}

func (c *Client) GetAllServerUserData(ctx context.Context, zone scw.Zone, serverID string) (map[string]io.Reader, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	resp, err := c.instance.GetAllServerUserData(&instance.GetAllServerUserDataRequest{
		Zone:     zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("GetAllServerUserData", err)
	}

	return resp.UserData, nil
}

func (c *Client) SetServerUserData(ctx context.Context, zone scw.Zone, serverID, key, content string) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if err := c.instance.SetServerUserData(&instance.SetServerUserDataRequest{
		Zone:     zone,
		ServerID: serverID,
		Key:      key,
		Content:  strings.NewReader(content),
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("SetServerUserData", err)
	}

	return nil
}

func (c *Client) DeleteServerUserData(ctx context.Context, zone scw.Zone, serverID, key string) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if err := c.instance.DeleteServerUserData(&instance.DeleteServerUserDataRequest{
		Zone:     zone,
		ServerID: serverID,
		Key:      key,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteServerUserData", err)
	}

	return nil
}

func (c *Client) ServerAction(ctx context.Context, zone scw.Zone, serverID string, action instance.ServerAction) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if _, err := c.instance.ServerAction(&instance.ServerActionRequest{
		Zone:     zone,
		ServerID: serverID,
		Action:   action,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("ServerAction", err)
	}

	return nil
}

func (c *Client) DetachVolume(ctx context.Context, zone scw.Zone, volumeID string) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if _, err := c.instance.DetachVolume(&instance.DetachVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DetachVolume", err)
	}

	return nil
}

func (c *Client) UpdateInstanceVolumeTags(ctx context.Context, zone scw.Zone, volumeID string, tags []string) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if _, err := c.instance.UpdateVolume(&instance.UpdateVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
		Tags:     &tags,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdateVolume", err)
	}

	return nil
}

func (c *Client) FindInstanceVolume(ctx context.Context, zone scw.Zone, tags []string) (*instance.Volume, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.instance.ListVolumes(&instance.ListVolumesRequest{
		Zone: zone,
		Tags: tags,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListVolumes", err)
	}

	// Filter out all volumes that have the wrong tags.
	volumes := slices.DeleteFunc(resp.Volumes, func(volume *instance.Volume) bool {
		return !matchTags(volume.Tags, tags)
	})

	switch len(volumes) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return volumes[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d volumes with tags %s", ErrTooManyItemsFound, len(volumes), tags)
	}
}

func (c *Client) DeleteInstanceVolume(ctx context.Context, zone scw.Zone, volumeID string) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if err := c.instance.DeleteVolume(&instance.DeleteVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteVolume", err)
	}

	return nil
}

func (c *Client) DeleteServer(ctx context.Context, zone scw.Zone, serverID string) error {
	if err := c.validateZone(c.instance, zone); err != nil {
		return err
	}

	if err := c.instance.DeleteServer(&instance.DeleteServerRequest{
		Zone:     zone,
		ServerID: serverID,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteServer", err)
	}

	return nil
}

func (c *Client) FindPlacementGroup(ctx context.Context, zone scw.Zone, name string) (*instance.PlacementGroup, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	resp, err := c.instance.ListPlacementGroups(&instance.ListPlacementGroupsRequest{
		Zone:    zone,
		Name:    &name,
		Project: &c.projectID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListPlacementGroups", err)
	}

	// Filter out all placement groups that have the wrong name.
	placementGroups := slices.DeleteFunc(resp.PlacementGroups, func(pg *instance.PlacementGroup) bool {
		return pg.Name != name
	})

	switch len(placementGroups) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return placementGroups[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d placement groups with name %s", ErrTooManyItemsFound, len(placementGroups), name)
	}
}
