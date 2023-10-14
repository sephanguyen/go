package mastermgmt

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/debezium"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/mongoclient"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/mastermgmt/configurations"
	academic_year_service "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/controller"
	academic_year_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure/repo"
	appsmith_commands "github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/application/commands"
	appsmith_service "github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/controller"
	appsmith_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/infrastructure/repo"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/middlewares"
	class_commands "github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/commands"
	class_queries "github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/queries"
	class_service "github.com/manabie-com/backend/internal/mastermgmt/modules/class/controller"
	class_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	config_service "github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/controller"
	config_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/infrastructure/repo"
	course_commands "github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/commands"
	course_queries "github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/queries"
	master_course "github.com/manabie-com/backend/internal/mastermgmt/modules/course/controller"
	master_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"
	custom_entity_service "github.com/manabie-com/backend/internal/mastermgmt/modules/custom_entity/controller"
	external_config_service "github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/controller"
	external_config_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/infrastructure/repo"
	grade_service "github.com/manabie-com/backend/internal/mastermgmt/modules/grade/controller"
	grade_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/grade/infrastructure/repo"
	internal_service "github.com/manabie-com/backend/internal/mastermgmt/modules/internal_service/controller"
	location_queries "github.com/manabie-com/backend/internal/mastermgmt/modules/location/application/queries"
	location_service "github.com/manabie-com/backend/internal/mastermgmt/modules/location/controller"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	org_subscription "github.com/manabie-com/backend/internal/mastermgmt/modules/location/subscriptions"
	orga_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/organization/repositories"
	orga_service "github.com/manabie-com/backend/internal/mastermgmt/modules/organization/services"
	schedule_class_service "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/controller"
	schedule_class_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure/repo"
	subject_service "github.com/manabie-com/backend/internal/mastermgmt/modules/subject/controller"
	subject_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/subject/infrastructure/repo"
	time_slot_service "github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/controller"
	time_slot_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/infrastructure/repo"
	version_control_service "github.com/manabie-com/backend/internal/mastermgmt/modules/version_control/services"
	working_hours_service "github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/controller"
	working_hours_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/infrastructure/repo"
	"github.com/manabie-com/backend/internal/mastermgmt/services"
	user_interceptors "github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	pb_fatima "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb_master "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/gin-gonic/gin"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

const localEnv, stagEnv = "local", "stag"

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
	bootstrap.DefaultMonitorService[configurations.Config]

	authInterceptor *user_interceptors.Auth
	masterDB        *pgxpool.Pool
	appsmithClient  *mongo.Client
	appsmithDB      *mongo.Database

	versionControlSvc    *version_control_service.VersionControlService
	organSvc             *orga_service.OrganizationService
	masterDataCourseSvc  *master_course.MasterDataCourseService
	courseTypeService    *master_course.CourseTypeService
	organizationModifier *org_subscription.OrganizationModifier
	gradeSrv             *grade_service.GradeService
	subjectSrv           *subject_service.SubjectService
	configSrv            *config_service.ConfigurationService
	externalConfigSrv    *external_config_service.ExternalConfigurationService
	classSvc             *class_service.ClassService
	academicYearSvc      *academic_year_service.AcademicYearService
	workingHoursSvc      *working_hours_service.WorkingHoursService
	timeSlotSvc          *time_slot_service.TimeSlotService
	customEntitySvc      *custom_entity_service.CustomEntityService
	scheduleClassSvc     *schedule_class_service.ScheduleClassService
	internalSvc          *internal_service.MasterInternalService

	locationManagementGRPCSvc *location_service.LocationManagementGRPCService
	courseAccessPathSvc       *master_course.CourseAccessPathService
	locationReaderSvc         *location_service.LocationReaderServices
	appsmithSvc               *appsmith_service.AppsmithService
	httpClient                *clients.HTTPClient
	fatimaConn                *grpc.ClientConn
}

