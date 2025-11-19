package scope

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

const (
	maxClusterNameLength = 100
	resourcePrefix       = "caps-"
)

type ManagedControlPlane struct {
	patchHelper *patch.Helper

	Client                      client.Client
	Cluster                     *clusterv1.Cluster
	ScalewayManagedCluster      *infrav1.ScalewayManagedCluster
	ScalewayManagedControlPlane *infrav1.ScalewayManagedControlPlane
	ScalewayClient              scwClient.Interface
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

	helper, err := patch.NewHelper(params.ManagedControlPlane, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayCluster: %w", err)
	}

	mcp := &ManagedControlPlane{
		patchHelper:                 helper,
		Client:                      params.Client,
		Cluster:                     params.Cluster,
		ScalewayManagedCluster:      params.ManagedCluster,
		ScalewayManagedControlPlane: params.ManagedControlPlane,
	}

	mcp.ScalewayClient, err = newScalewayClientForScalewayManagedCluster(ctx, params.Client, params.ManagedCluster)
	if err != nil {
		return nil, errors.Join(err, mcp.Close(ctx))
	}

	return mcp, nil
}

// PatchObject patches the ScalewayManagedControlPlane object.
func (m *ManagedControlPlane) PatchObject(ctx context.Context) error {
	summaryConditions := []string{
		infrav1.ScalewayManagedControlPlaneClusterReadyCondition,
	}

	if err := conditions.SetSummaryCondition(
		m.ScalewayManagedControlPlane, m.ScalewayManagedControlPlane,
		infrav1.ScalewayManagedControlPlaneReadyCondition,
		conditions.ForConditionTypes(summaryConditions),
	); err != nil {
		return err
	}

	return m.patchHelper.Patch(ctx, m.ScalewayManagedControlPlane, patch.WithOwnedConditions{
		Conditions: append(summaryConditions, infrav1.ScalewayManagedControlPlaneReadyCondition),
	})
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
			fmt.Sprintf("caps-namespace=%s", c.ScalewayManagedControlPlane.Namespace),
			fmt.Sprintf("caps-scalewaymanagedcontrolplane=%s", c.ScalewayManagedControlPlane.Name),
		}, additional...)
}

// PrivateNetworkID returns the Private Network ID that should be used when creating
// the managed cluster. It's nil if no Private Network ID is needed.
func (m *ManagedControlPlane) PrivateNetworkID() *string {
	if m.ScalewayManagedCluster.Status.Network.PrivateNetworkID == "" {
		return nil
	}

	return ptr.To(string(m.ScalewayManagedCluster.Status.Network.PrivateNetworkID))
}

// DeleteWithAdditionalResources returns true if we should tell Scaleway k8s API
// to delete additional resources when cluster is deleted.
func (m *ManagedControlPlane) DeleteWithAdditionalResources() bool {
	return ptr.Deref(m.ScalewayManagedControlPlane.Spec.OnDelete.WithAdditionalResources, false)
}

// SetControlPlaneEndpoint sets the control plane endpoint host and port.
func (m *ManagedControlPlane) SetControlPlaneEndpoint(host string, port int32) {
	m.ScalewayManagedControlPlane.Spec.ControlPlaneEndpoint.Host = host
	m.ScalewayManagedControlPlane.Spec.ControlPlaneEndpoint.Port = port
}

// SetStatusVersion sets the current cluster Kubernetes version in the status.
func (m *ManagedControlPlane) SetStatusVersion(version string) {
	m.ScalewayManagedControlPlane.Status.Version = "v" + version
}

// DesiredVersion returns the desired Kubernetes version, without leading "v".
func (m *ManagedControlPlane) DesiredVersion() string {
	version, _ := strings.CutPrefix(m.ScalewayManagedControlPlane.Spec.Version, "v")
	return version
}

// FixedVersion returns the desired Kubernetes version, with a leading "v" if it's missing.
func (m *ManagedControlPlane) FixedVersion() string {
	if !strings.HasPrefix(m.ScalewayManagedControlPlane.Spec.Version, "v") {
		return "v" + m.ScalewayManagedControlPlane.Spec.Version
	}

	return m.ScalewayManagedControlPlane.Spec.Version
}

func (m *ManagedControlPlane) DesiredTags() []string {
	return m.ResourceTags(m.ScalewayManagedControlPlane.Spec.AdditionalTags...)
}

func (m *ManagedControlPlane) ClusterName() string {
	if m.ScalewayManagedControlPlane.Spec.ClusterName == "" {
		name, err := GenerateClusterName(m.ScalewayManagedControlPlane)
		if err != nil {
			panic(err)
		}

		m.ScalewayManagedControlPlane.Spec.ClusterName = name
	}

	return m.ScalewayManagedControlPlane.Spec.ClusterName
}

func (m *ManagedControlPlane) DesiredCNI() (cni k8s.CNI) {
	if m.ScalewayManagedControlPlane.Spec.CNI != "" {
		cni = k8s.CNI(m.ScalewayManagedControlPlane.Spec.CNI)
	}

	return
}

func (m *ManagedControlPlane) DesiredType() string {
	return m.ScalewayManagedControlPlane.Spec.Type
}

