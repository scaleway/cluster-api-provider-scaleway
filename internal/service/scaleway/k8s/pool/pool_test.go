package pool

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const (
	clusterID        = "11111111-1111-1111-1111-111111111111"
	poolID           = "11111111-1111-1111-1111-111111111111"
	placementGroupID = "11111111-1111-1111-1111-111111111111"
	securityGroupID  = "11111111-1111-1111-1111-111111111111"
)

func TestService_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		ManagedMachinePool *scope.ManagedMachinePool
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
		asserts func(g *WithT, s *scope.ManagedMachinePool)
	}{
		{
			name: "creating pool",
			fields: fields{
				ManagedMachinePool: &scope.ManagedMachinePool{
					ScalewayManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
						Spec: infrav1.ScalewayManagedControlPlaneSpec{
							ClusterName: "default-controlplane",
							Version:     "v1.30.0",
						},
					},
					MachinePool: &clusterv1.MachinePool{
						Spec: clusterv1.MachinePoolSpec{
							Replicas: scw.Int32Ptr(2),
							Template: clusterv1.MachineTemplateSpec{
								Spec: clusterv1.MachineSpec{
									Version: "v1.30.0",
								},
							},
						},
					},
					ScalewayManagedMachinePool: &infrav1.ScalewayManagedMachinePool{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pool",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayManagedMachinePoolSpec{
							Zone:             infrav1.ScalewayZone(scw.ZoneFrPar1),
							PlacementGroupID: placementGroupID,
							NodeType:         "DEV1-M",
							Scaling: infrav1.Scaling{
								Autoscaling: ptr.To(true),
								MinSize:     scw.Int32Ptr(1),
								MaxSize:     scw.Int32Ptr(5),
							},
							Autohealing: ptr.To(true),
							UpgradePolicy: infrav1.UpgradePolicy{
								MaxUnavailable: scw.Int32Ptr(0),
								MaxSurge:       scw.Int32Ptr(2),
							},
							RootVolumeType:   "sbs_15k",
							RootVolumeSizeGB: 42,
							PublicIPDisabled: ptr.To(true),
							SecurityGroupID:  securityGroupID,
							AdditionalTags:   []string{"tag1"},
							KubeletArgs: map[string]string{
								"containerLogMaxFiles": "500",
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.FindCluster(gomock.Any(), "default-controlplane").Return(&k8s.Cluster{
					ID:     clusterID,
					Status: k8s.ClusterStatusReady,
				}, nil)
				i.FindPool(gomock.Any(), clusterID, "pool").Return(nil, client.ErrNoItemFound)
				i.CreatePool(
					gomock.Any(),
					scw.Zone("fr-par-1"),
					clusterID,
					"pool",
					"DEV1-M",
					ptr.To(placementGroupID),
					ptr.To(securityGroupID),
					true,
					true,
					true,
					uint32(2),
					scw.Uint32Ptr(1),
					scw.Uint32Ptr(5),
					[]string{"caps-namespace=default", "caps-scalewaymanagedmachinepool=pool", "tag1"},
					map[string]string{
						"containerLogMaxFiles": "500",
					},
					k8s.PoolVolumeType("sbs_15k"),
					scw.Uint64Ptr(42),
					&k8s.CreatePoolRequestUpgradePolicy{
						MaxUnavailable: scw.Uint32Ptr(0),
						MaxSurge:       scw.Uint32Ptr(2),
					},
				).Return(&k8s.Pool{
					ID:               poolID,
					Status:           k8s.PoolStatusReady,
					Version:          "1.30.0",
					NodeType:         "DEV1-M",
					Autoscaling:      true,
					Autohealing:      true,
					PublicIPDisabled: true,
					Name:             "pool",
					Size:             2,
					MinSize:          1,
					MaxSize:          5,
					Tags:             []string{"caps-namespace=default", "caps-scalewaymanagedmachinepool=pool", "tag1", "created-by=cluster-api-provider-scaleway"},
					PlacementGroupID: ptr.To(placementGroupID),
					SecurityGroupID:  securityGroupID,
					KubeletArgs: map[string]string{
						"containerLogMaxFiles": "500",
					},
					UpgradePolicy: &k8s.PoolUpgradePolicy{
						MaxUnavailable: 0,
						MaxSurge:       2,
					},
					RootVolumeType: k8s.PoolVolumeTypeSbs15k,
					RootVolumeSize: ptr.To(42 * scw.GB),
				}, nil)
				i.ListNodes(gomock.Any(), clusterID, poolID).Return([]*k8s.Node{
					{
						ProviderID: "providerID1",
					},
					{
						ProviderID: "providerID2",
					},
				}, nil)
			},
			asserts: func(g *WithT, s *scope.ManagedMachinePool) {
				g.Expect(s.ScalewayManagedMachinePool.Spec.ProviderIDList).To(Equal([]string{
					"providerID1", "providerID2",
				}))
				g.Expect(s.ScalewayManagedMachinePool.Status.Replicas).NotTo(BeNil())
				g.Expect(*s.ScalewayManagedMachinePool.Status.Replicas).To(BeEquivalentTo(2))
			},
		},
		{
			name: "pool exists and is up-to-date",
			fields: fields{
				ManagedMachinePool: &scope.ManagedMachinePool{
					ScalewayManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
						Spec: infrav1.ScalewayManagedControlPlaneSpec{
							ClusterName: "default-controlplane",
							Version:     "v1.30.0",
						},
					},
					MachinePool: &clusterv1.MachinePool{
						Spec: clusterv1.MachinePoolSpec{
							Replicas: scw.Int32Ptr(2),
							Template: clusterv1.MachineTemplateSpec{
								Spec: clusterv1.MachineSpec{
									Version: "v1.30.0",
								},
							},
						},
					},
					ScalewayManagedMachinePool: &infrav1.ScalewayManagedMachinePool{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pool",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayManagedMachinePoolSpec{
							Zone:             infrav1.ScalewayZone(scw.ZoneFrPar1),
							PlacementGroupID: placementGroupID,
							NodeType:         "DEV1-M",
							Scaling: infrav1.Scaling{
								Autoscaling: ptr.To(true),
								MinSize:     scw.Int32Ptr(1),
								MaxSize:     scw.Int32Ptr(5),
							},
							Autohealing: ptr.To(true),
							UpgradePolicy: infrav1.UpgradePolicy{
								MaxUnavailable: scw.Int32Ptr(0),
								MaxSurge:       scw.Int32Ptr(2),
							},
							RootVolumeType:   "sbs_15k",
							RootVolumeSizeGB: 42,
							PublicIPDisabled: ptr.To(true),
							SecurityGroupID:  securityGroupID,
							AdditionalTags:   []string{"tag1"},
							KubeletArgs: map[string]string{
								"containerLogMaxFiles": "500",
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.FindCluster(gomock.Any(), "default-controlplane").Return(&k8s.Cluster{
					ID:     clusterID,
					Status: k8s.ClusterStatusReady,
				}, nil)
				i.FindPool(gomock.Any(), clusterID, "pool").Return(&k8s.Pool{
					ID:               poolID,
					Status:           k8s.PoolStatusReady,
					Version:          "1.30.0",
					NodeType:         "DEV1-M",
					Autoscaling:      true,
					Autohealing:      true,
					PublicIPDisabled: true,
					Name:             "pool",
					Size:             2,
					MinSize:          1,
					MaxSize:          5,
					Tags:             []string{"caps-namespace=default", "caps-scalewaymanagedmachinepool=pool", "tag1", "created-by=cluster-api-provider-scaleway"},
					PlacementGroupID: ptr.To(placementGroupID),
					SecurityGroupID:  securityGroupID,
					KubeletArgs: map[string]string{
						"containerLogMaxFiles": "500",
					},
					UpgradePolicy: &k8s.PoolUpgradePolicy{
						MaxUnavailable: 0,
						MaxSurge:       2,
					},
					RootVolumeType: k8s.PoolVolumeTypeSbs15k,
					RootVolumeSize: ptr.To(42 * scw.GB),
				}, nil)
				i.ListNodes(gomock.Any(), clusterID, poolID).Return([]*k8s.Node{
					{
						ProviderID: "providerID1",
					},
					{
						ProviderID: "providerID2",
					},
				}, nil)
			},
			asserts: func(g *WithT, s *scope.ManagedMachinePool) {
				g.Expect(s.ScalewayManagedMachinePool.Spec.ProviderIDList).To(Equal([]string{
					"providerID1", "providerID2",
				}))
				g.Expect(s.ScalewayManagedMachinePool.Status.Replicas).NotTo(BeNil())
				g.Expect(*s.ScalewayManagedMachinePool.Status.Replicas).To(BeEquivalentTo(2))
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
				ManagedMachinePool: tt.fields.ManagedMachinePool,
			}
			s.ManagedMachinePool.ScalewayClient = scwMock
			if err := s.Reconcile(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.asserts(g, s.ManagedMachinePool)
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	type fields struct {
		ManagedMachinePool *scope.ManagedMachinePool
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
		expect  func(i *mock_client.MockInterfaceMockRecorder)
	}{
		{
			name: "delete pool",
			fields: fields{
				ManagedMachinePool: &scope.ManagedMachinePool{
					ScalewayManagedControlPlane: &infrav1.ScalewayManagedControlPlane{
						Spec: infrav1.ScalewayManagedControlPlaneSpec{
							ClusterName: "default-controlplane",
							Version:     "v1.30.0",
						},
					},
					ScalewayManagedMachinePool: &infrav1.ScalewayManagedMachinePool{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pool",
							Namespace: "default",
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: scaleway.WithTransientError(errors.New("pool is being deleted"), poolRetryTime),
			expect: func(i *mock_client.MockInterfaceMockRecorder) {
				i.FindCluster(gomock.Any(), "default-controlplane").Return(&k8s.Cluster{
					ID: clusterID,
				}, nil)
				i.FindPool(gomock.Any(), clusterID, "pool").Return(&k8s.Pool{
					ID: poolID,
				}, nil)
				i.DeletePool(gomock.Any(), poolID).Return(nil)
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
				ManagedMachinePool: tt.fields.ManagedMachinePool,
			}
			s.ScalewayClient = scwMock
			err := s.Delete(tt.args.ctx)
			if (err == nil) != (tt.wantErr == nil) {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Service.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
