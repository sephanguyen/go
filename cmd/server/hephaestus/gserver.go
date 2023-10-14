package hephaestus

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/hephaestus/configurations"
	"github.com/manabie-com/backend/internal/hephaestus/services"

	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	s := &server{}
	bootstrap.WithGRPC[configurations.Config](s).Register(s)
}

type server struct{}

func (*server) ServerName() string {
	return "hephaestus"
}

func (*server) SetupGRPC(_ context.Context, s *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	health.RegisterHealthServer(s, &services.HealthCheckService{})
	return nil
}

func (*server) InitDependencies(_ configurations.Config, _ *bootstrap.Resources) error { return nil }
func (*server) GracefulShutdown(context.Context)                                       {}
func (*server) WithUnaryServerInterceptors(_ configurations.Config, _ *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	return nil
}

func (*server) WithStreamServerInterceptors(_ configurations.Config, _ *bootstrap.Resources) []grpc.StreamServerInterceptor {
	return nil
}
func (*server) WithServerOptions() []grpc.ServerOption { return nil }
