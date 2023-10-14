package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleMatch(t *testing.T) {
	t.Parallel()
	type testcase struct {
		rule        rule
		inputs      []string
		matched     bool
		matchedVals []string
	}

	testcases := []testcase{
		{
			rule:    rule{Name: "1", Paths: []string{"skaffold.+"}},
			inputs:  []string{"skaffold.yaml", "README.md"},
			matched: true,
		},
		{
			rule:    rule{Name: "2", Paths: []string{"skaffold.ya?ml"}},
			inputs:  []string{"skaffold.yaml", "README.md"},
			matched: true,
		},
		{
			rule:    rule{Name: "3", Paths: []string{"skaffold.ya?ml"}, RunOnly: true},
			inputs:  []string{"skaffold.yaml", "README.md"},
			matched: false,
		},
		{
			rule:    rule{Name: "4", Paths: []string{"skaffold.ya?ml", "README.md"}, RunOnly: true},
			inputs:  []string{"skaffold.yaml", "README.md"},
			matched: true,
		},
		{
			rule:    rule{Name: "5", Paths: []string{"a/b/"}},
			inputs:  []string{"a/b/c/skaffold.yaml", "a/b/skaffold.yaml"},
			matched: true,
		},
		{
			rule:    rule{Name: "6", Paths: []string{`.+\.go`}},
			inputs:  []string{"a/b/c/d.go"},
			matched: true,
		},
		{
			rule:    rule{Name: "7", Paths: []string{`.+\.go`}},
			inputs:  []string{"a/b/c/some.file.yaml"},
			matched: false,
		},
		{
			rule:    rule{Name: "8", Paths: []string{`skaffold\..*\.yaml`}},
			inputs:  []string{"skaffold.a.yaml"},
			matched: true,
		},
		{
			rule:    rule{Name: "9", Paths: []string{`skaffold\..*\.yaml`}},
			inputs:  []string{"skaffold.yaml"},
			matched: false,
		},
		{
			rule:    rule{Name: "10", Paths: []string{`skaffold\..*\.yaml`}},
			inputs:  []string{"README.md"},
			matched: false,
		},
		{
			rule:        rule{Name: "11", Paths: []string{`skaffold\.{{.VALUE}}\.yaml`}, Values: []string{"a", "b", "c", "d"}},
			inputs:      []string{"skaffold.a.yaml", "skaffold.b.yaml", "skaffold.z.yaml"},
			matched:     true,
			matchedVals: []string{"a", "b"},
		},
		{
			rule:    rule{Name: "12", Paths: []string{`skaffold\.{{.VALUE}}\.yaml`}, Values: []string{"a", "b", "c", "d"}},
			inputs:  []string{"skaffold.z.yaml"},
			matched: false,
		},
		{
			rule:    rule{Name: "13", Paths: []string{`file.a`, `file.b`}, PathsIgnore: []string{`file.b`, `file.c`}},
			inputs:  []string{"file.a"},
			matched: true,
		},
		{
			rule:    rule{Name: "14", Paths: []string{`file.a`, `file.b`}, PathsIgnore: []string{`file.b`, `file.c`}, RunOnly: true},
			inputs:  []string{"file.a"},
			matched: true,
		},
		{
			rule:    rule{Name: "15", Paths: []string{`file.a`, `file.b`}, PathsIgnore: []string{`file.b`, `file.c`}, RunOnly: true},
			inputs:  []string{"file.a", "file.c"},
			matched: false,
		},
		{
			rule:    rule{Name: "16", Paths: []string{`file.a`, `file.b`}, PathsIgnore: []string{`file.b`, `file.c`}},
			inputs:  []string{"file.b"},
			matched: false,
		},
		{
			rule:    rule{Name: "17", Paths: []string{`file.a`, `file.b`}, PathsIgnore: []string{`file.b`, `file.c`}},
			inputs:  []string{"file.a", "file.b"},
			matched: true,
		},
		{
			rule:    rule{Name: "18", Paths: []string{`file.a`, `file.b`}, PathsIgnore: []string{`file.b`, `file.c`}, RunOnly: true},
			inputs:  []string{"file.a", "file.b"},
			matched: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.rule.Name, func(t *testing.T) {
			matched, matchedVals, err := tc.rule.match(tc.inputs, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.matched, matched)
			assert.ElementsMatch(t, tc.matchedVals, matchedVals)
		})
	}
}

