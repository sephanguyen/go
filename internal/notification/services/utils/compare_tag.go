package utils

import "golang.org/x/exp/slices"

func CompareTagArrays(newTags, oldTags []string) ([]string, []string) {
	var arrInsert, arrRemove []string
	for _, val := range newTags {
		// insert new tag
		if !slices.Contains(oldTags, val) {
			arrInsert = append(arrInsert, val)
		}
	}

	for _, val := range oldTags {
		// remove old tag
		if !slices.Contains(newTags, val) {
			arrRemove = append(arrRemove, val)
		}
	}
	return arrInsert, arrRemove
}
