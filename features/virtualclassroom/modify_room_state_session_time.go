package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userModifiesTheSessionTimeInTheLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_UpsertSessionTime{
			UpsertSessionTime: true,
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsSessionTimeInTheLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return nil, stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return nil, err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.LessonId); err != nil {
		return nil, err
	}

	if res.CurrentTime.AsTime().IsZero() {
		return nil, fmt.Errorf("expected lesson's current time but got empty")
	}

	// session time should not be empty or not before the current time (with 5 min buffer)
	now := time.Now().Add(-5 * time.Minute)
	if res.SessionTime.AsTime().IsZero() || res.SessionTime.AsTime().Before(now) {
		return nil, fmt.Errorf("expected session time but got empty or session time is in the past")
	}

	return StepStateToContext(ctx, stepState), nil
}
