package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userGetsAuthenticationTokenForZegoCloud(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetAuthenticationTokenRequest{}
	if len(stepState.CurrentUserID) != 0 {
		req.UserId = stepState.CurrentUserID
	}

	stepState.Response, stepState.ResponseErr = vpb.NewZegoCloudServiceClient(s.VirtualClassroomConn).
		GetAuthenticationToken(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesAuthenticationToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetAuthenticationTokenResponse)

	if len(response.GetAuthToken()) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected auth token but received empty")
	}

	if response.GetAppId() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected app id but received empty")
	}

	if len(response.GetAppSign()) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected app sign but received empty")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsAuthenticationTokenForZegoCloudUsingV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetAuthenticationTokenV2Request{}
	if len(stepState.CurrentUserID) != 0 {
		req.UserId = stepState.CurrentUserID
	}

	stepState.Response, stepState.ResponseErr = vpb.NewZegoCloudServiceClient(s.VirtualClassroomConn).
		GetAuthenticationTokenV2(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesAuthenticationTokenFromV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetAuthenticationTokenV2Response)

	if len(response.GetAuthToken()) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected auth token but received empty")
	}

	return StepStateToContext(ctx, stepState), nil
}
