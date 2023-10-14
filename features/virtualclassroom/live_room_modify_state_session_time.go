package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userModifiesTheSessionTimeInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command: &vpb.ModifyLiveRoomStateRequest_UpsertSessionTime{
			UpsertSessionTime: true,
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsSessionTimeInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return nil, stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// session time should not be empty or not before the current time (with 5 min buffer)
	now := time.Now().Add(-5 * time.Minute)
	if res.SessionTime.AsTime().IsZero() || res.SessionTime.AsTime().Before(now) {
		return nil, fmt.Errorf("expected session time but got empty or session time is in the past")
	}

	return StepStateToContext(ctx, stepState), nil
}
