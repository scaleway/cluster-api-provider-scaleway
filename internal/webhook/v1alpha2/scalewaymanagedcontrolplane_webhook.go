package v1alpha2

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
)

// nolint:unused
// log is for logging in this package.
var scalewaymanagedcontrolplanelog = logf.Log.WithName("scalewaymanagedcontrolplane-resource")

// SetupScalewayManagedControlPlaneWebhookWithManager registers the webhook for ScalewayManagedControlPlane in the manager.
func SetupScalewayManagedControlPlaneWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&infrav1.ScalewayManagedControlPlane{}).
		WithDefaulter(&ScalewayManagedControlPlaneCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1alpha2-scalewaymanagedcontrolplane,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=scalewaymanagedcontrolplanes,verbs=create;update,versions=v1alpha2,name=mscalewaymanagedcontrolplane-v1alpha2.kb.io,admissionReviewVersions=v1

// ScalewayManagedControlPlaneCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ScalewayManagedControlPlane when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ScalewayManagedControlPlaneCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &ScalewayManagedControlPlaneCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ScalewayManagedControlPlane.
func (d *ScalewayManagedControlPlaneCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	scalewaymanagedcontrolplane, ok := obj.(*infrav1.ScalewayManagedControlPlane)
	if !ok {
		return fmt.Errorf("expected an ScalewayManagedControlPlane object but got %T", obj)
	}

	scalewaymanagedcontrolplanelog.Info("Defaulting for ScalewayManagedControlPlane", "name", scalewaymanagedcontrolplane.GetName())

	if scalewaymanagedcontrolplane.Spec.ClusterName == "" {
		name, err := scope.GenerateClusterName(scalewaymanagedcontrolplane)
		if err != nil {
			return err
		}

		scalewaymanagedcontrolplane.Spec.ClusterName = name
	}

	return nil
}
