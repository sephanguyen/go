package question_tag

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
)

type StepState struct {
	Token       string
	Response    interface{}
	Request     interface{}
	ResponseErr error
	SchoolAdmin entity.SchoolAdmin
	Student     entity.Student
	Teacher     entity.Teacher
	Parent      entity.Parent
	HQStaff     entity.HQStaff

	QuestionTagIDs     []string
	QuestionTagTypeIDs []string
	QuestionTagNames   []string
	CSVContent         []byte
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<question_tag>a signed in "([^"]*)"`:                s.aSignedIn,
		`^some question tag types existed in database$`:       s.someQuestionTagTypesExistedInDatabase,
		`^a valid csv content with some valid question tags$`: s.aValidCSVContentWithSomeValidQuestionTags,
		`^<question_tag>returns "([^"]*)" status code$`:       s.returnsStatusCode,
		`^user upsert question tag$`:                          s.userUpsertQuestionTag,
		`^user create question tag$`:                          s.userCreateQuestionTag,
		`^question tag must be created$`:                      s.questionTagMustBeCreated,
		`^user update question tag$`:                          s.userUpdateQuestionTag,
		`^question tag must be updated$`:                      s.questionTagMustBeUpdated,
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
		stepState.Parent.Token = authToken
		stepState.Parent.ID = userID
	}
	stepState.Token = authToken
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsStatusCode(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), utils.ValidateStatusCode(stepState.ResponseErr, arg)
}
