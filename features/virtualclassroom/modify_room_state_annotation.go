package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) UserEnableAnnotationInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationEnable{
			AnnotationEnable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetAnnotationState(ctx context.Context) (context.Context, error) {
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

	expectedAnnotation := false
	req := stepState.Request.(*vpb.ModifyVirtualClassroomStateRequest)
	switch req.Command.(type) {
	case *vpb.ModifyVirtualClassroomStateRequest_AnnotationEnable:
		expectedAnnotation = true
	case *vpb.ModifyVirtualClassroomStateRequest_AnnotationDisable:
		expectedAnnotation = false
	case *vpb.ModifyVirtualClassroomStateRequest_StopSharingMaterial:
		expectedAnnotation = true
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

func (s *suite) UserDisableAnnotationInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationDisable{
			AnnotationDisable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
				Learners: stepState.StudentIds,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserDisableAllAnnotationInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationDisableAll{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AllAnnotationStateIsDisable(ctx context.Context) (context.Context, error) {
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

	actualLearnerSt := res.UsersState.Learners
	for _, l := range actualLearnerSt {
		if l.Annotation.Value {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's annotation state %v but got %v", false, l.Annotation.Value)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
