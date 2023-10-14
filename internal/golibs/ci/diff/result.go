package diff

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// result contains the result of the requirements in a map struct.
type result map[string]string

func (rs result) addFromRegexpmatches(allmatches [][]string) {
	for _, m := range allmatches {
		rs.addFromRegexpMatch2(m...)
	}
}

func (rs result) addFromRegexpMatch2(m ...string) {
	if len(m) <= 1 {
		panic("input's length should be at least 3 (full match and 2 submatches)")
	}
	ruleName := m[1]
	switch testType := m[1]; testType {
	case "integration-blocker-test":
		rs[ruleName] = "1"
		if m[2] != "" {
			extraTestKey := "extra_integration_test_args"
			if _, exists := rs[extraTestKey]; !exists {
				rs[extraTestKey] = ""
			}
			rs[extraTestKey] += (m[2] + " ")
		}
	case "integration-test":
		rs["integration-test"] = "1"
		if m[2] != "" { // test:integration:bob/maybe-with-a-file.feature
			if _, exists := rs["extra_integration_test_args"]; !exists {
				rs["extra_integration_test_args"] = ""
			}
			rs["extra_integration_test_args"] += (m[2] + " ")
		} else { // test:integration
			rs["run_all_integration_test"] = "1"
		}
	case "e2e-test":
		rs["e2e-test"] = "1"
		if m[2] != "" {
			if _, exists := rs["e2e_flags"]; !exists {
				rs["e2e_flags"] = ""
			}
			rs["e2e_flags"] += (m[2] + " ")
		}
	default:
		ruleName := m[1]
		rs[ruleName] = "1"
	}
}

func (rs result) combine(other result) {
	for k, v := range other {
		if v == "0" {
			continue
		}
		if v == "1" {
			rs[k] = "1"
			continue
		}
		if v != "" {
			rs[k] += (strings.TrimRight(v, " ") + " ")
		}
	}
}

// setPRDescOnlyMode adds "pr_desc_only=1" to the output to signify
// that --pr-desc-only is enabled.
func (rs result) setPRDescOnlyMode() {
	rs["pr_desc_only"] = "1"
}

// output writes the formatted result to f. Generally, this output
// is consumed by Github Action, e.g. echo $(go run main.go) >> $GITHUB_OUTPUT
// The output is sorted for easier testings.
func (rs result) output(f io.Writer) error {
	keys := make([]string, 0, len(rs))
	for k := range rs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, err := fmt.Fprintf(f, "%s=%s\n", k, rs[k]); err != nil {
			return fmt.Errorf("failed to write output: %s", err)
		}
	}
	return nil
}
