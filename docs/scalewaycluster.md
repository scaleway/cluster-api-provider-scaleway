# ScalewayCluster

The `ScalewayCluster` resource provisions the necessary Scaleway infrastructure
to make the Kubernetes workload cluster work. This may include [Load Balancers](https://www.scaleway.com/en/load-balancer/),
[Private Networks](https://www.scaleway.com/en/vpc/), [Public Gateways](https://www.scaleway.com/en/public-gateway/),
and more, depending on the configuration of the `ScalewayCluster`.

This document describes the various configuration options you can set to enable or disable
important features on a `ScalewayCluster`.

## Minimal ScalewayCluster

The `ScalewayCluster` with the minimum options looks like this:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
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
namespace of the `ScalewayCluster`. For more information about this secret, please refer
to the [Scaleway Secret documentation](secret.md).

## Failure domains

The `failureDomains` field allows to set the Scaleway availability zones where the
control-plane nodes will be deployed. The specified availability zones must be in
the same region as the region specified in the `region` field. When omitted, all
availability zones in the specified region are automatically eligible for hosting
control-plane nodes.

> [!WARNING]
> Pricing of Scaleway products may differ between availability zones. For better
> cost predictability, it is recommended to always set your desired `failureDomains`.

You can find the current failure domains of a `ScalewayCluster` by running this command:

```bash
$ kubectl get scalewaycluster my-cluster --template='{{.status.failureDomains}}'
map[fr-par-1:map[controlPlane:true] fr-par-2:map[controlPlane:true] fr-par-3:map[controlPlane:true]]
```

Here is an example of `ScalewayCluster` with the failure domains set to `fr-par-1` and `fr-par-2`:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  region: fr-par
  failureDomains:
    - fr-par-1
    - fr-par-2
  # some fields were omitted...
```

## Other features

### DNS

Set the `network.controlPlaneDNS` field to automatically configure DNS records that
will point to the Load Balancer address(es) of your workload cluster.

In this example, the FQDN `my-cluster.subdomain.your-domain.com` will have `A` record(s)
configured to point to the Load Balancer IP address(es).

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  network:
    controlPlaneDNS:
      domain: subdomain.your-domain.com
      name: my-cluster
  # some fields were omitted...
```

The `domain` must be an existing Scaleway DNS Zone. Please refer to the
[Scaleway Domains and DNS documentation](https://www.scaleway.com/en/docs/domains-and-dns/)
for more information. You may register an external domain by following
[this documentation](https://www.scaleway.com/en/docs/domains-and-dns/how-to/add-external-domain/).

The `network.controlPlaneDNS` field is **immutable**, it cannot be updated after creation.

### Load Balancer

When creating a `ScalewayCluster`, a "main" Load Balancer is always created.
It is also possible to specify "extra" Load Balancers to achieve regional redundancy.

#### Main Load Balancer

The main Load Balancer is always created by default, it is not possible to disable it.

Here is an example of main Load Balancer configuration:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  region: fr-par
  network:
    controlPlaneLoadBalancer:
      type: LB-S
      port: 443
      zone: fr-par-1
      ip: 42.42.42.42 # optional
  # some fields were omitted...
```

- The `type` field can be updated to migrate the Load Balancer to another type.
- The `port` field specifies the port of the Load Balancer frontend that exposes the kube-apiserver(s).
  This field is immutable.
- The `zone` field specifies where the Load Balancer will be created. Must be in the same region
  as the `ScalewayCluster` region. This defaults to the first availability zone of the region.
  This field is immutable.
- The `ip` field specifies an existing Load Balancer Flexible IP to use when creating
  the Load Balancer. If not set, a new IP will be created. This field is immutable.

#### Extra Load Balancers

To specify extra Load Balancers, it is required to also set the `network.controlPlaneDNS` field.

Here is an example that configures two extra Load Balancers in `nl-ams-2` and `nl-ams-3`.

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
metadata:
  name: my-cluster
spec:
  region: nl-ams
  network:
    controlPlaneLoadBalancer:
      type: LB-S
      zone: nl-ams-1
    controlPlaneExtraLoadBalancers:
      - type: LB-S
        zone: nl-ams-2
      - type: LB-S
        zone: nl-ams-3
    controlPlaneDNS:
      domain: subdomain.your-domain.com
      name: my-cluster
  # some fields were omitted...
```

> [!WARNING]
> You can freely add and remove extra Load Balancers, however, some requests to
> the workload cluster's API server may fail as the Load Balancers are reconfigured.

#### Allowed ranges (ACLs)

The workload cluster's API server is always exposed publicly though the Load Balancer(s).
To prevent unauthorized access to the API server, you may configure some allowed network ranges
in CIDR format. By default, when the `allowedRanges` is unset or set to an empty list (`[]`),
all network ranges are allowed.

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  network:
    controlPlaneLoadBalancer:
      allowedRanges:
        - 42.42.0.0/16
        - 1.1.1.1/32
  # some fields were omitted...
```

> [!NOTE]
> The public IPs of the nodes and Public Gateways will automatically be allowed.
> No additional configuration is required.

### VPC

#### Private Network

To automatically attach the resources of the cluster to a Private Network, simply
set the `network.privateNetwork.enabled` field to true:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  network:
    privateNetwork:
      enabled: true
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

#### Public Gateways

To create `ScalewayMachines` without a Public IP, your Private Network must contain
at least one Public Gateway that advertises its default route. You can configure one
manually or let the provider configure that for you:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1
kind: ScalewayCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  region: fr-par
  network:
    privateNetwork:
      enabled: true
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
> 🚮 Updating a Public Gateway will lead to its recreation, which will make its private IP change.
>
> ⏳ Because the default routes are advertised via DHCP, the DHCP leases of the nodes must
> be renewed for changes to be propagated (~24 hours). You can reboot the nodes or
> trigger a rollout restart of the `kubeadmcontrolplanes`/`machinedeployments` to force the propagation.
