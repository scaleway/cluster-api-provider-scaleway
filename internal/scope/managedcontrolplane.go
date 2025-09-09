package scope

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxClusterNameLength = 100
	resourcePrefix       = "caps-"
)

type ManagedControlPlane struct {
	patchHelper *patch.Helper

	Client              client.Client
	Cluster             *clusterv1.Cluster
	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane
	ScalewayClient      scwClient.Interface
}

// ClusterParams contains mandatory params for creating the Cluster scope.
type ManagedControlPlaneParams struct {
	Client              client.Client
	Cluster             *clusterv1.Cluster
	ManagedCluster      *infrav1.ScalewayManagedCluster
	ManagedControlPlane *infrav1.ScalewayManagedControlPlane
}

// NewCluster creates a new Cluster scope.
func NewManagedControlPlane(ctx context.Context, params *ManagedControlPlaneParams) (*ManagedControlPlane, error) {
	c, err := newScalewayClientForScalewayManagedCluster(ctx, params.Client, params.ManagedCluster)
	if err != nil {
		return nil, err
	}

	helper, err := patch.NewHelper(params.ManagedControlPlane, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayCluster: %w", err)
	}

	return &ManagedControlPlane{
		patchHelper:         helper,
		Client:              params.Client,
		ScalewayClient:      c,
		Cluster:             params.Cluster,
		ManagedCluster:      params.ManagedCluster,
		ManagedControlPlane: params.ManagedControlPlane,
	}, nil
}

// PatchObject patches the ScalewayManagedControlPlane object.
func (m *ManagedControlPlane) PatchObject(ctx context.Context) error {
	return m.patchHelper.Patch(ctx, m.ManagedControlPlane)
}

// Close closes the Machine scope by patching the ScalewayManagedControlPlane object.
func (m *ManagedControlPlane) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// ResourceTags returns the tags that resources created for the control-plane should have.
// It is possible to provide additional tags that will be added to the default tags.
func (c *ManagedControlPlane) ResourceTags(additional ...string) []string {
	return append(
		[]string{
			fmt.Sprintf("caps-namespace=%s", c.ManagedControlPlane.Namespace),
			fmt.Sprintf("caps-scalewaymanagedcontrolplane=%s", c.ManagedControlPlane.Name),
		}, additional...)
}

// PrivateNetworkID returns the Private Network ID that should be used when creating
// the managed cluster. It's nil if no Private Network ID is needed.
func (m *ManagedControlPlane) PrivateNetworkID() *string {
	if m.ManagedCluster.Status.Network == nil {
		return nil
	}

	return m.ManagedCluster.Status.Network.PrivateNetworkID
}

// DeleteWithAdditionalResources returns true if we should tell Scaleway k8s API
// to delete additional resources when cluster is deleted.
func (m *ManagedControlPlane) DeleteWithAdditionalResources() bool {
	if m.ManagedControlPlane.Spec.OnDelete == nil || m.ManagedControlPlane.Spec.OnDelete.WithAdditionalResources == nil {
		return false
	}

	return *m.ManagedControlPlane.Spec.OnDelete.WithAdditionalResources
}

// SetControlPlaneEndpoint sets the control plane endpoint host and port.
func (m *ManagedControlPlane) SetControlPlaneEndpoint(host string, port int32) {
	m.ManagedControlPlane.Spec.ControlPlaneEndpoint.Host = host
	m.ManagedControlPlane.Spec.ControlPlaneEndpoint.Port = port
}

// SetStatusVersion sets the current cluster Kubernetes version in the status.
func (m *ManagedControlPlane) SetStatusVersion(version string) {
	m.ManagedControlPlane.Status.Version = scw.StringPtr("v" + version)
}

// DesiredVersion returns the desired Kubernetes version, without leading "v".
func (m *ManagedControlPlane) DesiredVersion() string {
	version, _ := strings.CutPrefix(m.ManagedControlPlane.Spec.Version, "v")
	return version
}

