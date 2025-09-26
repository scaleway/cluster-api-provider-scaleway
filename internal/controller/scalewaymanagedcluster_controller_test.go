package controller

import (
	"context"
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
)

var _ = Describe("ScalewayManagedCluster Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		scalewaymanagedcluster := &infrav1.ScalewayManagedCluster{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ScalewayManagedCluster")
			err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedcluster)
			if err != nil && errors.IsNotFound(err) {
				resource := &infrav1.ScalewayManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: infrav1.ScalewayManagedClusterSpec{
						Region:             infrav1.ScalewayRegion(scw.RegionFrPar),
						ProjectID:          "11111111-1111-1111-1111-111111111111",
						ScalewaySecretName: "test-secret",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &infrav1.ScalewayManagedCluster{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ScalewayManagedCluster")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ScalewayManagedClusterReconciler{
				Client:                              k8sClient,
				createScalewayManagedClusterService: newScalewayManagedClusterService,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("ScalewayManagedCluster", func() {
	Context("When updating the resource", func() {
		When("Basic cluster", func() {
			const resourceName = "test-resource-1"
			ctx := context.Background()

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			scalewaymanagedcluster := &infrav1.ScalewayManagedCluster{}

			BeforeEach(func() {
				By("creating the custom resource for the Kind ScalewayManagedCluster")
				err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedcluster)
				if err != nil && errors.IsNotFound(err) {
					resource := &infrav1.ScalewayManagedCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: "default",
						},
						Spec: infrav1.ScalewayManagedClusterSpec{
							Region:             infrav1.ScalewayRegion(scw.RegionFrPar),
							ProjectID:          "11111111-1111-1111-1111-111111111111",
							ScalewaySecretName: "test-secret",
						},
					}
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				}
			})

			AfterEach(func() {
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance ScalewayManagedCluster")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})

			It("should fail to update projectID", func(ctx SpecContext) {
				By("Updating the projectID")
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ProjectID = "11111111-1111-1111-1111-111111111110"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to update region", func(ctx SpecContext) {
				By("Updating the region")
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.Region = "nl-ams"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should succeed to update scaleway secret name", func(ctx SpecContext) {
				By("Updating the region")
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ScalewaySecretName = "my-other-secret"
				Expect(k8sClient.Update(ctx, resource)).To(Succeed())
			})

			It("should fail to set private network params", func(ctx SpecContext) {
				By("Setting private network params")
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.Network = infrav1.ScalewayManagedClusterNetwork{
					PrivateNetwork: infrav1.PrivateNetwork{
						ID: "11111111-1111-1111-1111-111111111111",
					},
				}
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})
		})

		When("ControlPlaneEndpoint is set", func() {
			const resourceName = "test-resource-2"

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			scalewaymanagedcluster := &infrav1.ScalewayManagedCluster{}

			BeforeEach(func(ctx SpecContext) {
				By("creating the custom resource for the Kind ScalewayManagedCluster")
				err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedcluster)
				if err != nil && errors.IsNotFound(err) {
					resource := &infrav1.ScalewayManagedCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: "default",
						},
						Spec: infrav1.ScalewayManagedClusterSpec{
							ProjectID:          "11111111-1111-1111-1111-111111111111",
							Region:             infrav1.ScalewayRegion(scw.RegionFrPar),
							ScalewaySecretName: "secret",
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
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance ScalewayManagedCluster")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})

			It("should fail to update host", func(ctx SpecContext) {
				By("Updating the host")
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint.Host = "22.22.22.22"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to update port", func(ctx SpecContext) {
				By("Updating the port")
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint.Port = 443
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to remove ControlPlaneEndpoint", func(ctx SpecContext) {
				By("Removing ControlPlaneEndpoint")
				resource := &infrav1.ScalewayManagedCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{}
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})
		})
	})
})

var (
	managedEndpoint = clusterv1.APIEndpoint{
		Host: "clusterid.api.k8s.fr-par.scw.cloud",
		Port: 6443,
	}
	scalewayManagedClusterNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "scalewaymanagedcluster",
	}
	scalewayManagedControlPlaneNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "scalewaymanagedcontrolplane",
	}
)

func TestScalewayManagedClusterReconciler_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		createScalewayManagedClusterService scalewayManagedClusterServiceCreator
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
				createScalewayManagedClusterService: func(managedClusterScope *scope.ManagedCluster) *scalewayManagedClusterService {
					return &scalewayManagedClusterService{
						scope:     managedClusterScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayManagedClusterNamespacedName},
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
				},
				&infrav1.ScalewayManagedControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayManagedControlPlaneNamespacedName.Name,
						Namespace: scalewayManagedControlPlaneNamespacedName.Namespace,
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
						ControlPlaneRef: clusterv1.ContractVersionedObjectReference{
							Name: scalewayManagedControlPlaneNamespacedName.Name,
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
				// ScalewayManagedCluster checks
				sc := &infrav1.ScalewayManagedCluster{}
				g.Expect(c.Get(context.TODO(), scalewayManagedClusterNamespacedName, sc)).To(Succeed())
				g.Expect(sc.Status.Initialization.Provisioned).NotTo(BeNil())
				g.Expect(*sc.Status.Initialization.Provisioned).To(BeTrue())
				g.Expect(sc.Spec.ControlPlaneEndpoint.Host).To(Equal(managedEndpoint.Host))
				g.Expect(sc.Spec.ControlPlaneEndpoint.Port).To(Equal(managedEndpoint.Port))
				g.Expect(sc.Finalizers).To(ContainElement(infrav1.ScalewayManagedClusterFinalizer))

				// Secret checks
				s := &corev1.Secret{}
				g.Expect(c.Get(context.TODO(), secretNamespacedName, s)).To(Succeed())
				g.Expect(s.Finalizers).To(ContainElement(SecretFinalizer))
				g.Expect(s.OwnerReferences).NotTo(BeEmpty())
			},
		},
		{
			name: "should reconcile deletion",
			fields: fields{
				createScalewayManagedClusterService: func(managedClusterScope *scope.ManagedCluster) *scalewayManagedClusterService {
					return &scalewayManagedClusterService{
						scope:     managedClusterScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayManagedClusterNamespacedName},
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
						Finalizers:        []string{infrav1.ScalewayManagedClusterFinalizer},
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
					Spec: infrav1.ScalewayManagedClusterSpec{
						Region:             "fr-par",
						ScalewaySecretName: secretNamespacedName.Name,
						ProjectID:          "11111111-1111-1111-1111-111111111111",
					},
				},
				&clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterNamespacedName.Name,
						Namespace: clusterNamespacedName.Namespace,
					},
					Spec: clusterv1.ClusterSpec{
						ControlPlaneRef: clusterv1.ContractVersionedObjectReference{
							Name: scalewayManagedControlPlaneNamespacedName.Name,
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
				// ScalewayManagedCluster should not exist anymore if the finalizer was correctly removed.
				sc := &infrav1.ScalewayManagedCluster{}
				g.Expect(c.Get(context.TODO(), scalewayManagedClusterNamespacedName, sc)).NotTo(Succeed())

				// Secret checks
				s := &corev1.Secret{}
				g.Expect(c.Get(context.TODO(), secretNamespacedName, s)).To(Succeed())
				g.Expect(s.Finalizers).NotTo(ContainElement(SecretFinalizer))
				g.Expect(s.OwnerReferences).To(BeEmpty())
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

			r := &ScalewayManagedClusterReconciler{
				Client:                              c,
				createScalewayManagedClusterService: tt.fields.createScalewayManagedClusterService,
			}
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScalewayManagedClusterReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScalewayManagedClusterReconciler.Reconcile() = %v, want %v", got, tt.want)
			}
			tt.asserts(g, c)
		})
	}
}
