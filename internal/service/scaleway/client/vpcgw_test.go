package client

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const vpcgwID = "11111111-1111-1111-1111-111111111111"

func TestClient_FindGateways(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*vpcgw.Gateway
		wantErr bool
		expect  func(v *mock_client.MockVPCGWAPIMockRecorder)
	}{
		{
			name: "find gateways",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				tags: []string{"tag1", "tag2"},
			},
			want: []*vpcgw.Gateway{
				{
					ID:   vpcgwID,
					Tags: []string{"tag1", "tag2"},
				},
			},
			expect: func(v *mock_client.MockVPCGWAPIMockRecorder) {
				v.Zones()
				v.ListGateways(&vpcgw.ListGatewaysRequest{
					Zone:      scw.ZoneFrPar1,
					ProjectID: ptr.To(projectID),
					Tags:      []string{"tag1", "tag2"},
				}, gomock.Any(), gomock.Any(), gomock.Any()).Return(&vpcgw.ListGatewaysResponse{
					TotalCount: 1,
					Gateways: []*vpcgw.Gateway{
						{
							ID:   vpcgwID,
							Tags: []string{"tag1", "tag2"},
						},
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

			vpcgwMock := mock_client.NewMockVPCGWAPI(mockCtrl)

			tt.expect(vpcgwMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpcgw:     vpcgwMock,
			}
			got, err := c.FindGateways(tt.args.ctx, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindGateways() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindGateways() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteGateway(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		id       string
		deleteIP bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(v *mock_client.MockVPCGWAPIMockRecorder)
	}{
		{
			name: "delete gateway",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				id:       vpcgwID,
				deleteIP: true,
			},
			expect: func(v *mock_client.MockVPCGWAPIMockRecorder) {
				v.DeleteGateway(&vpcgw.DeleteGatewayRequest{
					Zone:      scw.ZoneFrPar1,
					GatewayID: vpcgwID,
					DeleteIP:  true,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			vpcgwMock := mock_client.NewMockVPCGWAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			vpcgwMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(vpcgwMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpcgw:     vpcgwMock,
			}
			if err := c.DeleteGateway(tt.args.ctx, tt.args.zone, tt.args.id, tt.args.deleteIP); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteGateway() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_FindGatewayIP(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		ip   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *vpcgw.IP
		wantErr bool
		expect  func(v *mock_client.MockVPCGWAPIMockRecorder)
	}{
		{
			name: "find gateway IP",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				ip:   "42.42.42.42",
			},
			want: &vpcgw.IP{
				Address: net.IPv4(42, 42, 42, 42),
			},
			expect: func(v *mock_client.MockVPCGWAPIMockRecorder) {
				v.ListIPs(&vpcgw.ListIPsRequest{
					Zone:      scw.ZoneFrPar1,
					IsFree:    ptr.To(true),
					ProjectID: ptr.To(projectID),
				}, gomock.Any(), gomock.Any()).Return(&vpcgw.ListIPsResponse{
					TotalCount: 1,
					IPs: []*vpcgw.IP{
						{Address: net.IPv4(42, 42, 42, 42)},
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

			vpcgwMock := mock_client.NewMockVPCGWAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			vpcgwMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(vpcgwMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpcgw:     vpcgwMock,
			}
			got, err := c.FindGatewayIP(tt.args.ctx, tt.args.zone, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindGatewayIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindGatewayIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateGateway(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx    context.Context
		zone   scw.Zone
		name   string
		gwType string
		tags   []string
		ipID   *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *vpcgw.Gateway
		wantErr bool
		expect  func(v *mock_client.MockVPCGWAPIMockRecorder)
	}{
		{
			name: "create gateway",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:    context.TODO(),
				zone:   scw.ZoneFrPar1,
				name:   "gateway",
				gwType: "VPC-GW-S",
				tags:   []string{"tag1", "tag2"},
				ipID:   ptr.To(ipID),
			},
			want: &vpcgw.Gateway{
				ID: vpcgwID,
			},
			expect: func(v *mock_client.MockVPCGWAPIMockRecorder) {
				v.CreateGateway(&vpcgw.CreateGatewayRequest{
					Zone: scw.ZoneFrPar1,
					Name: "gateway",
					Tags: []string{"tag1", "tag2", createdByTag},
					Type: "VPC-GW-S",
					IPID: ptr.To(ipID),
				}, gomock.Any()).Return(&vpcgw.Gateway{
					ID: vpcgwID,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			vpcgwMock := mock_client.NewMockVPCGWAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			vpcgwMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(vpcgwMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpcgw:     vpcgwMock,
			}
			got, err := c.CreateGateway(tt.args.ctx, tt.args.zone, tt.args.name, tt.args.gwType, tt.args.tags, tt.args.ipID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateGateway() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateGateway() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateGatewayNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		zone             scw.Zone
		gatewayID        string
		privateNetworkID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(v *mock_client.MockVPCGWAPIMockRecorder)
	}{
		{
			name: "create gateway network",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				gatewayID:        vpcgwID,
				privateNetworkID: privateNetworkID,
			},
			expect: func(v *mock_client.MockVPCGWAPIMockRecorder) {
				v.CreateGatewayNetwork(&vpcgw.CreateGatewayNetworkRequest{
					Zone:             scw.ZoneFrPar1,
					GatewayID:        vpcgwID,
					PrivateNetworkID: privateNetworkID,
					EnableMasquerade: true,
					PushDefaultRoute: true,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			vpcgwMock := mock_client.NewMockVPCGWAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			vpcgwMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(vpcgwMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpcgw:     vpcgwMock,
			}
			if err := c.CreateGatewayNetwork(tt.args.ctx, tt.args.zone, tt.args.gatewayID, tt.args.privateNetworkID); (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateGatewayNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ListGatewayTypes(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
		expect  func(v *mock_client.MockVPCGWAPIMockRecorder)
	}{
		{
			name: "list gateway types",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
			},
			want: []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
			expect: func(v *mock_client.MockVPCGWAPIMockRecorder) {
				v.ListGatewayTypes(&vpcgw.ListGatewayTypesRequest{
					Zone: scw.ZoneFrPar1,
				}, gomock.Any()).Return(&vpcgw.ListGatewayTypesResponse{
					Types: []*vpcgw.GatewayType{
						{Name: "VPC-GW-S"},
						{Name: "VPC-GW-M"},
						{Name: "VPC-GW-L"},
						{Name: "VPC-GW-XL"},
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

			vpcgwMock := mock_client.NewMockVPCGWAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			vpcgwMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(vpcgwMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpcgw:     vpcgwMock,
			}
			got, err := c.ListGatewayTypes(tt.args.ctx, tt.args.zone)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListGatewayTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListGatewayTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_UpgradeGateway(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx       context.Context
		zone      scw.Zone
		gatewayID string
		newType   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *vpcgw.Gateway
		wantErr bool
		expect  func(v *mock_client.MockVPCGWAPIMockRecorder)
	}{
		{
			name: "upgrade gateway",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:       context.TODO(),
				zone:      scw.ZoneFrPar1,
				gatewayID: vpcgwID,
				newType:   "VPC-GW-M",
			},
			want: &vpcgw.Gateway{
				ID: vpcgwID,
			},
			expect: func(v *mock_client.MockVPCGWAPIMockRecorder) {
				v.UpgradeGateway(&vpcgw.UpgradeGatewayRequest{
					Zone:      scw.ZoneFrPar1,
					GatewayID: vpcgwID,
					Type:      ptr.To("VPC-GW-M"),
				}, gomock.Any()).Return(&vpcgw.Gateway{
					ID: vpcgwID,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			vpcgwMock := mock_client.NewMockVPCGWAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			vpcgwMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(vpcgwMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpcgw:     vpcgwMock,
			}
			got, err := c.UpgradeGateway(tt.args.ctx, tt.args.zone, tt.args.gatewayID, tt.args.newType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.UpgradeGateway() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.UpgradeGateway() = %v, want %v", got, tt.want)
			}
		})
	}
}
