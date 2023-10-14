package student_latest_submissions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgx/v4"
)

func (s *Suite) validStudyPlanItemInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.StudyPlanID = idutil.ULIDNow()
	stepState.MasterStudyPlanID = stepState.StudyPlanID
	stepState.StudyPlanItemID = idutil.ULIDNow()
	stepState.AssignmentID = idutil.ULIDNow()
	stepState.Name = "Assignment " + idutil.ULIDNow()
	now := time.Now()
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO study_plans (study_plan_id, master_study_plan_id, created_at, updated_at) VALUES($1, $2, $3, $3);`,
			database.Text(stepState.StudyPlanID),
			database.Text(stepState.MasterStudyPlanID),
			database.Timestamptz(now),
		); err != nil {
			return fmt.Errorf("Insert study_plans: %w", err)
		}
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO study_plan_items (study_plan_item_id, study_plan_id, created_at, updated_at) VALUES($1, $2, $3, $3);`,
			database.Text(stepState.StudyPlanItemID),
			database.Text(stepState.StudyPlanID),
			database.Timestamptz(now),
		); err != nil {
			return fmt.Errorf("Insert study_plan_items: %w", err)
		}
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO student_study_plans (student_id, study_plan_id, created_at, updated_at) VALUES($1, $2, $3, $3);`,
			database.Text(idutil.ULIDNow()),
			database.Text(stepState.StudyPlanID),
			database.Timestamptz(now),
		); err != nil {
			return fmt.Errorf("Insert student_study_plans: %w", err)
		}
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO assignments (assignment_id, "name", created_at, updated_at) VALUES($1, $2, $3, $3);`,
			database.Text(stepState.AssignmentID),
			database.Text(stepState.Name),
			database.Timestamptz(now),
		); err != nil {
			return fmt.Errorf("Insert assignments: %w", err)
		}
		return nil
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("database.ExecInTx: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) insertAValidStudentLatestSubmission(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.StudentID = idutil.ULIDNow()
	now := time.Now()
	if _, err := s.DB.Exec(
		ctx,
		`
		INSERT INTO student_latest_submissions (study_plan_item_id, assignment_id, student_id, student_submission_id, created_at, updated_at)
			VALUES($1, $2, $3, $4, $5, $5);
		`,
		database.Text(stepState.StudyPlanItemID),
		database.Text(stepState.AssignmentID),
		database.Text(stepState.StudentID),
		database.Text(idutil.ULIDNow()),
		database.Timestamptz(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Insert student_latest_submissions: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentLatestSubmissionNewIdentityFilled(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studyPlanID := database.Text("")
	learningMaterialID := database.Text("")
	if err := s.DB.QueryRow(ctx, `
		SELECT
			study_plan_id,
			learning_material_id
		FROM student_latest_submissions
		WHERE
			study_plan_item_id = $1
			AND assignment_id = $2
			AND student_id = $3;
		`,
		database.Text(stepState.StudyPlanItemID),
		database.Text(stepState.AssignmentID),
		database.Text(stepState.StudentID),
	).Scan(&studyPlanID, &learningMaterialID); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if studyPlanID.String != stepState.StudyPlanID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Wrong study_plan_id, expected: %s, got: %s", stepState.StudyPlanID, studyPlanID.String)
	}
	if learningMaterialID.String != stepState.AssignmentID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Wrong study_plan_id, expected: %s, got: %s", stepState.AssignmentID, learningMaterialID.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
