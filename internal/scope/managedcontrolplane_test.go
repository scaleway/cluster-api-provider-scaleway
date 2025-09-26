package scope

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

func Test_generateScalewayK8sName(t *testing.T) {
	type args struct {
		resourceName string
		namespace    string
		maxLength    int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "escaped name",
			args: args{
				resourceName: "test.cluster",
				namespace:    "default",
				maxLength:    maxClusterNameLength,
			},
			want: "default-test-cluster",
		},
		{
			name: "hashed name",
			args: args{
				resourceName: "test-cluster-test-cluster-test-cluster-test-cluster-test-cluster-test-cluster-test-cluster-test-cluster",
				namespace:    "default",
				maxLength:    maxClusterNameLength,
			},
			want: "caps-hma5vzr1gzj7q6045d1dei138m14jwosabmbtcabyqt6kr33qfj9hs2nj3u",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateScalewayK8sName(tt.args.resourceName, tt.args.namespace, tt.args.maxLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateScalewayK8sName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateScalewayK8sName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManagedControlPlane_ClusterName(t *testing.T) {
	type fields struct {
		ManagedCluster      *infrav1.ScalewayManagedCluster
		ManagedControlPlane *infrav1.ScalewayManagedControlPlane
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "name already present",
			fields: fields{
				ManagedCluster: &infrav1.ScalewayManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster",
						Namespace: "default",
					},
				},
				ManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster",
						Namespace: "default",
					},
					Spec: infrav1.ScalewayManagedControlPlaneSpec{
						ClusterName: "mycluster",
					},
				},
			},
			want: "mycluster",
		},
		{
			name: "generate name",
			fields: fields{
				ManagedCluster: &infrav1.ScalewayManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster",
						Namespace: "default",
					},
				},
				ManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster",
						Namespace: "default",
					},
				},
			},
			want: "default-cluster",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ManagedControlPlane{
				ScalewayManagedCluster:      tt.fields.ManagedCluster,
				ScalewayManagedControlPlane: tt.fields.ManagedControlPlane,
			}
			if got := m.ClusterName(); got != tt.want {
				t.Errorf("ManagedControlPlane.ClusterName() = %v, want %v", got, tt.want)
			}
			if tt.want != m.ScalewayManagedControlPlane.Spec.ClusterName {
				t.Errorf("expected ManagedControlPlane.Spec.ClusterName to equal %v, got %v", tt.want, m.ScalewayManagedControlPlane.Spec.ClusterName)
			}
		})
	}
}
