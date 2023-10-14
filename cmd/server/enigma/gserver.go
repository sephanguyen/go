package enigma

import (
	"context"

	"github.com/manabie-com/backend/internal/enigma/configurations"
	"github.com/manabie-com/backend/internal/enigma/controllers"
	"github.com/manabie-com/backend/internal/enigma/middlewares"
	"github.com/manabie-com/backend/internal/enigma/repositories"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	s := &server{}
	bootstrap.WithHTTP[configurations.Config](s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct {
	bobConn *grpc.ClientConn
	bootstrap.DefaultMonitorService[configurations.Config]
}

func (s *server) ServerName() string {
	return "enigma"
}

// ServerName() string
func (s *server) GracefulShutdown(context.Context) {}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	s.bobConn = rsc.GRPCDial("bob")
	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	dbTrace := rsc.DBWith("bob")
	jsm := rsc.NATS()

	r.Use(tracingMiddleware)

	controllers.RegisterLBCheckerController(r.Group("lb-checker"), zapLogger, &c)
	controllers.RegisterImageController(r.Group("image"), bpb.NewUserServiceClient(s.bobConn), bpb.NewInternalClient(s.bobConn), zapLogger)
	controllers.RegisterHealthCheckController(r.Group("healthcheck"), zapLogger, dbTrace)
	controllers.RegisterUpdateReleaseController(r.Group("release"), zapLogger, &c)

	jprepRoute := r.Group("jprep")
	jprepRoute.Use(
		middlewares.VerifySignature(
			middlewares.JPREPHeaderKey,
			c.JPREPSignatureSecret,
		),
		middlewares.VerifyTimestamp(
			int64(c.JPREPPayloadExpiredSec),
		),
	)
	jPREPController := &controllers.JPREPController{
		Logger:                      zapLogger,
		JSM:                         jsm,
		DB:                          dbTrace,
		PartnerSyncDataLogRepo:      &repositories.PartnerSyncDataLogRepo{},
		PartnerSyncDataLogSplitRepo: &repositories.PartnerSyncDataLogSplitRepo{},
	}
	controllers.RegisterJPREPController(jprepRoute, jPREPController)

	cloudConvertRoute := r.Group("cloud-convert")
	cloudConvertRoute.Use(middlewares.VerifySignature(
		middlewares.CloudConvertSigKey,
		c.CloudConvertSigningSecret,
	))
	controllers.RegisterCloudConvertController(cloudConvertRoute, zapLogger, jsm)
	return nil
}

func tracingMiddleware(c *gin.Context) {
	tracingHeaders := []string{
		"X-Request-Id",
		"X-B3-Traceid",
		"X-B3-Spanid",
		"X-B3-Sampled",
		"X-B3-Parentspanid",
		"X-B3-Flags",
		"X-Ot-Span-Context",
	}
	ctx := c.Request.Context()
	for _, key := range tracingHeaders {
		if val := c.Request.Header.Get(key); val != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, key, val)
		}
	}

	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
