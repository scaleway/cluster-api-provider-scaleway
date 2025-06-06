package client

import (
	"context"
	"fmt"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

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

func (c *Client) FindVolume(ctx context.Context, zone scw.Zone, tags []string) (*block.Volume, error) {
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

	switch len(volumes) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return volumes[0], nil
	default:
		return nil, fmt.Errorf("%w: found %d block volumes with tags %s", ErrTooManyItemsFound, len(volumes), tags)
	}
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
