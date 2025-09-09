package controller

import (
	"context"
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("ScalewayManagedControlPlane Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		scalewaymanagedcontrolplane := &infrav1.ScalewayManagedControlPlane{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ScalewayManagedControlPlane")
			err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedcontrolplane)
			if err != nil && errors.IsNotFound(err) {
				resource := &infrav1.ScalewayManagedControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: infrav1.ScalewayManagedControlPlaneSpec{
						Type:    "kapsule",
						Version: "v1.30.0",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &infrav1.ScalewayManagedControlPlane{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ScalewayManagedControlPlane")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ScalewayManagedControlPlaneReconciler{
				Client:                                   k8sClient,
				createScalewayManagedControlPlaneService: newScalewayManagedControlPlaneService,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("ScalewayManagedControlPlane", func() {
	Context("When updating the resource", func() {
		When("Basic control plane", func() {
			const resourceName = "test-resource-1"
			ctx := context.Background()

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			scalewaymanagedcontrolplane := &infrav1.ScalewayManagedControlPlane{}

			BeforeEach(func() {
				By("creating the custom resource for the Kind ScalewayManagedControlPlane")
				err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedcontrolplane)
				if err != nil && errors.IsNotFound(err) {
					resource := &infrav1.ScalewayManagedControlPlane{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: "default",
						},
						Spec: infrav1.ScalewayManagedControlPlaneSpec{
							Type:    "kapsule",
							Version: "v1.30.0",
						},
					}
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				}
			})

			AfterEach(func() {
				resource := &infrav1.ScalewayManagedControlPlane{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance ScalewayManagedControlPlane")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})

			It("should fail to set CNI", func(ctx SpecContext) {
				By("Setting CNI")
				resource := &infrav1.ScalewayManagedControlPlane{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.CNI = scw.StringPtr("calico")
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})
		})

		When("ControlPlaneEndpoint is set", func() {
			const resourceName = "test-resource-2"

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			scalewaymanagedcontrolplane := &infrav1.ScalewayManagedControlPlane{}

			BeforeEach(func(ctx SpecContext) {
				By("creating the custom resource for the Kind ScalewayManagedControlPlane")
				err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedcontrolplane)
				if err != nil && errors.IsNotFound(err) {
					resource := &infrav1.ScalewayManagedControlPlane{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: "default",
						},
						Spec: infrav1.ScalewayManagedControlPlaneSpec{
							Type:    "kapsule",
							Version: "v1.30.0",
							ControlPlaneEndpoint: clusterv1.APIEndpoint{
								Host: "42.42.42.42",
								Port: 6443,
							},
						},
					}
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				}
			})

			AfterEach(func(ctx SpecContext) {
				resource := &infrav1.ScalewayManagedControlPlane{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance ScalewayManagedControlPlane")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})

			It("should fail to update host", func(ctx SpecContext) {
				By("Updating the host")
				resource := &infrav1.ScalewayManagedControlPlane{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint.Host = "33.33.33.33"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to update port", func(ctx SpecContext) {
				By("Updating the port")
				resource := &infrav1.ScalewayManagedControlPlane{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint.Port = 443
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to remove ControlPlaneEndpoint", func(ctx SpecContext) {
				By("Removing ControlPlaneEndpoint")
				resource := &infrav1.ScalewayManagedControlPlane{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{}
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})
		})
	})
})

func TestScalewayManagedControlPlaneReconciler_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		createScalewayManagedControlPlaneService scalewayManagedControlPlaneServiceCreator
	}
	type args struct {
		ctx context.Context
		req ctrl.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ctrl.Result
		wantErr bool
		objects []client.Object
		asserts func(g *WithT, c client.Client)
	}{
		{
			name: "should reconcile normally",
			fields: fields{
				createScalewayManagedControlPlaneService: func(managedControlPlaneScope *scope.ManagedControlPlane) *scalewayManagedControlPlaneService {
					return &scalewayManagedControlPlaneService{
						scope:     managedControlPlaneScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayManagedControlPlaneNamespacedName},
			},
			want: reconcile.Result{},
			objects: []client.Object{
				&infrav1.ScalewayManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayManagedClusterNamespacedName.Name,
						Namespace: scalewayManagedClusterNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       clusterNamespacedName.Name,
								Kind:       "Cluster",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
					},
					Spec: infrav1.ScalewayManagedClusterSpec{
						Region:             "fr-par",
						ScalewaySecretName: secretNamespacedName.Name,
						ProjectID:          "11111111-1111-1111-1111-111111111111",
					},
					Status: infrav1.ScalewayManagedClusterStatus{
						Ready: true,
					},
				},
				&infrav1.ScalewayManagedControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayManagedControlPlaneNamespacedName.Name,
						Namespace: scalewayManagedControlPlaneNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       clusterNamespacedName.Name,
								Kind:       "Cluster",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
					},
					Spec: infrav1.ScalewayManagedControlPlaneSpec{
						Type:    "kapsule",
						Version: "v1.30.0",
						ControlPlaneEndpoint: clusterv1.APIEndpoint{
							Host: managedEndpoint.Host,
							Port: managedEndpoint.Port,
						},
					},
				},
				&clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterNamespacedName.Name,
						Namespace: clusterNamespacedName.Namespace,
					},
					Spec: clusterv1.ClusterSpec{
						ControlPlaneRef: &corev1.ObjectReference{
							Name:      scalewayManagedControlPlaneNamespacedName.Name,
							Namespace: scalewayManagedControlPlaneNamespacedName.Namespace,
						},
						InfrastructureRef: &corev1.ObjectReference{
							Name:      scalewayManagedClusterNamespacedName.Name,
							Namespace: scalewayManagedClusterNamespacedName.Namespace,
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretNamespacedName.Name,
						Namespace: secretNamespacedName.Namespace,
					},
					Data: map[string][]byte{
						scw.ScwAccessKeyEnv: []byte("SCWXXXXXXXXXXXXXXXXX"),
						scw.ScwSecretKeyEnv: []byte("11111111-1111-1111-1111-111111111111"),
					},
				},
			},
			asserts: func(g *WithT, c client.Client) {
				// ScalewayManagedControlPlane checks
				smcp := &infrav1.ScalewayManagedControlPlane{}
				g.Expect(c.Get(context.TODO(), scalewayManagedControlPlaneNamespacedName, smcp)).To(Succeed())
				g.Expect(smcp.Status.Ready).To(BeTrue())
				g.Expect(smcp.Status.Initialized).To(BeTrue())
				g.Expect(smcp.Status.ExternalManagedControlPlane).To(BeTrue())
				g.Expect(smcp.Finalizers).To(ContainElement(infrav1.ManagedControlPlaneFinalizer))
			},
		},
		{
			name: "should reconcile deletion",
			fields: fields{
				createScalewayManagedControlPlaneService: func(managedControlPlaneScope *scope.ManagedControlPlane) *scalewayManagedControlPlaneService {
					return &scalewayManagedControlPlaneService{
						scope:     managedControlPlaneScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayManagedControlPlaneNamespacedName},
			},
			want: reconcile.Result{},
			objects: []client.Object{
				&infrav1.ScalewayManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayManagedClusterNamespacedName.Name,
						Namespace: scalewayManagedClusterNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       clusterNamespacedName.Name,
								Kind:       "Cluster",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
					},
					Spec: infrav1.ScalewayManagedClusterSpec{
						Region:             "fr-par",
						ScalewaySecretName: secretNamespacedName.Name,
						ProjectID:          "11111111-1111-1111-1111-111111111111",
					},
					Status: infrav1.ScalewayManagedClusterStatus{
						Ready: true,
					},
				},
				&infrav1.ScalewayManagedControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayManagedControlPlaneNamespacedName.Name,
						Namespace: scalewayManagedControlPlaneNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       clusterNamespacedName.Name,
								Kind:       "Cluster",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
						Finalizers:        []string{infrav1.ManagedControlPlaneFinalizer},
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
					Spec: infrav1.ScalewayManagedControlPlaneSpec{
						Type:    "kapsule",
						Version: "v1.30.0",
						ControlPlaneEndpoint: clusterv1.APIEndpoint{
							Host: managedEndpoint.Host,
							Port: managedEndpoint.Port,
						},
					},
				},
				&clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterNamespacedName.Name,
						Namespace: clusterNamespacedName.Namespace,
					},
					Spec: clusterv1.ClusterSpec{
						ControlPlaneRef: &corev1.ObjectReference{
							Name:      scalewayManagedControlPlaneNamespacedName.Name,
							Namespace: scalewayManagedControlPlaneNamespacedName.Namespace,
						},
						InfrastructureRef: &corev1.ObjectReference{
							Name:      scalewayManagedClusterNamespacedName.Name,
							Namespace: scalewayManagedClusterNamespacedName.Namespace,
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretNamespacedName.Name,
						Namespace: secretNamespacedName.Namespace,
					},
					Data: map[string][]byte{
						scw.ScwAccessKeyEnv: []byte("SCWXXXXXXXXXXXXXXXXX"),
						scw.ScwSecretKeyEnv: []byte("11111111-1111-1111-1111-111111111111"),
					},
				},
			},
			asserts: func(g *WithT, c client.Client) {
				// ScalewayManagedControlPlane should not exist anymore if the finalizer was correctly removed.
				smcp := &infrav1.ScalewayManagedControlPlane{}
				g.Expect(c.Get(context.TODO(), scalewayManagedControlPlaneNamespacedName, smcp)).NotTo(Succeed())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := NewWithT(t)
			sb := runtime.NewSchemeBuilder(
				corev1.AddToScheme,
				clusterv1.AddToScheme,
				infrav1.AddToScheme,
			)
			s := runtime.NewScheme()

			g.Expect(sb.AddToScheme(s)).To(Succeed())

			runtimeObjects := make([]runtime.Object, 0, len(tt.objects))
			for _, obj := range tt.objects {
				runtimeObjects = append(runtimeObjects, obj)
			}

			c := fake.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(runtimeObjects...).
				WithStatusSubresource(tt.objects...).
				Build()

			r := &ScalewayManagedControlPlaneReconciler{
				Client:                                   c,
				createScalewayManagedControlPlaneService: tt.fields.createScalewayManagedControlPlaneService,
			}
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScalewayManagedControlPlaneReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScalewayManagedControlPlaneReconciler.Reconcile() = %v, want %v", got, tt.want)
			}
			tt.asserts(g, c)
		})
	}
}
