package diff

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDifferParseOptions(t *testing.T) {
	t.Parallel()
	type testcase struct {
		desc         string
		d            Differ
		expectedErr  error
		validateFunc func(d Differ)
	}

	testcases := []testcase{
		{desc: "nothing specified", d: Differ{}, expectedErr: errMissingAll},
		{desc: "missing --head-ref", d: Differ{BaseRef: "abc"}, expectedErr: errMissingGitHeadRef},
		{desc: "missing --base-ref", d: Differ{HeadRef: "abc"}, expectedErr: errMissingGitBaseRef},
		{desc: "missing --pr-desc", d: Differ{PRDescOnly: true}, expectedErr: errMissingPRDesc},
		{desc: "normal mode", d: Differ{BaseRef: "a", HeadRef: "b", PRDesc: "c"}, expectedErr: nil},
		{desc: "--pr-desc-only mode", d: Differ{PRDesc: "c", PRDescOnly: true}, expectedErr: nil},
		{
			desc:        "parse --squads",
			d:           Differ{BaseRef: "a", HeadRef: "b", Squads: []string{"dev-b", "squad-a"}},
			expectedErr: nil,
			validateFunc: func(d Differ) {
				assert.Equal(t, "squad-a", *d.targetSquad)
			},
		},
		{
			desc:        "parse --squads with multiple valid inputs",
			d:           Differ{BaseRef: "a", HeadRef: "b", Squads: []string{"squad-a", "squad-b"}},
			expectedErr: nil,
			validateFunc: func(d Differ) {
				assert.Equal(t, "squad-a", *d.targetSquad)
			},
		},
		{
			desc:        "parse --squads with ignored squads",
			d:           Differ{BaseRef: "a", HeadRef: "b", Squads: []string{"squad-admin", "squad-release", "squad-ddd", "squad-data", "squad-b"}},
			expectedErr: nil,
			validateFunc: func(d Differ) {
				assert.Equal(t, "squad-b", *d.targetSquad)
			},
		},
		{
			desc:        "parse --squads with ignored squad suffixes",
			d:           Differ{BaseRef: "a", HeadRef: "b", Squads: []string{"squad-a-be", "squad-b-fe", "squad-c-me", "squad-d-red", "squad-c"}},
			expectedErr: nil,
			validateFunc: func(d Differ) {
				assert.Equal(t, "squad-c", *d.targetSquad)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			actualErr := tc.d.parseOptions()
			assert.Equal(t, tc.expectedErr, actualErr)
			if tc.validateFunc != nil {
				tc.validateFunc(tc.d)
			}
		})
	}
}

