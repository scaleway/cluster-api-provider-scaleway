package v1alpha2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

// nolint:unused
// log is for logging in this package.
var scalewaymachinetemplatelog = logf.Log.WithName("scalewaymachinetemplate-resource")

// SetupScalewayMachineTemplateWebhookWithManager registers the webhook for ScalewayMachineTemplate in the manager.
func SetupScalewayMachineTemplateWebhookWithManager(mgr ctrl.Manager) error {
<<<<<<< HEAD
	return ctrl.NewWebhookManagedBy(mgr, &infrastructurev1alpha2.ScalewayMachineTemplate{}).
=======
	return ctrl.NewWebhookManagedBy(mgr).For(&infrav1.ScalewayMachineTemplate{}).
>>>>>>> tmp-original-13-02-26-16-17
		Complete()
}
