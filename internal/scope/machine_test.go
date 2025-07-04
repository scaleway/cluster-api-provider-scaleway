package scope

import (
	"reflect"
	"testing"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func TestMachine_ResourceTags(t *testing.T) {
	t.Parallel()
	type fields struct {
		Cluster         *Cluster
		ScalewayMachine *infrav1.ScalewayMachine
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "machine tags",
			fields: fields{
				Cluster: &Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
					},
				},
				ScalewayMachine: &infrav1.ScalewayMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "machine",
						Namespace: "default",
					},
				},
			},
			want: []string{"caps-namespace=default", "caps-scalewaycluster=cluster", "caps-scalewaymachine=machine"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Machine{
				Cluster:         tt.fields.Cluster,
				ScalewayMachine: tt.fields.ScalewayMachine,
			}
			if got := m.ResourceTags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Machine.ResourceTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_RootVolumeSize(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayMachine *infrav1.ScalewayMachine
	}
	tests := []struct {
		name   string
		fields fields
		want   scw.Size
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{},
			},
			want: defaultRootVolumeSize,
		},
		{
			name: "50GB",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{
					Spec: infrav1.ScalewayMachineSpec{
						RootVolume: &infrav1.RootVolumeSpec{
							Size: scw.Int64Ptr(50),
						},
					},
				},
			},
			want: 50 * scw.GB,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Machine{
				ScalewayMachine: tt.fields.ScalewayMachine,
			}
			if got := m.RootVolumeSize(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Machine.RootVolumeSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_RootVolumeType(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayMachine *infrav1.ScalewayMachine
	}
	tests := []struct {
		name    string
		fields  fields
		want    instance.VolumeVolumeType
		wantErr bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{},
			},
			want: defaultRootVolumeType,
		},
		{
			name: "local",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{
					Spec: infrav1.ScalewayMachineSpec{
						RootVolume: &infrav1.RootVolumeSpec{
							Type: scw.StringPtr("local"),
						},
					},
				},
			},
			want: instance.VolumeVolumeTypeLSSD,
		},
		{
			name: "block",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{
					Spec: infrav1.ScalewayMachineSpec{
						RootVolume: &infrav1.RootVolumeSpec{
							Type: scw.StringPtr("block"),
						},
					},
				},
			},
			want: instance.VolumeVolumeTypeSbsVolume,
		},
		{
			name: "unknown",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{
					Spec: infrav1.ScalewayMachineSpec{
						RootVolume: &infrav1.RootVolumeSpec{
							Type: scw.StringPtr("unknown"),
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Machine{
				ScalewayMachine: tt.fields.ScalewayMachine,
			}
			got, err := m.RootVolumeType()
			if (err != nil) != tt.wantErr {
				t.Errorf("Machine.RootVolumeType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Machine.RootVolumeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_RootVolumeIOPS(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayMachine *infrav1.ScalewayMachine
	}
	tests := []struct {
		name   string
		fields fields
		want   *int64
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{},
			},
			want: nil,
		},
		{
			name: "15000",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{
					Spec: infrav1.ScalewayMachineSpec{
						RootVolume: &infrav1.RootVolumeSpec{
							IOPS: scw.Int64Ptr(15000),
						},
					},
				},
			},
			want: scw.Int64Ptr(15000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Machine{
				ScalewayMachine: tt.fields.ScalewayMachine,
			}
			if got := m.RootVolumeIOPS(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Machine.RootVolumeIOPS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_HasPublicIPv4(t *testing.T) {
	t.Parallel()
	type fields struct {
		Cluster         *Cluster
		ScalewayMachine *infrav1.ScalewayMachine
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "public cluster defaults to true",
			fields: fields{
				Cluster: &Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{},
				},
			},
			want: true,
		},
		{
			name: "private cluster defaults to false",
			fields: fields{
				Cluster: &Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
							},
						},
					},
				},
				ScalewayMachine: &infrav1.ScalewayMachine{},
			},
			want: false,
		},
		{
			name: "private cluster with ipv4 enabled",
			fields: fields{
				Cluster: &Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
							},
						},
					},
				},
				ScalewayMachine: &infrav1.ScalewayMachine{
					Spec: infrav1.ScalewayMachineSpec{
						PublicNetwork: &infrav1.PublicNetworkSpec{
							EnableIPv4: scw.BoolPtr(true),
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Machine{
				Cluster:         tt.fields.Cluster,
				ScalewayMachine: tt.fields.ScalewayMachine,
			}
			if got := m.HasPublicIPv4(); got != tt.want {
				t.Errorf("Machine.HasPublicIPv4() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_HasPublicIPv6(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayMachine *infrav1.ScalewayMachine
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{},
			},
			want: false,
		},
		{
			name: "ipv6 enabled",
			fields: fields{
				ScalewayMachine: &infrav1.ScalewayMachine{
					Spec: infrav1.ScalewayMachineSpec{
						PublicNetwork: &infrav1.PublicNetworkSpec{
							EnableIPv6: scw.BoolPtr(true),
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Machine{
				ScalewayMachine: tt.fields.ScalewayMachine,
			}
			if got := m.HasPublicIPv6(); got != tt.want {
				t.Errorf("Machine.HasPublicIPv6() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_HasJoinedCluster(t *testing.T) {
	t.Parallel()
	type fields struct {
		Machine *clusterv1.Machine
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "empty status",
			fields: fields{
				Machine: &clusterv1.Machine{},
			},
			want: false,
		},
		{
			name: "nodeRef is set",
			fields: fields{
				Machine: &clusterv1.Machine{
					Status: clusterv1.MachineStatus{
						NodeRef: &corev1.ObjectReference{
							Name: "node-name",
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Machine{
				Machine: tt.fields.Machine,
			}
			if got := m.HasJoinedCluster(); got != tt.want {
				t.Errorf("Machine.HasJoinedCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}
