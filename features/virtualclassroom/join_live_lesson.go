package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userJoinVirtualClassRoomInVirtualClassroom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		JoinLiveLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.JoinLiveLessonRequest{
			LessonId: stepState.CurrentLessonID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesRoomIDAndOtherTokens(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.JoinLiveLessonResponse)

	if len(response.RoomId) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing room id when %s join live lesson", user)
	}

	if len(response.WhiteboardToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard token when %s join live lesson", user)
	}

	if len(response.WhiteboardAppId) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard app ID when %s join live lesson", user)
	}

	if len(response.StmToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing stm token when %s join live lesson", user)
	}

	if len(response.StreamToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing stream token when %s join live lesson", user)
	}

	if user == "teacher" || user == "staff" {
		if len(response.VideoToken) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("missing video token when teacher join live lesson")
		}

		if len(response.ScreenRecordingToken) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("missing screen recording token when teacher join live lesson")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
