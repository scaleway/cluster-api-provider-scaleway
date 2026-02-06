package v1alpha1

import (
	"reflect"
	"testing"

	"k8s.io/utils/ptr"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

func Test_ptrIfNotZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		v    string
		want *string
	}{
		{
			name: "empty value",
			v:    "",
			want: nil,
		},
		{
			name: "value not empty",
			v:    "test",
			want: ptr.To("test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ptrIfNotZero(tt.v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ptrIfNotZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ScalewayClusters.
var (
	v1alpha1PrivateScalewayCluster = &ScalewayCluster{
		Spec: ScalewayClusterSpec{
			ProjectID:          "11111111-1111-1111-1111-111111111111",
			Region:             "fr-par",
			ScalewaySecretName: "scaleway-secret",
			FailureDomains:     []string{"fr-par-1", "fr-par-2", "fr-par-3"},
			ControlPlaneEndpoint: clusterv1beta1.APIEndpoint{
				Host: "testing-cluster.33333333-3333-3333-3333-333333333333.internal",
				Port: 6443,
			},
			Network: &NetworkSpec{
				ControlPlaneLoadBalancer: &ControlPlaneLoadBalancerSpec{
					LoadBalancerSpec: LoadBalancerSpec{
						Zone:      ptr.To("fr-par-2"),
						Type:      ptr.To("LB-S"),
						IP:        ptr.To("11.11.11.11"),
						PrivateIP: ptr.To("10.0.0.1"),
					},
					AllowedRanges: []CIDR{
						CIDR("10.0.0.0/8"),
						CIDR("0.0.0.0/0"),
					},
					Private: ptr.To(true),
				},
				ControlPlaneExtraLoadBalancers: []LoadBalancerSpec{
					{
						Zone:      ptr.To("fr-par-1"),
						Type:      ptr.To("LB-M"),
						IP:        ptr.To("22.22.22.22"),
						PrivateIP: ptr.To("10.0.0.2"),
					},
					{
						Zone:      ptr.To("fr-par-3"),
						Type:      ptr.To("LB-L"),
						IP:        ptr.To("33.33.33.33"),
						PrivateIP: ptr.To("10.0.0.3"),
					},
				},
				ControlPlanePrivateDNS: &ControlPlanePrivateDNSSpec{
					Name: "testing-cluster",
				},
				PrivateNetwork: &PrivateNetworkSpec{
					PrivateNetworkParams: PrivateNetworkParams{
						ID:     ptr.To("22222222-2222-2222-2222-222222222222"),
						VPCID:  ptr.To("33333333-3333-3333-3333-333333333333"),
						Subnet: ptr.To("192.168.0.0/16"),
					},
					Enabled: true,
				},
				PublicGateways: []PublicGatewaySpec{
					{
						Type: ptr.To("PGW-M"),
						IP:   ptr.To("44.44.44.44"),
						Zone: ptr.To("fr-par-1"),
					},
					{
						Type: ptr.To("PGW-S"),
						IP:   ptr.To("55.55.55.55"),
						Zone: ptr.To("fr-par-2"),
					},
				},
			},
		},
		Status: ScalewayClusterStatus{
			Ready: true,
		},
	}
	v1alpha2PrivateScalewayCluster = &infrav1.ScalewayCluster{
		Spec: infrav1.ScalewayClusterSpec{
			ProjectID:          infrav1.UUID("11111111-1111-1111-1111-111111111111"),
			Region:             infrav1.ScalewayRegion("fr-par"),
			ScalewaySecretName: "scaleway-secret",
			FailureDomains:     []infrav1.ScalewayZone{"fr-par-1", "fr-par-2", "fr-par-3"},
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "testing-cluster.33333333-3333-3333-3333-333333333333.internal",
				Port: 6443,
			},
			Network: infrav1.ScalewayClusterNetwork{
				ControlPlaneLoadBalancer: infrav1.ControlPlaneLoadBalancer{
					LoadBalancer: infrav1.LoadBalancer{
						Zone:      infrav1.ScalewayZone("fr-par-2"),
						Type:      "LB-S",
						IP:        infrav1.IPv4("11.11.11.11"),
						PrivateIP: infrav1.IPv4("10.0.0.1"),
					},
					AllowedRanges: []infrav1.CIDR{
						infrav1.CIDR("10.0.0.0/8"),
						infrav1.CIDR("0.0.0.0/0"),
					},
					Private: ptr.To(true),
				},
				ControlPlaneExtraLoadBalancers: []infrav1.LoadBalancer{
					{
						Zone:      infrav1.ScalewayZone("fr-par-1"),
						Type:      "LB-M",
						IP:        infrav1.IPv4("22.22.22.22"),
						PrivateIP: infrav1.IPv4("10.0.0.2"),
					},
					{
						Zone:      infrav1.ScalewayZone("fr-par-3"),
						Type:      "LB-L",
						IP:        infrav1.IPv4("33.33.33.33"),
						PrivateIP: infrav1.IPv4("10.0.0.3"),
					},
				},
				ControlPlaneDNS: infrav1.ControlPlaneDNS{
					Name: "testing-cluster",
				},
				PrivateNetwork: infrav1.PrivateNetworkSpec{
					PrivateNetwork: infrav1.PrivateNetwork{
						ID:     infrav1.UUID("22222222-2222-2222-2222-222222222222"),
						VPCID:  infrav1.UUID("33333333-3333-3333-3333-333333333333"),
						Subnet: infrav1.CIDR("192.168.0.0/16"),
					},
					Enabled: ptr.To(true),
				},
				PublicGateways: []infrav1.PublicGateway{
					{
						Type: "PGW-M",
						IP:   infrav1.IPv4("44.44.44.44"),
						Zone: infrav1.ScalewayZone("fr-par-1"),
					},
					{
						Type: "PGW-S",
						IP:   infrav1.IPv4("55.55.55.55"),
						Zone: "fr-par-2",
					},
				},
			},
		},
		Status: infrav1.ScalewayClusterStatus{
			Initialization: infrav1.ScalewayClusterInitializationStatus{
				Provisioned: ptr.To(true),
			},
		},
	}

	v1alpha1PublicScalewayCluster = &ScalewayCluster{
		Spec: ScalewayClusterSpec{
			ProjectID:          "11111111-1111-1111-1111-111111111111",
			Region:             "fr-par",
			ScalewaySecretName: "scaleway-secret",
			FailureDomains:     []string{"fr-par-1", "fr-par-2", "fr-par-3"},
			ControlPlaneEndpoint: clusterv1beta1.APIEndpoint{
				Host: "testing-cluster.example.com",
				Port: 6443,
			},
			Network: &NetworkSpec{
				ControlPlaneLoadBalancer: &ControlPlaneLoadBalancerSpec{
					LoadBalancerSpec: LoadBalancerSpec{
						Zone:      ptr.To("fr-par-2"),
						Type:      ptr.To("LB-S"),
						IP:        ptr.To("11.11.11.11"),
						PrivateIP: ptr.To("10.0.0.1"),
					},
					AllowedRanges: []CIDR{
						CIDR("10.0.0.0/8"),
						CIDR("0.0.0.0/0"),
					},
				},
				ControlPlaneExtraLoadBalancers: []LoadBalancerSpec{
					{
						Zone:      ptr.To("fr-par-1"),
						Type:      ptr.To("LB-M"),
						IP:        ptr.To("22.22.22.22"),
						PrivateIP: ptr.To("10.0.0.2"),
					},
					{
						Zone:      ptr.To("fr-par-3"),
						Type:      ptr.To("LB-L"),
						IP:        ptr.To("33.33.33.33"),
						PrivateIP: ptr.To("10.0.0.3"),
					},
				},
				ControlPlaneDNS: &ControlPlaneDNSSpec{
					Name:   "testing-cluster",
					Domain: "example.com",
				},
				PrivateNetwork: &PrivateNetworkSpec{
					PrivateNetworkParams: PrivateNetworkParams{
						ID:     ptr.To("22222222-2222-2222-2222-222222222222"),
						VPCID:  ptr.To("33333333-3333-3333-3333-333333333333"),
						Subnet: ptr.To("192.168.0.0/16"),
					},
					Enabled: true,
				},
				PublicGateways: []PublicGatewaySpec{
					{
						Type: ptr.To("PGW-M"),
						IP:   ptr.To("44.44.44.44"),
						Zone: ptr.To("fr-par-1"),
					},
					{
						Type: ptr.To("PGW-S"),
						IP:   ptr.To("55.55.55.55"),
						Zone: ptr.To("fr-par-2"),
					},
				},
			},
		},
		Status: ScalewayClusterStatus{
			Ready: true,
		},
	}
	v1alpha2PublicScalewayCluster = &infrav1.ScalewayCluster{
		Spec: infrav1.ScalewayClusterSpec{
			ProjectID:          infrav1.UUID("11111111-1111-1111-1111-111111111111"),
			Region:             infrav1.ScalewayRegion("fr-par"),
			ScalewaySecretName: "scaleway-secret",
			FailureDomains:     []infrav1.ScalewayZone{"fr-par-1", "fr-par-2", "fr-par-3"},
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "testing-cluster.example.com",
				Port: 6443,
			},
			Network: infrav1.ScalewayClusterNetwork{
				ControlPlaneLoadBalancer: infrav1.ControlPlaneLoadBalancer{
					LoadBalancer: infrav1.LoadBalancer{
						Zone:      infrav1.ScalewayZone("fr-par-2"),
						Type:      "LB-S",
						IP:        infrav1.IPv4("11.11.11.11"),
						PrivateIP: infrav1.IPv4("10.0.0.1"),
					},
					AllowedRanges: []infrav1.CIDR{
						infrav1.CIDR("10.0.0.0/8"),
						infrav1.CIDR("0.0.0.0/0"),
					},
				},
				ControlPlaneExtraLoadBalancers: []infrav1.LoadBalancer{
					{
						Zone:      infrav1.ScalewayZone("fr-par-1"),
						Type:      "LB-M",
						IP:        infrav1.IPv4("22.22.22.22"),
						PrivateIP: infrav1.IPv4("10.0.0.2"),
					},
					{
						Zone:      infrav1.ScalewayZone("fr-par-3"),
						Type:      "LB-L",
						IP:        infrav1.IPv4("33.33.33.33"),
						PrivateIP: infrav1.IPv4("10.0.0.3"),
					},
				},
				ControlPlaneDNS: infrav1.ControlPlaneDNS{
					Name:   "testing-cluster",
					Domain: "example.com",
				},
				PrivateNetwork: infrav1.PrivateNetworkSpec{
					PrivateNetwork: infrav1.PrivateNetwork{
						ID:     infrav1.UUID("22222222-2222-2222-2222-222222222222"),
						VPCID:  infrav1.UUID("33333333-3333-3333-3333-333333333333"),
						Subnet: infrav1.CIDR("192.168.0.0/16"),
					},
					Enabled: ptr.To(true),
				},
				PublicGateways: []infrav1.PublicGateway{
					{
						Type: "PGW-M",
						IP:   infrav1.IPv4("44.44.44.44"),
						Zone: infrav1.ScalewayZone("fr-par-1"),
					},
					{
						Type: "PGW-S",
						IP:   infrav1.IPv4("55.55.55.55"),
						Zone: "fr-par-2",
					},
				},
			},
		},
		Status: infrav1.ScalewayClusterStatus{
			Initialization: infrav1.ScalewayClusterInitializationStatus{
				Provisioned: ptr.To(true),
			},
		},
	}
)

func TestScalewayCluster_ConvertTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *ScalewayCluster
		want    *infrav1.ScalewayCluster
		wantErr bool
	}{
		{
			name: "v1alpha1 to v1alpha2, private lb",
			src:  v1alpha1PrivateScalewayCluster,
			want: v1alpha2PrivateScalewayCluster,
		},
		{
			name: "v1alpha1 to v1alpha2, public lb",
			src:  v1alpha1PublicScalewayCluster,
			want: v1alpha2PublicScalewayCluster,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &infrav1.ScalewayCluster{}
			gotErr := tt.src.ConvertTo(dst)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertTo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertTo() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

func TestScalewayCluster_ConvertFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *infrav1.ScalewayCluster
		want    *ScalewayCluster
		wantErr bool
	}{
		{
			name: "v1alpha2 to v1alpha1, private lb",
			src:  v1alpha2PrivateScalewayCluster,
			want: v1alpha1PrivateScalewayCluster,
		},
		{
			name: "v1alpha2 to v1alpha1, public lb",
			src:  v1alpha2PublicScalewayCluster,
			want: v1alpha1PublicScalewayCluster,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &ScalewayCluster{}
			gotErr := dst.ConvertFrom(tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertFrom() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

// ScalewayMachines.
var (
	v1alpha1ScalewayMachine = &ScalewayMachine{
		Spec: ScalewayMachineSpec{
			ProviderID:     ptr.To("scaleway://instance/fr-par-1/11111111-1111-1111-1111-111111111111"),
			CommercialType: "DEV1-S",
			Image: ImageSpec{
				Name: ptr.To("scaleway-image"),
			},
			RootVolume: &RootVolumeSpec{
				Size: ptr.To(int64(42)),
				Type: ptr.To("block"),
				IOPS: ptr.To(int64(15000)),
			},
			PublicNetwork: &PublicNetworkSpec{
				EnableIPv4: ptr.To(true),
				EnableIPv6: ptr.To(false),
			},
			PlacementGroup: &PlacementGroupSpec{
				Name: ptr.To("scaleway-placement-group"),
			},
			SecurityGroup: &SecurityGroupSpec{
				Name: ptr.To("scaleway-security-group"),
			},
		},
		Status: ScalewayMachineStatus{
			Ready: true,
		},
	}
	v1alpha2ScalewayMachine = &infrav1.ScalewayMachine{
		Spec: infrav1.ScalewayMachineSpec{
			ProviderID:     "scaleway://instance/fr-par-1/11111111-1111-1111-1111-111111111111",
			CommercialType: "DEV1-S",
			Image: infrav1.Image{
				IDOrName: infrav1.IDOrName{
					Name: "scaleway-image",
				},
			},
			RootVolume: infrav1.RootVolume{
				Size: 42,
				Type: "block",
				IOPS: 15000,
			},
			PublicNetwork: infrav1.PublicNetwork{
				EnableIPv4: ptr.To(true),
				EnableIPv6: ptr.To(false),
			},
			PlacementGroup: infrav1.IDOrName{
				Name: "scaleway-placement-group",
			},
			SecurityGroup: infrav1.IDOrName{
				Name: "scaleway-security-group",
			},
		},
		Status: infrav1.ScalewayMachineStatus{
			Initialization: infrav1.ScalewayMachineInitializationStatus{
				Provisioned: ptr.To(true),
			},
		},
	}
)

func TestScalewayMachine_ConvertTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *ScalewayMachine
		want    *infrav1.ScalewayMachine
		wantErr bool
	}{
		{
			name: "v1alpha1 to v1alpha2",
			src:  v1alpha1ScalewayMachine,
			want: v1alpha2ScalewayMachine,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &infrav1.ScalewayMachine{}
			gotErr := tt.src.ConvertTo(dst)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertTo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertTo() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

func TestScalewayMachine_ConvertFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *infrav1.ScalewayMachine
		want    *ScalewayMachine
		wantErr bool
	}{
		{
			name: "v1alpha2 to v1alpha1",
			src:  v1alpha2ScalewayMachine,
			want: v1alpha1ScalewayMachine,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &ScalewayMachine{}
			gotErr := dst.ConvertFrom(tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertFrom() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

// ScalewayManagedClusters.
var (
	v1alpha1ScalewayManagedCluster = &ScalewayManagedCluster{
		Spec: ScalewayManagedClusterSpec{
			Region:             "fr-par",
			ProjectID:          "11111111-1111-1111-1111-111111111111",
			ScalewaySecretName: "scaleway-secret",
			ControlPlaneEndpoint: clusterv1beta1.APIEndpoint{
				Host: "cluster.endpoint.example.com",
				Port: 6443,
			},
			Network: &ManagedNetworkSpec{
				PrivateNetwork: &PrivateNetworkParams{
					ID:     ptr.To("22222222-2222-2222-2222-222222222222"),
					VPCID:  ptr.To("33333333-3333-3333-3333-333333333333"),
					Subnet: ptr.To("192.168.0.0/16"),
				},
				PublicGateways: []PublicGatewaySpec{
					{
						Type: ptr.To("PGW-M"),
						IP:   ptr.To("44.44.44.44"),
						Zone: ptr.To("fr-par-1"),
					},
					{
						Type: ptr.To("PGW-S"),
						IP:   ptr.To("55.55.55.55"),
						Zone: ptr.To("fr-par-2"),
					},
				},
			},
		},
		Status: ScalewayManagedClusterStatus{
			Ready: true,
		},
	}
	v1alpha2ScalewayManagedCluster = &infrav1.ScalewayManagedCluster{
		Spec: infrav1.ScalewayManagedClusterSpec{
			Region:             infrav1.ScalewayRegion("fr-par"),
			ProjectID:          infrav1.UUID("11111111-1111-1111-1111-111111111111"),
			ScalewaySecretName: "scaleway-secret",
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "cluster.endpoint.example.com",
				Port: 6443,
			},
			Network: infrav1.ScalewayManagedClusterNetwork{
				PrivateNetwork: infrav1.PrivateNetwork{
					ID:     infrav1.UUID("22222222-2222-2222-2222-222222222222"),
					VPCID:  infrav1.UUID("33333333-3333-3333-3333-333333333333"),
					Subnet: infrav1.CIDR("192.168.0.0/16"),
				},
				PublicGateways: []infrav1.PublicGateway{
					{
						Type: "PGW-M",
						IP:   infrav1.IPv4("44.44.44.44"),
						Zone: infrav1.ScalewayZone("fr-par-1"),
					},
					{
						Type: "PGW-S",
						IP:   infrav1.IPv4("55.55.55.55"),
						Zone: "fr-par-2",
					},
				},
			},
		},
		Status: infrav1.ScalewayManagedClusterStatus{
			Initialization: infrav1.ScalewayManagedClusterInitializationStatus{
				Provisioned: ptr.To(true),
			},
		},
	}
)

func TestScalewayManagedCluster_ConvertTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *ScalewayManagedCluster
		want    *infrav1.ScalewayManagedCluster
		wantErr bool
	}{
		{
			name: "v1alpha1 to v1alpha2",
			src:  v1alpha1ScalewayManagedCluster,
			want: v1alpha2ScalewayManagedCluster,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &infrav1.ScalewayManagedCluster{}
			gotErr := tt.src.ConvertTo(dst)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertTo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertTo() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

func TestScalewayManagedCluster_ConvertFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *infrav1.ScalewayManagedCluster
		want    *ScalewayManagedCluster
		wantErr bool
	}{
		{
			name: "v1alpha2 to v1alpha1",
			src:  v1alpha2ScalewayManagedCluster,
			want: v1alpha1ScalewayManagedCluster,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &ScalewayManagedCluster{}
			gotErr := dst.ConvertFrom(tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertFrom() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

// ScalewayManagedControlPlanes.
var (
	v1alpha1ScalewayManagedControlPlane = &ScalewayManagedControlPlane{
		Spec: ScalewayManagedControlPlaneSpec{
			ClusterName:    ptr.To("my-cluster"),
			Type:           "kapsule",
			Version:        "1.23.4",
			CNI:            ptr.To("cilium"),
			AdditionalTags: []string{"tag1", "tag2"},
			Autoscaler: &AutoscalerSpec{
				ScaleDownDisabled:             ptr.To(true),
				ScaleDownDelayAfterAdd:        ptr.To("1m"),
				Estimator:                     ptr.To("binpacking"),
				Expander:                      ptr.To("random"),
				IgnoreDaemonsetsUtilization:   ptr.To(true),
				BalanceSimilarNodeGroups:      ptr.To(true),
				ExpendablePodsPriorityCutoff:  ptr.To(int32(5)),
				ScaleDownUnneededTime:         ptr.To("2m"),
				ScaleDownUtilizationThreshold: ptr.To("0.1"),
				MaxGracefulTerminationSec:     ptr.To(int32(60)),
			},
			AutoUpgrade: &AutoUpgradeSpec{
				Enabled: true,
				MaintenanceWindow: &MaintenanceWindowSpec{
					StartHour: ptr.To(int32(5)),
					Day:       ptr.To("monday"),
				},
			},
			FeatureGates:     []string{"fg1", "fg2"},
			AdmissionPlugins: []string{"ap1", "ap2"},
			OpenIDConnect: &OpenIDConnectSpec{
				IssuerURL:      "https://auth.example.com",
				ClientID:       "abcd",
				UsernameClaim:  ptr.To("username"),
				UsernamePrefix: ptr.To("username-"),
				GroupsClaim:    []string{"group", "gp"},
				GroupsPrefix:   ptr.To("group-"),
				RequiredClaim:  []string{"test"},
			},
			APIServerCertSANs: []string{"san1", "san2"},
			OnDelete: &OnDeleteSpec{
				WithAdditionalResources: ptr.To(true),
			},
			ACL: &ACLSpec{
				AllowedRanges: []CIDR{"0.0.0.0/0", "10.0.0.0/8"},
			},
			EnablePrivateEndpoint: ptr.To(true),
			ControlPlaneEndpoint: clusterv1beta1.APIEndpoint{
				Host: "private.internal",
				Port: 6443,
			},
		},
		Status: ScalewayManagedControlPlaneStatus{
			Ready:                       true,
			Initialized:                 true,
			ExternalManagedControlPlane: true,
			Version:                     ptr.To("1.23.4"),
		},
	}
	v1alpha2ScalewayManagedControlPlane = &infrav1.ScalewayManagedControlPlane{
		Spec: infrav1.ScalewayManagedControlPlaneSpec{
			ClusterName:    "my-cluster",
			Type:           "kapsule",
			Version:        "1.23.4",
			CNI:            "cilium",
			AdditionalTags: []string{"tag1", "tag2"},
			Autoscaler: infrav1.Autoscaler{
				ScaleDownDisabled:             ptr.To(true),
				ScaleDownDelayAfterAdd:        "1m",
				Estimator:                     "binpacking",
				Expander:                      "random",
				IgnoreDaemonsetsUtilization:   ptr.To(true),
				BalanceSimilarNodeGroups:      ptr.To(true),
				ExpendablePodsPriorityCutoff:  ptr.To(int32(5)),
				ScaleDownUnneededTime:         "2m",
				ScaleDownUtilizationThreshold: "0.1",
				MaxGracefulTerminationSec:     60,
			},
			AutoUpgrade: infrav1.AutoUpgrade{
				Enabled: ptr.To(true),
				MaintenanceWindow: infrav1.MaintenanceWindow{
					StartHour: ptr.To(int32(5)),
					Day:       "monday",
				},
			},
			FeatureGates:     []string{"fg1", "fg2"},
			AdmissionPlugins: []string{"ap1", "ap2"},
			OpenIDConnect: infrav1.OpenIDConnect{
				IssuerURL:      "https://auth.example.com",
				ClientID:       "abcd",
				UsernameClaim:  "username",
				UsernamePrefix: "username-",
				GroupsClaim:    []string{"group", "gp"},
				GroupsPrefix:   "group-",
				RequiredClaim:  []string{"test"},
			},
			APIServerCertSANs: []string{"san1", "san2"},
			OnDelete: infrav1.OnDelete{
				WithAdditionalResources: ptr.To(true),
			},
			ACL: &infrav1.ACL{
				AllowedRanges: []infrav1.CIDR{"0.0.0.0/0", "10.0.0.0/8"},
			},
			EnablePrivateEndpoint: ptr.To(true),
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "private.internal",
				Port: 6443,
			},
		},
		Status: infrav1.ScalewayManagedControlPlaneStatus{
			Version:                     "1.23.4",
			ExternalManagedControlPlane: ptr.To(true),
			Initialization: infrav1.ScalewayManagedControlPlaneInitializationStatus{
				ControlPlaneInitialized: ptr.To(true),
			},
		},
	}
)

func TestScalewayManagedControlPlane_ConvertTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *ScalewayManagedControlPlane
		want    *infrav1.ScalewayManagedControlPlane
		wantErr bool
	}{
		{
			name: "v1alpha1 to v1alpha2",
			src:  v1alpha1ScalewayManagedControlPlane,
			want: v1alpha2ScalewayManagedControlPlane,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &infrav1.ScalewayManagedControlPlane{}
			gotErr := tt.src.ConvertTo(dst)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertTo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertTo() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

func TestScalewayManagedControlPlane_ConvertFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *infrav1.ScalewayManagedControlPlane
		want    *ScalewayManagedControlPlane
		wantErr bool
	}{
		{
			name: "v1alpha2 to v1alpha1",
			src:  v1alpha2ScalewayManagedControlPlane,
			want: v1alpha1ScalewayManagedControlPlane,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &ScalewayManagedControlPlane{}
			gotErr := dst.ConvertFrom(tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertFrom() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

// ScalewayManagedMachinePools
var (
	v1alpha1ScalewayManagedMachinePool = &ScalewayManagedMachinePool{
		Spec: ScalewayManagedMachinePoolSpec{
			NodeType:         "DEV1-S",
			Zone:             "fr-par-1",
			PlacementGroupID: ptr.To("11111111-1111-1111-1111-111111111111"),
			Scaling: &ScalingSpec{
				Autoscaling: ptr.To(true),
				MinSize:     ptr.To(int32(1)),
				MaxSize:     ptr.To(int32(10)),
			},
			Autohealing:    ptr.To(true),
			AdditionalTags: []string{"tag1", "tag2"},
			KubeletArgs: map[string]string{
				"arg1": "val1",
				"arg2": "val2",
			},
			UpgradePolicy: &UpgradePolicySpec{
				MaxUnavailable: ptr.To(int32(0)),
				MaxSurge:       ptr.To(int32(5)),
			},
			RootVolumeType:   ptr.To("sbs_5k"),
			RootVolumeSizeGB: ptr.To(int64(42)),
			PublicIPDisabled: ptr.To(true),
			SecurityGroupID:  ptr.To("22222222-2222-2222-2222-222222222222"),
			ProviderIDList:   []string{"id1", "id2"},
		},
		Status: ScalewayManagedMachinePoolStatus{
			Ready:    true,
			Replicas: 2,
		},
	}
	v1alpha2ScalewayManagedMachinePool = &infrav1.ScalewayManagedMachinePool{
		Spec: infrav1.ScalewayManagedMachinePoolSpec{
			NodeType:         "DEV1-S",
			Zone:             infrav1.ScalewayZone("fr-par-1"),
			PlacementGroupID: infrav1.UUID("11111111-1111-1111-1111-111111111111"),
			Scaling: infrav1.Scaling{
				Autoscaling: ptr.To(true),
				MinSize:     ptr.To(int32(1)),
				MaxSize:     ptr.To(int32(10)),
			},
			Autohealing:    ptr.To(true),
			AdditionalTags: []string{"tag1", "tag2"},
			KubeletArgs: map[string]string{
				"arg1": "val1",
				"arg2": "val2",
			},
			UpgradePolicy: infrav1.UpgradePolicy{
				MaxUnavailable: ptr.To(int32(0)),
				MaxSurge:       ptr.To(int32(5)),
			},
			RootVolumeType:   "sbs_5k",
			RootVolumeSizeGB: 42,
			PublicIPDisabled: ptr.To(true),
			SecurityGroupID:  infrav1.UUID("22222222-2222-2222-2222-222222222222"),
			ProviderIDList:   []string{"id1", "id2"},
		},
		Status: infrav1.ScalewayManagedMachinePoolStatus{
			Ready: ptr.To(true),
			Initialization: infrav1.ScalewayManagedMachinePoolInitializationStatus{
				Provisioned: ptr.To(true),
			},
			Replicas: ptr.To(int32(2)),
		},
	}
)

func TestScalewayManagedMachinePool_ConvertTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *ScalewayManagedMachinePool
		want    *infrav1.ScalewayManagedMachinePool
		wantErr bool
	}{
		{
			name: "v1alpha1 to v1alpha2",
			src:  v1alpha1ScalewayManagedMachinePool,
			want: v1alpha2ScalewayManagedMachinePool,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &infrav1.ScalewayManagedMachinePool{}
			gotErr := tt.src.ConvertTo(dst)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertTo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertTo() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

func TestScalewayManagedMachinePool_ConvertFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *infrav1.ScalewayManagedMachinePool
		want    *ScalewayManagedMachinePool
		wantErr bool
	}{
		{
			name: "v1alpha2 to v1alpha1",
			src:  v1alpha2ScalewayManagedMachinePool,
			want: v1alpha1ScalewayManagedMachinePool,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &ScalewayManagedMachinePool{}
			gotErr := dst.ConvertFrom(tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertFrom() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

// ScalewayMachineTemplates
var (
	v1alpha1ScalewayMachineTemplate = &ScalewayMachineTemplate{
		Spec: ScalewayMachineTemplateSpec{
			Template: ScalewayMachineTemplateResource{
				Spec: v1alpha1ScalewayMachine.Spec,
			},
		},
	}

	v1alpha2ScalewayMachineTemplate = &infrav1.ScalewayMachineTemplate{
		Spec: infrav1.ScalewayMachineTemplateSpec{
			Template: infrav1.ScalewayMachineTemplateResource{
				Spec: v1alpha2ScalewayMachine.Spec,
			},
		},
	}
)

func TestScalewayMachineTemplate_ConvertTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *ScalewayMachineTemplate
		want    *infrav1.ScalewayMachineTemplate
		wantErr bool
	}{
		{
			name: "v1alpha1 to v1alpha2",
			src:  v1alpha1ScalewayMachineTemplate,
			want: v1alpha2ScalewayMachineTemplate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &infrav1.ScalewayMachineTemplate{}
			gotErr := tt.src.ConvertTo(dst)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertTo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertTo() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}

func TestScalewayMachineTemplate_ConvertFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string // description of this test case
		src     *infrav1.ScalewayMachineTemplate
		want    *ScalewayMachineTemplate
		wantErr bool
	}{
		{
			name: "v1alpha2 to v1alpha1",
			src:  v1alpha2ScalewayMachineTemplate,
			want: v1alpha1ScalewayMachineTemplate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := &ScalewayMachineTemplate{}
			gotErr := dst.ConvertFrom(tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ConvertFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ConvertFrom() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(dst, tt.want) {
				t.Fatalf("Conversion mismatch: expected %+v, got %+v", tt.want, dst)
			}
		})
	}
}
