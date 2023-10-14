package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// HealthcheckService implement gRPC Healthcheck standard
type HealthcheckService struct {
	DB database.Ext
}

// Check will ensure migration already success
func (s *HealthcheckService) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	resp := &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_NOT_SERVING,
	}

	logger := ctxzap.Extract(ctx)

	var selectResult int
	row := s.DB.QueryRow(ctx, "SELECT 1")
	err := row.Scan(&selectResult)
	if err != nil {
		logger.Error("HealthcheckService.Check selecting from DB", zap.Error(err))
		return resp, nil
	}

	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

// Watch only dummy implemented
func (*HealthcheckService) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, codes.Unimplemented.String())
}
