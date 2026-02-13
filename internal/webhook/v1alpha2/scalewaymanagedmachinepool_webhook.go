package v1alpha2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

// nolint:unused
// log is for logging in this package.
var scalewaymanagedmachinepoollog = logf.Log.WithName("scalewaymanagedmachinepool-resource")

// SetupScalewayManagedMachinePoolWebhookWithManager registers the webhook for ScalewayManagedMachinePool in the manager.
func SetupScalewayManagedMachinePoolWebhookWithManager(mgr ctrl.Manager) error {
<<<<<<< HEAD
	return ctrl.NewWebhookManagedBy(mgr, &infrastructurev1alpha2.ScalewayManagedMachinePool{}).
=======
	return ctrl.NewWebhookManagedBy(mgr).For(&infrav1.ScalewayManagedMachinePool{}).
>>>>>>> tmp-original-13-02-26-16-17
		Complete()
}
