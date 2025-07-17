package e2e

import (
	"context"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/util"
)

func Byf(format string, a ...any) {
	By(fmt.Sprintf(format, a...))
}

func setupSpecNamespace(ctx context.Context, specName string, clusterProxy framework.ClusterProxy, artifactFolder string) (*corev1.Namespace, context.CancelFunc) {
	Byf("Creating a namespace for hosting the %q test spec", specName)
	namespace, cancelWatches := framework.CreateNamespaceAndWatchEvents(ctx, framework.CreateNamespaceAndWatchEventsInput{
		Creator:   clusterProxy.GetClient(),
		ClientSet: clusterProxy.GetClientSet(),
		Name:      fmt.Sprintf("%s-%s", specName, util.RandomString(6)),
		LogFolder: filepath.Join(artifactFolder, "clusters", clusterProxy.GetName()),
	})

	return namespace, cancelWatches
}

func dumpSpecResourcesAndCleanup(ctx context.Context, specName string, clusterProxy framework.ClusterProxy, artifactFolder, clusterctlConfigPath string, namespace *corev1.Namespace, cancelWatches context.CancelFunc, cluster *clusterv1.Cluster, intervalsGetter func(spec, key string) []any, skipCleanup bool) {
	var clusterName string
	var clusterNamespace string
	if cluster != nil {
		clusterName = cluster.Name
		clusterNamespace = cluster.Namespace
		Byf("Dumping logs from the %q workload cluster", clusterName)

		// Dump all the logs from the workload cluster before deleting them.
		clusterProxy.CollectWorkloadClusterLogs(ctx, clusterNamespace, clusterName, filepath.Join(artifactFolder, "clusters", clusterName))

		Byf("Dumping all the Cluster API resources in the %q namespace", namespace.Name)

		// Dump all Cluster API related resources to artifacts before deleting them.
		framework.DumpAllResources(ctx, framework.DumpAllResourcesInput{
			Lister:               clusterProxy.GetClient(),
			KubeConfigPath:       clusterProxy.GetKubeconfigPath(),
			ClusterctlConfigPath: clusterctlConfigPath,
			Namespace:            namespace.Name,
			LogPath:              filepath.Join(artifactFolder, "clusters", clusterProxy.GetName(), "resources"),
		})
	} else {
		clusterName = "empty"
		clusterNamespace = "empty"
	}

	if !skipCleanup {
		Byf("Deleting cluster %s/%s", clusterNamespace, clusterName)
		// While https://github.com/kubernetes-sigs/cluster-api/issues/2955 is addressed in future iterations, there is a chance
		// that cluster variable is not set even if the cluster exists, so we are calling DeleteAllClustersAndWait
		// instead of DeleteClusterAndWait
		framework.DeleteAllClustersAndWait(ctx, framework.DeleteAllClustersAndWaitInput{
			ClusterProxy:         clusterProxy,
			ClusterctlConfigPath: clusterctlConfigPath,
			Namespace:            namespace.Name,
			ArtifactFolder:       artifactFolder,
		}, intervalsGetter(specName, "wait-delete-cluster")...)

		Byf("Deleting namespace used for hosting the %q test spec", specName)
		framework.DeleteNamespace(ctx, framework.DeleteNamespaceInput{
			Deleter: clusterProxy.GetClient(),
			Name:    namespace.Name,
		})
	}
	cancelWatches()
}
