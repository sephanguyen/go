package stresstest

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/lessonmgmt"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func Test_API(t *testing.T) {
	st, err := NewStressTest(
		&common.Config{
			FirebaseAPIKey:     "",
			BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
			IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
		},
		0,
		"",
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

	suite := st.NewSuite()
	ctx = common.StepStateToContext(ctx, suite.lessonSuite.CommonSuite.StepState)
	err = suite.ASignedInAsSchoolAdmin(ctx)
	require.NoError(t, err)
	ids, err := suite.GetLocationsByHasura(ctx, jwt, nil)
	require.NoError(t, err)
	fmt.Println("all locations: ", ids)

	err = suite.lessonSuite.RetrieveLowestLevelLocations(ctx)
	require.NoError(t, err)
	suite.lessonSuite.CommonSuite.CenterIDs = suite.lessonSuite.CommonSuite.LowestLevelLocationIDs
	fmt.Println("Lowest Level Locations: ", suite.lessonSuite.CommonSuite.LowestLevelLocationIDs)

	_, err = suite.lessonSuite.RetrieveListLessonManagement(ctx, "LESSON_TIME_FUTURE", "10", lessonmgmt.NIL_VALUE)
	require.NoError(t, err)
	fmt.Println("RetrieveListLessonManagement: ", suite.lessonSuite.CommonSuite.StepState.Response)

	ids, err = suite.GetTeachersByHasura(ctx, jwt, suite.lessonSuite.CommonSuite.CurrentSchoolID)
	require.NoError(t, err)
	fmt.Println("teachers: ", ids)
	suite.lessonSuite.CommonSuite.StepState.TeacherIDs = ids

	_, err = suite.lessonSuite.UserRetrieveStudentSubscription(ctx, 10, 0, "", "", "")
	require.NoError(t, err)
	fmt.Println("student subscriptions: ", suite.lessonSuite.CommonSuite.StepState.Response)
	subs := suite.lessonSuite.CommonSuite.Response.(*bpb.RetrieveStudentSubscriptionResponse)

	for _, item := range subs.Items {
		suite.lessonSuite.CommonSuite.StudentIDWithCourseID = append(
			suite.lessonSuite.CommonSuite.StudentIDWithCourseID,
			item.StudentId,
			item.CourseId,
		)
	}

	_, err = suite.lessonSuite.CommonSuite.UserCreateALessonWithMissingFields(ctx, "materials")
	require.NoError(t, err)
	fmt.Println("UserCreateALessonWithMissingFields: ", suite.lessonSuite.CommonSuite.StepState.ResponseErr)
	fmt.Println("UserCreateALessonWithMissingFields: ", suite.lessonSuite.CommonSuite.StepState.Response)

	stepState := common.StepStateFromContext(ctx)
	lesson, err := suite.GetLessonDetailByHasura(ctx, jwt, stepState.CurrentLessonID)
	require.NoError(t, err)
	fmt.Println("lesson:", lesson)
	err = suite.CheckCreatedLessonDetail(ctx, stepState.Request.(*bpb.CreateLessonRequest), lesson)
	require.NoError(t, err)
}

