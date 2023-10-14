package auth

import (
	"context"
	"fmt"

	health "google.golang.org/grpc/health/grpc_health_v1"
)

func (s *suite) healthCheckEndpointCalled(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = health.NewHealthClient(s.AuthConn).Check(contextWithValidVersion(ctx), &health.HealthCheckRequest{})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) everythingIsOK(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) authServiceShouldReturnWithStatus(ctx context.Context, statusCode, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CommonSuite.ReturnsStatusCode(ctx, statusCode)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp := stepState.Response.(*health.HealthCheckResponse)
	if resp.GetStatus().String() != status {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected service status: %s", resp.GetStatus())
	}

	return StepStateToContext(ctx, stepState), nil
}
