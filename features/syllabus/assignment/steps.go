package assignment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type StepState struct {
	Token                      string
	AdminToken                 string
	StudentID                  string
	UserID                     string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	BookID                     string
	TopicID                    string
	TopicIDs                   []string
	ChapterID                  string
	ChapterIDs                 []string
	SchoolAdmin                entity.SchoolAdmin
	Student                    entity.Student
	LearningMaterialID         string
	LearningMaterialIDs        []string
	TopicLODisplayOrderCounter int32
	StudyPlanID                string
	AssignmentIDs              []string
	CourseID                   string
	HighestScore               int32
	StudyPlanItemIDs           []string
	SubmissionID               string
	// student submission test
	Submissions []*sspb.SubmitAssignmentRequest
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		// BEGIN====common=====BEGIN
		`^<assignment>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<assignment>a valid book content$`:          s.aValidBookContent,
		`^<assignment>a valid assignment$`:            s.aValidAssignment,
		`^<assignment>a valid task assignment$`:       s.aValidTaskAssignment,
		`^<assignment>returns "([^"]*)" status code$`: s.returnsStatusCode,

		// // insert assignment
		`^user inserts an assignment$`: s.userInsertAssignment,
		`^assignment must be created$`: s.assignmentMustBeCreated,

		// // update assignment:
		`^user updates an assignment$`: s.userUpdateAssignment,
		`^assignment must be updated$`: s.assignmentMustBeUpdated,

		// // list assignment:
		`^there are assignments existed$`:                s.thereAreAssignmentsExisted,
		`^user list assignment$`:                         s.userListAssignment,
		`^our system must return assignments correctly$`: s.ourSystemMustReturnAssignmentCorrectly,

		// submit assignment
		`^a course and study plan with "([^"]*)" student$`:                           s.aCourseAndStudyPlanWithStudent,
		`^student submit unrelated assignment$`:                                      s.studentSubmitUnrelatedAssignment,
		`^user submit their assignment "([^"]*)" times$`:                             s.userSubmitTheirAssignmentTimes,
		`^our system must records all the submissions from student "([^"]*)" times$`: s.ourSystemMustRecordsAllTheSubmissionsFromStudent,
		`^our system must records highest grade from assignment$`:                    s.ourSystemMustRecordsHighestGradeFromAssignment,
		`^teacher grade the assignments$`:                                            s.teacherGradeTheAssignments,
		`^user submit their assignment with old submission endpoint$`:                s.userSubmitTheirAssignmentWithOldSubmissionEndpoint,
	}

	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, role string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	userID, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, role)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	switch role {
	case "school admin":
		stepState.AdminToken = authToken
	case "student":
		stepState.StudentID = userID
	}
	stepState.UserID = userID
	stepState.Token = authToken
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
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	loResult, err := utils.GenerateLearningObjectivesV2(
		s.AuthHelper.SignedCtx(ctx, stepState.Token),
		stepState.TopicIDs[0],
		1,
		cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_LEARNING,
		nil,
		s.EurekaConn,
	)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateLearningObjectivesV2: %w", err)
	}
	stepState.LearningMaterialID = loResult.LoIDs[0]
	assignmentResult, err := utils.GenerateAssignment(
		s.AuthHelper.SignedCtx(ctx, stepState.Token),
		stepState.TopicIDs[0],
		1,
		loResult.LoIDs,
		s.EurekaConn,
		nil,
	)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateAssignment: %w", err)
	}
	stepState.AssignmentIDs = assignmentResult.AssignmentIDs
	stepState.LearningMaterialID = assignmentResult.AssignmentIDs[0]
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	now := time.Now()
	req := &epb.UpsertAdHocAssignmentRequest{
		StudentId:   stepState.UserID,
		CourseId:    stepState.CourseID,
		ChapterName: "chapter example",
		TopicName:   "topic example",
		StartDate:   timestamppb.New(now.Add(-24 * time.Hour)),
		EndDate:     timestamppb.New(now.Add(24 * time.Hour)),
		Assignment: &epb.Assignment{
			AssignmentId: idutil.ULIDNow(),
			Name:         "Task assignment",
		},
	}
	resp, err := epb.NewAssignmentModifierServiceClient(s.EurekaConn).UpsertAdHocAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("epb.NewAssignmentModifierServiceClient(s.EurekaConn).UpsertAdHocAssignment: %w", err)
	}
	stepState.LearningMaterialID = resp.AssignmentId
	return utils.StepStateToContext(ctx, stepState), nil
}
