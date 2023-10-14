package lessonmgmt

import (
	"context"
	"fmt"
	"net/http"
	"time"

	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/clients"
	configs "github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	"github.com/manabie-com/backend/internal/lessonmgmt/healthcheck"
	allocation_controller "github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/controller"
	allocation_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/infrastructure/repo"
	assigned_student "github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student"
	classdo_controller "github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/controller"
	course_location_scheduling_controller "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/controller"
	course_location_schedule_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/infrastructure/repo"
	course_location_schedule_service "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/service"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/controller"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/mediaadapter"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/usermodadapter"
	lesson_report_controller "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/controller"
	lesson_report_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure/repo"
	academic_week_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/academic_week/repository"
	academic_year_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/academic_year/repository"
	class_controller "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/class/controller"
	class_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/class/repository"
	class_usecase "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/class/usecase"
	course_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/course/repository"
	lesson_media "github.com/manabie-com/backend/internal/lessonmgmt/modules/media"
	lesson_media_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user"
	user_controller "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/controller"
	user_infras_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	controller_zoom "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/controller"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	master_class_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	pb_lesson "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/gin-gonic/gin"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
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
	authInterceptor *interceptors.Auth
	bootstrap.DefaultMonitorService[configurations.Config]
	retryOptions        *configs.RetryOptions
	httpClient          *clients.HTTPClient
	configurationClient *clients.ConfigurationClient
	schedulerClient     *clients.SchedulerClient
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	bobDBTrace := rsc.DBWith("bob")
	lessonDBTrace := rsc.DBWith("lessonmgmt")
	unleashClient := rsc.Unleash()
	wrapperConnection := support.InitWrapperDBConnector(bobDBTrace, lessonDBTrace, unleashClient, c.Common.Environment)
	zapLogger := rsc.Logger()
	jsm := rsc.NATS()

	if err := controller.RegisterLessonStudentSubscriptionHandler(
		jsm,
		zapLogger,
		wrapperConnection,
		c.Common.Environment,
		unleashClient,
		&repo.LessonMemberRepo{},
		&lesson_report_repo.LessonReportRepo{},
		&repo.ReallocationRepo{},
		&repo.LessonRepo{},
		commands.LessonCommandHandler{
			MasterDataPort: &repo.MasterDataRepo{},
		},
	); err != nil {
		zapLogger.Fatal("RegisterLessonStudentSubscriptionHandler: ", zap.Error(err))
	}

	if err := controller.RegisterLockLessonSubscriptionHandler(
		jsm,
		zapLogger,
		wrapperConnection,
		&repo.LessonRepo{},
		c.Common.Environment,
		unleashClient,
	); err != nil {
		zapLogger.Fatal("RegisterLockLessonSubscriptionHandler: ", zap.Error(err))
	}

	if err := controller.RegisterStudentClassSubscriptionHandler(
		jsm,
		zapLogger,
		bobDBTrace,
		wrapperConnection,
		&repo.LessonRepo{},
		&repo.LessonMemberRepo{},
		&master_class_repo.ClassMemberRepo{},
		&lesson_report_repo.LessonReportRepo{},
	); err != nil {
		zapLogger.Fatal("RegisterStudentClassSubscriptionHandler: ", zap.Error(err))
	}

	if err := controller.RegisterReserveClassHandler(
		jsm,
		zapLogger,
		bobDBTrace,
		wrapperConnection,
		&repo.LessonRepo{},
		&repo.LessonMemberRepo{},
		&master_class_repo.ClassMemberRepo{},
		&lesson_report_repo.LessonReportRepo{},
	); err != nil {
		zapLogger.Fatal("RegisterReserveClassHandler: ", zap.Error(err))
	}

	return nil
}

