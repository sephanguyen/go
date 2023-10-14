package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CompareTagArrays(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name        string
		NewArray    []string
		OldArray    []string
		ArrayInsert []string
		ArrayRemove []string
	}{
		{
			Name:        "happy case",
			NewArray:    []string{"A", "C", "D", "F"},
			OldArray:    []string{"A", "B", "C", "T"},
			ArrayInsert: []string{"D", "F"},
			ArrayRemove: []string{"B", "T"},
		},
		{
			Name:        "happy case nothing to insert/remove",
			NewArray:    []string{"A", "B", "C", "D"},
			OldArray:    []string{"A", "B", "C", "D"},
			ArrayInsert: []string{},
			ArrayRemove: []string{},
		},
		{
			Name:        "happy case test",
			NewArray:    []string{"tag4", "tag1"},
			OldArray:    []string{"tag2", "tag3", "tag5"},
			ArrayInsert: []string{"tag4", "tag1"},
			ArrayRemove: []string{"tag2", "tag3", "tag5"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			insArr, rmArr := CompareTagArrays(testCase.NewArray, testCase.OldArray)
			if len(insArr) > 0 {
				assert.EqualValues(t, testCase.ArrayInsert, insArr)
			} else {
				assert.Nil(t, insArr)
			}
			if len(rmArr) > 0 {
				assert.EqualValues(t, testCase.ArrayRemove, rmArr)
			} else {
				assert.Nil(t, rmArr)
			}
		})
	}
}
