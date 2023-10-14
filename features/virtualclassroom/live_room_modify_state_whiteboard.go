package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userZoomWhiteboardInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	w := &vc_domain.WhiteboardZoomState{
		PdfScaleRatio: 23.32,
		CenterX:       243.5,
		CenterY:       -432.034,
		PdfWidth:      234.43,
		PdfHeight:     -0.33424,
	}
	stepState.CurrentWhiteboardZoomState = w

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
		Command: &vpb.ModifyLiveRoomStateRequest_WhiteboardZoomState_{
			WhiteboardZoomState: &vpb.ModifyLiveRoomStateRequest_WhiteboardZoomState{
				PdfScaleRatio: w.PdfScaleRatio,
				CenterX:       w.CenterX,
				CenterY:       w.CenterY,
				PdfWidth:      w.PdfWidth,
				PdfHeight:     w.PdfHeight,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsWhiteboardZoomStateInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	b := res.WhiteboardZoomState
	d := stepState.CurrentWhiteboardZoomState

	if d.CenterX != b.CenterX || d.CenterY != b.CenterY || d.PdfHeight != b.PdfHeight || d.PdfWidth != b.PdfWidth || d.PdfScaleRatio != b.PdfScaleRatio {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected whiteboard zoom state %v but got %v", d, b)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsWhiteboardZoomStateDefaultInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentWhiteboardZoomState = new(vc_domain.WhiteboardZoomState).SetDefault()

	return s.userGetsWhiteboardZoomStateInTheLiveRoom(ctx)
}
