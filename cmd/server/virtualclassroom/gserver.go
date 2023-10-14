package virtualclassroom

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/metrics"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/mediaadapter"
	lesson_media "github.com/manabie-com/backend/internal/lessonmgmt/modules/media"
	lesson_media_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/healthcheck"
	lr_commands "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/commands"
	lr_queries "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	lr_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/controller"
	lr_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure/repo"
	logger_svc_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	logger_svc_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/middlewares"
	vl_queries "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries"
	vl_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/controller"
	zg_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/zegocloud/controller"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func init() {
	s := &server{}
	bootstrap.WithGRPC[configurations.Config](s).
		WithNatsServicer(s).
		WithHTTP(s).
		WithMonitorServicer(s).
		Register(s)
}

var _ bootstrap.GRPCServicer[configurations.Config] = (*server)(nil) // ensure grpc-safety

type server struct {
	authInterceptor *interceptors.Auth
	fireStoreAgora  filestore.FileStore
	bootstrap.DefaultMonitorService[configurations.Config]

	bobDB             *database.DBTrace
	lessonmgmtDB      *database.DBTrace
	wrapperConnection *support.WrapperDBConnection

	retryOptions       *configs.RetryOptions
	conversationClient *clients.ConversationClient
	mediaModuleAdapter *mediaadapter.MediaModuleAdapter

	apiHandlerCollector *metrics.PrometheusCollector
}

