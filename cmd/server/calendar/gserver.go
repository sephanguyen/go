package calendar

import (
	"context"

	"github.com/manabie-com/backend/internal/calendar/configurations"
	cld_ctrl "github.com/manabie-com/backend/internal/calendar/controller"
	cld_repo "github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/calendar/support"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	healthcheck "github.com/manabie-com/backend/internal/golibs/healthcheck"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	cld_pb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
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

	calendarDB *pgxpool.Pool

	dateInfoReaderService    *cld_ctrl.DateInfoReaderService
	dateInfoModifierService  *cld_ctrl.DateInfoModifierService
	schedulerModifierService *cld_ctrl.SchedulerModifierService
	userReaderService        *cld_ctrl.UserReaderService
	lessonReaderService      *cld_ctrl.LessonReaderService
}

func (*server) ServerName() string {
	return "calendar"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	grpcUnary := []grpc.UnaryServerInterceptor{
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(rsc.Logger(), grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel)),
		s.authInterceptor.UnaryServerInterceptor,
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

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	bobDBTrace := rsc.DBWith("bob")
	calendarDBTrace := rsc.DBWith("calendar")
	lessonDBTrace := rsc.DBWith("lessonmgmt")
	unleashClient := rsc.Unleash()
	env := c.Common.Environment
	wrapperConnection := support.InitWrapperDBConnector(bobDBTrace, lessonDBTrace, unleashClient, env)

	s.authInterceptor = authInterceptor(&c, rsc.Logger(), bobDBTrace.DB)

	dateInfoRepo := &cld_repo.DateInfoRepo{}
	locationRepo := &cld_repo.LocationRepo{}
	schedulerRepo := &cld_repo.SchedulerRepo{}
	userRepo := &cld_repo.UserRepo{}
	lessonRepo := &cld_repo.LessonRepo{
		LessonmgmtLessonRepo: &lesson_repo.LessonRepo{},
	}
	lessonTeacherRepo := &cld_repo.LessonTeacherRepo{
		LessonmgmtLessonTeacherRepo: &lesson_repo.LessonTeacherRepo{},
	}
	lessonMemberRepo := &cld_repo.LessonMemberRepo{
		LessonmgmtLessonMemberRepo: &lesson_repo.LessonMemberRepo{},
	}
	lessonClassroomRepo := &cld_repo.LessonClassroomRepo{
		LessonmgmtLessonClassroom: &lesson_repo.LessonClassroomRepo{},
	}
	lessonGroupRepo := &cld_repo.LessonGroupRepo{
		LessonmgmtLessonGroupRepo: &lesson_repo.LessonGroupRepo{},
	}

	s.dateInfoReaderService = cld_ctrl.NewDateInfoReaderService(calendarDBTrace, dateInfoRepo)
	s.dateInfoModifierService = cld_ctrl.NewDateInfoModifierService(calendarDBTrace, dateInfoRepo, locationRepo)
	s.schedulerModifierService = cld_ctrl.NewSchedulerModifierService(schedulerRepo, calendarDBTrace)
	s.userReaderService = cld_ctrl.NewUserReaderService(userRepo, bobDBTrace, unleashClient, env)
	s.lessonReaderService = cld_ctrl.NewLessonReaderService(
		wrapperConnection,
		lessonRepo,
		lessonTeacherRepo,
		lessonMemberRepo,
		lessonClassroomRepo,
		lessonGroupRepo,
		schedulerRepo,
		userRepo,
		env,
		unleashClient,
	)

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcServer *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	cld_pb.RegisterDateInfoReaderServiceServer(grpcServer, s.dateInfoReaderService)
	cld_pb.RegisterDateInfoModifierServiceServer(grpcServer, s.dateInfoModifierService)
	cld_pb.RegisterSchedulerModifierServiceServer(grpcServer, s.schedulerModifierService)
	cld_pb.RegisterUserReaderServiceServer(grpcServer, s.userReaderService)
	cld_pb.RegisterLessonReaderServiceServer(grpcServer, s.lessonReaderService)
	health.RegisterHealthServer(grpcServer, &healthcheck.Service{DB: rsc.DB().DB.(*pgxpool.Pool)})

	return nil
}

func (s *server) GracefulShutdown(context.Context) {
	if s.calendarDB != nil {
		s.calendarDB.Close()
	}
}