func (*server) ServerName() string {
	return "mastermgmt"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()
	// DB must have locations table applied AC
	locationRestrictedInterceptor := locationRestrictedInterceptor(rsc.DBWith("bob").DB)
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		locationRestrictedInterceptor.UnaryServerInterceptor,
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

func (*server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	logger := rsc.Logger()
	jsm := rsc.NATS()

	bobDBTrace := rsc.DBWith("bob")
	masterDBTrace := rsc.DBWith("mastermgmt")

	s.fatimaConn = rsc.GRPCDial("fatima")

	s.authInterceptor = authInterceptor(&c, logger, bobDBTrace.DB)
	if len(c.AppsmithMongoDB.Connection) > 0 {
		s.appsmithClient, s.appsmithDB = mongoclient.GetMongoClient(ctx, logger, c.AppsmithMongoDB)
	}

	clientHTTPConfig := &clients.HTTPClientConfig{TimeOut: 60 * time.Second}
	s.httpClient = clients.InitHTTPClient(clientHTTPConfig, logger)

	subscriptionModifierServiceClient := pb_fatima.NewSubscriptionModifierServiceClient(s.fatimaConn)

	joinVersion := strings.Join(c.CheckClientVersions, ",")
	s.versionControlSvc = &version_control_service.VersionControlService{
		JoinClientVersions: joinVersion,
	}

	s.organSvc = &orga_service.OrganizationService{
		DB:               bobDBTrace,
		JSM:              jsm,
		OrganizationRepo: &orga_repo.OrganizationRepo{},
	}

	s.masterDataCourseSvc = &master_course.MasterDataCourseService{
		DB:             bobDBTrace,
		LocationRepo:   &location_repo.LocationRepo{},
		CourseTypeRepo: &master_repo.CourseTypeRepo{},
		StudentSubscriptionCommandHandler: course_queries.StudentSubscriptionQueryHandler{
			DB:                      bobDBTrace,
			StudentSubscriptionRepo: &master_repo.StudentSubscriptionRepo{},
		},
		CourseCommandHandler: course_commands.CourseCommandHandler{
			DB:                   bobDBTrace,
			CourseRepo:           &master_repo.CourseRepo{},
			CourseAccessPathRepo: &master_repo.CourseAccessPathRepo{},
		},
		CourseQueryHandler: course_queries.CourseQueryHandler{
			DB:         bobDBTrace,
			CourseRepo: &master_repo.CourseRepo{},
		},
		UnleashClientIns: rsc.Unleash(),
		Env:              c.Common.Environment,
	}

	s.courseTypeService = &master_course.CourseTypeService{
		DB:             bobDBTrace,
		CourseTypeRepo: &master_repo.CourseTypeRepo{},
		CourseTypeCommandHandler: course_commands.CourseTypeCommandHandler{
			DB:             bobDBTrace,
			CourseTypeRepo: &master_repo.CourseTypeRepo{},
		},
	}

	s.organizationModifier = &org_subscription.OrganizationModifier{
		DB:               bobDBTrace,
		Logger:           logger,
		JSM:              jsm,
		LocationTypeRepo: &location_repo.LocationTypeRepo{},
		LocationRepo:     &location_repo.LocationRepo{},
	}

	s.locationManagementGRPCSvc = location_service.NewLocationManagementGRPCService(
		bobDBTrace,
		jsm,
		&location_repo.LocationRepo{},
		&location_repo.LocationTypeRepo{},
		&location_repo.ImportLogRepo{},
		rsc.Unleash(),
		c.Common.Environment,
	)

	s.locationReaderSvc = &location_service.LocationReaderServices{
		DB:               bobDBTrace,
		LocationRepo:     &location_repo.LocationRepo{},
		LocationTypeRepo: &location_repo.LocationTypeRepo{},
		GetLocationQueryHandler: location_queries.GetLocationQueryHandler{
			DB:               bobDBTrace,
			LocationRepo:     &location_repo.LocationRepo{},
			Env:              c.Common.Environment,
			LocationTypeRepo: &location_repo.LocationTypeRepo{},
			UnleashClientIns: rsc.Unleash(),
		},
		ExportLocationQueryHandler: location_queries.ExportLocationQueryHandler{
			DB:               bobDBTrace,
			LocationRepo:     &location_repo.LocationRepo{},
			LocationTypeRepo: &location_repo.LocationTypeRepo{},
		},
	}

	s.gradeSrv = grade_service.NewGradeService(masterDBTrace, &grade_repo.GradeRepo{})
	s.subjectSrv = subject_service.NewSubjectService(bobDBTrace, &subject_repo.SubjectRepo{})
	s.configSrv = config_service.NewConfigurationService(masterDBTrace, &config_repo.ConfigRepo{}, &external_config_repo.ExternalConfigRepo{})
	s.externalConfigSrv = external_config_service.NewExternalConfigurationService(masterDBTrace, &external_config_repo.ExternalConfigRepo{})

	s.classSvc = &class_service.ClassService{
		DB:  bobDBTrace,
		JSM: jsm,
		ClassCommandHandler: class_commands.ClassCommandHandler{
			DB:        bobDBTrace,
			ClassRepo: &class_repo.ClassRepo{},
		},
		ClassQueryHandler: class_queries.ClassQueryHandler{
			DB:        bobDBTrace,
			ClassRepo: &class_repo.ClassRepo{},
		},
		ExportClassesQueryHandler: class_queries.ExportClassesQueryHandler{
			DB:        bobDBTrace,
			ClassRepo: &class_repo.ClassRepo{},
		},
		ClassRepo:    &class_repo.ClassRepo{},
		LocationRepo: &location_repo.LocationRepo{},
		CourseRepo:   &master_repo.CourseRepo{},
	}

	s.courseAccessPathSvc = master_course.NewCourseAccessPathService(bobDBTrace, &master_repo.CourseAccessPathRepo{}, &location_repo.LocationRepo{}, &master_repo.CourseRepo{})

	s.academicYearSvc = academic_year_service.NewAcademicYearService(masterDBTrace, bobDBTrace, &academic_year_repo.AcademicYearRepo{}, &academic_year_repo.AcademicWeekRepo{}, &academic_year_repo.AcademicClosedDayRepo{}, &location_repo.LocationRepo{}, &location_repo.LocationTypeRepo{}, &config_repo.ConfigRepo{})
	s.workingHoursSvc = working_hours_service.NewWorkingHoursService(masterDBTrace, bobDBTrace, &working_hours_repo.WorkingHoursRepo{}, &location_repo.LocationRepo{})
	s.timeSlotSvc = time_slot_service.NewTimeSlotService(masterDBTrace, bobDBTrace, &time_slot_repo.TimeSlotRepo{}, &location_repo.LocationRepo{})

	s.scheduleClassSvc = schedule_class_service.NewScheduleClassService(bobDBTrace, &schedule_class_repo.ReserveClassRepo{}, &schedule_class_repo.CourseRepo{}, &schedule_class_repo.ClassRepo{}, &schedule_class_repo.StudentPackageClassRepo{}, subscriptionModifierServiceClient, &class_repo.ClassMemberRepo{}, jsm)
	s.internalSvc = internal_service.NewMasterInternalService(bobDBTrace, &schedule_class_repo.ReserveClassRepo{}, &schedule_class_repo.CourseRepo{}, &schedule_class_repo.ClassRepo{}, &schedule_class_repo.StudentPackageClassRepo{}, &class_repo.ClassMemberRepo{}, jsm)
	httpClient := http.Client{Timeout: time.Duration(10) * time.Second}
	alertClient := &alert.SlackImpl{
		WebHookURL: c.AppsmithSlackWebHook,
		HTTPClient: httpClient,
	}

	if len(c.AppsmithMongoDB.Connection) > 0 {
		s.appsmithSvc = appsmith_service.NewAppsmithService(
			s.appsmithDB, &appsmith_repo.NewPageRepo{}, c.Common.Environment, c.Common.Organization, alertClient,
		)
	}
	if c.Common.Environment == stagEnv || c.Common.Environment == localEnv {
		s.customEntitySvc = &custom_entity_service.CustomEntityService{
			DB: masterDBTrace,
		}
	}

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcServer *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	bobDB := rsc.DBWith("bob")
	health.RegisterHealthServer(grpcServer, &healthcheck.Service{DB: bobDB.DB.(*pgxpool.Pool)})
	// check secondary db

	pb_master.RegisterVersionControlReaderServiceServer(grpcServer, s.versionControlSvc)
	pb_master.RegisterOrganizationServiceServer(grpcServer, s.organSvc)
	pb_master.RegisterMasterDataCourseServiceServer(grpcServer, s.masterDataCourseSvc)
	pb_master.RegisterLocationManagementGRPCServiceServer(grpcServer, s.locationManagementGRPCSvc)
	pb_master.RegisterMasterDataReaderServiceServer(grpcServer, s.locationReaderSvc)
	pb_master.RegisterCourseTypeServiceServer(grpcServer, s.courseTypeService)
	pb_master.RegisterGradeServiceServer(grpcServer, s.gradeSrv)
	pb_master.RegisterInternalServiceServer(grpcServer, s.configSrv)
	pb_master.RegisterConfigurationServiceServer(grpcServer, s.configSrv)
	pb_master.RegisterExternalConfigurationServiceServer(grpcServer, s.externalConfigSrv)
	pb_master.RegisterClassServiceServer(grpcServer, s.classSvc)
	pb_master.RegisterSubjectServiceServer(grpcServer, s.subjectSrv)
	pb_master.RegisterAcademicYearServiceServer(grpcServer, s.academicYearSvc)
	pb_master.RegisterCourseAccessPathServiceServer(grpcServer, s.courseAccessPathSvc)
	pb_master.RegisterWorkingHoursServiceServer(grpcServer, s.workingHoursSvc)
	pb_master.RegisterTimeSlotServiceServer(grpcServer, s.timeSlotSvc)
	pb_master.RegisterScheduleClassServiceServer(grpcServer, s.scheduleClassSvc)
	pb_master.RegisterMasterInternalServiceServer(grpcServer, s.internalSvc)
	if len(c.AppsmithMongoDB.Connection) > 0 {
		pb_master.RegisterAppsmithServiceServer(grpcServer, s.appsmithSvc)
	}
	if c.Common.Environment == stagEnv || c.Common.Environment == localEnv {
		pb_master.RegisterCustomEntityServiceServer(grpcServer, s.customEntitySvc)
	}

	return nil
}

func (s *server) GracefulShutdown(ctx context.Context) {
	if s.masterDB != nil {
		s.masterDB.Close()
	}
	if s.appsmithClient != nil {
		if err := s.appsmithClient.Disconnect(ctx); err != nil {
			fmt.Println("appsmithClient disconnect error", err)
		}
	}
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	organizationSubscription := services.OrganizationSubscription{
		JSM:                  rsc.NATS(),
		Logger:               rsc.Logger(),
		OrganizationModifier: s.organizationModifier,
	}
	err := organizationSubscription.Subscribe()
	if err != nil {
		return fmt.Errorf("RegisterNatsSubscribers: organizationSubscription.Subscribe: %w", err)
	}

	classRegister := class_service.RegisterSubscriber{
		JSM:             rsc.NATS(),
		Logger:          rsc.Logger(),
		DB:              rsc.DBWith("bob"),
		ClassMemberRepo: &class_repo.ClassMemberRepo{},
	}
	err = classRegister.Subscribe()
	if err != nil {
		return fmt.Errorf("RegisterNatsSubscribers: classRegister.Subscribe: %w", err)
	}

	// as source database, it will listen to incremental snapshot events which will trigger new captured table
	err = debezium.InitDebeziumIncrementalSnapshot(rsc.NATS(), rsc.Logger(), rsc.DBWith("bob"), c.Common.Name)
	if err != nil {
		return fmt.Errorf("initInternalDebeziumIncrementalSnapshot: %v", err)
	}

	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
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
			gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, &newError)
		}))
	err := setupGrpcGateway(mux, c.Common.Environment, rsc.GetGRPCPort(c.Common.Name))
	if err != nil {
		return fmt.Errorf("error setupGrpcGateway %s", err)
	}
	superGroup := r.Group("/mastermgmt/api/v1")
	{
		v1 := superGroup.Group("/appsmith")
		{
			v1.Use(middlewares.Authenticate(zapLogger))
			appsmithCtl := appsmith_service.AppsmithHTTPService{
				AppsmithAPI: c.AppsmithAPI,
				Logger:      zapLogger,
				AppsmithCommandHandler: appsmith_commands.AppsmithCommandHandler{
					LogRepo:    &appsmith_repo.LogRepo{},
					HTTPClient: s.httpClient,
				},
			}
			v1.POST("/track", appsmithCtl.Track)
			v1.GET("/pull", appsmithCtl.PullMetadata)
		}

		superGroup.Group("/proxy/*{grpc_gateway}").Any("", gin.WrapH(mux))
	}
	return nil
}

