package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"go.uber.org/multierr"

	"github.com/stretchr/testify/assert"
)

// the test case very hard to set up auto
func Test_Calculate(t *testing.T) {
	t.Parallel()

	reproduce2Start1CompletedV1Logs, err := reproduce2Start1CompletedV1()
	assert.NoError(t, err, "setup failed")
	reproduce2Start1CompletedV2Logs, err := reproduce2Start1CompletedV2()
	assert.NoError(t, err, "setup failed")
	happyCaseLogs, err := happyCase()
	assert.NoError(t, err, "setup failed")
	reproduce1Start2CompletedLogs, err := reproduce1Start2Completed()
	assert.NoError(t, err, "setup failed")
	reproduce1Start2CompletedAlotResumedLogs, err := reproduce1Start2CompletedAlotResumed()
	assert.NoError(t, err, "setup failed")
	reproduce2Start2Completed2ExitLogs, err := reproduce2Start2Completed2Exit()
	assert.NoError(t, err, "setup failed")
	reproduce2Start2CompletedLogs, err := reproduce2Start2Completed()
	assert.NoError(t, err, "setup failed")
	reproduce1Start1CompletedAlotBetweenLogs, err := reproduce1Start1CompletedAlotBetween()
	assert.NoError(t, err, "setup failed")
	reproduce1Start1Exited1CompletedLogs, err := reproduce1Start1Exited1Completed()
	assert.NoError(t, err, "setup failed")
	reproduce1Start1Exited1CompletedSametimeLogs, err := reproduce1Start1Exited1CompletedSametime()
	assert.NoError(t, err, "setup failed")

	testcases := []TestCase{
		{
			name:         "with 2 start 1 completed V1",
			ctx:          nil,
			req:          reproduce2Start1CompletedV1Logs,
			expectedResp: time.Duration(time.Minute * 20),
		},
		{
			name:         "with 2 start 1 completed V2",
			ctx:          nil,
			req:          reproduce2Start1CompletedV2Logs,
			expectedResp: time.Duration(time.Minute * 20),
		},
		{

			name:         "with 1 start 2 completed",
			ctx:          nil,
			req:          reproduce1Start2CompletedLogs,
			expectedResp: time.Duration(time.Minute * 30),
		},
		{
			name:         "happy case ",
			ctx:          nil,
			req:          happyCaseLogs,
			expectedResp: time.Duration(time.Minute*250 - time.Minute*120),
		},
		{
			name:         "with 1 start 2 completed + some resumed after ",
			ctx:          nil,
			req:          reproduce1Start2CompletedAlotResumedLogs,
			expectedResp: time.Duration(time.Minute * 30),
		},
		{
			name:         "with 2 start 2 completed 2 exit ",
			ctx:          nil,
			req:          reproduce2Start2Completed2ExitLogs,
			expectedResp: time.Duration(time.Minute * 30),
		},

		{
			name:         "with 2 start 2 completed/happy case two flow in 1 session id ",
			ctx:          nil,
			req:          reproduce2Start2CompletedLogs,
			expectedResp: time.Duration(time.Minute * 35),
		},
		{
			name:         "with 1 start 1 completed/happy case with a lot of paused in between started and completed",
			ctx:          nil,
			req:          reproduce1Start1CompletedAlotBetweenLogs,
			expectedResp: time.Duration(time.Minute * 50),
		},
		{
			name:         "with 1 start exited 1 completed",
			ctx:          nil,
			req:          reproduce1Start1Exited1CompletedLogs,
			expectedResp: time.Duration(time.Minute * 50),
		},
		{
			//reproduce1Start1Exited1CompletedSametimeLogs
			name:         "with 1 start exited 1 completed in same time",
			ctx:          nil,
			req:          reproduce1Start1Exited1CompletedSametimeLogs,
			expectedResp: time.Duration(time.Minute * 10),
		},
	}
	learningTimeCalculator := LearningTimeCalculator{}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			req := testcase.req.([]*entities.StudentEventLog)
			learningTime, _, err := learningTimeCalculator.Calculate(req)
			assert.NoError(t, err)
			assert.Equal(t, testcase.expectedResp, learningTime, "learning time have equal expected")
		})

	}
}

