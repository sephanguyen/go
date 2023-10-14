package exam_lo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

func (s *Suite) userListExamLOs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).ListExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListExamLORequest{
		LearningMaterialIds: stepState.LearningMaterialIDs,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnExamLOsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	response := stepState.Response.(*sspb.ListExamLOResponse)

	if len(response.ExamLos) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of exam LOs, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.ExamLos))
	}

	for _, examLo := range response.ExamLos {
		if !golibs.InArrayString(examLo.Base.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO id: %q in list %v of exam LOs: %q", examLo.Base.LearningMaterialId, stepState.LearningMaterialIDs, response.ExamLos)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validQuizSetForExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err1 := s.someStudentsAddedToCourseInSomeValidLocations(ctx)
	ctx, err2 := s.aQuizTestIncludeMultipleChoiceQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx, "2", "2")
	if err := multierr.Combine(err1, err2); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("validQuizSetForExamLO %w", err)
	}

	stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, stepState.LoID)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnExamLOHasTotalQuestion(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	response := stepState.Response.(*sspb.ListExamLOResponse)

	if len(response.ExamLos) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of exam LOs, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.ExamLos))
	}

	for _, examLo := range response.ExamLos {
		if examLo.Base.LearningMaterialId == stepState.LoID {
			if examLo.TotalQuestion == 0 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect total question of exam LOs, %d", examLo.TotalQuestion)
			}
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
