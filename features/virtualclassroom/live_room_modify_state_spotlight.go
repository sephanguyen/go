package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userSpotlightAUserInTheLiveRoom(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
	}

	switch action {
	case "adds":
		req.Command = &vpb.ModifyLiveRoomStateRequest_Spotlight_{
			Spotlight: &vpb.ModifyLiveRoomStateRequest_Spotlight{
				UserId:      stepState.StudentIds[0],
				IsSpotlight: true,
			},
		}
	case "removes":
		req.Command = &vpb.ModifyLiveRoomStateRequest_Spotlight_{
			Spotlight: &vpb.ModifyLiveRoomStateRequest_Spotlight{
				IsSpotlight: false,
			},
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("spotlight action %s in this step is not valid", action)
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsCorrectSpotlightStateInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*vpb.ModifyLiveRoomStateRequest)

	actualSpotlightState := res.GetSpotlight()
	expectedIsSpotlight := req.GetSpotlight().GetIsSpotlight()

	if expectedIsSpotlight != actualSpotlightState.GetIsSpotlight() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected spotlight state to be %v but got %v", expectedIsSpotlight, actualSpotlightState.GetIsSpotlight())
	}

	if expectedIsSpotlight && actualSpotlightState.GetUserId() == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting a spotlighted user but got %s", actualSpotlightState.GetUserId())
	}

	return StepStateToContext(ctx, stepState), nil
}

// user gets empty spotlight in the live room
func (s *suite) userGetsEmptySpotlightInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	actualSpotlightState := res.GetSpotlight()
	if actualSpotlightState.GetIsSpotlight() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected spotlight state to be %v but got %v", false, actualSpotlightState.GetIsSpotlight())
	}

	if actualSpotlightState.GetUserId() != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting empty spotlight user but got %s", actualSpotlightState.GetUserId())
	}

	return StepStateToContext(ctx, stepState), nil
}
