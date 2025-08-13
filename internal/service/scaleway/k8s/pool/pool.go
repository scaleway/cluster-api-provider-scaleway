package pool

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/common"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

const poolRetryTime = 30 * time.Second

type Service struct {
	*scope.ManagedMachinePool
}

func New(s *scope.ManagedMachinePool) *Service {
	return &Service{s}
}

func (s Service) Name() string {
	return "k8s_pool"
}

func (s *Service) Delete(ctx context.Context) error {
	clusterName, ok := s.ClusterName()
	if !ok {
		return nil
	}

	cluster, err := s.ScalewayClient.FindCluster(ctx, clusterName)
	if err != nil {
		if client.IsNotFoundError(err) {
			return nil
		}

		return err
	}

	pool, err := s.ScalewayClient.FindPool(ctx, cluster.ID, s.ResourceName())
	if err != nil {
		if client.IsNotFoundError(err) {
			return nil
		}

		return err
	}

	if pool.Status != k8s.PoolStatusDeleting {
		if err := s.ScalewayClient.DeletePool(ctx, pool.ID); err != nil {
			return err
		}
	}

	return scaleway.WithTransientError(errors.New("pool is being deleted"), poolRetryTime)
}

func (s *Service) Reconcile(ctx context.Context) error {
	clusterName, ok := s.ClusterName()
	if !ok {
		return scaleway.WithTransientError(errors.New("cluster name not set"), poolRetryTime)
	}

	cluster, err := s.ScalewayClient.FindCluster(ctx, clusterName)
	if err != nil {
		if client.IsNotFoundError(err) {
			return scaleway.WithTransientError(errors.New("cluster does not exist yet"), poolRetryTime)
		}
		return err
	}

	if !slices.Contains([]k8s.ClusterStatus{
		k8s.ClusterStatusReady,
		k8s.ClusterStatusPoolRequired,
	}, cluster.Status) {
		return scaleway.WithTransientError(fmt.Errorf("cluster %s is not yet ready: currently %s", cluster.ID, cluster.Status), poolRetryTime)
	}

	pool, err := s.getOrCreatePool(ctx, cluster)
	if err != nil {
		return err
	}

	if pool.Status != k8s.PoolStatusReady {
		return scaleway.WithTransientError(fmt.Errorf("pool %s is not yet ready: currently %s", pool.ID, pool.Status), poolRetryTime)
	}

	// Reconcile pool version.
	if desiredVersion := s.DesiredVersion(); desiredVersion != nil {
		poolUpToDate, err := common.IsUpToDate(pool.Version, *desiredVersion)
		if err != nil {
			return err
		}
		if !poolUpToDate {
			if err := s.ScalewayClient.UpgradePool(ctx, pool.ID, *desiredVersion); err != nil {
				return err
			}

			return scaleway.WithTransientError(fmt.Errorf("pool %s is upgrading to %s", cluster.ID, *desiredVersion), poolRetryTime)
		}
	}

	// Reconcile pools changes (size, tags, etc.).
	updated, err := s.updatePool(ctx, pool)
	if err != nil {
		return err
	}
	if updated {
		return scaleway.WithTransientError(fmt.Errorf("pool %s is being updated", cluster.ID), poolRetryTime)
	}

	nodes, err := s.ScalewayClient.ListNodes(ctx, cluster.ID, pool.ID)
	if err != nil {
		return err
	}

	s.SetProviderIDs(nodes)
	s.SetStatusReplicas(pool.Size)

	return nil
}

