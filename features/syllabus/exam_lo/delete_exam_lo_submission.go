package exam_lo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgx/v4"
)

func (s *Suite) userDeleteExamLoSubmission(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).DeleteExamLOSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.DeleteExamLOSubmissionRequest{
		SubmissionId: stepState.SubmissionID,
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) examLoSubmissionAndRelatedTablesHaveBeenDeletedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, err := s.checkAllExamLOSubmissionDeletedCorrectly(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	spi := entities.StudyPlanItem{}
	var totalSpi int
	queryCountStudyPlanItem := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE study_plan_item_id = $1::TEXT AND completed_at IS NULL`, spi.TableName())
	if err := s.EurekaDB.QueryRow(ctx, queryCountStudyPlanItem, stepState.StudyPlanItemID).Scan(&totalSpi); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("get study plan item failed with shuffled_quiz_set_id: %s, err: %w", stepState.ShuffledQuizSetID, err)
	}
	if totalSpi == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("set completed_at for study plan item failed with id: %s", stepState.StudyPlanItemID)
	}

	sel := entities.StudentEventLog{}
	var totalSel int
	queryCountStudentEventLog := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE learning_material_id = $1::TEXT AND student_id = $2::TEXT AND deleted_at IS NOT NULL`, sel.TableName())
	if err := s.EurekaDB.QueryRow(ctx, queryCountStudentEventLog, stepState.LearningMaterialID, stepState.CurrentStudentID).Scan(&totalSel); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("get student event log failed with learning_material_id and student_id: %s & %s, err: %w", stepState.LearningMaterialID, stepState.CurrentStudentID, err)
	}
	if totalSel == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete student event log failed with learning_material_id and student_id: %s & %s", stepState.LearningMaterialID, stepState.CurrentStudentID)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkAllExamLOSubmissionDeletedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// exam lo submission
	es := entities.ExamLOSubmission{}
	query := `SELECT COUNT(*) FROM %s WHERE submission_id = $1 AND deleted_at IS NOT NULL`
	querySubmission := fmt.Sprintf(query, es.TableName())
	var total int
	if err := s.EurekaDB.QueryRow(ctx, querySubmission, stepState.SubmissionID).Scan(&total); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete failed, exam lo submission submission_id: %s not deleted, err: %w", stepState.SubmissionID, err)
	}
	if total != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete failed, expected delete 1 exam lo submission, got %d", total)
	}

	// exam lo submission answer
	esa := entities.ExamLOSubmissionAnswer{}
	queryAnswer := fmt.Sprintf(query, esa.TableName())
	if err := s.EurekaDB.QueryRow(ctx, queryAnswer, stepState.SubmissionID).Scan(&total); err != nil && err != pgx.ErrNoRows {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete failed, submission_answer id: %s not deleted, err: %w", stepState.SubmissionID, err)
	}
	if total != len(stepState.ExternalIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete submission answers failed, expected delete %d got %d", len(stepState.ExternalIDs), total)
	}

	// exam lo submission score
	ess := entities.ExamLOSubmissionScore{}
	queryScore := fmt.Sprintf(query, ess.TableName())
	if err := s.EurekaDB.QueryRow(ctx, queryScore, stepState.SubmissionID).Scan(&total); err != nil && err != pgx.ErrNoRows {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete failed, submission_score id: %s not deleted, err: %w", stepState.SubmissionID, err)
	}
	if total != len(stepState.ExternalIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete submission scores failed, expected delete %d got %d", len(stepState.ExternalIDs), total)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createStudentEventLogsAfterDoQuiz(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if err := s.upsertStudentEventLogByEventType(ctx, "learning_objective"); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot create student event logs, err: %v", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) upsertStudentEventLogByEventType(ctx context.Context, eventType string) error {
	stepState := utils.StepStateFromContext[StepState](ctx)
	now := time.Now()

	types := []string{"study_guide_finished", "video_finished", "learning_objective", "quiz_answer_selected"}
	randomNo := rand.Intn(len(types))
	payload := &epb.StudentEventLogPayload{
		LoId:            stepState.LearningMaterialID,
		Event:           eventType,
		SessionId:       idutil.ULIDNow(),
		StudyPlanItemId: stepState.StudyPlanItemID,
	}

	if _, err := s.EurekaDB.Exec(
		ctx,
		`INSERT INTO student_event_logs (event_id, student_id, event_type, payload, study_plan_item_id, created_at) VALUES($1, $2, $3, $4, $5, $6);`,
		database.Varchar(idutil.ULIDNow()),
		database.Text(stepState.CurrentStudentID),
		database.Varchar(types[randomNo]),
		database.JSONB(payload),
		database.Text(stepState.StudyPlanItemID),
		database.Timestamptz(now),
	); err != nil {
		return err
	}

	return nil
}
