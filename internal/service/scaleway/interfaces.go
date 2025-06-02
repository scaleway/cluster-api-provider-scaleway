package scaleway

import "context"

// Reconciler is a generic interface for a controller reconciler which has Reconcile and Delete methods.
type Reconciler interface {
	Reconcile(ctx context.Context) error
	Delete(ctx context.Context) error
}

// ServiceReconciler is a Scaleway service reconciler which can reconcile a Scaleway service.
type ServiceReconciler interface {
	Name() string
	Reconciler
}
