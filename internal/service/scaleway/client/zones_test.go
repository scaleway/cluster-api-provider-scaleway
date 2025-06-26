package client

import (
	"reflect"
	"testing"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
)

func TestClient_GetZoneOrDefault(t *testing.T) {
	t.Parallel()
	type fields struct {
		region scw.Region
	}
	type args struct {
		zone *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    scw.Zone
		wantErr bool
	}{
		{
			name: "return default zone",
			fields: fields{
				region: scw.RegionFrPar,
			},
			args: args{},
			want: scw.ZoneFrPar1,
		},
		{
			name: "invalid zone",
			fields: fields{
				region: scw.RegionFrPar,
			},
			args: args{
				zone: scw.StringPtr("invalid-zone"),
			},
			wantErr: true,
		},
		{
			name: "provided zone",
			fields: fields{
				region: scw.RegionFrPar,
			},
			args: args{
				zone: scw.StringPtr("fr-par-3"),
			},
			want: scw.ZoneFrPar3,
		},
		{
			// GetZoneOrDefault does not enforce that zone must be in the same
			// region as the client's region. This should be validated using validateZone.
			name: "zone in another region",
			fields: fields{
				region: scw.RegionFrPar,
			},
			args: args{
				zone: scw.StringPtr("nl-ams-1"),
			},
			want: scw.ZoneNlAms1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{
				region: tt.fields.region,
			}
			got, err := c.GetZoneOrDefault(tt.args.zone)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetZoneOrDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetZoneOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetControlPlaneZones(t *testing.T) {
	t.Parallel()
	type fields struct {
		region scw.Region
	}
	tests := []struct {
		name   string
		fields fields
		want   []scw.Zone
		expect func(i *mock_client.MockInstanceAPIMockRecorder)
	}{
		{
			name: "fr-par",
			fields: fields{
				region: scw.RegionFrPar,
			},
			want: []scw.Zone{scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneFrPar3},
			expect: func(i *mock_client.MockInstanceAPIMockRecorder) {
				i.Zones().Return([]scw.Zone{
					scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneFrPar3,
					scw.ZoneNlAms1, scw.ZoneNlAms2, scw.ZoneNlAms3,
					scw.ZonePlWaw1, scw.ZonePlWaw2, scw.ZonePlWaw3,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			instanceMock := mock_client.NewMockInstanceAPI(mockCtrl)

			tt.expect(instanceMock.EXPECT())

			c := &Client{
				region:   tt.fields.region,
				instance: instanceMock,
			}
			if got := c.GetControlPlaneZones(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetControlPlaneZones() = %v, want %v", got, tt.want)
			}
		})
	}
}
