package version

import (
	"runtime/debug"
	"slices"
)

// Version is the version of cluster-api-provider-scaleway.
// This variable should be set during build.
var Version = defaultVersion

const (
	defaultVersion = "dev"
	modulePath     = "github.com/scaleway/cluster-api-provider-scaleway"
)

func init() {
	// Do nothing if Version was already overridden during build time.
	if Version != defaultVersion {
		return
	}

	// Find the module version in build info, in case it's imported by another module.
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range info.Deps {
			if dep.Path == modulePath {
				if slices.Contains([]string{"(devel)", ""}, dep.Version) {
					break
				}

				Version = dep.Version
			}
		}
	}
}
