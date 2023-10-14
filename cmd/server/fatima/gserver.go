package fatima

import (
	"context"
	"fmt"

	eureka_interceptors "github.com/manabie-com/backend/internal/eureka/golibs/interceptors"
	"github.com/manabie-com/backend/internal/fatima/configurations"
	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/fatima/services"
	"github.com/manabie-com/backend/internal/fatima/subscriptions"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/debezium"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithMonitorServicer(s).
		WithNatsServicer(s).
		Register(s)
}

type server struct {
	bootstrap.DefaultMonitorService[configurations.Config]

	userMgmtConn    *grpc.ClientConn
	authInterceptor *interceptors.Auth

	accessibilityReadSvc   *services.AccessibilityReadService
	courseReaderSvc        *services.CourseReaderService
	accessibilityModifySvc *services.AccessibilityModifyService
	subscriptionSvc        *services.SubscriptionServiceABAC
}

func (*server) ServerName() string {
	return "fatima"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		eureka_interceptors.UpdateUserIDForParent,
	}
	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	grpcUnary = append(grpcUnary, customs...)
	return grpcUnary
}

func (s *server) WithStreamServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	grpcStream := bootstrap.DefaultStreamServerInterceptor(rsc)
	grpcStream = append(grpcStream, s.authInterceptor.StreamServerInterceptor)

	return grpcStream
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()
	grpc_zap.ReplaceGrpcLoggerV2(rsc.Logger())
	s.authInterceptor = authInterceptor(&c, rsc.Logger(), db.DB)

	s.userMgmtConn = rsc.GRPCDial("usermgmt")

	userReaderClient := upb.NewUserReaderServiceClient(s.userMgmtConn)

	s.accessibilityReadSvc = &services.AccessibilityReadService{
		DB:                 db,
		StudentPackageRepo: &repositories.StudentPackageRepo{},
	}

	s.courseReaderSvc = &services.CourseReaderService{
		DB:                           db,
		StudentPackageAccessPathRepo: &repositories.StudentPackageAccessPathRepo{},
		UserMgmtUserReader:           userReaderClient,
	}

	s.accessibilityModifySvc = &services.AccessibilityModifyService{
		DB:                           db,
		StudentPackageRepo:           &repositories.StudentPackageRepo{},
		StudentPackageAccessPathRepo: &repositories.StudentPackageAccessPathRepo{},
		JSM:                          rsc.NATS(),
	}

	s.subscriptionSvc = &services.SubscriptionServiceABAC{
		SubscriptionModifyService: &services.SubscriptionModifyService{
			DB:                           db,
			PackageRepo:                  &repositories.PackageRepo{},
			StudentPackageRepo:           &repositories.StudentPackageRepo{},
			StudentPackageAccessPathRepo: &repositories.StudentPackageAccessPathRepo{},
			StudentPackageClassRepo:      &repositories.StudentPackageClassRepo{},
			JSM:                          rsc.NATS(),
		},
	}

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcServer *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	fpb.RegisterAccessibilityReadServiceServer(grpcServer, s.accessibilityReadSvc)
	fpb.RegisterCourseReaderServiceServer(grpcServer, s.courseReaderSvc)
	fpb.RegisterSubscriptionModifierServiceServer(grpcServer, s.subscriptionSvc)
	health.RegisterHealthServer(grpcServer, &healthcheck.Service{
		DB: rsc.DB().DB.(*pgxpool.Pool),
	})

	return nil
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	jprepStudentPackage := &subscriptions.JprepStudentPackage{
		JSM:           rsc.NATS(),
		Logger:        rsc.Logger(),
		CourseService: s.accessibilityModifySvc,
	}
	err := jprepStudentPackage.Subscribe()
	if err != nil {
		return fmt.Errorf("RegisterNatsSubscribers: jprepSyncStudentPackage.Subscribe: %w", err)
	}

	// as source database, it will listen to incremental snapshot events which will trigger new captured table
	err = debezium.InitDebeziumIncrementalSnapshot(rsc.NATS(), rsc.Logger(), rsc.DB(), c.Common.Name)
	if err != nil {
		return fmt.Errorf("initInternalDebeziumIncrementalSnapshot: %v", err)
	}
	return nil
}

func (s *server) GracefulShutdown(ctx context.Context) {
	s.userMgmtConn.Close()
}