func (*server) ServerName() string {
	return "virtualclassroom"
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	bobDB := s.bobDB
	lessonmgmtDB := s.lessonmgmtDB
	wrapperConnection := s.wrapperConnection
	mediaModuleAdapter := s.mediaModuleAdapter
	zapLogger := rsc.Logger()
	jsm := rsc.NATS()
	whiteboardSvc := whiteboard.New(&c.Whiteboard)

	if err := controller.RegisterLessonDeletedSubscriber(jsm, zapLogger, wrapperConnection, &repo.RecordedVideoRepo{}, s.fireStoreAgora, mediaModuleAdapter, c); err != nil {
		zapLogger.Fatal("registerLessonDeletedSubscriber: ", zap.Error(err))
	}

	if err := controller.RegisterLessonDefaultChatStateSubscriber(jsm, zapLogger, wrapperConnection, &repo.LessonMemberRepo{}); err != nil {
		zapLogger.Fatal("registerLessonDefaultChatStateSubscriber: ", zap.Error(err))
	}

	if err := controller.RegisterCreateLiveLessonRoomSubscriber(jsm, zapLogger, wrapperConnection, &repo.VirtualLessonRepo{}, whiteboardSvc); err != nil {
		zapLogger.Fatal("RegisterCreateLiveLessonRoomSubscriber: ", zap.Error(err))
	}

	if err := controller.RegisterUpcomingLiveLessonNotificationSubscriptionHandler(jsm, zapLogger, bobDB, wrapperConnection, &repo.VirtualLessonRepo{}, &repo.LiveLessonSentNotificationRepo{}, &repo.LessonMemberRepo{}, &repo.StudentParentRepo{}, &repo.UserRepo{}, c.Common.Environment, rsc.Unleash()); err != nil {
		zapLogger.Fatal("RegisterUpcomingLiveLessonNotificationSubscriptionHandler: ", zap.Error(err))
	}

	if err := lr_controller.RegisterLiveRoomSubscriber(jsm, zapLogger, lessonmgmtDB, &lr_repo.LiveRoomMemberStateRepo{}); err != nil {
		zapLogger.Fatal("RegisterLiveRoomSubscriber: ", zap.Error(err))
	}

	if err := controller.RegisterLessonUpdatedSubscriptionHandler(jsm, zapLogger, wrapperConnection, &repo.LiveLessonSentNotificationRepo{}); err != nil {
		zapLogger.Fatal("RegisterLessonUpdatedSubscriptionHandler: ", zap.Error(err))
	}

	return nil
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

func (s *server) WithPrometheusCollectors(rsc *bootstrap.Resources) []prometheus.Collector {
	return s.apiHandlerCollector.Collectors()
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	s.bobDB = rsc.DBWith("bob")
	s.lessonmgmtDB = rsc.DBWith("lessonmgmt")
	s.wrapperConnection = support.InitWrapperDBConnector(s.bobDB, s.lessonmgmtDB, rsc.Unleash(), c.Common.Environment)
	s.authInterceptor = authInterceptor(&c, rsc.Logger(), s.bobDB.DB)
	s.retryOptions = &configs.RetryOptions{}
	storageConfig := rsc.Storage()

	// Initialize file storage object.
	var (
		fs  filestore.FileStore
		err error
	)
	st := &configs.StorageConfig{
		Endpoint: storageConfig.Endpoint,
		Bucket:   storageConfig.Bucket,
	}
	if c.Agora.BucketName != "" {
		st.Bucket = c.Agora.BucketName
	}
	if c.Common.Environment != "stag" {
		st = &configs.StorageConfig{
			Endpoint:  storageConfig.Endpoint,
			Bucket:    storageConfig.Bucket,
			AccessKey: storageConfig.AccessKey,
			SecretKey: storageConfig.SecretKey,
		}
		if c.Agora.BucketName != "" && c.Agora.BucketAccessKey != "" && c.Agora.BucketSecretKey != "" {
			st.Bucket = c.Agora.BucketName
			st.AccessKey = c.Agora.BucketAccessKey
			st.SecretKey = c.Agora.BucketSecretKey
		}
	}
	if strings.Contains(c.Storage.Endpoint, "minio") {
		fs, err = filestore.NewFileStore(filestore.MinIOService, c.Common.ServiceAccountEmail, st)
		if err != nil {
			return fmt.Errorf("failed to init MinIO file storage %s", err)
		}
	} else {
		fs, err = filestore.NewFileStore(filestore.GoogleCloudStorageService, c.Common.ServiceAccountEmail, st)
		if err != nil {
			return fmt.Errorf("failed to init Google file storage %s", err)
		}
	}
	s.fireStoreAgora = fs
	s.apiHandlerCollector = metrics.NewMetricCollector()

	mediaModule := lesson_media.New(s.bobDB, &lesson_media_infrastructure.MediaRepo{})
	s.mediaModuleAdapter = &mediaadapter.MediaModuleAdapter{
		Module: mediaModule,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if c.Common.Environment != "uat" && c.Common.Environment != "prod" {
		s.conversationClient = clients.InitConversationClient(rsc.GRPCDialContext(ctx, "conversationmgmt", *s.retryOptions))
	}

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcServer *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	jsm := rsc.NATS()
	logger := rsc.Logger()
	lessonmgmtDB := s.lessonmgmtDB
	wrapperConnection := s.wrapperConnection
	mediaModuleAdapter := s.mediaModuleAdapter

	health.RegisterHealthServer(grpcServer, &healthcheck.Service{})
	vCrLog := &logger_svc_controller.VirtualClassRoomLogService{
		WrapperConnection: wrapperConnection,
		Repo:              &logger_svc_repo.VirtualClassroomLogRepo{},
	}
	liveRoomLog := &logger_svc_controller.LiveRoomLogService{
		DB:              lessonmgmtDB,
		LiveRoomLogRepo: &logger_svc_repo.LiveRoomLogRepo{},
	}
	whiteboardSvc := whiteboard.New(&c.Whiteboard)
	agoraTokenSvc := &controller.AgoraTokenService{
		AgoraCfg: c.Agora,
	}

	// Commands
	liveLessonCommand := &commands.LiveLessonCommand{
		WrapperDBConnection:      wrapperConnection,
		VideoTokenSuffix:         c.Agora.VideoTokenSuffix,
		MaximumLearnerStreamings: c.Agora.MaximumLearnerStreamings,
		WhiteboardSvc:            whiteboardSvc,
		AgoraTokenSvc:            agoraTokenSvc,
		VirtualLessonRepo:        &repo.VirtualLessonRepo{},
		LessonMemberRepo:         &repo.LessonMemberRepo{},
		ActivityLogRepo:          &repo.ActivityLogRepo{},
		StudentsRepo:             &repo.StudentsRepo{},
		CourseRepo:               &repo.CourseRepo{},
	}

	recordingCommand := commands.RecordingCommand{
		LessonmgmtDB:               lessonmgmtDB,
		WrapperDBConnection:        wrapperConnection,
		LessonRoomStateRepo:        &repo.LessonRoomStateRepo{},
		RecordedVideoRepo:          &repo.RecordedVideoRepo{},
		MediaModulePort:            mediaModuleAdapter,
		LiveRoomStateRepo:          &lr_repo.LiveRoomStateRepo{},
		LiveRoomRecordedVideosRepo: &lr_repo.LiveRoomRecordedVideosRepo{},
	}

	liveRoomCommand := &lr_commands.LiveRoomCommand{
		LessonmgmtDB:             lessonmgmtDB,
		WrapperDBConnection:      wrapperConnection,
		VideoTokenSuffix:         c.Agora.VideoTokenSuffix,
		WhiteboardAppID:          c.Whiteboard.AppID,
		MaximumLearnerStreamings: c.Agora.MaximumLearnerStreamings,
		WhiteboardSvc:            whiteboardSvc,
		AgoraTokenSvc:            agoraTokenSvc,
		LiveRoomRepo:             &lr_repo.LiveRoomRepo{},
		LiveRoomStateRepo:        &lr_repo.LiveRoomStateRepo{},
		LiveRoomActivityLogRepo:  &lr_repo.LiveRoomActivityLogRepo{},
		StudentsRepo:             &repo.StudentsRepo{},
		LessonRepo:               &repo.VirtualLessonRepo{},
	}

	chatServiceCommand := commands.ChatServiceCommand{
		LessonmgmtDB:               lessonmgmtDB,
		ConversationClient:         s.conversationClient,
		LiveLessonConversationRepo: &repo.LiveLessonConversationRepo{},
	}

	// Queries
	lessonRoomStateQuery := queries.LessonRoomStateQuery{
		WrapperDBConnection: wrapperConnection,
		VirtualLessonRepo:   &repo.VirtualLessonRepo{},
		LessonRoomStateRepo: &repo.LessonRoomStateRepo{},
		LessonMemberRepo:    &repo.LessonMemberRepo{},
		MediaModulePort:     mediaModuleAdapter,
		StudentsRepo:        &repo.StudentsRepo{},
	}

	userInfoQuery := queries.UserInfoQuery{
		WrapperDBConnection: wrapperConnection,
		UserBasicInfoRepo:   &repo.UserBasicInfoRepo{},
	}

	recordingQuery := queries.RecordedVideoQuery{
		WrapperDBConnection: wrapperConnection,
		RecordedVideoRepo:   &repo.RecordedVideoRepo{},
		MediaModulePort:     mediaModuleAdapter,
	}

	organizationQuery := queries.OrganizationQuery{
		WrapperDBConnection: wrapperConnection,
		OrganizationRepo:    &repo.OrganizationRepo{},
	}

	virtualLessonQuery := vl_queries.VirtualLessonQuery{
		LessonmgmtDB:                 lessonmgmtDB,
		WrapperDBConnection:          wrapperConnection,
		VirtualLessonRepo:            &repo.VirtualLessonRepo{},
		LessonMemberRepo:             &repo.LessonMemberRepo{},
		LessonTeacherRepo:            &repo.LessonTeacherRepo{},
		StudentEnrollmentHistoryRepo: &repo.StudentEnrollmentStatusHistoryRepo{},
		CourseClassRepo:              &repo.CourseClassRepo{},
		OldClassRepo:                 &repo.OldClassRepo{},
		StudentsRepo:                 &repo.StudentsRepo{},
		ConfigRepo:                   &repo.ConfigRepo{},
	}

	liveRoomStateQuery := lr_queries.LiveRoomStateQuery{
		LessonmgmtDB:            lessonmgmtDB,
		LiveRoomRepo:            &lr_repo.LiveRoomRepo{},
		LiveRoomStateRepo:       &lr_repo.LiveRoomStateRepo{},
		LiveRoomMemberStateRepo: &lr_repo.LiveRoomMemberStateRepo{},
		MediaModulePort:         mediaModuleAdapter,
	}

	// Controller / Service
	vcReaderSvc := &controller.VirtualClassroomReaderService{
		Cfg:                        c,
		LiveLessonCommand:          liveLessonCommand,
		LessonRoomStateQuery:       lessonRoomStateQuery,
		UserInfoQuery:              userInfoQuery,
		VirtualClassRoomLogService: vCrLog,
	}

	vcModifierSvc := &controller.VirtualClassroomModifierService{
		WrapperDBConnection:        wrapperConnection,
		JSM:                        jsm,
		Cfg:                        c,
		LiveLessonCommand:          liveLessonCommand,
		VirtualLessonRepo:          &repo.VirtualLessonRepo{},
		StudentsRepo:               &repo.StudentsRepo{},
		LessonGroupRepo:            &repo.LessonGroupRepo{},
		LessonMemberRepo:           &repo.LessonMemberRepo{},
		VirtualLessonPollingRepo:   &repo.VirtualLessonPollingRepo{},
		LessonRoomStateRepo:        &repo.LessonRoomStateRepo{},
		VirtualClassRoomLogService: vCrLog,
	}
	vcModifierSvc.RegisterMetric(s.apiHandlerCollector)

	vcChatSvc := &controller.VirtualClassroomChatService{
		ChatServiceCommand: chatServiceCommand,
		Logger:             logger,
	}

	lRecordingSvc := &controller.LessonRecordingService{
		Cfg:                  c,
		Logger:               logger,
		RecordingCommand:     recordingCommand,
		LessonRoomStateQuery: lessonRoomStateQuery,
		LiveRoomStateQuery:   liveRoomStateQuery,
		RecordingQuery:       recordingQuery,
		OrganizationQuery:    organizationQuery,
		FileStore:            s.fireStoreAgora,
	}

	vlReaderSvc := &vl_controller.VirtualLessonReaderService{
		WrapperDBConnection: wrapperConnection,
		JSM:                 jsm,
		Env:                 c.Common.Environment,
		VirtualLessonQuery:  virtualLessonQuery,
		VirtualLessonRepo:   &repo.VirtualLessonRepo{},
		LessonGroupRepo:     &repo.LessonGroupRepo{},
		UnleashClient:       rsc.Unleash(),
	}

	lrModifierSvc := &lr_controller.LiveRoomModifierService{
		LessonmgmtDB:            lessonmgmtDB,
		WrapperDBConnection:     wrapperConnection,
		JSM:                     jsm,
		Cfg:                     c,
		LiveRoomLogService:      liveRoomLog,
		LiveRoomCommand:         liveRoomCommand,
		LiveRoomStateQuery:      liveRoomStateQuery,
		StudentsRepo:            &repo.StudentsRepo{},
		LiveRoomStateRepo:       &lr_repo.LiveRoomStateRepo{},
		LiveRoomMemberStateRepo: &lr_repo.LiveRoomMemberStateRepo{},
		LiveRoomPoll:            &lr_repo.LiveRoomPollRepo{},
	}

	lrReaderSvc := &lr_controller.LiveRoomReaderService{
		LiveRoomCommand:    liveRoomCommand,
		LiveRoomStateQuery: liveRoomStateQuery,
		LiveRoomLogService: liveRoomLog,
	}

	zgSvc := &zg_controller.ZegoCloudService{
		ZegoCloudCfg: c.ZegoCloudConfig,
	}

	// Register to GRPC
	vpb.RegisterVirtualClassroomReaderServiceServer(grpcServer, vcReaderSvc)
	vpb.RegisterVirtualClassroomModifierServiceServer(grpcServer, vcModifierSvc)
	vpb.RegisterVirtualClassroomChatServiceServer(grpcServer, vcChatSvc)
	vpb.RegisterLessonRecordingServiceServer(grpcServer, lRecordingSvc)
	vpb.RegisterVirtualLessonReaderServiceServer(grpcServer, vlReaderSvc)
	vpb.RegisterLiveRoomModifierServiceServer(grpcServer, lrModifierSvc)
	vpb.RegisterLiveRoomReaderServiceServer(grpcServer, lrReaderSvc)
	vpb.RegisterZegoCloudServiceServer(grpcServer, zgSvc)

	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	bobDBTrace := rsc.DBWith("bob")
	lessonmgmtDBTrace := rsc.DBWith("lessonmgmt")
	wrapperConnection := support.InitWrapperDBConnector(bobDBTrace, lessonmgmtDBTrace, rsc.Unleash(), c.Common.Environment)

	r.Use(tracingMiddleware)

	v1 := r.Group("/api/virtualclassroom/v1")
	v1.Use(
		middlewares.VerifySignature(
			middlewares.AgoraHeaderKey,
			c.Agora.CallbackSignature,
		))

	agoraCallbackController := controller.AgoraCallbackService{
		Cfg:    c,
		Logger: zapLogger,
		RecordingCommand: commands.RecordingCommand{
			LessonmgmtDB:        lessonmgmtDBTrace,
			WrapperDBConnection: wrapperConnection,
			LessonRoomStateRepo: &repo.LessonRoomStateRepo{},
			LiveRoomStateRepo:   &lr_repo.LiveRoomStateRepo{},
		},
		LessonRoomStateQuery: queries.LessonRoomStateQuery{
			WrapperDBConnection: wrapperConnection,
			LessonRoomStateRepo: &repo.LessonRoomStateRepo{},
		},
		LiveRoomStateQuery: lr_queries.LiveRoomStateQuery{
			LessonmgmtDB:      lessonmgmtDBTrace,
			LiveRoomStateRepo: &lr_repo.LiveRoomStateRepo{},
		},
		OrganizationQuery: queries.OrganizationQuery{
			WrapperDBConnection: wrapperConnection,
			OrganizationRepo:    &repo.OrganizationRepo{},
		},
	}
	v1.POST("/agora-callback", agoraCallbackController.CallBack)
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

func (*server) GracefulShutdown(context.Context) {}
