package yasuo

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
)

func (s *suite) userFinishBrightcoveUploadUrlForVideo(ctx context.Context, videoName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.FinishUploadBrightCoveRequest{}
	if videoName != "" && stepState.Request.(*pb.CreateBrightCoveUploadUrlRequest).Name == videoName {
		uploadResp := stepState.Response.(*pb.CreateBrightCoveUploadUrlResponse)

		req = &pb.FinishUploadBrightCoveRequest{
			ApiRequestUrl: uploadResp.ApiRequestUrl,
			VideoId:       uploadResp.VideoId,
		}
	}

	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.Conn).FinishUploadBrightCove(s.signedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
