package notificationmgmt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/clients"
	firebaseLib "github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/notification/config"
	infra "github.com/manabie-com/backend/internal/notification/infra"
	metrics "github.com/manabie-com/backend/internal/notification/infra/metrics"
	"github.com/manabie-com/backend/internal/notification/mock"
	mediaController "github.com/manabie-com/backend/internal/notification/modules/media/controller"
	mediaRepo "github.com/manabie-com/backend/internal/notification/modules/media/infrastructure/repositories"
	systemNotificationController "github.com/manabie-com/backend/internal/notification/modules/system_notification/controller"
	systemNotificationKafka "github.com/manabie-com/backend/internal/notification/modules/system_notification/controller/kafka"
	tagServices "github.com/manabie-com/backend/internal/notification/modules/tagmgmt/services"
	"github.com/manabie-com/backend/internal/notification/services"
	"github.com/manabie-com/backend/internal/notification/subscribers"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	firebase "firebase.google.com/go/v4"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

const (
	localEnv string = "local"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[config.Config](s).
		WithNatsServicer(s).
		WithMonitorServicer(s).
		WithHTTP(s).
		WithKafkaServicer(s).
		Register(s)
}

type server struct {
	bootstrap.DefaultMonitorService[config.Config]

	authInterceptor *interceptors.Auth

	customMetrics metrics.NotificationMetrics
	s3Session     *session.Session

	notificationPusher firebaseLib.NotificationPusher
}

func (s *server) ServerName() string {
	return "notificationmgmt"
}

func (s *server) GracefulShutdown(_ context.Context) {}

func (s *server) RegisterNatsSubscribers(_ context.Context, c config.Config, rsc *bootstrap.Resources) error {
	pushNotificationService := infra.NewPushNotificationService(s.notificationPusher, s.customMetrics)

	storageConfig := rsc.Storage()

	newNotiModifierSvc := services.NewNotificationModifierService(rsc.DBWith("bob"), *storageConfig, s.s3Session, pushNotificationService, s.customMetrics, rsc.NATS(), c.Common.Environment)

	notiSubscriber := subscribers.NewNotificationSubscriber(newNotiModifierSvc)

	initEventPushNotification(rsc.NATS(), rsc.Logger(), notiSubscriber)

	initEventSyncStudentPackage(rsc.NATS(), rsc.Logger(), newNotiModifierSvc)

	initEventSyncJprepStudentPackage(rsc.NATS(), rsc.Logger(), newNotiModifierSvc)

	initEventSyncJprepStudentClass(rsc.NATS(), rsc.Logger(), newNotiModifierSvc)

	return nil
}

func (s *server) InitDependencies(c config.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	storageConfig := rsc.Storage()

	s.customMetrics = metrics.NewClientMetrics("notification")

	notificationMgmtDB := rsc.DBWith("notificationmgmt")
	if notificationMgmtDB == nil {
		zapLogger.Error("NotificationMgmt database connection is missing")
	}

	s.authInterceptor = authInterceptor(&c, zapLogger, rsc.DBWith("bob"))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.Storage.InsecureSkipVerify, //nolint:gosec
		},
	}
	s3Sess, err := session.NewSession(&aws.Config{
		HTTPClient:       &http.Client{Transport: tr},
		Region:           aws.String(storageConfig.Region),
		Credentials:      credentials.NewStaticCredentials(storageConfig.AccessKey, storageConfig.SecretKey, ""),
		Endpoint:         aws.String(storageConfig.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("session.NewSession %v", err)
	}
	s.notificationPusher = initNotificationPusher(context.Background(), &c, zapLogger)
	s.s3Session = s3Sess
	return nil
}

func (s *server) WithPrometheusCollectors(*bootstrap.Resources) []prometheus.Collector {
	return s.customMetrics.GetCollectors()
}

func (s *server) InitMetricsValue() {
	s.customMetrics.InitCounterValue()
}

func (s *server) SetupHTTP(c config.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	mux := gateway.NewServeMux(
		gateway.WithOutgoingHeaderMatcher(clients.IsHeaderAllowed),
		gateway.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			authHeader := request.Header.Get("Authorization")
			pkgHeader := request.Header.Get("pkg")
			versionHeader := request.Header.Get("version")

			// add defaults for pkg header
			if pkgHeader == "" {
				pkgHeader = "com.manabie.liz"
			}

			// add defaults for version header
			if versionHeader == "" {
				versionHeader = "1.0.0"
			}

			md := metadata.Pairs(
				"token", authHeader,
				"pkg", pkgHeader,
				"version", versionHeader,
			)
			return md
		}),
		gateway.WithErrorHandler(func(ctx context.Context, mux *gateway.ServeMux, marshaler gateway.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			newError := gateway.HTTPStatusError{
				HTTPStatus: 400,
				Err:        err,
			}
			gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, &newError)
		}))

	err := setupGrpcGateway(mux, rsc.GetGRPCPort(c.Common.Name))
	if err != nil {
		return fmt.Errorf("error setupGrpcGateway %s", err)
	}

	superGroup := r.Group("/notificationmgmt/api/v1")
	{
		superGroup.Group("/proxy/*{grpc_gateway}").Any("", gin.WrapH(mux))
	}
	return nil
}

