package flashcard

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) returnsLatestFlashcardStudyProgressCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.GetLastestProgressResponse)
	if resp.GetStudySetId().GetValue() != stepState.LatestFlashcardStudyProgressStudySetId {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("study set id is wrong: expect %s but got %s", stepState.LatestFlashcardStudyProgressStudySetId, resp.GetStudySetId().GetValue())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateSomeFlashcardStudies(ctx context.Context) (_ context.Context, err error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// nolint
	n := rand.Int()%5 + 2
	for i := 1; i <= n; i++ {
		ctx, err = s.userCreateFlashcardStudy(ctx)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create flashcard study: %w", err)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetLatestFlashcardStudyProgress(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.CreateFlashCardStudyResponse)
	stepState.LatestFlashcardStudyProgressStudySetId = resp.GetStudySetId()

	stepState.Response, stepState.ResponseErr = sspb.NewFlashcardClient(s.EurekaConn).
		GetLastestProgress(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.GetLastestProgressRequest{
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudyPlanId:        stepState.StudyPlanID,
				LearningMaterialId: stepState.FlashcardID,
				StudentId:          wrapperspb.String(stepState.StudentID),
			},
		})
	return utils.StepStateToContext(ctx, stepState), nil
}
