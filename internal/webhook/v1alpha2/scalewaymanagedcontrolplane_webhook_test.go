package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

var _ = Describe("ScalewayManagedControlPlane Webhook", func() {
	var (
		obj       *infrav1.ScalewayManagedControlPlane
		oldObj    *infrav1.ScalewayManagedControlPlane
		defaulter ScalewayManagedControlPlaneCustomDefaulter
	)

	BeforeEach(func() {
		obj = &infrav1.ScalewayManagedControlPlane{}
		oldObj = &infrav1.ScalewayManagedControlPlane{}
		defaulter = ScalewayManagedControlPlaneCustomDefaulter{}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
<<<<<<< HEAD
	})

	AfterEach(func() {
		// TODO (user): Add any teardown logic common to all tests
=======
>>>>>>> tmp-original-13-02-26-16-17
	})

	Context("When creating ScalewayManagedControlPlane under Defaulting Webhook", func() {
		It("Should initialize clusterName if not set", func() {
			obj.Name = "mycluster"
			obj.Namespace = "default"
			By("calling the Default method to apply defaults")
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			By("checking that the default values are set")
			Expect(obj.Spec.ClusterName).To(Equal("default-mycluster"))
		})

		It("Should keep existing clusterName", func() {
			obj.Name = "mycluster"
			obj.Namespace = "default"
			obj.Spec.ClusterName = "test-cluster-1"
			By("calling the Default method to apply defaults")
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			By("checking that the default values are set")
			Expect(obj.Spec.ClusterName).To(Equal("test-cluster-1"))
		})
	})
})
