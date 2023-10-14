package tom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	firebaseLib "github.com/manabie-com/backend/internal/golibs/firebase"
	healthcheck "github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/metrics"
	natsjs "github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	master_interceptors "github.com/manabie-com/backend/internal/mastermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/tom/app/core"
	"github.com/manabie-com/backend/internal/tom/app/lesson"
	"github.com/manabie-com/backend/internal/tom/app/support"
	"github.com/manabie-com/backend/internal/tom/configurations"
	chatinfra "github.com/manabie-com/backend/internal/tom/infra/chat"
	"github.com/manabie-com/backend/internal/tom/infra/migration"
	"github.com/manabie-com/backend/internal/tom/mock"
	"github.com/manabie-com/backend/internal/tom/repositories"
	tomgrpc "github.com/manabie-com/backend/internal/tom/transport/grpc"
	natstransport "github.com/manabie-com/backend/internal/tom/transport/nats"
	user_interceptors "github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	pb_v1 "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	firebase "firebase.google.com/go/v4"
	lru "github.com/hashicorp/golang-lru"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithNatsServicer(s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct {
	bootstrap.DefaultMonitorService[configurations.Config]

	eurekaConn                    *grpc.ClientConn
	masterMgmtConn                *grpc.ClientConn
	authInterceptor               *user_interceptors.Auth
	locationRestrictedInterceptor *master_interceptors.LocationRestricted
	apiHandlerCollector           *metrics.PrometheusCollector

	chatInfra *chatinfra.Server

	coreChatReader  *core.ChatReader
	coreChatService *core.ChatServiceImpl

	supportChatModifier *support.ChatModifier
	supportChatReader   *support.ChatReader
	supportChatReaderV2 *support.ChatReader
	elasticIndexer      *support.SearchIndexer
	deviceTokenModifier *support.DeviceTokenModifier

	resourcePathMigrator *migration.ResourcePathMigrator

	lessonChatModifier *lesson.ChatModifier
	lessonChatReader   *lesson.ChatReader
}

func (*server) ServerName() string {
	return "tom"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()
	ignoreTracingMap := map[string]struct{}{
		"/manabie.tom.ChatService/PingSubscribeV2": {},
	}

	ignoreTracingInterceptor := interceptors.NewIgnoredTraceInterceptor(ignoreTracingMap)
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		s.locationRestrictedInterceptor.UnaryServerInterceptor,
	}

	grpcUnary := []grpc.UnaryServerInterceptor{
		ignoreTracingInterceptor.UnaryServerInterceptor,
	}
	grpcUnary = append(grpcUnary, bootstrap.DefaultUnaryServerInterceptor(rsc)...)
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
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 10 * time.Second, // Ping the client if it is idle for 10 seconds to ensure the connection is still active
		}),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (s *server) WithPrometheusCollectors(rsc *bootstrap.Resources) []prometheus.Collector {
	customMetricsCollectors := natsjs.NewClientMetrics(s.ServerName(), rsc.NATS())
	customMetricsCollectors = append(customMetricsCollectors, s.apiHandlerCollector.Collectors()...)

	return customMetricsCollectors
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	logger := rsc.Logger()
	db := rsc.DB()
	nats := rsc.NATS()
	searchClient := rsc.Elastic()
	ctx := context.Background()

	s.authInterceptor = authInterceptor(&c, logger, db.DB)
	s.locationRestrictedInterceptor = locationRestrictedInterceptor(db.DB)
	s.apiHandlerCollector = metrics.NewMetricCollector()

	err := UpsertElasticsearchFields(searchClient)
	if err != nil {
		return fmt.Errorf("InitDependencies unable to upsert elasticsearch fields: %w", err)
	}

	s.eurekaConn = rsc.GRPCDial("eureka")
	s.masterMgmtConn = rsc.GRPCDial("mastermgmt")

	externalConfigurationService := mpb.NewExternalConfigurationServiceClient(s.masterMgmtConn)

	firebaseProject := c.Common.FirebaseProject
	if firebaseProject == "" {
		firebaseProject = c.Common.GoogleCloudProject
	}

	nf := &chatinfra.Notification{
		Pusher: initPusher(ctx, logger, firebaseProject, c.Common.Environment),
	}
	lruCache, err := lru.New(c.MaxCacheEntry)
	if err != nil {
		return fmt.Errorf("InitDependencies lruCache error: %w", err)
	}

	deviceTokenRepo := &repositories.UserDeviceTokenRepo{}
	onlineUserRepo := &repositories.OnlineUserRepo{
		OnlineUserDBRepo: &repositories.OnlineUserDBRepo{},
		OnlineUserCacheRepo: &repositories.OnlineUserCacheRepo{
			LRUCache: lruCache,
		},
	}

	s.chatInfra = chatinfra.NewChatServer(
		ctx,
		c.Common.Hostname(),
		logger,
		db,
		nats,
		deviceTokenRepo,
		onlineUserRepo,
		nf,
		s.apiHandlerCollector,
	)

	s.coreChatReader = &core.ChatReader{
		Logger:                 logger,
		DB:                     db,
		ConversationRepo:       &repositories.ConversationRepo{},
		ConversationMemberRepo: &repositories.ConversationMemberRepo{},
		MessageRepo:            &repositories.MessageRepo{},
	}

	// core chat
	coreChatService := core.NewChatService(logger, s.chatInfra, s.chatInfra, db, nats)
	coreChatService.ConversationMemberRepo = &repositories.ConversationMemberRepo{}
	coreChatService.MessageRepo = &repositories.MessageRepo{}
	coreChatService.ConversationRepo = &repositories.ConversationRepo{}
	s.coreChatService = coreChatService

	supportChatModifier := support.NewChatModifier(db, coreChatService, logger)
	supportChatModifier.ConversationMemberRepo = &repositories.ConversationMemberRepo{}
	supportChatModifier.ConversationStudentRepo = &repositories.ConversationStudentRepo{}
	supportChatModifier.ConversationRepo = &repositories.ConversationRepo{}
	supportChatModifier.ConversationLocationRepo = &repositories.ConversationLocationRepo{}
	supportChatModifier.JSM = nats
	supportChatModifier.LocationRepo = &repositories.LocationRepo{}
	supportChatModifier.GrantedPermissionRepo = &repositories.GrantedPermissionsRepo{}
	supportChatModifier.UserGroupMemberRepo = &repositories.UserGroupMembersRepo{}
	s.supportChatModifier = supportChatModifier

	searchRepoV2 := &repositories.SearchRepo{}
	searchRepoV2.V2()

	locationConfigResolver := &support.LocationConfigResolver{
		DB:                           db,
		LocationRepo:                 &repositories.LocationRepo{},
		ExternalConfigurationService: externalConfigurationService,
	}

	s.supportChatReaderV2 = &support.ChatReader{
		SearchClient:                 searchClient,
		DB:                           db,
		Logger:                       logger,
		ConversationMemberRepo:       &repositories.ConversationMemberRepo{},
		ConversationStudentRepo:      &repositories.ConversationStudentRepo{},
		ConversationRepo:             &repositories.ConversationRepo{},
		MessageRepo:                  &repositories.MessageRepo{},
		ConversationSearchRepo:       searchRepoV2,
		LocationRepo:                 &repositories.LocationRepo{},
		ExternalConfigurationService: externalConfigurationService,
		UnleashClientIns:             rsc.Unleash(),
		Env:                          c.Common.Environment,
		LocationConfigResolver:       locationConfigResolver,
	}

	s.supportChatReader = &support.ChatReader{
		SearchClient:                 searchClient,
		DB:                           db,
		Logger:                       logger,
		ConversationMemberRepo:       &repositories.ConversationMemberRepo{},
		ConversationStudentRepo:      &repositories.ConversationStudentRepo{},
		ConversationRepo:             &repositories.ConversationRepo{},
		MessageRepo:                  &repositories.MessageRepo{},
		ConversationSearchRepo:       &repositories.SearchRepo{},
		LocationRepo:                 &repositories.LocationRepo{},
		ExternalConfigurationService: externalConfigurationService,
		ConversationLocationRepo:     &repositories.ConversationLocationRepo{},
		UnleashClientIns:             rsc.Unleash(),
		Env:                          c.Common.Environment,
		LocationConfigResolver:       locationConfigResolver,
	}

	s.lessonChatReader = lesson.NewLessonChatReader(db)

	s.deviceTokenModifier = &support.DeviceTokenModifier{
		DB:                       db,
		JSM:                      nats,
		Logger:                   logger,
		UserDeviceTokenRepo:      &repositories.UserDeviceTokenRepo{},
		ConversationStudentRepo:  &repositories.ConversationStudentRepo{},
		ConversationRepo:         &repositories.ConversationRepo{},
		ConversationLocationRepo: &repositories.ConversationLocationRepo{},
		ConversationMemberRepo:   &repositories.ConversationMemberRepo{},
		GrantedPermissionRepo:    &repositories.GrantedPermissionsRepo{},
	}

	s.resourcePathMigrator = &migration.ResourcePathMigrator{
		DB:                     db,
		ConversationLessonRepo: &repositories.ConversationLessonRepo{},
		UserDeviceTokenRepo:    &repositories.UserDeviceTokenRepo{},
		ConversationRepo:       &repositories.ConversationRepo{},
	}

	s.lessonChatModifier = &lesson.ChatModifier{
		ConversationMemberRepo:        &repositories.ConversationMemberRepo{},
		ConversationRepo:              &repositories.ConversationRepo{},
		ConversationLessonRepo:        &repositories.ConversationLessonRepo{},
		PrivateConversationLessonRepo: &repositories.PrivateConversationLessonRepo{},
		MessageRepo:                   &repositories.MessageRepo{},
		UserRepo:                      &repositories.UsersRepo{},
		Logger:                        logger,
		DB:                            db,
		ChatService:                   coreChatService,
		ChatInfra:                     s.chatInfra,
	}

	eurekaCourseReaderClient := epb.NewCourseReaderServiceClient(s.eurekaConn)
	s.elasticIndexer = &support.SearchIndexer{
		SearchFactory:             searchClient,
		EurekaCourseReaderService: eurekaCourseReaderClient,
		ChatReader:                s.supportChatReader,
		DB:                        db,
		SearchRepo:                searchRepoV2,
		MessageRepo:               &repositories.MessageRepo{},
		ConversationMemberRepo:    &repositories.ConversationMemberRepo{},
		ConversationRepo:          &repositories.ConversationRepo{},
		ConversationStudentRepo:   &repositories.ConversationStudentRepo{},
		ConversationLocationRepo:  &repositories.ConversationLocationRepo{},
		Logger:                    logger,
	}
	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserver *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	genProtoChatService := tomgrpc.GenprotoChatService{
		Chat:              s.coreChatService,
		ChatReader:        s.coreChatReader,
		ChatInfra:         s.chatInfra,
		SupportChatReader: s.supportChatReader,
	}

	manabufChatReader := tomgrpc.ManabufV1ChatReader{
		SupportChatReader:   s.supportChatReader,
		SupportChatReaderV2: s.supportChatReaderV2,
		CoreChatReader:      s.coreChatReader,
	}

	// lessonChatReader := lesson.NewLessonChatReader(db)
	manabufConversationReader := tomgrpc.ManabufV1ConversationReader{
		SupportChatReader: s.supportChatReader,
		LessonChatReader:  s.lessonChatReader,
	}

	manbufChatModifier := &tomgrpc.ManabufV1ChatModifier{
		SupportChatModifier: s.supportChatModifier,
		Chat:                s.coreChatService,
	}

	health.RegisterHealthServer(grpcserver, &healthcheck.Service{DB: rsc.DB().DB.(*pgxpool.Pool)})
	pb.RegisterChatServiceServer(grpcserver, &genProtoChatService)
	pb_v1.RegisterChatModifierServiceServer(grpcserver, manbufChatModifier)
	pb_v1.RegisterChatReaderServiceServer(grpcserver, &manabufChatReader)
	pb_v1.RegisterConversationReaderServiceServer(grpcserver, &manabufConversationReader)
	pb_v1.RegisterLessonChatReaderServiceServer(grpcserver, s.lessonChatReader)
	pb_v1.RegisterLessonChatModifierServiceServer(grpcserver, s.lessonChatModifier)

	return nil
}

