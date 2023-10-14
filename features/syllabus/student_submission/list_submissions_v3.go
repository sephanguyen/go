package student_submission

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) someStudentsAddedToCourseInSomeValidLocations(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	m := rand.Int31n(3) + 1
	locationIDs := make([]string, 0, m)
	for i := int32(1); i <= m; i++ {
		id := idutil.ULIDNow()
		locationIDs = append(locationIDs, id)
	}

	for i := 1; i <= 10; i++ {
		studentID := idutil.ULIDNow()
		if err := s.AuthHelper.AValidUser(ctx, studentID, constants.RoleStudent); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.AuthHelper.AValidUser: %w", err)
		}
		stepState.StudentIDs = append(stepState.StudentIDs, studentID)
	}

	courseStudents, err := utils.AValidCourseWithIDs(ctx, s.EurekaDB, stepState.StudentIDs, stepState.CourseID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
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
	stepState.CourseStudents = courseStudents
	stepState.LocationIDs = locationIDs

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentsSubmitTheirAssignments(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for _, studentID := range stepState.StudentIDs {
		token, err := s.AuthHelper.GenerateExchangeToken(studentID, consta.RoleStudent)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		stepState.Token = token

		stmt := `
			SELECT learning_material_id FROM master_study_plan
			WHERE study_plan_id = $1 
			LIMIT 1
		`
		var learningMaterial pgtype.Text
		if err := s.EurekaDB.QueryRow(ctx, stmt, stepState.StudyPlanID).Scan(&learningMaterial); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("database.Select: %w", err)
		}
		submissionResp, err := sspb.NewAssignmentClient(s.EurekaConn).
			SubmitAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.SubmitAssignmentRequest{
				Submission: &sspb.StudentSubmission{
					StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
						LearningMaterialId: learningMaterial.String,
						StudyPlanId:        stepState.StudyPlanID,
						StudentId:          wrapperspb.String(studentID),
					},
					SubmissionContent: []*sspb.SubmissionContent{},
				},
			})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("submitAssignment: %w", err)
		}
		stepState.StudyPlanIdentities = append(stepState.StudyPlanIdentities, &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: learningMaterial.String,
			StudentId:          wrapperspb.String(studentID),
		})
		stepState.SubmissionIDs = append(stepState.SubmissionIDs, submissionResp.SubmissionId)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUsingListSubmissionsV(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = sspb.NewStudentSubmissionServiceClient(s.EurekaConn).ListSubmissionsV3(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListSubmissionsV3Request{
		LocationIds: stepState.LocationIDs,
		CourseId:    wrapperspb.String(stepState.CourseID),
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		Start: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		End:   timestamppb.New(time.Now().Add(24 * time.Hour)),
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsListStudentSubmissionsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.ListSubmissionsV3Response)
	for _, item := range resp.Items {
		studyPlanItem := item.StudyPlanItemIdentity
		if studyPlanItem.LearningMaterialId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong LearningMaterialId of %v", item.SubmissionId)
		}
		if studyPlanItem.StudyPlanId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong StudyPlanId of %v", item.SubmissionId)
		}
		if studyPlanItem.StudentId == nil || studyPlanItem.StudentId.Value == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong StudentId of %v", item.SubmissionId)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createAStudyPlanForThatCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studyPlanID, err := utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = studyPlanID
	repo := &repositories.StudyPlanItemRepo{}
	// Find master study plan Items
	masterStudyPlanItems, err := repo.FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	for _, masterStudyPlanItem := range masterStudyPlanItems {
		masterStudyPlanItem.AvailableFrom.Set(time.Now().Add(-24 * time.Hour))
		masterStudyPlanItem.AvailableTo.Set(time.Now().AddDate(0, 0, 10))
		masterStudyPlanItem.StartDate.Set(time.Now().Add(-23 * time.Hour))
		masterStudyPlanItem.EndDate.Set(time.Now().AddDate(0, 0, 1))
		masterStudyPlanItem.UpdatedAt.Set(time.Now())
	}
	if err := repo.BulkInsert(ctx, s.EurekaDB, masterStudyPlanItems); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update master study plan items: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
