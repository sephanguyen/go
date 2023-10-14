package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userModifiesLearnersAnnotationInTheLiveRoom(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
	}

	switch state {
	case Disables:
		req.Command = &vpb.ModifyLiveRoomStateRequest_AnnotationDisable{
			AnnotationDisable: &vpb.ModifyLiveRoomStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		}
	case Enables:
		req.Command = &vpb.ModifyLiveRoomStateRequest_AnnotationEnable{
			AnnotationEnable: &vpb.ModifyLiveRoomStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("modify annotation state step not supported: %s", state)
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedAnnotationStateLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*vpb.ModifyLiveRoomStateRequest)
	expectedAnnotation := false

	switch req.Command.(type) {
	case *vpb.ModifyLiveRoomStateRequest_AnnotationEnable:
		expectedAnnotation = true
	case *vpb.ModifyLiveRoomStateRequest_AnnotationDisable:
		expectedAnnotation = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("before request %T is invalid command", req.Command)
	}

	for _, actualLearner := range res.UsersState.Learners {
		if actualLearner.Annotation.Value != expectedAnnotation {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's annotation %v but got %v", expectedAnnotation, actualLearner.Annotation.Value)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDisablesAllAnnotationInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command:   &vpb.ModifyLiveRoomStateRequest_AnnotationDisableAll{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedAnnotationStateLiveRoomIs(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expectedAnnotation := false
	switch state {
	case Enabled:
		expectedAnnotation = true
	case Disabled:
		expectedAnnotation = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected annotation state is not supported in this step")
	}

	for _, actualLearner := range res.UsersState.Learners {
		if actualLearner.Annotation.Value != expectedAnnotation {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's annotation %v but got %v", expectedAnnotation, actualLearner.Annotation.Value)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedAnnotationStateLiveRoomIsWithWait(ctx context.Context, state string) (context.Context, error) {
	// sleep to make sure NATS sync data successfully for join live room scenario
	time.Sleep(10 * time.Second)

	return s.userGetsExpectedAnnotationStateLiveRoomIs(ctx, state)
}
