# Kubernetes Cluster API Provider Scaleway

> [!WARNING]
> **This project is currently in its alpha stage**, which means it is still under active development.
> As such, features are subject to change, and breaking changes may occur without notice.
> We recommend using it with caution in production environments and keeping up to date
> with the latest updates and documentation.

------

## What is the Cluster API Provider Scaleway (CAPS)

The [Cluster API](https://github.com/kubernetes-sigs/cluster-api) brings declarative, Kubernetes-style APIs to cluster creation, configuration and management.

CAPS is a Cluster API infrastructure provider that enables efficient management at
scale of self-managed clusters on Scaleway.

## Quick Start

Check out the [getting started](./docs/getting-started.md) to create your first
Kubernetes cluster on Scaleway using Cluster API.

## Getting Help

If you need help with CAPS, please visit the #cluster-api channel on
[Scaleway Slack community](https://slack.scaleway.com/) or open a GitHub issue.

------

## Compatibility

### Cluster API Versions

This provider's versions are compatible with the following versions of Cluster API:

| Scaleway Provider CRD version | Cluster API `v1beta1` (v1.0-v1.10) | Cluster API `v1beta2` (v1.11+) |
| ----------------------------- | ---------------------------------- | ------------------------------ |
| `v1alpha1` (v0.1.x)           | ✓                                  | ☓                              |
| `v1alpha2` (v0.2.x, main)     | ☓                                  | ✓                              |

### Kubernetes Versions

The Scaleway provider is able to install and manage the [versions of Kubernetes supported by the Cluster API (CAPI) project](https://cluster-api.sigs.k8s.io/reference/versions.html#supported-kubernetes-versions).

------

## Getting involved and contributing

Are you interested in contributing to cluster-api-provider-scaleway? We would love
your suggestions, contributions, and help!

To set up your environment checkout the [development guide](./docs/development.md).

## Github issues

### Bugs

If you think you have found a bug please follow the instructions below.

- Please spend a small amount of time giving due diligence to the issue tracker. Your issue might be a duplicate.
- Get the logs from the cluster controllers. Please paste this into your issue.
- Open a new issue.
- Remember users might be searching for your issue in the future, so please give it a meaningful title to help others.
- Feel free to reach out to the #cluster-api channel on [Scaleway Slack community](https://slack.scaleway.com/)
