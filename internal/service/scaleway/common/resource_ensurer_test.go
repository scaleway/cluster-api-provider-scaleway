package common

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/common/mock_common"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
)

type resource int
type desired int

func TestResourceEnsurer_Do(t *testing.T) {
	t.Parallel()
	type fields struct {
	}
	type args struct {
		ctx     context.Context
		desired []desired
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []resource
		wantErr bool
		expect  func(r *mock_common.MockResourceReconcilerMockRecorder[desired, resource])
	}{
		{
			name: "ensure new resources",
			args: args{
				ctx:     context.TODO(),
				desired: []desired{1, 2, 3},
			},
			wantErr: false,
			want:    []resource{1, 2, 3},
			expect: func(r *mock_common.MockResourceReconcilerMockRecorder[desired, resource]) {
				r.GetDesiredZone(desired(1)).Return(scw.ZoneFrPar1, nil)
				r.GetDesiredZone(desired(2)).Return(scw.ZoneFrPar2, nil)
				r.GetDesiredZone(desired(3)).Return(scw.ZoneFrPar3, nil)
				r.ListResources(gomock.Any())
				r.GetDesiredResourceName(0).Return("resource-0").Times(3)

				// Resources are created in a random order.
				r.CreateResource(gomock.Any(), scw.ZoneFrPar1, "resource-0", desired(1)).Return(resource(1), nil)
				r.CreateResource(gomock.Any(), scw.ZoneFrPar2, "resource-0", desired(2)).Return(resource(2), nil)
				r.CreateResource(gomock.Any(), scw.ZoneFrPar3, "resource-0", desired(3)).Return(resource(3), nil)
			},
		},
		{
			name: "ensure existing resources",
			args: args{
				ctx:     context.TODO(),
				desired: []desired{1, 2, 3},
			},
			wantErr: false,
			want:    []resource{1, 2, 3},
			expect: func(r *mock_common.MockResourceReconcilerMockRecorder[desired, resource]) {
				r.GetDesiredZone(desired(1)).Return(scw.ZoneFrPar1, nil)
				r.GetDesiredZone(desired(2)).Return(scw.ZoneFrPar2, nil)
				r.GetDesiredZone(desired(3)).Return(scw.ZoneFrPar3, nil)

				r.ListResources(gomock.Any()).Return([]resource{1, 2, 3}, nil)

				r.GetDesiredResourceName(0).Return("resource-0").Times(2 * 3)

				r.GetResourceZone(resource(1)).Return(scw.ZoneFrPar1).AnyTimes()
				r.GetResourceName(resource(1)).Return("resource-0").AnyTimes()
				r.ShouldKeepResource(gomock.Any(), resource(1), desired(1)).Return(true, nil)
				r.UpdateResource(gomock.Any(), resource(1), desired(1)).Return(resource(1), nil)

				r.GetResourceZone(resource(2)).Return(scw.ZoneFrPar2).AnyTimes()
				r.GetResourceName(resource(2)).Return("resource-0").AnyTimes()
				r.ShouldKeepResource(gomock.Any(), resource(2), desired(2)).Return(true, nil)
				r.UpdateResource(gomock.Any(), resource(2), desired(2)).Return(resource(2), nil)

				r.GetResourceZone(resource(3)).Return(scw.ZoneFrPar3).AnyTimes()
				r.GetResourceName(resource(3)).Return("resource-0").AnyTimes()
				r.ShouldKeepResource(gomock.Any(), resource(3), desired(3)).Return(true, nil)
				r.UpdateResource(gomock.Any(), resource(3), desired(3)).Return(resource(3), nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			reconcilerMock := mock_common.NewMockResourceReconciler[desired, resource](mockCtrl)

			tt.expect(reconcilerMock.EXPECT())

			r := &ResourceEnsurer[desired, resource]{
				ResourceReconciler: reconcilerMock,
			}

			got, err := r.Do(tt.args.ctx, tt.args.desired)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResourceEnsurer.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			slices.Sort(got) // values are returned out-of-order
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceEnsurer.Do() = %v, want %v", got, tt.want)
			}
		})
	}
}
