package task_assignment

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
)

type StepState struct {
	UserID                     string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	BookID                     string
	TopicIDs                   []string
	ChapterIDs                 []string
	CourseID                   string
	Token                      string
	SchoolAdmin                entity.SchoolAdmin
	Student                    entity.Student
	Teacher                    entity.Teacher
	Parent                     entity.Parent
	HQStaff                    entity.HQStaff
	TopicID                    string
	LearningMaterialID         string
	LearningMaterialIDs        []string
	TopicLODisplayOrderCounter int32
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<task_assignment>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<task_assignment>returns "([^"]*)" status code$`: s.returnsStatusCode,
		`^<task_assignment>a valid book content$`:          s.aValidBookContent,
		`^<task_assignment>a valid course$`:                s.aValidCourse,
		`^there are task assignments existed in topic$`:    s.thereAreTaskAssignmentsExistedInTopic,

		// insert task assignment
		`^user insert a valid task assignment`:                                               s.userInsertAValidTaskAssignment,
		`^task assignment must be created`:                                                   s.taskAssignmentMustBeCreated,
		`^our system generates a correct display order for task assignment`:                  s.ourSystemGeneratesACorrectDisplayOrderForTaskAssignment,
		`^our system updates topic LODisplayOrderCounter correctly with new task assignment`: s.ourSystemUpdatesTopicLODisplayOrderCounterCorrectlyWithNewTaskAssignment,

		// list task assignment
		`^user list task assignment$`:                        s.userListTaskAssignment,
		`^our system must return task assignment correctly$`: s.ourSystemMustReturnTaskAssignmentCorrectly,

		`^user update valid task assignment$`:                s.userUpdateValidTaskAssignment,
		`^our system updates the task assignment correctly$`: s.ourSystemUpdatesTheTaskAssignmentCorrectly,

		// upsert adhoc task assignment
		`^user creates a valid adhoc task assignment$`:         s.userCreatesAValidAdhocTaskAssignment,
		`^user updates the adhoc task assignment$`:             s.userUpdatesTheAdhocTaskAssignment,
		`^our system creates adhoc task assignment correctly$`: s.ourSystemCreatesAdhocTaskAssignmentCorrectly,
		`^our system updates adhoc task assignment correctly$`: s.ourSystemUpdatesAdhocTaskAssignmentCorrectly,
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

func (s *Suite) aValidBookContent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookID, chapterIDs, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, s.EurekaDB, constant.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}
	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.TopicIDs = topicIDs
	stepState.TopicID = topicIDs[0]
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourse: %w", err)
	}
	stepState.CourseID = courseID

	return utils.StepStateToContext(ctx, stepState), nil
}
