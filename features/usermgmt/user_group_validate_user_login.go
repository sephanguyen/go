package usermgmt

import (
	"context"
	"fmt"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

var responseMapStatus = map[string]bool{
	"able":   true,
	"unable": false,
}

func (s *suite) genValidationPayload(platformName string) *pb.ValidateUserLoginRequest {
	return &pb.ValidateUserLoginRequest{
		Platform: cpb.Platform(cpb.Platform_value[platformName]),
	}
}

func (s *suite) checkThisUserIsAbleToAccessPlatform(ctx context.Context, platform string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := s.genValidationPayload(platform)
	resp, err := pb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).ValidateUserLogin(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to call validate service: %w", stepState.ResponseErr)
	}

	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userHasPermissionToAccessPlatform(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.ValidateUserLoginResponse)

	// check correct response
	if responseMapStatus[status] != resp.Allowable {
		return StepStateToContext(ctx, stepState), fmt.Errorf("status not match expected %t, got %t", responseMapStatus[status], resp.Allowable)
	}

	return StepStateToContext(ctx, stepState), nil
}
