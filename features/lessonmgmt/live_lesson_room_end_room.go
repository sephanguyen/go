package lessonmgmt

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	pbb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) EndLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pbb.EndLiveLessonRequest{
		LessonId: stepState.CurrentLessonID,
	}
	stepState.Response, stepState.ResponseErr = pbb.NewClassModifierServiceClient(s.BobConn).
		EndLiveLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}
