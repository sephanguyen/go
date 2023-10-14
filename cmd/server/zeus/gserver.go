package zeus

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/zeus/configurations"
	"github.com/manabie-com/backend/internal/zeus/repositories"
	"github.com/manabie-com/backend/internal/zeus/services"
	"github.com/manabie-com/backend/internal/zeus/subscriptions"
	pb "github.com/manabie-com/backend/pkg/manabuf/zeus/v1"

	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithNatsServicer(s).
		Register(s)
}

type server struct{}

func (*server) RegisterNatsSubscribers(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	lg := rsc.Logger()
	db := rsc.DB()
	jsm := rsc.NATS()

	centralizeLogsService := &services.CentralizeLogsService{
		DB:              db,
		ActivityLogRepo: &repositories.ActivityLogRepo{},
	}

	activityLogSubscriber := &subscriptions.ActivityLogCreatedEventSubscriber{
		CentralizeLogsService: centralizeLogsService,
		JSM:                   jsm,
		Logger:                lg,
		Configs:               &c,
	}

	// if err := activityLogSubscriber.Subscribe(); err != nil {
	// 	return fmt.Errorf("activityLogSubscriber.Subscribe: %w", err)
	// }

	if err := activityLogSubscriber.PullConsumer(); err != nil {
		return fmt.Errorf("activityLogSubscriber.PullConsumer: %w", err)
	}

	return nil
}

func (*server) ServerName() string {
	return "zeus"
}

func (*server) SetupGRPC(ctx context.Context, s *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()
	centralizeLogsService := &services.CentralizeLogsService{
		DB:              db,
		ActivityLogRepo: &repositories.ActivityLogRepo{},
	}
	pb.RegisterCentralizeLogsServiceServer(s, centralizeLogsService)
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
