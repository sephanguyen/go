package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userGetsChatConfigForZegoCloud(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetChatConfigRequest{}
	stepState.Response, stepState.ResponseErr = vpb.NewZegoCloudServiceClient(s.VirtualClassroomConn).
		GetChatConfig(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesChatConfiguration(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetChatConfigResponse)

	if response.GetAppId() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected app id but received empty")
	}

	if len(response.GetAppSign()) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected app sign but received empty")
	}

	return StepStateToContext(ctx, stepState), nil
}
