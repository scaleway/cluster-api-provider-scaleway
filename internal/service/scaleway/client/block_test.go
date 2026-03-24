package client

import (
	"context"
	"reflect"
	"testing"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
)

const volumeID = "22222222-2222-2222-2222-222222222222"

func TestClient_UpdateVolumeIOPS(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx      context.Context
		zone     scw.Zone
		volumeID string
		iops     int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(b *mock_client.MockBlockAPIMockRecorder)
	}{
		{
			name: "unknown zone",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {},
			args: args{
				zone: "fr-par-999",
			},
			wantErr: true,
		},
		{
			name: "update iops to 15000",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.UpdateVolume(&block.UpdateVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
					PerfIops: scw.Uint32Ptr(15000),
				}, gomock.Any())
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				volumeID: volumeID,
				iops:     15000,
			},
		},
		{
			name: "API error",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.UpdateVolume(&block.UpdateVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
					PerfIops: scw.Uint32Ptr(15000),
				}, gomock.Any()).Return(nil, errAPI)
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				volumeID: volumeID,
				iops:     15000,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			blockMock := mock_client.NewMockBlockAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			blockMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(blockMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				block:     blockMock,
			}
			if err := c.UpdateVolumeIOPS(tt.args.ctx, tt.args.zone, tt.args.volumeID, tt.args.iops); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateVolumeIOPS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_UpdateVolumeTags(t *testing.T) {
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
		expect  func(b *mock_client.MockBlockAPIMockRecorder)
	}{
		{
			name: "unknown zone",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {},
			args: args{
				zone: "fr-par-999",
			},
			wantErr: true,
		},
		{
			name: "update tags",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.UpdateVolume(&block.UpdateVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
					Tags:     &[]string{"tag1", "tag2", "tag3"},
				}, gomock.Any())
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				volumeID: volumeID,
				tags:     []string{"tag1", "tag2", "tag3"},
			},
		},
		{
			name: "API error",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.UpdateVolume(&block.UpdateVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
					Tags:     &[]string{"tag1", "tag2", "tag3"},
				}, gomock.Any()).Return(nil, errAPI)
			},
			args: args{
				ctx:      context.TODO(),
				zone:     scw.ZoneFrPar1,
				volumeID: volumeID,
				tags:     []string{"tag1", "tag2", "tag3"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			blockMock := mock_client.NewMockBlockAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			blockMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(blockMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				block:     blockMock,
			}
			if err := c.UpdateVolumeTags(tt.args.ctx, tt.args.zone, tt.args.volumeID, tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateVolumeTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreateVolume(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx  context.Context
		zone scw.Zone
		name string
		size scw.Size
		iops int64
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *block.Volume
		wantErr bool
		expect  func(b *mock_client.MockBlockAPIMockRecorder)
	}{
		{
			name: "unknown zone",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {},
			args: args{
				zone: "fr-par-999",
			},
			wantErr: true,
		},
		{
			name: "create volume without iops",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "my-volume",
				size: 20 * scw.GB,
				iops: 0,
				tags: []string{"tag1", "tag2"},
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.CreateVolume(&block.CreateVolumeRequest{
					Zone: scw.ZoneFrPar1,
					Name: "my-volume",
					FromEmpty: &block.CreateVolumeRequestFromEmpty{
						Size: 20 * scw.GB,
					},
					Tags: []string{"tag1", "tag2", createdByTag},
				}, gomock.Any()).Return(&block.Volume{Name: "my-volume"}, nil)
			},
			want: &block.Volume{Name: "my-volume"},
		},
		{
			name: "create volume with iops",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "my-volume",
				size: 20 * scw.GB,
				iops: 5000,
				tags: []string{"tag1", "tag2"},
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.CreateVolume(&block.CreateVolumeRequest{
					Zone: scw.ZoneFrPar1,
					Name: "my-volume",
					FromEmpty: &block.CreateVolumeRequestFromEmpty{
						Size: 20 * scw.GB,
					},
					Tags:     []string{"tag1", "tag2", createdByTag},
					PerfIops: scw.Uint32Ptr(5000),
				}, gomock.Any()).Return(&block.Volume{Name: "my-volume"}, nil)
			},
			want: &block.Volume{Name: "my-volume"},
		},
		{
			name: "API error",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				name: "my-volume",
				size: 20 * scw.GB,
				tags: []string{"tag1", "tag2"},
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.CreateVolume(&block.CreateVolumeRequest{
					Zone: scw.ZoneFrPar1,
					Name: "my-volume",
					FromEmpty: &block.CreateVolumeRequestFromEmpty{
						Size: 20 * scw.GB,
					},
					Tags: []string{"tag1", "tag2", createdByTag},
				}, gomock.Any()).Return(nil, errAPI)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			blockMock := mock_client.NewMockBlockAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			blockMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(blockMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				block:     blockMock,
			}
			got, err := c.CreateVolume(tt.args.ctx, tt.args.zone, tt.args.name, tt.args.size, tt.args.iops, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteVolume(t *testing.T) {
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
		expect  func(b *mock_client.MockBlockAPIMockRecorder)
	}{
		{
			name: "unknown zone",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {},
			args: args{
				zone: "fr-par-999",
			},
			wantErr: true,
		},
		{
			name: "delete volume",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.DeleteVolume(&block.DeleteVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
				}, gomock.Any())
			},
			args: args{
				zone:     scw.ZoneFrPar1,
				ctx:      context.TODO(),
				volumeID: volumeID,
			},
		},
		{
			name: "API error",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.DeleteVolume(&block.DeleteVolumeRequest{
					Zone:     scw.ZoneFrPar1,
					VolumeID: volumeID,
				}, gomock.Any()).Return(errAPI)
			},
			args: args{
				zone:     scw.ZoneFrPar1,
				ctx:      context.TODO(),
				volumeID: volumeID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			blockMock := mock_client.NewMockBlockAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			blockMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(blockMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				block:     blockMock,
			}
			if err := c.DeleteVolume(tt.args.ctx, tt.args.zone, tt.args.volumeID); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_FindVolumes(t *testing.T) {
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
		want    []*block.Volume
		wantErr bool
		expect  func(b *mock_client.MockBlockAPIMockRecorder)
	}{
		{
			name: "unknown zone",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {},
			args: args{
				zone: "fr-par-999",
			},
			wantErr: true,
		},
		{
			name: "fail with empty tags",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{},
			},
			wantErr: true,
			expect:  func(b *mock_client.MockBlockAPIMockRecorder) {},
		},
		{
			name: "multiple volumes found",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			wantErr: false,
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.ListVolumes(&block.ListVolumesRequest{
					Zone: scw.ZoneFrPar1,
					Tags: []string{"tag1", "tag2", "tag3"},
				}, gomock.Any(), gomock.Any()).Return(&block.ListVolumesResponse{
					Volumes: []*block.Volume{
						{Tags: []string{"tag1", "tag2", "tag3", "tag4"}},
						{Tags: []string{"tag1", "tag2", "tag3", "tag4"}},
					},
					TotalCount: 2,
				}, nil)
			},
			want: []*block.Volume{
				{Tags: []string{"tag1", "tag2", "tag3", "tag4"}},
				{Tags: []string{"tag1", "tag2", "tag3", "tag4"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			blockMock := mock_client.NewMockBlockAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			blockMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(blockMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				region:    tt.fields.region,
				block:     blockMock,
			}
			got, err := c.FindVolumes(tt.args.ctx, tt.args.zone, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindVolumes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindVolumes() = %v, want %v", got, tt.want)
			}
		})
	}
}