func (s *server) GracefulShutdown(ctx context.Context) {
	s.chatInfra.HubStop()
	s.chatInfra.DeleteOnlineUser(ctx)
}

func initPusher(ctx context.Context, zapLogger *zap.Logger, firebaseProject string, env string) (n chatinfra.Pusher) {
	if env != "local" {
		firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{
			ProjectID: firebaseProject,
		})
		if err != nil {
			zapLogger.Fatal("failed to initialize Firebase App", zap.Error(err))
		}
		fcmClient, err := firebaseApp.Messaging(ctx)
		if err != nil {
			zapLogger.Fatal("failed to get FCM client firebaseApp.Messaging", zap.Error(err))
		}
		n = firebaseLib.NewNotificationPusher(fcmClient)
		zapLogger.Info("Init Firebase Cloud Messaging successfully!")
		return
	}
	n = mock.NewNotifier()
	zapLogger.Info("init mock fcm successfully in local env")
	return
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	jsm := rsc.NATS()
	messageChatModifierSubscription := natstransport.MessageChatModifierSubscription{
		HostName:  c.Common.Hostname(),
		JSM:       jsm,
		Logger:    zapLogger,
		ChatInfra: s.chatInfra,
	}
	err := messageChatModifierSubscription.SubscribeMessageChatCreated()
	if err != nil {
		return fmt.Errorf("sendMessageChatSubscription.Subscribe: %w", err)
	}

	err = messageChatModifierSubscription.SubscribeMessageChatDeleted()
	if err != nil {
		return fmt.Errorf("messageChatDeletedSubscription.Subscribe: %w", err)
	}

	userInfoSubscription := natstransport.UserUpdateSubscription{
		JSM:                 jsm,
		Logger:              zapLogger,
		DeviceTokenModifier: s.deviceTokenModifier,
		ChatModifier:        s.supportChatModifier,
	}
	err = userInfoSubscription.Subscribe()
	if err != nil {
		return fmt.Errorf("userInfoSubscription.Subscribe: %w", err)
	}
	resourcePathMigratorSub := natstransport.ResourcePathMigration{
		Config:   &c,
		JSM:      jsm,
		Logger:   zapLogger,
		Migrator: s.resourcePathMigrator,
	}
	err = resourcePathMigratorSub.Subscribe()
	if err != nil {
		return fmt.Errorf("resourcepathMigrator.Subscribe: %w", err)
	}

	lessonSubscription := natstransport.LessonEventSubscription{
		Config:             &c,
		Logger:             zapLogger,
		LessonChatModifier: s.lessonChatModifier,
		JSM:                jsm,
	}
	err = lessonSubscription.Subscribe()
	if err != nil {
		return fmt.Errorf("lessonSubscription.Subscribe: %w", err)
	}

	studentLessonSub := natstransport.StudentLessonsSubscriptions{
		Config:             &c,
		JSM:                jsm,
		Logger:             zapLogger,
		LessonChatModifier: s.lessonChatModifier,
	}
	err = studentLessonSub.Subscribe()
	if err != nil {
		return fmt.Errorf("studentLessonSubs.Subscribe: %w", err)
	}

	studentConversationSubscription := natstransport.UserEventSubscription{
		Config:       &c,
		Logger:       zapLogger,
		ChatModifier: s.supportChatModifier,
		JSM:          jsm,
	}
	err = studentConversationSubscription.Subscribe()
	if err != nil {
		return fmt.Errorf("studentConversationSubscription.Subscribe: %w", err)
	}

	esindexSub := natstransport.ElasticsearchReindexSubscription{
		Config:  &c,
		Logger:  zapLogger,
		JSM:     jsm,
		Indexer: s.elasticIndexer,
	}
	err = esindexSub.SubscribeConversationInternal()
	if err != nil {
		return fmt.Errorf("esindexSub.SubscribeConversationInternal: %w", err)
	}

	err = esindexSub.SubscribeCourseStudent()
	if err != nil {
		return fmt.Errorf("esindexSub.SubscribeCourseStudent: %w", err)
	}

	staffSubscription := natstransport.StaffUpsertSubscription{
		ChatModifier: s.supportChatModifier,
		JSM:          jsm,
		Logger:       zapLogger,
	}

	err = staffSubscription.Subscribe()
	if err != nil {
		return fmt.Errorf("staffSubscription.Subscribe: %w", err)
	}

	userGroupSubscription := natstransport.UserGroupUpsertSubscription{
		ChatModifier: s.supportChatModifier,
		JSM:          jsm,
		Logger:       zapLogger,
	}

	err = userGroupSubscription.Subscribe()
	if err != nil {
		return fmt.Errorf("userGroupSubscription.Subscribe: %w", err)
	}

	return nil
}
