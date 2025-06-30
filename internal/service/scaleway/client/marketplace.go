package client

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type MarketplaceAPI interface {
	GetLocalImageByLabel(req *marketplace.GetLocalImageByLabelRequest, opts ...scw.RequestOption) (*marketplace.LocalImage, error)
}

type Marketplace interface {
	GetLocalImageByLabel(
		ctx context.Context,
		zone scw.Zone,
		commercialType,
		imageLabel string,
		imageType marketplace.LocalImageType,
	) (*marketplace.LocalImage, error)
}

func (c *Client) GetLocalImageByLabel(
	ctx context.Context,
	zone scw.Zone,
	commercialType,
	imageLabel string,
	imageType marketplace.LocalImageType,
) (*marketplace.LocalImage, error) {
	if err := c.validateZone(c.instance, zone); err != nil {
		return nil, err
	}

	image, err := c.marketplace.GetLocalImageByLabel(&marketplace.GetLocalImageByLabelRequest{
		Zone:           zone,
		CommercialType: commercialType,
		ImageLabel:     imageLabel,
		Type:           imageType,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("GetLocalImageByLabel", err)
	}

	return image, nil
}
