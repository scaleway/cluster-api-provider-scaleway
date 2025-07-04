package scope

import (
	"fmt"
	"reflect"
	"testing"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/scaleway-sdk-go/scw"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/util/patch"
)

const (
	privateNetworkID = "11111111-1111-1111-1111-111111111111"
	vpcID            = "11111111-1111-1111-1111-111111111111"
	lbIP             = "42.42.42.42"
)

func TestCluster_ResourceName(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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

func TestCluster_ResourceTags(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	type args struct {
		additional []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "base tags",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					ObjectMeta: v1.ObjectMeta{
						Name:      "my-cluster",
						Namespace: "default",
					},
				},
			},
			args: args{},
			want: []string{"caps-namespace=default", "caps-scalewaycluster=my-cluster"},
		},
		{
			name: "with additional tag",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					ObjectMeta: v1.ObjectMeta{
						Name:      "my-cluster",
						Namespace: "default",
					},
				},
			},
			args: args{
				additional: []string{"additional-tag"},
			},
			want: []string{"caps-namespace=default", "caps-scalewaycluster=my-cluster", "additional-tag"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			if got := c.ResourceTags(tt.args.additional...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cluster.ResourceTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_HasPrivateNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			want: false,
		},
		{
			name: "enabled",
			fields: fields{
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
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			if got := c.HasPrivateNetwork(); got != tt.want {
				t.Errorf("Cluster.HasPrivateNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_ShouldManagePrivateNetwork(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			want: false,
		},
		{
			name: "existing private network",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							PrivateNetwork: &infrav1.PrivateNetworkSpec{
								Enabled: true,
								ID:      scw.StringPtr(privateNetworkID),
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "managed",
			fields: fields{
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
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			if got := c.ShouldManagePrivateNetwork(); got != tt.want {
				t.Errorf("Cluster.ShouldManagePrivateNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_PrivateNetworkID(t *testing.T) {
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			wantErr: true, // cluster has no Private Network
		},
		{
			name: "missing status",
			fields: fields{
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
			wantErr: true, // PrivateNetworkID not found in ScalewayCluster status
		},
		{
			name: "found private network ID",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							PrivateNetwork: &infrav1.PrivateNetworkSpec{
								Enabled: true,
							},
						},
					},
					Status: infrav1.ScalewayClusterStatus{
						Network: &infrav1.NetworkStatus{
							PrivateNetworkID: scw.StringPtr(privateNetworkID),
						},
					},
				},
			},
			want: privateNetworkID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			got, err := c.PrivateNetworkID()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cluster.PrivateNetworkID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Cluster.PrivateNetworkID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_ControlPlaneLoadBalancerPort(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name   string
		fields fields
		want   int32
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			want: defaultFrontendControlPlanePort,
		},
		{
			name: "override with 443",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							ControlPlaneLoadBalancer: &infrav1.ControlPlaneLoadBalancerSpec{
								Port: scw.Int32Ptr(443),
							},
						},
					},
				},
			},
			want: 443,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			if got := c.ControlPlaneLoadBalancerPort(); got != tt.want {
				t.Errorf("Cluster.ControlPlaneLoadBalancerPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_ControlPlaneLoadBalancerAllowedRanges(t *testing.T) {
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			want: nil,
		},
		{
			name: "allowed ranges set",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							ControlPlaneLoadBalancer: &infrav1.ControlPlaneLoadBalancerSpec{
								AllowedRanges: []infrav1.CIDR{"127.0.0.1/32", "10.0.0.0/8"},
							},
						},
					},
				},
			},
			want: []string{"127.0.0.1/32", "10.0.0.0/8"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			if got := c.ControlPlaneLoadBalancerAllowedRanges(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cluster.ControlPlaneLoadBalancerAllowedRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_HasControlPlaneDNS(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			want: false,
		},
		{
			name: "public dns",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							ControlPlaneDNS: &infrav1.ControlPlaneDNSSpec{
								Domain: "example.com",
								Name:   "domain",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "private dns",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							PrivateNetwork: &infrav1.PrivateNetworkSpec{
								Enabled: true,
							},
							ControlPlaneLoadBalancer: &infrav1.ControlPlaneLoadBalancerSpec{
								Private: scw.BoolPtr(true),
							},
							ControlPlanePrivateDNS: &infrav1.ControlPlanePrivateDNSSpec{
								Name: "domain",
							},
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
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			if got := c.HasControlPlaneDNS(); got != tt.want {
				t.Errorf("Cluster.HasControlPlaneDNS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_ControlPlaneDNSZoneAndName(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			wantErr: true,
		},
		{
			name: "public DNS",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							ControlPlaneDNS: &infrav1.ControlPlaneDNSSpec{
								Domain: "example.com",
								Name:   "domain",
							},
						},
					},
				},
			},
			want:  "example.com",
			want1: "domain",
		},
		{
			name: "private DNS",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							PrivateNetwork: &infrav1.PrivateNetworkSpec{
								Enabled: true,
							},
							ControlPlaneLoadBalancer: &infrav1.ControlPlaneLoadBalancerSpec{
								Private: scw.BoolPtr(true),
							},
							ControlPlanePrivateDNS: &infrav1.ControlPlanePrivateDNSSpec{
								Name: "domain",
							},
						},
					},
					Status: infrav1.ScalewayClusterStatus{
						Network: &infrav1.NetworkStatus{
							VPCID:            scw.StringPtr(vpcID),
							PrivateNetworkID: scw.StringPtr(privateNetworkID),
						},
					},
				},
			},
			want:  fmt.Sprintf("%s.%s.privatedns", vpcID, privateNetworkID),
			want1: "domain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			got, got1, err := c.ControlPlaneDNSZoneAndName()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cluster.ControlPlaneDNSZoneAndName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Cluster.ControlPlaneDNSZoneAndName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Cluster.ControlPlaneDNSZoneAndName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCluster_ControlPlaneHost(t *testing.T) {
	t.Parallel()
	type fields struct {
		ScalewayCluster *infrav1.ScalewayCluster
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "empty spec",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{},
			},
			wantErr: true,
		},
		{
			name: "public DNS",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							ControlPlaneDNS: &infrav1.ControlPlaneDNSSpec{
								Domain: "example.com",
								Name:   "domain",
							},
						},
					},
				},
			},
			want: "domain.example.com",
		},
		{
			name: "private DNS",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Spec: infrav1.ScalewayClusterSpec{
						Network: &infrav1.NetworkSpec{
							PrivateNetwork: &infrav1.PrivateNetworkSpec{
								Enabled: true,
							},
							ControlPlaneLoadBalancer: &infrav1.ControlPlaneLoadBalancerSpec{
								Private: scw.BoolPtr(true),
							},
							ControlPlanePrivateDNS: &infrav1.ControlPlanePrivateDNSSpec{
								Name: "domain",
							},
						},
					},
					Status: infrav1.ScalewayClusterStatus{
						Network: &infrav1.NetworkStatus{
							VPCID:            scw.StringPtr(vpcID),
							PrivateNetworkID: scw.StringPtr(privateNetworkID),
						},
					},
				},
			},
			want: fmt.Sprintf("domain.%s.internal", privateNetworkID),
		},
		{
			name: "ip",
			fields: fields{
				ScalewayCluster: &infrav1.ScalewayCluster{
					Status: infrav1.ScalewayClusterStatus{
						Network: &infrav1.NetworkStatus{
							LoadBalancerIP: scw.StringPtr(lbIP),
						},
					},
				},
			},
			want: lbIP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Cluster{
				ScalewayCluster: tt.fields.ScalewayCluster,
			}
			got, err := c.ControlPlaneHost()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cluster.ControlPlaneHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Cluster.ControlPlaneHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
