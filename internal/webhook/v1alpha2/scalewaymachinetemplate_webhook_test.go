package v1alpha2

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

var _ = Describe("ScalewayMachineTemplate Webhook", func() {
	var (
		obj       *infrav1.ScalewayMachineTemplate
		oldObj    *infrav1.ScalewayMachineTemplate
		validator ScalewayMachineTemplateCustomValidator
	)

	var (
		v1alpha2ScalewayMachine = &infrav1.ScalewayMachine{
			Spec: infrav1.ScalewayMachineSpec{
				ProviderID:     "scaleway://instance/fr-par-1/11111111-1111-1111-1111-111111111111",
				CommercialType: "DEV1-S",
				Image: infrav1.Image{
					IDOrName: infrav1.IDOrName{
						Name: "scaleway-image",
					},
				},
				RootVolume: infrav1.RootVolume{
					Size: 42,
					Type: "block",
					IOPS: 15000,
				},
				AdditionalVolumes: []infrav1.AdditionalVolume{{
					Size: 20,
					Type: "block",
					IOPS: 15000,
				}},
				PublicNetwork: infrav1.PublicNetwork{
					EnableIPv4: ptr.To(true),
					EnableIPv6: ptr.To(false),
				},
				SecurityGroup: infrav1.IDOrName{
					Name: "scaleway-security-group",
				},
			},
			Status: infrav1.ScalewayMachineStatus{
				Initialization: infrav1.ScalewayMachineInitializationStatus{
					Provisioned: ptr.To(true),
				},
			},
		}
		v1alpha2ScalewayMachineTemplate = &infrav1.ScalewayMachineTemplate{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-machine-template",
			},
			Spec: infrav1.ScalewayMachineTemplateSpec{
				Template: infrav1.ScalewayMachineTemplateResource{
					Spec: v1alpha2ScalewayMachine.Spec,
				},
			},
		}
	)

	BeforeEach(func() {
		obj = &infrav1.ScalewayMachineTemplate{}
		oldObj = &infrav1.ScalewayMachineTemplate{}
		validator = ScalewayMachineTemplateCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	Context("When updating scalewayMachinetemplate", func() {
		It("Should reject if spec updated", func() {
			oldObj = v1alpha2ScalewayMachineTemplate
			obj := oldObj.DeepCopy()
			obj.Spec.Template.Spec.Image = infrav1.Image{
				IDOrName: infrav1.IDOrName{
					Name: "scaleway-test",
				},
			}
			ctx := context.Background()
			ctx = admission.NewContextWithRequest(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{DryRun: ptr.To(false)}})
			By("calling the validateUpdate method")
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(HaveOccurred())
		})
		It("Should pass if spec updated in dry-run mode", func() {
			oldObj = v1alpha2ScalewayMachineTemplate
			obj := oldObj.DeepCopy()
			obj.Spec.Template.Spec.Image = infrav1.Image{
				IDOrName: infrav1.IDOrName{
					Name: "scaleway-test",
				},
			}
			obj.SetAnnotations(map[string]string{clusterv1.TopologyDryRunAnnotation: ""})
			ctx := context.Background()
			ctx = admission.NewContextWithRequest(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{DryRun: ptr.To(true)}})
			By("calling the validateUpdate method in dry-run")
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should pass if template metadata well updated", func() {
			oldObj = v1alpha2ScalewayMachineTemplate
			obj := oldObj.DeepCopy()
			obj.Spec.Template.ObjectMeta.Labels = nil
			obj.SetAnnotations(map[string]string{clusterv1.TopologyDryRunAnnotation: ""})
			ctx := context.Background()
			ctx = admission.NewContextWithRequest(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{DryRun: ptr.To(false)}})
			By("calling the validateUpdate method with template metadata changed")
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should failed if template metadata bad update", func() {
			oldObj = v1alpha2ScalewayMachineTemplate
			obj := oldObj.DeepCopy()
			obj.Spec.Template.ObjectMeta = clusterv1.ObjectMeta{
				Labels: map[string]string{
					"foo":          "$invalid-key",
					"bar":          strings.Repeat("a", 64) + "too-long-value",
					"/invalid-key": "foo",
				},
				Annotations: map[string]string{
					"/invalid-key": "foo",
				},
			}
			obj.SetAnnotations(map[string]string{clusterv1.TopologyDryRunAnnotation: ""})
			ctx := context.Background()
			ctx = admission.NewContextWithRequest(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{DryRun: ptr.To(false)}})
			By("calling the validateUpdate method with bad template metadata change")
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(HaveOccurred())
		})
		It("Should failed if add element to template spec ", func() {
			oldObj = v1alpha2ScalewayMachineTemplate
			obj := oldObj.DeepCopy()
			obj.Spec.Template.Spec.PlacementGroup = infrav1.IDOrName{
				Name: "scaleway-placement-group",
			}
			obj.SetAnnotations(map[string]string{clusterv1.TopologyDryRunAnnotation: ""})
			ctx := context.Background()
			ctx = admission.NewContextWithRequest(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{DryRun: ptr.To(false)}})
			By("calling the validateUpdate method with adding element to spec")
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(HaveOccurred())
		})
	})
})
