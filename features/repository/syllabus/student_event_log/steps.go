package student_event_log

import (
	"github.com/jackc/pgtype"
)

type StepState struct {
	EventID            string
	StudyPlanID        string
	MasterStudyPlanID  string
	StudyPlanItemID    string
	StudentID          string
	LearningMaterialID string
	CompletedDate      pgtype.Timestamptz
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<student_event_logs>valid study plan item in DB$`: s.validStudyPlanItemInDB,
		`^insert a valid student event log$`:                s.insertAValidStudentEventLog,
		`^student event log new identity filled$`:           s.studentEventLogNewIdentityFilled,

		// check student completion learning material
		`^a valid "([^"]*)" student event log in DB at "([^"]*)"$`:          s.validStudentEventLogInDB,
		`^a valid "([^"]*)" student event log previous at "([^"]*)"$`:       s.validUpdateEventStudentEventLogInDB,
		`^<student_event_log>checking completed time of learning material$`: s.getCompletedTimeOfLM,
		`^<student_event_log>a valid "([^"]*)" of learning material$`:       s.validCompletedTime,
		`^<student_event_log>a valid exam lo in DB$`:                        s.validExamLOInDB,
		`^<student_event_log>a student "([^"]*)" exam lo$`:                  s.aStudentSubmitExamLO,
	}
	return steps
}
