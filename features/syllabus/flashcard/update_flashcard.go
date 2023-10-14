package flashcard

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

var (
	flashcardUpdatedName = "Flashcard updated"
)

func (s *Suite) userUpdateAFlashcard(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	fc := &sspb.FlashcardBase{
		Base: &sspb.LearningMaterialBase{
			LearningMaterialId: stepState.LearningMaterialIDs[0],
			Name:               flashcardUpdatedName,
		},
	}
	stepState.Response, stepState.ResponseErr = sspb.NewFlashcardClient(s.EurekaConn).UpdateFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpdateFlashcardRequest{
		Flashcard: fc,
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesTheFlashcardCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	fc := &entities.Flashcard{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.LearningMaterialIDs[0]),
		},
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(fc), ","), fc.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, fc.ID).ScanOne(fc); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if fc.Name.String != flashcardUpdatedName {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect Flashcard name: expected %s, got %s", flashcardUpdatedName, fc.Name.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
