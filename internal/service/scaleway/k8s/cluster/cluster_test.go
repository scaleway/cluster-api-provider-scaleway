package cluster

import (
	"context"
	"fmt"
	"net"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	clusterID        = "11111111-1111-1111-1111-111111111111"
	projectID        = "11111111-1111-1111-1111-111111111111"
	privateNetworkID = "11111111-1111-1111-1111-111111111111"
)

func TestService_Reconcile(t *testing.T) {
	t.Parallel()

	range0 := "0.0.0.0/0"
	_, ipNet0, err := net.ParseCIDR(range0)
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		ManagedControlPlane *scope.ManagedControlPlane
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		objects []runtime.Object
		expect  func(i *mock_client.MockInterfaceMockRecorder)
		asserts func(g *WithT, s *scope.ManagedControlPlane)
	}{
		{
			name: "create control-plane",
			fields: fields{
				ManagedControlPlane: &scope.ManagedControlPlane{
					Cluster: &v1beta1.Cluster{
						ObjectMeta: v1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: v1beta1.ClusterSpec{},
					},
					ManagedCluster: &v1alpha1.ScalewayManagedCluster{
						ObjectMeta: v1.ObjectMeta{
							Name:      "managedcluster",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayManagedClusterSpec{
							ProjectID: projectID,
						},
						Status: v1alpha1.ScalewayManagedClusterStatus{
							Ready: true,
							Network: &v1alpha1.ManagedNetworkStatus{
								PrivateNetworkID: scw.StringPtr("11111111-1111-1111-1111-111111111111"),
							},
						},
					},
					ManagedControlPlane: &v1alpha1.ScalewayManagedControlPlane{
						ObjectMeta: v1.ObjectMeta{
							Name:      "controlplane",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayManagedControlPlaneSpec{
							Type:    "kapsule",
							Version: "v1.31.1",
							CNI:     scw.StringPtr("cilium"),
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			objects: []runtime.Object{},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.FindCluster(gomock.Any(), "default-controlplane").Return(nil, client.ErrNoItemFound)
				i.CreateCluster(
					gomock.Any(),
					"default-controlplane",
					"kapsule",
					"1.31.1",
					scw.StringPtr(privateNetworkID),
					[]string{"caps-namespace=default", "caps-scalewaymanagedcontrolplane=controlplane"},
					nil,
					nil,
					nil,
					k8s.CNICilium,
					&k8s.CreateClusterRequestAutoscalerConfig{
						ScaleDownDisabled:             scw.BoolPtr(false),
						ScaleDownDelayAfterAdd:        scw.StringPtr("10m"),
						Estimator:                     k8s.AutoscalerEstimatorBinpacking,
						Expander:                      k8s.AutoscalerExpanderRandom,
						IgnoreDaemonsetsUtilization:   scw.BoolPtr(false),
						BalanceSimilarNodeGroups:      scw.BoolPtr(false),
						ExpendablePodsPriorityCutoff:  scw.Int32Ptr(-10),
						ScaleDownUnneededTime:         scw.StringPtr("10m"),
						ScaleDownUtilizationThreshold: scw.Float32Ptr(0.5),
						MaxGracefulTerminationSec:     scw.Uint32Ptr(600),
					},
					&k8s.CreateClusterRequestAutoUpgrade{
						Enable: false,
						MaintenanceWindow: &k8s.MaintenanceWindow{
							StartHour: 0,
							Day:       k8s.MaintenanceWindowDayOfTheWeekAny,
						},
					},
					&k8s.CreateClusterRequestOpenIDConnectConfig{
						UsernameClaim:  scw.StringPtr(""),
						UsernamePrefix: scw.StringPtr(""),
						GroupsPrefix:   scw.StringPtr(""),
						GroupsClaim:    &[]string{},
						RequiredClaim:  &[]string{},
					},
					scw.IPNet{},
					scw.IPNet{},
				).Return(&k8s.Cluster{
					ID:         clusterID,
					Status:     k8s.ClusterStatusReady,
					Type:       "kapsule",
					Version:    "1.31.1",
					Tags:       []string{"caps-namespace=default", "caps-scalewaymanagedcontrolplane=controlplane", "created-by=cluster-api-provider-scaleway"},
					Cni:        k8s.CNICilium,
					ClusterURL: fmt.Sprintf("https://%s.api.k8s.fr-par.scw.cloud:6443", clusterID),
					AutoscalerConfig: &k8s.ClusterAutoscalerConfig{
						ScaleDownDisabled:             false,
						ScaleDownDelayAfterAdd:        "10m",
						Estimator:                     k8s.AutoscalerEstimatorBinpacking,
						Expander:                      k8s.AutoscalerExpanderRandom,
						IgnoreDaemonsetsUtilization:   false,
						BalanceSimilarNodeGroups:      false,
						ExpendablePodsPriorityCutoff:  -10,
						ScaleDownUnneededTime:         "10m",
						ScaleDownUtilizationThreshold: 0.5,
						MaxGracefulTerminationSec:     600,
					},
					AutoUpgrade: &k8s.ClusterAutoUpgrade{
						Enabled: false,
						MaintenanceWindow: &k8s.MaintenanceWindow{
							StartHour: 0,
							Day:       k8s.MaintenanceWindowDayOfTheWeekAny,
						},
					},
					OpenIDConnectConfig: &k8s.ClusterOpenIDConnectConfig{},
				}, nil)
				i.ListClusterACLRules(gomock.Any(), clusterID).Return([]*k8s.ACLRule{
					{IP: &scw.IPNet{IPNet: *ipNet0}},
				}, nil)
				i.GetClusterKubeConfig(gomock.Any(), clusterID).Return(&k8s.Kubeconfig{
					Clusters: []*k8s.KubeconfigClusterWithName{
						{
							Name: "default-controlplane",
							Cluster: k8s.KubeconfigCluster{
								CertificateAuthorityData: "fake",
							},
						},
					},
				}, nil)
				i.GetSecretKey().Return("secret-key")
			},
			asserts: func(g *WithT, s *scope.ManagedControlPlane) {
				g.Expect(s.ManagedControlPlane.Spec.ClusterName).To(HaveValue(Equal("default-controlplane")))
				g.Expect(s.ManagedControlPlane.Status.Version).To(HaveValue(Equal("v1.31.1")))
				g.Expect(s.ManagedControlPlane.Spec.ControlPlaneEndpoint.Host).To(Equal(fmt.Sprintf("%s.api.k8s.fr-par.scw.cloud", clusterID)))
				g.Expect(s.ManagedControlPlane.Spec.ControlPlaneEndpoint.Port).To(BeEquivalentTo(6443))

				kubeconfig := &corev1.Secret{}
				g.Expect(s.Client.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "cluster-kubeconfig",
				}, kubeconfig)).To(Succeed())
				g.Expect(kubeconfig.Data).To(HaveKey("value"))

				kubeconfig = &corev1.Secret{}
				g.Expect(s.Client.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "cluster-user-kubeconfig",
				}, kubeconfig)).To(Succeed())
				g.Expect(kubeconfig.Data).To(HaveKey("value"))
			},
		},
		{
			name: "control-plane is already created and up-to-date",
			fields: fields{
				ManagedControlPlane: &scope.ManagedControlPlane{
					Cluster: &v1beta1.Cluster{
						Spec: v1beta1.ClusterSpec{},
					},
					ManagedCluster: &v1alpha1.ScalewayManagedCluster{
						ObjectMeta: v1.ObjectMeta{
							Name:      "managedcluster",
							Namespace: "default",
						},
						Status: v1alpha1.ScalewayManagedClusterStatus{
							Ready: true,
							Network: &v1alpha1.ManagedNetworkStatus{
								PrivateNetworkID: scw.StringPtr("11111111-1111-1111-1111-111111111111"),
							},
						},
					},
					ManagedControlPlane: &v1alpha1.ScalewayManagedControlPlane{
						ObjectMeta: v1.ObjectMeta{
							Name:      "controlplane",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayManagedControlPlaneSpec{
							Type:        "kapsule",
							Version:     "v1.31.1",
							CNI:         scw.StringPtr("cilium"),
							ClusterName: scw.StringPtr("default-controlplane"),
							ControlPlaneEndpoint: v1beta1.APIEndpoint{
								Host: fmt.Sprintf("%s.api.k8s.fr-par.scw.cloud", clusterID),
								Port: 6443,
							},
						},
						Status: v1alpha1.ScalewayManagedControlPlaneStatus{
							Ready:                       true,
							Initialized:                 true,
							ExternalManagedControlPlane: true,
							Version:                     scw.StringPtr("v1.31.1"),
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			objects: []runtime.Object{},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.FindCluster(gomock.Any(), "default-controlplane").Return(&k8s.Cluster{
					ID:         clusterID,
					Status:     k8s.ClusterStatusReady,
					Type:       "kapsule",
					Version:    "1.31.1",
					Tags:       []string{"caps-namespace=default", "caps-scalewaymanagedcontrolplane=controlplane", "created-by=cluster-api-provider-scaleway"},
					Cni:        k8s.CNICilium,
					ClusterURL: fmt.Sprintf("https://%s.api.k8s.fr-par.scw.cloud:6443", clusterID),
					AutoscalerConfig: &k8s.ClusterAutoscalerConfig{
						ScaleDownDisabled:             false,
						ScaleDownDelayAfterAdd:        "10m",
						Estimator:                     k8s.AutoscalerEstimatorBinpacking,
						Expander:                      k8s.AutoscalerExpanderRandom,
						IgnoreDaemonsetsUtilization:   false,
						BalanceSimilarNodeGroups:      false,
						ExpendablePodsPriorityCutoff:  -10,
						ScaleDownUnneededTime:         "10m",
						ScaleDownUtilizationThreshold: 0.5,
						MaxGracefulTerminationSec:     600,
					},
					AutoUpgrade: &k8s.ClusterAutoUpgrade{
						Enabled: false,
						MaintenanceWindow: &k8s.MaintenanceWindow{
							StartHour: 0,
							Day:       k8s.MaintenanceWindowDayOfTheWeekAny,
						},
					},
					OpenIDConnectConfig: &k8s.ClusterOpenIDConnectConfig{},
				}, nil)
				i.ListClusterACLRules(gomock.Any(), clusterID).Return([]*k8s.ACLRule{
					{IP: &scw.IPNet{IPNet: *ipNet0}},
				}, nil)
				i.GetClusterKubeConfig(gomock.Any(), clusterID).Return(&k8s.Kubeconfig{
					Clusters: []*k8s.KubeconfigClusterWithName{
						{
							Name: "default-controlplane",
							Cluster: k8s.KubeconfigCluster{
								CertificateAuthorityData: "fake",
							},
						},
					},
				}, nil)
				i.GetSecretKey().Return("secret-key")
			},
			asserts: func(g *WithT, s *scope.ManagedControlPlane) {},
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
				ManagedControlPlane: tt.fields.ManagedControlPlane,
			}
			s.Client = fake.NewFakeClient(tt.objects...)
			s.ScalewayClient = scwMock
			if err := s.Reconcile(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.asserts(g, s.ManagedControlPlane)
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	type fields struct {
		ManagedControlPlane *scope.ManagedControlPlane
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
			name: "delete cluster",
			fields: fields{
				ManagedControlPlane: &scope.ManagedControlPlane{
					ManagedControlPlane: &v1alpha1.ScalewayManagedControlPlane{
						ObjectMeta: v1.ObjectMeta{
							Name:      "controlplane",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayManagedControlPlaneSpec{
							Type:        "kapsule",
							Version:     "v1.31.1",
							CNI:         scw.StringPtr("cilium"),
							ClusterName: scw.StringPtr("default-controlplane"),
							ControlPlaneEndpoint: v1beta1.APIEndpoint{
								Host: fmt.Sprintf("%s.api.k8s.fr-par.scw.cloud", clusterID),
								Port: 6443,
							},
						},
						Status: v1alpha1.ScalewayManagedControlPlaneStatus{
							Ready:                       true,
							Initialized:                 true,
							ExternalManagedControlPlane: true,
							Version:                     scw.StringPtr("v1.31.1"),
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.FindCluster(gomock.Any(), "default-controlplane").Return(&k8s.Cluster{
					ID: clusterID,
				}, nil)
				i.DeleteCluster(gomock.Any(), clusterID, false)
			},
		},
		{
			name: "delete cluster with additional resources",
			fields: fields{
				ManagedControlPlane: &scope.ManagedControlPlane{
					ManagedControlPlane: &v1alpha1.ScalewayManagedControlPlane{
						ObjectMeta: v1.ObjectMeta{
							Name:      "controlplane",
							Namespace: "default",
						},
						Spec: v1alpha1.ScalewayManagedControlPlaneSpec{
							Type:        "kapsule",
							Version:     "v1.31.1",
							CNI:         scw.StringPtr("cilium"),
							ClusterName: scw.StringPtr("default-controlplane"),
							ControlPlaneEndpoint: v1beta1.APIEndpoint{
								Host: fmt.Sprintf("%s.api.k8s.fr-par.scw.cloud", clusterID),
								Port: 6443,
							},
							OnDelete: &v1alpha1.OnDeleteSpec{
								WithAdditionalResources: scw.BoolPtr(true),
							},
						},
						Status: v1alpha1.ScalewayManagedControlPlaneStatus{
							Ready:                       true,
							Initialized:                 true,
							ExternalManagedControlPlane: true,
							Version:                     scw.StringPtr("v1.31.1"),
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.FindCluster(gomock.Any(), "default-controlplane").Return(&k8s.Cluster{
					ID: clusterID,
				}, nil)
				i.DeleteCluster(gomock.Any(), clusterID, true).Return(nil)
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
				ManagedControlPlane: tt.fields.ManagedControlPlane,
			}
			s.ScalewayClient = scwMock
			if err := s.Delete(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
