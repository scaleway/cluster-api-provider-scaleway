package v1alpha2

import (
	"context"
	"fmt"

	compare "github.com/scaleway/cluster-api-provider-scaleway/internal/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api/util/topology"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

// nolint:unused
// log is for logging in this package.
var scalewaymachinetemplatelog = logf.Log.WithName("scalewaymachinetemplate-resource")

// ScalewayMachineTemplate implements a custom validation webhook for ScalewayMachineTemplate.
// +kubebuilder:object:generate=false
type ScalewayMachineTemplateValidator struct{}

// SetupScalewayMachineTemplateWebhookWithManager registers the webhook for ScalewayMachineTemplate in the manager.
func SetupScalewayMachineTemplateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &infrav1.ScalewayMachineTemplate{}).
		WithValidator(&ScalewayMachineTemplateValidator{}).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1alpha2-scalewaymachinetemplate,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=scalewaymachinetemplates,versions=v1alpha2,name=validation.scalewaymachinetemplate.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1

var _ admission.Validator[*infrav1.ScalewayMachineTemplate] = &ScalewayMachineTemplateValidator{}

func (webhook *ScalewayMachineTemplateValidator) ValidateCreate(_ context.Context, obj *infrav1.ScalewayMachineTemplate) (admission.Warnings, error) {
	scalewaymachinetemplatelog.Info("ValidateCreate ScalewayMachineTemplateValidator", "name", obj.GetName())
	// Validate the metadata of the template.
	allErrs := obj.Spec.Template.ObjectMeta.Validate(field.NewPath("spec", "template", "metadata"))
	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(infrav1.GroupVersion.WithKind("ScalewayMachineTemplate").GroupKind(), obj.Name, allErrs)
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (webhook *ScalewayMachineTemplateValidator) ValidateUpdate(ctx context.Context, oldObj *infrav1.ScalewayMachineTemplate, newObj *infrav1.ScalewayMachineTemplate) (admission.Warnings, error) {
	scalewaymachinetemplatelog.Info("ValidateUpdate ScalewayMachineTemplateValidator", "name", oldObj.GetName())
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a admission.Request inside context: %v", err))
	}

	var allErrs field.ErrorList
	if !topology.IsDryRunRequest(req, newObj) {
		scalewaymachinetemplatelog.Info("ValidateUpdate ScalewayMachineTemplateValidator Not DryRun", "name", oldObj.GetName())
		equal, diff, err := compare.Diff(oldObj.Spec.Template.Spec, newObj.Spec.Template.Spec)
		if err != nil {
			return nil, apierrors.NewBadRequest(fmt.Sprintf("failed to compare old and new ScalewayMachineTemplate: %v", err))
		}
		if !equal {
			allErrs = append(allErrs,
				field.Invalid(field.NewPath("spec", "template", "spec"), newObj, fmt.Sprintf("ScalewayMachineTemplate spec.template.spec field is immutable. Please create a new resource instead. Diff: %s", diff)),
			)
		}
	} else {
		scalewaymachinetemplatelog.Info("ValidateUpdate ScalewayMachineTemplateValidator DryRun", "name", oldObj.GetName())
	}

	// Validate the metadata of the template.
	allErrs = append(allErrs, newObj.Spec.Template.ObjectMeta.Validate(field.NewPath("spec", "template", "metadata"))...)

	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(infrav1.GroupVersion.WithKind("ScalewayMachineTemplate").GroupKind(), newObj.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (webhook *ScalewayMachineTemplateValidator) ValidateDelete(_ context.Context, obj *infrav1.ScalewayMachineTemplate) (admission.Warnings, error) {
	scalewaymachinetemplatelog.Info("ValidateDelete ScalewayMachineTemplateValidator", "name", obj.GetName())
	return nil, nil
}
