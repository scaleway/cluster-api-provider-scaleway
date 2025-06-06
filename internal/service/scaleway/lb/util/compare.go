package util

import (
	"slices"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func compareACLs(a, b *lb.ACL) int {
	return strings.Compare(a.Name, b.Name)
}

func removePtr[T any](list []*T) []T {
	result := make([]T, 0, len(list))

	for _, l := range list {
		result = append(result, *l)
	}

	return result
}

// IPsEqual compares two slices of pointers to strings representing IP addresses.
// It returns true if both slices contain the same IP addresses, regardless of order.
func IPsEqual(a, b []*string) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := slices.Sorted(slices.Values(removePtr(a)))
	sortedB := slices.Sorted(slices.Values(removePtr(b)))

	for i, a := range sortedA {
		if a != sortedB[i] {
			return false
		}
	}

	return true
}

// ACLEqual compares two slices of pointers to lb.ACL objects.
// It returns true if both slices contain the same ACLs, regardless of order.
func ACLEqual(a, b []*lb.ACL) bool {
	if len(a) != len(b) {
		return false
	}

	// Sort both lists by name
	slices.SortFunc(a, compareACLs)
	slices.SortFunc(b, compareACLs)

	for i, acl := range a {
		if acl.Name != b[i].Name {
			return false
		}

		if acl.Index != b[i].Index {
			return false
		}

		if (acl.Action == nil) != (b[i].Action == nil) {
			return false
		}

		if acl.Action != nil && b[i].Action != nil && acl.Action.Type != b[i].Action.Type {
			return false
		}

		if (acl.Match == nil) != (b[i].Match == nil) {
			return false
		}

		if acl.Match != nil && b[i].Match != nil && !IPsEqual(acl.Match.IPSubnet, b[i].Match.IPSubnet) {
			return false
		}
	}

	return true
}
