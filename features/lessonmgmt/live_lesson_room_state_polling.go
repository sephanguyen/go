package lessonmgmt

import (
	"context"
	"fmt"
	"reflect"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) UserStartPollingInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	Options := []*bpb.ModifyLiveLessonStateRequest_PollingOption{
		{
			Answer:    "A",
			IsCorrect: true,
		},
		{
			Answer:    "B",
			IsCorrect: false,
		},
		{
			Answer:    "C",
			IsCorrect: false,
		},
	}
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_StartPolling{
			StartPolling: &bpb.ModifyLiveLessonStateRequest_PollingOptions{
				Options: Options,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserStartPollingInLiveLessonRoomWithNumOption(ctx context.Context, numOpt int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	Options := []*bpb.ModifyLiveLessonStateRequest_PollingOption{
		{
			Answer:    "A",
			IsCorrect: true,
		},
	}
	for i := 0; i < numOpt-1; i++ {
		Options = append(Options, &bpb.ModifyLiveLessonStateRequest_PollingOption{
			Answer:    idutil.ULIDNow(),
			IsCorrect: false,
		})
	}

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_StartPolling{
			StartPolling: &bpb.ModifyLiveLessonStateRequest_PollingOptions{
				Options: Options,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetCurrentPollingStateOfLiveLessonRoomStarted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson's current time but got empty")
	}
	if res.CurrentPolling == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling but got empty")
	}

	if res.CurrentPolling.Status != bpb.PollingState_POLLING_STATE_STARTED {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current polling is not start")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserSubmitPollingAnswerInLiveLessonRoom(ctx context.Context, answer string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	answers := []string{answer}
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_SubmitPollingAnswer{
			SubmitPollingAnswer: &bpb.ModifyLiveLessonStateRequest_PollingAnswer{
				StringArrayValue: answers,
			},
		},
	}
	stepState.SubmitPollingAnswer = answers

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetPollingAnswerState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expectedAnswer := stepState.SubmitPollingAnswer
	actualLearnerSt := res.UsersState.Learners[0]
	if !reflect.DeepEqual(actualLearnerSt.PollingAnswer.StringArrayValue, expectedAnswer) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected polling answer %v but got %v", expectedAnswer, actualLearnerSt.PollingAnswer.StringArrayValue)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserStopPollingInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_StopPolling{},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetCurrentPollingStateOfLiveLessonRoomStopped(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson's current time but got empty")
	}
	if res.CurrentPolling == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling but got empty")
	}

	if res.CurrentPolling.Status != bpb.PollingState_POLLING_STATE_STOPPED {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current polling is not stop")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserEndPollingInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_EndPolling{},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetCurrentPollingStateOfLiveLessonRoomIsEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson's current time but got empty")
	}
	if res.CurrentPolling != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling is not empty")
	}

	return StepStateToContext(ctx, stepState), nil
}
