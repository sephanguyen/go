package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name     string
	expected string
	input    string
}

func TestReplaceSpecialChars(t *testing.T) {
	testCases := []testCase{
		{
			name:     "backward compatibility",
			input:    "exam lo",
			expected: "exam lo",
		},
		{
			name:     "search with % character",
			input:    "exam lo 100%",
			expected: `exam lo 100\%`,
		},
		{
			name:     `search with \ character`,
			input:    `exam\lo`,
			expected: `exam\\lo`,
		},
		{
			name:     "search with _ character",
			input:    "exam_lo",
			expected: `exam\_lo`,
		},
		{
			name:     "search with multiple special characters",
			input:    "exam_lo_100%",
			expected: `exam\_lo\_100\%`,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := ReplaceSpecialChars(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
