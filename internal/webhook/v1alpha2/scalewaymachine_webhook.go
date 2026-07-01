package v1alpha2

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	compare "github.com/scaleway/cluster-api-provider-scaleway/internal/util"
)

// nolint:unused
// log is for logging in this package.
var scalewaymachinelog = logf.Log.WithName("scalewaymachine-resource")

type ScalewayMachineValidator struct{}

// SetupScalewayMachineWebhookWithManager registers the webhook for ScalewayMachineTemplate in the manager.
func SetupScalewayMachineWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &infrav1.ScalewayMachine{}).
		WithValidator(&ScalewayMachineValidator{}).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1alpha2-scalewaymachine,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=scalewaymachines,versions=v1alpha2,name=validation.scalewaymachine.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1

var _ admission.Validator[*infrav1.ScalewayMachine] = &ScalewayMachineValidator{}

func (webhook *ScalewayMachineValidator) ValidateCreate(_ context.Context, obj *infrav1.ScalewayMachine) (admission.Warnings, error) {
	scalewaymachinelog.Info("ValidateCreate ScalewayMachineTemplateValidator", "name", obj.GetName())
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (webhook *ScalewayMachineValidator) ValidateUpdate(ctx context.Context, oldObj *infrav1.ScalewayMachine, newObj *infrav1.ScalewayMachine) (admission.Warnings, error) {
	scalewaymachinelog.Info("ValidateUpdate ScalewayMachineValidator", "name", oldObj.GetName())

	var allErrs field.ErrorList

	scalewaymachinelog.Info("ValidateUpdate ScalewayMachineValidator Not DryRun", "name", oldObj.GetName())
	equal, diff, err := compare.Diff(oldObj.Spec, newObj.Spec)
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

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (webhook *ScalewayMachineValidator) ValidateDelete(_ context.Context, obj *infrav1.ScalewayMachine) (admission.Warnings, error) {
	scalewaymachinelog.Info("ValidateDelete ScalewayMachineValidator", "name", obj.GetName())
	return nil, nil
}
