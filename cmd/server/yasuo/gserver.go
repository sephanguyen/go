package yasuo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	bobRepo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services/classes"
	enigmaRepo "github.com/manabie-com/backend/internal/enigma/repositories"
	enigmaService "github.com/manabie-com/backend/internal/enigma/services"
	utils "github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	brightcove_service "github.com/manabie-com/backend/internal/golibs/brightcove"
	"github.com/manabie-com/backend/internal/golibs/caching"
	"github.com/manabie-com/backend/internal/golibs/clients"
	firebaseLib "github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	masterClassRepo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	newNotiMetrics "github.com/manabie-com/backend/internal/notification/infra/metrics"
	"github.com/manabie-com/backend/internal/notification/mock"
	usermgmtRepo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	virtual_lesson_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	"github.com/manabie-com/backend/internal/yasuo/repositories"
	"github.com/manabie-com/backend/internal/yasuo/services"
	"github.com/manabie-com/backend/internal/yasuo/subscriptions"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bpb_v1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	firebaseV4 "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

const (
	localEnv, stagEnv string = "local", "stag"
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
	bobConn        *grpc.ClientConn
	eurekaConn     *grpc.ClientConn
	tomConn        *grpc.ClientConn
	mastermgmtConn *grpc.ClientConn
	fatimaConn     *grpc.ClientConn

	brightcoveService     *services.BrightcoveService
	userService           *services.UserService
	courseService         *services.CourseService
	schoolService         *services.SchoolService
	courseReaderService   services.CourseReaderService
	uploadReaderService   services.UploadReaderService
	uploadModifierService services.UploadModifierService

	authInterceptor *interceptors.Auth

	customMetrics      newNotiMetrics.NotificationMetrics
	notificationPusher firebaseLib.NotificationPusher
	s3Sess             *session.Session
	tenantManager      multitenant.TenantManager
	firebaseAuthClient multitenant.TenantClient
	firebaseClient     *auth.Client
}

func (s *server) WithOpencensusViews() []*view.View {
	return []*view.View{
		caching.CacheCounterView,
	}
}

func (*server) ServerName() string {
	return "yasuo"
}

