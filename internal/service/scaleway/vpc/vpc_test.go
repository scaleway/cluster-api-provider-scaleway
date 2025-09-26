package vpc

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const (
	privateNetworkID = "11111111-1111-1111-1111-111111111111"
	vpcID            = "22222222-2222-2222-2222-222222222222"
)

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
		asserts func(g *WithT, s Scope)
	}{
		{
			name: "no private network",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect:  func(i *mock_client.MockInterfaceMockRecorder) {},
			asserts: func(g *WithT, c Scope) {},
		},
		{
			name: "IDs already set in status",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						Spec: infrav1.ScalewayClusterSpec{
							Network: infrav1.ScalewayClusterNetwork{
								PrivateNetwork: infrav1.PrivateNetworkSpec{
									Enabled: ptr.To(true),
								},
							},
						},
						Status: infrav1.ScalewayClusterStatus{
							Network: infrav1.ScalewayClusterNetworkStatus{
								VPCID:            infrav1.UUID(vpcID),
								PrivateNetworkID: infrav1.UUID(privateNetworkID),
							},
							Conditions: []metav1.Condition{
								{
									Type:   infrav1.PrivateNetworkReadyCondition,
									Status: metav1.ConditionTrue,
									Reason: infrav1.ReadyReason,
								},
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect:  func(i *mock_client.MockInterfaceMockRecorder) {},
			asserts: func(g *WithT, s Scope) {},
		},
		{
			name: "managed private network",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: infrav1.ScalewayClusterNetwork{
								PrivateNetwork: infrav1.PrivateNetworkSpec{
									Enabled: ptr.To(true),
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
				tags := []string{
					"caps-namespace=default",
					"caps-scalewaycluster=cluster",
				}

				i.FindPrivateNetwork(gomock.Any(), tags, nil).Return(nil, client.ErrNoItemFound)
				i.CreatePrivateNetwork(gomock.Any(), "cluster", nil, nil, tags).Return(&vpc.PrivateNetwork{
					ID:          privateNetworkID,
					VpcID:       vpcID,
					DHCPEnabled: true,
				}, nil)
			},
			asserts: func(g *WithT, s Scope) {
				clusterScope, ok := s.(*scope.Cluster)
				g.Expect(ok).To(BeTrue())
				g.Expect(clusterScope.ScalewayCluster.Status.Network).NotTo(BeNil())
				g.Expect(clusterScope.ScalewayCluster.Status.Network.PrivateNetworkID).To(BeEquivalentTo(privateNetworkID))
				g.Expect(clusterScope.ScalewayCluster.Status.Network.VPCID).To(BeEquivalentTo(vpcID))
			},
		},
		{
			name: "existing private network",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: infrav1.ScalewayClusterNetwork{
								PrivateNetwork: infrav1.PrivateNetworkSpec{
									Enabled: ptr.To(true),
									PrivateNetwork: infrav1.PrivateNetwork{
										ID: infrav1.UUID(privateNetworkID),
									},
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
				i.GetPrivateNetwork(gomock.Any(), privateNetworkID).Return(&vpc.PrivateNetwork{
					ID:    privateNetworkID,
					VpcID: vpcID,
				}, nil)
			},
			asserts: func(g *WithT, s Scope) {
				clusterScope, ok := s.(*scope.Cluster)
				g.Expect(ok).To(BeTrue())

				g.Expect(clusterScope.ScalewayCluster.Status.Network).NotTo(BeNil())
				g.Expect(clusterScope.ScalewayCluster.Status.Network.PrivateNetworkID).To(BeEquivalentTo(privateNetworkID))
				g.Expect(clusterScope.ScalewayCluster.Status.Network.VPCID).To(BeEquivalentTo(vpcID))
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
				Scope: tt.fields.Scope,
			}
			s.SetCloud(scwMock)
			if err := s.Reconcile(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.asserts(g, s.Scope)
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
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {},
		},
		{
			name: "find and delete",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: infrav1.ScalewayClusterNetwork{
								PrivateNetwork: infrav1.PrivateNetworkSpec{
									Enabled: ptr.To(true),
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
				tags := []string{
					"caps-namespace=default",
					"caps-scalewaycluster=cluster",
				}

				i.FindPrivateNetwork(gomock.Any(), tags, nil).Return(&vpc.PrivateNetwork{
					ID: privateNetworkID,
				}, nil)
				i.CleanAvailableIPs(gomock.Any(), privateNetworkID)
				i.DeletePrivateNetwork(gomock.Any(), privateNetworkID)
			},
		},
		{
			name: "do not remove user-provided private network",
			fields: fields{
				Scope: &scope.Cluster{
					ScalewayCluster: &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster",
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							Network: infrav1.ScalewayClusterNetwork{
								PrivateNetwork: infrav1.PrivateNetworkSpec{
									Enabled: ptr.To(true),
									PrivateNetwork: infrav1.PrivateNetwork{
										ID: infrav1.UUID(privateNetworkID),
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			expect: func(i *mock_client.MockInterfaceMockRecorder) {},
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
