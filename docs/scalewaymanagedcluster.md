# ScalewayManagedCluster

The `ScalewayManagedCluster` resource provisions the necessary Scaleway infrastructure
to make the Scaleway managed workload cluster work. This may include [Private Networks](https://www.scaleway.com/en/vpc/),
[Public Gateways](https://www.scaleway.com/en/public-gateway/), and more, depending on the configuration of the `ScalewayManagedCluster`.

This document describes the various configuration options you can set to enable or disable
important features on a `ScalewayManagedCluster`.

## Minimal ScalewayManagedCluster

The `ScalewayManagedCluster` with the minimum options looks like this:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayManagedCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  projectID: 11111111-1111-1111-1111-111111111111
  region: fr-par
  scalewaySecretName: my-scaleway-secret
```

The `projectID`, `region` and `scalewaySecretName` fields are **required**.

The `projectID` and `region` fields are **immutable**, they cannot be updated after creation.

The `scalewaySecretName` field must contain the name of an existing `Secret` inside the
namespace of the `ScalewayManagedCluster`. For more information about this secret, please refer
to the [Scaleway Secret documentation](secret.md).

## VPC

### Private Network

If the `ScalewayManagedCluster` is associated with a Kapsule `ScalewayManagedControlPlane`,
a new Private Network is automatically created if none is provided.

It is possible to re-use an existing Private Network or configure the VPC where the
Private Network will be created.

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayManagedCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  network:
    privateNetwork:
      # id: 11111111-1111-1111-1111-111111111111
      # vpcID: 11111111-1111-1111-1111-111111111111
      # subnet: 192.168.0.0/22
  # some fields were omitted...
```

- The `id` field can be set to use an existing Private Network. If not set, the provider
  will create a new Private Network and manage it.
- The `vpcID` field can be set to tell the provider to create a new Private Network inside a
  specific VPC. If not set, Private Networks are created in the default VPC.
- The `subnet` field can be set to use a specific subnet. Make sure the subnet does not
  overlap with the subnet of another Private Network in the VPC.

### Public Gateways

To create `ScalewayManagedMachinePools` without a Public IP, your Private Network must contain
at least one Public Gateway that advertises its default route. You can configure one
manually or let the provider configure that for you:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayManagedCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  region: fr-par
  network:
    publicGateways:
      - type: VPC-GW-S
        zone: fr-par-1
      - type: VPC-GW-S
        zone: fr-par-2
        # ip: 42.42.42.42
      # Note: the Public Gateway product is currently not available in fr-par-3.
  # some fields were omitted...
```

The `ip` field can be set on the spec of a Public Gateway to use an existing Public IP.
If not set, a new IP will be created.

> [!CAUTION]
> The `publicGateways` field is fully mutable, but changes should be avoided as much as possible.
>
> 🚫📶 Updating existing Public Gateways can lead to a loss of network on the nodes, be
> very careful when updating this field.
>
> 🚮 Updating a Public Gateway will lead to its re-creation, which will make its private IP change.
> The only change that won't lead to a re-creation of the Public Gateway is a type upgrade
> (e.g. VPC-GW-S to VPC-GW-M). Downgrading a Public Gateway is only possible through a re-creation.
>
> ⏳ Because the default routes are advertised via DHCP, the DHCP leases of the nodes must
> be renewed for changes to be propagated (~24 hours). You can reboot the nodes or
> re-create a new `MachinePool` to force the propagation.