func (s *server) GracefulShutdown(context.Context) {}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	dbTrace := rsc.DBWith("bob")
	lessonDB := rsc.DBWith("lessonmgmt")
	eurekaDBTrace := rsc.DBWith("eureka")
	unleashClientInstance := rsc.Unleash()
	storageConfig := rsc.Storage()

	s.authInterceptor = authInterceptor(&c, zapLogger, dbTrace.DB)

	var err error
	s.bobConn = rsc.GRPCDial("bob")
	s.eurekaConn = rsc.GRPCDial("eureka")
	s.tomConn = rsc.GRPCDial("tom")
	s.mastermgmtConn = rsc.GRPCDial("mastermgmt")
	s.fatimaConn = rsc.GRPCDial("fatima")

	firebaseProject := c.Common.FirebaseProject
	if firebaseProject == "" {
		firebaseProject = c.Common.GoogleCloudProject
	}
	firebaseAppV4, err := firebaseV4.NewApp(context.Background(), &firebaseV4.Config{
		ProjectID: firebaseProject,
	})
	if err != nil {
		return fmt.Errorf("error initializing v4 app %s", err)
	}

	firebaseClient, err := firebaseAppV4.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("error getting Auth client %s", err)
	}
	s.firebaseClient = firebaseClient

	singleTenantGCPApp, err := gcp.NewApp(context.Background(), "", firebaseProject)
	if err != nil {
		return fmt.Errorf("failed to initialize gcp app for single tenant env %s", err)
	}
	firebaseAuthClient, err := multitenant.NewFirebaseAuthClientFromGCP(context.Background(), singleTenantGCPApp)
	if err != nil {
		return fmt.Errorf("failed to initialize firebase auth client for single tenant env %s", err)
	}
	s.firebaseAuthClient = firebaseAuthClient

	identityPlatformProject := c.Common.IdentityPlatformProject
	if identityPlatformProject == "" {
		identityPlatformProject = c.Common.GoogleCloudProject
	}
	multiTenantGCPApp, err := gcp.NewApp(context.Background(), "", identityPlatformProject)
	if err != nil {
		return fmt.Errorf("failed to initialize gcp app for multi tenant env %s", err)
	}

	tenantManager, err := multitenant.NewTenantManagerFromGCP(context.Background(), multiTenantGCPApp)
	if err != nil {
		return fmt.Errorf("failed to initialize identity platform tenant manager for multi tenant env %s", err)
	}
	s.tenantManager = tenantManager

	userServiceClient := bpb.NewUserServiceClient(s.bobConn)
	subscriptionModifierServiceClient := fpb.NewSubscriptionModifierServiceClient(s.fatimaConn)
	userModifierService := services.NewUserModifierService(&c, dbTrace, rsc.NATS(), firebaseClient, firebaseAuthClient, tenantManager, subscriptionModifierServiceClient)
	s.userService = &services.UserService{
		DBPgx:               dbTrace,
		UserRepo:            &bobRepo.UserRepo{},
		TeacherRepo:         &bobRepo.TeacherRepo{},
		SchoolAdminRepo:     &bobRepo.SchoolAdminRepo{},
		SchoolRepo:          &repositories.SchoolRepo{},
		UserGroupRepo:       &bobRepo.UserGroupRepo{},
		StudentRepo:         &bobRepo.StudentRepo{},
		FirebaseClient:      firebaseClient,
		UserController:      userServiceClient,
		UserModifierService: userModifierService,
		UserGroupV2Repo:     &usermgmtRepo.UserGroupV2Repo{},
	}

	brightcoveExtService := brightcove_service.NewBrightcoveService(
		c.Brightcove.ClientID,
		c.Brightcove.Secret,
		c.Brightcove.AccountID,
		c.Brightcove.PolicyKey,
		c.Brightcove.PolicyKeyWithSearch,
		c.Brightcove.Profile,
	)
	if c.FakeBrightcoveServer != "" {
		brightcoveExtService.AccessTokenURL = c.FakeBrightcoveServer + "/v4/access_token"
		brightcoveExtService.CreateVideoURL = c.FakeBrightcoveServer + "/v1/accounts/%s/videos/"
		brightcoveExtService.UploadURLsURL = c.FakeBrightcoveServer + "/v1/accounts/%s/videos/%s/upload-urls/%s"
		brightcoveExtService.DynamicIngestURL = c.FakeBrightcoveServer + "/v1/accounts/%s/videos/%s/ingest-requests"
		brightcoveExtService.PlaybackURL = c.FakeBrightcoveServer + "/playback/v1/accounts/%s/videos/%s"
	}

	masterInternallMgmtClient := mpb.NewInternalServiceClient(s.mastermgmtConn)

	s.brightcoveService = &services.BrightcoveService{
		BrightcoveExtService:       brightcoveExtService,
		MastermgmtInternalServices: masterInternallMgmtClient,
		Env:                        c.Common.Environment,
		UnleashClientIns:           unleashClientInstance,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: storageConfig.InsecureSkipVerify, //nolint:gosec
		},
	}

	s3Sess, _ := session.NewSession(&aws.Config{
		HTTPClient:       &http.Client{Transport: tr},
		Region:           aws.String(storageConfig.Region),
		Credentials:      credentials.NewStaticCredentials(storageConfig.AccessKey, storageConfig.SecretKey, ""),
		Endpoint:         aws.String(storageConfig.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	s.s3Sess = s3Sess

	eurekaInternalModifierClient := epb.NewInternalModifierServiceClient(s.eurekaConn)

	activityLogRepo := &repositories.ActivityLogRepo{}

	s.courseService = &services.CourseService{
		Env:                            c.Common.Environment,
		EurekaDBTrace:                  eurekaDBTrace,
		Config:                         &c,
		DBTrace:                        dbTrace,
		LessonDBTrace:                  lessonDB,
		UnleashClientIns:               unleashClientInstance,
		JSM:                            rsc.NATS(),
		Logger:                         zapLogger,
		TopicQuestionPublish:           c.QuestionPublishedTopic,
		SubQuestionRenderFinish:        c.QuestionRenderedSubscriber,
		BrightCoveProfile:              c.Brightcove.Profile,
		BrightcoveService:              *s.brightcoveService,
		LimitQuestionsPullPerTime:      100,
		MaxWaitTimePullQuestion:        10 * time.Second,
		BrightcoveExtService:           brightcoveExtService,
		UserRepo:                       &bobRepo.UserRepo{},
		ChapterRepo:                    &bobRepo.ChapterRepo{},
		CourseRepo:                     &bobRepo.CourseRepo{},
		CourseAccessPathRepo:           &bobRepo.CourseAccessPathRepo{},
		CourseClassRepo:                &repositories.CourseClassRepo{},
		ClassRepo:                      &repositories.ClassRepo{},
		TeacherRepo:                    &repositories.TeacherRepo{},
		SchoolAdminRepo:                &bobRepo.SchoolAdminRepo{},
		TopicRepo:                      &repositories.TopicRepo{},
		LoRepo:                         &repositories.LearningObjectiveRepo{},
		PresetStudyPlanRepo:            &repositories.PresetStudyPlanRepo{},
		PresetStudyPlanWeeklyRepo:      &repositories.PresetStudyPlanWeeklyRepo{},
		LessonRepo:                     &repositories.LessonRepo{},
		ActivityLogRepo:                activityLogRepo,
		CourseBookRepo:                 &bobRepo.CourseBookRepo{},
		BookRepo:                       &bobRepo.BookRepo{},
		BookChapterRepo:                &bobRepo.BookChapterRepo{},
		QuizRepo:                       &bobRepo.QuizRepo{},
		QuizSetRepo:                    &bobRepo.QuizSetRepo{},
		LessonMemberRepo:               &bobRepo.LessonMemberRepo{},
		LessonGroupRepo:                &bobRepo.LessonGroupRepo{},
		Uploader:                       s3manager.NewUploader(s3Sess),
		AcademicYearRepo:               &bobRepo.AcademicYearRepo{},
		EurekaInternalModifierService:  eurekaInternalModifierClient,
		MediaModifierService:           bpb_v1.NewMediaModifierServiceClient(s.bobConn),
		SpeechesRepo:                   &repositories.SpeechesRepository{},
		TopicLearningObjectiveRepo:     &bobRepo.TopicsLearningObjectivesRepo{},
		LiveLessonSentNotificationRepo: &virtual_lesson_repo.LiveLessonSentNotificationRepo{},
	}

	s.schoolService = &services.SchoolService{
		DBTrace:         dbTrace,
		UserRepo:        &bobRepo.UserRepo{},
		SchoolRepo:      &repositories.SchoolRepo{},
		SchoolAdminRepo: &bobRepo.SchoolAdminRepo{},
		TeacherRepo:     &repositories.TeacherRepo{},
		ClassMemberRepo: &repositories.ClassMemberRepo{},
		ClassRepo:       &repositories.ClassRepo{},
		ActivityLogRepo: activityLogRepo,
		JSM:             rsc.NATS(),
	}

	s.courseReaderService = services.CourseReaderService{
		DB:               dbTrace,
		OldCourseService: s.courseService,
		UserRepo:         &bobRepo.UserRepo{},
		TeacherRepo:      &repositories.TeacherRepo{},
		SchoolAdminRepo:  &bobRepo.SchoolAdminRepo{},
	}

	s.uploadReaderService = services.UploadReaderService{
		DBTrace:  dbTrace.DB,
		Logger:   zapLogger,
		Config:   &c,
		Uploader: s3manager.NewUploader(s3Sess),
	}

	s.uploadModifierService = services.UploadModifierService{
		DBTrace:  dbTrace.DB,
		Logger:   zapLogger,
		Config:   &c,
		Uploader: s3manager.NewUploader(s3Sess),
	}

	multiTenantFCMClient, err := multiTenantGCPApp.Messaging(context.Background())
	if err != nil {
		return fmt.Errorf("firebaseApp.Messaging %s", err)
	}

	if c.Common.Environment != localEnv {
		s.notificationPusher = firebaseLib.NewNotificationPusher(multiTenantFCMClient)
	} else {
		s.notificationPusher = mock.NewNotificationPusher()
	}
	s.customMetrics = newNotiMetrics.NewClientMetrics("yasuo")

	return nil
}

func (s *server) WithPrometheusCollectors(_ *bootstrap.Resources) []prometheus.Collector {
	return s.customMetrics.GetCollectors()
}

func (s *server) InitMetricsValue() {
	s.customMetrics.GetCollectors()
}

func (s *server) SetupGRPC(_ context.Context, grpcserv *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	dbTrace := rsc.DBWith("bob")
	lessonDBTrace := rsc.DBWith("lessonmgmt")
	unleashClient := rsc.Unleash()
	zapLogger := rsc.Logger()

	pb.RegisterUserServiceServer(grpcserv, s.userService)
	ypb.RegisterBrightcoveServiceServer(grpcserv, s.brightcoveService)
	pb.RegisterCourseServiceServer(grpcserv, &services.CourseAbac{CourseService: s.courseService})
	pb.RegisterSchoolServiceServer(grpcserv, s.schoolService)
	ypb.RegisterCourseReaderServiceServer(grpcserv, &s.courseReaderService)
	ypb.RegisterUploadReaderServiceServer(grpcserv, &s.uploadReaderService)
	ypb.RegisterUploadModifierServiceServer(grpcserv, &s.uploadModifierService)
	health.RegisterHealthServer(grpcserv, &healthcheck.Service{DB: dbTrace.DB.(*pgxpool.Pool)})
	initV1Yasuo(&c, grpcserv, lessonDBTrace, dbTrace, rsc.NATS(), unleashClient, s.firebaseClient, s.firebaseAuthClient, s.tenantManager, s.fatimaConn, *rsc.Storage(), s.s3Sess, s.notificationPusher, s.courseService, s.customMetrics, zapLogger)
	return nil
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	dbTrace := rsc.DBWith("bob")
	zapLogger := rsc.Logger()
	jsm := rsc.NATS()
	configService := &services.ConfigService{
		DB:         dbTrace.DB,
		ConfigRepo: &bobRepo.ConfigRepo{},
	}
	bobClassService := &classes.ClassService{
		ClassCodeLength:       c.ClassCodeLength,
		DB:                    dbTrace,
		ClassRepo:             &bobRepo.ClassRepo{},
		ClassMemberRepo:       &bobRepo.ClassMemberRepo{},
		ConfigRepo:            &bobRepo.ConfigRepo{},
		SchoolConfigRepo:      &bobRepo.SchoolConfigRepo{},
		TeacherRepo:           &bobRepo.TeacherRepo{},
		YasuoCourseClassRepo:  &repositories.CourseClassRepo{},
		UserRepo:              &bobRepo.UserRepo{},
		JSM:                   rsc.NATS(),
		MasterClassRepo:       &masterClassRepo.ClassRepo{},
		MasterClassMemberRepo: &masterClassRepo.ClassMemberRepo{},
		CourseRepo:            &bobRepo.CourseRepo{},
	}
	partnerSyncDataLogService := &enigmaService.PartnerSyncDataLogService{
		DB:                          dbTrace,
		PartnerSyncDataLogSplitRepo: &enigmaRepo.PartnerSyncDataLogSplitRepo{},
		PartnerSyncDataLogRepo:      &enigmaRepo.PartnerSyncDataLogRepo{},
	}

	jprepSyncUserCourse := &subscriptions.JprepSyncUserCourse{
		JSM:                       jsm,
		Logger:                    zapLogger,
		CourseService:             s.courseService,
		PartnerSyncDataLogService: partnerSyncDataLogService,
		ConfigService:             configService,
	}
	err := jprepSyncUserCourse.Subscribe()
	if err != nil {
		return fmt.Errorf("jprepSyncUserCourse.Subscribe: %w", err)
	}

	jprepUserRegistration := &subscriptions.JprepUserRegistration{
		Logger:                    zapLogger,
		UserService:               s.userService,
		PartnerSyncDataLogService: partnerSyncDataLogService,
		ClassService:              bobClassService,
		JSM:                       jsm,
		ConfigService:             configService,
	}
	err = jprepUserRegistration.Subscribe()
	if err != nil {
		return fmt.Errorf("jprepUserRegistration.Subscribe: %w", err)
	}

	jprepMasterRegistration := &subscriptions.JprepMasterRegistration{
		Logger:                    zapLogger,
		CourseService:             s.courseService,
		ClassService:              bobClassService,
		PartnerSyncDataLogService: partnerSyncDataLogService,
		JSM:                       jsm,
		ConfigService:             configService,
	}
	err = jprepMasterRegistration.Subscribe()
	if err != nil {
		return fmt.Errorf("jprepMasterRegistration.Subscribe: %w", err)
	}
	return nil
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	zapLogger := rsc.Logger()
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()
	activityLogRepo := &repositories.ActivityLogRepo{}
	customs := []grpc.UnaryServerInterceptor{
		grpc_zap.PayloadUnaryServerInterceptor(zapLogger, func(ctx context.Context, fullMethod string, _ interface{}) bool {
			if !c.Common.Log.LogPayload {
				return false
			}

			// don't need to log internal APIs
			if utils.InArrayString(fullMethod, ignoreAuthEndpoint) {
				return false
			}
			return true
		}),
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		UnaryServerActivityLogRequestInterceptor(activityLogRepo, rsc.DBWith("bob"), zapLogger),
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
			gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, &newError)
		}))
	err := setupGrpcGateway(mux, c.Common.Environment, rsc.GetGRPCPort(c.Common.Name))
	if err != nil {
		return fmt.Errorf("error setupGrpcGateway %s", err)
	}
	superGroup := r.Group("/yasuo/api/v1")
	{
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
		"BrightcoveService": ypb.RegisterBrightcoveServiceHandlerFromEndpoint,
	}

	for _, registerFunc := range serviceMap {
		if !(env == stagEnv || env == localEnv) {
			continue
		}

		err := registerFunc(context.Background(), mux, fmt.Sprintf("localhost%s", port), dialOpts)
		if err != nil {
			return err
		}
	}

	return nil
}
