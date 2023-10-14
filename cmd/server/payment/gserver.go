package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/payment/configurations"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/search"
	courseMgMt "github.com/manabie-com/backend/internal/payment/services/course_mgmt"
	echoService "github.com/manabie-com/backend/internal/payment/services/echo_service"
	exportService "github.com/manabie-com/backend/internal/payment/services/export_service"
	"github.com/manabie-com/backend/internal/payment/services/file_service"
	serviceForTest "github.com/manabie-com/backend/internal/payment/services/for_test_only"
	importService "github.com/manabie-com/backend/internal/payment/services/import_service"
	internalService "github.com/manabie-com/backend/internal/payment/services/internal_service"
	orderMgmt "github.com/manabie-com/backend/internal/payment/services/order_mgmt"
	eventHandler "github.com/manabie-com/backend/internal/payment/services/order_mgmt/event_handler"
	userInterceptors "github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	fatimaPb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
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

	fatimaConn *grpc.ClientConn

	orderSvc        *orderMgmt.OrderMgMt
	internalSvc     *internalService.InternalService
	authInterceptor *userInterceptors.Auth
	exportSvc       *exportService.ExportService
	fileSvc         *file_service.FileService
	courseSvc       *courseMgMt.CourseMgMt
}

func (*server) ServerName() string {
	return "payment"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()

	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
	}
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

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) (err error) {
	grpcZap.ReplaceGrpcLoggerV2(rsc.Logger())
	dbTrace := rsc.DBWith("fatima")
	s.authInterceptor = authInterceptor(&c, rsc.Logger(), dbTrace.DB)
	storageConfig := &c.Storage

	if c.Common.Organization != "jprep" {
		storageConfig = rsc.Storage()
	}

	s.fatimaConn = rsc.GRPCDial("fatima")

	elasticSearch := search.NewElasticSearch(rsc.Elastic())
	subscriptionModifierServiceClient := fatimaPb.NewSubscriptionModifierServiceClient(s.fatimaConn)
	s.orderSvc = orderMgmt.NewOrderMgMt(dbTrace, elasticSearch, rsc.NATS(), subscriptionModifierServiceClient, rsc.Kafka(), c.Common)
	s.internalSvc = internalService.NewInternalService(dbTrace, rsc.NATS(), rsc.Kafka(), c.Common)
	s.exportSvc = exportService.NewExportService(dbTrace)

	s.fileSvc, err = file_service.NewFileService(dbTrace, *storageConfig)
	if err != nil {
		return fmt.Errorf("creating file service have err : %w", err)
	}

	s.courseSvc = courseMgMt.NewCourseMgMt(dbTrace, rsc.NATS(), rsc.Kafka(), c.Common)
	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcServer *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	dbTrace := rsc.DBWith("fatima")
	paymentPb.RegisterImportMasterDataServiceServer(grpcServer, &importService.ImportMasterDataService{
		DB:                             dbTrace,
		AccountingCategoryRepo:         &repositories.AccountingCategoryRepo{},
		BillingScheduleRepo:            &repositories.BillingScheduleRepo{},
		BillingSchedulePeriodRepo:      &repositories.BillingSchedulePeriodRepo{},
		BillingRatioRepo:               &repositories.BillingRatioRepo{},
		DiscountRepo:                   &repositories.DiscountRepo{},
		TaxRepo:                        &repositories.TaxRepo{},
		ProductAccountingCategoryRepo:  &repositories.ProductAccountingCategoryRepo{},
		ProductGradeRepo:               &repositories.ProductGradeRepo{},
		FeeRepo:                        &repositories.FeeRepo{},
		MaterialRepo:                   &repositories.MaterialRepo{},
		PackageRepo:                    &repositories.PackageRepo{},
		ProductPriceRepo:               &repositories.ProductPriceRepo{},
		PackageCourseRepo:              &repositories.PackageCourseRepo{},
		ProductLocationRepo:            &repositories.ProductLocationRepo{},
		LeavingReasonRepo:              &repositories.LeavingReasonRepo{},
		PackageQuantityTypeMappingRepo: &repositories.PackageQuantityTypeMappingRepo{},
		ProductSettingRepo:             &repositories.ProductSettingRepo{},
		PackageCourseMaterialRepo:      &repositories.PackageCourseMaterialRepo{},
		PackageCourseFeeRepo:           &repositories.PackageCourseFeeRepo{},
		ProductDiscountRepo:            &repositories.ProductDiscountRepo{},
		NotificationDateRepo:           &repositories.NotificationDateRepo{},
	})

	paymentPb.RegisterImportMasterDataForTestServiceServer(grpcServer, &serviceForTest.ImportMasterDataForTestService{
		DB:          dbTrace,
		ForTestRepo: &serviceForTest.ForTestRepo{},
	})

	paymentPb.RegisterOrderServiceServer(grpcServer, s.orderSvc)
	paymentPb.RegisterEchoServiceServer(grpcServer, &echoService.EchoService{})
	paymentPb.RegisterInternalServiceServer(grpcServer, s.internalSvc)
	paymentPb.RegisterExportServiceServer(grpcServer, s.exportSvc)
	paymentPb.RegisterFileServiceServer(grpcServer, s.fileSvc)
	paymentPb.RegisterCourseServiceServer(grpcServer, s.courseSvc)

	health.RegisterHealthServer(grpcServer, &healthcheck.Service{DB: dbTrace.DB.(*pgxpool.Pool)})

	return nil
}

func (s *server) RegisterNatsSubscribers(_ context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zlogger := rsc.Logger()
	jsm := rsc.NATS()
	dbTrace := rsc.DBWith("fatima")

	err := eventHandler.RegisterDiscountEventHandler(
		jsm,
		zlogger,
		dbTrace,
		s.orderSvc,
	)
	if err != nil {
		zlogger.Fatal(fmt.Sprintf("RegisterDiscountEventHandler failed, err: %v", err))
	} else {
		zlogger.Info(fmt.Sprintf("RegisterDiscountEventHandler subscribed: %v", time.Now()))
	}

	return nil
}

func (*server) GracefulShutdown(context.Context) {}
