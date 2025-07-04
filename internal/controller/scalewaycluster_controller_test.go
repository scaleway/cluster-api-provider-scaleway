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

var _ = Describe("ScalewayCluster Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		scalewaycluster := &infrav1.ScalewayCluster{}

		BeforeEach(func(ctx SpecContext) {
			By("creating the custom resource for the Kind ScalewayCluster")
			err := k8sClient.Get(ctx, typeNamespacedName, scalewaycluster)
			if err != nil && errors.IsNotFound(err) {
				resource := &infrav1.ScalewayCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: infrav1.ScalewayClusterSpec{
						ProjectID: "11111111-1111-1111-1111-111111111111",
						Region:    string(scw.RegionFrPar),
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func(ctx SpecContext) {
			resource := &infrav1.ScalewayCluster{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ScalewayCluster")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func(ctx SpecContext) {
			By("Reconciling the created resource")
			controllerReconciler := &ScalewayClusterReconciler{
				Client:                       k8sClient,
				createScalewayClusterService: newScalewayClusterService,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("ScalewayCluster", func() {
	Context("When updating the resource", func() {
		When("Basic cluster", func() {
			const resourceName = "test-resource-1"

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			scalewaycluster := &infrav1.ScalewayCluster{}

			BeforeEach(func(ctx SpecContext) {
				By("creating the custom resource for the Kind ScalewayCluster")
				err := k8sClient.Get(ctx, typeNamespacedName, scalewaycluster)
				if err != nil && errors.IsNotFound(err) {
					resource := &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							ProjectID:          "11111111-1111-1111-1111-111111111111",
							Region:             string(scw.RegionFrPar),
							ScalewaySecretName: "my-secret",
						},
					}
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				}
			})

			AfterEach(func(ctx SpecContext) {
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance ScalewayCluster")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})

			It("should fail to update projectID", func(ctx SpecContext) {
				By("Updating the projectID")
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ProjectID = "11111111-1111-1111-1111-111111111110"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to update region", func(ctx SpecContext) {
				By("Updating the region")
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.Region = "nl-ams"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should succeed to update scaleway secret name", func(ctx SpecContext) {
				By("Updating the region")
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ScalewaySecretName = "my-other-secret"
				Expect(k8sClient.Update(ctx, resource)).To(Succeed())
			})

			It("should fail to enable private network", func(ctx SpecContext) {
				By("Enabling private network")
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.Network = &infrav1.NetworkSpec{
					PrivateNetwork: &infrav1.PrivateNetworkSpec{
						Enabled: true,
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
			scalewaycluster := &infrav1.ScalewayCluster{}

			BeforeEach(func(ctx SpecContext) {
				By("creating the custom resource for the Kind ScalewayCluster")
				err := k8sClient.Get(ctx, typeNamespacedName, scalewaycluster)
				if err != nil && errors.IsNotFound(err) {
					resource := &infrav1.ScalewayCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: "default",
						},
						Spec: infrav1.ScalewayClusterSpec{
							ProjectID: "11111111-1111-1111-1111-111111111111",
							Region:    string(scw.RegionFrPar),
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
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance ScalewayCluster")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})

			It("should fail to update host", func(ctx SpecContext) {
				By("Updating the host")
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint.Host = "12.12.12.12"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to update port", func(ctx SpecContext) {
				By("Updating the port")
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint.Port = 443
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to remove ControlPlaneEndpoint", func(ctx SpecContext) {
				By("Removing ControlPlaneEndpoint")
				resource := &infrav1.ScalewayCluster{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{}
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})
		})
	})
})

var (
	scalewayClusterNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "scalewaycluster",
	}
	clusterNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "cluster",
	}
	secretNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "scaleway-secret",
	}
)

func TestScalewayClusterReconciler_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		createScalewayClusterService scalewayClusterServiceCreator
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
				createScalewayClusterService: func(clusterScope *scope.Cluster) *scalewayClusterService {
					return &scalewayClusterService{
						scope: clusterScope,
						Reconcile: func(ctx context.Context) error {
							clusterScope.ScalewayCluster.Status.Network = &infrav1.NetworkStatus{
								LoadBalancerIP: scw.StringPtr("42.42.42.42"),
							}
							return nil
						},
						Delete: func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayClusterNamespacedName},
			},
			want: reconcile.Result{},
			objects: []client.Object{
				&infrav1.ScalewayCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayClusterNamespacedName.Name,
						Namespace: scalewayClusterNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       clusterNamespacedName.Name,
								Kind:       "Cluster",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
					},
					Spec: infrav1.ScalewayClusterSpec{
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
				// ScalewayCluster checks
				sc := &infrav1.ScalewayCluster{}
				g.Expect(c.Get(context.TODO(), scalewayClusterNamespacedName, sc)).To(Succeed())
				g.Expect(sc.Status.Ready).To(BeTrue())
				g.Expect(sc.Spec.ControlPlaneEndpoint.Host).NotTo(BeEmpty())
				g.Expect(sc.Spec.ControlPlaneEndpoint.Port).NotTo(BeZero())
				g.Expect(sc.Finalizers).To(ContainElement(infrav1.ClusterFinalizer))

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
				createScalewayClusterService: func(clusterScope *scope.Cluster) *scalewayClusterService {
					return &scalewayClusterService{
						scope:     clusterScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayClusterNamespacedName},
			},
			want: reconcile.Result{},
			objects: []client.Object{
				&infrav1.ScalewayCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayClusterNamespacedName.Name,
						Namespace: scalewayClusterNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       clusterNamespacedName.Name,
								Kind:       "Cluster",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
						Finalizers:        []string{infrav1.ClusterFinalizer},
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
					Spec: infrav1.ScalewayClusterSpec{
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
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:       secretNamespacedName.Name,
						Namespace:  secretNamespacedName.Namespace,
						Finalizers: []string{SecretFinalizer},
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       scalewayClusterNamespacedName.Name,
								Kind:       "ScalewayCluster",
								APIVersion: infrav1.GroupVersion.String(),
							},
						},
					},
					Data: map[string][]byte{
						scw.ScwAccessKeyEnv: []byte("SCWXXXXXXXXXXXXXXXXX"),
						scw.ScwSecretKeyEnv: []byte("11111111-1111-1111-1111-111111111111"),
					},
				},
			},
			asserts: func(g *WithT, c client.Client) {
				// ScalewayCluster should not exist anymore if the finalizer was correctly removed.
				sc := &infrav1.ScalewayCluster{}
				g.Expect(c.Get(context.TODO(), scalewayClusterNamespacedName, sc)).NotTo(Succeed())

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

			r := &ScalewayClusterReconciler{
				Client:                       c,
				createScalewayClusterService: tt.fields.createScalewayClusterService,
			}

			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScalewayClusterReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScalewayClusterReconciler.Reconcile() = %v, want %v", got, tt.want)
			}

			tt.asserts(g, c)
		})
	}
}
