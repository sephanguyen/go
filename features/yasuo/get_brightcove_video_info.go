package yasuo

import (
	"context"
	"fmt"
	"time"

	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *suite) userGetsInfoOfAVideo(ctx context.Context, videoType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &ypb.GetBrightCoveVideoInfoRequest{
		AccountId: "account-id",
		VideoId:   "invalid-video-id",
	}
	if videoType == "valid" {
		req.VideoId = "video-id"
	}
	if videoType == "not_playable" {
		req.VideoId = "video_not_playable"
	}

	stepState.Response, stepState.ResponseErr = ypb.NewBrightcoveServiceClient(s.Conn).GetBrightcoveVideoInfo(contextWithToken(s, ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theCorrectInfoOfTheVideoIsReturned(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, ok := stepState.Response.(*ypb.GetBrightCoveVideoInfoResponse)
	if !ok {
		return ctx, fmt.Errorf("expected *ypb.GetBrightCoveVideoInfoResponse type for response, got %T", stepState.Response)
	}

	expectedResp := &ypb.GetBrightCoveVideoInfoResponse{
		Id:             "video-id",
		Name:           "video-name",
		Thumbnail:      "https://link/to/some/image.jpg",
		Duration:       durationpb.New(time.Millisecond * 1234),
		OfflineEnabled: true,
	}

	if resp.Id != expectedResp.Id {
		return ctx, fmt.Errorf("expected %q for ID, got %q", expectedResp.Id, resp.Id)
	}
	if resp.Name != expectedResp.Name {
		return ctx, fmt.Errorf("expected %q for Name, got %q", expectedResp.Name, resp.Name)
	}
	if resp.Thumbnail != expectedResp.Thumbnail {
		return ctx, fmt.Errorf("expected %q for Thumbnail, got %q", expectedResp.Thumbnail, resp.Thumbnail)
	}
	if resp.Duration.AsDuration() != expectedResp.Duration.AsDuration() {
		return ctx, fmt.Errorf("expected %q for Name, got %q", expectedResp.Duration, resp.Duration)
	}
	if resp.OfflineEnabled != expectedResp.OfflineEnabled {
		return ctx, fmt.Errorf("expected %v for Name, got %v", expectedResp.OfflineEnabled, resp.OfflineEnabled)
	}

	return StepStateToContext(ctx, stepState), nil
}
