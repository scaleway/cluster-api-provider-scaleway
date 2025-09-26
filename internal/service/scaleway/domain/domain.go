package domain

import (
	"context"
	"errors"
	"fmt"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/util/conditions"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

type Service struct {
	*scope.Cluster
}

func New(clusterScope *scope.Cluster) *Service {
	return &Service{Cluster: clusterScope}
}

func (s *Service) Name() string {
	return "domain"
}

func (s *Service) Delete(ctx context.Context) error {
	if !s.ScalewayCluster.Spec.Network.ControlPlaneDNS.IsDefined() {
		return nil
	}

	zone, name, err := s.ControlPlaneDNSZoneAndName()
	if err != nil {
		return err
	}

	records, err := s.ScalewayClient.ListDNSZoneRecords(ctx, zone, name)
	if err != nil {
		// Domain API returns forbidden error when domain is not found.
		if client.IsForbiddenError(err) {
			return nil
		}

		return err
	}

	if len(records) == 0 {
		return nil
	}

	logf.FromContext(ctx).Info("Deleting zone records", "zone", zone, "name", name)

	if err := s.ScalewayClient.DeleteDNSZoneRecords(ctx, zone, name); err != nil {
		return fmt.Errorf("failed to delete dns records: %w", err)
	}

	return nil
}

func (s *Service) Reconcile(ctx context.Context) (retErr error) {
	if !s.ScalewayCluster.Spec.Network.ControlPlaneDNS.IsDefined() {
		conditions.Set(s.ScalewayCluster, metav1.Condition{
			Type:   infrav1.ScalewayClusterDomainReadyCondition,
			Status: metav1.ConditionTrue,
			Reason: infrav1.ScalewayClusterNoDomainReason,
		})
		return nil
	}

	defer func() {
		if retErr != nil {
			conditions.Set(s.ScalewayCluster, metav1.Condition{
				Type:    infrav1.ScalewayClusterDomainReadyCondition,
				Status:  metav1.ConditionFalse,
				Reason:  infrav1.ScalewayClusterDomainReconciliationFailedReason,
				Message: retErr.Error(),
			})
		}
	}()

	zone, name, err := s.ControlPlaneDNSZoneAndName()
	if err != nil {
		return err
	}

	records, err := s.ScalewayClient.ListDNSZoneRecords(ctx, zone, name)
	if err != nil {
		return err
	}

	recordIPs := make([]string, 0, len(records))
	for _, record := range records {
		recordIPs = append(recordIPs, record.Data)
	}
	slices.Sort(recordIPs)

	controlPlaneIPs := s.ControlPlaneLoadBalancerIPs()
	if len(controlPlaneIPs) == 0 {
		return errors.New("no control plane ips found")
	}

	if !slices.Equal(recordIPs, controlPlaneIPs) {
		logf.FromContext(ctx).Info("Updating zone records", "zone", zone, "name", name, "controlPlaneIPs", controlPlaneIPs)

		if err := s.ScalewayClient.SetDNSZoneRecords(ctx, zone, name, controlPlaneIPs); err != nil {
			return fmt.Errorf("failed to set dns records: %w", err)
		}
	}

	conditions.Set(s.ScalewayCluster, metav1.Condition{
		Type:   infrav1.ScalewayClusterDomainReadyCondition,
		Status: metav1.ConditionTrue,
		Reason: infrav1.ScalewayClusterDomainZoneConfiguredReason,
	})

	return nil
}
