package student_submissions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func (s *Suite) validStudentSubmissionInDB(ctx context.Context, status string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.StudyPlanID = idutil.ULIDNow()
	stepState.MasterStudyPlanID = stepState.StudyPlanID
	stepState.StudyPlanItemID = idutil.ULIDNow()
	stepState.AssignmentID = idutil.ULIDNow()
	stepState.StudentID = idutil.ULIDNow()
	stepState.Name = "Assignment " + idutil.ULIDNow()
	now := time.Now()

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO study_plans (study_plan_id, master_study_plan_id, created_at, updated_at) VALUES($1, $2, $3, $3);`,
		database.Text(stepState.StudyPlanID),
		database.Text(stepState.MasterStudyPlanID),
		database.Timestamptz(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert study_plans: %w", err)
	}
	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO study_plan_items (study_plan_item_id, study_plan_id, created_at, updated_at) VALUES($1, $2, $3, $3);`,
		database.Text(stepState.StudyPlanItemID),
		database.Text(stepState.StudyPlanID),
		database.Timestamptz(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert study_plan_items: %w", err)
	}
	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO assignments (assignment_id, "name", created_at, updated_at) VALUES($1, $2, $3, $3);`,
		database.Text(stepState.AssignmentID),
		database.Text(stepState.Name),
		database.Timestamptz(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert assignments: %w", err)
	}

	completed := database.TimestamptzFromPb(nil)
	if status == "completed" {
		completed = database.Timestamptz(now)
	}

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO student_submissions (study_plan_item_id, assignment_id, student_id, student_submission_id, created_at, updated_at, complete_date)
			VALUES($1, $2, $3, $4, $5, $5, $6);`,
		database.Text(stepState.StudyPlanItemID),
		database.Text(stepState.AssignmentID),
		database.Text(stepState.StudentID),
		database.Text(idutil.ULIDNow()),
		database.Timestamptz(now),
		completed,
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert assignments: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getCompletedTimeOfLM(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if err := s.DB.QueryRow(ctx, `
		SELECT completed_at 
		FROM get_student_completion_learning_material()
		WHERE study_plan_id = $1 AND student_id = $2 AND learning_material_id = $3
		LIMIT 1`,
		stepState.StudyPlanID, stepState.StudentID, stepState.AssignmentID,
	).Scan(&stepState.CompletedDate); err != nil && err != pgx.ErrNoRows {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validCompletedTime(ctx context.Context, completionStatus string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	status := map[pgtype.Status]string{
		pgtype.Null:    "null",
		pgtype.Present: "valid",
	}

	out, ok := status[stepState.CompletedDate.Status]
	if !ok {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect completed_at: got: %s", string(stepState.CompletedDate.Status))
	}

	if out != completionStatus {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_at: expected: %s, got: %s", completionStatus, out)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
