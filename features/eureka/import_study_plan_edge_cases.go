package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) studyPlanItemsWrongBookID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count(*)
	FROM study_plan_items spi 
	JOIN study_plans sp ON sp.study_plan_id  = spi.study_plan_id 
	WHERE spi.content_structure ->> 'book_id' <> sp.book_id AND sp.course_id = $1`

	count := 0
	if err := s.DB.QueryRow(ctx, query, stepState.CourseID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("internal error: %w", err)
	}

	if count != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of study plan items (wrong bookID): expected 0, got %v", count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlanItemsForStudentOfLOsAndAssignmentsMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.UpsertLOsAndAssignmentsResponse)
	assignmentIDs := resp.AssignmentIds
	loIDs := resp.LoIds

	try.Do(func(attempt int) (bool, error) {
		query := "SELECT count(study_plan_item_id) FROM study_plan_items WHERE (content_structure->>'lo_id' = ANY($1) OR content_structure->>'assignment_id' = ANY($2)) AND deleted_at IS NULL"

		var countLOs pgtype.Int8
		if err := s.DB.QueryRow(ctx, query, &loIDs, &assignmentIDs).Scan(&countLOs); err != nil {
			return true, fmt.Errorf("internal error: %w", err)
		}

		if countLOs.Status == pgtype.Null || countLOs.Int != int64(len(loIDs)+len(assignmentIDs)) {
			time.Sleep(1 * time.Second)
			return attempt < 3, fmt.Errorf("study plan items of LOs and Assignments weren't created")
		}

		return false, nil
	})

	return StepStateToContext(ctx, stepState), nil
}
