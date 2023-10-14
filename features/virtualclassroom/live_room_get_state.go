package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) GetCurrentStateOfLiveRoom(ctx context.Context, channelID string) (*vpb.GetLiveRoomStateResponse, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)

	req := &vpb.GetLiveRoomStateRequest{
		ChannelId: channelID,
	}

	res, err := vpb.NewLiveRoomReaderServiceClient(s.VirtualClassroomConn).
		GetLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return nil, fmt.Errorf("failed to get live room state, channel %s: %s", channelID, err)
	}
	if res.ChannelId != channelID {
		return nil, fmt.Errorf("expected live room state for channel %s but got %s", channelID, res.ChannelId)
	}
	if res.CurrentTime.AsTime().IsZero() {
		return nil, fmt.Errorf("expected live room's current time but got empty")
	}

	return res, nil
}

func (s *suite) userGetsLiveRoomState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomReaderServiceClient(s.VirtualClassroomConn).
		GetLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.GetLiveRoomStateRequest{
			ChannelId: stepState.CurrentChannelID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) liveRoomStateIsInDefaultEmptyState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetLiveRoomStateResponse)

	if len(response.ChannelId) == 0 || response.ChannelId != stepState.CurrentChannelID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect channel id (%s) when get live room state, expecting %s", response.ChannelId, stepState.CurrentChannelID)
	}

	if response.CurrentTime.AsTime().IsZero() {
		return nil, fmt.Errorf("expected current time but got empty")
	}

	// current material
	if response.CurrentMaterial != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current material should be empty by default")
	}

	// current polling
	if response.CurrentPolling != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current polling should be empty by default")
	}

	// spotlighted user
	if response.Spotlight.GetIsSpotlight() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("is spotlight should be false by default")
	}
	if len(response.Spotlight.GetUserId()) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("spotlighted user should be empty by default")
	}

	// recording
	if response.Recording.GetIsRecording() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("is recording should be false by default")
	}
	if len(response.Recording.GetCreator()) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("recording creator should be empty by default")
	}

	// whiteboard zoom state
	wbZoomState := (&domain.WhiteboardZoomState{}).SetDefault()
	actualWBZoomState := response.GetWhiteboardZoomState()
	if wbZoomState.CenterX != actualWBZoomState.CenterX || wbZoomState.CenterY != actualWBZoomState.CenterY ||
		wbZoomState.PdfHeight != actualWBZoomState.PdfHeight || wbZoomState.PdfWidth != actualWBZoomState.PdfWidth ||
		wbZoomState.PdfScaleRatio != actualWBZoomState.PdfScaleRatio {
		return StepStateToContext(ctx, stepState), fmt.Errorf("one or more of the whiteboard zoom state is not in its default values")
	}

	return StepStateToContext(ctx, stepState), nil
}
