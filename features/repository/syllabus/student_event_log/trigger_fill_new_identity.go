package student_event_log

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *Suite) validStudyPlanItemInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if err := s.insertStudyPlanItem(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert StudyPlanItem: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) insertAValidStudentEventLog(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.EventID = idutil.ULIDNow()
	stepState.StudentID = idutil.ULIDNow()
	stepState.LearningMaterialID = idutil.ULIDNow()
	now := time.Now()

	types := []string{"study_guide_finished", "video_finished", "learning_objective", "quiz_answer_selected"}
	randomNo := rand.Intn(len(types))
	payload := &epb.StudentEventLogPayload{
		LoId:            stepState.LearningMaterialID,
		Event:           "started",
		SessionId:       idutil.ULIDNow(),
		StudyPlanItemId: stepState.StudyPlanItemID,
	}

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO student_event_logs (event_id, student_id, event_type, payload, study_plan_item_id, created_at) VALUES($1, $2, $3, $4, $5, $6);`,
		database.Varchar(stepState.EventID),
		database.Text(stepState.StudentID),
		database.Varchar(types[randomNo]),
		database.JSONB(payload),
		database.Text(stepState.StudyPlanItemID),
		database.Timestamptz(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert student_event_logs: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentEventLogNewIdentityFilled(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studyPlanID := database.Text("")
	learningMaterialID := database.Text("")

	if err := s.DB.QueryRow(ctx, `
		SELECT 
			study_plan_id, learning_material_id
		FROM student_event_logs
		WHERE
			study_plan_item_id = $1
			AND event_id = $2
			AND student_id = $3;
		`,
		database.Text(stepState.StudyPlanItemID),
		database.Text(stepState.EventID),
		database.Text(stepState.StudentID),
	).Scan(&studyPlanID, &learningMaterialID); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if studyPlanID.String != stepState.StudyPlanID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong study_plan_id, expected: %s, got: %s", stepState.StudyPlanID, studyPlanID.String)
	}

	if learningMaterialID.String != stepState.LearningMaterialID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong learning_material_id, expected: %s, got: %s", stepState.LearningMaterialID, learningMaterialID.String)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
