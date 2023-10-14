package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userEndTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		EndLiveRoom(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.EndLiveRoomRequest{
			ChannelId: stepState.CurrentChannelID,
		})

	return StepStateToContext(ctx, stepState), nil
}
