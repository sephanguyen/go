package bob

import (
	"context"
	"fmt"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *suite) retrieveWhiteboardToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.Conn).RetrieveWhiteboardToken(contextWithToken(s, ctx), &bpb.RetrieveWhiteboardTokenRequest{
		LessonId: stepState.CurrentLessonID,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receiveWhiteboardToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*bpb.RetrieveWhiteboardTokenResponse)
	if resp.RoomId == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing room id when student join live lesson")
	}
	if resp.WhiteboardToken == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing WhiteboardToken")
	}
	if resp.WhiteboardAppId == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing WhiteboardAppId")
	}
	return StepStateToContext(ctx, stepState), nil
}
