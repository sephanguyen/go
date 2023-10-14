package entryexitmgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/bob/services/uploads"
	"github.com/manabie-com/backend/internal/entryexitmgmt/configurations"
	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/repositories"
	services "github.com/manabie-com/backend/internal/entryexitmgmt/services"
	eeFS "github.com/manabie-com/backend/internal/entryexitmgmt/services/filestore"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services/uploader"
	natstransport "github.com/manabie-com/backend/internal/entryexitmgmt/transport/nats"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/curl"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithNatsServicer(s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct {
	qrCodeService   *services.EntryExitModifierService
	authInterceptor *interceptors.Auth
	fireStore       eeFS.FileStore
	bootstrap.DefaultMonitorService[configurations.Config]
	mastermgmtConn *grpc.ClientConn
}

func (*server) ServerName() string {
	return "entryexitmgmt"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
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
	zapLogger := rsc.Logger()
	sugar := zapLogger.Sugar()
	jsm := rsc.NATS()
	storageConfig := rsc.Storage()
	s.mastermgmtConn = rsc.GRPCDial("mastermgmt")

	mastermgmtService := mpb.NewInternalServiceClient(s.mastermgmtConn)

	var (
		bobFsName filestore.ServiceName
		fsName    constant.FileStoreName
	)

	// Initialize the file store to be used. The default storage is google cloud
	switch {
	case strings.Contains(storageConfig.Endpoint, "minio"):
		bobFsName = filestore.MinIOService
		fsName = constant.MinIOService
	default:
		bobFsName = filestore.GoogleCloudStorageService
		fsName = constant.GoogleCloudStorageService
	}

	fs, err := filestore.NewFileStore(bobFsName, c.Common.ServiceAccountEmail, storageConfig)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to init %s file storage", bobFsName), zap.Error(err))
	}

	newFs, err := eeFS.GetFileStore(fsName, storageConfig)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to init %s file storage", fsName), zap.Error(err))
	}
	s.fireStore = newFs

	sdkFileUploader := &uploader.SDKUploaderService{FileStore: newFs, Cfg: storageConfig}
	uploadReaderService := &uploads.UploadReaderService{FileStore: fs, Cfg: *storageConfig}
	curlFileUploader := &uploader.CurlUploaderService{UploadReaderService: uploadReaderService, HTTP: &curl.HTTP{}}

	uploadServiceSelector := &services.UploadServiceSelector{
		SdkUploadService:  sdkFileUploader,
		CurlUploadService: curlFileUploader,
		UnleashClient:     rsc.Unleash(),
		Env:               c.Common.Environment,
	}

	db := rsc.DB()
	studentQRRepo := &repositories.StudentQRRepo{}
	studentEntryExitRecordsRepo := &repositories.StudentEntryExitRecordsRepo{}
	entryExitQueueRepo := &repositories.EntryExitQueueRepo{}

	// Check encryption key
	if strings.TrimSpace(c.QrCodeEncryption.SecretKey) == "" {
		zapLogger.Fatal("encryption secret key cannot be empty")
	}

	studentRepo := &repositories.StudentRepo{}
	studentParentRepo := &repositories.StudentParentRepo{}
	userRepo := &repositories.UserRepo{}
	s.qrCodeService = services.NewEntryExitModifierService(
		&services.Libraries{
			Logger:                *sugar,
			JSM:                   jsm,
			DB:                    db,
			UploadServiceSelector: uploadServiceSelector,
		},
		&services.Repositories{
			StudentQRRepo:               studentQRRepo,
			StudentEntryExitRecordsRepo: studentEntryExitRecordsRepo,
			EntryExitQueueRepo:          entryExitQueueRepo,
			StudentRepo:                 studentRepo,
			StudentParentRepo:           studentParentRepo,
			UserRepo:                    userRepo,
		},
		&services.QREncryptionSecretKeys{
			EncryptionKey:         c.QrCodeEncryption.SecretKey,
			EncryptionKeyTokyo:    c.QrCodeEncryptionTokyo.SecretKey,
			EncryptionKeySynersia: c.QrCodeEncryptionSynersia.SecretKey,
		},
		mastermgmtService,
	)

	s.authInterceptor = authInterceptor(&c, zapLogger, db)

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserver *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	health.RegisterHealthServer(grpcserver, &services.HealthcheckService{})
	eepb.RegisterEntryExitServiceServer(grpcserver, s.qrCodeService)
	return nil
}

func (s *server) GracefulShutdown(ctx context.Context) {
	s.fireStore.Close()
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	studentCreateSubscription := natstransport.UserEventSubscription{
		Config:           &c,
		JSM:              rsc.NATS(),
		Logger:           rsc.Logger(),
		EntryExitService: s.qrCodeService,
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
