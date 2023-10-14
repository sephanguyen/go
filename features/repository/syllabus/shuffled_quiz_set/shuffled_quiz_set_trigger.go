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
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *Suite) ourSystemStoredStudyPlanItemIdentityOfShuffleQuizSetCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	e := &entities.ShuffledQuizSet{}
	stmt := fmt.Sprintf(`SELECT study_plan_id, learning_material_id FROM %s WHERE shuffled_quiz_set_id = $1`, e.TableName())

	var (
		studyPlanID        pgtype.Text
		learningMaterialID pgtype.Text
	)

	if err := s.DB.QueryRow(ctx, stmt, stepState.ShuffledQuizSetID).Scan(&studyPlanID, &learningMaterialID); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve shuffled quiz set: %w", err)
	}

	if stepState.LoID != learningMaterialID.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("shuffled quiz set stored learning material id wrong, expect %v but got %v", stepState.LoID, learningMaterialID.String)
	}

	if stepState.StudyPlanID != studyPlanID.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("shuffled quiz set stored study plan id wrong, expect %v but got %v", stepState.StudyPlanID, studyPlanID.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateAShuffleQuizSet(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studentID := idutil.ULIDNow()
	ctx, err := s.aValidUserInDB(ctx, s.DBTrace, studentID, constant.RoleStudent, constant.RoleStudent)
	if err != nil {
		return ctx, err
	}
	now := time.Now()

	e := &entities.ShuffledQuizSet{}
	database.AllNullEntity(e)

	stepState.ShuffledQuizSetID = idutil.ULIDNow()
	if err := multierr.Combine(
		e.ID.Set(stepState.ShuffledQuizSetID),
		e.StudentID.Set(studentID),
		e.StudyPlanItemID.Set(stepState.StudyPlanItemID),
		e.TotalCorrectness.Set(1),
		e.SubmissionHistory.Set(database.JSONB([]*entities.QuizAnswer{})),
		e.QuizExternalIDs.Set([]string{}),
		e.OriginalQuizSetID.Set(idutil.ULIDNow()),
		e.Status.Set(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),

		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value: %w", err)
	}

	shuffledQuizSetRepo := &repositories.ShuffledQuizSetRepo{}
	if _, err := shuffledQuizSetRepo.Create(ctx, s.DB, e); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create shuffled quiz set: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
