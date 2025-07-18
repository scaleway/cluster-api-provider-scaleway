package client

import (
	"context"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
)

const (
	imageID          = "11111111-1111-1111-1111-111111111111"
	placementGroupID = "11111111-1111-1111-1111-111111111111"
	securityGroupID  = "11111111-1111-1111-1111-111111111111"
	ipID             = "11111111-1111-1111-1111-111111111111"
	serverID         = "11111111-1111-1111-1111-111111111111"
	privateNetworkID = "11111111-1111-1111-1111-111111111111"

	rootVolumeSize    = 42 * scw.GB
	scratchVolumeSize = 100 * scw.GB
)

func TestClient_FindServer(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.Server
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "no server found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2"},
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListServers(&instance.ListServersRequest{
					Zone:    scw.ZoneFrPar1,
					Tags:    []string{"tag1", "tag2"},
					Project: scw.StringPtr(projectID),
				}, gomock.Any()).Return(&instance.ListServersResponse{}, nil)
			},
		},
		{
			name: "server found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListServers(&instance.ListServersRequest{
					Zone:    scw.ZoneFrPar1,
					Tags:    []string{"tag1", "tag2"},
					Project: scw.StringPtr(projectID),
				}, gomock.Any()).Return(&instance.ListServersResponse{
					TotalCount: 1,
					Servers: []*instance.Server{
						{
							Name: "server",
							Tags: []string{"misc", "tag1", "tag2"},
						},
					},
				}, nil)
			},
			want: &instance.Server{
				Name: "server",
				Tags: []string{"misc", "tag1", "tag2"},
			},
		},
		{
			name: "duplicate servers found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListServers(&instance.ListServersRequest{
					Zone:    scw.ZoneFrPar1,
					Tags:    []string{"tag1", "tag2"},
					Project: scw.StringPtr(projectID),
				}, gomock.Any()).Return(&instance.ListServersResponse{
					TotalCount: 2,
					Servers: []*instance.Server{
						{
							Name: "server",
							Tags: []string{"misc", "tag1", "tag2"},
						},
						{
							Name: "server1",
							Tags: []string{"misc", "tag1", "tag2"},
						},
					},
				}, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.FindServer(tt.args.ctx, tt.args.zone, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateServer(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		zone             scw.Zone
		name             string
		commercialType   string
		imageID          string
		placementGroupID *string
		securityGroupID  *string
		rootVolumeSize   scw.Size
		rootVolumeType   instance.VolumeVolumeType
		tags             []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.Server
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "create server",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				name:             "server",
				commercialType:   "DEV1-S",
				imageID:          imageID,
				placementGroupID: scw.StringPtr(placementGroupID),
				securityGroupID:  scw.StringPtr(securityGroupID),
				rootVolumeSize:   rootVolumeSize,
				rootVolumeType:   instance.VolumeVolumeTypeBSSD,
				tags:             []string{"tag1", "tag2", "tag3"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.GetServerType(&instance.GetServerTypeRequest{
					Zone: scw.ZoneFrPar1,
					Name: "DEV1-S",
				}).Return(&instance.ServerType{}, nil)

				d.CreateServer(&instance.CreateServerRequest{
					Zone:              scw.ZoneFrPar1,
					Name:              "server",
					CommercialType:    "DEV1-S",
					DynamicIPRequired: scw.BoolPtr(false),
					Image:             scw.StringPtr(imageID),
					PlacementGroup:    scw.StringPtr(placementGroupID),
					SecurityGroup:     scw.StringPtr(securityGroupID),
					Volumes: map[string]*instance.VolumeServerTemplate{
						"0": {
							Size:       scw.SizePtr(rootVolumeSize),
							VolumeType: instance.VolumeVolumeTypeBSSD,
							Boot:       scw.BoolPtr(true),
						},
					},
					Tags: []string{"tag1", "tag2", "tag3", createdByTag},
				}, gomock.Any()).Return(&instance.CreateServerResponse{
					Server: &instance.Server{
						Name: "server",
					},
				}, nil)
			},
			want: &instance.Server{
				Name: "server",
			},
		},
		{
			name: "create server with scratch storage",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar2,
				name:             "server",
				commercialType:   "H100-1-80G",
				imageID:          imageID,
				placementGroupID: scw.StringPtr(placementGroupID),
				securityGroupID:  scw.StringPtr(securityGroupID),
				rootVolumeSize:   rootVolumeSize,
				rootVolumeType:   instance.VolumeVolumeTypeBSSD,
				tags:             []string{"tag1", "tag2", "tag3"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.GetServerType(&instance.GetServerTypeRequest{
					Zone: scw.ZoneFrPar2,
					Name: "H100-1-80G",
				}).Return(&instance.ServerType{
					ScratchStorageMaxSize: scw.SizePtr(scratchVolumeSize),
				}, nil)

				d.CreateServer(&instance.CreateServerRequest{
					Zone:              scw.ZoneFrPar2,
					Name:              "server",
					CommercialType:    "H100-1-80G",
					DynamicIPRequired: scw.BoolPtr(false),
					Image:             scw.StringPtr(imageID),
					PlacementGroup:    scw.StringPtr(placementGroupID),
					SecurityGroup:     scw.StringPtr(securityGroupID),
					Volumes: map[string]*instance.VolumeServerTemplate{
						"0": {
							Size:       scw.SizePtr(rootVolumeSize),
							VolumeType: instance.VolumeVolumeTypeBSSD,
							Boot:       scw.BoolPtr(true),
						},
						"1": {
							Name:       scw.StringPtr("server-scratch"),
							Size:       scw.SizePtr(scratchVolumeSize),
							VolumeType: instance.VolumeVolumeTypeScratch,
						},
					},
					Tags: []string{"tag1", "tag2", "tag3", createdByTag},
				}, gomock.Any()).Return(&instance.CreateServerResponse{
					Server: &instance.Server{
						Name: "server",
					},
				}, nil)
			},
			want: &instance.Server{
				Name: "server",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.CreateServer(tt.args.ctx, tt.args.zone, tt.args.name, tt.args.commercialType, tt.args.imageID, tt.args.placementGroupID, tt.args.securityGroupID, tt.args.rootVolumeSize, tt.args.rootVolumeType, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindImage(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.Image
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "no image found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "my-image",
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListImages(&instance.ListImagesRequest{
					Zone:    scw.ZoneFrPar1,
					Project: scw.StringPtr(projectID),
					Name:    scw.StringPtr("my-image"),
					Public:  scw.BoolPtr(false),
				}, gomock.Any()).Return(&instance.ListImagesResponse{}, nil)
			},
		},
		{
			name: "image found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "my-image",
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListImages(&instance.ListImagesRequest{
					Zone:    scw.ZoneFrPar1,
					Project: scw.StringPtr(projectID),
					Name:    scw.StringPtr("my-image"),
					Public:  scw.BoolPtr(false),
				}, gomock.Any()).Return(&instance.ListImagesResponse{
					TotalCount: 1,
					Images: []*instance.Image{
						{
							Name: "my-image",
						},
					},
				}, nil)
			},
			want: &instance.Image{
				Name: "my-image",
			},
		},
		{
			name: "duplicate images found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "my-image",
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListImages(&instance.ListImagesRequest{
					Zone:    scw.ZoneFrPar1,
					Project: scw.StringPtr(projectID),
					Name:    scw.StringPtr("my-image"),
					Public:  scw.BoolPtr(false),
				}, gomock.Any()).Return(&instance.ListImagesResponse{
					TotalCount: 2,
					Images: []*instance.Image{
						{
							Name: "my-image",
						},
						{
							Name: "my-image",
						},
					},
				}, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.FindImage(tt.args.ctx, tt.args.zone, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindIPs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*instance.IP
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "no ip found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListIPs(&instance.ListIPsRequest{
					Zone:    scw.ZoneFrPar1,
					Project: scw.StringPtr(projectID),
					Tags:    []string{"tag1", "tag2", "tag3"},
				}, gomock.Any(), gomock.Any()).Return(&instance.ListIPsResponse{}, nil)
			},
		},
		{
			name: "some ips found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListIPs(&instance.ListIPsRequest{
					Zone:    scw.ZoneFrPar1,
					Project: scw.StringPtr(projectID),
					Tags:    []string{"tag1", "tag2", "tag3"},
				}, gomock.Any(), gomock.Any()).Return(&instance.ListIPsResponse{
					TotalCount: 2,
					IPs: []*instance.IP{
						{
							Address: net.IPv4(42, 42, 42, 42),
							Tags:    []string{"tag1", "tag2", "tag3"},
							Type:    instance.IPTypeRoutedIPv4,
						},
						{
							Address: net.IPv4(43, 43, 43, 43),
							Tags:    []string{"tag1", "tag2", "tag3"},
							Type:    instance.IPTypeRoutedIPv4,
						},
					},
				}, nil)
			},
			want: []*instance.IP{
				{
					Address: net.IPv4(42, 42, 42, 42),
					Tags:    []string{"tag1", "tag2", "tag3"},
					Type:    instance.IPTypeRoutedIPv4,
				},
				{
					Address: net.IPv4(43, 43, 43, 43),
					Tags:    []string{"tag1", "tag2", "tag3"},
					Type:    instance.IPTypeRoutedIPv4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.FindIPs(tt.args.ctx, tt.args.zone, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateIP(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx    context.Context
		zone   scw.Zone
		ipType instance.IPType
		tags   []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.IP
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "create IP",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:    context.TODO(),
				zone:   scw.ZoneFrPar1,
				ipType: instance.IPTypeRoutedIPv4,
				tags:   []string{"tag1", "tag2"},
			},
			want: &instance.IP{
				Address: net.IPv4(42, 42, 42, 42),
				Tags:    []string{"tag1", "tag2", createdByTag},
				Type:    instance.IPTypeRoutedIPv4,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.CreateIP(&instance.CreateIPRequest{
					Zone: scw.ZoneFrPar1,
					Tags: []string{"tag1", "tag2", createdByTag},
					Type: instance.IPTypeRoutedIPv4,
				}, gomock.Any()).Return(&instance.CreateIPResponse{
					IP: &instance.IP{
						Address: net.IPv4(42, 42, 42, 42),
						Tags:    []string{"tag1", "tag2", createdByTag},
						Type:    instance.IPTypeRoutedIPv4,
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

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.CreateIP(tt.args.ctx, tt.args.zone, tt.args.ipType, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteIP(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		ipID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "delete IP",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				ipID: ipID,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.DeleteIP(&instance.DeleteIPRequest{
					Zone: scw.ZoneFrPar1,
					IP:   ipID,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.DeleteIP(tt.args.ctx, tt.args.zone, tt.args.ipID); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteIP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreatePrivateNIC(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx              context.Context
		zone             scw.Zone
		serverID         string
		privateNetworkID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.PrivateNIC
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "create private nic",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				serverID:         serverID,
				privateNetworkID: privateNetworkID,
			},
			want: &instance.PrivateNIC{
				ServerID:         serverID,
				PrivateNetworkID: privateNetworkID,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.CreatePrivateNIC(&instance.CreatePrivateNICRequest{
					Zone:             scw.ZoneFrPar1,
					ServerID:         serverID,
					PrivateNetworkID: privateNetworkID,
					Tags:             []string{createdByTag},
				}, gomock.Any()).Return(&instance.CreatePrivateNICResponse{
					PrivateNic: &instance.PrivateNIC{
						ServerID:         serverID,
						PrivateNetworkID: privateNetworkID,
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

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.CreatePrivateNIC(tt.args.ctx, tt.args.zone, tt.args.serverID, tt.args.privateNetworkID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreatePrivateNIC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreatePrivateNIC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetAllServerUserData(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		serverID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]io.Reader
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "get all server user data",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				serverID: serverID,
			},
			want: map[string]io.Reader{
				"test": strings.NewReader("test value"),
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.GetAllServerUserData(&instance.GetAllServerUserDataRequest{
					Zone:     scw.ZoneFrPar1,
					ServerID: serverID,
				}, gomock.Any()).Return(&instance.GetAllServerUserDataResponse{
					UserData: map[string]io.Reader{
						"test": strings.NewReader("test value"),
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

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.GetAllServerUserData(tt.args.ctx, tt.args.zone, tt.args.serverID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetAllServerUserData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetAllServerUserData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SetServerUserData(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		serverID string
		key      string
		content  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "set server user data",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				serverID: serverID,
				key:      "test",
				content:  "value",
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.SetServerUserData(&instance.SetServerUserDataRequest{
					Zone:     scw.ZoneFrPar1,
					ServerID: serverID,
					Key:      "test",
					Content:  strings.NewReader("value"),
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.SetServerUserData(tt.args.ctx, tt.args.zone, tt.args.serverID, tt.args.key, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("Client.SetServerUserData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteServerUserData(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		serverID string
		key      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "delete server user data",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				serverID: serverID,
				key:      "test",
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.DeleteServerUserData(&instance.DeleteServerUserDataRequest{
					Zone:     scw.ZoneFrPar1,
					ServerID: serverID,
					Key:      "test",
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.DeleteServerUserData(tt.args.ctx, tt.args.zone, tt.args.serverID, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteServerUserData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ServerAction(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		serverID string
		action   instance.ServerAction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "poweroff server action",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				serverID: serverID,
				action:   instance.ServerActionPoweroff,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ServerAction(&instance.ServerActionRequest{
					Zone:     scw.ZoneFrPar1,
					ServerID: serverID,
					Action:   instance.ServerActionPoweroff,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.ServerAction(tt.args.ctx, tt.args.zone, tt.args.serverID, tt.args.action); (err != nil) != tt.wantErr {
				t.Errorf("Client.ServerAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DetachVolume(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		volumeID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "detach volume",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				volumeID: volumeID,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.DetachVolume(&instance.DetachVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.DetachVolume(tt.args.ctx, tt.args.zone, tt.args.volumeID); (err != nil) != tt.wantErr {
				t.Errorf("Client.DetachVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_UpdateInstanceVolumeTags(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		volumeID string
		tags     []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "update instance volume tags",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				volumeID: volumeID,
				tags:     []string{"tag1", "tag2", "tag3"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.UpdateVolume(&instance.UpdateVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
					Tags:     &[]string{"tag1", "tag2", "tag3"},
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.UpdateInstanceVolumeTags(tt.args.ctx, tt.args.zone, tt.args.volumeID, tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateInstanceVolumeTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_FindInstanceVolume(t *testing.T) {
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.Volume
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "no volume found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListVolumes(&instance.ListVolumesRequest{
					Zone: scw.ZoneFrPar1,
					Tags: []string{"tag1", "tag2", "tag3"},
				}, gomock.Any(), gomock.Any()).Return(&instance.ListVolumesResponse{}, nil)
			},
		},
		{
			name: "volume found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			want: &instance.Volume{
				ID:   volumeID,
				Name: "volume",
				Tags: []string{"tag1", "tag2", "tag3", "tag4"},
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListVolumes(&instance.ListVolumesRequest{
					Zone: scw.ZoneFrPar1,
					Tags: []string{"tag1", "tag2", "tag3"},
				}, gomock.Any(), gomock.Any()).Return(&instance.ListVolumesResponse{
					TotalCount: 1,
					Volumes: []*instance.Volume{
						{
							ID:   volumeID,
							Name: "volume",
							Tags: []string{"tag1", "tag2", "tag3", "tag4"},
						},
					},
				}, nil)
			},
		},
		{
			name: "duplicate volumes found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListVolumes(&instance.ListVolumesRequest{
					Zone: scw.ZoneFrPar1,
					Tags: []string{"tag1", "tag2", "tag3"},
				}, gomock.Any(), gomock.Any()).Return(&instance.ListVolumesResponse{
					TotalCount: 2,
					Volumes: []*instance.Volume{
						{
							Name: "volume",
							Tags: []string{"tag1", "tag2", "tag3", "tag4"},
						},
						{
							Name: "volume-duplicate",
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

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.FindInstanceVolume(tt.args.ctx, tt.args.zone, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindInstanceVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindInstanceVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteInstanceVolume(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		volumeID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "delete instance volume",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				volumeID: volumeID,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.DeleteVolume(&instance.DeleteVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.DeleteInstanceVolume(tt.args.ctx, tt.args.zone, tt.args.volumeID); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteInstanceVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteServer(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		serverID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "delete server",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				serverID: serverID,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.DeleteServer(&instance.DeleteServerRequest{
					Zone:     scw.ZoneFrPar1,
					ServerID: serverID,
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			if err := c.DeleteServer(tt.args.ctx, tt.args.zone, tt.args.serverID); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_FindPlacementGroup(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.PlacementGroup
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "placement group not found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "test-placement-group",
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListPlacementGroups(&instance.ListPlacementGroupsRequest{
					Zone:    scw.ZoneFrPar1,
					Name:    scw.StringPtr("test-placement-group"),
					Project: scw.StringPtr(projectID),
				}, gomock.Any(), gomock.Any()).Return(&instance.ListPlacementGroupsResponse{}, nil)
			},
		},
		{
			name: "placement group found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "test-placement-group",
			},
			want: &instance.PlacementGroup{
				ID:   placementGroupID,
				Name: "test-placement-group",
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListPlacementGroups(&instance.ListPlacementGroupsRequest{
					Zone:    scw.ZoneFrPar1,
					Name:    scw.StringPtr("test-placement-group"),
					Project: scw.StringPtr(projectID),
				}, gomock.Any(), gomock.Any()).Return(&instance.ListPlacementGroupsResponse{
					TotalCount: 1,
					PlacementGroups: []*instance.PlacementGroup{
						{
							ID:   placementGroupID,
							Name: "test-placement-group",
						},
					},
				}, nil)
			},
		},
		{
			name: "duplicate placement groups found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "test-placement-group",
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListPlacementGroups(&instance.ListPlacementGroupsRequest{
					Zone:    scw.ZoneFrPar1,
					Name:    scw.StringPtr("test-placement-group"),
					Project: scw.StringPtr(projectID),
				}, gomock.Any(), gomock.Any()).Return(&instance.ListPlacementGroupsResponse{
					TotalCount: 2,
					PlacementGroups: []*instance.PlacementGroup{
						{
							Name: "test-placement-group",
						},
						{
							Name: "test-placement-group",
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

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.FindPlacementGroup(tt.args.ctx, tt.args.zone, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindPlacementGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindPlacementGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindSecurityGroup(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.SecurityGroup
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "security group not found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "test-security-group",
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListSecurityGroups(&instance.ListSecurityGroupsRequest{
					Zone:    scw.ZoneFrPar1,
					Name:    scw.StringPtr("test-security-group"),
					Project: scw.StringPtr(projectID),
				}, gomock.Any(), gomock.Any()).Return(&instance.ListSecurityGroupsResponse{}, nil)
			},
		},
		{
			name: "security group found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "test-security-group",
			},
			want: &instance.SecurityGroup{
				ID:   securityGroupID,
				Name: "test-security-group",
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListSecurityGroups(&instance.ListSecurityGroupsRequest{
					Zone:    scw.ZoneFrPar1,
					Name:    scw.StringPtr("test-security-group"),
					Project: scw.StringPtr(projectID),
				}, gomock.Any(), gomock.Any()).Return(&instance.ListSecurityGroupsResponse{
					TotalCount: 1,
					SecurityGroups: []*instance.SecurityGroup{
						{
							ID:   securityGroupID,
							Name: "test-security-group",
						},
					},
				}, nil)
			},
		},
		{
			name: "duplicate security groups found",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "test-security-group",
			},
			wantErr: true,
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.ListSecurityGroups(&instance.ListSecurityGroupsRequest{
					Zone:    scw.ZoneFrPar1,
					Name:    scw.StringPtr("test-security-group"),
					Project: scw.StringPtr(projectID),
				}, gomock.Any(), gomock.Any()).Return(&instance.ListSecurityGroupsResponse{
					TotalCount: 2,
					SecurityGroups: []*instance.SecurityGroup{
						{
							Name: "test-security-group",
						},
						{
							Name: "test-security-group",
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

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.FindSecurityGroup(tt.args.ctx, tt.args.zone, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindSecurityGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindSecurityGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_UpdateServerPublicIPs(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx         context.Context
		zone        scw.Zone
		id          string
		publicIPIDs []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instance.Server
		wantErr bool
		expect  func(d *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "update public IPs",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:         context.TODO(),
				zone:        scw.ZoneFrPar1,
				id:          serverID,
				publicIPIDs: []string{ipID},
			},
			want: &instance.Server{
				ID: serverID,
			},
			expect: func(d *mock_client.MockInstanceAPIMockRecorder) {
				d.UpdateServer(&instance.UpdateServerRequest{
					Zone:      scw.ZoneFrPar1,
					ServerID:  serverID,
					PublicIPs: &[]string{ipID},
				}, gomock.Any()).Return(&instance.UpdateServerResponse{
					Server: &instance.Server{
						ID: serverID,
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

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				instance:  instanceMock,
			}
			got, err := c.UpdateServerPublicIPs(tt.args.ctx, tt.args.zone, tt.args.id, tt.args.publicIPIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateServerPublicIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.UpdateServerPublicIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}
