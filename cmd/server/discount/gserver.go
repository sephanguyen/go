package discount

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/discount/configurations"
	"github.com/manabie-com/backend/internal/discount/repositories"
	discountService "github.com/manabie-com/backend/internal/discount/services"
	exportService "github.com/manabie-com/backend/internal/discount/services/export_service"
	services "github.com/manabie-com/backend/internal/discount/services/import_services"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	discountPb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

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
		Register(s)
}

type server struct {
	bootstrap.DefaultMonitorService[configurations.Config]

	authInterceptor *interceptors.Auth
	discountService *discountService.DiscountService
	internalService *discountService.InternalService
	exportService   *exportService.ExportService
}

func (s *server) WithUnaryServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()

	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
	}
	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	grpcUnary = append(grpcUnary, customs...)

	return grpcUnary
}

func (s *server) WithStreamServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	grpcStream := bootstrap.DefaultStreamServerInterceptor(rsc)
	grpcStream = append(grpcStream, s.authInterceptor.StreamServerInterceptor)

	return grpcStream
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (*server) ServerName() string {
	return "discount"
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	fatimaDBTrace := rsc.DBWith("fatima")

	s.authInterceptor = authInterceptor(&c, zapLogger, fatimaDBTrace.DB)
	s.discountService = discountService.NewDiscountService(fatimaDBTrace, rsc.NATS(), zapLogger)
	s.internalService = discountService.NewInternalService(fatimaDBTrace, rsc.NATS(), zapLogger, rsc.Kafka())
	s.exportService = exportService.NewExportService(fatimaDBTrace)

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcServer *grpc.Server, _ configurations.Config, rsc *bootstrap.Resources) error {
	fatimaDBTrace := rsc.DBWith("fatima")

	health.RegisterHealthServer(grpcServer, &healthcheck.Service{DB: fatimaDBTrace.DB.(*pgxpool.Pool)})
	discountPb.RegisterDiscountServiceServer(grpcServer, s.discountService)
	discountPb.RegisterInternalServiceServer(grpcServer, s.internalService)

	discountPb.RegisterImportMasterDataServiceServer(grpcServer, &services.ImportMasterDataService{
		UnimplementedImportMasterDataServiceServer: discountPb.UnimplementedImportMasterDataServiceServer{},
		DB:                               fatimaDBTrace,
		DiscountTagRepo:                  &repositories.DiscountTagRepo{},
		ProductGroupRepo:                 &repositories.ProductGroupRepo{},
		ProductGroupMappingRepo:          &repositories.ProductGroupMappingRepo{},
		PackageDiscountSettingRepo:       &repositories.PackageDiscountSettingRepo{},
		PackageDiscountCourseMappingRepo: &repositories.PackageDiscountCourseMappingRepo{},
	})
	discountPb.RegisterExportServiceServer(grpcServer, s.exportService)

	return nil
}

func (s *server) RegisterNatsSubscribers(_ context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	if err := s.discountService.SubscribeToOrderWithProductInfoLog(); err != nil {
		return fmt.Errorf("orderEventSubscription: failed to subscribe order event %w", err)
	}

	return nil
}

func (*server) GracefulShutdown(context.Context) {}
