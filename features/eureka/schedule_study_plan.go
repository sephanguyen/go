package eureka

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"golang.org/x/net/context"
)

func (s *suite) studyPlanAndAssignmentExistsInDb(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())
	if stepState.BookID == "" {
		if ctx, err := s.insertBookIntoBob(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("insertBookIntoBob: %w", err)
		}

		if ctx, err := s.insertChapterIntoBob(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("insertChapterIntoBob: %w", err)
		}

		if ctx, err := s.insertBookChapterIntoBob(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("insertBookChapterIntoBob: %w", err)
		}

		if ctx, err := s.insertTopicIntoBob(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("insertTopicIntoBob: %w", err)
		}
	}

	_, _ = s.DB.Exec(ctx, "UPDATE study_plans SET book_id = $1 WHERE study_plan_id = $2 OR master_study_plan_id = $2", stepState.BookID, stepState.StudyPlanID)

	for i := 0; i < 13; i++ {
		stepState.StudyPlanItemIDs = append(stepState.StudyPlanItemIDs, "study-plan-item-id-"+strconv.Itoa(i))
	}
	if ctx, err := s.userCreateNewAssignments(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.userCreateNewAssignments: %w", err)
	}
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create assignment: %w", stepState.ResponseErr)
	}
	stepState.AssignmentIDs = stepState.Response.(*pb.UpsertAssignmentsResponse).AssignmentIds

	values := entities.StudyPlanItems{}
	if err := try.Do(func(attempt int) (bool, error) {
		fields, _ := (&entities.StudyPlanItem{}).FieldMap()
		query := fmt.Sprintf(`SELECT %s
		FROM study_plan_items
		WHERE study_plan_id = $1 AND deleted_at IS NULL`,
			strings.Join(fields, ","))

		if err := database.Select(ctx, db, query, &stepState.StudyPlanID).ScanAll(&values); err != nil {
			return true, err
		}

		if len(values) > 0 {
			return false, nil
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("studyPlanItems of assignments not created")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, value := range values {
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, s.convertStudyPlanItemEntitiesToPb(value))
	}

	if ctx, err := s.userUpsertAListOfStudyPlanItemWithExistStudyPlanItems(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to upsert study plan item: %w", err)
	}
	stepState.StudyPlanItemIDs = stepState.Response.(*pb.UpsertStudyPlanItemResponse).StudyPlanItemIds

	if ctx, err := s.returnsStatusCode(ctx, "OK"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
