# Scaleway Secret

When creating a `ScalewayCluster`, it is required to specify in the `scalewaySecretName` field
the name of an existing `Secret` in the same namespace as the `ScalewayCluster`.

## Secret specification

The Secret can have the following data keys:

| Key            | Required | Description                                               |
| -------------- | -------- | --------------------------------------------------------- |
| SCW_ACCESS_KEY | yes      | Your Scaleway Access Key                                  |
| SCW_SECRET_KEY | yes      | You Scaleway Secret Key                                   |
| SCW_API_URL    | no       | Scaleway API URL. Defaults to <https://api.scaleway.com> |

**All other keys inside the secret will be ignored.**

## Permission sets

Your Scaleway API Key must have the following permission sets:

> [!WARNING]
> This list may change when new features are added to the provider. Make sure you
> read the changelogs before upgrading the provider.

- `IPAMFullAccess`
- `LoadBalancersFullAccess`
- `VPCFullAccess`
- `PrivateNetworksFullAccess`
- `VPCGatewayFullAccess`
- `BlockStorageFullAccess`
- `DomainsDNSFullAccess`
- `InstancesFullAccess`

If a permission set is missing, you may encounter reconcile errors in the logs of the provider.

## Example

Here is an example of valid secret, please update the values of `SCW_ACCESS_KEY`
and `SCW_SECRET_KEY` before applying it:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-scaleway-secret
  namespace: default
type: Opaque
stringData:
  SCW_ACCESS_KEY: SCW11111111111111111
  SCW_SECRET_KEY: 11111111-1111-1111-1111-111111111111
```
