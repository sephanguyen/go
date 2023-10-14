package auth

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/usermgmt/modules/auth/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/modules/auth/core/service"
	grpcPort "github.com/manabie-com/backend/internal/usermgmt/modules/auth/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	apb "github.com/manabie-com/backend/pkg/manabuf/auth/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

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
	shamirConn      *grpc.ClientConn

	authService *grpcPort.AuthService
}

func (*server) ServerName() string {
	return "auth"
}

func (s *server) WithUnaryServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(nil, rsc.Logger(), s.ServerName()),
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

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	_, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	db := rsc.DB()
	s.authInterceptor = authInterceptor(&c, rsc.Logger(), db)
	s.shamirConn = rsc.GRPCDial("shamir")

	domainAuth := service.DomainAuthService{
		ShamirClient: spb.NewTokenReaderServiceClient(s.shamirConn),
	}

	authService := grpcPort.AuthService{
		DomainAuthService: &domainAuth,
	}

	s.authService = &authService

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserv *grpc.Server, _ configurations.Config, _ *bootstrap.Resources) error {
	health.RegisterHealthServer(grpcserv, &grpcPort.HealthcheckService{})

	apb.RegisterAuthServiceServer(grpcserv, s.authService)
	return nil
}

func (s *server) GracefulShutdown(_ context.Context) {
	s.shamirConn.Close()
}
