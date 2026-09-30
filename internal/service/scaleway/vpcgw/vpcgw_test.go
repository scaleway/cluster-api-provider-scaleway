package vpcgw

import (
	"context"
	"slices"
	"testing"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	privateNetworkID = "11111111-1111-1111-1111-111111111111"
	gwID1            = "11111111-1111-1111-1111-111111111111"
	gwID2            = "11111111-1111-1111-1111-111111111111"
	gwID3            = "11111111-1111-1111-1111-111111111111"
	gwID4            = "11111111-1111-1111-1111-111111111111"
	ipID             = "11111111-1111-1111-1111-111111111111"
)

func Test_canUpgradeTypes(t *testing.T) {
	t.Parallel()

	type args struct {
		types   []string
		current string
		desired string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "upgrade from VPC-GW-S to VPC-GW-XL",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "VPC-GW-XL",
			},
			want: true,
		},
		{
			name: "upgrade from VPC-GW-S to VPC-GW-M",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "VPC-GW-XL",
			},
			want: true,
		},
		{
			name: "current equals desired, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "VPC-GW-S",
			},
			want: false,
		},
		{
			name: "unknown current type, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "UNKNOWN-S",
				desired: "VPC-GW-L",
			},
			want: false,
		},
		{
			name: "unknown current and desired type, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "UNKNOWN-S",
				desired: "UNKNOWN-M",
			},
			want: false,
		},
		{
			name: "unknown desired type, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "UNKNOWN-M",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := canUpgradeTypes(tt.args.types, tt.args.current, tt.args.desired); got != tt.want {
				t.Errorf("canUpgradeTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		Scope
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(i *mock_client.MockInterfaceMockRecorder)
	}{
		{
			name: "no private network",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{},
				},
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {},
		},
		{
			name: "no gateway configured",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}
				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{}, nil)
			},
		},
		{
			name: "no gateway configured: delete existing",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}
				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{ID: gwID1, Zone: scw.ZoneFrPar1, Tags: []string{capsManagedIPTag}},
					{ID: gwID2, Zone: scw.ZoneFrPar2},
				}, nil)
				i.DeleteGateway(gomock.Any(), scw.ZoneFrPar1, gwID1, true)
				i.DeleteGateway(gomock.Any(), scw.ZoneFrPar2, gwID2, false)
			},
		},
		{
			name: "gateways configured: up-to-date",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-2")},
									{Zone: scw.StringPtr("fr-par-3")},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				// Indexing desired gateways by zone.
				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)
				i.GetZoneOrDefault(scw.StringPtr("fr-par-2")).Return(scw.ZoneFrPar2, nil)
				i.GetZoneOrDefault(scw.StringPtr("fr-par-3")).Return(scw.ZoneFrPar3, nil)

				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						ID:              gwID1,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
					{
						ID:              gwID2,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar2,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
					{
						ID:              gwID3,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar3,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
				}, nil)
			},
		},
		{
			name: "gateways configured: create missing",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-1"), IP: scw.StringPtr("42.42.42.42")},
									{Zone: scw.StringPtr("fr-par-1")},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"} //nolint:prealloc

				// Indexing desired gateways by zone.
				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil).Times(4)

				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						ID:              gwID1,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
					{
						ID:              gwID2,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-1",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
				}, nil)

				i.FindGatewayIP(gomock.Any(), scw.ZoneFrPar1, "42.42.42.42").Return(&vpcgw.IP{
					ID: ipID,
				}, nil)

				i.CreateGateway(
					gomock.Any(),
					scw.ZoneFrPar1,
					"cluster-2", "",
					tags,
					scw.StringPtr(ipID),
					false,
				).Return(&vpcgw.Gateway{
					ID:     gwID3,
					Name:   "cluster-2",
					Status: vpcgw.GatewayStatusRunning,
					Zone:   scw.ZoneFrPar1,
				}, nil)

				i.CreateGateway(
					gomock.Any(),
					scw.ZoneFrPar1,
					"cluster-3", "",
					append(tags, capsManagedIPTag),
					nil,
					false,
				).Return(&vpcgw.Gateway{
					ID:     gwID4,
					Name:   "cluster-3",
					Status: vpcgw.GatewayStatusRunning,
					Zone:   scw.ZoneFrPar1,
				}, nil)

				i.CreateGatewayNetwork(gomock.Any(), scw.ZoneFrPar1, gwID3, privateNetworkID)
				i.CreateGatewayNetwork(gomock.Any(), scw.ZoneFrPar1, gwID4, privateNetworkID)
			},
		},
		{
			name: "gateways configured: upgrade",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1"), Type: scw.StringPtr("VPC-GW-S")},
									{Zone: scw.StringPtr("fr-par-2"), Type: scw.StringPtr("VPC-GW-S")},
									{Zone: scw.StringPtr("fr-par-3"), Type: scw.StringPtr("VPC-GW-M")},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				// Indexing desired gateways by zone.
				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)
				i.GetZoneOrDefault(scw.StringPtr("fr-par-2")).Return(scw.ZoneFrPar2, nil)
				i.GetZoneOrDefault(scw.StringPtr("fr-par-3")).Return(scw.ZoneFrPar3, nil)

				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						ID:              gwID1,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
						Type:            "VPC-GW-S",
					},
					{
						ID:              gwID2,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar2,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
						Type:            "VPC-GW-S",
					},
					{
						ID:              gwID3,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar3,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
						Type:            "VPC-GW-S",
					},
				}, nil)

				i.ListGatewayTypes(gomock.Any(), scw.ZoneFrPar3).Return([]string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L"}, nil)

				i.UpgradeGateway(gomock.Any(), scw.ZoneFrPar3, gwID3, "VPC-GW-M").Return(&vpcgw.Gateway{
					ID:              gwID3,
					Status:          vpcgw.GatewayStatusRunning,
					Name:            "cluster-0",
					Zone:            scw.ZoneFrPar3,
					Tags:            []string{capsManagedIPTag},
					IPv4:            &vpcgw.IP{},
					GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					Type:            "VPC-GW-M",
				}, nil)
			},
		},
		{
			name: "gateways configured: create with smtp enabled",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1"), EnableSMTP: scw.BoolPtr(true)},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)

				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{}, nil)

				i.CreateGateway(
					gomock.Any(),
					scw.ZoneFrPar1,
					"cluster-0", "",
					slices.Concat(tags, []string{capsManagedIPTag}),
					nil,
					true,
				).Return(&vpcgw.Gateway{
					ID:          gwID1,
					Name:        "cluster-0",
					Status:      vpcgw.GatewayStatusRunning,
					Zone:        scw.ZoneFrPar1,
					Tags:        []string{capsManagedIPTag},
					IPv4:        &vpcgw.IP{},
					SMTPEnabled: true,
				}, nil)

				i.CreateGatewayNetwork(gomock.Any(), scw.ZoneFrPar1, gwID1, privateNetworkID)
			},
		},
		{
			name: "gateways configured: enable smtp on existing gateway",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1"), EnableSMTP: scw.BoolPtr(true)},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)

				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						ID:              gwID1,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
						SMTPEnabled:     false,
					},
				}, nil)

				i.UpdateGateway(gomock.Any(), scw.ZoneFrPar1, gwID1, scw.BoolPtr(true)).Return(&vpcgw.Gateway{
					ID:              gwID1,
					Status:          vpcgw.GatewayStatusRunning,
					Name:            "cluster-0",
					Zone:            scw.ZoneFrPar1,
					Tags:            []string{capsManagedIPTag},
					IPv4:            &vpcgw.IP{},
					GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					SMTPEnabled:     true,
				}, nil)
			},
		},
		{
			name: "gateways configured: disable smtp on existing gateway",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1"), EnableSMTP: scw.BoolPtr(false)},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)

				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						ID:              gwID1,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
						SMTPEnabled:     true,
					},
				}, nil)

				i.UpdateGateway(gomock.Any(), scw.ZoneFrPar1, gwID1, scw.BoolPtr(false)).Return(&vpcgw.Gateway{
					ID:              gwID1,
					Status:          vpcgw.GatewayStatusRunning,
					Name:            "cluster-0",
					Zone:            scw.ZoneFrPar1,
					Tags:            []string{capsManagedIPTag},
					IPv4:            &vpcgw.IP{},
					GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					SMTPEnabled:     false,
				}, nil)
			},
		},
		{
			name: "gateways configured: smtp unmanaged when field unset",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1")},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)

				// Gateway has SMTP enabled but spec doesn't set EnableSMTP → no UpdateGateway call.
				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						ID:              gwID1,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
						SMTPEnabled:     true,
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			scwMock := mock_client.NewMockInterface(mockCtrl)

			tt.expect(scwMock.EXPECT())

			s := &Service{
				Scope: tt.fields.Scope,
			}
			s.SetCloud(scwMock)
			if err := s.Reconcile(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	type fields struct {
		Scope
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expect  func(i *mock_client.MockInterfaceMockRecorder)
	}{
		{
			name: "no private network",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{},
				},
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {},
		},
		{
			name: "delete gateways",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								PublicGateways: []infrav1.PublicGatewaySpec{
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-2")},
									{Zone: scw.StringPtr("fr-par-3")},
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
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						ID:              gwID1,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar1,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
					{
						ID:              gwID2,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar2,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
					{
						ID:              gwID3,
						Status:          vpcgw.GatewayStatusRunning,
						Name:            "cluster-0",
						Zone:            scw.ZoneFrPar3,
						Tags:            []string{capsManagedIPTag},
						IPv4:            &vpcgw.IP{},
						GatewayNetworks: []*vpcgw.GatewayNetwork{{PrivateNetworkID: privateNetworkID}},
					},
				}, nil)

				i.DeleteGateway(gomock.Any(), scw.ZoneFrPar1, gwID1, true)
				i.DeleteGateway(gomock.Any(), scw.ZoneFrPar2, gwID2, true)
				i.DeleteGateway(gomock.Any(), scw.ZoneFrPar3, gwID3, true)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			scwMock := mock_client.NewMockInterface(mockCtrl)

			tt.expect(scwMock.EXPECT())

			s := &Service{
				Scope: tt.fields.Scope,
			}
			s.SetCloud(scwMock)
			if err := s.Delete(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
