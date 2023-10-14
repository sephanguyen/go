package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userLeavesTheCurrentLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		LeaveLiveRoom(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.LeaveLiveRoomRequest{
			ChannelId: stepState.CurrentChannelID,
			UserId:    stepState.CurrentUserID,
		})

	return StepStateToContext(ctx, stepState), nil
}
