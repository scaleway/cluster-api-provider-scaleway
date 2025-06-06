package controller

import (
	"context"
	"fmt"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/instance"
)

type scalewayMachineService struct {
	scope *scope.Machine
	// services is the list of services that are reconciled by this controller.
	// The order of the services is important as it determines the order in which the services are reconciled.
	services  []scaleway.ServiceReconciler
	Reconcile func(context.Context) error
	Delete    func(context.Context) error
}

func newScalewayMachineService(s *scope.Machine) *scalewayMachineService {
	scs := &scalewayMachineService{
		scope: s,
		services: []scaleway.ServiceReconciler{
			instance.New(s),
		},
	}

	scs.Reconcile = scs.reconcile
	scs.Delete = scs.delete

	return scs
}

// Reconcile reconciles all the services in a predetermined order.
func (s *scalewayMachineService) reconcile(ctx context.Context) error {
	for _, service := range s.services {
		if err := service.Reconcile(ctx); err != nil {
			return fmt.Errorf("failed to reconcile ScalewayMachine service %s: %w", service.Name(), err)
		}
	}

	return nil
}

// Delete reconciles all the services in a predetermined order.
func (s *scalewayMachineService) delete(ctx context.Context) error {
	for i := len(s.services) - 1; i >= 0; i-- {
		if err := s.services[i].Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete ScalewayMachine service %s: %w", s.services[i].Name(), err)
		}
	}

	return nil
}
