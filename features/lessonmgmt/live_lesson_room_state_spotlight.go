package lessonmgmt

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) enableSpotlight(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	students := stepState.StudentIds
	var userID string
	if len(students) > 0 {
		userID = students[0]
	}
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_Spotlight_{
			Spotlight: &bpb.ModifyLiveLessonStateRequest_Spotlight{
				UserId:      userID,
				IsSpotlight: true,
			},
		},
	}
	stepState.CurrentStudentID = userID
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Connections.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) disableSpotlight(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_Spotlight_{
			Spotlight: &bpb.ModifyLiveLessonStateRequest_Spotlight{
				IsSpotlight: false,
			},
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Connections.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}
