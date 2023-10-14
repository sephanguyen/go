package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) UserRequestRecordingLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_RequestRecording{
			RequestRecording: true,
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetCurrentRecordingLiveLessonPermissionToStartRecording(ctx context.Context) (context.Context, error) {
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

	if res.Recording == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson have recording status but got nil")
	}
	if !res.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected live lesson is recording but it's not")
	}
	if res.Recording.Creator != stepState.CurrentUserID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected user %s get recording permission but creator is %s", stepState.CurrentUserID, res.Recording.Creator)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userHaveNoCurrentRecordingLiveLessonPermissionToStartRecording(ctx context.Context) (context.Context, error) {
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

	if res.Recording != nil && res.Recording.Creator == stepState.CurrentUserID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected user %s have no recording permission but got it", stepState.CurrentUserID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserStopRecordingLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_StopRecording{
			StopRecording: true,
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) liveLessonIsStillRecording(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.CommonSuite.ASignedInAdmin(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Recording == nil || !res.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected live lesson %s is still recording but it don't", stepState.CurrentLessonID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) liveLessonIsNotRecording(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.CommonSuite.ASignedInAdmin(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Recording != nil && res.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected live lesson %s is not recording but it do", stepState.CurrentLessonID)
	}

	return StepStateToContext(ctx, stepState), nil
}
