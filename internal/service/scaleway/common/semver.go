package common

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// IsUpToDate compares current and desired semver and returns true if current >= desired.
func IsUpToDate(current, desired string) (bool, error) {
	curr, err := semver.StrictNewVersion(current)
	if err != nil {
		return false, fmt.Errorf("failed to parse current version: %w", err)
	}
	desi, err := semver.StrictNewVersion(desired)
	if err != nil {
		return false, fmt.Errorf("failed to parse desired version: %w", err)
	}

	return curr.GreaterThanEqual(desi), nil
}
