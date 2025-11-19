package client

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client/mock_client"
)

const (
	clusterID = "11111111-1111-1111-1111-111111111111"
	poolID    = "11111111-1111-1111-1111-111111111111"
	aclID1    = "11111111-1111-1111-1111-111111111111"
	aclID2    = "22222222-1111-1111-1111-111111111111"
)

func TestClient_FindCluster(t *testing.T) {
	t.Parallel()
	type fields struct {
		projectID string
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *k8s.Cluster
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "cluster found",
			fields: fields{
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				name: "mycluster",
			},
			want: &k8s.Cluster{
				ID:   clusterID,
				Name: "mycluster",
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.ListClusters(&k8s.ListClustersRequest{
					ProjectID: ptr.To(projectID),
					Name:      ptr.To("mycluster"),
				}, gomock.Any()).Return(&k8s.ListClustersResponse{
					TotalCount: 1,
					Clusters: []*k8s.Cluster{
						{
							ID:   clusterID,
							Name: "mycluster",
						},
					},
				}, nil)
			},
		},
		{
			name: "no cluster found",
			fields: fields{
				projectID: projectID,
			},
			args: args{
				ctx:  context.TODO(),
				name: "mycluster",
			},
			wantErr: true,
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.ListClusters(&k8s.ListClustersRequest{
					ProjectID: ptr.To(projectID),
					Name:      ptr.To("mycluster"),
				}, gomock.Any(), gomock.Any()).Return(&k8s.ListClustersResponse{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				projectID: tt.fields.projectID,
				k8s:       k8sMock,
			}
			got, err := c.FindCluster(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreateCluster(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                 context.Context
		name                string
		clusterType         string
		version             string
		pnID                *string
		tags                []string
		featureGates        []string
		admissionPlugins    []string
		apiServerCertSANs   []string
		cni                 k8s.CNI
		autoscalerConfig    *k8s.CreateClusterRequestAutoscalerConfig
		autoUpgrade         *k8s.CreateClusterRequestAutoUpgrade
		openIDConnectConfig *k8s.CreateClusterRequestOpenIDConnectConfig
		podCIDR             scw.IPNet
		serviceCIDR         scw.IPNet
	}
	tests := []struct {
		name    string
		args    args
		want    *k8s.Cluster
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "create cluster",
			args: args{
				ctx:               context.TODO(),
				name:              "test",
				clusterType:       "kapsule",
				version:           "1.30.4",
				pnID:              ptr.To(privateNetworkID),
				tags:              []string{"tag1", "tag2"},
				featureGates:      []string{"HPAScaleToZero"},
				admissionPlugins:  []string{"AlwaysPullImages"},
				apiServerCertSANs: []string{"my-cluster.test"},
				cni:               k8s.CNICilium,
				autoscalerConfig: &k8s.CreateClusterRequestAutoscalerConfig{
					ScaleDownDisabled:             ptr.To(false),
					ScaleDownDelayAfterAdd:        ptr.To("1m"),
					Estimator:                     k8s.AutoscalerEstimatorBinpacking,
					Expander:                      k8s.AutoscalerExpanderMostPods,
					IgnoreDaemonsetsUtilization:   ptr.To(true),
					BalanceSimilarNodeGroups:      ptr.To(true),
					ExpendablePodsPriorityCutoff:  scw.Int32Ptr(1),
					ScaleDownUnneededTime:         ptr.To("1m"),
					ScaleDownUtilizationThreshold: scw.Float32Ptr(1),
					MaxGracefulTerminationSec:     scw.Uint32Ptr(30),
				},
				autoUpgrade: &k8s.CreateClusterRequestAutoUpgrade{
					Enable: true,
					MaintenanceWindow: &k8s.MaintenanceWindow{
						StartHour: 1,
						Day:       k8s.MaintenanceWindowDayOfTheWeekFriday,
					},
				},
				openIDConnectConfig: &k8s.CreateClusterRequestOpenIDConnectConfig{
					IssuerURL:      "http://oidcprovider.test",
					ClientID:       "abcd",
					UsernameClaim:  ptr.To("username"),
					UsernamePrefix: ptr.To("usernameprefix"),
					GroupsClaim:    &[]string{"groups"},
					GroupsPrefix:   ptr.To("groupsprefix"),
					RequiredClaim:  &[]string{"verified"},
				},
				podCIDR:     scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(192, 168, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)}},
				serviceCIDR: scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)}},
			},
			want: &k8s.Cluster{
				ID:   clusterID,
				Name: "test",
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.CreateCluster(&k8s.CreateClusterRequest{
					Type:        "kapsule",
					Name:        "test",
					Description: createdByDescription,
					Tags:        []string{"tag1", "tag2", createdByTag},
					Version:     "1.30.4",
					Cni:         k8s.CNICilium,
					AutoscalerConfig: &k8s.CreateClusterRequestAutoscalerConfig{
						ScaleDownDisabled:             ptr.To(false),
						ScaleDownDelayAfterAdd:        ptr.To("1m"),
						Estimator:                     k8s.AutoscalerEstimatorBinpacking,
						Expander:                      k8s.AutoscalerExpanderMostPods,
						IgnoreDaemonsetsUtilization:   ptr.To(true),
						BalanceSimilarNodeGroups:      ptr.To(true),
						ExpendablePodsPriorityCutoff:  scw.Int32Ptr(1),
						ScaleDownUnneededTime:         ptr.To("1m"),
						ScaleDownUtilizationThreshold: scw.Float32Ptr(1),
						MaxGracefulTerminationSec:     scw.Uint32Ptr(30),
					},
					AutoUpgrade: &k8s.CreateClusterRequestAutoUpgrade{
						Enable: true,
						MaintenanceWindow: &k8s.MaintenanceWindow{
							StartHour: 1,
							Day:       k8s.MaintenanceWindowDayOfTheWeekFriday,
						},
					},
					FeatureGates:     []string{"HPAScaleToZero"},
					AdmissionPlugins: []string{"AlwaysPullImages"},
					OpenIDConnectConfig: &k8s.CreateClusterRequestOpenIDConnectConfig{
						IssuerURL:      "http://oidcprovider.test",
						ClientID:       "abcd",
						UsernameClaim:  ptr.To("username"),
						UsernamePrefix: ptr.To("usernameprefix"),
						GroupsClaim:    &[]string{"groups"},
						GroupsPrefix:   ptr.To("groupsprefix"),
						RequiredClaim:  &[]string{"verified"},
					},
					ApiserverCertSans: []string{"my-cluster.test"},
					PrivateNetworkID:  ptr.To(privateNetworkID),
					PodCidr:           &scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(192, 168, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)}},
					ServiceCidr:       &scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)}},
				}, gomock.Any()).Return(&k8s.Cluster{
					ID:   clusterID,
					Name: "test",
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			got, err := c.CreateCluster(tt.args.ctx, tt.args.name, tt.args.clusterType, tt.args.version, tt.args.pnID, tt.args.tags, tt.args.featureGates, tt.args.admissionPlugins, tt.args.apiServerCertSANs, tt.args.cni, tt.args.autoscalerConfig, tt.args.autoUpgrade, tt.args.openIDConnectConfig, tt.args.podCIDR, tt.args.serviceCIDR)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteCluster(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                     context.Context
		id                      string
		withAdditionalResources bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "delete cluster",
			args: args{
				ctx:                     context.TODO(),
				id:                      clusterID,
				withAdditionalResources: true,
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.DeleteCluster(&k8s.DeleteClusterRequest{
					ClusterID:               clusterID,
					WithAdditionalResources: true,
				}, gomock.Any()).Return(&k8s.Cluster{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.DeleteCluster(tt.args.ctx, tt.args.id, tt.args.withAdditionalResources); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_GetClusterKubeConfig(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *k8s.Kubeconfig
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "get cluster kubeconfig",
			args: args{
				ctx: context.TODO(),
				id:  clusterID,
			},
			want: &k8s.Kubeconfig{Kind: "Config"},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.GetClusterKubeConfig(&k8s.GetClusterKubeConfigRequest{
					ClusterID: clusterID,
				}, gomock.Any()).Return(&k8s.Kubeconfig{
					Kind: "Config",
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			got, err := c.GetClusterKubeConfig(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetClusterKubeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetClusterKubeConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_UpdateCluster(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                 context.Context
		id                  string
		tags                *[]string
		featureGates        *[]string
		admissionPlugins    *[]string
		apiServerCertSANs   *[]string
		autoscalerConfig    *k8s.UpdateClusterRequestAutoscalerConfig
		autoUpgrade         *k8s.UpdateClusterRequestAutoUpgrade
		openIDConnectConfig *k8s.UpdateClusterRequestOpenIDConnectConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "update cluster",
			args: args{
				ctx:               context.TODO(),
				id:                clusterID,
				tags:              &[]string{"tag1", "tag2"},
				featureGates:      &[]string{"HPAScaleToZero"},
				admissionPlugins:  &[]string{"AlwaysPullImages"},
				apiServerCertSANs: &[]string{"mycluster.test"},
				autoscalerConfig: &k8s.UpdateClusterRequestAutoscalerConfig{
					ScaleDownDisabled:             ptr.To(false),
					ScaleDownDelayAfterAdd:        ptr.To("1m"),
					Estimator:                     k8s.AutoscalerEstimatorBinpacking,
					Expander:                      k8s.AutoscalerExpanderMostPods,
					IgnoreDaemonsetsUtilization:   ptr.To(true),
					BalanceSimilarNodeGroups:      ptr.To(true),
					ExpendablePodsPriorityCutoff:  scw.Int32Ptr(1),
					ScaleDownUnneededTime:         ptr.To("1m"),
					ScaleDownUtilizationThreshold: scw.Float32Ptr(1),
					MaxGracefulTerminationSec:     scw.Uint32Ptr(30),
				},
				autoUpgrade: &k8s.UpdateClusterRequestAutoUpgrade{
					Enable: ptr.To(true),
					MaintenanceWindow: &k8s.MaintenanceWindow{
						StartHour: 1,
						Day:       k8s.MaintenanceWindowDayOfTheWeekFriday,
					},
				},
				openIDConnectConfig: &k8s.UpdateClusterRequestOpenIDConnectConfig{
					IssuerURL:      ptr.To("http://oidcprovider.test"),
					ClientID:       ptr.To("abcd"),
					UsernameClaim:  ptr.To("username"),
					UsernamePrefix: ptr.To("usernameprefix"),
					GroupsClaim:    &[]string{"groups"},
					GroupsPrefix:   ptr.To("groupsprefix"),
					RequiredClaim:  &[]string{"verified"},
				},
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.UpdateCluster(&k8s.UpdateClusterRequest{
					ClusterID: clusterID,
					Tags:      &[]string{"tag1", "tag2", createdByTag},
					AutoscalerConfig: &k8s.UpdateClusterRequestAutoscalerConfig{
						ScaleDownDisabled:             ptr.To(false),
						ScaleDownDelayAfterAdd:        ptr.To("1m"),
						Estimator:                     k8s.AutoscalerEstimatorBinpacking,
						Expander:                      k8s.AutoscalerExpanderMostPods,
						IgnoreDaemonsetsUtilization:   ptr.To(true),
						BalanceSimilarNodeGroups:      ptr.To(true),
						ExpendablePodsPriorityCutoff:  scw.Int32Ptr(1),
						ScaleDownUnneededTime:         ptr.To("1m"),
						ScaleDownUtilizationThreshold: scw.Float32Ptr(1),
						MaxGracefulTerminationSec:     scw.Uint32Ptr(30),
					},
					AutoUpgrade: &k8s.UpdateClusterRequestAutoUpgrade{
						Enable: ptr.To(true),
						MaintenanceWindow: &k8s.MaintenanceWindow{
							StartHour: 1,
							Day:       k8s.MaintenanceWindowDayOfTheWeekFriday,
						},
					},
					FeatureGates:      &[]string{"HPAScaleToZero"},
					AdmissionPlugins:  &[]string{"AlwaysPullImages"},
					ApiserverCertSans: &[]string{"mycluster.test"},
					OpenIDConnectConfig: &k8s.UpdateClusterRequestOpenIDConnectConfig{
						IssuerURL:      ptr.To("http://oidcprovider.test"),
						ClientID:       ptr.To("abcd"),
						UsernameClaim:  ptr.To("username"),
						UsernamePrefix: ptr.To("usernameprefix"),
						GroupsClaim:    &[]string{"groups"},
						GroupsPrefix:   ptr.To("groupsprefix"),
						RequiredClaim:  &[]string{"verified"},
					},
				}, gomock.Any()).Return(&k8s.Cluster{
					ID: clusterID,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.UpdateCluster(tt.args.ctx, tt.args.id, tt.args.tags, tt.args.featureGates, tt.args.admissionPlugins, tt.args.apiServerCertSANs, tt.args.autoscalerConfig, tt.args.autoUpgrade, tt.args.openIDConnectConfig); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_UpgradeCluster(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx     context.Context
		id      string
		version string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "upgrade cluster",
			args: args{
				ctx:     context.TODO(),
				id:      clusterID,
				version: "1.31.5",
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.UpgradeCluster(&k8s.UpgradeClusterRequest{
					ClusterID:    clusterID,
					Version:      "1.31.5",
					UpgradePools: false,
				}, gomock.Any()).Return(&k8s.Cluster{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.UpgradeCluster(tt.args.ctx, tt.args.id, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpgradeCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_SetClusterType(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx         context.Context
		id          string
		clusterType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "set cluster type",
			args: args{
				ctx:         context.TODO(),
				id:          clusterID,
				clusterType: "kapsule-dedicated-4",
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.SetClusterType(&k8s.SetClusterTypeRequest{
					ClusterID: clusterID,
					Type:      "kapsule-dedicated-4",
				}, gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.SetClusterType(tt.args.ctx, tt.args.id, tt.args.clusterType); (err != nil) != tt.wantErr {
				t.Errorf("Client.SetClusterType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_FindPool(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx       context.Context
		clusterID string
		name      string
	}
	tests := []struct {
		name    string
		args    args
		want    *k8s.Pool
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "found pool",
			args: args{
				ctx:       context.TODO(),
				clusterID: clusterID,
				name:      "mypool",
			},
			want: &k8s.Pool{
				ID:   poolID,
				Name: "mypool",
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.ListPools(&k8s.ListPoolsRequest{
					ClusterID: clusterID,
					Name:      ptr.To("mypool"),
				}, gomock.Any(), gomock.Any()).Return(&k8s.ListPoolsResponse{
					TotalCount: 1,
					Pools: []*k8s.Pool{
						{
							ID:   poolID,
							Name: "mypool",
						},
					},
				}, nil)
			},
		},
		{
			name: "no pool found",
			args: args{
				ctx:       context.TODO(),
				clusterID: clusterID,
				name:      "mypool",
			},
			wantErr: true,
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.ListPools(&k8s.ListPoolsRequest{
					ClusterID: clusterID,
					Name:      ptr.To("mypool"),
				}, gomock.Any(), gomock.Any()).Return(&k8s.ListPoolsResponse{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			got, err := c.FindPool(tt.args.ctx, tt.args.clusterID, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FindPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FindPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CreatePool(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx              context.Context
		zone             scw.Zone
		clusterID        string
		name             string
		nodeType         string
		placementGroupID *string
		securityGroupID  *string
		autoscaling      bool
		autohealing      bool
		publicIPDisabled bool
		size             uint32
		minSize          *uint32
		maxSize          *uint32
		tags             []string
		kubeletArgs      map[string]string
		rootVolumeType   k8s.PoolVolumeType
		rootVolumeSizeGB *uint64
		upgradePolicy    *k8s.CreatePoolRequestUpgradePolicy
	}
	tests := []struct {
		name    string
		args    args
		want    *k8s.Pool
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "create pool",
			args: args{
				ctx:              context.TODO(),
				zone:             scw.ZoneFrPar1,
				clusterID:        clusterID,
				name:             "mypool",
				nodeType:         "DEV1-S",
				placementGroupID: ptr.To(placementGroupID),
				securityGroupID:  ptr.To(securityGroupID),
				autoscaling:      true,
				autohealing:      true,
				publicIPDisabled: true,
				size:             1,
				minSize:          scw.Uint32Ptr(1),
				maxSize:          scw.Uint32Ptr(5),
				tags:             []string{"tag1", "tag2"},
				kubeletArgs: map[string]string{
					"containerLogMaxFiles":  "100",
					"maxParallelImagePulls": "5",
				},
				rootVolumeType:   k8s.PoolVolumeTypeBSSD,
				rootVolumeSizeGB: scw.Uint64Ptr(30),
				upgradePolicy: &k8s.CreatePoolRequestUpgradePolicy{
					MaxUnavailable: scw.Uint32Ptr(0),
					MaxSurge:       scw.Uint32Ptr(1),
				},
			},
			want: &k8s.Pool{
				ID:   poolID,
				Name: "mypool",
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.CreatePool(&k8s.CreatePoolRequest{
					ClusterID:        clusterID,
					Name:             "mypool",
					NodeType:         "DEV1-S",
					PlacementGroupID: ptr.To(placementGroupID),
					Autoscaling:      true,
					Size:             1,
					MinSize:          scw.Uint32Ptr(1),
					MaxSize:          scw.Uint32Ptr(5),
					Autohealing:      true,
					Tags:             []string{"tag1", "tag2", createdByTag},
					KubeletArgs: map[string]string{
						"containerLogMaxFiles":  "100",
						"maxParallelImagePulls": "5",
					},
					UpgradePolicy: &k8s.CreatePoolRequestUpgradePolicy{
						MaxUnavailable: scw.Uint32Ptr(0),
						MaxSurge:       scw.Uint32Ptr(1),
					},
					Zone:             scw.ZoneFrPar1,
					RootVolumeType:   k8s.PoolVolumeTypeBSSD,
					RootVolumeSize:   ptr.To(30 * scw.GB),
					PublicIPDisabled: true,
					SecurityGroupID:  ptr.To(securityGroupID),
				}, gomock.Any()).Return(&k8s.Pool{
					ID:   poolID,
					Name: "mypool",
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			got, err := c.CreatePool(tt.args.ctx, tt.args.zone, tt.args.clusterID, tt.args.name, tt.args.nodeType, tt.args.placementGroupID, tt.args.securityGroupID, tt.args.autoscaling, tt.args.autohealing, tt.args.publicIPDisabled, tt.args.size, tt.args.minSize, tt.args.maxSize, tt.args.tags, tt.args.kubeletArgs, tt.args.rootVolumeType, tt.args.rootVolumeSizeGB, tt.args.upgradePolicy)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreatePool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreatePool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_UpdatePool(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx           context.Context
		id            string
		autoscaling   *bool
		autohealing   *bool
		size          *uint32
		minSize       *uint32
		maxSize       *uint32
		tags          *[]string
		kubeletArgs   *map[string]string
		upgradePolicy *k8s.UpdatePoolRequestUpgradePolicy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "update pool",
			args: args{
				ctx:         context.TODO(),
				id:          poolID,
				autoscaling: ptr.To(true),
				autohealing: ptr.To(true),
				size:        scw.Uint32Ptr(1),
				minSize:     scw.Uint32Ptr(1),
				maxSize:     scw.Uint32Ptr(5),
				tags:        &[]string{"tag1", "tag2"},
				kubeletArgs: &map[string]string{
					"containerLogMaxFiles":  "100",
					"maxParallelImagePulls": "5",
				},
				upgradePolicy: &k8s.UpdatePoolRequestUpgradePolicy{
					MaxUnavailable: scw.Uint32Ptr(0),
					MaxSurge:       scw.Uint32Ptr(1),
				},
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.UpdatePool(&k8s.UpdatePoolRequest{
					PoolID:      poolID,
					Autoscaling: ptr.To(true),
					Autohealing: ptr.To(true),
					Size:        scw.Uint32Ptr(1),
					MinSize:     scw.Uint32Ptr(1),
					MaxSize:     scw.Uint32Ptr(5),
					Tags:        &[]string{"tag1", "tag2", createdByTag},
					KubeletArgs: &map[string]string{
						"containerLogMaxFiles":  "100",
						"maxParallelImagePulls": "5",
					},
					UpgradePolicy: &k8s.UpdatePoolRequestUpgradePolicy{
						MaxUnavailable: scw.Uint32Ptr(0),
						MaxSurge:       scw.Uint32Ptr(1),
					},
				}, gomock.Any()).Return(&k8s.Pool{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.UpdatePool(tt.args.ctx, tt.args.id, tt.args.autoscaling, tt.args.autohealing, tt.args.size, tt.args.minSize, tt.args.maxSize, tt.args.tags, tt.args.kubeletArgs, tt.args.upgradePolicy); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdatePool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_UpgradePool(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx     context.Context
		id      string
		version string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "upgrade pool",
			args: args{
				ctx:     context.TODO(),
				id:      poolID,
				version: "1.31.1",
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.UpgradePool(&k8s.UpgradePoolRequest{
					PoolID:  poolID,
					Version: "1.31.1",
				}, gomock.Any()).Return(&k8s.Pool{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.UpgradePool(tt.args.ctx, tt.args.id, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("Client.UpgradePool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeletePool(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "delete pool",
			args: args{
				ctx: context.TODO(),
				id:  poolID,
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.DeletePool(&k8s.DeletePoolRequest{
					PoolID: poolID,
				}, gomock.Any()).Return(&k8s.Pool{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.DeletePool(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeletePool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ListNodes(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx       context.Context
		clusterID string
		poolID    string
	}
	tests := []struct {
		name    string
		args    args
		want    []*k8s.Node
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "list nodes",
			args: args{
				ctx:       context.TODO(),
				clusterID: clusterID,
				poolID:    poolID,
			},
			want: []*k8s.Node{
				{
					Name: "node1",
				},
				{
					Name: "node2",
				},
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.ListNodes(&k8s.ListNodesRequest{
					ClusterID: clusterID,
					PoolID:    ptr.To(poolID),
				}, gomock.Any(), gomock.Any()).Return(&k8s.ListNodesResponse{
					TotalCount: 2,
					Nodes: []*k8s.Node{
						{
							Name: "node1",
						},
						{
							Name: "node2",
						},
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

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			got, err := c.ListNodes(tt.args.ctx, tt.args.clusterID, tt.args.poolID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListClusterACLRules(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx       context.Context
		clusterID string
	}
	tests := []struct {
		name    string
		args    args
		want    []*k8s.ACLRule
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "list cluster acls",
			args: args{
				ctx:       context.TODO(),
				clusterID: clusterID,
			},
			want: []*k8s.ACLRule{
				{
					ID: aclID1,
				},
				{
					ID: aclID2,
				},
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.ListClusterACLRules(&k8s.ListClusterACLRulesRequest{
					ClusterID: clusterID,
				}, gomock.Any(), gomock.Any()).Return(&k8s.ListClusterACLRulesResponse{
					TotalCount: 2,
					Rules: []*k8s.ACLRule{
						{
							ID: aclID1,
						},
						{
							ID: aclID2,
						},
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

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			got, err := c.ListClusterACLRules(tt.args.ctx, tt.args.clusterID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListClusterACLRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListClusterACLRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SetClusterACLRules(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx       context.Context
		clusterID string
		rules     []*k8s.ACLRuleRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  func(d *mock_client.MockK8sAPIMockRecorder)
	}{
		{
			name: "set cluster acls",
			args: args{
				ctx:       context.TODO(),
				clusterID: clusterID,
				rules: []*k8s.ACLRuleRequest{
					{
						ScalewayRanges: ptr.To(true),
					},
					{
						IP: &scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)}},
					},
				},
			},
			expect: func(d *mock_client.MockK8sAPIMockRecorder) {
				d.SetClusterACLRules(&k8s.SetClusterACLRulesRequest{
					ClusterID: clusterID,
					ACLs: []*k8s.ACLRuleRequest{
						{
							ScalewayRanges: ptr.To(true),
							Description:    createdByDescription,
						},
						{
							IP:          &scw.IPNet{IPNet: net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)}},
							Description: createdByDescription,
						},
					},
				}, gomock.Any()).Return(&k8s.SetClusterACLRulesResponse{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			k8sMock := mock_client.NewMockK8sAPI(mockCtrl)

			tt.expect(k8sMock.EXPECT())

			c := &Client{
				k8s: k8sMock,
			}
			if err := c.SetClusterACLRules(tt.args.ctx, tt.args.clusterID, tt.args.rules); (err != nil) != tt.wantErr {
				t.Errorf("Client.SetClusterACLRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
