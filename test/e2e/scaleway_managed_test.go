package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	capi_e2e "sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	"sigs.k8s.io/cluster-api/util"
)

var _ = Describe("Managed workload cluster creation", func() {
	var (
		ctx                 = context.TODO()
		specName            = "managed" // TODO: set to "create-managed-workload-cluster" when tag issue is fixed.
		namespace           *corev1.Namespace
		cancelWatches       context.CancelFunc
		result              *ApplyManagedClusterTemplateAndWaitResult
		clusterName         string
		clusterctlLogFolder string
	)

	BeforeEach(func() {
		Expect(e2eConfig).ToNot(BeNil(), "Invalid argument. e2eConfig can't be nil when calling %s spec", specName)
		Expect(clusterctlConfigPath).To(BeAnExistingFile(), "Invalid argument. clusterctlConfigPath must be an existing file when calling %s spec", specName)
		Expect(bootstrapClusterProxy).ToNot(BeNil(), "Invalid argument. bootstrapClusterProxy can't be nil when calling %s spec", specName)
		Expect(os.MkdirAll(artifactFolder, 0o755)).To(Succeed(), "Invalid argument. artifactFolder can't be created for %s spec", specName)
		Expect(e2eConfig.Variables).To(HaveKey(capi_e2e.KubernetesVersion))

		clusterName = fmt.Sprintf("caps-e2e-%s", util.RandomString(6))

		// Setup a Namespace where to host objects for this spec and create a watcher for the namespace events.
		namespace, cancelWatches = setupSpecNamespace(ctx, specName, bootstrapClusterProxy, artifactFolder)

		result = new(ApplyManagedClusterTemplateAndWaitResult)

		// We need to override clusterctl apply log folder to avoid getting our credentials exposed.
		clusterctlLogFolder = filepath.Join(os.TempDir(), "clusters", bootstrapClusterProxy.GetName())
	})

	AfterEach(func() {
		cleanInput := cleanupInput{
			SpecName:             specName,
			Cluster:              result.Cluster,
			ClusterProxy:         bootstrapClusterProxy,
			ClusterctlConfigPath: clusterctlConfigPath,
			Namespace:            namespace,
			CancelWatches:        cancelWatches,
			IntervalsGetter:      e2eConfig.GetIntervals,
			SkipCleanup:          skipCleanup,
			ArtifactFolder:       artifactFolder,
		}

		dumpSpecResourcesAndCleanup(ctx, cleanInput)
	})

	Context("Creating a Scaleway Kapsule cluster", func() {
		It("Should create a cluster with 1 machine pool and scale", func() {
			By("Initializes with 1 machine pool")

			ApplyManagedClusterTemplateAndWait(ctx, ApplyManagedClusterTemplateAndWaitInput{
				ClusterProxy: bootstrapClusterProxy,
				ConfigCluster: clusterctl.ConfigClusterInput{
					LogFolder:                clusterctlLogFolder,
					ClusterctlConfigPath:     clusterctlConfigPath,
					KubeconfigPath:           bootstrapClusterProxy.GetKubeconfigPath(),
					InfrastructureProvider:   clusterctl.DefaultInfrastructureProvider,
					Flavor:                   "managed",
					Namespace:                namespace.Name,
					ClusterName:              clusterName,
					KubernetesVersion:        e2eConfig.MustGetVariable(capi_e2e.KubernetesVersion),
					ControlPlaneMachineCount: ptr.To[int64](1),
					WorkerMachineCount:       ptr.To[int64](3),
					ClusterctlVariables: map[string]string{
						"WORKER_PUBLIC_IP_DISABLED": "true",
						"PUBLIC_GATEWAYS":           "[{}]",
					},
				},
				WaitForClusterIntervals:      e2eConfig.GetIntervals(specName, "wait-cluster"),
				WaitForControlPlaneIntervals: e2eConfig.GetIntervals(specName, "wait-control-plane"),
				WaitForMachinePools:          e2eConfig.GetIntervals(specName, "wait-worker-machine-pools"),
			}, result)

			By("Scaling the machine pool up")
			framework.ScaleMachinePoolAndWait(ctx, framework.ScaleMachinePoolAndWaitInput{
				ClusterProxy:              bootstrapClusterProxy,
				Cluster:                   result.Cluster,
				Replicas:                  4,
				MachinePools:              result.MachinePools,
				WaitForMachinePoolToScale: e2eConfig.GetIntervals(specName, "wait-worker-machine-pools"),
			})

			By("Scaling the machine pool down")
			framework.ScaleMachinePoolAndWait(ctx, framework.ScaleMachinePoolAndWaitInput{
				ClusterProxy:              bootstrapClusterProxy,
				Cluster:                   result.Cluster,
				Replicas:                  3,
				MachinePools:              result.MachinePools,
				WaitForMachinePoolToScale: e2eConfig.GetIntervals(specName, "wait-worker-machine-pools"),
			})
		})
	})
})
