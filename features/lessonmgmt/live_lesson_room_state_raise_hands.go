package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) UserRaiseHandInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_RaiseHand{},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetHandsUpState(ctx context.Context) (context.Context, error) {
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

	if len(res.UsersState.Learners) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected get 1 learner's state but got %d", len(res.UsersState.Learners))
	}

	expectedHandsUp := false
	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
	switch req.Command.(type) {
	case *bpb.ModifyLiveLessonStateRequest_RaiseHand:
		expectedHandsUp = true
	case *bpb.ModifyLiveLessonStateRequest_HandOff:
		expectedHandsUp = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("before request %T is invalid command", req.Command)
	}

	actualLearnerSt := res.UsersState.Learners[0]
	if actualLearnerSt.UserId != stepState.CurrentStudentID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner id %s but got %s", stepState.CurrentStudentID, actualLearnerSt.UserId)
	}

	if actualLearnerSt.HandsUp.Value != expectedHandsUp {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's hands up state %v but got %v", expectedHandsUp, actualLearnerSt.HandsUp.Value)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userFoldALearnersHandInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_FoldUserHand{
			FoldUserHand: stepState.StudentIds[0],
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetAllLearnersHandsUpStatesWhoAllHaveValueIsOff(ctx context.Context) (context.Context, error) {
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

	for _, learner := range res.UsersState.Learners {
		if learner.HandsUp.Value {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected all learner's hands up is off but %s is not", learner.UserId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserFoldHandAllLearner(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_FoldHandAll{},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserHandOffInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_HandOff{},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
