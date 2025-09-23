package controller

import "github.com/scaleway/cluster-api-provider-scaleway/internal/controller"

// Following variables provide access to reconcilers implemented in internal/controller.
var (
	NewScalewayClusterReconciler             = controller.NewScalewayClusterReconciler
	NewScalewayMachineReconciler             = controller.NewScalewayMachineReconciler
	NewScalewayManagedClusterReconciler      = controller.NewScalewayManagedClusterReconciler
	NewScalewayManagedControlPlaneReconciler = controller.NewScalewayManagedControlPlaneReconciler
	NewScalewayManagedMachinePoolReconciler  = controller.NewScalewayManagedMachinePoolReconciler
)
