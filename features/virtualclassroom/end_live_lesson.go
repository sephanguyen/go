package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userEndTheLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.EndLiveLessonRequest{
		LessonId: stepState.CurrentLessonID,
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		EndLiveLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}
