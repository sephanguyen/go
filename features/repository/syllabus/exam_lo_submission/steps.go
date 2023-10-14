package exam_lo_submission

import "github.com/manabie-com/backend/internal/eureka/repositories"

type StepState struct {
	Request                     interface{}
	Response                    interface{}
	ResponseError               error
	SubmissionIDs               []string
	ExtendedExamLOSubmissionMap map[string]*repositories.ExtendedExamLOSubmission
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^user create a set of exam lo submissions to database$`: s.userCreateASetOfExamLOSubmissionsToDatabase,
		`^user call function exam lo submission list$`:           s.userCallFunctionExamLOSubmissionList,
		`^system returns correct list exam lo submissions$`:      s.systemReturnsCorrectListExamLOSubmissions,
	}
	return steps
}