func setupGrpcGateway(mux *gateway.ServeMux, env string, port string) error {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(&tracer.B3Handler{ClientHandler: &ocgrpc.ClientHandler{}}),
	}

	serviceMap := map[string]func(context.Context, *gateway.ServeMux, string, []grpc.DialOption) error{
		"MasterDataReaderService":       pb_master.RegisterMasterDataReaderServiceHandlerFromEndpoint,
		"AppsmithService":               pb_master.RegisterAppsmithServiceHandlerFromEndpoint,
		"CustomEntityService":           pb_master.RegisterCustomEntityServiceHandlerFromEndpoint,
		"GradeService":                  pb_master.RegisterGradeServiceHandlerFromEndpoint,
		"LocationManagementGRPCService": pb_master.RegisterLocationManagementGRPCServiceHandlerFromEndpoint,
		"MasterDataCourseService":       pb_master.RegisterMasterDataCourseServiceHandlerFromEndpoint,
		"ClassService":                  pb_master.RegisterClassServiceHandlerFromEndpoint,
		"CourseTypeService":             pb_master.RegisterCourseTypeServiceHandlerFromEndpoint,
		"SubjectService":                pb_master.RegisterSubjectServiceHandlerFromEndpoint,
		"AcademicYearService":           pb_master.RegisterAcademicYearServiceHandlerFromEndpoint,
		"WorkingHoursService":           pb_master.RegisterWorkingHoursServiceHandlerFromEndpoint,
		"CourseAccessPathService":       pb_master.RegisterCourseAccessPathServiceHandlerFromEndpoint,
		"TimeSlotService":               pb_master.RegisterTimeSlotServiceHandlerFromEndpoint,
		"MasterInternalService":         pb_master.RegisterMasterDataCourseServiceHandlerFromEndpoint,
	}

	for serviceName, registerFunc := range serviceMap {
		if serviceName == "CustomEntityService" && !(env == stagEnv || env == localEnv) {
			continue
		}

		err := registerFunc(context.Background(), mux, fmt.Sprintf("localhost%s", port), dialOpts)
		if err != nil {
			return err
		}
	}

	return nil
}
