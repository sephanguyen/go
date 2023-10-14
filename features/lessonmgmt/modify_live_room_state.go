package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) GetCurrentStateOfLiveLessonRoom(ctx context.Context, lessonID string) (*bpb.LiveLessonStateResponse, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.LiveLessonStateRequest{Id: lessonID}
	res, err := bpb.NewLessonReaderServiceClient(s.Connections.BobConn).
		GetLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Suite) userSignedAsStudentWhoBelongToLesson(ctx context.Context) (context.Context, error) {
	return s.CommonSuite.ASignedInStudentInStudentList(ctx)
}

func (s *Suite) EndOneOfTheLiveLessonV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.EndLiveLessonRequest{
		LessonId: stepState.CurrentLessonID,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.BobConn).
		EndLiveLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetSpotlightedUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
	if !req.GetSpotlight().IsSpotlight {
		if res.Spotlight.UserId != "" && res.Spotlight.IsSpotlight {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected empty user but got %s", res.Spotlight.UserId)
		}
	} else {
		if res.Spotlight.UserId != stepState.CurrentStudentID && !res.Spotlight.IsSpotlight {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected user %s but got %s", stepState.CurrentStudentID, res.Spotlight.UserId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
