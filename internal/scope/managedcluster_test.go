package scope

import (
	"testing"

	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

func TestManagedCluster_HasPrivateNetwork(t *testing.T) {
	type fields struct {
		patchHelper         *patch.Helper
		ManagedCluster      *infrav1.ScalewayManagedCluster
		ManagedControlPlane *infrav1.ScalewayManagedControlPlane
		ScalewayClient      scwClient.Interface
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "multicloud has no private network",
			fields: fields{
				ManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
					Spec: infrav1.ScalewayManagedControlPlaneSpec{
						Type: "multicloud",
					},
				},
			},
			want: false,
		},
		{
			name: "multicloud-dedicated-4 has no private network",
			fields: fields{
				ManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
					Spec: infrav1.ScalewayManagedControlPlaneSpec{
						Type: "multicloud-dedicated-4",
					},
				},
			},
			want: false,
		},
		{
			name: "kapsule has private network",
			fields: fields{
				ManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
					Spec: infrav1.ScalewayManagedControlPlaneSpec{
						Type: "kapsule",
					},
				},
			},
			want: true,
		},
		{
			name: "assume a private network if ManagedControlPlane is nil",
			fields: fields{
				ManagedControlPlane: nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ManagedCluster{
				patchHelper:                 tt.fields.patchHelper,
				ScalewayManagedCluster:      tt.fields.ManagedCluster,
				ScalewayManagedControlPlane: tt.fields.ManagedControlPlane,
				ScalewayClient:              tt.fields.ScalewayClient,
			}
			if got := c.HasPrivateNetwork(); got != tt.want {
				t.Errorf("ManagedCluster.HasPrivateNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}