func setupGrpcGateway(mux *gateway.ServeMux, port string) error {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(&tracer.B3Handler{ClientHandler: &ocgrpc.ClientHandler{}}),
	}

	serviceMap := map[string]func(context.Context, *gateway.ServeMux, string, []grpc.DialOption) error{
		"SystemNotificationReaderService":   npb.RegisterSystemNotificationReaderServiceHandlerFromEndpoint,
		"SystemNotificationModifierService": npb.RegisterSystemNotificationModifierServiceHandlerFromEndpoint,
	}

	for _, registerFunc := range serviceMap {
		err := registerFunc(context.Background(), mux, fmt.Sprintf("localhost%s", port), dialOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserv *grpc.Server, c config.Config, rsc *bootstrap.Resources) error {
	bobDB := rsc.DBWith("bob")
	notificationmgmtDB := rsc.DBWith("notificationmgmt")
	storageConfig := rsc.Storage()
	health.RegisterHealthServer(grpcserv, &healthcheck.Service{DB: bobDB.DB.(*pgxpool.Pool)})

	pushNotificationService := infra.NewPushNotificationService(s.notificationPusher, s.customMetrics)

	newNotiModifierSvc := services.NewNotificationModifierService(bobDB, *storageConfig, s.s3Session, pushNotificationService, s.customMetrics, rsc.NATS(), c.Common.Environment)
	newNotiReaderSvc := services.NewNotificationReaderService(bobDB, c.Common.Environment)

	newTagModifierSvc := tagServices.NewTagModifierService(bobDB)
	newTagReaderSvc := tagServices.NewTagReaderService(bobDB)

	internalSvc := services.NewInternalService(pushNotificationService)

	newMediaModifierSvc := mediaController.NewMediaModifierService(bobDB, &mediaRepo.MediaRepo{})

	newSystemNotificationReaderSvc := systemNotificationController.NewSystemNotificationReaderService(notificationmgmtDB)
	newSystemNotificationModifierSvc := systemNotificationController.NewSystemNotificationModifierService(notificationmgmtDB)

	initServersFromOldBob(&c, grpcserv, bobDB)
	initServersFromOldYasuo(&c, grpcserv, bobDB, s.s3Session, s.notificationPusher, s.customMetrics, rsc.Logger(), rsc.NATS())
	initNewNotificationServer(grpcserv, newNotiModifierSvc, newNotiReaderSvc)
	initNewTagmgmtServer(grpcserv, newTagReaderSvc, newTagModifierSvc)
	initBobUserServerService(grpcserv, newNotiModifierSvc)
	initInternalServerService(grpcserv, internalSvc)
	initNewMediaServer(grpcserv, newMediaModifierSvc)
	initNewNotificationV2Service(grpcserv, newNotiReaderSvc)
	initNewSystemNotificationService(grpcserv, newSystemNotificationReaderSvc, newSystemNotificationModifierSvc)
	return nil
}

func (s *server) WithUnaryServerInterceptors(c config.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		UnaryAccessControlErrorHandlingInterceptor,
	}
	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	grpcUnary = append(grpcUnary, customs...)

	return grpcUnary
}

func (s *server) WithStreamServerInterceptors(c config.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	grpcStream := bootstrap.DefaultStreamServerInterceptor(rsc)
	grpcStream = append(grpcStream, s.authInterceptor.StreamServerInterceptor)
	return grpcStream
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func initNotificationPusher(ctx context.Context, config *config.Config, zapLogger *zap.Logger) (notificationPusher firebaseLib.NotificationPusher) {
	firebaseProject := config.Common.FirebaseProject
	if firebaseProject == "" {
		zapLogger.Fatal("missing config for firebase project")
	}

	if config.Common.Environment != localEnv {
		firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{
			ProjectID: firebaseProject,
		})
		if err != nil {
			zapLogger.Fatal("failed to initialize Firebase App", zap.Error(err))
		}
		fcmClient, err := firebaseApp.Messaging(ctx)
		if err != nil {
			zapLogger.Fatal("failed to get FCM client firebaseApp.Messaging", zap.Error(err))
		}
		notificationPusher = firebaseLib.NewNotificationPusher(fcmClient)
		zapLogger.Info("Init Firebase Cloud Messaging successfully!")
		return
	}
	notificationPusher = mock.NewNotificationPusher()
	return
}

func (s *server) InitKafkaConsumers(_ context.Context, c config.Config, rsc *bootstrap.Resources) error {
	notificationDB := rsc.DBWith("notificationmgmt")
	consumersRegistered := systemNotificationKafka.NewConsumersRegistered(notificationDB, rsc.Kafka(), rsc.Logger(), s.customMetrics)
	err := consumersRegistered.Consume()
	if err != nil {
		return err
	}

	return nil
}
