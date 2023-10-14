package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) UserRaiseHandInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_RaiseHand{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetHandsUpState(ctx context.Context) (context.Context, error) {
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

	if len(res.UsersState.Learners) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected get at least 1 learner's state but got 0")
	}

	expectedHandsUp := false
	req := stepState.Request.(*vpb.ModifyVirtualClassroomStateRequest)
	switch req.Command.(type) {
	case *vpb.ModifyVirtualClassroomStateRequest_RaiseHand:
		expectedHandsUp = true
	case *vpb.ModifyVirtualClassroomStateRequest_HandOff:
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

func (s *suite) userFoldALearnersHandInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_FoldUserHand{
			FoldUserHand: stepState.StudentIds[0],
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetAllLearnersHandsUpStatesWhoAllHaveValueIsOff(ctx context.Context) (context.Context, error) {
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

	for _, learner := range res.UsersState.Learners {
		if learner.HandsUp.Value {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected all learner's hands up is off but %s is not", learner.UserId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserFoldHandAllLearner(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_FoldHandAll{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserHandOffInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_HandOff{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
