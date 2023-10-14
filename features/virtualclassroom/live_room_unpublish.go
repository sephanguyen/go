package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userUnpublishInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.UnpublishLiveRoomRequest{
		ChannelId: stepState.CurrentChannelID,
		LearnerId: stepState.CurrentStudentID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		UnpublishLiveRoom(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsUnpublishInTheLiveRoom(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.UnpublishLiveRoomResponse)
	var expectedStatus vpb.UnpublishStatus
	switch status {
	case StatusNone:
		expectedStatus = vpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_NONE
	case "unpublished before":
		expectedStatus = vpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_BEFORE
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported expected status")
	}

	if response.Status != expectedStatus {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected status %s does not match with actual status %s", expectedStatus.String(), response.Status.String())
	}

	return StepStateToContext(ctx, stepState), nil
}
