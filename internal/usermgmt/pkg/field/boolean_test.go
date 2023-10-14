package field

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoolean(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          bool
		expectedOutput Boolean
	}{
		{
			name:  "init Boolean with zero value",
			input: false,
			expectedOutput: Boolean{
				value:  false,
				status: StatusPresent,
			},
		},
		{
			name:  "init Boolean with valid value",
			input: true,
			expectedOutput: Boolean{
				value:  true,
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewBoolean(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullBoolean(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput Boolean
	}{
		{
			name: "init null Boolean",
			expectedOutput: Boolean{
				value:  false,
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullBoolean() == testCase.expectedOutput)
		})
	}
}

func TestBoolean_Ptr(t *testing.T) {
	boolField := NewBoolean(true)

	assert.Equal(t, &boolField, boolField.Ptr())
}

func TestBoolean_Boolean(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Boolean
		expectedOutput bool
	}{
		{
			name:           "undefined Boolean field must return false when enforced returning value",
			input:          NewUndefinedBoolean(),
			expectedOutput: false,
		},
		{
			name:           "null Boolean field must return false when enforced returning value",
			input:          NewNullBoolean(),
			expectedOutput: false,
		},
		{
			name:           "present false Boolean field must return false when enforced returning value",
			input:          NewBoolean(false),
			expectedOutput: false,
		},
		{
			name:           "present true Boolean field must return false when enforced returning value",
			input:          NewBoolean(true),
			expectedOutput: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, testCase.input.Boolean())
		})
	}
}

func TestBooleans_Booleans(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		input          Booleans
		expectedOutput []bool
	}

	testCases := []testCase{
		{
			name:           "nil slice",
			input:          nil,
			expectedOutput: []bool{},
		},
		{
			name:           "empty slice",
			input:          Booleans{},
			expectedOutput: []bool{},
		},
		{
			name: "there are values in slice",
			input: Booleans{
				NewBoolean(true),
				NewBoolean(false),
				NewNullBoolean(),
			},
			expectedOutput: []bool{true, false, false},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, len(testCase.expectedOutput), len(testCase.input.Booleans()))
		assert.Equal(t, testCase.expectedOutput, testCase.input.Booleans())
	}
}
