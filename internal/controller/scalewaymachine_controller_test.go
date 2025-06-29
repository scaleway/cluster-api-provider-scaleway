package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ = Describe("ScalewayMachine Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		scalewaymachine := &infrav1.ScalewayMachine{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ScalewayMachine")
			err := k8sClient.Get(ctx, typeNamespacedName, scalewaymachine)
			if err != nil && errors.IsNotFound(err) {
				resource := &infrav1.ScalewayMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: infrav1.ScalewayMachineSpec{
						Image: infrav1.ImageSpec{
							Label: scw.StringPtr("ubuntu_focal"),
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &infrav1.ScalewayMachine{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ScalewayMachine")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ScalewayMachineReconciler{
				Client:                       k8sClient,
				createScalewayMachineService: newScalewayMachineService,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
