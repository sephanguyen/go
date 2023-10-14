package study_plan

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// nolint

func (s *Suite) assignStudyPlanToAStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	err := utils.UserAssignStudyPlanToAStudent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.Student.ID, stepState.StudyPlanID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot assign study plan to student")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validCourseAndStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var err error

	stepState.CourseID, err = utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't generate course: %v", err)
	}

	stepState.StudyPlanID, err = utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't generate study plan: %v", err)
	}

	courseStudents, err := utils.AValidCourseWithIDs(ctx, s.EurekaDB, []string{stepState.Student.ID}, stepState.CourseID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}

	if err := utils.GenerateCourseBooks(s.AuthHelper.SignedCtx(ctx, stepState.SchoolAdminToken), stepState.CourseID, []string{stepState.BookID}, s.EurekaConn); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourseBooks: %w", err)
	}
	m := rand.Int31n(3) + 1
	locationIDs := make([]string, 0, m)
	for i := int32(1); i <= m; i++ {
		id := idutil.ULIDNow()
		locationIDs = append(locationIDs, id)
	}
	for _, courseStudent := range courseStudents {
		for _, locationID := range locationIDs {
			now := time.Now()
			e := &entities.CourseStudentsAccessPath{}
			database.AllNullEntity(e)
			if err := multierr.Combine(
				e.CourseStudentID.Set(courseStudent.ID.String),
				e.CourseID.Set(courseStudent.CourseID.String),
				e.StudentID.Set(courseStudent.StudentID.String),
				e.LocationID.Set(locationID),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
			); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
			}
			if _, err := database.Insert(ctx, e, s.EurekaDB.Exec); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("database.Insert: %w", err)
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userInsertIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for _, loID := range stepState.LoIDs {
		req := utils.GenerateIndividualStudyPlanRequest(stepState.StudyPlanID, loID, stepState.Student.ID)
		_, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertIndividual(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
		if stepState.ResponseErr != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert individual study plan to student:%s ", stepState.ResponseErr.Error())
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) dataForList(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, err := s.validCourseAndStudyPlan(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.validCourseAndStudyPlanInDB: %w", err)
	}

	ctx, err = s.assignStudyPlanToAStudent(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.adminAssignStudyPlanToAStudent: %w", err)
	}

	ctx, err = s.userInsertIndividualStudyPlan(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.adminInsertIndividualStudyPlan: %w", err)
	}

	ctx, err = s.aValidAssignmentV2(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidAssignment: %w", err)
	}

	ctx, err = s.studentSubmitAssignmentV2(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.studentSubmitAssignment: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ListToDoItemStructuredBookTree(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	paging := &cpb.Paging{
		Limit: uint32(2) + 10,
	}
	for {
		stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).ListToDoItemStructuredBookTree(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListToDoItemStructuredBookTreeRequest{
			StudyPlanIdentity: &sspb.StudyPlanIdt{
				StudyPlanId: stepState.StudyPlanID,
				StudentId:   &wrapperspb.StringValue{Value: stepState.Student.ID},
			},
			Page: paging,
		})
		if stepState.ResponseErr != nil {
			return utils.StepStateToContext(ctx, stepState), nil
		}
		resp := stepState.Response.(*sspb.ListToDoItemStructuredBookTreeResponse)
		if len(resp.TodoItems) == 0 {
			break
		}
		if len(resp.TodoItems) > int(paging.Limit) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total study plan items: got: %d, want: %d", len(resp.TodoItems), paging.Limit)
		}
		stepState.PaginatedStudentStudyPlanItem = append(stepState.PaginatedStudentStudyPlanItem, resp.TodoItems)
		stepState.PaginatedTopicProgress = append(stepState.PaginatedTopicProgress, resp.TopicProgresses)
		paging = resp.NextPage
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnCorrectlyToDoItem(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.PaginatedStudentStudyPlanItem) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of paginated. studentID: %s, studyPlanID: %s", stepState.Student.ID, stepState.StudyPlanID)
	}
	if stepState.PaginatedStudentStudyPlanItem[0][0].LearningMaterial.TopicId != stepState.TopicID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected topic id want %s, got %s", stepState.TopicID, stepState.PaginatedStudentStudyPlanItem[0][0].LearningMaterial.TopicId)
	}

	items := stepState.PaginatedStudentStudyPlanItem
	for _, paginatedItems := range items {
		for _, pi := range paginatedItems {
			if pi.LearningMaterial.TopicId != stepState.TopicID {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected topic id want %s, got %s", stepState.TopicID, pi.LearningMaterial.TopicId)
			}
		}

		// check order by (learning_material_id ASC)
		for i := 1; i < len(paginatedItems); i++ {
			prevItem := paginatedItems[i-1].LearningMaterial
			item := paginatedItems[i].LearningMaterial

			prevTopicID := prevItem.TopicId

			TopicID := item.TopicId

			if prevTopicID > TopicID {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("topic_id %v must be less than %v", prevTopicID, TopicID)
			}
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentSubmitAssignmentV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := s.getAssignmentSubmissionV2(ctx)
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewAssignmentClient(s.EurekaConn).
		SubmitAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Student.Token), req)
	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("studentSubmitAssignment: %w", stepState.ResponseErr)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getAssignmentSubmissionV2(ctx context.Context) *sspb.SubmitAssignmentRequest {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.SubmitAssignmentRequest{
		Submission: &sspb.StudentSubmission{
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudyPlanId:        stepState.StudyPlanID,
				LearningMaterialId: stepState.LearningMaterialID,
				StudentId:          wrapperspb.String(stepState.Student.ID),
			},
			SubmissionContent:  []*sspb.SubmissionContent{},
			Note:               "submit",
			CompleteDate:       timestamppb.Now(),
			Duration:           int32(rand.Intn(99) + 1),
			CorrectScore:       wrapperspb.Float(rand.Float32() * 10),
			TotalScore:         wrapperspb.Float(rand.Float32() * 100),
			UnderstandingLevel: sspb.SubmissionUnderstandingLevel(rand.Intn(len(sspb.SubmissionUnderstandingLevel_value))),
		},
	}

	return req
}

func (s *Suite) aValidAssignmentV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.LearningMaterialID = stepState.LoIDs[0]
	assignmentResult, err := utils.GenerateAssignment(
		s.AuthHelper.SignedCtx(ctx, stepState.Student.Token),
		stepState.TopicIDs[0],
		1,
		stepState.LoIDs,
		s.EurekaConn,
		nil,
	)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateAssignment: %w", err)
	}
	stepState.AssignmentIDs = assignmentResult.AssignmentIDs
	return utils.StepStateToContext(ctx, stepState), nil
}