func genStudentEvenLog(id, timeSpent int, studentID, loID, sessionID, evtType, event, studyPlanItemID, createdAt string) (*entities.StudentEventLog, error) {
	layout := "2006-01-02 15:04:05.000000 +00:00"
	layout2 := "2006-01-02 15:04:05.000000"
	studentEvtLog := &entities.StudentEventLog{}
	database.AllNullEntity(studentEvtLog)
	studentEvtLog.EventID = database.Varchar(idutil.ULIDNow())
	payload := &GenericPayload{
		Event:           event,
		LoID:            loID,
		SessionID:       sessionID,
		StudyPlanItemID: studyPlanItemID,
		TimeSpent:       timeSpent,
	}
	studentEvtLog.Payload.Set(payload)

	err := multierr.Combine(studentEvtLog.CreatedAt.Set(timeutil.Now()),
		studentEvtLog.EventType.Set(evtType),
		studentEvtLog.StudentID.Set(studentID),
		studentEvtLog.ID.Set(id),
		studentEvtLog.CreatedAt.Set(timeutil.Now()))
	if err != nil {
		return nil, fmt.Errorf("unable to set value student event: %w", err)
	}
	if createdAt != "" {
		parsedTime, err := time.Parse(layout, createdAt)
		if err != nil {
			parsedTime, err = time.Parse(layout2, createdAt)
			if err != nil {
				return nil, fmt.Errorf("unable to parse time: %w", err)
			}
		}
		studentEvtLog.CreatedAt.Set(parsedTime)
	}
	return studentEvtLog, nil
}

type GenericPayload struct {
	Event           string `json:"event,omitempty"`
	LoID            string `json:"lo_id,omitempty"`
	SessionID       string `json:"session_id,omitempty"`
	StudyPlanItemID string `json:"study_plan_item_id,omitempty"`
	TimeSpent       int    `json:"time_spent,omitempty"`
}

// with a lot of pause and resume
func happyCase() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	index := 153428
	layout := "2006-01-02 15:04:05.000000 +00:00"
	createdAtStr := "2021-12-24 09:00:00.559322 +00:00"
	parsedTime, err := time.Parse(layout, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse time: %w", err)
	}

	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", createdAtStr)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	for i := 1; i < 50; i++ {
		parsedTime = parsedTime.Add(time.Minute * 5)
		strTimeArr := strings.Split(parsedTime.String(), " ")
		parsedTimeStr := strTimeArr[0] + " " + strTimeArr[1]
		switch i % 2 {
		case 1:
			index += i
			e, err := genStudentEvenLog(index, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", parsedTimeStr)
			if err != nil {
				return nil, err
			}
			studentEvtLogs = append(studentEvtLogs, e)

		case 0:
			index += i
			e, err := genStudentEvenLog(index, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", parsedTimeStr)
			if err != nil {
				return nil, err
			}
			studentEvtLogs = append(studentEvtLogs, e)
		}
	}
	parsedTime = parsedTime.Add(time.Minute * 5)
	strTimeArr := strings.Split(parsedTime.String(), " ")
	parsedTimeStr := strTimeArr[0] + " " + strTimeArr[1]
	e, err = genStudentEvenLog(index+1, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", parsedTimeStr)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	return studentEvtLogs, nil
}

// 2 start 2 completed happy
func reproduce2Start2Completed() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:20:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:30:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:40:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:45:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:50:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 07:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

// 2 start 2 completed happy
func reproduce2Start2Completed2Exit() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:20:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:30:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "exited", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:32:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:40:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:45:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:55:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 07:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "exited", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 07:02:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

// reproduce2Start1CompletedV1 start after completed one time with order start -> completed -> start
func reproduce2Start1CompletedV1() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:20:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:30:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:40:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:55:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

// reproduce2Start1CompletedV2 start after completed one time with order start -> start - > completed
func reproduce2Start1CompletedV2() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:20:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:40:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:55:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:30:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	return studentEvtLogs, nil
}

// reproduce1Start2Completed -> start -> completed -> completed
func reproduce1Start2Completed() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:20:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:30:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:35:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:45:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:50:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

// a lot resumed inherite reproduce1Start2Completed
func reproduce1Start2CompletedAlotResumed() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:20:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:30:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:35:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:45:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:50:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:55:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "resumed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:59:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

func reproduce1Start1CompletedAlotBetween() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:15:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:20:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:25:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:30:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "paused", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:35:00.000000 +00:00")
	if err != nil {
		return nil, err
	}

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:50:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

func reproduce1Start1Exited1Completed() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "exited", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:50:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

func reproduce1Start1Exited1CompletedSametime() ([]*entities.StudentEventLog, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	e, err := genStudentEvenLog(153428, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "started", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:00:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "exited", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, "student-id", "mock-lo-id", "mock-lo-id", "learning_objective", "completed", "01FBV29DGXB90MB26SH07V2W7W", "2021-12-20 06:10:00.000000 +00:00")
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	return studentEvtLogs, nil
}

// TODO:
func Test_CalculateLearningTimeByEventLogs(t *testing.T) {

}
