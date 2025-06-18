# ScalewayMachine

The `ScalewayMachine` resource provisions the necessary Scaleway infrastructure
to make a Kubernetes node of your workload cluster work. This includes
an [Instance server](https://www.scaleway.com/en/virtual-instances/).

This document describes the various configuration options you can set to enable or disable
important features on a `ScalewayMachine`.

The infrastucture resources for the `ScalewayMachine` will be created in a
Scaleway availability zone that is based on the associated `Machine`'s `failureDomain`.

You will usually never create a `ScalewayMachine` directly, `ScalewayMachineTemplate` should be used instead:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayMachineTemplate
metadata:
  name: my-machine-template
  namespace: default
spec:
  template:
    spec: # Put your ScalewayMachine spec here:
      image:
        name: cluster-api-rockylinux-9-v1.32.4
      commercialType: DEV1-S
      rootVolume:
        type: block
```

## Commercial Type

The `commercialType` field can be any Scaleway Instance commercial type, as long as
it's available in the selected availability zone. This field is required and
cannot be updated later. For a list of available commercial types, you may refer
to the following pages:

- [Scaleway Instances datasheet](https://www.scaleway.com/en/docs/instances/reference-content/instances-datasheet/)
- [Choosing the right GPU Instance type](https://www.scaleway.com/en/docs/gpu/reference-content/choosing-gpu-instance-type/)

> [!WARNING]
>
> - The price of Scaleway Instances is based on commercial type AND availability zone.
> - Some commercial types may not be available in all availability zones.

## Image

An Instance image will be used to provision the Instance servers. This image must
exist and be compatible with the chosen Instance commercial type (e.g. same CPU architecture).
It must also be available in the availability zone of the server.

The `image` field must contain one of the following:

- An image ID:

  ```yaml
  apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
  kind: ScalewayMachine
  metadata:
    name: my-machine
    namespace: default
  spec:
    image:
      id: 11111111-1111-1111-1111-111111111111
    # some fields were omitted...
  ```

- An image name:

  ```yaml
  apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
  kind: ScalewayMachine
  metadata:
    name: my-machine
    namespace: default
  spec:
    image:
      name: cluster-api-rockylinux-9-v1.32.4
    # some fields were omitted...
  ```

  Make sure this image exists in the zones where you plan to deploy your nodes.
  You can list images by name with this command:

  ```bash
  scw instance image list name=${IMAGE_NAME} zone=${SCW_ZONE}
  ```

- An image [Marketplace label](https://www.scaleway.com/en/developers/api/marketplace/):

  ```yaml
  apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
  kind: ScalewayMachine
  metadata:
    name: my-machine
    namespace: default
  spec:
    image:
      label: ubuntu_jammy
    # some fields were omitted...
  ```

To build your own image, you can use the [Kubernetes image-builder project](https://image-builder.sigs.k8s.io/capi/quickstart).

## Root Volume

During machine creation, a root volume will be provisioned based on the chosen image.

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayMachine
metadata:
  name: my-machine
  namespace: default
spec:
  rootVolume:
    size: 20
    type: block # can be block or local
    iops: 15000 # can be 5000 or 15000
  # some fields were omitted...
```

The default size of the root volume is 20 GB, this corresponds to the `rootVolume.size`
field set to `20`. You should adjust this value depending on the expected disk space usage.

The `type` field defaults to `block`, which uses the Scaleway Block Storage (SBS) product
to provision a root volume. This field can be set to `local` to use Instance SSD
Local Storage (`l_ssd`), note however that not all commercial types support local storage.

> [!TIP]
> The type of the root volume MUST match the type of Instance image you specified in
> the `image` field of the `ScalewayMachine` spec. For example, if your image is
> based on an SBS Snapshot, you can't use a `local` root volume.

The `iops` field is only applicable for `block` root volumes. By default, `block`
volumes will have 5000 IOPS. Currently, only `5000` and `15000` IOPS are supported.

For GPU Instances that support [scratch storage](https://www.scaleway.com/en/docs/gpu/how-to/use-scratch-storage-h100-instances/),
an additional scratch volume is automatically created and attached to the Instance.

## Public Network

The `publicNetwork` field defines if an IPv4 and/or IPv6 should be created and attached
during the Instance server creation.

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayMachine
metadata:
  name: my-machine
  namespace: default
spec:
  publicNetwork:
    enableIPv4: true
    enableIPv6: true
  # some fields were omitted...
```

By default, no IPv6 is created unless `publicNetwork.enableIPv6` is set to `true`.

The default behavior for IPv4 creation depends on the configuration of a
Private Network (`network.privateNetwork.enabled`) in the `ScalewayCluster`.

- If `network.privateNetwork.enabled` is `false` or not set, a public IPv4 will **always**
  be created, even if `publicNetwork.enableIPv4` is set to `false`. This is to prevent
  the creation of a node that can't access the control-plane Load Balancer.
- If `network.privateNetwork.enabled` is `true`, no public IPv4 is created, unless
  `publicNetwork.enableIPv4` is also set to `true`. If you do not enable public IPv4
  connectivity, make sure a Public Gateway advertises its default route in the
  Private Network as nodes will not be able to access the control-plane Load Balancer
  without public connectivity.

## Placement Group

It is possible to attach an existing Placement group to the Instance server that will be created.

> [!WARNING]
> A Placement group can be attached to at most 20 Instance servers.

Placement groups allow you to define if you want certain Instances to run on
different physical hypervisors for maximum availability or as physically close
together as possible for minimum latency. For more information about Placement
groups, please refer to the [Scaleway documentation](https://www.scaleway.com/en/docs/instances/how-to/use-placement-groups/).

The `placementGroup` field must contain one of the following:

- A Placement group ID:

  ```yaml
  apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
  kind: ScalewayMachine
  metadata:
    name: my-machine
    namespace: default
  spec:
    placementGroup:
      id: 11111111-1111-1111-1111-111111111111
    # some fields were omitted...
  ```

- A Placement group name:

  ```yaml
  apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
  kind: ScalewayMachine
  metadata:
    name: my-machine
    namespace: default
  spec:
    placementGroup:
      name: my-placement-group
    # some fields were omitted...
  ```

  Make sure this Placement group exists in the zones where you plan to deploy your nodes.
  You can list Placement groups by name with this command:

  ```bash
  scw instance placement-group list name=${IMAGE_NAME} zone=${SCW_ZONE}
  ```
