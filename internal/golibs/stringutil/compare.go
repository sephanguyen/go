package stringutil

import (
	"sort"
)

// SliceEqual returns true if two string slices are equal, ordered.
func SliceEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, item1 := range s1 {
		if item1 != s2[i] {
			return false
		}
	}
	return true
}

// SliceElementsMatch returns true if two string slices contains
// the same elements, regardless of order.
// The complexity is O(NlogN).
func SliceElementsMatch(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	sort.Strings(s1)
	sort.Strings(s2)
	for i, item1 := range s1 {
		if item1 != s2[i] {
			return false
		}
	}
	return true
}

// SliceElementsDiff return elements in a that is not in b
func SliceElementsDiff(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var removed []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			removed = append(removed, x)
		}
	}
	return removed
}
