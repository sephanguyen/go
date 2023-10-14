package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userSpotlightAUser(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
	}

	switch action {
	case "adds":
		req.Command = &vpb.ModifyVirtualClassroomStateRequest_Spotlight_{
			Spotlight: &vpb.ModifyVirtualClassroomStateRequest_Spotlight{
				UserId:      stepState.StudentIds[0],
				IsSpotlight: true,
			},
		}
		stepState.IsExpectingASpotlightedUser = true
	case "removes":
		req.Command = &vpb.ModifyVirtualClassroomStateRequest_Spotlight_{
			Spotlight: &vpb.ModifyVirtualClassroomStateRequest_Spotlight{
				IsSpotlight: false,
			},
		}
		stepState.IsExpectingASpotlightedUser = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("spotlight action %s is not valid", action)
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsCorrectSpotlightState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.LessonId); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	actualSpotlightState := res.GetSpotlight()
	expectedIsSpotlight := stepState.IsExpectingASpotlightedUser

	if expectedIsSpotlight != actualSpotlightState.GetIsSpotlight() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected spotlight state to be %v but got %v", expectedIsSpotlight, actualSpotlightState.GetIsSpotlight())
	}

	if expectedIsSpotlight && actualSpotlightState.GetUserId() == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting a spotlighted user but got %s", actualSpotlightState.GetUserId())
	}

	return StepStateToContext(ctx, stepState), nil
}
