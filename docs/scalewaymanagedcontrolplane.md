# ScalewayManagedControlPlane

The `ScalewayManagedControlPlane` resource provisions a Scaleway Managed Kubernetes cluster
using [Kapsule](https://www.scaleway.com/en/kubernetes-kapsule/) or [Kosmos](https://www.scaleway.com/fr/kubernetes-kosmos/).

This document describes the various configuration options you can set to configure a `ScalewayManagedControlPlane`.

## Minimal ScalewayManagedControlPlane

The `ScalewayManagedControlPlane` with the minimum options looks like this:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  type: kapsule
  version: v1.32.0
```

The `type` field must be set to the desired cluster type (e.g. `kapsule`, `kapsule-dedicated-4`, `multicloud`, etc.).
You can list the available cluster types using the Scaleway CLI: `$ scw k8s cluster-type list`.
The cluster is automatically updated to the desired type when the `type` field is updated.

The `version` field must be set to one of the supported Kubernetes version.
You can list the supported Kubernetes versions using the Scaleway CLI: `$ scw k8s version list`.
The cluster is automatically upgraded when the `version` field is bumped to a
version that is above the current version of the cluster. It is not possible to
downgrade the version of a cluster.

## Additional tags

You can configure additional tags that will be set on the created Scaleway Managed Kubernetes cluster:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  additionalTags:
    - "test"
    - "test1"
```

> [!WARNING]
> Do not attempt to update the tags directly via the Scaleway API as tags will always
> be overwritten by the provider during ScalewayManagedControlPlane reconciliation.

## ACL

You can configure the IPs allowed to access the public endpoint of the cluster:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  acl:
    allowedRanges:
      - "10.10.10.0/24"
      - "20.20.0.0/16"
```

If the `acl` field is not set, the provider will ensure that the ACL rule with
IP range `0.0.0.0/0` is set on the cluster.

> [!WARNING]
> Make sure the nodes of the management cluster are allowed to access the cluster.

## Autoscaler configuration

You can configure the autoscaler of the cluster:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  autoscaler:
    scaleDownDisabled: false
    scaleDownDelayAfterAdd: 10m
    expander: most_pods
    ignoreDaemonsetsUtilization: false
    balanceSimilarNodeGroups: false
    scaleDownUtilizationThreshold: "0.5"
    maxGracefulTerminationSec: 600
```

## Auto Upgrade configuration

You can set the auto upgrade configuration of the cluster:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  autoUpgrade:
    enabled: true
    maintenanceWindow:
      startHour: 0
      day: any
```

## Feature Gates

You can enable Kubernetes feature gates on the cluster:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  featureGates:
    - "HPAScaleToZero"
    - "PodLevelResources"
```

You can use the Scaleway CLI to list the available feature gates: `$ scw k8s version list -o json | jq`.

## Admission Plugins

You can enable Kubernetes admission plugins on the cluster:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  admissionPlugins:
    - "AlwaysPullImages"
    - "PodNodeSelector"
```

You can use the Scaleway CLI to list the available admission plugins: `$ scw k8s version list -o json | jq`.

## API Server Cert SANs

You can add additional API Server Cert SANs:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  apiServerCertSANs:
    - "mycluster.com"
```

## Open ID Connect configuration

You can set the OIDC configuration of the cluster:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  openIDConnect:
    issuerURL: "https://oidc-provider.example.com"
    clientID: "test"
    usernameClaim: "email"
    usernamePrefix: "myusernameprefix"
    groupsClaim:
      - "groups"
    groupsPrefix: "mygroupprefix"
    requiredClaim:
      - "yourkey=yourvalue"
```

## Cluster deletion behavior

You can enable the deletion of additional resources (e.g. Load Balancers, Persistent Volumes, etc.)
when the cluster is deleted:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: ScalewayManagedControlPlane
metadata:
  name: my-cluster-control-plane
  namespace: default
spec:
  # some fields were omitted...
  onDelete:
    withAdditionalResources: true
```
