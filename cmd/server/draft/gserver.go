package draft

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/draft/configurations"
	"github.com/manabie-com/backend/internal/draft/repositories"
	"github.com/manabie-com/backend/internal/draft/service"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	pb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithHTTP(s).
		WithNatsServicer(s).
		Register(s)
}

type server struct {
	githubCollectDataController *service.GithubCollectDataController
	dataCleanerController       *service.DataCleanerController
}

func (*server) ServerName() string {
	return "draft"
}

func (*server) SetupGRPC(_ context.Context, s *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()
	pb.RegisterSendCoverageServiceServer(s, &service.SendCoverageServer{DraftRepo: &repositories.DraftRepo{}, DB: db})
	pb.RegisterBDDSuiteServiceServer(s, &service.BDDSuite{DB: db, Repo: &repositories.BDDSuite{}})
	health.RegisterHealthServer(s, &service.HealthcheckService{})
	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	service.RegisterGithubCollectDataController(r, s.githubCollectDataController)
	service.RegisterDataCleanerController(&c, r, s.dataCleanerController)
	return nil
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	cleanDataRegister := RegisterSubscriber{JSM: rsc.NATS(), Config: c, Rsc: rsc}
	err := cleanDataRegister.Subscribe()
	if err != nil {
		return fmt.Errorf("RegisterNatsSubscribers: cleanDataRegister.Subscribe: %w", err)
	}
	return nil
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	cl := repositories.NewClient(c.Github.AppID, c.Github.InstallationID, []byte(c.GithubPrivateKey))

	s.githubCollectDataController = &service.GithubCollectDataController{
		CFG:             &c,
		DB:              rsc.DB(),
		GithubEventRepo: &repositories.GithubEvent{},
		MergeStatusRepo: &repositories.GithubMergeStatusRepo{},
		GithubClient:    cl,
		Logger:          rsc.Logger().Sugar(),
	}
	s.dataCleanerController = &service.DataCleanerController{JSM: rsc.NATS()}
	return nil
}
func (*server) GracefulShutdown(context.Context) {}
func (*server) WithUnaryServerInterceptors(_ configurations.Config, _ *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	return nil
}

func (s *server) WithStreamServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	grpcStream := bootstrap.DefaultStreamServerInterceptor(rsc)
	return grpcStream
}
func (*server) WithServerOptions() []grpc.ServerOption { return nil }
