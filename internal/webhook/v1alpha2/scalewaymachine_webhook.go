package v1alpha2

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/util/compare"
)

// nolint:unused
// log is for logging in this package.
var scalewaymachinelog = logf.Log.WithName("scalewaymachine-resource")

// ScalewayMachineCustomValidator struct is responsible for validating the ScalewayMachine resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ScalewayMachineCustomValidator struct{}

// SetupScalewayMachineWebhookWithManager registers the webhook for ScalewayMachine in the manager.
func SetupScalewayMachineWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &infrav1.ScalewayMachine{}).
		WithValidator(&ScalewayMachineCustomValidator{}).
		Complete()
}

func ignoreProviderID(old infrav1.ScalewayMachineSpec) cmp.Option {
	if old.ProviderID == "" {
		return cmpopts.IgnoreFields(infrav1.ScalewayMachineSpec{}, "ProviderID")
	}
	return cmp.AllowUnexported()
}

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha2-scalewaymachine,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=scalewaymachines,verbs=create;update,versions=v1alpha2,name=vscalewaymachine-v1alpha2.kb.io,admissionReviewVersions=v1

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ScalewayMachine .
func (webhook *ScalewayMachineCustomValidator) ValidateCreate(_ context.Context, obj *infrav1.ScalewayMachine) (admission.Warnings, error) {
	scalewaymachinelog.Info("Validation for ScalewayMachine upon creation", "name", obj.GetName())
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ScalewayMachine.
func (webhook *ScalewayMachineCustomValidator) ValidateUpdate(_ context.Context, oldObj *infrav1.ScalewayMachine, newObj *infrav1.ScalewayMachine) (admission.Warnings, error) {
	scalewaymachinelog.Info("Validation for ScalewayMachine upon update", "name", oldObj.GetName())

	var allErrs field.ErrorList

	opts := []cmp.Option{}
	opts = append(opts, ignoreProviderID(oldObj.Spec))

	equal, diff, err := compare.Diff(oldObj.Spec, newObj.Spec, opts...)
	if err != nil {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("failed to compare old and new ScalewayMachine: %v", err))
	}
	if !equal {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec"), newObj, fmt.Sprintf("ScalewayMachine spec. field is immutable. Please create a new resource instead. Diff: %s", diff)),
		)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(infrav1.GroupVersion.WithKind("ScalewayMachine").GroupKind(), newObj.Name, allErrs)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ScalewayMachine.
func (webhook *ScalewayMachineCustomValidator) ValidateDelete(_ context.Context, obj *infrav1.ScalewayMachine) (admission.Warnings, error) {
	scalewaymachinelog.Info("Validation for ScalewayMachine upon deletion", "name", obj.GetName())
	return nil, nil
}
