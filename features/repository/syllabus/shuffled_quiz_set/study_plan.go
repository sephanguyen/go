package shuffled_quiz_set

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *Suite) userCreateAStudyPlanOfAssignmentToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.aValidStudyPlanInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aValidAssignmentInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aValidStudyPlanItemInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aValidAssignmentStudyPlanItemInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aValidLoToStudyPlanItemInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidStudyPlanInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.StudyPlanID = idutil.ULIDNow()
	studyplan := &entities.StudyPlan{}
	database.AllNullEntity(studyplan)
	now := timeutil.Now()
	if err := multierr.Combine(studyplan.ID.Set(stepState.StudyPlanID),
		studyplan.Name.Set(fmt.Sprintf("StudyPlan_name+%s", stepState.StudyPlanID)),
		studyplan.StudyPlanType.Set(fmt.Sprintf("%d", 2)),
		studyplan.SchoolID.Set(int32(1)),
		studyplan.CourseID.Set(idutil.ULIDNow()),
		studyplan.BookID.Set(idutil.ULIDNow()),
		studyplan.Status.Set(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE),
		studyplan.TrackSchoolProgress.Set(true),
		studyplan.Grades.Set(int32(1)),
		studyplan.UpdatedAt.Set(now),
		studyplan.CreatedAt.Set(now),
		studyplan.MasterStudyPlan.Set(stepState.StudyPlanID)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup StudyPlan: %w", err)
	}
	studyplanRepo := repositories.StudyPlanRepo{}
	if err := studyplanRepo.BulkUpsert(ctx, s.DB, []*entities.StudyPlan{studyplan}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat StudyPlan: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidStudyPlanItemInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.StudyPlanItemID = idutil.ULIDNow()
	studyplanItem := &entities.StudyPlanItem{}
	database.AllNullEntity(studyplanItem)
	now := timeutil.Now()
	if err := multierr.Combine(
		studyplanItem.ID.Set(stepState.StudyPlanItemID),
		studyplanItem.StudyPlanID.Set(stepState.StudyPlanID),
		studyplanItem.AvailableFrom.Set(now),
		studyplanItem.AvailableTo.Set(now),
		studyplanItem.StartDate.Set(now),
		studyplanItem.EndDate.Set(now),
		studyplanItem.CompletedAt.Set(now),
		studyplanItem.ContentStructure.Set(entities.ContentStructure{
			CourseID:  idutil.ULIDNow(),
			BookID:    idutil.ULIDNow(),
			ChapterID: idutil.ULIDNow(),
			TopicID:   stepState.TopicID,
			LoID:      stepState.LoID,
		}),
		studyplanItem.ContentStructureFlatten.Set("a"),
		studyplanItem.DisplayOrder.Set(0),
		studyplanItem.CopyStudyPlanItemID.Set(stepState.StudyPlanItemID),
		studyplanItem.Status.Set(epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE),
		studyplanItem.SchoolDate.Set(now),

		studyplanItem.UpdatedAt.Set(now),
		studyplanItem.CreatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup StudyPlanItem: %w", err)
	}
	studyplanItemRepo := repositories.StudyPlanItemRepo{}
	if err := studyplanItemRepo.BulkInsert(ctx, s.DB, []*entities.StudyPlanItem{studyplanItem}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat StudyPlanItem: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidAssignmentInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.AssignmentID = idutil.ULIDNow()
	stepState.LoID = idutil.ULIDNow()
	stepState.TopicID = idutil.ULIDNow()
	assignment := &entities.Assignment{}
	database.AllNullEntity(assignment)
	if err := multierr.Combine(
		assignment.CreatedAt.Set(time.Now()),
		assignment.UpdatedAt.Set(time.Now()),
		assignment.ID.Set(stepState.AssignmentID),
		assignment.Content.Set(epb.AssignmentContent{
			TopicId: stepState.TopicID,
			LoId:    []string{stepState.LoID},
		}),
		assignment.Attachment.Set([]string{"media-id-1", "media-id-2"}),
		assignment.Settings.Set(&epb.AssignmentSetting{
			AllowLateSubmission: false,
			AllowResubmission:   false,
			RequireAttachment:   false,
		}),
		assignment.CheckList.Set(&epb.CheckList{
			Items: []*epb.CheckListItem{{Content: "Complete all learning objectives", IsChecked: true}, {Content: "Submitted required videos", IsChecked: false}},
		}),
		assignment.Name.Set(fmt.Sprintf("assignment-%s", idutil.ULIDNow())),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup assignment: %w", err)
	}

	assignmentRepo := repositories.AssignmentRepo{}
	if err := assignmentRepo.BulkUpsert(ctx, s.DB, []*entities.Assignment{assignment}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat assignmentStudyPlanItem: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidAssignmentStudyPlanItemInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignmentStudyplanItem := &entities.AssignmentStudyPlanItem{}
	database.AllNullEntity(assignmentStudyplanItem)
	now := timeutil.Now()

	if err := multierr.Combine(
		assignmentStudyplanItem.AssignmentID.Set(stepState.AssignmentID),
		assignmentStudyplanItem.StudyPlanItemID.Set(stepState.StudyPlanItemID),
		assignmentStudyplanItem.UpdatedAt.Set(now),
		assignmentStudyplanItem.CreatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup assignment studyplan item: %w", err)
	}

	assignmentStudyplanItemRepo := repositories.AssignmentStudyPlanItemRepo{}

	if err := assignmentStudyplanItemRepo.BulkInsert(ctx, s.DB, []*entities.AssignmentStudyPlanItem{assignmentStudyplanItem}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat AssignmentStudyPlanItem: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidLoToStudyPlanItemInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	loStudyplanItem := &entities.LoStudyPlanItem{}
	database.AllNullEntity(loStudyplanItem)
	now := timeutil.Now()
	if err := multierr.Combine(
		loStudyplanItem.UpdatedAt.Set(now),
		loStudyplanItem.CreatedAt.Set(now),
		loStudyplanItem.LoID.Set(stepState.LoID),
		loStudyplanItem.StudyPlanItemID.Set(stepState.StudyPlanItemID),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup lo studyplan item: %w", err)
	}

	loStudyplanItemRepo := repositories.LoStudyPlanItemRepo{}

	if err := loStudyplanItemRepo.BulkInsert(ctx, s.DB, []*entities.LoStudyPlanItem{loStudyplanItem}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat lo studyplan item: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
