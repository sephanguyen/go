package lessonmgmt

import (
	"context"
	"fmt"

	health "google.golang.org/grpc/health/grpc_health_v1"
)

func (s *Suite) healthCheckEndpointCalled(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = health.NewHealthClient(s.LessonMgmtConn).Check(s.CommonSuite.ContextWithValidVersion(ctx), &health.HealthCheckRequest{})
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) everythingIsOK(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonMgmtShouldReturnWithStatus(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CommonSuite.ReturnsStatusCode(ctx, arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp := stepState.Response.(*health.HealthCheckResponse)
	if resp.GetStatus().String() != arg2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected service status: %s", resp.GetStatus())
	}

	return StepStateToContext(ctx, stepState), nil
}
