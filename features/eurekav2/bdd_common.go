package eurekav2

import (
	"context"
	"fmt"

	"google.golang.org/grpc/status"
)

func (s *suite) checkStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}
