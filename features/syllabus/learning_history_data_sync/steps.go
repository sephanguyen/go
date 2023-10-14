// nolint
package learning_history_data_sync

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
)

type StepState struct {
	Response    interface{}
	Request     interface{}
	ResponseErr error
	BookID      string
	TopicIDs    []string
	ChapterIDs  []string
	Token       string
	SchoolAdmin entity.SchoolAdmin
	Student     entity.Student
	Teacher     entity.Teacher
	Parent      entity.Parent
	HQStaff     entity.HQStaff

	NumberOfStudyPlan   int
	LearningMaterialIDs []string
	SubmissionIDs       []string
	CompletedLmIDs      []string
	CourseID            string
	StudentID           string
	QuestionTagIDs      []string
	ExamLOID            string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<learning_history_data_sync>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<learning_history_data_sync>returns "([^"]*)" status code$`: s.returnsStatusCode,
		`^user download mapping file$`:                                s.userDownloadMappingFile,
		`^returns url of mapping file correctly$`:                     s.returnURLCorrectly,

		`^<learning_history_data_sync>valid course, exam_lo, question_tag in db$`: s.validCourseExamLoQuestionTagInDB,
		`^user upload mapping file$`: s.userUploadMappingFile,
		`^csv file is uploaded$`:     s.svFileIsUploaded,
	}

	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	userID, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, arg)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	//TODO: no need if you're not use it. Just an example.
	switch arg {
	case "student":
		stepState.Student.Token = authToken
		stepState.Student.ID = userID
	case "school admin", "admin":
		stepState.SchoolAdmin.Token = authToken
		stepState.SchoolAdmin.ID = userID
	case "teacher", "current teacher":
		stepState.Teacher.Token = authToken
		stepState.Teacher.ID = userID
	case "parent":
		stepState.Parent.Token = authToken
		stepState.Parent.ID = userID
	case "hq staff":
		stepState.HQStaff.Token = authToken
		stepState.HQStaff.ID = userID
	default:
		stepState.Student.Token = authToken
		stepState.Student.ID = userID
	}
	stepState.Token = authToken
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsStatusCode(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), utils.ValidateStatusCode(stepState.ResponseErr, arg)
}
