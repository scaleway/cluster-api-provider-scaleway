package domain

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
	if !s.HasControlPlaneDNS() {
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

func (s *Service) Reconcile(ctx context.Context) error {
	if !s.HasControlPlaneDNS() {
		return nil
	}

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

	if slices.Equal(recordIPs, controlPlaneIPs) {
		return nil
	}

	logf.FromContext(ctx).Info("Updating zone records", "zone", zone, "name", name, "controlPlaneIPs", controlPlaneIPs)

	if err := s.ScalewayClient.SetDNSZoneRecords(ctx, zone, name, controlPlaneIPs); err != nil {
		return fmt.Errorf("failed to set dns records: %w", err)
	}

	return nil
}
