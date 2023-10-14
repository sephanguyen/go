package conversationmgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/configurations"
	agora_usermgmt_grpc "github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/controller/grpc"
	convo_nats "github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/controller/nats"
	convo_grpc "github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/controller/grpc"
	convo_http "github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/controller/http"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/service"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/middleware"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
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
		WithHTTP(s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct {
	bootstrap.DefaultMonitorService[configurations.Config]
	authInterceptor *interceptors.Auth
}

func (*server) ServerName() string {
	return "conversationmgmt"
}

func (s *server) SetupGRPC(_ context.Context, grpcserver *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	tomDB := rsc.DBWith("tom")
	logger := rsc.Logger()

	health.RegisterHealthServer(grpcserver, &healthcheck.Service{DB: tomDB.DB.(*pgxpool.Pool)})

	chatVendor := initChatProvider(c, logger)

	agoraUserMgmtGRPC := agora_usermgmt_grpc.NewAgoraUserMgmtService(tomDB, chatVendor, logger)

	conversationModifierService := service.NewConversationModifierService(tomDB, logger, c.Common.Environment, chatVendor)
	conversationModifierGRPC := convo_grpc.NewNotificationModifierGRPC(conversationModifierService)

	conversationReaderService := service.NewConversationReaderService(tomDB, logger, c.Common.Environment, chatVendor)
	conversationReaderGRPC := convo_grpc.NewConversationReaderGRPC(conversationReaderService)

	initAgoraUserMgmtServer(grpcserver, agoraUserMgmtGRPC)
	initConversationModifierServer(grpcserver, conversationModifierGRPC)
	initConversationReaderServer(grpcserver, conversationReaderGRPC)
	return nil
}

func (s *server) GracefulShutdown(ctx context.Context) {}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (s *server) WithUnaryServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
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

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	tomDBTrace := rsc.DBWith("tom")

	s.authInterceptor = authInterceptor(&c, zapLogger, tomDBTrace.DB)
	return nil
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	chatVendor := initChatProvider(c, rsc.Logger())
	zapLogger.Sugar().Info("[agora]: init Agora Client successfully.")

	if c.Common.Environment == localEnv || c.Common.Environment == stagEnv {
		subscribersRegistered := convo_nats.NewSubscribersRegistered(rsc.DBWith("tom"), rsc.NATS(), chatVendor, rsc.Logger())
		err := subscribersRegistered.StartSubscribeForAllSubscribers()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()

	chatVendor := initChatProvider(c, rsc.Logger())
	conversationModifierService := service.NewConversationModifierService(rsc.DBWith("tom"), rsc.Logger(), c.Common.Environment, chatVendor)
	notificationHandlerService := service.NewNotificationHandlerService(rsc.DBWith("tom"), rsc.Logger(), c.Common.Environment)
	conversationModifierGRPC := convo_http.NewNotificationModifierGTTP(rsc.DBWith("tom"), rsc.Logger(), conversationModifierService, notificationHandlerService)

	superGroupV1 := r.Group("/conversationmgmt/api/v1")
	{
		conversationMgmtAPIGroup := superGroupV1.Group("/conversationmgmt")
		{
			messageEvent := conversationMgmtAPIGroup.Group("/message_event")
			{
				messageEvent.Use(middleware.VerifyWebhookRequest(c.Agora, zapLogger, agora.NewWebhookVerifier()))
				messageEvent.POST("/", conversationModifierGRPC.HandleMessageEvent)
			}
		}
	}
	return nil
}
