package student_latest_submissions

type StepState struct {
	StudyPlanID       string
	MasterStudyPlanID string
	StudyPlanItemID   string
	AssignmentID      string
	Name              string
	StudentID         string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<student_latest_submissions>valid study plan item in DB$`: s.validStudyPlanItemInDB,
		`^insert a valid student latest submission$`:                s.insertAValidStudentLatestSubmission,
		`^student latest submission new identity filled$`:           s.studentLatestSubmissionNewIdentityFilled,
	}
	return steps
}
