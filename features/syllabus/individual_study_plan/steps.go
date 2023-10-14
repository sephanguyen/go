package individual_study_plan

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StepState struct {
	StartDate                  *timestamppb.Timestamp
	Token                      string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	BookID                     string
	StudentID                  string
	StudyPlanID                string
	CourseID                   string
	StudentIDs                 []string
	StudyPlanIDs               []string
	LoIDs                      []string
	SchoolDate                 *timestamp.Timestamp
	TopicIDs                   []string
	ChapterIDs                 []string
	SchoolAdmin                entity.SchoolAdmin
	Student                    entity.Student
	Teacher                    entity.Teacher
	Parent                     entity.Parent
	HQStaff                    entity.HQStaff
	LearningMaterialID         string
	FlashcardID                string
	TopicLODisplayOrderCounter int32
	LearningMaterialIDs        []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<individual_study_plan>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<individual_study_plan>a valid book content$`:          s.aValidBookContent,
		`^<individual_study_plan>returns "([^"]*)" status code$`: s.returnsStatusCode,

		`^there is a flashcard existed in topic$`: s.thereIsAFlashcardExistedInTopic,

		// individual study plan
		`^admin insert individual study plan$`:                                s.adminInsertIndividualStudyPlan,
		`^our system stores individual study plan correctly$`:                 s.ourSystemStoresIndividualStudyPlanCorrectly,
		`^admin update start date for individual study plan$`:                 s.adminUpdateStartDateForIndividualStudyPlan,
		`^our system updates start date for individual study plan correctly$`: s.ourSystemUpdatesStartDateForIndividualStudyPlanCorrectly,

		// study plan
		`^admin insert study plan$`: s.adminInsertStudyPlan,
		`^a valid study plan$`:      s.aValidStudyPlan,

		// school date
		`^user update school date$`:                 s.userUpdateSchoolDateV2,
		`^our system stores school date correctly$`: s.ourSystemStoresSchoolDateCorrectly,
		`^"([^"]*)" has created a study plan exact match with the book content for (\d+) student$`: s.hasCreatedAStudyPlanExactMatchWithTheBookContentForMultipleStudent,
		`^our system triggers data to individual study plan table correctly$`:                      s.ourSystemTriggersDataToIndividualStudyPlanTableCorrectly,
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

func (s *Suite) aValidBookContent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookID, chapterIDs, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, s.EurekaDB, constant.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}

	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.TopicIDs = topicIDs

	lo := utils.GenerateLearningObjective(stepState.TopicIDs[0])

	// Old alias
	stepState.LoIDs = append(stepState.LoIDs, lo.Info.Id)

	stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, stepState.LoIDs...)

	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.EurekaConn).UpsertLOs(s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&epb.UpsertLOsRequest{
			LearningObjectives: []*cpb.LearningObjective{
				lo,
			},
		}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create los: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
