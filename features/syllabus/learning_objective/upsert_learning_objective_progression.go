package learning_objective

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) upsertLOProgressionWithAnswers(ctx context.Context, answerNo int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var (
		quizAnswers = make([]*sspb.QuizAnswer, 0)
		lastIdx     = 0
	)

	for i, quiz := range stepState.QuizLOList {
		if len(quizAnswers) >= answerNo {
			break
		}

		lastIdx = i + 1
		quizAnswers = append(quizAnswers, &sspb.QuizAnswer{
			QuizId: quiz.Quiz.ExternalId,
			Answer: []*sspb.Answer{
				{
					Format: &sspb.Answer_SelectedIndex{
						SelectedIndex: uint32(i + 1),
					},
				},
			},
		})
	}

	stepState.Response, stepState.ResponseErr = sspb.NewLearningObjectiveClient(s.EurekaConn).UpsertLOProgression(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpsertLOProgressionRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LearningMaterialID,
			StudentId:          wrapperspb.String(stepState.StudentIDs[0]),
		},
		ShuffledQuizSetId: stepState.ShuffledQuizSetID,
		LastIndex:         uint32(lastIdx),
		QuizAnswer:        quizAnswers,
		SessionId:         stepState.SessionID,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}
