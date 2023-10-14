package virtualclassroom

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

const (
	Started string = "started"
	Stopped string = "stopped"
	Empty   string = "empty"
)

func (s *suite) userStartPollingInTheLiveRoomWithNumOption(ctx context.Context, numOptStr string, numCorrectAnswerStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOpt, _ := strconv.Atoi(numOptStr)
	numCorrectAnswer, _ := strconv.Atoi(numCorrectAnswerStr)

	Options := []*vpb.ModifyLiveRoomStateRequest_PollingOption{
		{
			Answer:  "A",
			Content: "Runeterra",
		},
		{
			Answer:  "B",
			Content: "Kingdom",
		},
	}

	for i := 0; i < numOpt-2; i++ {
		op := &vpb.ModifyLiveRoomStateRequest_PollingOption{
			Answer:  idutil.ULIDNow(),
			Content: idutil.ULIDNow(),
		}

		Options = append(Options, op)
	}

	for i, v := range Options {
		if i < numCorrectAnswer {
			v.IsCorrect = true
		}
	}

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command: &vpb.ModifyLiveRoomStateRequest_StartPolling{
			StartPolling: &vpb.ModifyLiveRoomStateRequest_PollingOptions{
				Options:  Options,
				Question: "Can I ask you a question?",
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStartPollingInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	Options := []*vpb.ModifyLiveRoomStateRequest_PollingOption{
		{
			Answer:    "A",
			IsCorrect: true,
			Content:   "Runeterra",
		},
		{
			Answer:    "B",
			IsCorrect: false,
			Content:   "",
		},
		{
			Answer:    "C",
			IsCorrect: false,
			Content:   "Kingdom",
		},
	}

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command: &vpb.ModifyLiveRoomStateRequest_StartPolling{
			StartPolling: &vpb.ModifyLiveRoomStateRequest_PollingOptions{
				Options:  Options,
				Question: "Can I ask you a question?",
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStopPollingInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command:   &vpb.ModifyLiveRoomStateRequest_StopPolling{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userSharingThePollingInTheLiveRoom(ctx context.Context, isStartShared string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isShared := isStartShared == Started

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command: &vpb.ModifyLiveRoomStateRequest_SharePolling{
			SharePolling: isShared,
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userEndPollingInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command:   &vpb.ModifyLiveRoomStateRequest_EndPolling{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userSubmitPollingAnswerInTheLiveRoomPolling(ctx context.Context, answer string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	answers := strings.Split(answer, ",")

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command: &vpb.ModifyLiveRoomStateRequest_SubmitPollingAnswer{
			SubmitPollingAnswer: &vpb.ModifyLiveRoomStateRequest_PollingAnswer{
				StringArrayValue: answers,
			},
		},
	}
	stepState.SubmitPollingAnswer = answers

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentPollingStateInTheLiveRoomHasStarted(ctx context.Context) (context.Context, error) {
	return s.userGetCurrentPollingStateInTheLiveRoomState(ctx, Started)
}

func (s *suite) userGetCurrentPollingStateInTheLiveRoomHasStopped(ctx context.Context) (context.Context, error) {
	return s.userGetCurrentPollingStateInTheLiveRoomState(ctx, Stopped)
}

func (s *suite) userGetCurrentPollingStateInTheLiveRoomIsEmpty(ctx context.Context) (context.Context, error) {
	return s.userGetCurrentPollingStateInTheLiveRoomState(ctx, Empty)
}

func (s *suite) userGetCurrentPollingStateInTheLiveRoomState(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch state {
	case Started:
		if res.CurrentPolling == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling but got empty")
		}

		if res.CurrentPolling.Status != vpb.PollingState_POLLING_STATE_STARTED {
			return StepStateToContext(ctx, stepState), fmt.Errorf("live room polling is not started but found %s", res.CurrentPolling.Status)
		}

		req := stepState.Request.(*vpb.ModifyLiveRoomStateRequest).GetStartPolling()

		if res.CurrentPolling.Question != req.Question {
			return StepStateToContext(ctx, stepState), fmt.Errorf("the question in the request is different what was received")
		}

		options := req.GetOptions()
		for i, v := range res.CurrentPolling.Options {
			if v.Answer != options[i].Answer {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the answer of response must be same")
			}
			if v.Content != options[i].Content {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the content of response must be same")
			}
			if v.IsCorrect != options[i].IsCorrect {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the IsCorrect of response must be same")
			}
		}
	case Stopped:
		if res.CurrentPolling == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling but got empty")
		}

		if res.CurrentPolling.Status != vpb.PollingState_POLLING_STATE_STOPPED {
			return StepStateToContext(ctx, stepState), fmt.Errorf("live room polling is not stopped but found %s", res.CurrentPolling.Status)
		}
	case Empty:
		if res.CurrentPolling != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling is not empty")
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("live room polling state step is not supported")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetPollingAnswerStateInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expectedAnswer := stepState.SubmitPollingAnswer
	actualLearner := sliceutils.Filter(res.UsersState.Learners, func(learner *vpb.GetLiveLessonStateResponse_UsersState_LearnerState) bool {
		return (learner.UserId == stepState.CurrentStudentID)
	})

	if len(actualLearner) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected 1 info for learner id %s but got %v", stepState.CurrentStudentID, len(actualLearner))
	}

	if !reflect.DeepEqual(actualLearner[0].PollingAnswer.StringArrayValue, expectedAnswer) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected polling answer %v but got %v", expectedAnswer, actualLearner[0].PollingAnswer.StringArrayValue)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentPollingStateOfTheLiveRoomContainingSharePolling(ctx context.Context, isStartShared string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	isShared := isStartShared == Started
	if res.CurrentPolling.IsShared.IsShared != isShared {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the current polling sharing must be %s", isStartShared)
	}

	return StepStateToContext(ctx, stepState), nil
}
