package diff

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/logger"
)

type rule struct {
	// Name is the name of the rule.
	Name string `yaml:"name"`

	// Paths contains the list of regexp paths. When there are files changed on
	// one of those paths, this rule is activated.
	Paths []string `yaml:"paths"`

	// PathsIgnore contains the list of regexp paths. Files in any of these paths
	// never trigger this rule, effectively ignoring Paths if matched.
	PathsIgnore []string `yaml:"paths-ignore"`

	// RunOnly controls whether all changed files should match paths to activate the rule.
	// When false, this rule is activated when any of the changed files matches.
	// When true, all files must match to activate this rule.
	RunOnly bool `default:"false" yaml:"run_only"`

	// Values contains the list of values to replace {{.VALUE}} in rule's Paths.
	// When the rule with a value is activated, that value will be appended
	// to the output of this rule (and thus, this rule's output will not be a boolean
	// but will be an array instead).
	Values []string `default:"[]" yaml:"values"`

	// ForceValue is the value of this rule in --force mode.
	ForceValue *string `yaml:"force_value"`

	// EnabledSquads contains a list of squad names that this rule is enabled for.
	// Affects only `git diff` method. Mutually exclusive with DisabledSquads.
	// When the target squad is not specified, or both EnabledSquads and DisabledSquads are empty,
	// then the rule is enabled by default.
	EnabledSquads []string `yaml:"enabled_squads"`

	// DisabledSquads contains a list of squad names that this rule is disabled for.
	// Affects only `git diff` method. Mutually exclusive with EnabledSquads.
	// When the target squad is not specified, or both EnabledSquads and DisabledSquads are empty,
	// then the rule is enabled by default.
	DisabledSquads []string `yaml:"disabled_squads"`
}

func (r rule) parseTemplatedRule(val string) *rule {
	child := rule{
		Name:       r.Name,
		Paths:      make([]string, 0, len(r.Paths)),
		RunOnly:    r.RunOnly,
		Values:     nil,
		ForceValue: r.ForceValue,
	}
	for i := range r.Paths {
		child.Paths = append(child.Paths, strings.ReplaceAll(r.Paths[i], "{{.VALUE}}", val))
	}
	return &child
}

func (r *rule) match(changedFiles []string, squad *string) (bool, []string, error) {
	if len(r.Values) == 0 {
		matched, err := r.matchWithoutValues(changedFiles, squad)
		return matched, nil, err
	}
	matchedValues := []string{}
	for _, val := range r.Values {
		childRule := r.parseTemplatedRule(val)
		matched, err := childRule.matchWithoutValues(changedFiles, squad)
		if err != nil {
			return false, nil, err
		}
		if matched {
			matchedValues = append(matchedValues, val)
		}
	}
	return len(matchedValues) > 0, matchedValues, nil
}

func (r *rule) matchWithoutValues(changedFiles []string, squad *string) (bool, error) {
	// early exit if this rule is disabled for this squad
	if !r.matchSquad(squad) {
		logger.Debugf("rule %q is will NOT be triggered (disabled for squad %s)", r.Name, *squad)
		return false, nil
	}

	// match the paths-ignore first
	allRegexpsIgnore := make([]*regexp.Regexp, 0, len(r.PathsIgnore))
	for _, p := range r.PathsIgnore {
		re, err := regexp.Compile(p)
		if err != nil {
			return false, fmt.Errorf("failed to compile regexp %q: %s", p, err)
		}
		allRegexpsIgnore = append(allRegexpsIgnore, re)
	}

	filteredChangedFiles := make([]string, 0, len(changedFiles))
	for _, f := range changedFiles {
		ignoreMatched := false
		for i := range allRegexpsIgnore {
			matched := allRegexpsIgnore[i].MatchString(f)
			if !matched { // if not ignored, add to list to process later
				continue
			}
			logger.Debugf("ignored: rule %q ignores file %q by path %q", r.Name, f, r.PathsIgnore[i])
			if r.RunOnly {
				logger.Infof("result: rule %q will NOT be trigger", r.Name)
				return false, nil
			}
			ignoreMatched = true
			break
		}

		if !ignoreMatched {
			filteredChangedFiles = append(filteredChangedFiles, f)
		}
	}

	// match the paths from the filtered file list
	// again, RunOnly determines match-all or match-any strategy
	allRegexps := make([]*regexp.Regexp, 0, len(r.Paths))
	for _, p := range r.Paths {
		re, err := regexp.Compile(p)
		if err != nil {
			return false, fmt.Errorf("failed to compile regexp %q: %s", p, err)
		}
		allRegexps = append(allRegexps, re)
	}

	if !r.RunOnly {
		for _, f := range filteredChangedFiles {
			for i := range allRegexps {
				matched := allRegexps[i].MatchString(f)
				if matched {
					logger.Debugf("match: rule %q matched file %q by path %q", r.Name, f, r.Paths[i])
					logger.Debugf("result: rule %q will be triggered", r.Name)
					return true, nil
				}
			}
		}
		logger.Debugf("match: rule %q failed to match any file", r.Name)
		logger.Debugf("result: rule %q will NOT be triggered", r.Name)
		return false, nil
	}

	for _, f := range filteredChangedFiles {
		matchedAtLeastOnce := false
		for _, re := range allRegexps {
			matched := re.MatchString(f)
			matchedAtLeastOnce = matchedAtLeastOnce || matched
			if matchedAtLeastOnce {
				break
			}
		}
		if !matchedAtLeastOnce {
			logger.Debugf("match: rule %q failed to match file %q by any of its paths", r.Name, f)
			logger.Debugf("result: rule %q will NOT be triggered", r.Name)
			return false, nil
		}
	}
	logger.Debugf("result: rule %q will be triggered", r.Name)
	return true, nil
}

