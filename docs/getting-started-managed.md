# Getting started (Kapsule / Kosmos)

This document will help you provision a management cluster and a Scaleway managed workload cluster.

## Setup a management cluster

### Provision the cluster

You can use any existing Kubernetes cluster as a management cluster. If you don't
have one, you can use one of the following methods to provision a cluster. At the
end of this section, you must have the kubeconfig of your future management cluster.

#### Method 1: Create a Scaleway Kapsule cluster

Follow this documentation to create a Scaleway Kapsule cluster: [Kubernetes - Quickstart](https://www.scaleway.com/en/docs/kubernetes/quickstart/)

Make sure the `KUBECONFIG` environment variable points to the cluster's kubeconfig:

```console
export KUBECONFIG=/path/to/your/kubeconfig
```

#### Method 2: Create a cluster in Docker with kind

1. Follow this documentation to install Docker: [Install Docker Engine](https://docs.docker.com/engine/install/)
2. Follow this documentation to install kind: [Quick Start](https://kind.sigs.k8s.io/docs/user/quick-start/)
3. Create a kind cluster:

   ```console
   $ kind create cluster
   Creating cluster "kind" ...
   ✓ Ensuring node image (kindest/node:v1.31.2) 🖼
   ✓ Preparing nodes 📦
   ✓ Writing configuration 📜
   ✓ Starting control-plane 🕹️
   ✓ Installing CNI 🔌
   ✓ Installing StorageClass 💾
   Set kubectl context to "kind-kind"
   You can now use your cluster with:

   kubectl cluster-info --context kind-kind

   Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂
   ```

4. Get the kubeconfig:

   ```console
   kind get kubeconfig > mgmt.yaml
   export KUBECONFIG=mgmt.yaml
   ```

### Install cluster API and the Scaleway provider

1. Follow these instructions to install the `clusterctl` command-line tool: [Install clusterctl](https://cluster-api.sigs.k8s.io/user/quick-start#install-clusterctl)

2. Initialize the management cluster:

   ```console
   $ clusterctl init --infrastructure scaleway
   Fetching providers
   Installing cert-manager version="v1.17.2"
   Waiting for cert-manager to be available...
   Installing provider="cluster-api" version="v1.10.2" targetNamespace="capi-system"
   Installing provider="bootstrap-kubeadm" version="v1.10.2" targetNamespace="capi-kubeadm-bootstrap-system"
   Installing provider="control-plane-kubeadm" version="v1.10.2" targetNamespace="capi-kubeadm-control-plane-system"
   Installing provider="infrastructure-scaleway" version="v0.1.0" targetNamespace="caps-system"

   Your management cluster has been initialized successfully!

   You can now create your first workload cluster by running the following:

   clusterctl generate cluster [name] --kubernetes-version [version] | kubectl apply -f -
   ```

## Create a Scaleway managed workload cluster

1. Replace the placeholder values and set the following environment variables:

   ```bash
   export CLUSTER_NAME="my-cluster"

   # Scaleway credentials, project ID and region.
   export SCW_ACCESS_KEY="<ACCESS_KEY>"
   export SCW_SECRET_KEY="<SECRET_KEY>"
   export SCW_PROJECT_ID="<PROJECT_ID>"
   export SCW_REGION="fr-par"
   ```

2. Generate the cluster manifests (update the flags if needed):

   ```bash
   clusterctl generate cluster ${CLUSTER_NAME} \
      --kubernetes-version v1.32.4 \
      --flavor managed \
      --worker-machine-count 1 > my-cluster.yaml
   ```

3. Review and edit the `my-cluster.yaml` file as needed.
   For configuring the CAPS CRDs, refer to the [ScalewayManagedCluster](scalewaymanagedcluster.md),
   [ScalewayManagedControlPlane](scalewaymanagedcontrolplane.md) and
   [ScalewayManagedMachinePool](scalewaymanagedmachinepool.md) documentations.
4. Apply the `my-cluster.yaml` file to create the workload cluster.
5. Wait for the cluster and machines to be ready.

   ```bash
   $ clusterctl describe cluster ${CLUSTER_NAME}
   NAME                                                                          READY  SEVERITY  REASON  SINCE  MESSAGE
   Cluster/my-cluster                                                            True                     2m59s
   ├─ClusterInfrastructure - ScalewayManagedCluster/my-cluster
   ├─ControlPlane - ScalewayManagedControlPlane/my-cluster-control-plane
   └─Workers
   └─MachinePool/my-cluster-mp-0                                               True                     2m
       └─MachinePoolInfrastructure - ScalewayManagedMachinePool/my-cluster-mp-0
   ```

6. Fetch the kubeconfig of the cluster.

   ```bash
   clusterctl get kubeconfig ${CLUSTER_NAME} > kubeconfig.yaml
   export KUBECONFIG=kubeconfig.yaml
   ```

7. List nodes.

   ```bash
   $ kubectl get nodes
   NAME                                             STATUS   ROLES    AGE     VERSION
   scw-default-my-cluster-control-my-clust-c8e009   Ready    <none>   4m13s   v1.32.4
   ```
