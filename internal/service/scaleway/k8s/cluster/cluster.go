package cluster

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/common"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

const clusterRetryTime = 30 * time.Second

type kubeconfigGetter func() (*k8s.Kubeconfig, error)

type Service struct {
	*scope.ManagedControlPlane
}

func New(s *scope.ManagedControlPlane) *Service {
	return &Service{ManagedControlPlane: s}
}

func (s *Service) Name() string {
	return "k8s_cluster"
}

func (s *Service) Reconcile(ctx context.Context) error {
	cluster, err := s.getOrCreateCluster(ctx)
	if err != nil {
		return err
	}

	if cluster.Status != k8s.ClusterStatusReady {
		return scaleway.WithTransientError(fmt.Errorf("cluster %s is not yet ready: currently %s", cluster.ID, cluster.Status), clusterRetryTime)
	}

	// Reconcile cluster type.
	if desiredType := s.DesiredType(); desiredType != cluster.Type {
		if err := s.ScalewayClient.SetClusterType(ctx, cluster.ID, desiredType); err != nil {
			return err
		}

		return scaleway.WithTransientError(fmt.Errorf("cluster %s is changing type to %s", cluster.ID, desiredType), clusterRetryTime)
	}

	// Reconcile cluster version.
	desiredVersion := s.DesiredVersion()
	clusterUpToDate, err := common.IsUpToDate(cluster.Version, desiredVersion)
	if err != nil {
		return err
	}
	if !clusterUpToDate {
		if err := s.ScalewayClient.UpgradeCluster(ctx, cluster.ID, desiredVersion); err != nil {
			return err
		}

		return scaleway.WithTransientError(fmt.Errorf("cluster %s is upgrading to %s", cluster.ID, desiredVersion), clusterRetryTime)
	}

	// Reconcile cluster changes (tags, autoscaler, etc.).
	updated, err := s.updateCluster(ctx, cluster)
	if err != nil {
		return err
	}
	if updated {
		return scaleway.WithTransientError(fmt.Errorf("cluster %s is being updated", cluster.ID), clusterRetryTime)
	}

	// Reconcile cluster ACL.
	updated, err = s.updateClusterACLs(ctx, cluster)
	if err != nil {
		return err
	}
	if updated {
		return scaleway.WithTransientError(fmt.Errorf("cluster %s is updating ACLs", cluster.ID), clusterRetryTime)
	}

	// Reconcile kubeconfig.
	getKubeconfigOnce := sync.OnceValues(func() (*k8s.Kubeconfig, error) {
		return s.ScalewayClient.GetClusterKubeConfig(ctx, cluster.ID)
	})
	if err := s.reconcileKubeconfig(ctx, cluster, getKubeconfigOnce); err != nil {
		return err
	}
	if err := s.reconcileAdditionalKubeconfigs(ctx, cluster, getKubeconfigOnce); err != nil {
		return err
	}

	host, port, err := urlToHostPort(s.ClusterEndpoint(cluster))
	if err != nil {
		return err
	}

	s.SetControlPlaneEndpoint(host, port)
	s.SetStatusVersion(cluster.Version)

	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	clusterName := s.ManagedControlPlane.ManagedControlPlane.Spec.ClusterName
	if clusterName == nil {
		return nil
	}

	cluster, err := s.ScalewayClient.FindCluster(ctx, *clusterName)
	if err != nil {
		if client.IsNotFoundError(err) {
			return nil
		}

		return err
	}

	if err := s.ScalewayClient.DeleteCluster(ctx, cluster.ID, s.DeleteWithAdditionalResources()); err != nil {
		return err
	}

	return nil
}

