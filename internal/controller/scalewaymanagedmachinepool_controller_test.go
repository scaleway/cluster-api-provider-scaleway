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
	expclusterv1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("ScalewayManagedMachinePool Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		scalewaymanagedmachinepool := &infrav1.ScalewayManagedMachinePool{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ScalewayManagedMachinePool")
			err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedmachinepool)
			if err != nil && errors.IsNotFound(err) {
				resource := &infrav1.ScalewayManagedMachinePool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: infrav1.ScalewayManagedMachinePoolSpec{
						NodeType: "DEV1-S",
						Zone:     string(scw.ZoneFrPar1),
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &infrav1.ScalewayManagedMachinePool{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ScalewayManagedMachinePool")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ScalewayManagedMachinePoolReconciler{
				Client:                                  k8sClient,
				createScalewayManagedMachinePoolService: newScalewayManagedMachinePoolService,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("ScalewayManagedMachinePool", func() {
	Context("When updating the resource", func() {
		When("Basic machine pool", func() {
			const resourceName = "test-resource-1"
			ctx := context.Background()

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			scalewaymanagedmachinepool := &infrav1.ScalewayManagedMachinePool{}

			BeforeEach(func() {
				By("creating the custom resource for the Kind ScalewayManagedMachinePool")
				err := k8sClient.Get(ctx, typeNamespacedName, scalewaymanagedmachinepool)
				if err != nil && errors.IsNotFound(err) {
					resource := &infrav1.ScalewayManagedMachinePool{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: "default",
						},
						Spec: infrav1.ScalewayManagedMachinePoolSpec{
							NodeType: "DEV1-S",
							Zone:     string(scw.ZoneFrPar1),
						},
					}
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				}
			})

			AfterEach(func() {
				resource := &infrav1.ScalewayManagedMachinePool{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance ScalewayManagedMachinePool")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			})

			It("should fail to update Node Type", func(ctx SpecContext) {
				By("Setting Node Type")
				resource := &infrav1.ScalewayManagedMachinePool{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.NodeType = "DEV1-M"
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})

			It("should fail to update Zone", func(ctx SpecContext) {
				By("Setting Zone")
				resource := &infrav1.ScalewayManagedMachinePool{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				Expect(err).NotTo(HaveOccurred())

				resource.Spec.Zone = string(scw.ZoneFrPar2)
				Expect(k8sClient.Update(ctx, resource)).NotTo(Succeed())
			})
		})
	})
})

var (
	scalewayManagedMachinePoolNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "scalewaymanagedmachinepool",
	}
	machinePoolNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "machinepool",
	}
)

func TestScalewayManagedMachinePoolReconciler_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		createScalewayManagedMachinePoolService scalewayManagedMachinePoolServiceCreator
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
				createScalewayManagedMachinePoolService: func(managedMachinePoolScope *scope.ManagedMachinePool) *scalewayManagedMachinePoolService {
					return &scalewayManagedMachinePoolService{
						scope:     managedMachinePoolScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayManagedMachinePoolNamespacedName},
			},
			want: reconcile.Result{},
			objects: []client.Object{
				&expclusterv1.MachinePool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      machinePoolNamespacedName.Name,
						Namespace: machinePoolNamespacedName.Namespace,
						Labels: map[string]string{
							clusterv1.ClusterNameLabel: clusterNamespacedName.Name,
						},
					},
					Spec: expclusterv1.MachinePoolSpec{
						ClusterName: clusterNamespacedName.Name,
						Template: clusterv1.MachineTemplateSpec{
							Spec: clusterv1.MachineSpec{
								ClusterName: clusterNamespacedName.Name,
								InfrastructureRef: corev1.ObjectReference{
									Name:      scalewayManagedMachinePoolNamespacedName.Name,
									Namespace: scalewayManagedMachinePoolNamespacedName.Namespace,
								},
							},
						},
					},
				},
				&infrav1.ScalewayManagedMachinePool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayManagedMachinePoolNamespacedName.Name,
						Namespace: scalewayManagedMachinePoolNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       machinePoolNamespacedName.Name,
								Kind:       "MachinePool",
								APIVersion: expclusterv1.GroupVersion.String(),
							},
						},
					},
					Spec: infrav1.ScalewayManagedMachinePoolSpec{
						NodeType: "DEV1-S",
						Zone:     scw.ZoneFrPar1.String(),
					},
				},
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
				// ScalewayManagedMachinePool checks
				smmp := &infrav1.ScalewayManagedMachinePool{}
				g.Expect(c.Get(context.TODO(), scalewayManagedMachinePoolNamespacedName, smmp)).To(Succeed())
				g.Expect(smmp.Status.Ready).To(BeTrue())
				g.Expect(smmp.Finalizers).To(ContainElement(infrav1.ManagedMachinePoolFinalizer))
			},
		},
		{
			name: "should reconcile deletion",
			fields: fields{
				createScalewayManagedMachinePoolService: func(managedMachinePoolScope *scope.ManagedMachinePool) *scalewayManagedMachinePoolService {
					return &scalewayManagedMachinePoolService{
						scope:     managedMachinePoolScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{NamespacedName: scalewayManagedMachinePoolNamespacedName},
			},
			want: reconcile.Result{},
			objects: []client.Object{
				&expclusterv1.MachinePool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      machinePoolNamespacedName.Name,
						Namespace: machinePoolNamespacedName.Namespace,
						Labels: map[string]string{
							clusterv1.ClusterNameLabel: clusterNamespacedName.Name,
						},
					},
					Spec: expclusterv1.MachinePoolSpec{
						ClusterName: clusterNamespacedName.Name,
						Template: clusterv1.MachineTemplateSpec{
							Spec: clusterv1.MachineSpec{
								ClusterName: clusterNamespacedName.Name,
								InfrastructureRef: corev1.ObjectReference{
									Name:      scalewayManagedMachinePoolNamespacedName.Name,
									Namespace: scalewayManagedMachinePoolNamespacedName.Namespace,
								},
							},
						},
					},
				},
				&infrav1.ScalewayManagedMachinePool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayManagedMachinePoolNamespacedName.Name,
						Namespace: scalewayManagedMachinePoolNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       machinePoolNamespacedName.Name,
								Kind:       "MachinePool",
								APIVersion: expclusterv1.GroupVersion.String(),
							},
						},
						Finalizers:        []string{infrav1.ManagedMachinePoolFinalizer},
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
					Spec: infrav1.ScalewayManagedMachinePoolSpec{
						NodeType: "DEV1-S",
						Zone:     scw.ZoneFrPar1.String(),
					},
				},
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
				// ScalewayManagedMachinePool should not exist anymore if the finalizer was correctly removed.
				smmp := &infrav1.ScalewayManagedMachinePool{}
				g.Expect(c.Get(context.TODO(), scalewayManagedMachinePoolNamespacedName, smmp)).NotTo(Succeed())
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
				expclusterv1.AddToScheme,
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

			r := &ScalewayManagedMachinePoolReconciler{
				Client:                                  c,
				createScalewayManagedMachinePoolService: tt.fields.createScalewayManagedMachinePoolService,
			}
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScalewayManagedMachinePoolReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScalewayManagedMachinePoolReconciler.Reconcile() = %v, want %v", got, tt.want)
			}
			tt.asserts(g, c)
		})
	}
}
