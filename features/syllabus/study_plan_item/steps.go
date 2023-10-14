package study_plan_item

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
)

type StepState struct {
	Token                      string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	BookID                     string
	TopicIDs                   []string
	ChapterIDs                 []string
	SchoolAdmin                entity.SchoolAdmin
	Student                    entity.Student
	Teacher                    entity.Teacher
	Parent                     entity.Parent
	HQStaff                    entity.HQStaff
	LearningMaterialID         string
	LearningMaterialIDs        []string
	TopicLODisplayOrderCounter int32
	SchoolID                   string
	StudyPlanID                string
	StudyPlanIDs               []string
	TopicID                    string
	LoIDs                      []string
	AssignmentID               string
	AssignmentIDs              []string
	CourseID                   string
	Grade                      int32
	AvailableStudyPlanIDs      []string
	AssignedStudentIDs         []string
	StudentIDs                 []string
	SchoolIDInt                int32
	StudyPlanItemStartDate     time.Time
	StudyPlanItemEndDate       time.Time
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		// BEGIN====common=====BEGIN
		`^<study_plan_item> a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<study_plan_item> returns "([^"]*)" status code$`: s.returnsStatusCode,

		// // edit assignment time
		`^<study_plan_item> "([^"]*)" has created a content book$`:                                                   s.hasCreatedAContentBook,
		`^<study_plan_item> "([^"]*)" has created a study plan exact match with the book content for (\d+) student$`: s.hasCreatedAStudyPlanExactMatchWithTheBookContentForMultipleStudent,
		`^admin update assignment time with "([^"]*)"$`:                                                              s.adminUpdateStudyPlanItemsStartEndDateWith,
		`^admin update assignment time with null data and "([^"]*)"$`:                                                s.adminUpdateStudyPlanItemsStartEndDateWithNullData,
		`^assignment time was updated with according update_type "([^"]*)"$`:                                         s.studyPlanItemsTimeWasUpdatedWithAccordingUpdateType,
		`^assignment time was updated with null data and according update_type "([^"]*)"$`:                           s.assignmentTimeWasUpdatedWithNullDataAndAccordingUpdate_type,
		`^<study_plan_item> user send update assignment time request$`:                                               s.userSendUpdateStudyPlanItemsStartEndDateRequest,
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
	return utils.StepStateToContext(ctx, stepState), nil
}