func (s *server) ServerName() string {
	return "lessonmgmt"
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

func (s *server) SetupGRPC(_ context.Context, server *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	bobDBTrace := rsc.DBWith("bob")
	lessonDBTrace := rsc.DBWith("lessonmgmt")
	unleashClient := rsc.Unleash()
	wrapperConnection := support.InitWrapperDBConnector(bobDBTrace, lessonDBTrace, unleashClient, c.Common.Environment)
	jsm := rsc.NATS()
	externalConfigService := service.InitExternalConfigService(s.configurationClient, c.Zoom.SecretKey)
	zoomService := service.InitZoomService(&c.Zoom, externalConfigService, s.httpClient)
	health.RegisterHealthServer(server, &healthcheck.Service{DBBob: bobDBTrace.DB.(*pgxpool.Pool), DBLesson: lessonDBTrace.DB.(*pgxpool.Pool)})

	userModule := user.New(server, bobDBTrace, wrapperConnection, c.Common.Environment, unleashClient)
	umAdapter := &usermodadapter.UserModuleAdapter{
		Module: userModule,
	}
	mediaModule := lesson_media.New(bobDBTrace, &lesson_media_infrastructure.MediaRepo{})
	mediaModuleAdapter := &mediaadapter.MediaModuleAdapter{
		Module: mediaModule,
	}
	lessonModuleWriter := lesson.NewModuleWriter(server, wrapperConnection, jsm, umAdapter, mediaModuleAdapter, c.Common.Environment, unleashClient, zoomService, s.schedulerClient)
	assignedStudentModule := assigned_student.New(server, wrapperConnection, c.Common.Environment, unleashClient)
	lessonModuleReader := lesson.NewModuleReader(server, wrapperConnection, c.Common.Environment, unleashClient)

	pb_lesson.RegisterLessonReaderServiceServer(server, lessonModuleReader.LessonReaderService)

	pb_lesson.RegisterLessonModifierServiceServer(server, lessonModuleWriter.LessonModifierService)

	pb_lesson.RegisterAssignedStudentListServiceServer(server, assignedStudentModule.AssignedStudentGRPCService)

	pb_lesson.RegisterUserServiceServer(server, userModule.UserGRPCService)

	pb_lesson.RegisterStudentSubscriptionServiceServer(server, user_controller.NewStudentSubscriptionGRPCService(
		wrapperConnection,
		&user_infras_repo.StudentSubscriptionRepo{},
		&user_infras_repo.StudentSubscriptionAccessPathRepo{},
		&master_class_repo.ClassMemberRepo{},
		&master_class_repo.ClassRepo{},
		c.Common.Environment,
		unleashClient,
	),
	)
	courseLocationScheduleService := course_location_schedule_service.NewCourseLocationScheduleService(wrapperConnection,
		&course_location_schedule_repo.CourseLocationScheduleRepo{})

	pb_lesson.RegisterCourseLocationScheduleServiceServer(server,
		course_location_scheduling_controller.NewCourseLocationControllerController(wrapperConnection,
			courseLocationScheduleService),
	)

	pb_lesson.RegisterLessonReportModifierServiceServer(server, lesson_report_controller.NewLessonReportModifierService(
		wrapperConnection,
		&repo.LessonRepo{},
		&repo.LessonMemberRepo{},
		&lesson_report_repo.LessonReportRepo{},
		&lesson_report_repo.LessonReportDetailRepo{},
		&lesson_report_repo.PartnerFormConfigRepo{},
		&repo.ReallocationRepo{},
		lessonModuleWriter.LessonModifierService.UpdateLessonSchedulingStatus,
		unleashClient,
		c.Common.Environment,
		&repo.MasterDataRepo{},
	))

	pb_lesson.RegisterLessonReportReaderServiceServer(server,
		lesson_report_controller.NewLessonReportReaderService(wrapperConnection, &c, &bob_repo.ConfigRepo{}))

	lessonZoomController := controller_zoom.InitZoomController(&c.Zoom,
		wrapperConnection,
		externalConfigService,
		s.httpClient)
	pb_lesson.RegisterLessonZoomServiceServer(server, lessonZoomController)

	lessonZoomAccountController := controller_zoom.InitZoomAccountController(wrapperConnection, service.InitZoomService(&c.Zoom,
		externalConfigService,
		s.httpClient))
	pb_lesson.RegisterZoomAccountServiceServer(server, lessonZoomAccountController)

	pb_lesson.RegisterLessonExecutorServiceServer(server, lessonModuleWriter.LessonExecutorService)

	pb_lesson.RegisterClassroomReaderServiceServer(server, controller.NewClassroomReaderService(
		wrapperConnection,
		&repo.ClassroomRepo{},
		&repo.LessonClassroomRepo{},
		c.Common.Environment,
		unleashClient,
	))

	pb_lesson.RegisterLessonAllocationReaderServiceServer(server, allocation_controller.NewLessonAllocationReaderService(
		wrapperConnection,
		&allocation_repo.LessonAllocationRepo{},
		&academic_week_repo.AcademicWeekRepository{},
		&user_infras_repo.StudentSubscriptionRepo{},
		&user_infras_repo.StudentSubscriptionAccessPathRepo{},
		&course_location_schedule_repo.CourseLocationScheduleRepo{},
		&academic_year_repo.AcademicYearRepository{},
		&user_infras_repo.StudentCourseRepo{},
	))

	pb_lesson.RegisterClassReaderServiceServer(server, &class_controller.ClassReaderService{
		ClassUseCase: &class_usecase.ClassUseCase{
			ClassRepo:               &class_repo.ClassRepository{},
			StudentSubscriptionRepo: &user_infras_repo.StudentSubscriptionRepo{},
			CourseRepo:              &course_repo.CourseRepository{},
		},
		WrapperConnection: wrapperConnection,
	})

	classDoAccountController := classdo_controller.InitClassDoAccountController(&c.ClassDo, lessonDBTrace)
	pb_lesson.RegisterClassDoAccountServiceServer(server, classDoAccountController)
	portForwardClassDoController := classdo_controller.InitPortForwardClassDoController(&c.ClassDo, lessonDBTrace, s.httpClient)
	pb_lesson.RegisterPortForwardClassDoServiceServer(server, portForwardClassDoController)

	return nil
}

func (s *server) GracefulShutdown(context.Context) {}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	bobDBTrace := rsc.DBWith("bob")

	s.authInterceptor = authInterceptor(&c, zapLogger, bobDBTrace.DB)

	clientHTTPConfig := &clients.HTTPClientConfig{TimeOut: 5 * time.Second}
	s.httpClient = clients.InitHTTPClient(clientHTTPConfig, zapLogger)
	s.retryOptions = &configs.RetryOptions{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if c.Common.Organization != "jprep" {
		s.configurationClient = clients.InitConfigurationClient(rsc.GRPCDialContext(ctx, "mastermgmt", *s.retryOptions))
		s.schedulerClient = clients.InitSchedulerClient(rsc.GRPCDialContext(ctx, "calendar", *s.retryOptions))
	}

	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	mux := gateway.NewServeMux(
		gateway.WithOutgoingHeaderMatcher(clients.IsHeaderAllowed),
		gateway.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			token := request.Header.Get("Authorization")
			pkgHeader := request.Header.Get("pkg")
			versionHeader := request.Header.Get("version")
			if pkgHeader == "" {
				pkgHeader = "com.manabie.liz"
			}
			if versionHeader == "" {
				versionHeader = "1.0.0"
			}

			md := metadata.Pairs(
				"token", token,
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
			// using default handler to do the rest of heavy lifting of marshaling error and adding headers
			gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, &newError)
		}))
	err := setupGrpcGateway(mux, rsc.GetGRPCPort(c.Common.Name))
	if err != nil {
		return fmt.Errorf("error setupGrpcGateway %s", err)
	}
	v1 := r.Group("/lessonmgmt/api/v1")
	{
		v1.Group("/proxy/*{grpc_gateway}").Any("", gin.WrapH(mux))
	}
	return nil
}

func setupGrpcGateway(mux *gateway.ServeMux, port string) error {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(&tracer.B3Handler{ClientHandler: &ocgrpc.ClientHandler{}}),
	}

	serviceMap := map[string]func(context.Context, *gateway.ServeMux, string, []grpc.DialOption) error{
		"LessonReaderService": pb_lesson.RegisterLessonReaderServiceHandlerFromEndpoint,
	}

	for _, registerFunc := range serviceMap {
		err := registerFunc(context.Background(), mux, fmt.Sprintf("localhost%s", port), dialOpts)
		if err != nil {
			return err
		}
	}

	return nil
}
