package statuscheck

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckerRun(t *testing.T) {
	t.Parallel()
	c := Checker{
		Data: `{
	"requirements": {"result": "success", "outputs": {"unit-test": "1", "lint": "0", "run-blocker-test": "1"}},
	"unit-test": {"result": "success", "outputs": {}},
	"lint": {"result": "skipped", "outputs": {}},
	"run-blocker-test": {"result": "success", "outputs": {}}
}`,
		RequiredJobs: []string{"unit-test", "lint", "run-blocker-test"},
	}

	err := c.Run()
	require.NoError(t, err)
}

func TestCheckerRunWithFailedJobs(t *testing.T) {
	t.Parallel()
	c := Checker{
		Data: `{
	"requirements": {"result": "success", "outputs": {"unit-test": "1", "lint": "0", "run-blocker-test": "1"}},
	"unit-test": {"result": "success", "outputs": {}},
	"lint": {"result": "skipped", "outputs": {}},
	"run-blocker-test": {"result": "failure", "outputs": {}}
}`,
		RequiredJobs: []string{"unit-test", "lint", "run-blocker-test"},
	}

	err := c.Run()
	require.EqualError(t, err, `expected output "success" for job "run-blocker-test", got "failure"`)
}

func TestCheckerRunWithMissingStatuses(t *testing.T) {
	t.Parallel()
	c := Checker{
		Data: `{
	"requirements": {"result": "success", "outputs": {"unit-test": "1", "lint": "0", "run-blocker-test": "1"}},
	"unit-test": {"result": "success", "outputs": {}},
	"lint": {"result": "skipped", "outputs": {}}
}`,
		RequiredJobs: []string{"unit-test", "lint", "run-blocker-test"},
	}

	err := c.Run()
	require.EqualError(t, err, `missing required job "run-blocker-test" from input data`)
}

func TestGhDataUnmarshalJSON(t *testing.T) {
	t.Parallel()
	input := []byte(`{
	"requirements": {
		"result": "success",
		"outputs": {
			"run_a": "0",
			"run_b": "1"
		}
	},
	"unit-test": {
		"result": "success",
		"outputs": {}
	},
	"lint": {
		"result": "cancelled",
		"outputs": {}
	},
	"run-blocker-test": {
		"result": "failure",
		"outputs": {}
	}
}`)
	actual := &ghData{}
	err := json.Unmarshal(input, actual)
	require.NoError(t, err)

	expected := &ghData{
		"requirements": {
			Result: statusSuccess,
			Outputs: map[string]string{
				"run_a": "0",
				"run_b": "1",
			},
		},
		"unit-test": {
			Result:  statusSuccess,
			Outputs: map[string]string{},
		},
		"lint": {
			Result:  statusCancelled,
			Outputs: map[string]string{},
		},
		"run-blocker-test": {
			Result:  statusFailure,
			Outputs: map[string]string{},
		},
	}
	assert.Equal(t, expected, actual)
}

func c(v ghData) *ghData {
	return &v
}

func TestGhDataGetExpectedResult(t *testing.T) {
	t.Parallel()
	type testcase struct {
		desc        string
		input       ghData
		expected    *ghData
		expectedErr error
	}

	mockRequiredJobs := []string{"a", "b"}

	testcases := []testcase{
		{
			desc: "normal",
			input: map[string]ghJobOutput{
				"requirements": {Outputs: map[string]string{"a": "1", "b": "1"}},
			},
			expected: c(map[string]ghJobOutput{"a": {Result: statusSuccess}, "b": {Result: statusSuccess}}),
		},
		{
			desc: "some are skipped",
			input: map[string]ghJobOutput{
				"requirements": {Outputs: map[string]string{"a": "1", "b": "0"}},
			},
			expected: c(map[string]ghJobOutput{"a": {Result: statusSuccess}, "b": {Result: statusSkipped}}),
		},
		{
			desc: "all are skipped",
			input: map[string]ghJobOutput{
				"requirements": {Outputs: map[string]string{"a": "0", "b": "0"}},
			},
			expected: c(map[string]ghJobOutput{"a": {Result: statusSkipped}, "b": {Result: statusSkipped}}),
		},
		{
			desc:        "missing activation field",
			input:       map[string]ghJobOutput{},
			expectedErr: errMissingActivation,
		},
		{
			desc: "missing some steps",
			input: map[string]ghJobOutput{
				"requirements": {Outputs: map[string]string{"a": "0"}},
			},
			expectedErr: errors.New(`missing required activation flag "b" in activation job`),
		},
	}
	for _, tc := range testcases {
		actual, err := tc.input.getExpectedResults(mockRequiredJobs)
		require.Equal(t, tc.expectedErr, err, "test case %q failed", tc.desc)
		if tc.expected != nil {
			assert.Equal(t, tc.expected, actual, "test case %q failed", tc.desc)
		}
	}
}

func TestGhDataVerify(t *testing.T) {
	t.Parallel()
	expectation := ghData{
		"a": ghJobOutput{Result: statusSuccess},
		"b": ghJobOutput{Result: statusSuccess},
		"c": ghJobOutput{Result: statusSkipped},
	}
	type testcase struct {
		desc        string
		input       ghData
		expectedErr string
	}

	testcases := []testcase{
		{
			desc: "success",
			input: map[string]ghJobOutput{
				"a": {Result: statusSuccess},
				"b": {Result: statusSuccess},
				"c": {Result: statusSkipped},
			},
		},
		{
			desc: "fail a",
			input: map[string]ghJobOutput{
				"a": {Result: statusFailure},
				"b": {Result: statusSuccess},
				"c": {Result: statusSkipped},
			},
			expectedErr: `expected output "success" for job "a", got "failure"`,
		},
		{
			desc: "unexpected skipped a",
			input: map[string]ghJobOutput{
				"a": {Result: statusSkipped},
				"b": {Result: statusSuccess},
				"c": {Result: statusSkipped},
			},
			expectedErr: `expected output "success" for job "a", got "skipped"`,
		},
		{
			desc: "missing d",
			input: map[string]ghJobOutput{
				"b": {Result: statusSuccess},
				"c": {Result: statusSkipped},
			},
			expectedErr: `missing required job "a" from input data`,
		},
		{
			desc: "a failed that caused b to be cancelled",
			input: map[string]ghJobOutput{
				"a": {Result: statusFailure},
				"b": {Result: statusCancelled},
				"c": {Result: statusSkipped},
			},
			expectedErr: `expected output "success" for job "a", got "failure"`,
		},
		{
			desc: "a & b somehow got cancelled",
			input: map[string]ghJobOutput{
				"a": {Result: statusCancelled},
				"b": {Result: statusCancelled},
				"c": {Result: statusSkipped},
			},
			expectedErr: `expected output "success" for job "a", got "cancelled"; expected output "success" for job "b", got "cancelled"`,
		},
	}
	for _, tc := range testcases {
		err := tc.input.verify(expectation)
		if tc.expectedErr != "" {
			assert.EqualError(t, err, tc.expectedErr, "test case %q failed", tc.desc)
		}
	}
}
