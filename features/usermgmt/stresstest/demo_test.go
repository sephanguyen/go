package stresstest

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/manabie-com/backend/features/common"

	"github.com/stretchr/testify/require"
)

func Test_API(t *testing.T) {
	st, err := NewStressTest(
		&common.Config{
			FirebaseAPIKey:     "",
			BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
			IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
		},
		0,
		"./accounts.json",
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := st.UserSignInWithPassword(ctx, &AccountInfo{
		Email:    "example@manabie.com",
		Password: "example_password",
	}, true)
	require.NoError(t, err)
	fmt.Println("res: ", res)

	jwt, err := st.ExchangeUserToken(ctx, res.IdToken)
	require.NoError(t, err)
	fmt.Println("jwt: ", jwt)

}

type ScenarioReport struct {
	StartAt  time.Time
	Duration time.Duration
	Error    error
}

type ScenarioReports []*ScenarioReport

func (scenarioReports ScenarioReports) GetSuccessfulScenarioReports() ScenarioReports {
	successScenarioReports := make(ScenarioReports, 0, len(scenarioReports))
	for _, scenarioReport := range scenarioReports {
		if scenarioReport.Error != nil {
			continue
		}
		successScenarioReports = append(successScenarioReports, scenarioReport)
	}
	return successScenarioReports
}

func (scenarioReports ScenarioReports) MinMaxAvg() (time.Duration, time.Duration, time.Duration) {
	min := scenarioReports[0].Duration
	max := min
	sum := time.Duration(0)
	for _, scenarioReport := range scenarioReports {
		if scenarioReport.Duration < min {
			min = scenarioReport.Duration
		}
		if scenarioReport.Duration > max {
			max = scenarioReport.Duration
		}
		sum += scenarioReport.Duration
	}
	return min, max, sum / time.Duration(len(scenarioReports))
}

// RunScenarioExchangeTokenWithValidAuthenticationToken runs ExchangeTokenWithValidAuthenticationToken scenario
// and returns ScenarioReport
func (s *Suite) RunScenarioExchangeTokenWithValidAuthenticationToken(ctx context.Context) *ScenarioReport {
	scenarioReport := &ScenarioReport{}
	scenarioReport.StartAt = time.Now()
	err := s.Scenario_ExchangeTokenWithValidAuthenticationToken(ctx)
	scenarioReport.Duration = time.Now().Sub(scenarioReport.StartAt)
	if err != nil {
		scenarioReport.Error = err
	}
	return scenarioReport
}

// Config for staging:
//	BobSrvAddr:         "api.staging.manabie.io:31500",
//	FirebaseAPIKey:     "AIzaSyA7h5F1D1irKjtxd5Uj8A1OTMRmoc1ANRs",
//	BobHasuraAdminURL:  "https://admin.staging-green.manabie.io:31600",
//	IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",

// Accounts.json can use that account:  thu.vo+jprep@manabie.com / M@nabie123

func Test_SimpleStressTest_ExchangeTokenWithValidAuthenticationToken(t *testing.T) {
	ctx := context.Background()

	st, err := NewStressTest(
		&common.Config{
			FirebaseAPIKey:     "",
			BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
			IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
		},
		-2147483647,
		"./accounts.json",
	)
	require.NoError(t, err)

	const num = 500

	type TestReport struct {
		TotalDuration time.Duration
		Success       int
		Avg           time.Duration
		Max           time.Duration
		Min           time.Duration
	}
	tr := &TestReport{}

	// setup suites
	suites := make([]*Suite, 0, num)
	for i := 0; i < num; i++ {
		suites = append(suites, st.NewSuite())
	}

	// execute Scenarios of suites in parallel
	start := time.Now()
	reportChan := make(chan *ScenarioReport, len(suites))

	//Run suite.RunScenarioExchangeTokenWithValidAuthenticationToken() in parallel
	var wg sync.WaitGroup
	for _, suite := range suites {
		wg.Add(1)
		go func(suite *Suite) {
			defer wg.Done()
			reportChan <- suite.RunScenarioExchangeTokenWithValidAuthenticationToken(ctx)
		}(suite)
	}
	wg.Wait()

	close(reportChan)
	tr.TotalDuration = time.Now().Sub(start)
	fmt.Println("Total time:", tr.TotalDuration)

	// store reports
	reports := make(ScenarioReports, 0, num)
	for r := range reportChan {
		reports = append(reports, r)
	}

	successfulScenarioReports := reports.GetSuccessfulScenarioReports()
	min, max, avg := successfulScenarioReports.MinMaxAvg()

	tr.Success = len(successfulScenarioReports)
	tr.Min = min
	tr.Max = max
	tr.Avg = avg

	fmt.Printf("Total success: %d/%d (%v%%) \n", tr.Success, num, (float64(tr.Success)/float64(num))*100)
	if tr.Success == 0 {
		return
	}

	fmt.Println("average: ", tr.Avg)
	fmt.Println("min: ", tr.Min)
	fmt.Println("max: ", tr.Max)

	const resolution = 5 // 20%
	durationResolution := (tr.Max - tr.Min) / resolution
	for i := 0; i < resolution; i++ {
		lowerLimit := tr.Min + durationResolution*time.Duration(i)
		upperLimit := lowerLimit + durationResolution

		//avoid incorrect rounding number when divide floating number
		if i == resolution-1 {
			upperLimit = tr.Max
		}

		count := 0
		for _, report := range successfulScenarioReports {
			if report.Duration >= lowerLimit && report.Duration <= upperLimit {
				count++
			}
		}
		percent := (float64(count) / float64(num)) * 100
		fmt.Printf("There are %v%% (%d) scenarios completed in %v to %v \n", percent, count, lowerLimit, upperLimit)
	}
}
