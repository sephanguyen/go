package support

import (
	"reflect"
	"testing"
)

func TestUtils_ConvertAcademicWeeks(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string

		expected []string
	}{
		{
			input:    "1-10",
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
		{
			input:    "1_2_3-12_17",
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "17"},
		},
		{
			input:    "1_2_3-12",
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		},
		{
			input:    "1",
			expected: []string{"1"},
		},
		{
			input:    "1_6_7",
			expected: []string{"1", "6", "7"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()
			data, _ := ConvertAcademicWeeks(tc.input)
			if !reflect.DeepEqual(data, tc.expected) {
				t.Errorf("ConvertAcademicWeeks(%v) = %v, expected: %v", tc.input, data, tc.expected)
			}
		})
	}
}
