package v1alpha2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

// nolint:unused
// log is for logging in this package.
var scalewayclusterlog = logf.Log.WithName("scalewaycluster-resource")

// SetupScalewayClusterWebhookWithManager registers the webhook for ScalewayCluster in the manager.
func SetupScalewayClusterWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&infrav1.ScalewayCluster{}).
		Complete()
}
