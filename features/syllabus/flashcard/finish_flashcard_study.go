package flashcard

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userFinishFlashcardStudyWith(ctx context.Context, option string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.FinishFlashCardStudyRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.FlashcardID,
			StudentId:          wrapperspb.String(stepState.StudentID),
		},
		StudySetId: stepState.StudySetID,
	}

	switch option {
	case "no restart":
		req.IsRestart = false
	case "restart":
		req.IsRestart = true
	}

	stepState.Response, stepState.ResponseErr = sspb.NewFlashcardClient(s.EurekaConn).
		FinishFlashCardStudy(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesFlashcardStudyCorrectlyWith(ctx context.Context, option string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	e := &entities.FlashcardProgression{}
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE study_set_id = $1", e.TableName())
	switch option {
	case "no restart":
		query = fmt.Sprintf("%s AND completed_at IS NOT NULL", query)
	case "restart":
		query = fmt.Sprintf("%s AND deleted_at IS NOT NULL", query)
	}

	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, database.Text(stepState.StudySetID)).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.QueryRow: %w", err)
	}

	if count != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("finish flashcard study with option: '%s' is failed", option)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
