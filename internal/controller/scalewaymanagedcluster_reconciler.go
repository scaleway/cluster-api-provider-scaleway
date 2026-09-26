package controller

import (
	"context"
	"fmt"
	"slices"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/vpc"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/vpcgw"
)

type scalewayManagedClusterService struct {
	scope *scope.ManagedCluster
	// services is the list of services that are reconciled by this controller.
	// The order of the services is important as it determines the order in which the services are reconciled.
	services  []scaleway.ServiceReconciler
	Reconcile func(context.Context) error
	Delete    func(context.Context) error
}

func newScalewayManagedClusterService(s *scope.ManagedCluster) *scalewayManagedClusterService {
	scs := &scalewayManagedClusterService{
		scope: s,
		services: []scaleway.ServiceReconciler{
			vpc.New(s),
			vpcgw.New(s),
		},
	}

	scs.Reconcile = scs.reconcile
	scs.Delete = scs.delete

	return scs
}

func (s *scalewayManagedClusterService) reconcile(ctx context.Context) error {
	s.setFailureDomains()

	for _, service := range s.services {
		if err := service.Reconcile(ctx); err != nil {
			return fmt.Errorf("failed to reconcile ScalewayManagedCluster service %s: %w", service.Name(), err)
		}
	}

	return nil
}

func (s *scalewayManagedClusterService) delete(ctx context.Context) error {
	for _, service := range slices.Backward(s.services) {
		if err := service.Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete ScalewayManagedCluster service %s: %w", service.Name(), err)
		}
	}

	return nil
}

// setFailureDomains sets the ScalewayManagedCluster Status failure domains
// based on the zones where the pools of the cluster can be created.
func (s *scalewayManagedClusterService) setFailureDomains() {
	// Pools of multicloud clusters can be created in any zone. For other cluster
	// types, pools must be in the same region as the control plane.
	if s.scope.IsMulticloud() {
		s.scope.SetFailureDomains(s.scope.ScalewayClient.GetAllZones())
		return
	}

	s.scope.SetFailureDomains(s.scope.ScalewayClient.GetZones())
}