func TestRuleListResultFromPRDesc2(t *testing.T) {
	t.Parallel()
	r := ruleList{
		Rules: []rule{
			{Name: "integration-test"},
			{Name: "integration-blocker-test"},
			{Name: "unit-test"},
			{Name: "e2e-test"},
			{Name: "lint"},
			{Name: "svcs_change", Values: []string{"bob", "tom"}},
		},
	}

	type testcase struct {
		description string
		input       string
		expected    map[string]string
	}

	testcases := []testcase{
		{
			description: "test case 1", input: "- test:unit-test",
			expected: map[string]string{"integration-test": "0", "integration-blocker-test": "0", "run_all_integration_test": "0", "unit-test": "1", "lint": "0", "svcs_change": "", "e2e-test": "0"},
		},
		{
			description: "test case 2", input: "- test:lint",
			expected: map[string]string{"integration-test": "0", "integration-blocker-test": "0", "run_all_integration_test": "0", "unit-test": "0", "lint": "1", "svcs_change": "", "e2e-test": "0"},
		},
		{
			description: "test case 3", input: "- test:e2e-test",
			expected: map[string]string{"integration-test": "0", "integration-blocker-test": "0", "run_all_integration_test": "0", "unit-test": "0", "lint": "0", "svcs_change": "", "e2e-test": "1"},
		},
		{
			description: "test case 4", input: "- test:integration-test",
			expected: map[string]string{"integration-test": "1", "integration-blocker-test": "0", "run_all_integration_test": "1", "unit-test": "0", "lint": "0", "svcs_change": "", "e2e-test": "0"},
		},
		{
			description: "test case 5", input: "- test:integration-blocker-test",
			expected: map[string]string{"integration-test": "0", "integration-blocker-test": "1", "run_all_integration_test": "0", "unit-test": "0", "lint": "0", "svcs_change": "", "e2e-test": "0"},
		},
		{
			description: "test case 6", input: "- test:integration-test:bob\n- test:integration-test:eureka",
			expected: map[string]string{"integration-test": "1", "integration-blocker-test": "0", "run_all_integration_test": "0", "unit-test": "0", "lint": "0", "svcs_change": "", "extra_integration_test_args": "bob eureka ", "e2e-test": "0"},
		},
		{
			description: "test case 7", input: "- test:integration-blocker-test:bob\n- test:integration-blocker-test:eureka",
			expected: map[string]string{"integration-test": "0", "integration-blocker-test": "1", "run_all_integration_test": "0", "unit-test": "0", "lint": "0", "svcs_change": "", "extra_integration_test_args": "bob eureka ", "e2e-test": "0"},
		},
		{
			description: "test case 8", input: "- test:lint\n-test:unit-test\n-test:integration-test",
			expected: map[string]string{"integration-test": "1", "integration-blocker-test": "0", "run_all_integration_test": "1", "unit-test": "1", "lint": "1", "svcs_change": "", "e2e-test": "0"},
		},
		{
			description: "test case 9", input: "- test:lint\n-test:unit-test\n-test:integration-blocker-test",
			expected: map[string]string{"integration-test": "0", "integration-blocker-test": "1", "run_all_integration_test": "0", "unit-test": "1", "lint": "1", "svcs_change": "", "e2e-test": "0"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			actual, err := r.resultFromPRDesc2(tc.input)
			require.NoError(t, err)
			assert.Equal(t, result(tc.expected), *actual)
		})
	}
}

func TestRuleListForceValues(t *testing.T) {
	t.Parallel()
	type testcase struct {
		desc   string // description for the test case
		input  ruleList
		expect result
	}

	testcases := []testcase{
		{
			desc: "normal case",
			input: ruleList{[]rule{
				{Name: "r1"},
				{Name: "r2", RunOnly: true},
				{Name: "r3", Values: []string{"a", "b"}},
				{Name: "r3", Values: []string{"b", "c"}},
			}},
			expect: map[string]string{
				"r1": "1",
				"r2": "0",
				"r3": "a b c",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			result := tc.input.forceValues()
			assert.Equal(t, tc.expect, result)
		})
	}
}
