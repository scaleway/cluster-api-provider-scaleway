package scope

import (
	"testing"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/util/patch"
)

func TestCluster_ResourceName(t *testing.T) {
	type fields struct {
		patchHelper     *patch.Helper
		ScalewayCluster *infrav1.ScalewayCluster
		ScalewayClient  *scwClient.Client
	}
	type args struct {
		suffixes []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "no suffix provided",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					ObjectMeta: v1.ObjectMeta{
						Name: "cluster-name",
					},
				},
			},
			args: args{},
			want: "cluster-name",
		},
		{
			name: "suffix provided",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					ObjectMeta: v1.ObjectMeta{
						Name: "cluster-name",
					},
				},
			},
			args: args{suffixes: []string{"0"}},
			want: "cluster-name-0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				patchHelper:     tt.fields.patchHelper,
				ScalewayCluster: tt.fields.ScalewayCluster,
				ScalewayClient:  tt.fields.ScalewayClient,
			}
			if got := c.ResourceName(tt.args.suffixes...); got != tt.want {
				t.Errorf("Cluster.ResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