func (s *Service) getOrCreateCluster(ctx context.Context) (*k8s.Cluster, error) {
	cluster, err := s.ScalewayClient.FindCluster(ctx, s.ClusterName())
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return nil, err
	}

	if cluster == nil {
		smcp := s.ManagedControlPlane.ManagedControlPlane
		autoscalerConfig, err := s.DesiredClusterAutoscalerConfig()
		if err != nil {
			return nil, err
		}

		autoUpgrade := s.DesiredAutoUpgrade()
		oidcConfig := s.DesiredClusterOpenIDConnectConfig()

		var podCIDR, serviceCIDR scw.IPNet
		if clusterNetwork := s.Cluster.Spec.ClusterNetwork; clusterNetwork != nil {
			if clusterNetwork.Pods != nil && len(clusterNetwork.Pods.CIDRBlocks) > 0 {
				_, podCIDRIPNet, err := net.ParseCIDR(clusterNetwork.Pods.CIDRBlocks[0])
				if err != nil {
					return nil, err
				}

				podCIDR.IPNet = *podCIDRIPNet
			}

			if clusterNetwork.Services != nil && len(clusterNetwork.Services.CIDRBlocks) > 0 {
				_, podCIDRIPNet, err := net.ParseCIDR(clusterNetwork.Services.CIDRBlocks[0])
				if err != nil {
					return nil, err
				}

				serviceCIDR.IPNet = *podCIDRIPNet
			}
		}

		cluster, err = s.ScalewayClient.CreateCluster(
			ctx,
			s.ClusterName(),
			smcp.Spec.Type,
			s.DesiredVersion(),
			s.PrivateNetworkID(),
			s.DesiredTags(),
			smcp.Spec.FeatureGates,
			smcp.Spec.AdmissionPlugins,
			smcp.Spec.APIServerCertSANs,
			s.DesiredCNI(),
			&k8s.CreateClusterRequestAutoscalerConfig{
				ScaleDownDisabled:             &autoscalerConfig.ScaleDownDisabled,
				ScaleDownDelayAfterAdd:        &autoscalerConfig.ScaleDownDelayAfterAdd,
				Estimator:                     autoscalerConfig.Estimator,
				Expander:                      autoscalerConfig.Expander,
				IgnoreDaemonsetsUtilization:   &autoscalerConfig.IgnoreDaemonsetsUtilization,
				BalanceSimilarNodeGroups:      &autoscalerConfig.BalanceSimilarNodeGroups,
				ExpendablePodsPriorityCutoff:  &autoscalerConfig.ExpendablePodsPriorityCutoff,
				ScaleDownUnneededTime:         &autoscalerConfig.ScaleDownUnneededTime,
				ScaleDownUtilizationThreshold: &autoscalerConfig.ScaleDownUtilizationThreshold,
				MaxGracefulTerminationSec:     &autoscalerConfig.MaxGracefulTerminationSec,
			},
			&k8s.CreateClusterRequestAutoUpgrade{
				Enable: autoUpgrade.Enabled,
				MaintenanceWindow: &k8s.MaintenanceWindow{
					StartHour: autoUpgrade.MaintenanceWindow.StartHour,
					Day:       autoUpgrade.MaintenanceWindow.Day,
				},
			},
			&k8s.CreateClusterRequestOpenIDConnectConfig{
				IssuerURL:      oidcConfig.IssuerURL,
				ClientID:       oidcConfig.ClientID,
				UsernameClaim:  &oidcConfig.UsernameClaim,
				UsernamePrefix: &oidcConfig.UsernamePrefix,
				GroupsClaim:    &oidcConfig.GroupsClaim,
				GroupsPrefix:   &oidcConfig.GroupsPrefix,
				RequiredClaim:  &oidcConfig.RequiredClaim,
			},
			podCIDR,
			serviceCIDR,
		)
		if err != nil {
			return nil, err
		}
	}

	return cluster, nil
}

