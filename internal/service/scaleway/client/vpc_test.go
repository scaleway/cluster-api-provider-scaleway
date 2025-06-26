package client

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
)

const (
	vpcID      = "11111111-1111-1111-1111-111111111111"
	projectID2 = "22222222-2222-2222-2222-222222222222"
)

func TestClient_FindPrivateNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx   context.Context
		tags  []string
		vpcID *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *vpc.PrivateNetwork
		wantErr bool
		expect  func(v *mock_client.MockVPCAPIMockRecorder)
	}{
		{
			name: "no private network found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:   context.TODO(),
				tags:  []string{"tag1", "tag2", "tag3"},
				vpcID: scw.StringPtr(vpcID),
			},
			wantErr: true,
			expect: func(v *mock_client.MockVPCAPIMockRecorder) {
				v.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
					Tags:      []string{"tag1", "tag2", "tag3"},
					ProjectID: scw.StringPtr(projectID),
					VpcID:     scw.StringPtr(vpcID),
				}, gomock.Any(), gomock.Any()).Return(&vpc.ListPrivateNetworksResponse{
					PrivateNetworks: []*vpc.PrivateNetwork{},
				}, nil)
			},
		},
		{
			name: "private network found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:   context.TODO(),
				tags:  []string{"tag1", "tag2", "tag3"},
				vpcID: scw.StringPtr(vpcID),
			},
			want: &vpc.PrivateNetwork{
				ID:   privateNetworkID,
				Tags: []string{"tag1", "tag2", "tag3", "tag4"},
			},
			expect: func(v *mock_client.MockVPCAPIMockRecorder) {
				v.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
					Tags:      []string{"tag1", "tag2", "tag3"},
					ProjectID: scw.StringPtr(projectID),
					VpcID:     scw.StringPtr(vpcID),
				}, gomock.Any(), gomock.Any()).Return(&vpc.ListPrivateNetworksResponse{
					PrivateNetworks: []*vpc.PrivateNetwork{
						{
							ID:   privateNetworkID,
							Tags: []string{"tag1", "tag2", "tag3", "tag4"},
						},
					},
				}, nil)
			},
		},
		{
			name: "duplicate private networks",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:   context.TODO(),
				tags:  []string{"tag1", "tag2", "tag3"},
				vpcID: scw.StringPtr(vpcID),
			},
			wantErr: true,
			expect: func(v *mock_client.MockVPCAPIMockRecorder) {
				v.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
					Tags:      []string{"tag1", "tag2", "tag3"},
					ProjectID: scw.StringPtr(projectID),
					VpcID:     scw.StringPtr(vpcID),
				}, gomock.Any(), gomock.Any()).Return(&vpc.ListPrivateNetworksResponse{
					PrivateNetworks: []*vpc.PrivateNetwork{
						{
							Tags: []string{"tag1", "tag2", "tag3", "tag4"},
						},
						{
							Tags: []string{"tag1", "tag2", "tag3", "tag5"},
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

			vpcMock := mock_client.NewMockVPCAPI(mockCtrl)

			tt.expect(vpcMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpc:       vpcMock,
			}
			got, err := c.FindPrivateNetwork(tt.args.ctx, tt.args.tags, tt.args.vpcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindPrivateNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindPrivateNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeletePrivateNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(v *mock_client.MockVPCAPIMockRecorder)
	}{
		{
			name: "delete private network",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx: context.TODO(),
				id:  privateNetworkID,
			},
			expect: func(v *mock_client.MockVPCAPIMockRecorder) {
				v.DeletePrivateNetwork(&vpc.DeletePrivateNetworkRequest{
					PrivateNetworkID: privateNetworkID,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			vpcMock := mock_client.NewMockVPCAPI(mockCtrl)

			tt.expect(vpcMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpc:       vpcMock,
			}
			if err := c.DeletePrivateNetwork(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeletePrivateNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreatePrivateNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx    context.Context
		name   string
		vpcID  *string
		subnet *string
		tags   []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *vpc.PrivateNetwork
		wantErr bool
		expect  func(v *mock_client.MockVPCAPIMockRecorder)
	}{
		{
			name: "create private network with vpcID and subnet",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:    context.TODO(),
				name:   "privatenetwork",
				vpcID:  scw.StringPtr(vpcID),
				subnet: scw.StringPtr("192.168.1.0/24"),
				tags:   []string{"tag1", "tag2"},
			},
			want: &vpc.PrivateNetwork{
				ID:    privateNetworkID,
				VpcID: vpcID,
			},
			expect: func(v *mock_client.MockVPCAPIMockRecorder) {
				_, ipNet, err := net.ParseCIDR("192.168.1.0/24")
				if err != nil {
					panic(err)
				}

				v.CreatePrivateNetwork(&vpc.CreatePrivateNetworkRequest{
					Name:    "privatenetwork",
					VpcID:   scw.StringPtr(vpcID),
					Tags:    []string{"tag1", "tag2", createdByTag},
					Subnets: []scw.IPNet{{IPNet: *ipNet}},
				}, gomock.Any()).Return(&vpc.PrivateNetwork{
					ID:    privateNetworkID,
					VpcID: vpcID,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			vpcMock := mock_client.NewMockVPCAPI(mockCtrl)

			tt.expect(vpcMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpc:       vpcMock,
			}
			got, err := c.CreatePrivateNetwork(tt.args.ctx, tt.args.name, tt.args.vpcID, tt.args.subnet, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreatePrivateNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreatePrivateNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetPrivateNetwork(t *testing.T) {
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
		want    *vpc.PrivateNetwork
		wantErr bool
		expect  func(v *mock_client.MockVPCAPIMockRecorder)
	}{
		{
			name: "get private network",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				privateNetworkID: privateNetworkID,
			},
			want: &vpc.PrivateNetwork{
				ID:        privateNetworkID,
				ProjectID: projectID,
			},
			expect: func(v *mock_client.MockVPCAPIMockRecorder) {
				v.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
					PrivateNetworkID: privateNetworkID,
				}, gomock.Any()).Return(&vpc.PrivateNetwork{
					ID:        privateNetworkID,
					ProjectID: projectID,
				}, nil)
			},
		},
		{
			name: "wrong project ID",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				privateNetworkID: privateNetworkID,
			},
			wantErr: true,
			expect: func(v *mock_client.MockVPCAPIMockRecorder) {
				v.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
					PrivateNetworkID: privateNetworkID,
				}, gomock.Any()).Return(&vpc.PrivateNetwork{
					ID:        privateNetworkID,
					ProjectID: projectID2,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			vpcMock := mock_client.NewMockVPCAPI(mockCtrl)

			tt.expect(vpcMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				vpc:       vpcMock,
			}
			got, err := c.GetPrivateNetwork(tt.args.ctx, tt.args.privateNetworkID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetPrivateNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetPrivateNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}
