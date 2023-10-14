package eureka

import (
	"context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) removeLastQuizFromLo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, _ = s.aSignedIn(ctx, "school admin")
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).RemoveQuizFromLO(s.signedCtx(ctx), &epb.RemoveQuizFromLORequest{
		LoId:   stepState.LoIDs[0],
		QuizId: stepState.QuizIDs[len(stepState.QuizIDs)-1],
	})

	return StepStateToContext(ctx, stepState), nil
}
