package bob

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) userCreateBrightcoveUploadUrlForVideo(ctx context.Context, videoName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.CreateBrightCoveUploadUrlRequest{
		Name: videoName,
	}

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).CreateBrightCoveUploadUrl(
		contextWithToken(s, ctx),
		req,
	)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnAVideoUploadUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.CreateBrightCoveUploadUrlResponse)

	if resp.ApiRequestUrl == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("apiRequestUrl should not empty")
	}

	if resp.SignedUrl == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("signedUrl should not empty")
	}

	return StepStateToContext(ctx, stepState), nil
}
