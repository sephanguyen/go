package statuscheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/manabie-com/backend/internal/golibs/logger"

	"go.uber.org/multierr"
)

type Checker struct {
	Data         string
	RequiredJobs []string
}

func (c *Checker) Run() error {
	d := ghData{}
	if err := json.Unmarshal([]byte(c.Data), &d); err != nil {
		return fmt.Errorf("invalid input data: %s", err)
	}

	expectedJobOutput, err := d.getExpectedResults(c.RequiredJobs)
	if err != nil {
		return err
	}

	return d.verify(*expectedJobOutput)
}

type ghData map[string]ghJobOutput

var errMissingActivation = errors.New("missing \"requirements\" field in input data")

func (d ghData) getExpectedResults(requiredJobs []string) (*ghData, error) {
	res := ghData{}
	activation, exists := d["requirements"]
	if !exists {
		return nil, errMissingActivation
	}

	for _, jobName := range requiredJobs {
		flagVal, exists := activation.Outputs[jobName]
		if !exists {
			return nil, fmt.Errorf("missing required activation flag %q in activation job", jobName)
		}
		if flagVal == "1" {
			res[jobName] = ghJobOutput{Result: statusSuccess}
		} else {
			res[jobName] = ghJobOutput{Result: statusSkipped}
		}
	}
	return &res, nil
}

// verify verifies that the actual job outcomes (contained in d) match the expected
// outcomes (contained in expected).
//
// For some cases, if the job's result is "cancelled" vs the expected "success",
// it is likely because the workflow itself cancelled that job to save resources.
// This happens when any job fails, so we must look for it first.
func (d ghData) verify(expected ghData) error {
	var p1Errs error // stores the root errors that dev must fix
	var p2Errs error // stores the errors that might come from another error

	// make the output deterministic, for testing
	sortedJobNames := d.sortJobName(expected)
	for _, jobName := range sortedJobNames {
		expectedJobVal := expected[jobName]
		actualJobVal, exists := d[jobName]
		if !exists {
			err := fmt.Errorf("missing required job %q from input data", jobName)
			logger.Errorf("verify error: %s", err)
			p1Errs = multierr.Append(p1Errs, err)
			continue
		}
		if actualJobVal.Result != expectedJobVal.Result {
			err := fmt.Errorf("expected output %q for job %q, got %q", expectedJobVal.Result, jobName, actualJobVal.Result)
			if actualJobVal.Result == statusCancelled && expectedJobVal.Result == statusSuccess {
				// likely the case where the workflow cancel this job to save resources
				logger.Warnf("verify error, cancelled case: %s", err)
				p2Errs = multierr.Append(p2Errs, err)
			} else {
				logger.Errorf("verify error: %s", err)
				p1Errs = multierr.Append(p1Errs, err)
			}
		}
	}

	// Only return the low priority errors when there are no high priority errors
	if p1Errs != nil {
		return p1Errs
	}
	return p2Errs
}

func (d ghData) sortJobName(data ghData) []string {
	res := make([]string, 0, len(data))
	for name := range data {
		res = append(res, name)
	}
	sort.Strings(res)
	return res
}

type ghJobOutput struct {
	Result  githubJobStatus   `json:"result"`
	Outputs map[string]string `json:"outputs"`
}
