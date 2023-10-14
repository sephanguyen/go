package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) getBookIdsBelongToStudentStudyPlanItems(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg1 == "invalid" {
		stepState.AuthToken = ""
	}
	ctx = contextWithToken(s, ctx)
	req := &epb.GetBookIDsBelongsToStudentStudyPlanRequest{
		StudentId: stepState.StudentID,
		BookIds:   []string{stepState.BookID},
	}
	resp, err := epb.NewStudyPlanReaderServiceClient(s.Conn).GetBookIDsBelongsToStudentStudyPlan(ctx, req)
	stepState.ResponseErr = err
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHasToGetBookIdsBelongToStudentStudyPlanItemsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, ok := stepState.Response.(*epb.GetBookIDsBelongsToStudentStudyPlanResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to cast the response")
	}
	var counter pgtype.Int4

	stmt := `SELECT count(DISTINCT content_structure->>'book_id')::INT
	FROM study_plan_items JOIN student_study_plans USING (study_plan_id) 
	WHERE content_structure->>'book_id' = ANY($1) AND student_id = $2
	`
	err := s.DB.QueryRow(ctx, stmt, database.TextArray(resp.BookIds), database.Text(stepState.StudentID)).Scan(&counter)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to get course_students: %w", err)
	}
	if int(counter.Int) != 1 { // we only use one book
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong number of books, want %d, actual %d", 1, counter.Int)
	}

	return StepStateToContext(ctx, stepState), nil
}
