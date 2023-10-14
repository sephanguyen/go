package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userZoomWhiteboardInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	w := &domain.WhiteboardZoomState{
		PdfScaleRatio: 23.32,
		CenterX:       243.5,
		CenterY:       -432.034,
		PdfWidth:      234.43,
		PdfHeight:     -0.33424,
	}
	stepState.CurrentWhiteboardZoomState = w

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_WhiteboardZoomState_{
			WhiteboardZoomState: &vpb.ModifyVirtualClassroomStateRequest_WhiteboardZoomState{
				PdfScaleRatio: w.PdfScaleRatio,
				CenterX:       w.CenterX,
				CenterY:       w.CenterY,
				PdfWidth:      w.PdfWidth,
				PdfHeight:     w.PdfHeight,
			},
		},
	}
	stepState.Request = req
	return s.userModifyWhiteboardState(StepStateToContext(ctx, stepState), req)
}

func (s *suite) userModifyWhiteboardState(ctx context.Context, req *vpb.ModifyVirtualClassroomStateRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetWhiteboardZoomState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.LessonId); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	b := res.WhiteboardZoomState
	d := stepState.CurrentWhiteboardZoomState

	if d.CenterX != b.CenterX || d.CenterY != b.CenterY || d.PdfHeight != b.PdfHeight || d.PdfWidth != b.PdfWidth || d.PdfScaleRatio != b.PdfScaleRatio {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected whiteboard zoom state %v but got %v", d, b)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetWhiteboardZoomStateWithDefaultValue(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentWhiteboardZoomState = new(domain.WhiteboardZoomState).SetDefault()

	return s.userGetWhiteboardZoomState(ctx)
}
