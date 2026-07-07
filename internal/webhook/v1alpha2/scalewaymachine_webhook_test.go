package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

var _ = Describe("ScalewayMachine Webhook", func() {
	var (
		obj       *infrav1.ScalewayMachine
		oldObj    *infrav1.ScalewayMachine
		validator ScalewayMachineCustomValidator
	)

	BeforeEach(func() {
		obj = &infrav1.ScalewayMachine{}
		oldObj = &infrav1.ScalewayMachine{}
		validator = ScalewayMachineCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	Context("When creating or updating ScalewayMachine under Validating Webhook", func() {
		It("Should validate updates correctly", func() {
			By("simulating a valid update scenario")
			obj.Spec.ProviderID = "scaleway://hello"
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).To(BeNil())
		})
	})
	Context("When creating or updating ScalewayMachine under Validating Webhook", func() {
		It("Should reject updates correctly", func() {
			By("simulating a bad update scenario")
			oldObj.Spec.ProviderID = "test"
			obj.Spec.ProviderID = "scaleway://hello"
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(HaveOccurred())
		})
	})
})
