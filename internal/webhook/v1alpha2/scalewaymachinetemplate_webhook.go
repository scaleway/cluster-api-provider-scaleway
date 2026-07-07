package v1alpha2

import (
	"context"
	"fmt"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/util/compare"
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

// ScalewayMachineTemplateCustomValidator struct is responsible for validating the ScalewayMachineTemplate resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ScalewayMachineTemplateCustomValidator struct{}

// SetupScalewayMachineTemplateWebhookWithManager registers the webhook for ScalewayMachineTemplate in the manager.
func SetupScalewayMachineTemplateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &infrav1.ScalewayMachineTemplate{}).
		WithValidator(&ScalewayMachineTemplateCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha2-scalewaymachinetemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=scalewaymachinetemplates,verbs=create;update,versions=v1alpha2,name=vscalewaymachinetemplate-v1alpha2.kb.io,admissionReviewVersions=v1

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ScalewayMachineTemplate.
func (v *ScalewayMachineTemplateCustomValidator) ValidateCreate(_ context.Context, obj *infrav1.ScalewayMachineTemplate) (admission.Warnings, error) {
	scalewaymachinetemplatelog.Info("Validation for ScalewayMachineTemplate upon creation", "name", obj.GetName())
	// Validate the metadata of the template.
	allErrs := obj.Spec.Template.ObjectMeta.Validate(field.NewPath("spec", "template", "metadata"))
	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(infrav1.GroupVersion.WithKind("ScalewayMachineTemplate").GroupKind(), obj.Name, allErrs)
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ScalewayMachineTemplate.
func (v *ScalewayMachineTemplateCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *infrav1.ScalewayMachineTemplate) (admission.Warnings, error) {
	scalewaymachinetemplatelog.Info("Validation for ScalewayMachineTemplate upon update", "name", newObj.GetName())
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

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ScalewayMachineTemplate.
func (v *ScalewayMachineTemplateCustomValidator) ValidateDelete(_ context.Context, obj *infrav1.ScalewayMachineTemplate) (admission.Warnings, error) {
	scalewaymachinetemplatelog.Info("Validation for ScalewayMachineTemplate upon deletion", "name", obj.GetName())
	return nil, nil
}
