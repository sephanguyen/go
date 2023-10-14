package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLowerCaseFirstLetter(t *testing.T) {
	type testCase struct {
		name           string
		input          string
		expectedOutput string
	}

	testCases := []testCase{
		{
			name:           "input is empty",
			input:          "",
			expectedOutput: "",
		},
		{
			name:           "length of input is 1",
			input:          "A",
			expectedOutput: "a",
		},
		{
			name:           "length of input is 2",
			input:          "AB",
			expectedOutput: "aB",
		},
		{
			name:           "length of input is larger than 2",
			input:          "ABC",
			expectedOutput: "aBC",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, LowerCaseFirstLetter(testCase.input))
		})
	}
}

func TestUpperCaseFirstLetter(t *testing.T) {
	type testCase struct {
		name           string
		input          string
		expectedOutput string
	}

	testCases := []testCase{
		{
			name:           "input is empty",
			input:          "",
			expectedOutput: "",
		},
		{
			name:           "length of input is 1",
			input:          "a",
			expectedOutput: "A",
		},
		{
			name:           "length of input is 2",
			input:          "ab",
			expectedOutput: "Ab",
		},
		{
			name:           "length of input is larger than 2",
			input:          "abc",
			expectedOutput: "Abc",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, UpperCaseFirstLetter(testCase.input))
		})
	}
}
