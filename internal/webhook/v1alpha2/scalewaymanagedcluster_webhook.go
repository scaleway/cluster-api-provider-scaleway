package v1alpha2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

// nolint:unused
// log is for logging in this package.
var scalewaymanagedclusterlog = logf.Log.WithName("scalewaymanagedcluster-resource")

// SetupScalewayManagedClusterWebhookWithManager registers the webhook for ScalewayManagedCluster in the manager.
func SetupScalewayManagedClusterWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&infrav1.ScalewayManagedCluster{}).
		Complete()
}
