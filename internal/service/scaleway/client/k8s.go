package client

import (
	"context"
	"fmt"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type K8sAPI interface {
	ListClusters(req *k8s.ListClustersRequest, opts ...scw.RequestOption) (*k8s.ListClustersResponse, error)
	CreateCluster(req *k8s.CreateClusterRequest, opts ...scw.RequestOption) (*k8s.Cluster, error)
	DeleteCluster(req *k8s.DeleteClusterRequest, opts ...scw.RequestOption) (*k8s.Cluster, error)
	GetClusterKubeConfig(req *k8s.GetClusterKubeConfigRequest, opts ...scw.RequestOption) (*k8s.Kubeconfig, error)
	UpdateCluster(req *k8s.UpdateClusterRequest, opts ...scw.RequestOption) (*k8s.Cluster, error)
	UpgradeCluster(req *k8s.UpgradeClusterRequest, opts ...scw.RequestOption) (*k8s.Cluster, error)
	SetClusterType(req *k8s.SetClusterTypeRequest, opts ...scw.RequestOption) (*k8s.Cluster, error)
	ListPools(req *k8s.ListPoolsRequest, opts ...scw.RequestOption) (*k8s.ListPoolsResponse, error)
	CreatePool(req *k8s.CreatePoolRequest, opts ...scw.RequestOption) (*k8s.Pool, error)
	UpdatePool(req *k8s.UpdatePoolRequest, opts ...scw.RequestOption) (*k8s.Pool, error)
	UpgradePool(req *k8s.UpgradePoolRequest, opts ...scw.RequestOption) (*k8s.Pool, error)
	DeletePool(req *k8s.DeletePoolRequest, opts ...scw.RequestOption) (*k8s.Pool, error)
	ListNodes(req *k8s.ListNodesRequest, opts ...scw.RequestOption) (*k8s.ListNodesResponse, error)
	ListClusterACLRules(req *k8s.ListClusterACLRulesRequest, opts ...scw.RequestOption) (*k8s.ListClusterACLRulesResponse, error)
	SetClusterACLRules(req *k8s.SetClusterACLRulesRequest, opts ...scw.RequestOption) (*k8s.SetClusterACLRulesResponse, error)
}

type K8s interface {
	FindCluster(ctx context.Context, name string) (*k8s.Cluster, error)
	CreateCluster(
		ctx context.Context,
		name, clusterType, version string,
		pnID *string,
		tags, featureGates, admissionPlugins, apiServerCertSANs []string,
		cni k8s.CNI,
		autoscalerConfig *k8s.CreateClusterRequestAutoscalerConfig,
		autoUpgrade *k8s.CreateClusterRequestAutoUpgrade,
		openIDConnectConfig *k8s.CreateClusterRequestOpenIDConnectConfig,
		podCIDR, serviceCIDR scw.IPNet,
	) (*k8s.Cluster, error)
	DeleteCluster(ctx context.Context, id string, withAdditionalResources bool) error
	GetClusterKubeConfig(ctx context.Context, id string) (*k8s.Kubeconfig, error)
	UpdateCluster(
		ctx context.Context,
		id string,
		tags, featureGates, admissionPlugins, apiServerCertSANs *[]string,
		autoscalerConfig *k8s.UpdateClusterRequestAutoscalerConfig,
		autoUpgrade *k8s.UpdateClusterRequestAutoUpgrade,
		openIDConnectConfig *k8s.UpdateClusterRequestOpenIDConnectConfig,
	) error
	UpgradeCluster(ctx context.Context, id, version string) error
	SetClusterType(ctx context.Context, id, clusterType string) error
	FindPool(ctx context.Context, clusterID, name string) (*k8s.Pool, error)
	CreatePool(
		ctx context.Context,
		zone scw.Zone,
		clusterID, name, nodeType string,
		placementGroupID, securityGroupID *string,
		autoscaling, autohealing, publicIPDisabled bool,
		size uint32,
		minSize, maxSize *uint32,
		tags []string,
		kubeletArgs map[string]string,
		rootVolumeType k8s.PoolVolumeType,
		rootVolumeSizeGB *uint64,
		upgradePolicy *k8s.CreatePoolRequestUpgradePolicy,
	) (*k8s.Pool, error)
	UpdatePool(
		ctx context.Context,
		id string,
		autoscaling, autohealing *bool,
		size, minSize, maxSize *uint32,
		tags *[]string,
		kubeletArgs *map[string]string,
		upgradePolicy *k8s.UpdatePoolRequestUpgradePolicy,
	) error
	UpgradePool(ctx context.Context, id, version string) error
	DeletePool(ctx context.Context, id string) error
	ListNodes(ctx context.Context, clusterID, poolID string) ([]*k8s.Node, error)
	ListClusterACLRules(ctx context.Context, clusterID string) ([]*k8s.ACLRule, error)
	SetClusterACLRules(ctx context.Context, clusterID string, rules []*k8s.ACLRuleRequest) error
}

func (c *Client) FindCluster(ctx context.Context, name string) (*k8s.Cluster, error) {
	resp, err := c.k8s.ListClusters(&k8s.ListClustersRequest{
		ProjectID: &c.projectID,
		Name:      &name,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListClusters", err)
	}

	// Filter out all clusters that have the wrong name.
	clusters := slices.DeleteFunc(resp.Clusters, func(cluster *k8s.Cluster) bool {
		return cluster.Name != name
	})

	switch len(clusters) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return clusters[0], nil
	default:
		// This case should never happen as k8s API prevents the creation of
		// multiple clusters with the same name.
		return nil, fmt.Errorf("%w: found %d clusters with name %s", ErrTooManyItemsFound, len(clusters), name)
	}
}

func (c *Client) CreateCluster(
	ctx context.Context,
	name, clusterType, version string,
	pnID *string,
	tags, featureGates, admissionPlugins, apiServerCertSANs []string,
	cni k8s.CNI,
	autoscalerConfig *k8s.CreateClusterRequestAutoscalerConfig,
	autoUpgrade *k8s.CreateClusterRequestAutoUpgrade,
	openIDConnectConfig *k8s.CreateClusterRequestOpenIDConnectConfig,
	podCIDR, serviceCIDR scw.IPNet,
) (*k8s.Cluster, error) {
	cluster, err := c.k8s.CreateCluster(&k8s.CreateClusterRequest{
		Name:                name,
		Type:                clusterType,
		Description:         createdByDescription,
		Tags:                append(tags, createdByTag),
		Version:             version,
		Cni:                 cni,
		PrivateNetworkID:    pnID,
		AutoscalerConfig:    autoscalerConfig,
		AutoUpgrade:         autoUpgrade,
		FeatureGates:        featureGates,
		AdmissionPlugins:    admissionPlugins,
		OpenIDConnectConfig: openIDConnectConfig,
		ApiserverCertSans:   apiServerCertSANs,
		PodCidr:             &podCIDR,
		ServiceCidr:         &serviceCIDR,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreateCluster", err)
	}

	return cluster, nil
}

func (c *Client) DeleteCluster(ctx context.Context, id string, withAdditionalResources bool) error {
	if _, err := c.k8s.DeleteCluster(&k8s.DeleteClusterRequest{
		ClusterID:               id,
		WithAdditionalResources: withAdditionalResources,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeleteCluster", err)
	}

	return nil
}

func (c *Client) GetClusterKubeConfig(ctx context.Context, id string) (*k8s.Kubeconfig, error) {
	kubeconfig, err := c.k8s.GetClusterKubeConfig(&k8s.GetClusterKubeConfigRequest{
		ClusterID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("GetClusterKubeConfig", err)
	}

	return kubeconfig, nil
}

func (c *Client) UpdateCluster(
	ctx context.Context,
	id string,
	tags, featureGates, admissionPlugins, apiServerCertSANs *[]string,
	autoscalerConfig *k8s.UpdateClusterRequestAutoscalerConfig,
	autoUpgrade *k8s.UpdateClusterRequestAutoUpgrade,
	openIDConnectConfig *k8s.UpdateClusterRequestOpenIDConnectConfig,
) error {
	if tags != nil {
		*tags = append(*tags, createdByTag)
	}

	if _, err := c.k8s.UpdateCluster(&k8s.UpdateClusterRequest{
		ClusterID:           id,
		Tags:                tags,
		AutoscalerConfig:    autoscalerConfig,
		AutoUpgrade:         autoUpgrade,
		FeatureGates:        featureGates,
		AdmissionPlugins:    admissionPlugins,
		OpenIDConnectConfig: openIDConnectConfig,
		ApiserverCertSans:   apiServerCertSANs,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdateCluster", err)
	}

	return nil
}

func (c *Client) UpgradeCluster(ctx context.Context, id, version string) error {
	if _, err := c.k8s.UpgradeCluster(&k8s.UpgradeClusterRequest{
		ClusterID:    id,
		Version:      version,
		UpgradePools: false,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpgradeCluster", err)
	}

	return nil
}

func (c *Client) SetClusterType(ctx context.Context, id, clusterType string) error {
	if _, err := c.k8s.SetClusterType(&k8s.SetClusterTypeRequest{
		ClusterID: id,
		Type:      clusterType,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("SetClusterType", err)
	}

	return nil
}

func (c *Client) FindPool(ctx context.Context, clusterID, name string) (*k8s.Pool, error) {
	resp, err := c.k8s.ListPools(&k8s.ListPoolsRequest{
		ClusterID: clusterID,
		Name:      scw.StringPtr(name),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListPools", err)
	}

	// Filter out all pools that have the wrong name.
	pools := slices.DeleteFunc(resp.Pools, func(pool *k8s.Pool) bool {
		return pool.Name != name
	})

	switch len(pools) {
	case 0:
		return nil, ErrNoItemFound
	case 1:
		return pools[0], nil
	default:
		// This case should never happen as k8s API prevents the creation of
		// multiple pools with the same name.
		return nil, fmt.Errorf("%w: found %d pools with name %s", ErrTooManyItemsFound, len(pools), name)
	}
}

func (c *Client) CreatePool(
	ctx context.Context,
	zone scw.Zone,
	clusterID, name, nodeType string,
	placementGroupID, securityGroupID *string,
	autoscaling, autohealing, publicIPDisabled bool,
	size uint32,
	minSize, maxSize *uint32,
	tags []string,
	kubeletArgs map[string]string,
	rootVolumeType k8s.PoolVolumeType,
	rootVolumeSizeGB *uint64,
	upgradePolicy *k8s.CreatePoolRequestUpgradePolicy,
) (*k8s.Pool, error) {
	var rootVolumeSize *scw.Size
	if rootVolumeSizeGB != nil {
		rootVolumeSize = scw.SizePtr(scw.Size(*rootVolumeSizeGB) * scw.GB)
	}

	pool, err := c.k8s.CreatePool(&k8s.CreatePoolRequest{
		Zone:             zone,
		ClusterID:        clusterID,
		Name:             name,
		NodeType:         nodeType,
		PlacementGroupID: placementGroupID,
		Autoscaling:      autoscaling,
		Autohealing:      autohealing,
		PublicIPDisabled: publicIPDisabled,
		Size:             size,
		MinSize:          minSize,
		MaxSize:          maxSize,
		Tags:             append(tags, createdByTag),
		KubeletArgs:      kubeletArgs,
		RootVolumeType:   rootVolumeType,
		RootVolumeSize:   rootVolumeSize,
		SecurityGroupID:  securityGroupID,
		UpgradePolicy:    upgradePolicy,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, newCallError("CreatePool", err)
	}

	return pool, nil
}

func (c *Client) UpdatePool(
	ctx context.Context,
	id string,
	autoscaling, autohealing *bool,
	size, minSize, maxSize *uint32,
	tags *[]string,
	kubeletArgs *map[string]string,
	upgradePolicy *k8s.UpdatePoolRequestUpgradePolicy,
) error {
	if tags != nil {
		*tags = append(*tags, createdByTag)
	}

	if _, err := c.k8s.UpdatePool(&k8s.UpdatePoolRequest{
		PoolID:        id,
		Autoscaling:   autoscaling,
		Size:          size,
		MinSize:       minSize,
		MaxSize:       maxSize,
		Autohealing:   autohealing,
		Tags:          tags,
		KubeletArgs:   kubeletArgs,
		UpgradePolicy: upgradePolicy,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpdatePool", err)
	}

	return nil
}

func (c *Client) UpgradePool(ctx context.Context, id, version string) error {
	if _, err := c.k8s.UpgradePool(&k8s.UpgradePoolRequest{
		PoolID:  id,
		Version: version,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("UpgradePool", err)
	}

	return nil
}

func (c *Client) DeletePool(ctx context.Context, id string) error {
	if _, err := c.k8s.DeletePool(&k8s.DeletePoolRequest{
		PoolID: id,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("DeletePool", err)
	}

	return nil
}

func (c *Client) ListNodes(ctx context.Context, clusterID, poolID string) ([]*k8s.Node, error) {
	resp, err := c.k8s.ListNodes(&k8s.ListNodesRequest{
		ClusterID: clusterID,
		PoolID:    &poolID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListNodes", err)
	}

	return resp.Nodes, nil
}

func (c *Client) ListClusterACLRules(ctx context.Context, clusterID string) ([]*k8s.ACLRule, error) {
	resp, err := c.k8s.ListClusterACLRules(&k8s.ListClusterACLRulesRequest{
		ClusterID: clusterID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, newCallError("ListClusterACLRules", err)
	}

	return resp.Rules, nil
}

func (c *Client) SetClusterACLRules(ctx context.Context, clusterID string, rules []*k8s.ACLRuleRequest) error {
	for _, rule := range rules {
		rule.Description = createdByDescription
	}

	if _, err := c.k8s.SetClusterACLRules(&k8s.SetClusterACLRulesRequest{
		ClusterID: clusterID,
		ACLs:      rules,
	}, scw.WithContext(ctx)); err != nil {
		return newCallError("SetClusterACLRules", err)
	}

	return nil
}
