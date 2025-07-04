package domain

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
)

const (
	zone             = "zone.example.com"
	name             = "cluster"
	lbIP             = "42.42.42.42"
	privateNetworkID = "11111111-1111-1111-1111-111111111111"
	vpcID            = "11111111-1111-1111-1111-111111111111"
)

var (
	extraLBIPs = []string{"11.11.11.11", "22.22.22.22"}
)

func TestService_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		Cluster *scope.Cluster
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
			name: "no dns",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect:  func(i *mock_client.MockInterfaceMockRecorder) {},
		},
		{
			name: "public dns: set zone records",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								ControlPlaneDNS: &v1alpha1.ControlPlaneDNSSpec{
									Domain: zone,
									Name:   name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{}, nil)
				i.SetDNSZoneRecords(gomock.Any(), zone, name, append(extraLBIPs, lbIP))
			},
		},
		{
			name: "public dns: up-to-date",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								ControlPlaneDNS: &v1alpha1.ControlPlaneDNSSpec{
									Domain: zone,
									Name:   name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{
					{Data: extraLBIPs[0]},
					{Data: extraLBIPs[1]},
					{Data: lbIP},
				}, nil)
			},
		},
		{
			name: "private dns: set zone records",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								PrivateNetwork: &v1alpha1.PrivateNetworkSpec{
									Enabled: true,
								},
								ControlPlaneLoadBalancer: &v1alpha1.ControlPlaneLoadBalancerSpec{
									Private: scw.BoolPtr(true),
								},
								ControlPlanePrivateDNS: &v1alpha1.ControlPlanePrivateDNSSpec{
									Name: name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								VPCID:                scw.StringPtr(vpcID),
								PrivateNetworkID:     scw.StringPtr(privateNetworkID),
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				zone := fmt.Sprintf("%s.%s.privatedns", privateNetworkID, vpcID)
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{}, nil)
				i.SetDNSZoneRecords(gomock.Any(), zone, name, append(extraLBIPs, lbIP))
			},
		},
		{
			name: "private dns: up-to-date",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								PrivateNetwork: &v1alpha1.PrivateNetworkSpec{
									Enabled: true,
								},
								ControlPlaneLoadBalancer: &v1alpha1.ControlPlaneLoadBalancerSpec{
									Private: scw.BoolPtr(true),
								},
								ControlPlanePrivateDNS: &v1alpha1.ControlPlanePrivateDNSSpec{
									Name: name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								VPCID:                scw.StringPtr(vpcID),
								PrivateNetworkID:     scw.StringPtr(privateNetworkID),
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				zone := fmt.Sprintf("%s.%s.privatedns", privateNetworkID, vpcID)
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{
					{Data: extraLBIPs[0]},
					{Data: extraLBIPs[1]},
					{Data: lbIP},
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
				Cluster: tt.fields.Cluster,
			}
			s.Cluster.ScalewayClient = scwMock
			if err := s.Reconcile(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	type fields struct {
		Cluster *scope.Cluster
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
			name: "no dns",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect:  func(i *mock_client.MockInterfaceMockRecorder) {},
		},
		{
			name: "public dns: ignore missing zone",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								ControlPlaneDNS: &v1alpha1.ControlPlaneDNSSpec{
									Domain: zone,
									Name:   name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return(nil, &scw.ResponseError{
					StatusCode: http.StatusForbidden,
				})
			},
		},
		{
			name: "public dns: already deleted",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								ControlPlaneDNS: &v1alpha1.ControlPlaneDNSSpec{
									Domain: zone,
									Name:   name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{}, nil)
			},
		},
		{
			name: "public dns: delete records",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								ControlPlaneDNS: &v1alpha1.ControlPlaneDNSSpec{
									Domain: zone,
									Name:   name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{
					{Data: extraLBIPs[0]},
					{Data: extraLBIPs[1]},
					{Data: lbIP},
				}, nil)
				i.DeleteDNSZoneRecords(gomock.Any(), zone, name)
			},
		},
		{
			name: "private dns: already deleted",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								PrivateNetwork: &v1alpha1.PrivateNetworkSpec{
									Enabled: true,
								},
								ControlPlaneLoadBalancer: &v1alpha1.ControlPlaneLoadBalancerSpec{
									Private: scw.BoolPtr(true),
								},
								ControlPlanePrivateDNS: &v1alpha1.ControlPlanePrivateDNSSpec{
									Name: name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								VPCID:                scw.StringPtr(vpcID),
								PrivateNetworkID:     scw.StringPtr(privateNetworkID),
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				zone := fmt.Sprintf("%s.%s.privatedns", privateNetworkID, vpcID)
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{}, nil)
			},
		},
		{
			name: "private dns: delete records",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						Spec: v1alpha1.ScalewayClusterSpec{
							Network: &v1alpha1.NetworkSpec{
								PrivateNetwork: &v1alpha1.PrivateNetworkSpec{
									Enabled: true,
								},
								ControlPlaneLoadBalancer: &v1alpha1.ControlPlaneLoadBalancerSpec{
									Private: scw.BoolPtr(true),
								},
								ControlPlanePrivateDNS: &v1alpha1.ControlPlanePrivateDNSSpec{
									Name: name,
								},
							},
						},
						Status: v1alpha1.ScalewayClusterStatus{
							Network: &v1alpha1.NetworkStatus{
								VPCID:                scw.StringPtr(vpcID),
								PrivateNetworkID:     scw.StringPtr(privateNetworkID),
								LoadBalancerIP:       scw.StringPtr(lbIP),
								ExtraLoadBalancerIPs: *scw.StringsPtr(extraLBIPs),
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				zone := fmt.Sprintf("%s.%s.privatedns", privateNetworkID, vpcID)
				i.ListDNSZoneRecords(gomock.Any(), zone, name).Return([]*domain.Record{
					{Data: extraLBIPs[0]},
					{Data: extraLBIPs[1]},
					{Data: lbIP},
				}, nil)
				i.DeleteDNSZoneRecords(gomock.Any(), zone, name)
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
				Cluster: tt.fields.Cluster,
			}
			s.ScalewayClient = scwMock
			if err := s.Delete(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
