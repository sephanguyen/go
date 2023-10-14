package flashcard

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userListFlashcard(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ListFlashcardReq := &sspb.ListFlashcardRequest{
		LearningMaterialIds: stepState.LearningMaterialIDs,
	}
	stepState.Response, stepState.ResponseErr = sspb.NewFlashcardClient(s.EurekaConn).ListFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), ListFlashcardReq)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnFlashcardsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.ListFlashcardResponse)
	if len(response.Flashcards) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of flashcards, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.Flashcards))
	}

	for _, flashcard := range response.Flashcards {
		if !golibs.InArrayString(flashcard.Base.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected flashcard id: %q in list %v of flashcards: %q", flashcard.Base.LearningMaterialId, stepState.LearningMaterialIDs, response.Flashcards)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validFCQuizzesByLoIDs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	for _, loID := range stepState.LearningMaterialIDs {
		// create quizzes
		if err := utils.GenerateQuizzes(s.AuthHelper.SignedCtx(ctx, stepState.Token), loID, rand.Intn(7)+3, nil, s.EurekaConn); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuizzes: %w", err)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnFCHasTotalQuestion(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	response := stepState.Response.(*sspb.ListFlashcardResponse)
	if len(response.Flashcards) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of flashcards, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.Flashcards))
	}

	for _, flashcard := range response.Flashcards {
		if flashcard.TotalQuestion == 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect total question of flashcard, %d", flashcard.TotalQuestion)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
