package invoicemgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/invoicemgmt/configurations"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	dataMigrationService "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration"
	exportService "github.com/manabie-com/backend/internal/invoicemgmt/services/export_service"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	http_port "github.com/manabie-com/backend/internal/invoicemgmt/services/http"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/http/middleware"
	importService "github.com/manabie-com/backend/internal/invoicemgmt/services/import_service"
	invoiceService "github.com/manabie-com/backend/internal/invoicemgmt/services/invoice"
	openAPIService "github.com/manabie-com/backend/internal/invoicemgmt/services/open_api"
	paymentService "github.com/manabie-com/backend/internal/invoicemgmt/services/payment"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/payment_detail"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	"github.com/manabie-com/backend/internal/invoicemgmt/subscriptions"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	pb_mastermgmt "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/gin-gonic/gin"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithHTTP(s).
		WithNatsServicer(s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct {
	authInterceptor      *interceptors.Auth
	orderMgmtConn        *grpc.ClientConn
	invoiceSvc           *invoiceService.InvoiceModifierService
	importSvc            *importService.ImportMasterDataService
	paymentSvc           *paymentService.PaymentModifierService
	editPaymentDetailSvc *payment_detail.EditPaymentDetailService
	exportSvc            *exportService.ExportMasterDataService
	dataMigrationSvc     *dataMigrationService.DataMigrationModifierService
	fileStorage          filestorage.FileStorage
	bootstrap.DefaultMonitorService[configurations.Config]
	openAPISvc     *openAPIService.OpenAPIModifierService
	shamirConn     *grpc.ClientConn
	mastermgmtConn *grpc.ClientConn
}

func (s *server) ServerName() string {
	return "invoicemgmt"
}

func (s *server) GracefulShutdown(context.Context) {
	s.fileStorage.Close()
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	sugar := zapLogger.Sugar()
	dbTrace := rsc.DB()
	unleashClient := rsc.Unleash()
	jsm := rsc.NATS()

	s.orderMgmtConn = rsc.GRPCDial("payment")
	s.shamirConn = rsc.GRPCDial("shamir")
	s.mastermgmtConn = rsc.GRPCDial("mastermgmt")

	s.authInterceptor = authInterceptor(&c, zapLogger, dbTrace)

	internalOrderService := payment_pb.NewInternalServiceClient(s.orderMgmtConn)

	mastermgmtService := pb_mastermgmt.NewInternalServiceClient(s.mastermgmtConn)

	// Initialize the file store to be used. The default storage is Google Cloud Storage
	fileStorageName := filestorage.GoogleCloudStorageService
	if strings.Contains(c.Storage.Endpoint, "minio") {
		fileStorageName = filestorage.MinIOService
	}

	fileStorage, err := filestorage.GetFileStorage(fileStorageName, &c.Storage)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to init %s file storage", fileStorageName), zap.Error(err))
	}
	s.fileStorage = fileStorage

	// Initialize all repositories
	repositories := initRepositories()

	s.invoiceSvc = invoiceService.NewInvoiceModifierService(
		*sugar,
		dbTrace,
		internalOrderService,
		fileStorage,
		getInvoiceServiceRepositories(repositories),
		unleashClient,
		c.Common.Environment,
		&utils.TempFileCreator{TempDirPattern: constant.InvoicemgmtTemporaryDir},
	)

	s.paymentSvc = paymentService.NewPaymentModifierService(
		*sugar,
		dbTrace,
		getPaymentServiceRepositories(repositories),
		fileStorage,
		unleashClient,
		&utils.TempFileCreator{TempDirPattern: constant.InvoicemgmtTemporaryDir},
		c.Common.Environment,
	)

	s.importSvc = importService.NewImportMasterDataService(
		*sugar,
		dbTrace,
		getImportMasterDataServiceRepositories(repositories),
		unleashClient,
		c.Common.Environment,
	)

	s.editPaymentDetailSvc = &payment_detail.EditPaymentDetailService{
		Logger:                            sugar,
		DB:                                dbTrace,
		PrefectureRepo:                    repositories.PrefectureRepo,
		BillingAddressRepo:                repositories.BillingAddressRepo,
		StudentPaymentDetailRepo:          repositories.StudentPaymentDetailRepo,
		BankAccountRepo:                   repositories.BankAccountRepo,
		BankRepo:                          repositories.BankRepo,
		BankBranchRepo:                    repositories.BankBranchRepo,
		StudentPaymentDetailActionLogRepo: repositories.StudentPaymentDetailActionLogRepo,
	}

	s.exportSvc = exportService.NewExportMasterDataService(
		*sugar,
		dbTrace,
		getExportMasterDataServiceRepositories(repositories),
	)

	s.dataMigrationSvc = dataMigrationService.NewDataMigrationModifierService(
		*sugar,
		dbTrace,
		getDataMigrationServiceRepositories(repositories),
	)

	s.openAPISvc = openAPIService.NewOpenAPIModifierService(
		*sugar,
		jsm,
		dbTrace,
		getOpenAPIServiceRepositories(repositories),
		unleashClient,
		mastermgmtService,
		c.Common.Environment,
	)

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserv *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	health.RegisterHealthServer(grpcserv, &healthcheck.Service{DB: rsc.DB().DB.(*pgxpool.Pool)})
	invoice_pb.RegisterInvoiceServiceServer(grpcserv, s.invoiceSvc)
	invoice_pb.RegisterPaymentServiceServer(grpcserv, s.paymentSvc)
	invoice_pb.RegisterImportMasterDataServiceServer(grpcserv, s.importSvc)
	invoice_pb.RegisterInternalServiceServer(grpcserv, s.invoiceSvc)
	invoice_pb.RegisterEditPaymentDetailServiceServer(grpcserv, s.editPaymentDetailSvc)
	invoice_pb.RegisterExportMasterDataServiceServer(grpcserv, s.exportSvc)
	invoice_pb.RegisterDataMigrationServiceServer(grpcserv, s.dataMigrationSvc)

	return nil
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()

	// don't use UnaryActivityLogRequestInterceptor, it does not have nats account yet
	grpcUnary := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(rsc.Logger(), grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel)),
	}

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

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	studentCreateSubscription := subscriptions.UserEventSubscription{
		Config:         &c,
		JSM:            rsc.NATS(),
		Logger:         rsc.Logger(),
		OpenAPIService: s.openAPISvc,
	}
	// err := studentCreateSubscription.Subscribe()
	// if err != nil {
	// 	return fmt.Errorf("studentCreateSubscription.Subscribe: %w", err)
	// }

	err := studentCreateSubscription.PullSubscribe()
	if err != nil {
		return fmt.Errorf("studentCreateSubscription.PullSubscribe: %w", err)
	}
	return nil
}

func (s *server) SetupHTTP(_ configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()

	healthCheckHTTP := http_port.HealthCheckService{}
	groupDecider := middleware.NewGroupDecider(rsc.DBWith("invoicemgmt"))

	r.Use(middleware.VerifySignature(zapLogger, groupDecider, spb.NewTokenReaderServiceClient(s.shamirConn)))
	r.GET(constant.HealthCheckStatusEndpoint, healthCheckHTTP.Status)
	r.PUT(constant.StudentBankInfoEndpoint, s.openAPISvc.UpsertStudentBankAccountInfo)
	return nil
}