func (s *Service) getOrCreatePool(ctx context.Context, cluster *k8s.Cluster) (*k8s.Pool, error) {
	pool, err := s.ScalewayClient.FindPool(ctx, cluster.ID, s.ResourceName())
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return nil, err
	}

	if pool == nil {
		mmp := s.ManagedMachinePool.ManagedMachinePool

		autoscaling, size, min, max := s.Scaling()
		pup := s.DesiredPoolUpgradePolicy()

		pool, err = s.ScalewayClient.CreatePool(
			ctx,
			scw.Zone(mmp.Spec.Zone),
			cluster.ID,
			s.ResourceName(),
			mmp.Spec.NodeType,
			mmp.Spec.PlacementGroupID,
			mmp.Spec.SecurityGroupID,
			autoscaling,
			s.Autohealing(),
			s.PublicIPDisabled(),
			size,
			&min,
			&max,
			s.DesiredTags(),
			mmp.Spec.KubeletArgs,
			s.RootVolumeType(),
			s.RootVolumeSizeGB(),
			&k8s.CreatePoolRequestUpgradePolicy{
				MaxUnavailable: &pup.MaxUnavailable,
				MaxSurge:       &pup.MaxSurge,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return pool, nil
}

func (s *Service) updatePool(ctx context.Context, pool *k8s.Pool) (bool, error) {
	updateNeeded := false

	var autohealing *bool
	if pool.Autohealing != s.Autohealing() {
		updateNeeded = true
		autohealing = scw.BoolPtr(s.Autohealing())
	}

	var autoscaling *bool
	var size, minSize, maxSize *uint32

	if pool.NodeType != "external" {
		desiredAutoscaling, desiredSize, desiredMin, desiredMax := s.Scaling()

		if pool.Autoscaling != desiredAutoscaling {
			updateNeeded = true
			autoscaling = &desiredAutoscaling
		}

		// Only reconcile minSize and maxSize when autoscaling is enabled.
		if desiredAutoscaling {
			if pool.MinSize != desiredMin {
				updateNeeded = true
				minSize = &desiredMin
			}

			if pool.MaxSize != desiredMax {
				updateNeeded = true
				maxSize = &desiredMax
			}
		} else {
			// Only reconcile size when autoscaling is disabled.
			if pool.Size != desiredSize {
				updateNeeded = true
				size = &desiredSize
			}
		}
	}

	var tags *[]string
	if !common.SlicesEqualIgnoreOrder(client.TagsWithoutCreatedBy(pool.Tags), s.DesiredTags()) {
		updateNeeded = true
		tags = scw.StringsPtr(s.DesiredTags())
	}

	var kubeletArgs *map[string]string
	if !maps.Equal(pool.KubeletArgs, s.ManagedMachinePool.ManagedMachinePool.Spec.KubeletArgs) {
		updateNeeded = true
		kubeletArgs = &s.ManagedMachinePool.ManagedMachinePool.Spec.KubeletArgs
		if *kubeletArgs == nil {
			kubeletArgs = &map[string]string{}
		}
	}

	var upgradePolicy *k8s.UpdatePoolRequestUpgradePolicy
	desiredPoolUpgradePolicy := s.DesiredPoolUpgradePolicy()
	if !poolUpgradePolicyMatchesDesired(pool.UpgradePolicy, desiredPoolUpgradePolicy) {
		updateNeeded = true

		upgradePolicy = &k8s.UpdatePoolRequestUpgradePolicy{
			MaxUnavailable: &desiredPoolUpgradePolicy.MaxUnavailable,
			MaxSurge:       &desiredPoolUpgradePolicy.MaxSurge,
		}
	}

	if !updateNeeded {
		return false, nil
	}

	if err := s.ScalewayClient.UpdatePool(
		ctx,
		pool.ID,
		autoscaling, autohealing,
		size, minSize, maxSize,
		tags,
		kubeletArgs,
		upgradePolicy,
	); err != nil {
		return false, fmt.Errorf("failed to update pool: %w", err)
	}

	return true, nil
}

func poolUpgradePolicyMatchesDesired(current, desired *k8s.PoolUpgradePolicy) bool {
	if current == nil || desired == nil {
		return true
	}

	return current.MaxSurge == desired.MaxSurge &&
		current.MaxUnavailable == desired.MaxUnavailable
}
