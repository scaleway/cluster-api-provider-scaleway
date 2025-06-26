package client

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
)

const privateNICID = "11111111-1111-1111-1111-111111111111"

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
					ProjectID:    scw.StringPtr(projectID),
					ResourceType: ipam.ResourceTypeInstancePrivateNic,
					ResourceID:   scw.StringPtr(privateNICID),
					IsIPv6:       scw.BoolPtr(false),
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
