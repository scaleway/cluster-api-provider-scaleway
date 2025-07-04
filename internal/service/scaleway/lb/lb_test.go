package lb

import (
	"context"
	"net"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	privateNetworkID = "11111111-1111-1111-1111-111111111111"

	lbID  = "00000000-0000-0000-0000-000000000000"
	lbID1 = "11111111-1111-1111-1111-111111111111"
	lbID2 = "22222222-2222-2222-2222-222222222222"
	lbID3 = "33333333-3333-3333-3333-333333333333"

	backendID  = "11111111-1111-1111-1111-111111111111"
	backendID1 = "11111111-1111-1111-1111-111111111111"
	backendID2 = "11111111-1111-1111-1111-111111111111"
	backendID3 = "11111111-1111-1111-1111-111111111111"

	frontendID  = "55555555-5555-5555-5555-555555555555"
	frontendID1 = "66666666-6666-6666-6666-666666666666"
	frontendID2 = "77777777-7777-7777-7777-777777777777"
	frontendID3 = "88888888-8888-8888-8888-888888888888"

	lbIP  = "1.1.1.1"
	lbIP1 = "2.2.2.2"
	lbIP2 = "3.3.3.3"
	lbIP3 = "4.4.4.4"
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
		asserts func(g *WithT, c *scope.Cluster)
	}{
		{
			name: "public LB, no extra LB, no Private Network, no ACL: create",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &v1alpha1.ScalewayCluster{
						ObjectMeta: v1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				// Main LB
				i.GetZoneOrDefault(nil).Return(scw.ZoneFrPar1, nil)
				i.FindLB(gomock.Any(), scw.ZoneFrPar1, append(tags, CAPSMainLBTag)).Return(nil, client.ErrNoItemFound)
				i.CreateLB(gomock.Any(), scw.ZoneFrPar1, "cluster", "LB-S", nil, false, append(tags, CAPSMainLBTag)).Return(&lb.LB{
					ID:     lbID,
					Name:   "cluster",
					Status: lb.LBStatusReady,
					Zone:   scw.ZoneFrPar1,
					IP:     []*lb.IP{{IPAddress: "42.42.42.42"}},
				}, nil)

				// Extra LBs
				i.FindLBs(gomock.Any(), append(tags, CAPSExtraLBTag)).Return([]*lb.LB{}, nil)

				// Backend
				i.FindBackend(gomock.Any(), scw.ZoneFrPar1, lbID, BackendName).Return(nil, client.ErrNoItemFound)
				i.CreateBackend(gomock.Any(), scw.ZoneFrPar1, lbID, BackendName, nil, backendControlPlanePort).Return(&lb.Backend{
					ID: backendID,
					LB: &lb.LB{
						ID:   lbID,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)

				// Frontend
				i.FindFrontend(gomock.Any(), scw.ZoneFrPar1, lbID, FrontendName).Return(nil, client.ErrNoItemFound)
				i.CreateFrontend(gomock.Any(), scw.ZoneFrPar1, lbID, FrontendName, backendID, int32(6443)).Return(&lb.Frontend{
					ID: frontendID,
					LB: &lb.LB{
						ID:   lbID,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)

				// ACL
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, allowedRangesACLName).Return(nil, client.ErrNoItemFound)
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, publicGatewayACLName).Return(nil, client.ErrNoItemFound)
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, denyAllACLName).Return(nil, client.ErrNoItemFound)
			},
			asserts: func(g *WithT, c *scope.Cluster) {
				g.Expect(c.ScalewayCluster.Status.Network).ToNot(BeNil())
				g.Expect(c.ScalewayCluster.Status.Network.LoadBalancerIP).To(Equal(scw.StringPtr("42.42.42.42")))
			},
		},
		{
			name: "public LB, extra LBs, Private Network, ACL: up-to-date",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: v1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								ControlPlaneLoadBalancer: &infrav1.ControlPlaneLoadBalancerSpec{
									AllowedRanges: []infrav1.CIDR{"10.10.0.0/16"},
								},
								ControlPlaneExtraLoadBalancers: []infrav1.LoadBalancerSpec{
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-2")},
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
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				// Main LB
				i.GetZoneOrDefault(nil).Return(scw.ZoneFrPar1, nil)
				i.FindLB(gomock.Any(), scw.ZoneFrPar1, append(tags, CAPSMainLBTag)).Return(&lb.LB{
					ID:     lbID,
					Name:   "cluster",
					Status: lb.LBStatusReady,
					Zone:   scw.ZoneFrPar1,
					IP:     []*lb.IP{{IPAddress: lbIP}},
					Type:   "LB-S",
				}, nil)

				// Extra LBs
				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil).Times(2)
				i.GetZoneOrDefault(scw.StringPtr("fr-par-2")).Return(scw.ZoneFrPar2, nil)
				i.FindLBs(gomock.Any(), append(tags, CAPSExtraLBTag)).Return([]*lb.LB{
					{
						ID:     lbID1,
						Name:   "cluster-0",
						Status: lb.LBStatusReady,
						Zone:   scw.ZoneFrPar1,
						IP:     []*lb.IP{{IPAddress: lbIP1}},
						Tags:   append(tags, capsManagedIPTag, CAPSExtraLBTag),
					},
					{
						ID:     lbID2,
						Name:   "cluster-1",
						Status: lb.LBStatusReady,
						Zone:   scw.ZoneFrPar1,
						IP:     []*lb.IP{{IPAddress: lbIP2}},
						Tags:   append(tags, capsManagedIPTag, CAPSExtraLBTag),
					}, {
						ID:     lbID3,
						Name:   "cluster-0",
						Status: lb.LBStatusReady,
						Zone:   scw.ZoneFrPar2,
						IP:     []*lb.IP{{IPAddress: lbIP3}},
						Tags:   append(tags, capsManagedIPTag, CAPSExtraLBTag),
					},
				}, nil)

				// Private Network
				i.FindLBPrivateNetwork(gomock.Any(), scw.ZoneFrPar1, lbID1, privateNetworkID).Return(&lb.PrivateNetwork{PrivateNetworkID: privateNetworkID}, nil)
				i.FindLBPrivateNetwork(gomock.Any(), scw.ZoneFrPar1, lbID2, privateNetworkID).Return(&lb.PrivateNetwork{PrivateNetworkID: privateNetworkID}, nil)
				i.FindLBPrivateNetwork(gomock.Any(), scw.ZoneFrPar2, lbID3, privateNetworkID).Return(&lb.PrivateNetwork{PrivateNetworkID: privateNetworkID}, nil)
				i.FindLBPrivateNetwork(gomock.Any(), scw.ZoneFrPar1, lbID, privateNetworkID).Return(&lb.PrivateNetwork{PrivateNetworkID: privateNetworkID}, nil)
				i.FindLBServersIPs(gomock.Any(), privateNetworkID, []string{lbID1, lbID2, lbID3, lbID}).Return([]*ipam.IP{
					{
						Resource: &ipam.Resource{ID: lbID1},
						Address:  scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}},
					},
					{
						Resource: &ipam.Resource{ID: lbID2},
						Address:  scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 2), Mask: net.CIDRMask(24, 32)}},
					},
					{
						Resource: &ipam.Resource{ID: lbID3},
						Address:  scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 3), Mask: net.CIDRMask(24, 32)}},
					},
					{
						Resource: &ipam.Resource{ID: lbID},
						Address:  scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 4), Mask: net.CIDRMask(24, 32)}},
					},
				}, nil)

				// // Backends
				i.FindBackend(gomock.Any(), scw.ZoneFrPar1, lbID, BackendName).Return(&lb.Backend{
					ID: backendID,
					LB: &lb.LB{
						ID:   lbID,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)
				i.FindBackend(gomock.Any(), scw.ZoneFrPar1, lbID1, BackendName).Return(&lb.Backend{
					ID: backendID1,
					LB: &lb.LB{
						ID:   lbID1,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)
				i.FindBackend(gomock.Any(), scw.ZoneFrPar1, lbID2, BackendName).Return(&lb.Backend{
					ID: backendID2,
					LB: &lb.LB{
						ID:   lbID2,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)
				i.FindBackend(gomock.Any(), scw.ZoneFrPar2, lbID3, BackendName).Return(&lb.Backend{
					ID: backendID3,
					LB: &lb.LB{
						ID:   lbID3,
						Zone: scw.ZoneFrPar2,
					},
				}, nil)

				// Frontends
				i.FindFrontend(gomock.Any(), scw.ZoneFrPar1, lbID, FrontendName).Return(&lb.Frontend{
					ID: frontendID,
					LB: &lb.LB{
						ID:   lbID,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)
				i.FindFrontend(gomock.Any(), scw.ZoneFrPar1, lbID1, FrontendName).Return(&lb.Frontend{
					ID: frontendID1,
					LB: &lb.LB{
						ID:   lbID1,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)
				i.FindFrontend(gomock.Any(), scw.ZoneFrPar1, lbID2, FrontendName).Return(&lb.Frontend{
					ID: frontendID2,
					LB: &lb.LB{
						ID:   lbID2,
						Zone: scw.ZoneFrPar1,
					},
				}, nil)
				i.FindFrontend(gomock.Any(), scw.ZoneFrPar2, lbID3, FrontendName).Return(&lb.Frontend{
					ID: frontendID3,
					LB: &lb.LB{
						ID:   lbID3,
						Zone: scw.ZoneFrPar2,
					},
				}, nil)

				// ACL for main LB
				i.FindGateways(gomock.Any(), tags).Return([]*vpcgw.Gateway{
					{
						IPv4: &vpcgw.IP{
							Address: net.IPv4(42, 42, 42, 42),
						},
					},
				}, nil)
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, allowedRangesACLName).Return(&lb.ACL{
					Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
					Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"10.10.0.0/16"})},
				}, nil)
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, publicGatewayACLName).Return(&lb.ACL{
					Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
					Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"42.42.42.42"})},
				}, nil)
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, denyAllACLName).Return(&lb.ACL{
					Action: &lb.ACLAction{Type: lb.ACLActionTypeDeny},
					Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"0.0.0.0/0", "::/0"})},
				}, nil)

				// Duplicate ACLs from main LB to extra LBs
				i.ListLBACLs(gomock.Any(), scw.ZoneFrPar1, frontendID).Return([]*lb.ACL{
					{
						Name:   allowedRangesACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"10.10.0.0/16"})},
						Index:  aclIndex,
					},
					{
						Name:   publicGatewayACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"42.42.42.42"})},
						Index:  aclIndex,
					},
					{
						Name:   denyAllACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeDeny},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"0.0.0.0/0", "::/0"})},
						Index:  denyAllACLIndex,
					},
				}, nil)
				i.ListLBACLs(gomock.Any(), scw.ZoneFrPar1, frontendID1).Return([]*lb.ACL{
					{
						Name:   allowedRangesACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"10.10.0.0/16"})},
						Index:  aclIndex,
					},
					{
						Name:   publicGatewayACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"42.42.42.42"})},
						Index:  aclIndex,
					},
					{
						Name:   denyAllACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeDeny},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"0.0.0.0/0", "::/0"})},
						Index:  denyAllACLIndex,
					},
				}, nil)
				i.ListLBACLs(gomock.Any(), scw.ZoneFrPar1, frontendID2).Return([]*lb.ACL{
					{
						Name:   allowedRangesACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"10.10.0.0/16"})},
						Index:  aclIndex,
					},
					{
						Name:   publicGatewayACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"42.42.42.42"})},
						Index:  aclIndex,
					},
					{
						Name:   denyAllACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeDeny},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"0.0.0.0/0", "::/0"})},
						Index:  denyAllACLIndex,
					},
				}, nil)
				i.ListLBACLs(gomock.Any(), scw.ZoneFrPar2, frontendID3).Return([]*lb.ACL{
					{
						Name:   allowedRangesACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"10.10.0.0/16"})},
						Index:  aclIndex,
					},
					{
						Name:   publicGatewayACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"42.42.42.42"})},
						Index:  aclIndex,
					},
					{
						Name:   denyAllACLName,
						Action: &lb.ACLAction{Type: lb.ACLActionTypeDeny},
						Match:  &lb.ACLMatch{IPSubnet: scw.StringSlicePtr([]string{"0.0.0.0/0", "::/0"})},
						Index:  denyAllACLIndex,
					},
				}, nil)
			},
			asserts: func(g *WithT, c *scope.Cluster) {
				g.Expect(c.ScalewayCluster.Status.Network).ToNot(BeNil())
				g.Expect(c.ScalewayCluster.Status.Network.LoadBalancerIP).To(Equal(scw.StringPtr(lbIP)))
				g.Expect(c.ScalewayCluster.Status.Network.ExtraLoadBalancerIPs).To(Equal([]string{lbIP1, lbIP2, lbIP3}))
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

			s := &Service{
				Cluster: tt.fields.Cluster,
			}
			s.ScalewayClient = scwMock
			if err := s.Reconcile(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.asserts(g, s.Cluster)
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
			name: "delete all",
			fields: fields{
				Cluster: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: v1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: &infrav1.NetworkSpec{
								PrivateNetwork: &infrav1.PrivateNetworkSpec{
									Enabled: true,
								},
								ControlPlaneLoadBalancer: &infrav1.ControlPlaneLoadBalancerSpec{
									AllowedRanges: []infrav1.CIDR{"10.10.0.0/16"},
								},
								ControlPlaneExtraLoadBalancers: []infrav1.LoadBalancerSpec{
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-1")},
									{Zone: scw.StringPtr("fr-par-2")},
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
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				tags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}

				// Main LB
				i.GetZoneOrDefault(nil).Return(scw.ZoneFrPar1, nil)
				i.FindLB(gomock.Any(), scw.ZoneFrPar1, append(tags, CAPSMainLBTag)).Return(&lb.LB{
					ID:     lbID,
					Name:   "cluster",
					Status: lb.LBStatusReady,
					Zone:   scw.ZoneFrPar1,
					IP:     []*lb.IP{{IPAddress: lbIP}},
					Tags:   append(tags, CAPSMainLBTag, capsManagedIPTag),
					Type:   "LB-S",
				}, nil)
				i.DeleteLB(gomock.Any(), scw.ZoneFrPar1, lbID, true)

				// Extra LBs
				i.FindLBs(gomock.Any(), append(tags, CAPSExtraLBTag)).Return([]*lb.LB{
					{
						ID:     lbID1,
						Name:   "cluster-0",
						Status: lb.LBStatusReady,
						Zone:   scw.ZoneFrPar1,
						IP:     []*lb.IP{{IPAddress: lbIP1}},
						Tags:   append(tags, capsManagedIPTag, CAPSExtraLBTag),
					},
					{
						ID:     lbID2,
						Name:   "cluster-1",
						Status: lb.LBStatusReady,
						Zone:   scw.ZoneFrPar1,
						IP:     []*lb.IP{{IPAddress: lbIP2}},
						Tags:   append(tags, capsManagedIPTag, CAPSExtraLBTag),
					}, {
						ID:     lbID3,
						Name:   "cluster-0",
						Status: lb.LBStatusReady,
						Zone:   scw.ZoneFrPar2,
						IP:     []*lb.IP{{IPAddress: lbIP3}},
						Tags:   append(tags, capsManagedIPTag, CAPSExtraLBTag),
					},
				}, nil)
				i.DeleteLB(gomock.Any(), scw.ZoneFrPar1, lbID1, true)
				i.DeleteLB(gomock.Any(), scw.ZoneFrPar1, lbID2, true)
				i.DeleteLB(gomock.Any(), scw.ZoneFrPar2, lbID3, true)
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
