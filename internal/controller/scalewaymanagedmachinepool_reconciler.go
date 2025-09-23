package controller

import (
	"context"
	"fmt"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/k8s/pool"
)

type scalewayManagedMachinePoolService struct {
	scope *scope.ManagedMachinePool
	// services is the list of services that are reconciled by this controller.
	// The order of the services is important as it determines the order in which the services are reconciled.
	services  []scaleway.ServiceReconciler
	Reconcile func(context.Context) error
	Delete    func(context.Context) error
}

func newScalewayManagedMachinePoolService(s *scope.ManagedMachinePool) *scalewayManagedMachinePoolService {
	svc := &scalewayManagedMachinePoolService{
		scope: s,
		services: []scaleway.ServiceReconciler{
			pool.New(s),
		},
	}

	svc.Reconcile = svc.reconcile
	svc.Delete = svc.delete

	return svc
}

func (s *scalewayManagedMachinePoolService) reconcile(ctx context.Context) error {
	for _, service := range s.services {
		if err := service.Reconcile(ctx); err != nil {
			return fmt.Errorf("failed to reconcile ScalewayManagedMachinePool service %s: %w", service.Name(), err)
		}
	}

	return nil
}

func (s *scalewayManagedMachinePoolService) delete(ctx context.Context) error {
	for i := len(s.services) - 1; i >= 0; i-- {
		if err := s.services[i].Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete ScalewayManagedMachinePool service %s: %w", s.services[i].Name(), err)
		}
	}

	return nil
}