// FixedVersion returns the desired Kubernetes version, with a leading "v" if it's missing.
func (m *ManagedControlPlane) FixedVersion() string {
	if !strings.HasPrefix(m.ManagedControlPlane.Spec.Version, "v") {
		return "v" + m.ManagedControlPlane.Spec.Version
	}

	return m.ManagedControlPlane.Spec.Version
}

func (m *ManagedControlPlane) DesiredTags() []string {
	return m.ResourceTags(m.ManagedControlPlane.Spec.AdditionalTags...)
}

func (m *ManagedControlPlane) ClusterName() string {
	if m.ManagedControlPlane.Spec.ClusterName == nil {
		name, err := generateScalewayK8sName(m.ManagedControlPlane.Name, m.ManagedControlPlane.Namespace, maxClusterNameLength)
		if err != nil {
			panic(err)
		}

		m.ManagedControlPlane.Spec.ClusterName = &name
	}

	return *m.ManagedControlPlane.Spec.ClusterName
}

func (m *ManagedControlPlane) DesiredCNI() k8s.CNI {
	var cni k8s.CNI
	if m.ManagedControlPlane.Spec.CNI != nil {
		cni = k8s.CNI(*m.ManagedControlPlane.Spec.CNI)
	}

	return cni
}

func (m *ManagedControlPlane) DesiredType() string {
	return m.ManagedControlPlane.Spec.Type
}

func (m *ManagedControlPlane) DesiredClusterAutoscalerConfig() (*k8s.ClusterAutoscalerConfig, error) {
	config := &k8s.ClusterAutoscalerConfig{
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
	}

	autoscaler := m.ManagedControlPlane.Spec.Autoscaler
	if autoscaler == nil {
		return config, nil
	}

	if autoscaler.ScaleDownDisabled != nil {
		config.ScaleDownDisabled = *autoscaler.ScaleDownDisabled
	}

	if autoscaler.ScaleDownDelayAfterAdd != nil {
		config.ScaleDownDelayAfterAdd = *autoscaler.ScaleDownDelayAfterAdd
	}

	if autoscaler.Estimator != nil && *autoscaler.Estimator != k8s.AutoscalerEstimatorUnknownEstimator.String() {
		config.Estimator = k8s.AutoscalerEstimator(*autoscaler.Estimator)
	}

	if autoscaler.Expander != nil && *autoscaler.Expander != k8s.AutoscalerEstimatorUnknownEstimator.String() {
		config.Expander = k8s.AutoscalerExpander(*autoscaler.Expander)
	}

	if autoscaler.IgnoreDaemonsetsUtilization != nil {
		config.IgnoreDaemonsetsUtilization = *autoscaler.IgnoreDaemonsetsUtilization
	}

	if autoscaler.BalanceSimilarNodeGroups != nil {
		config.BalanceSimilarNodeGroups = *autoscaler.BalanceSimilarNodeGroups
	}

	if autoscaler.ExpendablePodsPriorityCutoff != nil {
		config.ExpendablePodsPriorityCutoff = *autoscaler.ExpendablePodsPriorityCutoff
	}

	if autoscaler.ScaleDownUnneededTime != nil {
		config.ScaleDownUnneededTime = *autoscaler.ScaleDownUnneededTime
	}

	if autoscaler.ScaleDownUtilizationThreshold != nil {
		value, err := strconv.ParseFloat(*autoscaler.ScaleDownUtilizationThreshold, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse scaleDownUtilizationThreshold as float32: %w", err)
		}

		config.ScaleDownUtilizationThreshold = float32(value)
	}

	if autoscaler.MaxGracefulTerminationSec != nil {
		config.MaxGracefulTerminationSec = uint32(*autoscaler.MaxGracefulTerminationSec)
	}

	return config, nil
}

