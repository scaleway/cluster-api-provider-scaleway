# Getting started

This document will help you provision a management cluster and a workload cluster.

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

## Prepare the OS image

In order to provision a workload cluster, you will need to create an OS image
with all the necessary dependencies pre-installed (kubeadm, containerd, etc.).
This OS image will be used to provision the nodes of your workload cluster. At the
end of this section, you must have the name of the Instance image that will be used to provision
the machines.

### Method 1: Reuse an existing image

Set the following environment variable:

```bash
export SCW_PROJECT_ID="<PROJECT_ID>"
```

Use the following command to import the `cluster-api-ubuntu-2404-v1.32.4` image provided by Scaleway:

> [!WARNING]
> This image is provided only for testing and should not be used in production.

```bash
export SNAPSHOT_ID=$(scw block snapshot import-from-object-storage \
    name=cluster-api-ubuntu-2404-v1.32.4 \
    bucket=scwcaps \
    key=images/cluster-api-ubuntu-2404-v1.32.4.qcow2 \
    project-id=${SCW_PROJECT_ID} \
    -o json | jq -r .id)
```

Wait for the snapshot to have the `available` status:

```bash
watch scw block snapshot get ${SNAPSHOT_ID}
```

Finally, create an Instance image with the name `cluster-api-ubuntu-2404-v1.32.4` from the previously imported snapshot:

```bash
scw instance image create \
   name=cluster-api-ubuntu-2404-v1.32.4 \
   arch=x86_64 \
   snapshot-id=${SNAPSHOT_ID} \
   project-id=${SCW_PROJECT_ID}
```

### Method 2: Build an OS image with `image-builder`

The [image-builder](https://github.com/kubernetes-sigs/image-builder) project allows
you to build a CAPI-ready OS image using Packer and Ansible.

To begin, please clone the `image-builder` repository:

```bash
git clone https://github.com/kubernetes-sigs/image-builder.git
```

Then, please follow the [Building Images for Scaleway documentation](https://image-builder.sigs.k8s.io/capi/providers/scaleway).

## Create a basic workload cluster

1. Replace the placeholder values and set the following environment variables:

   ```bash
   export CLUSTER_NAME="my-cluster"

   # Scaleway credentials, project ID and region.
   export SCW_ACCESS_KEY="<ACCESS_KEY>"
   export SCW_SECRET_KEY="<SECRET_KEY>"
   export SCW_PROJECT_ID="<PROJECT_ID>"
   export SCW_REGION="fr-par"

   # Scaleway Instance image names that will be used to provision servers.
   export CONTROL_PLANE_MACHINE_IMAGE="<IMAGE_NAME>"
   export WORKER_MACHINE_IMAGE="<IMAGE_NAME>"
   ```

2. Generate the cluster manifests (update the flags if needed):

   ```bash
   clusterctl generate cluster ${CLUSTER_NAME} \
      --kubernetes-version v1.32.4 \
      --control-plane-machine-count 1 \
      --worker-machine-count 1 > my-cluster.yaml
   ```

3. Review and edit the `my-cluster.yaml` file as needed.
   For configuring the CAPS CRDs, refer to the [ScalewayCluster](scalewaycluster.md)
   and [ScalewayMachine](scalewaymachine.md) documentations.
4. Apply the `my-cluster.yaml` file to create the workload cluster.
5. Wait for the cluster and machines to be ready.

   ```bash
   $ clusterctl describe cluster ${CLUSTER_NAME}
   NAME                                                                          READY  SEVERITY  REASON                       SINCE  MESSAGE
   Cluster/my-cluster                                                            True                                          2m19s
   ├─ClusterInfrastructure - ScalewayCluster/my-cluster
   ├─ControlPlane - KubeadmControlPlane/my-cluster-control-plane                 True                                          2m19s
   │ └─Machine/my-cluster-control-plane-pxpdl                                    True                                          3m19s
   │   └─MachineInfrastructure - ScalewayMachine/my-cluster-control-plane-pxpdl
   └─Workers
   └─MachineDeployment/my-cluster-md-0                                         False  Warning   WaitingForAvailableMachines  3m31s  Minimum availability requires 1 replicas, current 0 available
      └─Machine/my-cluster-md-0-bgzv8-5k96v                                     True                                          2m15s
         └─MachineInfrastructure - ScalewayMachine/my-cluster-md-0-bgzv8-5k96v
   ```

6. Fetch the kubeconfig of the cluster.

   ```bash
   clusterctl get kubeconfig ${CLUSTER_NAME} > kubeconfig.yaml
   export KUBECONFIG=kubeconfig.yaml
   ```

7. List nodes.

   ```bash
   $ kubectl get nodes
   NAME                             STATUS     ROLES           AGE     VERSION
   my-cluster-control-plane-pxpdl   NotReady   control-plane   3m46s   v1.32.4
   my-cluster-md-0-bgzv8-5k96v      NotReady   <none>          2m57s   v1.32.4
   ```

> [!NOTE]
> Nodes will have the `NotReady` status until a CNI is installed in the cluster.

### Setup the workload cluster

The workload cluster is ready to use. You should now:

- Install a CNI plugin
- (Optional) Install the [Scaleway CCM](https://github.com/scaleway/scaleway-cloud-controller-manager) to manage Nodes and LoadBalancers
- (Optional) Install the [Scaleway CSI](https://github.com/scaleway/scaleway-csi) driver to manage block volumes and snapshots
