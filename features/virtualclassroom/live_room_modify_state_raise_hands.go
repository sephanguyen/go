package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userModifiesHandInTheLiveRoom(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
	}

	switch state {
	case "raise":
		req.Command = &vpb.ModifyLiveRoomStateRequest_RaiseHand{}
	case "lowers":
		req.Command = &vpb.ModifyLiveRoomStateRequest_HandOff{}
	case "folds all":
		req.Command = &vpb.ModifyLiveRoomStateRequest_FoldHandAll{}
	case "folds another learner":
		req.Command = &vpb.ModifyLiveRoomStateRequest_FoldUserHand{
			FoldUserHand: stepState.StudentIds[0],
		}
	default:
		return nil, fmt.Errorf("state is not supported in the modify hand state step")
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetAllLearnersHandsUpStatesToOffInTHeLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, learner := range res.UsersState.Learners {
		if learner.HandsUp.Value {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected all learner's hands up is off but %s is not", learner.UserId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetHandsUpStateInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(res.UsersState.Learners) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected get at least 1 learners state but got 0")
	}

	expectedHandsUp := false
	req := stepState.Request.(*vpb.ModifyLiveRoomStateRequest)
	switch req.Command.(type) {
	case *vpb.ModifyLiveRoomStateRequest_RaiseHand:
		expectedHandsUp = true
	case *vpb.ModifyLiveRoomStateRequest_HandOff:
		expectedHandsUp = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("before request %T is invalid command", req.Command)
	}

	actualLearner := sliceutils.Filter(res.UsersState.Learners, func(learner *vpb.GetLiveLessonStateResponse_UsersState_LearnerState) bool {
		return (learner.UserId == stepState.CurrentStudentID)
	})

	if len(actualLearner) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected 1 info for learner id %s but got %v", stepState.CurrentStudentID, len(actualLearner))
	}

	if actualLearner[0].HandsUp.Value != expectedHandsUp {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's hands up state %v but got %v", expectedHandsUp, actualLearner[0].HandsUp.Value)
	}

	return StepStateToContext(ctx, stepState), nil
}
