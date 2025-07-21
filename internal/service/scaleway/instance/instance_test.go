package instance

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	servicelb "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/lb"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	extraVolumeID    = "22222222-2222-2222-2222-222222222222"
	bootVolumeID     = "11111111-1111-1111-1111-111111111111"
	privateNetworkID = "11111111-1111-1111-1111-111111111111"
	privateNICID     = "11111111-1111-1111-1111-111111111111"
	frontendID       = "11111111-1111-1111-1111-111111111111"
	backendID        = "11111111-1111-1111-1111-111111111111"
	serverID         = "11111111-1111-1111-1111-111111111111"
	imageID          = "11111111-1111-1111-1111-111111111111"
	ipv4ID           = "11111111-1111-1111-1111-111111111111"
	ipv6ID           = "11111111-1111-1111-1111-111111111111"
	lbID             = "11111111-1111-1111-1111-111111111111"
	lbACLID          = "11111111-1111-1111-1111-111111111111"

	cloudInitBootstrap = `#cloud-config

bootcmd:
  - echo [[[ .NodeIP ]]]
`
	cloudInitData = `#cloud-config

bootcmd:
  - echo 10.0.0.1
`
)

func TestService_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		Machine *scope.Machine
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
		objects []runtime.Object
		asserts func(g *WithT, m *scope.Machine)
	}{
		{
			name: "create control-plane machine",
			fields: fields{
				Machine: &scope.Machine{
					Machine: &v1beta1.Machine{
						ObjectMeta: v1.ObjectMeta{
							Name:      "machine",
							Namespace: "default",
							Labels:    map[string]string{clusterv1.MachineControlPlaneLabel: ""},
						},
						Spec: v1beta1.MachineSpec{
							FailureDomain: scw.StringPtr("fr-par-1"),
							Bootstrap: clusterv1.Bootstrap{
								DataSecretName: scw.StringPtr("bootstrap"),
							},
						},
					},
					ScalewayMachine: &v1alpha1.ScalewayMachine{
						ObjectMeta: v1.ObjectMeta{
							Name:      "machine",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayMachineSpec{
							CommercialType: "DEV1-S",
							Image: v1alpha1.ImageSpec{
								ID: scw.StringPtr(imageID),
							},
							PublicNetwork: &v1alpha1.PublicNetworkSpec{
								EnableIPv4: scw.BoolPtr(true),
								EnableIPv6: scw.BoolPtr(true),
							},
							RootVolume: &v1alpha1.RootVolumeSpec{
								Size: scw.Int64Ptr(42),
							},
						},
					},
					Cluster: &scope.Cluster{
						ScalewayCluster: &v1alpha1.ScalewayCluster{
							ObjectMeta: v1.ObjectMeta{
								Name:      "cluster",
								Namespace: "default",
							},
							Spec: v1alpha1.ScalewayClusterSpec{
								Network: &v1alpha1.NetworkSpec{
									PrivateNetwork: &v1alpha1.PrivateNetworkSpec{
										Enabled: true,
									},
								},
							},
							Status: v1alpha1.ScalewayClusterStatus{
								Network: &v1alpha1.NetworkStatus{
									PrivateNetworkID: scw.StringPtr(privateNetworkID),
								},
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			objects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      "bootstrap",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"value": []byte(cloudInitBootstrap),
					},
				},
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				clusterTags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}
				tags := append(clusterTags, "caps-scalewaymachine=machine")

				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)
				i.FindServer(gomock.Any(), scw.ZoneFrPar1, tags).Return(nil, client.ErrNoItemFound)
				i.CreateServer(
					gomock.Any(),
					scw.ZoneFrPar1,
					"machine",
					"DEV1-S",
					imageID,
					nil,
					nil,
					42*scw.GB,
					instance.VolumeVolumeTypeSbsVolume,
					tags,
				).Return(&instance.Server{
					Name:     "machine",
					Hostname: "machine",
					ID:       serverID,
					Zone:     scw.ZoneFrPar1,
					State:    instance.ServerStateStopped,
				}, nil)
				i.FindIPs(gomock.Any(), scw.ZoneFrPar1, tags).Return([]*instance.IP{}, nil)
				i.CreateIP(gomock.Any(), scw.ZoneFrPar1, instance.IPTypeRoutedIPv4, tags).Return(&instance.IP{
					ID:      ipv4ID,
					Address: net.IPv4(42, 42, 42, 42),
				}, nil)
				i.CreateIP(gomock.Any(), scw.ZoneFrPar1, instance.IPTypeRoutedIPv6, tags).Return(&instance.IP{
					ID:      ipv6ID,
					Address: net.IP{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42},
				}, nil)
				i.UpdateServerPublicIPs(gomock.Any(), scw.ZoneFrPar1, serverID, []string{ipv4ID, ipv6ID}).Return(&instance.Server{
					Name:     "machine",
					Hostname: "machine",
					ID:       serverID,
					Zone:     scw.ZoneFrPar1,
					State:    instance.ServerStateStopped,
					PublicIPs: []*instance.ServerIP{
						{Address: net.IPv4(42, 42, 42, 42)},
						{Address: net.IP{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42}},
					},
				}, nil)
				i.CreatePrivateNIC(gomock.Any(), scw.ZoneFrPar1, serverID, privateNetworkID).Return(&instance.PrivateNIC{
					ID: privateNICID,
				}, nil)
				i.FindPrivateNICIPs(gomock.Any(), privateNICID).Return([]*ipam.IP{
					{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
				}, nil)

				// LB configuration
				i.GetZoneOrDefault(nil).Return(scw.ZoneFrPar1, nil) // Get main LB zone.
				i.FindLB(gomock.Any(), scw.ZoneFrPar1, append(clusterTags, servicelb.CAPSMainLBTag)).Return(&lb.LB{
					ID:   lbID,
					Zone: scw.ZoneFrPar1,
				}, nil)
				i.FindLBs(gomock.Any(), append(clusterTags, servicelb.CAPSExtraLBTag)).Return(nil, nil)
				i.FindBackend(gomock.Any(), scw.ZoneFrPar1, lbID, servicelb.BackendName).Return(&lb.Backend{
					ID: backendID,
				}, nil)
				i.AddBackendServer(gomock.Any(), scw.ZoneFrPar1, backendID, "10.0.0.1")
				i.FindFrontend(gomock.Any(), scw.ZoneFrPar1, lbID, servicelb.FrontendName).Return(&lb.Frontend{
					ID: frontendID,
				}, nil)
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, "machine").Return(nil, client.ErrNoItemFound)
				i.CreateLBACL(gomock.Any(), scw.ZoneFrPar1, frontendID, "machine", machineACLIndex, lb.ACLActionTypeAllow, []string{"42.42.42.42", "2a00::2a"})

				// Cloud Init
				i.GetAllServerUserData(gomock.Any(), scw.ZoneFrPar1, serverID).Return(map[string]io.Reader{}, nil)
				i.SetServerUserData(gomock.Any(), scw.ZoneFrPar1, serverID, cloudInitUserDataKey, cloudInitData)

				// Start
				i.ServerAction(gomock.Any(), scw.ZoneFrPar1, serverID, instance.ServerActionPoweron)
			},
			asserts: func(g *WithT, m *scope.Machine) {
				g.Expect(m.ScalewayMachine.Status.Addresses).To(Equal([]clusterv1.MachineAddress{
					{Type: clusterv1.MachineHostName, Address: "machine"},
					{Type: clusterv1.MachineExternalIP, Address: "42.42.42.42"},
					{Type: clusterv1.MachineExternalIP, Address: "2a00::2a"},
					{Type: clusterv1.MachineExternalDNS, Address: "11111111-1111-1111-1111-111111111111.pub.instances.scw.cloud"},
					{Type: clusterv1.MachineInternalIP, Address: "10.0.0.1"},
				}))
				g.Expect(m.ScalewayMachine.Spec.ProviderID).To(Equal(scw.StringPtr("scaleway://instance/fr-par-1/11111111-1111-1111-1111-111111111111")))
			},
		},
		{
			name: "node has joined cluster: need to clean userdata",
			fields: fields{
				Machine: &scope.Machine{
					Machine: &v1beta1.Machine{
						ObjectMeta: v1.ObjectMeta{
							Name:      "machine",
							Namespace: "default",
							Labels:    map[string]string{clusterv1.MachineControlPlaneLabel: ""},
						},
						Spec: v1beta1.MachineSpec{
							FailureDomain: scw.StringPtr("fr-par-1"),
							Bootstrap: clusterv1.Bootstrap{
								DataSecretName: scw.StringPtr("bootstrap"),
							},
						},
						Status: clusterv1.MachineStatus{
							NodeRef: &corev1.ObjectReference{
								Name: "cluster",
							},
						},
					},
					ScalewayMachine: &v1alpha1.ScalewayMachine{
						ObjectMeta: v1.ObjectMeta{
							Name:      "machine",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayMachineSpec{
							CommercialType: "DEV1-S",
							Image: v1alpha1.ImageSpec{
								ID: scw.StringPtr(imageID),
							},
							PublicNetwork: &v1alpha1.PublicNetworkSpec{
								EnableIPv4: scw.BoolPtr(true),
								EnableIPv6: scw.BoolPtr(true),
							},
							RootVolume: &v1alpha1.RootVolumeSpec{
								Size: scw.Int64Ptr(42),
							},
							ProviderID: scw.StringPtr("scaleway://instance/fr-par-1/11111111-1111-1111-1111-111111111111"),
						},
					},
					Cluster: &scope.Cluster{
						ScalewayCluster: &v1alpha1.ScalewayCluster{
							ObjectMeta: v1.ObjectMeta{
								Name:      "cluster",
								Namespace: "default",
							},
							Spec: v1alpha1.ScalewayClusterSpec{
								Network: &v1alpha1.NetworkSpec{
									PrivateNetwork: &v1alpha1.PrivateNetworkSpec{
										Enabled: true,
									},
								},
							},
							Status: v1alpha1.ScalewayClusterStatus{
								Network: &v1alpha1.NetworkStatus{
									PrivateNetworkID: scw.StringPtr(privateNetworkID),
								},
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			objects: []runtime.Object{},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				clusterTags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}
				tags := append(clusterTags, "caps-scalewaymachine=machine")

				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)
				i.FindServer(gomock.Any(), scw.ZoneFrPar1, tags).Return(&instance.Server{
					Name:     "machine",
					Hostname: "machine",
					ID:       serverID,
					Zone:     scw.ZoneFrPar1,
					State:    instance.ServerStateStopped,
					PublicIPs: []*instance.ServerIP{
						{Address: net.IPv4(42, 42, 42, 42)},
						{Address: net.IP{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42}},
					},
				}, nil)
				i.GetAllServerUserData(gomock.Any(), scw.ZoneFrPar1, serverID).Return(map[string]io.Reader{
					cloudInitUserDataKey: strings.NewReader(cloudInitData),
				}, nil)
				i.DeleteServerUserData(gomock.Any(), scw.ZoneFrPar1, serverID, cloudInitUserDataKey)
			},
			asserts: func(g *WithT, m *scope.Machine) {},
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
				Machine: tt.fields.Machine,
			}
			s.ScalewayClient = scwMock
			s.Client = fake.NewFakeClient(tt.objects...)
			if err := s.Reconcile(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.asserts(g, s.Machine)
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	type fields struct {
		Machine *scope.Machine
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
			name: "invalid zone, do nothing",
			fields: fields{
				Machine: &scope.Machine{
					Machine: &clusterv1.Machine{
						Spec: clusterv1.MachineSpec{
							FailureDomain: scw.StringPtr("invalidvalue"),
						},
					},
					Cluster: &scope.Cluster{},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.GetZoneOrDefault(scw.StringPtr("invalidvalue")).Return(scw.Zone(""), errors.New("invalid zone"))
			},
		},
		{
			name: "delete control-plane machine",
			fields: fields{
				Machine: &scope.Machine{
					Machine: &v1beta1.Machine{
						ObjectMeta: v1.ObjectMeta{
							Name:      "machine",
							Namespace: "default",
							Labels:    map[string]string{clusterv1.MachineControlPlaneLabel: ""},
						},
						Spec: v1beta1.MachineSpec{
							FailureDomain: scw.StringPtr("fr-par-1"),
							Bootstrap: clusterv1.Bootstrap{
								DataSecretName: scw.StringPtr("bootstrap"),
							},
						},
						Status: clusterv1.MachineStatus{
							NodeRef: &corev1.ObjectReference{
								Name: "cluster",
							},
						},
					},
					ScalewayMachine: &v1alpha1.ScalewayMachine{
						ObjectMeta: v1.ObjectMeta{
							Name:      "machine",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayMachineSpec{
							CommercialType: "DEV1-S",
							Image: v1alpha1.ImageSpec{
								ID: scw.StringPtr(imageID),
							},
							PublicNetwork: &v1alpha1.PublicNetworkSpec{
								EnableIPv4: scw.BoolPtr(true),
								EnableIPv6: scw.BoolPtr(true),
							},
							RootVolume: &v1alpha1.RootVolumeSpec{
								Size: scw.Int64Ptr(42),
							},
							ProviderID: scw.StringPtr("scaleway://instance/fr-par-1/11111111-1111-1111-1111-111111111111"),
						},
					},
					Cluster: &scope.Cluster{
						ScalewayCluster: &v1alpha1.ScalewayCluster{
							ObjectMeta: v1.ObjectMeta{
								Name:      "cluster",
								Namespace: "default",
							},
							Spec: v1alpha1.ScalewayClusterSpec{
								Network: &v1alpha1.NetworkSpec{
									PrivateNetwork: &v1alpha1.PrivateNetworkSpec{
										Enabled: true,
									},
								},
							},
							Status: v1alpha1.ScalewayClusterStatus{
								Network: &v1alpha1.NetworkStatus{
									PrivateNetworkID: scw.StringPtr(privateNetworkID),
								},
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				clusterTags := []string{"caps-namespace=default", "caps-scalewaycluster=cluster"}
				tags := append(clusterTags, "caps-scalewaymachine=machine")

				i.GetZoneOrDefault(scw.StringPtr("fr-par-1")).Return(scw.ZoneFrPar1, nil)
				i.FindServer(gomock.Any(), scw.ZoneFrPar1, tags).Return(&instance.Server{
					Name:     "machine",
					Hostname: "machine",
					ID:       serverID,
					Zone:     scw.ZoneFrPar1,
					State:    instance.ServerStateStopped,
					PublicIPs: []*instance.ServerIP{
						{ID: ipv4ID, Address: net.IPv4(42, 42, 42, 42)},
						{ID: ipv6ID, Address: net.IP{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42}},
					},
					PrivateNics: []*instance.PrivateNIC{
						{ID: privateNICID, PrivateNetworkID: privateNetworkID},
					},
					Volumes: map[string]*instance.VolumeServer{
						"0": {
							ID:         bootVolumeID,
							Boot:       true,
							VolumeType: instance.VolumeServerVolumeTypeLSSD,
						},
						"1": {
							ID:         extraVolumeID,
							VolumeType: instance.VolumeServerVolumeTypeLSSD,
						},
					},
				}, nil)

				// LB config
				i.GetZoneOrDefault(nil).Return(scw.ZoneFrPar1, nil)
				i.FindLB(gomock.Any(), scw.ZoneFrPar1, append(clusterTags, servicelb.CAPSMainLBTag)).Return(&lb.LB{
					ID:   lbID,
					Zone: scw.ZoneFrPar1,
				}, nil)
				i.FindLBs(gomock.Any(), append(clusterTags, servicelb.CAPSExtraLBTag)).Return(nil, nil)
				i.FindFrontend(gomock.Any(), scw.ZoneFrPar1, lbID, servicelb.FrontendName).Return(&lb.Frontend{
					ID:   frontendID,
					Name: servicelb.FrontendName,
				}, nil)
				i.FindLBACLByName(gomock.Any(), scw.ZoneFrPar1, frontendID, "machine").Return(&lb.ACL{
					ID:   lbACLID,
					Name: "machine",
				}, nil)
				i.DeleteLBACL(gomock.Any(), scw.ZoneFrPar1, lbACLID)
				i.FindPrivateNICIPs(gomock.Any(), privateNICID).Return([]*ipam.IP{
					{Address: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)}}},
				}, nil)
				i.FindBackend(gomock.Any(), scw.ZoneFrPar1, lbID, servicelb.BackendName).Return(&lb.Backend{
					ID:   backendID,
					Pool: []string{"10.0.0.1"},
				}, nil)
				i.RemoveBackendServer(gomock.Any(), scw.ZoneFrPar1, backendID, "10.0.0.1")

				// Cleanup public IPs
				i.FindIPs(gomock.Any(), scw.ZoneFrPar1, tags).Return([]*instance.IP{{ID: ipv4ID}, {ID: ipv6ID}}, nil)
				i.DeleteIP(gomock.Any(), scw.ZoneFrPar1, ipv4ID)
				i.DeleteIP(gomock.Any(), scw.ZoneFrPar1, ipv6ID)

				// Volumes detach and remove
				i.UpdateInstanceVolumeTags(gomock.Any(), scw.ZoneFrPar1, bootVolumeID, tags)
				i.DetachVolume(gomock.Any(), scw.ZoneFrPar1, bootVolumeID)
				i.FindVolume(gomock.Any(), scw.ZoneFrPar1, tags).Return(&block.Volume{
					ID:     bootVolumeID,
					Status: block.VolumeStatusAvailable,
				}, nil)
				i.DeleteVolume(gomock.Any(), scw.ZoneFrPar1, bootVolumeID)

				// Delete serever
				i.DeleteServer(gomock.Any(), scw.ZoneFrPar1, serverID)
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
				Machine: tt.fields.Machine,
			}
			s.ScalewayClient = scwMock
			if err := s.Delete(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
