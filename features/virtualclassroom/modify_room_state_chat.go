package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userUpdatesChatOfLearnersInVirtualClassroom(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var request *vpb.ModifyVirtualClassroomStateRequest
	switch state {
	case "enables":
		request = &vpb.ModifyVirtualClassroomStateRequest{
			Id: stepState.CurrentLessonID,
			Command: &vpb.ModifyVirtualClassroomStateRequest_ChatEnable{
				ChatEnable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
					Learners: stepState.StudentIds,
				},
			},
		}
	case "disables":
		request = &vpb.ModifyVirtualClassroomStateRequest{
			Id: stepState.CurrentLessonID,
			Command: &vpb.ModifyVirtualClassroomStateRequest_ChatDisable{
				ChatDisable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
					Learners: stepState.StudentIds,
				},
			},
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("state entered is not supported")
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), request)
	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsLearnersChatPermissionWithWait(ctx context.Context, state string) (context.Context, error) {
	// sleep to make sure NATS sync data successfully for join live lesson scenario
	time.Sleep(10 * time.Second)

	return s.userGetsLearnersChatPermission(ctx, state)
}

func (s *suite) userGetsLearnersChatPermission(ctx context.Context, state string) (context.Context, error) {
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

	var expectedChatPermission bool
	switch state {
	case "enabled":
		expectedChatPermission = true
	case "disabled":
		expectedChatPermission = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported permission state")
	}

	for _, actualLearner := range res.UsersState.Learners {
		if actualLearner.Chat.Value != expectedChatPermission {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's chat permission %v but got %v", expectedChatPermission, actualLearner.Chat.Value)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
