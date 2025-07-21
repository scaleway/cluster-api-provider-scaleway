package e2e

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Workload cluster creation", func() {
	Context("Running the CAPSClusterDeploymentSpec with the default flavor", func() {
		CAPSClusterDeploymentSpec(func() CAPSClusterDeploymentSpecInput {
			return CAPSClusterDeploymentSpecInput{
				E2EConfig:                e2eConfig,
				ClusterctlConfigPath:     clusterctlConfigPath,
				BootstrapClusterProxy:    bootstrapClusterProxy,
				ArtifactFolder:           artifactFolder,
				SkipCleanup:              skipCleanup,
				ControlPlaneMachineCount: 3,
				WorkerMachineCount:       2,
				Flavor:                   "",
			}
		})
	})

	Context("Running the CAPSClusterDeploymentSpec with the private-network flavor", func() {
		CAPSClusterDeploymentSpec(func() CAPSClusterDeploymentSpecInput {
			return CAPSClusterDeploymentSpecInput{
				E2EConfig:                e2eConfig,
				ClusterctlConfigPath:     clusterctlConfigPath,
				BootstrapClusterProxy:    bootstrapClusterProxy,
				ArtifactFolder:           artifactFolder,
				SkipCleanup:              skipCleanup,
				ControlPlaneMachineCount: 3,
				WorkerMachineCount:       2,
				Flavor:                   "private-network",
			}
		})
	})

	Context("Running the CAPSClusterDeploymentSpec with the private-network flavor and no public IPs", func() {
		CAPSClusterDeploymentSpec(func() CAPSClusterDeploymentSpecInput {
			return CAPSClusterDeploymentSpecInput{
				E2EConfig:                e2eConfig,
				ClusterctlConfigPath:     clusterctlConfigPath,
				BootstrapClusterProxy:    bootstrapClusterProxy,
				ArtifactFolder:           artifactFolder,
				SkipCleanup:              skipCleanup,
				ControlPlaneMachineCount: 3,
				WorkerMachineCount:       2,
				Flavor:                   "private-network",
				ClusterctlVariables: map[string]string{
					"CONTROL_PLANE_MACHINE_IPV4": "false",
					"WORKER_MACHINE_IPV4":        "false",
					"PUBLIC_GATEWAYS":            "[{}]",
				},
			}
		})
	})
})