func (s *Service) updateCluster(ctx context.Context, cluster *k8s.Cluster) (bool, error) {
	updateNeeded := false
	smmp := s.ManagedControlPlane.ManagedControlPlane

	var tags *[]string
	if !common.SlicesEqualIgnoreOrder(client.TagsWithoutCreatedBy(cluster.Tags), s.DesiredTags()) {
		updateNeeded = true
		tags = scw.StringsPtr(s.DesiredTags())
	}

	var featureGates *[]string
	if !common.SlicesEqualIgnoreOrder(cluster.FeatureGates, smmp.Spec.FeatureGates) {
		updateNeeded = true
		featureGates = scw.StringsPtr(makeSliceIfNeeded(smmp.Spec.FeatureGates))
	}

	var admissionPlugins *[]string
	if !common.SlicesEqualIgnoreOrder(cluster.AdmissionPlugins, smmp.Spec.AdmissionPlugins) {
		updateNeeded = true
		admissionPlugins = scw.StringsPtr(makeSliceIfNeeded(smmp.Spec.AdmissionPlugins))
	}

	var apiServerCertSANs *[]string
	if !common.SlicesEqualIgnoreOrder(cluster.ApiserverCertSans, smmp.Spec.APIServerCertSANs) {
		updateNeeded = true
		apiServerCertSANs = scw.StringsPtr(makeSliceIfNeeded(smmp.Spec.APIServerCertSANs))
	}

	var autoscalerConfig *k8s.UpdateClusterRequestAutoscalerConfig
	desiredAutoscalerConfig, err := s.DesiredClusterAutoscalerConfig()
	if err != nil {
		return false, err
	}
	if !autoscalerConfigMatchesDesired(cluster.AutoscalerConfig, desiredAutoscalerConfig) {
		updateNeeded = true
		autoscalerConfig = &k8s.UpdateClusterRequestAutoscalerConfig{
			ScaleDownDisabled:             &desiredAutoscalerConfig.ScaleDownDisabled,
			ScaleDownDelayAfterAdd:        &desiredAutoscalerConfig.ScaleDownDelayAfterAdd,
			Estimator:                     desiredAutoscalerConfig.Estimator,
			Expander:                      desiredAutoscalerConfig.Expander,
			IgnoreDaemonsetsUtilization:   &desiredAutoscalerConfig.IgnoreDaemonsetsUtilization,
			BalanceSimilarNodeGroups:      &desiredAutoscalerConfig.BalanceSimilarNodeGroups,
			ExpendablePodsPriorityCutoff:  &desiredAutoscalerConfig.ExpendablePodsPriorityCutoff,
			ScaleDownUnneededTime:         &desiredAutoscalerConfig.ScaleDownUnneededTime,
			ScaleDownUtilizationThreshold: &desiredAutoscalerConfig.ScaleDownUtilizationThreshold,
			MaxGracefulTerminationSec:     &desiredAutoscalerConfig.MaxGracefulTerminationSec,
		}
	}

	var autoUpgrade *k8s.UpdateClusterRequestAutoUpgrade
	desiredAutoUpgrade := s.DesiredAutoUpgrade()
	if !clusterAutoUpgradeMatchesDesired(cluster.AutoUpgrade, desiredAutoUpgrade) {
		updateNeeded = true
		autoUpgrade = &k8s.UpdateClusterRequestAutoUpgrade{
			Enable:            &desiredAutoUpgrade.Enabled,
			MaintenanceWindow: desiredAutoUpgrade.MaintenanceWindow,
		}
	}

	var oidcConfig *k8s.UpdateClusterRequestOpenIDConnectConfig
	desiredOIDCConfig := s.DesiredClusterOpenIDConnectConfig()
	if !clusterOpenIDConnectConfigMatchesDesired(cluster.OpenIDConnectConfig, desiredOIDCConfig) {
		updateNeeded = true
		oidcConfig = &k8s.UpdateClusterRequestOpenIDConnectConfig{
			IssuerURL:      &desiredOIDCConfig.IssuerURL,
			ClientID:       &desiredOIDCConfig.ClientID,
			UsernameClaim:  &desiredOIDCConfig.UsernameClaim,
			UsernamePrefix: &desiredOIDCConfig.UsernamePrefix,
			GroupsClaim:    &desiredOIDCConfig.GroupsClaim,
			GroupsPrefix:   &desiredOIDCConfig.GroupsPrefix,
			RequiredClaim:  &desiredOIDCConfig.RequiredClaim,
		}
	}

	if !updateNeeded {
		return false, nil
	}

	if err := s.ScalewayClient.UpdateCluster(
		ctx,
		cluster.ID,
		tags, featureGates, admissionPlugins, apiServerCertSANs,
		autoscalerConfig,
		autoUpgrade,
		oidcConfig,
	); err != nil {
		return false, fmt.Errorf("failed to update cluster: %w", err)
	}

	return true, nil
}

