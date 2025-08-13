package common

import (
	"cmp"
	"slices"
)

// SlicesEqualIgnoreOrder returns true if both slices are equal, regardless of order.
func SlicesEqualIgnoreOrder[T cmp.Ordered](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	return slices.Equal(slices.Sorted(slices.Values(a)), slices.Sorted(slices.Values(b)))
}
