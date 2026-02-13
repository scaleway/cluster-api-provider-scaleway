# Kubernetes Cluster API Provider Scaleway

> [!WARNING]
> **This project is currently in its alpha stage**, which means it is still under active development.
> As such, features are subject to change, and breaking changes may occur without notice.
> We recommend using it with caution in production environments and keeping up to date
> with the latest updates and documentation.

------

<<<<<<< HEAD
### Prerequisites
- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.
=======
## What is the Cluster API Provider Scaleway (CAPS)
>>>>>>> tmp-original-13-02-26-16-17

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

<<<<<<< HEAD
### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/cluster-api-provider-scaleway:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/cluster-api-provider-scaleway/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
=======
If you think you have found a bug please follow the instructions below.
>>>>>>> tmp-original-13-02-26-16-17

- Please spend a small amount of time giving due diligence to the issue tracker. Your issue might be a duplicate.
- Get the logs from the cluster controllers. Please paste this into your issue.
- Open a new issue.
- Remember users might be searching for your issue in the future, so please give it a meaningful title to help others.
- Feel free to reach out to the #cluster-api channel on [Scaleway Slack community](https://slack.scaleway.com/)
