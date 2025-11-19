package client

import (
	"context"
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const marketplaceImageID = "11111111-1111-1111-1111-111111111111"

func TestClient_GetLocalImageByLabel(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
		region    scw.Region
	}
	type args struct {
		ctx            context.Context
		zone           scw.Zone
		commercialType string
		imageLabel     string
		imageType      marketplace.LocalImageType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *marketplace.LocalImage
		wantErr bool
		expect  func(m *mock_client.MockMarketplaceAPIMockRecorder)
	}{
		{
			name: "get local image by label",
			fields: fields{
				projectID: projectID,
				region:    scw.RegionFrPar,
			},
			args: args{
				ctx:            context.TODO(),
				zone:           scw.ZoneFrPar1,
				commercialType: "DEV1-S",
				imageLabel:     "ubuntu_noble",
				imageType:      marketplace.LocalImageTypeInstanceSbs,
			},
			want: &marketplace.LocalImage{ID: marketplaceImageID},
			expect: func(m *mock_client.MockMarketplaceAPIMockRecorder) {
				m.GetLocalImageByLabel(&marketplace.GetLocalImageByLabelRequest{
					Zone:           scw.ZoneFrPar1,
					ImageLabel:     "ubuntu_noble",
					CommercialType: "DEV1-S",
					Type:           marketplace.LocalImageTypeInstanceSbs,
				}, gomock.Any()).Return(&marketplace.LocalImage{
					ID: marketplaceImageID,
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
			marketplaceMock := mock_client.NewMockMarketplaceAPI(mockCtrl)

			// Every API call must be preceded by a zone check.
			instanceMock.EXPECT().Zones().Return(tt.fields.region.GetZones())

			tt.expect(marketplaceMock.EXPECT())

			c := &Client{
				projectID:   tt.fields.projectID,
				region:      tt.fields.region,
				instance:    instanceMock,
				marketplace: marketplaceMock,
			}
			got, err := c.GetLocalImageByLabel(tt.args.ctx, tt.args.zone, tt.args.commercialType, tt.args.imageLabel, tt.args.imageType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetLocalImageByLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetLocalImageByLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
