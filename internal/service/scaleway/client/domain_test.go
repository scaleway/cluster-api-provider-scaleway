package client

import (
	"context"
	"reflect"
	"testing"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const (
	zone       = "my-scw-zone.example.com"
	recordName = "my-name"
)

var (
	record1 = domain.Record{
		Name: recordName,
		Data: "127.0.0.1",
	}
	record2 = domain.Record{
		Name: recordName,
		Data: "127.0.0.2",
	}
)

func TestClient_ListDNSZoneRecords(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*domain.Record
		wantErr bool
		expect  func(d *mock_client.MockDomainAPIMockRecorder)
	}{
		{
			name: "empty list of records",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: zone,
				name: recordName,
			},
			wantErr: false,
			want:    []*domain.Record{},
			expect: func(d *mock_client.MockDomainAPIMockRecorder) {
				d.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
					DNSZone: zone,
					Name:    recordName,
					Type:    domain.RecordTypeA,
				}, gomock.Any(), gomock.Any()).Return(&domain.ListDNSZoneRecordsResponse{
					Records: []*domain.Record{},
				}, nil)
			},
		},
		{
			name: "records found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: zone,
				name: recordName,
			},
			wantErr: false,
			want:    []*domain.Record{&record1, &record2},
			expect: func(d *mock_client.MockDomainAPIMockRecorder) {
				d.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
					DNSZone: zone,
					Name:    recordName,
					Type:    domain.RecordTypeA,
				}, gomock.Any(), gomock.Any()).Return(&domain.ListDNSZoneRecordsResponse{
					Records:    []*domain.Record{&record1, &record2},
					TotalCount: 2,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			domainMock := mock_client.NewMockDomainAPI(mockCtrl)

			tt.expect(domainMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				domain:    domainMock,
			}
			got, err := c.ListDNSZoneRecords(tt.args.ctx, tt.args.zone, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListDNSZoneRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListDNSZoneRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteDNSZoneRecords(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockDomainAPIMockRecorder)
	}{
		{
			name: "delete dns zone records",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: zone,
				name: recordName,
			},
			wantErr: false,
			expect: func(d *mock_client.MockDomainAPIMockRecorder) {
				d.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
					DNSZone:                 zone,
					DisallowNewZoneCreation: true,
					Changes: []*domain.RecordChange{
						{
							Delete: &domain.RecordChangeDelete{
								IDFields: &domain.RecordIdentifier{
									Name: recordName,
									Type: domain.RecordTypeA,
								},
							},
						},
					},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			domainMock := mock_client.NewMockDomainAPI(mockCtrl)

			tt.expect(domainMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				domain:    domainMock,
			}
			if err := c.DeleteDNSZoneRecords(tt.args.ctx, tt.args.zone, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteDNSZoneRecords() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_SetDNSZoneRecords(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone string
		name string
		ips  []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockDomainAPIMockRecorder)
	}{
		{
			name: "set dns zone records",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: zone,
				name: recordName,
				ips:  []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
			},
			wantErr: false,
			expect: func(d *mock_client.MockDomainAPIMockRecorder) {
				d.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
					DNSZone:                 zone,
					DisallowNewZoneCreation: true,
					Changes: []*domain.RecordChange{
						{
							Set: &domain.RecordChangeSet{
								IDFields: &domain.RecordIdentifier{
									Name: recordName,
									Type: domain.RecordTypeA,
								},
								Records: []*domain.Record{
									{
										Data:     "127.0.0.1",
										Name:     recordName,
										Priority: 0,
										TTL:      60,
										Type:     domain.RecordTypeA,
										Comment:  ptr.To(createdByDescription),
									},
									{
										Data:     "127.0.0.2",
										Name:     recordName,
										Priority: 0,
										TTL:      60,
										Type:     domain.RecordTypeA,
										Comment:  ptr.To(createdByDescription),
									},
									{
										Data:     "127.0.0.3",
										Name:     recordName,
										Priority: 0,
										TTL:      60,
										Type:     domain.RecordTypeA,
										Comment:  ptr.To(createdByDescription),
									},
								},
							},
						},
					},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			domainMock := mock_client.NewMockDomainAPI(mockCtrl)

			tt.expect(domainMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				domain:    domainMock,
			}
			if err := c.SetDNSZoneRecords(tt.args.ctx, tt.args.zone, tt.args.name, tt.args.ips); (err != nil) != tt.wantErr {
				t.Errorf("Client.SetDNSZoneRecords() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