func (s *Service) updateClusterACLs(ctx context.Context, cluster *k8s.Cluster) (bool, error) {
	acls, err := s.ScalewayClient.ListClusterACLRules(ctx, cluster.ID)
	if err != nil {
		return false, err
	}

	desired := s.DesiredAllowedRanges()
	currentRanges, currentScalewayRanges := currentAllowedRanges(acls)

	if common.SlicesEqualIgnoreOrder(desired, currentRanges) {
		return false, nil
	}

	request := make([]*k8s.ACLRuleRequest, 0, len(currentRanges)+1)

	for _, cidr := range desired {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return false, fmt.Errorf("failed to parse range: %w", err)
		}

		request = append(request, &k8s.ACLRuleRequest{
			IP: &scw.IPNet{IPNet: *ipNet},
		})
	}

	if currentScalewayRanges {
		request = append(request, &k8s.ACLRuleRequest{
			ScalewayRanges: scw.BoolPtr(true),
		})
	}

	if err := s.ScalewayClient.SetClusterACLRules(ctx, cluster.ID, request); err != nil {
		return false, fmt.Errorf("failed to set ACLs: %w", err)
	}

	return true, nil
}

func currentAllowedRanges(rules []*k8s.ACLRule) (ranges []string, scalewayRanges bool) {
	ranges = make([]string, 0, len(rules))

	for _, rule := range rules {
		if rule.ScalewayRanges != nil {
			scalewayRanges = *rule.ScalewayRanges
		} else if rule.IP != nil {
			ranges = append(ranges, rule.IP.String())
		}
	}

	return
}

func urlToHostPort(s string) (string, int32, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return "", 0, err
	}

	return u.Hostname(), int32(port), nil
}

func makeSliceIfNeeded[T any](s []T) []T {
	if s == nil {
		return make([]T, 0)
	}

	return s
}

func autoscalerConfigMatchesDesired(current, desired *k8s.ClusterAutoscalerConfig) bool {
	if current == nil || desired == nil {
		return true
	}

	return current.ScaleDownDisabled == desired.ScaleDownDisabled &&
		current.ScaleDownDelayAfterAdd == desired.ScaleDownDelayAfterAdd &&
		current.Estimator == desired.Estimator &&
		current.Expander == desired.Expander &&
		current.IgnoreDaemonsetsUtilization == desired.IgnoreDaemonsetsUtilization &&
		current.BalanceSimilarNodeGroups == desired.BalanceSimilarNodeGroups &&
		current.ExpendablePodsPriorityCutoff == desired.ExpendablePodsPriorityCutoff &&
		current.ScaleDownUnneededTime == desired.ScaleDownUnneededTime &&
		current.ScaleDownUtilizationThreshold == desired.ScaleDownUtilizationThreshold &&
		current.MaxGracefulTerminationSec == desired.MaxGracefulTerminationSec
}

func clusterAutoUpgradeMatchesDesired(current, desired *k8s.ClusterAutoUpgrade) bool {
	if current == nil || desired == nil || current.MaintenanceWindow == nil || desired.MaintenanceWindow == nil {
		return true
	}

	return current.Enabled == desired.Enabled &&
		current.MaintenanceWindow.Day == desired.MaintenanceWindow.Day &&
		current.MaintenanceWindow.StartHour == desired.MaintenanceWindow.StartHour
}

func clusterOpenIDConnectConfigMatchesDesired(current, desired *k8s.ClusterOpenIDConnectConfig) bool {
	if current == nil || desired == nil {
		return true
	}

	return current.IssuerURL == desired.IssuerURL &&
		current.ClientID == desired.ClientID &&
		current.UsernameClaim == desired.UsernameClaim &&
		current.UsernamePrefix == desired.UsernamePrefix &&
		common.SlicesEqualIgnoreOrder(current.GroupsClaim, desired.GroupsClaim) &&
		current.GroupsPrefix == desired.GroupsPrefix &&
		common.SlicesEqualIgnoreOrder(current.RequiredClaim, desired.RequiredClaim)
}