// matchSquad determines whether this rule is enabled for squad.
func (r *rule) matchSquad(squad *string) bool {
	if squad == nil {
		return true
	}
	if len(r.EnabledSquads) > 0 {
		for _, v := range r.EnabledSquads {
			if v == *squad {
				return true
			}
		}
		return false
	}
	if len(r.DisabledSquads) > 0 {
		for _, v := range r.DisabledSquads {
			if v == *squad {
				return false
			}
		}
		return true
	}
	return true
}

// ruleList is contains a list of ruleList to trigger requirements
// base on a list of changed files.
type ruleList struct {
	Rules []rule `yaml:"rules"`
}

func (r ruleList) resultFromFileChanges(changedFiles []string, squad *string) (*result, error) {
	res := r.generateDisableAllResult()
	for _, rule := range r.Rules {
		matched, matchedValues, err := rule.match(changedFiles, squad)
		if err != nil {
			return nil, err
		}
		if matched {
			if len(matchedValues) > 0 {
				res[rule.Name] = res[rule.Name] + strings.Join(matchedValues, " ") + " "
			} else {
				res[rule.Name] = "1"
			}
		}
	}
	logger.Debugf("requirement result from git diff: %+v", res)
	return &res, nil
}

func (r ruleList) resultFromPRDesc2(prDesc string) (*result, error) {
	logger.Debugf("received PR description: %q", prDesc)

	res := r.generateDisableAllResult()
	// Note that Github uses CLRF, see https://github.com/actions/runner/issues/1462#issuecomment-1030124116
	regexpFormat := `(?m:^-\s*test:(%s)(?::([\w\/\.]+))?(?:\r?\n)?)`
	for _, rule := range r.Rules {
		ruleRegexp := fmt.Sprintf(regexpFormat, rule.Name)
		triggerRe, err := regexp.Compile(ruleRegexp)
		if err != nil {
			logger.Errorf("regexp.Compile failed: %s (regexp string: %s)", err, triggerRe)
			return nil, err
		}

		matches := triggerRe.FindAllStringSubmatch(prDesc, -1)
		res.addFromRegexpmatches(matches)
	}

	logger.Debugf("requirement result from PR description: %+v", res)
	return &res, nil
}

// forceValues returns a DiffResult when "force-test" is enabled.
func (r ruleList) forceValues() result {
	res := make(result)
	for _, v := range r.Rules {
		switch {
		case v.ForceValue != nil:
			res[v.Name] = *v.ForceValue
		case v.RunOnly:
			res[v.Name] = "0"
		case len(v.Values) > 0:
			arr := []string{}
			if _, ok := res[v.Name]; ok {
				arr = strings.Split(res[v.Name], " ")
			}
			r.mergeArray(&arr, &v.Values)
			res[v.Name] = strings.Join(arr, " ")
		default:
			res[v.Name] = "1"
		}
	}
	logger.Debugf("requirement result from --force: %+v", res)
	return res
}

func (r ruleList) mergeArray(dst, other *[]string) {
	for _, v := range *other {
		if !r.exists(*dst, v) {
			*dst = append(*dst, v)
		}
	}
}

func (r ruleList) exists(s []string, v string) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}

func (r ruleList) generateDisableAllResult() result {
	res := make(result)
	for _, v := range r.Rules {
		if len(v.Values) > 0 {
			res[v.Name] = ""
		} else {
			res[v.Name] = "0"
		}
	}
	// some custom vals
	res["svcs_change"] = ""
	res["run_all_integration_test"] = "0"
	return res
}
