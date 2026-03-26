package client

import (
	"context"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type BlockAPI interface {
	zonesGetter

	CreateVolume(req *block.CreateVolumeRequest, opts ...scw.RequestOption) (*block.Volume, error)
	UpdateVolume(req *block.UpdateVolumeRequest, opts ...scw.RequestOption) (*block.Volume, error)
	ListVolumes(req *block.ListVolumesRequest, opts ...scw.RequestOption) (*block.ListVolumesResponse, error)
	DeleteVolume(req *block.DeleteVolumeRequest, opts ...scw.RequestOption) error
}

type Block interface {
	CreateVolume(ctx context.Context, zone scw.Zone, name string, size scw.Size, iops int64, tags []string) (*block.Volume, error)
	UpdateVolumeIOPS(ctx context.Context, zone scw.Zone, volumeID string, iops int64) error
	UpdateVolumeTags(ctx context.Context, zone scw.Zone, volumeID string, tags []string) error
	FindVolumes(ctx context.Context, zone scw.Zone, tags []string) ([]*block.Volume, error)
	DeleteVolume(ctx context.Context, zone scw.Zone, volumeID string) error
}

func (c *Client) CreateVolume(ctx context.Context, zone scw.Zone, name string, size scw.Size, iops int64, tags []string) (*block.Volume, error) {
	if err := c.validateZone(c.block, zone); err != nil {
		return nil, err
	}

	req := &block.CreateVolumeRequest{
		Zone: zone,
		Name: name,
		FromEmpty: &block.CreateVolumeRequestFromEmpty{
			Size: size,
		},
		Tags: append(tags, createdByTag),
	}

	if iops != 0 {
		req.PerfIops = scw.Uint32Ptr(uint32(iops))
	}

	volume, err := c.block.CreateVolume(req, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateVolume", err)
	}

	return volume, nil
}

func (c *Client) UpdateVolumeIOPS(ctx context.Context, zone scw.Zone, volumeID string, iops int64) error {
	if err := c.validateZone(c.block, zone); err != nil {
		return err
	}

	if _, err := c.block.UpdateVolume(&block.UpdateVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
		PerfIops: scw.Uint32Ptr(uint32(iops)),
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdateVolume", err)
	}

	return nil
}

func (c *Client) UpdateVolumeTags(ctx context.Context, zone scw.Zone, volumeID string, tags []string) error {
	if err := c.validateZone(c.block, zone); err != nil {
		return err
	}

	if _, err := c.block.UpdateVolume(&block.UpdateVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
		Tags:     &tags,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdateVolume", err)
	}

	return nil
}

func (c *Client) FindVolumes(ctx context.Context, zone scw.Zone, tags []string) ([]*block.Volume, error) {
	if err := c.validateZone(c.block, zone); err != nil {
		return nil, err
	}

	if err := validateTags(tags); err != nil {
		return nil, err
	}

	resp, err := c.block.ListVolumes(&block.ListVolumesRequest{
		Zone: zone,
		Tags: tags,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListVolumes", err)
	}

	// Filter out all volumes that have the wrong tags.
	volumes := slices.DeleteFunc(resp.Volumes, func(volume *block.Volume) bool {
		return !matchTags(volume.Tags, tags)
	})

	return volumes, nil
}

func (c *Client) DeleteVolume(ctx context.Context, zone scw.Zone, volumeID string) error {
	if err := c.validateZone(c.block, zone); err != nil {
		return err
	}

	if err := c.block.DeleteVolume(&block.DeleteVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteVolume", err)
	}

	return nil
}
