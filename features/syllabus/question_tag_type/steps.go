package question_tag_type

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
)

type StepState struct {
	Response           interface{}
	Request            interface{}
	ResponseErr        error
	BookID             string
	TopicIDs           []string
	ChapterIDs         []string
	Token              string
	SchoolAdmin        entity.SchoolAdmin
	Student            entity.Student
	Teacher            entity.Teacher
	Parent             entity.Parent
	HQStaff            entity.HQStaff
	CSVContent         []byte
	QuestionTagTypeIDs []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`<question_tag_type>a signed in "([^"]*)"$`:         s.aSignedIn,
		`<question_tag_type>returns "([^"]*)" status code$`: s.returnsStatusCode,
		`^user update existed question tag types$`:          s.userUpdateQuestionTagType,
		`^user insert some question tag types$`:             s.userUpsertQuestionTagType,
		`^question tag type must be created$`:               s.questionTagTypeMustBeCreated,
		`^question tag type must be updated$`:               s.questionTagTypeMustBeUpdated,
		`^a valid csv content with "([^"]*)"$`:              s.aValidCSVContentWith,
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
