package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) UserEnableAnnotationInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_AnnotationEnable{
			AnnotationEnable: &bpb.ModifyLiveLessonStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetAnnotationState(ctx context.Context) (context.Context, error) {
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

	expectedAnnotation := false
	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
	switch req.Command.(type) {
	case *bpb.ModifyLiveLessonStateRequest_AnnotationEnable:
		expectedAnnotation = true
	case *bpb.ModifyLiveLessonStateRequest_AnnotationDisable:
	case *bpb.ModifyLiveLessonStateRequest_StopSharingMaterial:
		expectedAnnotation = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("before request %T is invalid command", req.Command)
	}
	actualLearnerSt := res.UsersState.Learners[0]
	if actualLearnerSt.Annotation.Value != expectedAnnotation {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's hands up state %v but got %v", expectedAnnotation, actualLearnerSt.Annotation.Value)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserDisableAnnotationInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_AnnotationDisable{
			AnnotationDisable: &bpb.ModifyLiveLessonStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
