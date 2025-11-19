# Advanced configurations

## Use private IP in Kubeadm configurations

When your nodes have both public and private IPs, Kubeadm will always advertise
the public IP by default. This can be an issue if you want the control-plane to
communicate through the Private Network and then block all public ingress traffic
on your nodes using an Instance security group.

To solve this, you can include the node's private IP in your Kubeadm configurations using
the `[[[ .NodeIP ]]]` placeholder value. This placeholder value will be replaced
by the provider with the private IP of the node. If a Private Network is not enabled
in the `ScalewayCluster`, `[[[ .NodeIP ]]]` will be replaced with the public IP
of the node instead.

Here is an example of `KubeadmControlPlane` configuration:

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta2
kind: KubeadmControlPlane
metadata:
  name: my-kubeadm-controlplane
  namespace: default
spec:
  kubeadmConfigSpec:
    initConfiguration:
      localAPIEndpoint:
        advertiseAddress: "[[[ .NodeIP ]]]"
      nodeRegistration:
        kubeletExtraArgs:
          - name: node-ip
            value: "[[[ .NodeIP ]]]"
    joinConfiguration:
      controlPlane:
        localAPIEndpoint:
          advertiseAddress: "[[[ .NodeIP ]]]"
      nodeRegistration:
        kubeletExtraArgs:
          - name: node-ip
            value: "[[[ .NodeIP ]]]"
  # important: some fields were omitted...
```

Here is an example of `KubeadmConfigTemplate` configuration:

```yaml
apiVersion: bootstrap.cluster.x-k8s.io/v1beta2
kind: KubeadmConfigTemplate
metadata:
  name: my-kubeadmconfig-template
  namespace: default
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          kubeletExtraArgs:
            - name: node-ip
              value: "[[[ .NodeIP ]]]"
  # important: some fields were omitted...
```
