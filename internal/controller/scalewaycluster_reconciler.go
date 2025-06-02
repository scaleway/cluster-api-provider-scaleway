package controller

import (
	"context"
	"fmt"
	"slices"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/domain"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/lb"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/vpc"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/vpcgw"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type scalewayClusterService struct {
	scope *scope.Cluster
	// services is the list of services that are reconciled by this controller.
	// The order of the services is important as it determines the order in which the services are reconciled.
	services  []scaleway.ServiceReconciler
	Reconcile func(context.Context) error
	Delete    func(context.Context) error
}

func newScalewayClusterService(s *scope.Cluster) *scalewayClusterService {
	scs := &scalewayClusterService{
		scope: s,
		services: []scaleway.ServiceReconciler{
			vpc.New(s),
			vpcgw.New(s),
			lb.New(s),
			domain.New(s),
		},
	}

	scs.Reconcile = scs.reconcile
	scs.Delete = scs.delete

	return scs
}

// Reconcile reconciles all the services in a predetermined order.
func (s *scalewayClusterService) reconcile(ctx context.Context) error {
	if err := s.setFailureDomainsForLocation(); err != nil {
		return scaleway.WithTerminalError(fmt.Errorf("failed to set failure domains in status: %w", err))
	}

	for _, service := range s.services {
		if err := service.Reconcile(ctx); err != nil {
			return fmt.Errorf("failed to reconcile ScalewayCluster service %s: %w", service.Name(), err)
		}
	}

	return nil
}

// Delete reconciles all the services in a predetermined order.
func (s *scalewayClusterService) delete(ctx context.Context) error {
	for i := len(s.services) - 1; i >= 0; i-- {
		if err := s.services[i].Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete ScalewayCluster service %s: %w", s.services[i].Name(), err)
		}
	}

	return nil
}

// setFailureDomainsForLocation sets the ScalewayCluster Status failure domains
// based on which Scaleway Availability Zones are available in the cluster location
// and the FailureDomains requested by user.
func (s *scalewayClusterService) setFailureDomainsForLocation() error {
	availableZones := s.scope.ScalewayClient.GetControlPlaneZones()

	var failureDomains []scw.Zone

	if len(s.scope.ScalewayCluster.Spec.FailureDomains) > 0 {
		for _, failureDomain := range s.scope.ScalewayCluster.Spec.FailureDomains {
			requestedZone, err := scw.ParseZone(failureDomain)
			if err != nil {
				return fmt.Errorf("failed to parse failureDomain %s as Scaleway zone: %w", failureDomain, err)
			}

			if !slices.Contains(availableZones, requestedZone) {
				return fmt.Errorf(
					"failureDomain %s is not allowed, you must use one of the following: %v",
					failureDomain, availableZones,
				)
			}

			failureDomains = append(failureDomains, requestedZone)
		}
	} else {
		failureDomains = availableZones
	}

	s.scope.SetFailureDomains(failureDomains)

	return nil
}
