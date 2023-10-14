package student_event_log

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

func (s *Suite) validStudentEventLogInDB(ctx context.Context, eventType, createdAt string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.StudentID = idutil.ULIDNow()
	stepState.LearningMaterialID = idutil.ULIDNow()

	if err := s.insertStudyPlanItem(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert StudyPlanItem: %w", err)
	}
	if err := s.upsertStudentEventLogByEventType(ctx, eventType, createdAt); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("upsert StudentEventLog: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validExamLOInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.LearningMaterialID = idutil.ULIDNow()
	topicID := idutil.ULIDNow()
	now := time.Now()

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO topics (topic_id, name, grade, subject, topic_type, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $6);`,
		database.Text(topicID),
		database.Text(idutil.ULIDNow()),
		database.Int2(10),
		database.Text("SUBJECT_NONE"),
		database.Text("TOPIC_STATUS_NONE"),
		database.Timestamptz(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert exam_lo: %w", err)
	}

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO exam_lo (learning_material_id, topic_id, name, type, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $5);`,
		database.Text(stepState.LearningMaterialID),
		database.Text(topicID),
		database.Text(stepState.LearningMaterialID),
		database.Text("LEARNING_MATERIAL_EXAM_LO"),
		database.Timestamptz(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert exam_lo: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aStudentSubmitExamLO(ctx context.Context, submit string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.StudentID = idutil.ULIDNow()
	now := time.Now()

	// Study Plan
	if err := s.insertStudyPlanItem(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert StudyPlanItem: %w", err)
	}

	if submit == "submitted" {
		// Exam LO Submission
		if _, err := s.DB.Exec(
			ctx,
			`INSERT INTO exam_lo_submission (
				submission_id, student_id, study_plan_id, learning_material_id, shuffled_quiz_set_id, created_at, updated_at
			)
			VALUES($1, $2, $3, $4, $5, $6, $6);`,
			database.Text(idutil.ULIDNow()),
			database.Text(stepState.StudentID),
			database.Text(stepState.StudyPlanID),
			database.Text(stepState.LearningMaterialID),
			database.Text(idutil.ULIDNow()),
			database.Timestamptz(now),
		); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert exam_lo: %w", err)
		}
	}

	if err := s.upsertStudentEventLogByEventType(ctx, "completed", "2023-01-03T08:00:00Z"); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("upsert StudentEventLog: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validUpdateEventStudentEventLogInDB(ctx context.Context, eventType, createdAt string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if err := s.upsertStudentEventLogByEventType(ctx, eventType, createdAt); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("upsert StudentEventLog: %w", err)
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
		stepState.StudyPlanID, stepState.StudentID, stepState.LearningMaterialID,
	).Scan(&stepState.CompletedDate); err != nil && err != pgx.ErrNoRows {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validCompletedTime(ctx context.Context, completedAt string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if got := stepState.CompletedDate; got.Status == pgtype.Null && completedAt != "" || got.Status == pgtype.Present && got.Time.UTC().Format(time.RFC3339) != completedAt {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_at: expected: %s, got: %s", completedAt, got.Time.UTC().Format(time.RFC3339))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
