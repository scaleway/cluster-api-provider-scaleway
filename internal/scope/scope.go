package scope

import "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"

type Interface interface {
	Cloud() client.Interface
	SetCloud(client.Interface) // SetCloud is used for testing.
	ResourceName(suffixes ...string) string
	ResourceTags(additional ...string) []string
}
