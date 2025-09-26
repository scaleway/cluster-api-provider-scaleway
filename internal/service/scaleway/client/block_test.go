package client

import (
	"context"
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
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

func TestClient_FindVolume(t *testing.T) {
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
			name: "no volume found",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			wantErr: true,
			expect: func(b *mock_client.MockBlockAPIMockRecorder) {
				b.ListVolumes(&block.ListVolumesRequest{
					Zone: scw.ZoneFrPar1,
					Tags: []string{"tag1", "tag2", "tag3"},
				}, gomock.Any(), gomock.Any()).Return(&block.ListVolumesResponse{}, nil)
			},
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
			name: "too many volumes found",
			fields: fields{
				region:    scw.RegionFrPar,
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				zone: scw.ZoneFrPar1,
				tags: []string{"tag1", "tag2", "tag3"},
			},
			wantErr: true,
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
		},
		{
			name: "one volume found",
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
						{
							Name: "my-volume",
							Tags: []string{"tag1", "tag2", "tag3", "tag4"},
						},
					},
					TotalCount: 1,
				}, nil)
			},
			want: &block.Volume{
				Name: "my-volume",
				Tags: []string{"tag1", "tag2", "tag3", "tag4"},
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
			got, err := c.FindVolume(tt.args.ctx, tt.args.zone, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
