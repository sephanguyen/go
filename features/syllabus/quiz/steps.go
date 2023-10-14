package quiz

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

type StepState struct {
	Token       string
	UserID      string
	Request     interface{}
	Response    interface{}
	ResponseErr error
	SessionID   string

	BookID               string
	ChapterID            string
	TopicID              string
	CourseID             string
	StudentIDs           []string
	CourseStudents       []*entities.CourseStudent
	StudyPlanID          string
	LearningMaterialID   string
	StudyPlanItemID      string
	StudyPlanItems       []*entities.StudyPlanItem
	StudyPlanItemsIDs    []string
	LoIDs                []string
	AssignedStudentIDs   []string
	ShuffledQuizSetID    string
	ExternalIDs          []string
	TotalPoint           int32
	LearningMaterialType sspb.LearningMaterialType
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<quiz>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<quiz>returns "([^"]*)" status code$`: s.returnsStatusCode,

		`^<quiz>user creates a valid book content$`:                      s.userCreatesAValidBookContent,
		`^<quiz>user creates a course and add students into the course$`: s.userCreatesACourseAndAddStudentsIntoTheCourse,
		`^<quiz>user adds a master study plan with the created book$`:    s.userAddsAMasterStudyPlanWithTheCreatedBook,

		`^<quiz>user create a learning material in "([^"]*)" type$`: s.userCreateALearningMaterialInType,
		`^user creates a quiz in "([^"]*)" type$`:                   s.userCreatesAQuizInType,
		`^user updates study plan for the learning material$`:       s.userUpdatesStudyPlanForTheLearningMaterial,
		`^user starts and submits a "([^"]*)" answer in "([^"]*)"$`: s.userStartsAndSubmitsAAnswerInKind,
		`^our system returns "([^"]*)" and "([^"]*)" correctly$`:    s.ourSystemReturnsCorrectnessAndIscorrectallCorrectly,
		`^user retry and submits a "([^"]*)" answer in "([^"]*)"$`:  s.userRetryAndSubmitsAAnswerInKind,
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
	stepState.Token = authToken
	stepState.UserID = userID
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsStatusCode(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), utils.ValidateStatusCode(stepState.ResponseErr, arg)
}