func (m *ManagedControlPlane) DesiredClusterAutoscalerConfig() (*k8s.ClusterAutoscalerConfig, error) {
	autoscaler := m.ScalewayManagedControlPlane.Spec.Autoscaler

	config := &k8s.ClusterAutoscalerConfig{
		ScaleDownDisabled:             ptr.Deref(autoscaler.ScaleDownDisabled, false),
		IgnoreDaemonsetsUtilization:   ptr.Deref(autoscaler.IgnoreDaemonsetsUtilization, false),
		BalanceSimilarNodeGroups:      ptr.Deref(autoscaler.BalanceSimilarNodeGroups, false),
		ScaleDownDelayAfterAdd:        "10m",
		Estimator:                     k8s.AutoscalerEstimatorBinpacking,
		Expander:                      k8s.AutoscalerExpanderRandom,
		ExpendablePodsPriorityCutoff:  ptr.Deref(autoscaler.ExpendablePodsPriorityCutoff, -10),
		ScaleDownUnneededTime:         "10m",
		ScaleDownUtilizationThreshold: 0.5,
		MaxGracefulTerminationSec:     600,
	}

	if autoscaler.ScaleDownDelayAfterAdd != "" {
		config.ScaleDownDelayAfterAdd = autoscaler.ScaleDownDelayAfterAdd
	}

	if autoscaler.Estimator != "" && autoscaler.Estimator != k8s.AutoscalerEstimatorUnknownEstimator.String() {
		config.Estimator = k8s.AutoscalerEstimator(autoscaler.Estimator)
	}

	if autoscaler.Expander != "" && autoscaler.Expander != k8s.AutoscalerEstimatorUnknownEstimator.String() {
		config.Expander = k8s.AutoscalerExpander(autoscaler.Expander)
	}

	if autoscaler.ScaleDownUnneededTime != "" {
		config.ScaleDownUnneededTime = autoscaler.ScaleDownUnneededTime
	}

	if autoscaler.ScaleDownUtilizationThreshold != "" {
		value, err := strconv.ParseFloat(autoscaler.ScaleDownUtilizationThreshold, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse scaleDownUtilizationThreshold as float32: %w", err)
		}

		config.ScaleDownUtilizationThreshold = float32(value)
	}

	if autoscaler.MaxGracefulTerminationSec > 0 {
		config.MaxGracefulTerminationSec = uint32(autoscaler.MaxGracefulTerminationSec)
	}

	return config, nil
}

func (m *ManagedControlPlane) DesiredAutoUpgrade() *k8s.ClusterAutoUpgrade {
	autoUpgrade := m.ScalewayManagedControlPlane.Spec.AutoUpgrade

	config := &k8s.ClusterAutoUpgrade{
		Enabled: ptr.Deref(autoUpgrade.Enabled, false),
		MaintenanceWindow: &k8s.MaintenanceWindow{
			StartHour: uint32(ptr.Deref(autoUpgrade.MaintenanceWindow.StartHour, 0)),
			Day:       k8s.MaintenanceWindowDayOfTheWeekAny,
		},
	}

	if autoUpgrade.MaintenanceWindow.Day != "" {
		config.MaintenanceWindow.Day = k8s.MaintenanceWindowDayOfTheWeek(autoUpgrade.MaintenanceWindow.Day)
	}

	return config
}

func (m *ManagedControlPlane) DesiredClusterOpenIDConnectConfig() *k8s.ClusterOpenIDConnectConfig {
	oidc := m.ScalewayManagedControlPlane.Spec.OpenIDConnect
	return &k8s.ClusterOpenIDConnectConfig{
		IssuerURL:      oidc.IssuerURL,
		ClientID:       oidc.ClientID,
		GroupsClaim:    oidc.GroupsClaim,
		RequiredClaim:  oidc.RequiredClaim,
		UsernamePrefix: oidc.UsernamePrefix,
		UsernameClaim:  oidc.UsernameClaim,
		GroupsPrefix:   oidc.GroupsPrefix,
	}
}

func (m *ManagedControlPlane) DesiredAllowedRanges() []string {
	// If ACL is not configured, we want all ranges to be allowed.
	if m.ScalewayManagedControlPlane.Spec.ACL == nil {
		return []string{"0.0.0.0/0"}
	}

	ranges := make([]string, 0, len(m.ScalewayManagedControlPlane.Spec.ACL.AllowedRanges))

	for _, r := range m.ScalewayManagedControlPlane.Spec.ACL.AllowedRanges {
		ranges = append(ranges, string(r))
	}

	return ranges
}

func (m *ManagedControlPlane) ClusterEndpoint(cluster *k8s.Cluster) string {
	if ptr.Deref(m.ScalewayManagedControlPlane.Spec.EnablePrivateEndpoint, false) && cluster.PrivateNetworkID != nil {
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

func GenerateClusterName(smcp *infrav1.ScalewayManagedControlPlane) (string, error) {
	if smcp.Name == "" || smcp.Namespace == "" {
		return "", errors.New("can't generate clusterName if name or namespace is not set")
	}

	return generateScalewayK8sName(smcp.Name, smcp.Namespace, maxClusterNameLength)
}
