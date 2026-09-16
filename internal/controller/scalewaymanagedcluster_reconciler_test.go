package controller

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

func TestScalewayManagedClusterService_setFailureDomains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		clusterType string
		expect      func(i *mock_client.MockInterfaceMockRecorder)
		want        []clusterv1.FailureDomain
	}{
		{
			name:        "kapsule cluster uses the zones of its region",
			clusterType: "kapsule",
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.GetZones().Return([]scw.Zone{scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneFrPar3})
			},
			want: []clusterv1.FailureDomain{
				{Name: "fr-par-1"},
				{Name: "fr-par-2"},
				{Name: "fr-par-3"},
			},
		},
		{
			name:        "multicloud cluster uses the zones of all regions",
			clusterType: "multicloud",
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.GetAllZones().Return([]scw.Zone{scw.ZoneFrPar1, scw.ZoneNlAms1, scw.ZonePlWaw1})
			},
			want: []clusterv1.FailureDomain{
				{Name: "fr-par-1"},
				{Name: "nl-ams-1"},
				{Name: "pl-waw-1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			scwMock := mock_client.NewMockInterface(mockCtrl)
			tt.expect(scwMock.EXPECT())

			s := &scalewayManagedClusterService{
				scope: &scope.ManagedCluster{
					ScalewayManagedCluster: &infrav1.ScalewayManagedCluster{},
					ScalewayManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
						Spec: infrav1.ScalewayManagedControlPlaneSpec{
							Type: tt.clusterType,
						},
					},
					ScalewayClient: scwMock,
				},
			}

			s.setFailureDomains()

			g.Expect(s.scope.ScalewayManagedCluster.Status.FailureDomains).To(Equal(tt.want))
		})
	}
}
