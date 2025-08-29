package controller

import (
	"context"
	"fmt"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/k8s/cluster"
)

type scalewayManagedControlPlaneService struct {
	scope *scope.ManagedControlPlane
	// services is the list of services that are reconciled by this controller.
	// The order of the services is important as it determines the order in which the services are reconciled.
	services  []scaleway.ServiceReconciler
	Reconcile func(context.Context) error
	Delete    func(context.Context) error
}

func newScalewayManagedControlPlaneService(s *scope.ManagedControlPlane) *scalewayManagedControlPlaneService {
	scs := &scalewayManagedControlPlaneService{
		scope: s,
		services: []scaleway.ServiceReconciler{
			cluster.New(s),
		},
	}

	scs.Reconcile = scs.reconcile
	scs.Delete = scs.delete

	return scs
}

func (s *scalewayManagedControlPlaneService) reconcile(ctx context.Context) error {
	for _, service := range s.services {
		if err := service.Reconcile(ctx); err != nil {
			return fmt.Errorf("failed to reconcile ScalewayManagedControlPlane service %s: %w", service.Name(), err)
		}
	}

	return nil
}

func (s *scalewayManagedControlPlaneService) delete(ctx context.Context) error {
	for i := len(s.services) - 1; i >= 0; i-- {
		if err := s.services[i].Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete ScalewayManagedControlPlane service %s: %w", s.services[i].Name(), err)
		}
	}

	return nil
}
