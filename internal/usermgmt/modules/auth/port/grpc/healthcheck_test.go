package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func TestHealthCheckService_Check(t *testing.T) {
	s := HealthcheckService{}
	req := &health.HealthCheckRequest{
		Service: "auth",
	}

	res, err := s.Check(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, health.HealthCheckResponse_SERVING, res.Status)
}
