package client

import (
	"context"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type DomainAPI interface {
	ListDNSZoneRecords(req *domain.ListDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.ListDNSZoneRecordsResponse, error)
	UpdateDNSZoneRecords(req *domain.UpdateDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.UpdateDNSZoneRecordsResponse, error)
}

type Domain interface {
	ListDNSZoneRecords(ctx context.Context, zone, name string) ([]*domain.Record, error)
	DeleteDNSZoneRecords(ctx context.Context, zone string, name string) error
	SetDNSZoneRecords(ctx context.Context, zone string, name string, ips []string) error
}

func (c *Client) ListDNSZoneRecords(ctx context.Context, zone, name string) ([]*domain.Record, error) {
	resp, err := c.domain.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: zone,
		Type:    domain.RecordTypeA,
		Name:    name,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListDNSZoneRecords", err)
	}

	return resp.Records, nil
}

func (c *Client) DeleteDNSZoneRecords(ctx context.Context, zone, name string) error {
	if _, err := c.domain.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
		DNSZone:                 zone,
		DisallowNewZoneCreation: true,
		Changes: []*domain.RecordChange{
			{
				Delete: &domain.RecordChangeDelete{
					IDFields: &domain.RecordIdentifier{
						Name: name,
						Type: domain.RecordTypeA,
					},
				},
			},
		},
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdateDNSZoneRecords", err)
	}

	return nil
}

func (c *Client) SetDNSZoneRecords(ctx context.Context, zone, name string, ips []string) error {
	recordsToSet := make([]*domain.Record, 0, len(ips))

	for _, ip := range ips {
		recordsToSet = append(recordsToSet, &domain.Record{
			Data:     ip,
			Name:     name,
			Priority: 0,
			TTL:      60,
			Type:     domain.RecordTypeA,
			Comment:  scw.StringPtr(createdByDescription),
		})
	}

	if _, err := c.domain.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
		DNSZone:                 zone,
		DisallowNewZoneCreation: true,
		Changes: []*domain.RecordChange{
			{
				Set: &domain.RecordChangeSet{
					IDFields: &domain.RecordIdentifier{
						Name: name,
						Type: domain.RecordTypeA,
					},
					Records: recordsToSet,
				},
			},
		},
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdateDNSZoneRecords", err)
	}

	return nil
}
