# ScalewayManagedMachinePool

The `ScalewayManagedMachinePool` resource provisions a pool in a Scaleway Managed Kubernetes cluster.

This document describes the various configuration options you can set to configure a `ScalewayManagedMachinePool`.

## Minimal ScalewayManagedMachinePool

The `ScalewayManagedMachinePool` with the minimum options looks like this:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  nodeType: GP1-XS
  zone: fr-srr-1
```

## Additional tags

You can configure additional tags that will be set on the created pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  additionalTags:
    - "test"
    - "test1"
```

> [!WARNING]
> Do not attempt to update the tags directly via the Scaleway API as tags will always
> be overwritten by the provider during ScalewayManagedMachinePool reconciliation.

## Autohealing

You can enable autohealing in the pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  autohealing: true
```

## Security Group and Placement Group

You can set specify a Security Group and Placement Group ID during the creation of the pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  placementGroupID: 11111111-1111-1111-1111-111111111111
  securityGroupID: 11111111-1111-1111-1111-111111111111
```

## Autoscaling configuration

You can enable autoscaling on the pool and set the min/max size of the pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  scaling:
    autoscaling: true
    minSize: 0
    maxSize: 5
```

## Upgrade policy

You can set the upgrade policy of the pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  upgradePolicy:
    maxUnavailable: 0
    maxSurge: 2
```

## Kubelet args

You can set Kubelet args on the pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  kubeletArgs:
    containerLogMaxFiles: "10"
    registryPullQPS: "10"
```

You can use the Scaleway CLI to list the available kubelet args: `$ scw k8s version list -o json | jq`.

## Root volume configuration

You can configure the root volume of the nodes of the pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  rootVolumeSizeGB: 40
  rootVolumeType: "sbs_15k"
```

## Full isolation pool

You can disable adding a public IP on the nodes of the pool:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedMachinePool
metadata:
  name: my-cluster-managed-machine-pool
  namespace: default
spec:
  # some fields were omitted...
  publicIPDisabled: true
```

Setting `publicIPDisabled: true` is only possible with a Kapsule cluster.
The Private Network of the cluster must also have at least one public gateway that
advertises a default route.
