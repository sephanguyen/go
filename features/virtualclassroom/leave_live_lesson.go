package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userLeaveVirtualClassRoomInVirtualClassroom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		LeaveLiveLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.LeaveLiveLessonRequest{
			LessonId: stepState.CurrentLessonID,
			UserId:   stepState.CurrentUserID,
		})

	return StepStateToContext(ctx, stepState), nil
}
