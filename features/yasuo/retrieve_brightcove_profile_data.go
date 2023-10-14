package yasuo

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func (s *suite) getBrightcoveProfileData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.RetrieveBrightCoveProfileDataRequest{}
	stepState.Response, stepState.ResponseErr = pb.NewBrightcoveServiceClient(s.Conn).RetrieveBrightCoveProfileData(s.signedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