func TestDifferRun(t *testing.T) {
	type testcase struct {
		desc     string
		d        Differ
		expected []string
	}

	prDesc := "<!-- some\nmultiline\ncomment\n-->\n\n### Summary of your PR\n- test:integration-blocker-test:bob/delete_lesson.feature\n- test:unit-test"
	testcases := []testcase{
		{
			desc: "--force",
			d: Differ{
				Force:      true,
				ConfigPath: getTestDiffConfig("1.force_test.yaml"),
			},
			expected: []string{"rule1=1", "rule2=0", "rule3=a b c d"},
		},
		{
			desc: "--force",
			d: Differ{
				Force:      true,
				ConfigPath: getTestDiffConfig("1.force_test.yaml"),
				PRDesc:     "- test:rule2",
				PRDescOnly: true,
			},
			expected: []string{"rule1=1", "rule2=0", "rule3=a b c d"},
		},
		{
			desc: "--pr-desc-only",
			expected: []string{
				"e2e-test=0",
				"extra_integration_test_args=bob/delete_lesson.feature ",
				"integration-blocker-test=1",
				"lint=0",
				"run_all_integration_test=0",
				"svcs_change=",
				"unit-test=1",
				"pr_desc_only=1",
			},
			d: Differ{
				PRDesc:     prDesc,
				PRDescOnly: true,
				ConfigPath: getTestDiffConfig("2.pr_desc_only.yaml"),
			},
		},
		{
			desc: "--pr-desc-only should ignore --head-ref and --base-ref",
			expected: []string{
				"all=0",
				"e2e-test=0",
				"extra_integration_test_args=bob/delete_lesson.feature ",
				"go=0",
				"goOnly=0",
				"integration-blocker-test=1",
				"multipleValues=",
				"run_all_integration_test=0",
				"svcs_change=",
				"unit-test=1",
				"pr_desc_only=1",
			},
			d: Differ{
				PRDesc:     prDesc,
				PRDescOnly: true,
				ConfigPath: getTestDiffConfig("3.git_diff.yaml"),
				BaseRef:    "45b09733a9959d704cba0d01641ba5d400243230",
				HeadRef:    "5c126559ff5a252dc51fbaac64a3dcbbf96c0844",
			},
		},
		{
			desc: "--pr-desc-only with base64 PR description input",
			expected: []string{
				"e2e-test=0",
				"extra_integration_test_args=bob/delete_lesson.feature ",
				"integration-blocker-test=1",
				"lint=1",
				"run_all_integration_test=0",
				"svcs_change=",
				"unit-test=0",
				"pr_desc_only=1",
			},
			d: Differ{
				PRDesc:     "Q0k6Ci0gdGVzdDppbnRlZ3JhdGlvbi1ibG9ja2VyLXRlc3Q6Ym9iL2RlbGV0ZV9sZXNzb24uZmVhdHVyZQotIHRlc3Q6bGludAo=",
				PRDescOnly: true,
				ConfigPath: getTestDiffConfig("2.pr_desc_only.yaml"),
			},
		},
		{
			desc: "work with git diff",
			d: Differ{
				BaseRef:    "45b09733a9959d704cba0d01641ba5d400243230",
				HeadRef:    "5c126559ff5a252dc51fbaac64a3dcbbf96c0844",
				ConfigPath: getTestDiffConfig("3.git_diff.yaml"),
			},
			expected: []string{
				"all=1",
				"e2e-test=0",
				"go=1",
				"goOnly=0",
				"integration-blocker-test=0",
				"multipleValues=README.md helloworld.go ",
				"run_all_integration_test=0",
				"svcs_change=",
				"unit-test=0",
			},
		},
		{
			desc: "work with git diff plus --pr-desc-only",
			d: Differ{
				BaseRef:    "45b09733a9959d704cba0d01641ba5d400243230",
				HeadRef:    "5c126559ff5a252dc51fbaac64a3dcbbf96c0844",
				PRDesc:     "CI\n\n- test:unit-test\n- test:e2e-test\n- test:integration-blocker-test:bob",
				ConfigPath: getTestDiffConfig("3.git_diff.yaml"),
			},
			expected: []string{
				"all=1",
				"e2e-test=1",
				"extra_integration_test_args=bob ",
				"go=1",
				"goOnly=0",
				"integration-blocker-test=1",
				"multipleValues=README.md helloworld.go ",
				"run_all_integration_test=0",
				"svcs_change=",
				"unit-test=1",
			},
		},
		{
			desc: "--squads=A",
			d: Differ{
				BaseRef:    "45b09733a9959d704cba0d01641ba5d400243230",
				HeadRef:    "5c126559ff5a252dc51fbaac64a3dcbbf96c0844",
				ConfigPath: getTestDiffConfig("4.squad_based.yaml"),
				Squads:     []string{"squad-a"},
			},
			expected: []string{"r1=1", "r2=1", "r3=0", "run_all_integration_test=0", "svcs_change="},
		},
		{
			desc: "--squads=B",
			d: Differ{
				BaseRef:    "45b09733a9959d704cba0d01641ba5d400243230",
				HeadRef:    "5c126559ff5a252dc51fbaac64a3dcbbf96c0844",
				ConfigPath: getTestDiffConfig("4.squad_based.yaml"),
				Squads:     []string{"squad-b"},
			},
			expected: []string{"r1=1", "r2=1", "r3=0", "run_all_integration_test=0", "svcs_change="},
		},
		{
			desc: "--squads=C",
			d: Differ{
				BaseRef:    "45b09733a9959d704cba0d01641ba5d400243230",
				HeadRef:    "5c126559ff5a252dc51fbaac64a3dcbbf96c0844",
				ConfigPath: getTestDiffConfig("4.squad_based.yaml"),
				Squads:     []string{"squad-c"},
			},
			expected: []string{"r1=1", "r2=0", "r3=1", "run_all_integration_test=0", "svcs_change="},
		},
		{
			desc: "empty --squads",
			d: Differ{
				BaseRef:    "45b09733a9959d704cba0d01641ba5d400243230",
				HeadRef:    "5c126559ff5a252dc51fbaac64a3dcbbf96c0844",
				ConfigPath: getTestDiffConfig("4.squad_based.yaml"),
				Squads:     nil,
			},
			expected: []string{"r1=1", "r2=1", "r3=1", "run_all_integration_test=0", "svcs_change="},
		},
	}

	// PRDescOnly should ignore the base/head-ref comparison
	execwrapper.SetGitDir(t, execwrapper.MockedGitDir())
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			actual, err := tc.d.Output()
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, parseOutput(actual))
		})
	}
}

func getTestDiffConfig(filename string) string {
	return filepath.Join(execwrapper.RootDirectory(), "internal/golibs/ci/diff/testdata", filename)
}

// parseOutput breaks down the output string into one per line,
// ignoring the empty lines.
func parseOutput(raw string) []string {
	splitByNewline := strings.Split(raw, "\n")
	res := make([]string, 0, len(splitByNewline))
	for _, v := range splitByNewline {
		if v != "" {
			res = append(res, v)
		}
	}
	return res
}
