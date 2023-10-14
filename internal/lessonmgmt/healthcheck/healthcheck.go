package healthcheck

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Service implement gRPC HealthCheck standard
type Service struct {
	DBBob    Pinger
	DBLesson Pinger
}
type Pinger interface {
	Ping(context.Context) error
	Stat() *pgxpool.Stat
}

func (s *Service) Check(ctx context.Context, _ *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	logger := ctxzap.Extract(ctx)

	err := s.DBBob.Ping(ctx)
	if err != nil {
		logger.Error("HealthcheckService.Check Ping", zap.Error(err))

		stat := s.DBBob.Stat()

		maxConns := stat.MaxConns()
		idleConns := stat.IdleConns()
		activeConns := stat.AcquiredConns() - idleConns

		if maxConns > 0 && activeConns == maxConns {
			logger.Warn("HealthcheckService.Check Maximum number of connections reached", zap.Int32("idleConns", idleConns), zap.Int32("activeConns", activeConns), zap.Int32("maxConns", maxConns))
		}

		return &health.HealthCheckResponse{
			Status: health.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	err = s.DBLesson.Ping(ctx)
	if err != nil {
		logger.Error("HealthcheckService.Check Ping", zap.Error(err))

		stat := s.DBLesson.Stat()

		maxConns := stat.MaxConns()
		idleConns := stat.IdleConns()
		activeConns := stat.AcquiredConns() - idleConns

		if maxConns > 0 && activeConns == maxConns {
			logger.Warn("HealthcheckService.Check Maximum number of connections reached", zap.Int32("idleConns", idleConns), zap.Int32("activeConns", activeConns), zap.Int32("maxConns", maxConns))
		}

		return &health.HealthCheckResponse{
			Status: health.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

// Watch only dummy implemented
func (*Service) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, codes.Unimplemented.String())
}
