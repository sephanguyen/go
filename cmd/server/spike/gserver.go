package spike

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/spike/configurations"
	email_grpc "github.com/manabie-com/backend/internal/spike/modules/email/controller/grpc"
	httpController "github.com/manabie-com/backend/internal/spike/modules/email/controller/http"
	email_kafka "github.com/manabie-com/backend/internal/spike/modules/email/controller/kafka"
	"github.com/manabie-com/backend/internal/spike/modules/email/metrics"
	"github.com/manabie-com/backend/internal/spike/modules/email/middlewares"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	localEnv = "local"
	stagEnv  = "stag"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithKafkaServicer(s).
		WithMonitorServicer(s).
		WithHTTP(s).
		Register(s)
}

type server struct {
	bootstrap.DefaultMonitorService[configurations.Config]
	authInterceptor *interceptors.Auth

	customMetrics metrics.EmailMetrics
}

func (*server) ServerName() string {
	return "spike"
}

func (s *server) SetupGRPC(_ context.Context, grpcserver *grpc.Server, configs configurations.Config, rsc *bootstrap.Resources) error {
	health.RegisterHealthServer(grpcserver, &healthcheck.Service{DB: rsc.DBWith("notificationmgmt").DB.(*pgxpool.Pool)})

	notificationmgmtDB := rsc.DBWith("notificationmgmt")

	emailSvc := email_grpc.NewEmailModifierService(notificationmgmtDB, rsc.Kafka(), s.customMetrics, configs.Common.Environment)
	initNewEmailServer(grpcserver, emailSvc)

	return nil
}

func (s *server) GracefulShutdown(_ context.Context) {}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (s *server) WithPrometheusCollectors(*bootstrap.Resources) []prometheus.Collector {
	return s.customMetrics.GetCollectors()
}

func (s *server) InitMetricsValue() {
	s.customMetrics.InitCounterValue()
}

func (s *server) WithUnaryServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()

	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
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

func (s *server) InitKafkaConsumers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	emailProvider := initEmailProvider(c, rsc.Logger())

	// NOTED [IMPORTANT]: If enable it on PROD, be careful for synersia cluster because it doesn't have any Kafka deployment (at this time this comment are written)
	if c.Common.Environment == localEnv || c.Common.Environment == stagEnv {
		consumersRegistered := email_kafka.NewConsumersRegistered(rsc.DBWith("notificationmgmt"), rsc.Kafka(), emailProvider, rsc.Logger())
		err := consumersRegistered.Consume()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	notificationmgmtDBTrace := rsc.DBWith("notificationmgmt")

	s.customMetrics = metrics.NewClientMetrics("spike")

	s.authInterceptor = authInterceptor(&c, zapLogger, notificationmgmtDBTrace.DB)
	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	notificationmgmtDBTrace := rsc.DBWith("notificationmgmt")
	emailProvider := initEmailProvider(c, zapLogger)
	emailHTTPService := httpController.NewEmailHTTPService(notificationmgmtDBTrace, zapLogger, s.customMetrics, c.EmailWebhookConfig)

	superGroup := r.Group("/spike/api/v1")
	{
		webhookAPIGroup := superGroup.Group("/spike/email_status_receiver")
		{
			webhookAPIGroup.Use(middlewares.AuthenticateWebhookRequest(zapLogger, emailProvider))
			webhookAPIGroup.POST("/", emailHTTPService.EmailStatusReceiver)
		}

		// Examples
		// otherAPIGroup := superGroup.Group("/spike")
		// {
		// 	otherAPIGroup.GET("/get_sent_email", func(ctx *gin.Context) {
		// 		ctx.JSON(http.StatusOK, gin.H{
		// 			"resp": "ok",
		// 		})
		// 	})

		// 	otherAPIGroup.GET("/get_recipients", func(ctx *gin.Context) {
		// 		ctx.JSON(http.StatusOK, gin.H{
		// 			"resp": "ok",
		// 		})
		// 	})
		// }
	}
	return nil
}
