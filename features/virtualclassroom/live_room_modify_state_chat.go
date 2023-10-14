package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

const (
	Enabled  string = "enabled"
	Disabled string = "disabled"
	Enables  string = "enables"
	Disables string = "disables"
)

func (s *suite) userModifiesLearnersChatPermissionInTheLiveRoom(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
	}

	switch state {
	case Disables:
		req.Command = &vpb.ModifyLiveRoomStateRequest_ChatDisable{
			ChatDisable: &vpb.ModifyLiveRoomStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		}
	case Enables:
		req.Command = &vpb.ModifyLiveRoomStateRequest_ChatEnable{
			ChatEnable: &vpb.ModifyLiveRoomStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("modify chat permission state step not supported: %s", state)
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedChatPermissionStateLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*vpb.ModifyLiveRoomStateRequest)
	expectedChatPermission := false

	switch req.Command.(type) {
	case *vpb.ModifyLiveRoomStateRequest_ChatEnable:
		expectedChatPermission = true
	case *vpb.ModifyLiveRoomStateRequest_ChatDisable:
		expectedChatPermission = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("before request %T is invalid command", req.Command)
	}

	for _, actualLearner := range res.UsersState.Learners {
		if actualLearner.Chat.Value != expectedChatPermission {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's chat permission %v but got %v", expectedChatPermission, actualLearner.Chat.Value)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedChatPermissionStateLiveRoomIs(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expectedChatPermission := false
	switch state {
	case Enabled:
		expectedChatPermission = true
	case Disabled:
		expectedChatPermission = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected chat state is not supported in this step")
	}

	for _, actualLearner := range res.UsersState.Learners {
		if actualLearner.Chat.Value != expectedChatPermission {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's chat permission %v but got %v", expectedChatPermission, actualLearner.Chat.Value)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedChatPermissionStateLiveRoomIsWithWait(ctx context.Context, state string) (context.Context, error) {
	// sleep to make sure NATS sync data successfully for join live room scenario
	time.Sleep(10 * time.Second)

	return s.userGetsExpectedChatPermissionStateLiveRoomIs(ctx, state)
}
