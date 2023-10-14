package student_submissions

import "github.com/jackc/pgtype"

type StepState struct {
	StudyPlanID       string
	MasterStudyPlanID string
	StudyPlanItemID   string
	AssignmentID      string
	Name              string
	StudentID         string
	CompletedDate     pgtype.Timestamptz
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<student_submissions>valid study plan item in DB$`: s.validStudyPlanItemInDB,
		`^insert a valid student submission$`:                s.insertAValidStudentSubmission,
		`^student submission new identity filled$`:           s.studentSubmissionNewIdentityFilled,

		// check student completion learning material
		`^a valid "([^"]*)" student submission in DB$`:                        s.validStudentSubmissionInDB,
		`^<student_submissions>checking completed time of learning material$`: s.getCompletedTimeOfLM,
		`^<student_submissions>a valid "([^"]*)" of learning material$`:       s.validCompletedTime,
	}
	return steps
}
