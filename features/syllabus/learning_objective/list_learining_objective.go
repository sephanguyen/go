package learning_objective

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userListLearningObjectives(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewLearningObjectiveClient(s.EurekaConn).ListLearningObjective(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListLearningObjectiveRequest{
		LearningMaterialIds: stepState.LearningMaterialIDs,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnLearningObjectivesCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	response := stepState.Response.(*sspb.ListLearningObjectiveResponse)

	if len(response.LearningObjectives) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of learning objectives, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.LearningObjectives))
	}

	for _, lo := range response.LearningObjectives {
		if !golibs.InArrayString(lo.Base.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective id: %q in list %v of learning objectives: %q", lo.Base.LearningMaterialId, stepState.LearningMaterialIDs, response.LearningObjectives)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validLearningObjectivesQuizzesByLoIDs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	for _, loID := range stepState.LearningMaterialIDs {
		// create quizzes
		if err := utils.GenerateQuizzes(s.AuthHelper.SignedCtx(ctx, stepState.Token), loID, rand.Intn(7)+3, nil, s.EurekaConn); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuizzes: %w", err)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnLOHasTotalQuestion(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.ListLearningObjectiveResponse)

	if len(response.LearningObjectives) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of learning objectives, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.LearningObjectives))
	}

	for _, lo := range response.LearningObjectives {
		if lo.TotalQuestion == 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect total question of Learning Objectives, %d", lo.TotalQuestion)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
