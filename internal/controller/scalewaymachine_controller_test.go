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

var _ = Describe("ScalewayMachine Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
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
		})
	})
})

var (
	scalewayMachineNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "scalewaymachine",
	}
	machineNamespacedName = types.NamespacedName{
		Namespace: "caps",
		Name:      "machine",
	}
)

func TestScalewayMachineReconciler_Reconcile(t *testing.T) {
	t.Parallel()
	type fields struct {
		createScalewayMachineService scalewayMachineServiceCreator
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
				createScalewayMachineService: func(machineScope *scope.Machine) *scalewayMachineService {
					return &scalewayMachineService{
						scope:     machineScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{
					NamespacedName: scalewayMachineNamespacedName,
				},
			},
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
					Spec: clusterv1.ClusterSpec{
						InfrastructureRef: &corev1.ObjectReference{
							Name: scalewayClusterNamespacedName.Name,
						},
					},
					Status: clusterv1.ClusterStatus{
						InfrastructureReady: true,
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
				&infrav1.ScalewayMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayMachineNamespacedName.Name,
						Namespace: scalewayMachineNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       machineNamespacedName.Name,
								Kind:       "Machine",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
					},
				},
				&clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      machineNamespacedName.Name,
						Namespace: machineNamespacedName.Namespace,
						Labels: map[string]string{
							clusterv1.ClusterNameLabel: clusterNamespacedName.Name,
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: scw.StringPtr("bootstrap"),
						},
					},
				},
			},
			asserts: func(g *WithT, c client.Client) {
				// ScalewayMachine checks
				sc := &infrav1.ScalewayMachine{}
				g.Expect(c.Get(context.TODO(), scalewayMachineNamespacedName, sc)).To(Succeed())
				g.Expect(sc.Status.Ready).To(BeTrue())
				g.Expect(sc.Finalizers).To(ContainElement(infrav1.MachineFinalizer))
			},
		},
		{
			name: "should reconcile deletion",
			fields: fields{
				createScalewayMachineService: func(machineScope *scope.Machine) *scalewayMachineService {
					return &scalewayMachineService{
						scope:     machineScope,
						Reconcile: func(ctx context.Context) error { return nil },
						Delete:    func(ctx context.Context) error { return nil },
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				req: reconcile.Request{
					NamespacedName: scalewayMachineNamespacedName,
				},
			},
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
					Spec: clusterv1.ClusterSpec{
						InfrastructureRef: &corev1.ObjectReference{
							Name: scalewayClusterNamespacedName.Name,
						},
					},
					Status: clusterv1.ClusterStatus{
						InfrastructureReady: true,
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
				&infrav1.ScalewayMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scalewayMachineNamespacedName.Name,
						Namespace: scalewayMachineNamespacedName.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:       machineNamespacedName.Name,
								Kind:       "Machine",
								APIVersion: clusterv1.GroupVersion.String(),
							},
						},
						Finalizers:        []string{infrav1.MachineFinalizer},
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
				&clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      machineNamespacedName.Name,
						Namespace: machineNamespacedName.Namespace,
						Labels: map[string]string{
							clusterv1.ClusterNameLabel: clusterNamespacedName.Name,
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: scw.StringPtr("bootstrap"),
						},
					},
				},
			},
			asserts: func(g *WithT, c client.Client) {
				// ScalewayMachine should not exist anymore if the finalizer was correctly removed.
				sc := &infrav1.ScalewayMachine{}
				g.Expect(c.Get(context.TODO(), scalewayMachineNamespacedName, sc)).NotTo(Succeed())
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

			r := &ScalewayMachineReconciler{
				Client:                       c,
				createScalewayMachineService: tt.fields.createScalewayMachineService,
			}
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScalewayMachineReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScalewayMachineReconciler.Reconcile() = %v, want %v", got, tt.want)
			}

			tt.asserts(g, c)
		})
	}
}