func (m *ManagedControlPlane) DesiredAutoUpgrade() *k8s.ClusterAutoUpgrade {
	config := &k8s.ClusterAutoUpgrade{
		Enabled: false,
		MaintenanceWindow: &k8s.MaintenanceWindow{
			StartHour: 0,
			Day:       k8s.MaintenanceWindowDayOfTheWeekAny,
		},
	}

	autoUpgrade := m.ManagedControlPlane.Spec.AutoUpgrade
	if autoUpgrade == nil {
		return config
	}

	config.Enabled = autoUpgrade.Enabled

	if autoUpgrade.MaintenanceWindow != nil {
		config.MaintenanceWindow = &k8s.MaintenanceWindow{}

		if autoUpgrade.MaintenanceWindow.StartHour != nil {
			config.MaintenanceWindow.StartHour = *scw.Uint32Ptr(uint32(*autoUpgrade.MaintenanceWindow.StartHour))
		}

		if autoUpgrade.MaintenanceWindow.Day != nil {
			config.MaintenanceWindow.Day = k8s.MaintenanceWindowDayOfTheWeek(*autoUpgrade.MaintenanceWindow.Day)
		}
	}

	return config
}

func (m *ManagedControlPlane) DesiredClusterOpenIDConnectConfig() *k8s.ClusterOpenIDConnectConfig {
	config := &k8s.ClusterOpenIDConnectConfig{
		GroupsClaim:   []string{},
		RequiredClaim: []string{},
	}

	oidc := m.ManagedControlPlane.Spec.OpenIDConnect
	if oidc == nil {
		return config
	}

	config.IssuerURL = oidc.IssuerURL
	config.ClientID = oidc.ClientID

	if config.GroupsClaim != nil {
		config.GroupsClaim = oidc.GroupsClaim
	}

	if config.RequiredClaim != nil {
		config.RequiredClaim = oidc.RequiredClaim
	}

	if oidc.UsernameClaim != nil {
		config.UsernameClaim = *oidc.UsernameClaim
	}

	if oidc.UsernamePrefix != nil {
		config.UsernamePrefix = *oidc.UsernamePrefix
	}

	if oidc.UsernameClaim != nil {
		config.UsernameClaim = *oidc.UsernameClaim
	}

	if oidc.GroupsPrefix != nil {
		config.GroupsPrefix = *oidc.GroupsPrefix
	}

	return config
}

func (m *ManagedControlPlane) DesiredAllowedRanges() []string {
	// If ACL is not configured, we want all ranges to be allowed.
	if m.ManagedControlPlane.Spec.ACL == nil {
		return []string{"0.0.0.0/0"}
	}

	ranges := make([]string, 0, len(m.ManagedControlPlane.Spec.ACL.AllowedRanges))

	for _, r := range m.ManagedControlPlane.Spec.ACL.AllowedRanges {
		ranges = append(ranges, string(r))
	}

	return ranges
}

func (m *ManagedControlPlane) ClusterEndpoint(cluster *k8s.Cluster) string {
	if m.ManagedControlPlane.Spec.EnablePrivateEndpoint != nil &&
		*m.ManagedControlPlane.Spec.EnablePrivateEndpoint &&
		cluster.PrivateNetworkID != nil {
		return fmt.Sprintf("https://%s.%s.internal:6443", cluster.ID, *cluster.PrivateNetworkID)
	}

	return cluster.ClusterURL
}

func generateScalewayK8sName(resourceName, namespace string, maxLength int) (string, error) {
	escapedName := strings.ReplaceAll(resourceName, ".", "-")
	name := fmt.Sprintf("%s-%s", namespace, escapedName)

	if len(name) < maxLength {
		return name, nil
	}

	hashLength := 64 - len(resourcePrefix)
	hashedName, err := base36TruncatedHash(name, hashLength)
	if err != nil {
		return "", fmt.Errorf("creating hash from name: %w", err)
	}

	return fmt.Sprintf("%s%s", resourcePrefix, hashedName), nil
}