func Test_SchoolAdminCanCreateLessonWithAllRequiredFields(t *testing.T) {
	st, err := NewStressTest(
		&common.Config{
			FirebaseAPIKey:     "",
			BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
			IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
		},
		0,
		"manabie-p7muf",
		"./accounts.json",
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	suite := st.NewSuite()
	err = suite.Scenario_SchoolAdminCanCreateLessonWithAllRequiredFields(ctx)
	require.NoError(t, err)
}

func Test_SimpleStressTest_SchoolAdminCanCreateLessonWithAllRequiredFields(t *testing.T) {
	st, err := NewStressTest(
		&common.Config{
			FirebaseAPIKey:     "",
			BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
			IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
		},
		0,
		"manabie-p7muf",
		"./accounts.json",
	)
	require.NoError(t, err)

	const num = 10

	type TestReport struct {
		TotalDuration time.Duration
		Success       int
		Avg           time.Duration
		Max           time.Duration
		Min           time.Duration
	}
	tr := &TestReport{}
	type ScenarioReport struct {
		StartAt  time.Time
		Duration time.Duration
		Error    error
	}
	reports := make([]*ScenarioReport, 0, num)
	reportChan := make(chan *ScenarioReport)
	var wg sync.WaitGroup
	wg.Add(num + 1)
	// store reports
	go func() {
		defer wg.Done()
		for {
			r := <-reportChan
			reports = append(reports, r)
			if len(reports) >= num {
				break
			}
		}
	}()

	// setup suites
	suites := make([]*Suite, 0, num)
	for i := 0; i < num; i++ {
		suites = append(suites, st.NewSuite())
	}

	// execute Scenarios of suites
	start := time.Now()
	for _, suite := range suites {
		go func(suite *Suite) {
			ctx := context.Background()
			defer wg.Done()
			r := &ScenarioReport{}
			r.StartAt = time.Now()
			err = suite.Scenario_SchoolAdminCanCreateLessonWithAllRequiredFields(ctx)
			r.Duration = time.Now().Sub(r.StartAt)
			if err != nil {
				r.Error = err
			}
			reportChan <- r
		}(suite)
	}
	wg.Wait()
	tr.TotalDuration = time.Now().Sub(start)
	fmt.Println("Total time:", tr.TotalDuration)

	for _, r := range reports {
		if r.Error != nil {
			fmt.Println("err:", r.Error)
			continue
		}
		tr.Success++
		tr.Avg += r.Duration
		if r.Duration > tr.Max {
			tr.Max = r.Duration
		}
	}
	tr.Min = tr.Max
	for _, r := range reports {
		if r.Error != nil {
			continue
		}
		if r.Duration < tr.Min {
			tr.Min = r.Duration
		}
	}

	fmt.Printf("Total success: %d/%d (%v%%) \n", tr.Success, num, (float64(tr.Success)/float64(num))*100)
	if tr.Success == 0 {
		return
	}
	tr.Avg = tr.Avg / time.Duration(tr.Success)
	fmt.Println("average: ", tr.Avg)
	fmt.Println("min: ", tr.Min)
	fmt.Println("max: ", tr.Max)

	const resolution = 5 // 20%
	durationResolution := (tr.Max - tr.Min) / resolution
	for i := 0; i < resolution; i++ {
		lowerLimit := tr.Min + durationResolution*time.Duration(i)
		upperLimit := lowerLimit + durationResolution
		count := 0
		for _, report := range reports {
			if report.Error != nil {
				continue
			}
			if i == resolution-1 {
				// get [lowerLimit,upperLimit]
				if report.Duration >= lowerLimit && report.Duration <= upperLimit {
					count++
				}
			} else {
				// get [lowerLimit,upperLimit)
				if report.Duration >= lowerLimit && report.Duration < upperLimit {
					count++
				}
			}
		}
		percent := (float64(count) / float64(num)) * 100
		fmt.Printf("There are %v%% (%d) scenarios completed in %v to %v \n", percent, count, lowerLimit, upperLimit)
	}
}

func Test_SimpleStressTest_ScenarioSimulateAStudyOnLiveLessonRoom(t *testing.T) {
	// some config to execute this test
	const courseID = ""
	const locationID = ""
	const schoolID = 0
	const numStudent = 10
	const numTeacher = 2
	const tenantID = ""
	cfg := &common.Config{
		FirebaseAPIKey:     "",
		BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
		IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
	}

	st, err := NewStressTest(
		cfg,
		schoolID,
		tenantID,
		"./accounts.json",
	)
	require.NoError(t, err)

	const num = 1

	type TestReport struct {
		TotalDuration time.Duration
		Success       int
		Avg           time.Duration
		Max           time.Duration
		Min           time.Duration
	}
	tr := &TestReport{}
	type ScenarioReport struct {
		StartAt  time.Time
		Duration time.Duration
		Error    error
	}
	reports := make([]*ScenarioReport, 0, num)
	reportChan := make(chan *ScenarioReport)
	var wg sync.WaitGroup

	// store reports
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			r := <-reportChan
			reports = append(reports, r)
			if len(reports) >= num {
				break
			}
		}
	}()

	// because we only test after user login and join lesson,
	// so preparing will do sequentially
	scns := make([]*ScenarioSimulateAStudyOnLiveLessonRoom, 0, num)
	fmt.Println("Preparing...")
	for i := 0; i < num; i++ {
		scn, err := NewScenarioSimulateAStudyOnLiveLessonRoom(
			st,
			courseID,
			locationID,
			numTeacher,
			numStudent,
		)
		if err != nil {
			require.NoError(t, fmt.Errorf("NewScenarioSimulateAStudyOnLiveLessonRoom: %w", err))
		}
		err = scn.Prepare(context.Background())
		if err != nil {
			require.NoError(t, fmt.Errorf("preparing: %w", err))
		}
		err = scn.EnterSiteAndPrepareJoinRoom(context.Background())
		if err != nil {
			require.NoError(t, fmt.Errorf("preparing: %w", err))
		}
		scns = append(scns, scn)
	}

	// execute Scenarios of suites
	start := time.Now()
	fmt.Println("begin running scenario")
	wg.Add(len(scns))
	for _, scn := range scns {
		go func(scn *ScenarioSimulateAStudyOnLiveLessonRoom) {
			r := &ScenarioReport{}
			r.StartAt = time.Now()
			var err error
			defer wg.Done()
			defer func() {
				r.Duration = time.Now().Sub(r.StartAt)
				if err != nil {
					r.Error = err
				}
				if scn.logs.NumberFetchFailedRoomState != 0 {
					fmt.Printf("NumberFetchFailedRoomState id %s : %d", scn.id, scn.logs.NumberFetchFailedRoomState)
				}
				reportChan <- r
			}()

			fmt.Println("Executing... ", scn.id)
			err = scn.Execute(context.Background())
			if err != nil {
				return
			}
		}(scn)
	}
	wg.Wait()
	tr.TotalDuration = time.Now().Sub(start)
	fmt.Println("Total time:", tr.TotalDuration)

	for _, r := range reports {
		if r.Error != nil {
			fmt.Println("err:", r.Error)
			continue
		}
		tr.Success++
		tr.Avg += r.Duration
		if r.Duration > tr.Max {
			tr.Max = r.Duration
		}
	}
	tr.Min = tr.Max
	for _, r := range reports {
		if r.Error != nil {
			continue
		}
		if r.Duration < tr.Min {
			tr.Min = r.Duration
		}
	}

	fmt.Printf("Total success: %d/%d (%v%%) \n", tr.Success, num, (float64(tr.Success)/float64(num))*100)
	if tr.Success == 0 {
		return
	}
	tr.Avg = tr.Avg / time.Duration(tr.Success)
	fmt.Println("average: ", tr.Avg)
	fmt.Println("min: ", tr.Min)
	fmt.Println("max: ", tr.Max)

	const resolution = 5 // 20%
	durationResolution := (tr.Max - tr.Min) / resolution
	for i := 0; i < resolution; i++ {
		lowerLimit := tr.Min + durationResolution*time.Duration(i)
		upperLimit := lowerLimit + durationResolution
		count := 0
		for _, report := range reports {
			if report.Error != nil {
				continue
			}
			if i == resolution-1 {
				// get [lowerLimit,upperLimit]
				if report.Duration >= lowerLimit && report.Duration <= upperLimit {
					count++
				}
			} else {
				// get [lowerLimit,upperLimit)
				if report.Duration >= lowerLimit && report.Duration < upperLimit {
					count++
				}
			}
		}
		percent := (float64(count) / float64(num)) * 100
		fmt.Printf("There are %v%% (%d) scenarios completed in %v to %v \n", percent, count, lowerLimit, upperLimit)
	}
}
