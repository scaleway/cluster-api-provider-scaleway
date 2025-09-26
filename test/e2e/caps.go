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

// CAPSClusterDeploymentSpecInput is the input for CAPSClusterDeploymentSpec.
type CAPSClusterDeploymentSpecInput struct {
	E2EConfig                *clusterctl.E2EConfig
	ClusterctlConfigPath     string
	BootstrapClusterProxy    framework.ClusterProxy
	ArtifactFolder           string
	SkipCleanup              bool
	Flavor                   string
	WorkerMachineCount       int64
	ControlPlaneMachineCount int64
	ClusterctlVariables      map[string]string
}

// CAPSClusterDeploymentSpec implements a test that verifies that MachineDeployment rolling updates are successful.
func CAPSClusterDeploymentSpec(inputGetter func() CAPSClusterDeploymentSpecInput) {
	var (
		ctx              = context.TODO()
		specName         = "create-workload-cluster"
		input            CAPSClusterDeploymentSpecInput
		namespace        *corev1.Namespace
		cancelWatches    context.CancelFunc
		clusterResources *clusterctl.ApplyClusterTemplateAndWaitResult
		clusterName      string
	)

	BeforeEach(func() {
		input = inputGetter()
		Expect(input.E2EConfig).ToNot(BeNil(), "Invalid argument. input.E2EConfig can't be nil when calling %s spec", specName)
		Expect(input.ClusterctlConfigPath).To(BeAnExistingFile(), "Invalid argument. input.ClusterctlConfigPath must be an existing file when calling %s spec", specName)
		Expect(input.BootstrapClusterProxy).ToNot(BeNil(), "Invalid argument. input.BootstrapClusterProxy can't be nil when calling %s spec", specName)
		Expect(os.MkdirAll(input.ArtifactFolder, 0o750)).To(Succeed(), "Invalid argument. input.ArtifactFolder can't be created for %s spec", specName)
		Expect(input.E2EConfig.Variables).To(HaveKey(capi_e2e.KubernetesVersion))

		// Setup a Namespace where to host objects for this spec and create a watcher for the namespace events.
		namespace, cancelWatches = setupSpecNamespace(ctx, specName, input.BootstrapClusterProxy, input.ArtifactFolder)
		clusterResources = new(clusterctl.ApplyClusterTemplateAndWaitResult)

		clusterName = fmt.Sprintf("%s-%s", specName, util.RandomString(6))
	})

	It("Should successfully create a cluster", func() {
		By("Creating a workload cluster")
		clusterctl.ApplyClusterTemplateAndWait(ctx, clusterctl.ApplyClusterTemplateAndWaitInput{
			ClusterProxy: input.BootstrapClusterProxy,
			ConfigCluster: clusterctl.ConfigClusterInput{
				LogFolder:                filepath.Join(input.ArtifactFolder, "clusters", input.BootstrapClusterProxy.GetName()),
				ClusterctlConfigPath:     input.ClusterctlConfigPath,
				KubeconfigPath:           input.BootstrapClusterProxy.GetKubeconfigPath(),
				InfrastructureProvider:   clusterctl.DefaultInfrastructureProvider,
				Flavor:                   input.Flavor,
				Namespace:                namespace.Name,
				ClusterName:              clusterName,
				KubernetesVersion:        input.E2EConfig.MustGetVariable(capi_e2e.KubernetesVersion),
				ControlPlaneMachineCount: ptr.To(input.ControlPlaneMachineCount),
				WorkerMachineCount:       ptr.To(input.WorkerMachineCount),
				ClusterctlVariables:      input.ClusterctlVariables,
			},
			WaitForClusterIntervals:      input.E2EConfig.GetIntervals(specName, "wait-cluster"),
			WaitForControlPlaneIntervals: input.E2EConfig.GetIntervals(specName, "wait-control-plane"),
			WaitForMachineDeployments:    input.E2EConfig.GetIntervals(specName, "wait-worker-nodes"),
		}, clusterResources)

		By("PASSED!")
	})

	AfterEach(func() {
		// Dumps all the resources in the spec namespace, then cleanups the cluster object and the spec namespace itself.
		cleanInput := cleanupInput{
			SpecName:             specName,
			Cluster:              clusterResources.Cluster,
			ClusterProxy:         input.BootstrapClusterProxy,
			ClusterctlConfigPath: input.ClusterctlConfigPath,
			Namespace:            namespace,
			CancelWatches:        cancelWatches,
			IntervalsGetter:      input.E2EConfig.GetIntervals,
			SkipCleanup:          input.SkipCleanup,
			ArtifactFolder:       input.ArtifactFolder,
		}
		dumpSpecResourcesAndCleanup(ctx, cleanInput)
	})
}
