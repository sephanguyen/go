package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	eu_v1 "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func (s *suite) assignmentWasSuccessfullyDeletedInSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, token := range []string{stepState.StudentToken, stepState.TeacherToken} {
		stepState.AuthToken = token
		res, err := eu_v1.NewAssignmentReaderServiceClient(s.Conn).RetrieveAssignments(s.signedCtx(ctx), &eu_v1.RetrieveAssignmentsRequest{
			Ids: stepState.AssignmentIDs,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to get assignments: %v", err)
		}

		if len(res.Items) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("assignments were not deleted")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminDeleteAssignment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := try.Do(func(attempt int) (bool, error) {
		var err error
		query := `
		SELECT COUNT(*) FROM assignment_study_plan_items WHERE assignment_id = ANY($1::TEXT[])
	`
		var count pgtype.Int8
		if err := s.DB.QueryRow(ctx, query, database.TextArray(stepState.AssignmentIDs)).Scan(&count); err != nil {
			return false, fmt.Errorf("unable to check assignment study plan items: %v", err)
		}
		if err != nil {
			if err == pgx.ErrNoRows {
				time.Sleep(time.Millisecond * 50)
				return attempt < 10, fmt.Errorf("no row found")
			}
			return false, err
		}

		if count.Int == 2 { // why 2, 1 for master, 1 for student
			return true, nil
		}
		time.Sleep(time.Millisecond * 50)
		return attempt < 10, fmt.Errorf("timeout sync study plan item")
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the pre-data is not correct: %v", err)
	}
	if _, err := eu_v1.NewAssignmentModifierServiceClient(s.Conn).DeleteAssignments(s.signedCtx(ctx), &eu_v1.DeleteAssignmentsRequest{
		AssignmentIds: stepState.AssignmentIDs,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable delete assignment study plan items: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlanItemsBelongToAssignmentWereSuccessfullyDeleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	assigmentStudyPlanQuery := `
		SELECT COUNT(*) FROM assignment_study_plan_items WHERE assignment_id = ANY($1::TEXT[]) AND deleted_at IS NULL
	`

	var count pgtype.Int8
	if err := try.Do(func(attempt int) (retry bool, err error) {
		if err := s.DB.QueryRow(ctx, assigmentStudyPlanQuery, database.TextArray(stepState.AssignmentIDs)).Scan(&count); err != nil {
			return false, err
		}
		if count.Int == 0 {
			return false, nil
		}
		time.Sleep(time.Second)
		return attempt < 5, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable delete assignment study plan items: %v", err)
	}

	if count.Int != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("0 assignment study plan items deleted")
	}

	studyPlanItemQuery := `
		SELECT COUNT(*) FROM study_plan_items spi 
			JOIN assignment_study_plan_items aspi ON spi.study_plan_item_id = aspi.study_plan_item_id
				AND aspi.assignment_id = ANY($1::TEXT[])
				AND aspi.deleted_at IS NULL
				AND spi.deleted_at IS NULL
	`

	if err := s.DB.QueryRow(ctx, studyPlanItemQuery, database.TextArray(stepState.AssignmentIDs)).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable delete study plan items: %v", err)
	}

	if count.Int != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("0 study plan items deleted")
	}

	return StepStateToContext(ctx, stepState), nil
}
