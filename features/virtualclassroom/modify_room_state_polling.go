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
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) UserStartPollingInVirtualClassroom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	Options := []*vpb.ModifyVirtualClassroomStateRequest_PollingOption{
		{
			Answer:    "A",
			IsCorrect: true,
			Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
		},
		{
			Answer:    "B",
			IsCorrect: false,
			Content:   "",
		},
		{
			Answer:    "C",
			IsCorrect: false,
			Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
		},
	}
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_StartPolling{
			StartPolling: &vpb.ModifyVirtualClassroomStateRequest_PollingOptions{
				Options:  Options,
				Question: "What is the question?",
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserStartPollingInVirtualClassroomWithNumOption(ctx context.Context, numOptStr string, numCorrectAnswerStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOpt, _ := strconv.Atoi(numOptStr)
	numCorrectAnswer, _ := strconv.Atoi(numCorrectAnswerStr)
	Options := []*vpb.ModifyVirtualClassroomStateRequest_PollingOption{
		{
			Answer:  "A",
			Content: "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
		},
		{
			Answer:  "B",
			Content: "",
		},
	}
	for i := 0; i < numOpt-2; i++ {
		op := &vpb.ModifyVirtualClassroomStateRequest_PollingOption{
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

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_StartPolling{
			StartPolling: &vpb.ModifyVirtualClassroomStateRequest_PollingOptions{
				Options:  Options,
				Question: "What is the question?",
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentPollingStateOfVirtualClassroomStarted(ctx context.Context) (context.Context, error) {
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

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected virtual classroom's current time but got empty")
	}
	if res.CurrentPolling == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling but got empty")
	}

	if res.CurrentPolling.Status != vpb.PollingState_POLLING_STATE_STARTED {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current polling is not start")
	}

	req := stepState.Request.(*vpb.ModifyVirtualClassroomStateRequest).GetStartPolling()
	if res.CurrentPolling.Question != req.Question {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the question of response must be same")
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

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserSubmitPollingAnswerInVirtualClassroom(ctx context.Context, answer string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	answers := strings.Split(answer, ",")
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_SubmitPollingAnswer{
			SubmitPollingAnswer: &vpb.ModifyVirtualClassroomStateRequest_PollingAnswer{
				StringArrayValue: answers,
			},
		},
	}
	stepState.SubmitPollingAnswer = answers

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetPollingAnswerState(ctx context.Context) (context.Context, error) {
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

func (s *suite) UserStopPollingInVirtualClassroom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_StopPolling{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentPollingStateOfVirtualClassroomStopped(ctx context.Context) (context.Context, error) {
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
	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected virtual classroom's current time but got empty")
	}
	if res.CurrentPolling == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling but got empty")
	}
	if res.CurrentPolling.Status != vpb.PollingState_POLLING_STATE_STOPPED {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current polling is not stop")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserEndPollingInVirtualClassroom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_EndPolling{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentPollingStateOfVirtualClassroomIsEmpty(ctx context.Context) (context.Context, error) {
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

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected virtual classroom's current time but got empty")
	}
	if res.CurrentPolling != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current polling is not empty")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserSharePollingInVirtualClassroom(ctx context.Context, isStartShared string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isShared := isStartShared == "start"

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_SharePolling{
			SharePolling: isShared,
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentPollingStateOfVirtualClassroomSharePolling(ctx context.Context, isStartShared string) (context.Context, error) {
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

	isShared := isStartShared == "start"
	if res.CurrentPolling.IsShared.IsShared != isShared {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the current polling sharing must be %s", isStartShared)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userEndTheLiveLessonInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.EndLiveLessonRequest{
		LessonId: stepState.CurrentLessonID,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.BobConn).EndLiveLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}
