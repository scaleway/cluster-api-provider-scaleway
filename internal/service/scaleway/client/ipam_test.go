package client

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const (
	privateNICID = "11111111-1111-1111-1111-111111111111"
	ipamIPID1    = "11111111-1111-1111-1111-111111111111"
	ipamIPID2    = "22222222-2222-2222-2222-222222222222"
)

func TestClient_FindPrivateNICIPs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx          context.Context
		privateNICID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*ipam.IP
		wantErr bool
		expect  func(d *mock_client.MockIPAMAPIMockRecorder)
	}{
		{
			name: "find private NIC IPs",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:          context.TODO(),
				privateNICID: privateNICID,
			},
			want: []*ipam.IP{
				{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
				{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)}}},
			},
			expect: func(d *mock_client.MockIPAMAPIMockRecorder) {
				d.ListIPs(&ipam.ListIPsRequest{
					ProjectID:    ptr.To(projectID),
					ResourceType: ipam.ResourceTypeInstancePrivateNic,
					ResourceID:   ptr.To(privateNICID),
					IsIPv6:       ptr.To(false),
				}, gomock.Any(), gomock.Any()).Return(&ipam.ListIPsResponse{
					TotalCount: 2,
					IPs: []*ipam.IP{
						{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
						{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)}}},
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			ipamMock := mock_client.NewMockIPAMAPI(mockCtrl)

			tt.expect(ipamMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				ipam:      ipamMock,
			}
			got, err := c.FindPrivateNICIPs(tt.args.ctx, tt.args.privateNICID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindPrivateNICIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindPrivateNICIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindLBServersIPs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		privateNetworkID string
		lbIDs            []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*ipam.IP
		wantErr bool
		expect  func(d *mock_client.MockIPAMAPIMockRecorder)
	}{
		{
			name: "find lb server ips",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				privateNetworkID: privateNetworkID,
				lbIDs:            []string{lbID},
			},
			want: []*ipam.IP{
				{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
			},
			expect: func(d *mock_client.MockIPAMAPIMockRecorder) {
				d.ListIPs(&ipam.ListIPsRequest{
					ProjectID:        ptr.To(projectID),
					ResourceType:     ipam.ResourceTypeLBServer,
					ResourceIDs:      []string{lbID},
					PrivateNetworkID: ptr.To(privateNetworkID),
					IsIPv6:           ptr.To(false),
				}, gomock.Any(), gomock.Any()).Return(&ipam.ListIPsResponse{
					TotalCount: 1,
					IPs: []*ipam.IP{
						{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			ipamMock := mock_client.NewMockIPAMAPI(mockCtrl)

			tt.expect(ipamMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				ipam:      ipamMock,
			}
			got, err := c.FindLBServersIPs(tt.args.ctx, tt.args.privateNetworkID, tt.args.lbIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindLBServersIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindLBServersIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindAvailableIPs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		privateNetworkID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*ipam.IP
		wantErr bool
		expect  func(d *mock_client.MockIPAMAPIMockRecorder)
	}{
		{
			name: "find available ips",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				privateNetworkID: privateNetworkID,
			},
			want: []*ipam.IP{
				{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
				{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)}}},
			},
			expect: func(d *mock_client.MockIPAMAPIMockRecorder) {
				d.ListIPs(&ipam.ListIPsRequest{
					ProjectID:        ptr.To(projectID),
					PrivateNetworkID: ptr.To(privateNetworkID),
					IsIPv6:           ptr.To(false),
					Attached:         ptr.To(false),
				}, gomock.Any(), gomock.Any()).Return(&ipam.ListIPsResponse{
					TotalCount: 2,
					IPs: []*ipam.IP{
						{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
						{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)}}},
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			ipamMock := mock_client.NewMockIPAMAPI(mockCtrl)

			tt.expect(ipamMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				ipam:      ipamMock,
			}
			got, err := c.FindAvailableIPs(tt.args.ctx, tt.args.privateNetworkID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindAvailableIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindAvailableIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CleanAvailableIPs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		privateNetworkID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockIPAMAPIMockRecorder)
	}{
		{
			name: "clean available ips",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				privateNetworkID: privateNetworkID,
			},
			expect: func(d *mock_client.MockIPAMAPIMockRecorder) {
				d.ListIPs(&ipam.ListIPsRequest{
					ProjectID:        ptr.To(projectID),
					PrivateNetworkID: ptr.To(privateNetworkID),
					Attached:         ptr.To(false),
				}, gomock.Any(), gomock.Any()).Return(&ipam.ListIPsResponse{
					TotalCount: 2,
					IPs: []*ipam.IP{
						{
							ID:      ipamIPID1,
							Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}},
						},
						{
							ID:      ipamIPID2,
							Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)}},
						},
					},
				}, nil)

				d.ReleaseIPSet(&ipam.ReleaseIPSetRequest{
					IPIDs: []string{ipamIPID1, ipamIPID2},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			ipamMock := mock_client.NewMockIPAMAPI(mockCtrl)

			tt.expect(ipamMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				ipam:      ipamMock,
			}
			if err := c.CleanAvailableIPs(tt.args.ctx, tt.args.privateNetworkID); (err != nil) != tt.wantErr {
				t.Errorf("Client.CleanAvailableIPs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
