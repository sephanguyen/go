package bob

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/bob/services/classes"
	"github.com/manabie-com/backend/internal/bob/services/filestore"
	lesson_reports "github.com/manabie-com/backend/internal/bob/services/lesson_reports"
	lessons "github.com/manabie-com/backend/internal/bob/services/lessons"
	log_service "github.com/manabie-com/backend/internal/bob/services/log"
	master_data "github.com/manabie-com/backend/internal/bob/services/master_data"
	"github.com/manabie-com/backend/internal/bob/services/media"
	students "github.com/manabie-com/backend/internal/bob/services/students"
	uploads "github.com/manabie-com/backend/internal/bob/services/uploads"
	"github.com/manabie-com/backend/internal/bob/services/users"
	bob_support "github.com/manabie-com/backend/internal/bob/support"
	"github.com/manabie-com/backend/internal/golibs"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	brightcove_service "github.com/manabie-com/backend/internal/golibs/brightcove"
	"github.com/manabie-com/backend/internal/golibs/caching"
	"github.com/manabie-com/backend/internal/golibs/cloudconvert"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/debezium"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	metrics "github.com/manabie-com/backend/internal/golibs/metrics"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	lesson_allocation_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/infrastructure/repo"
	cls_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/mediaadapter"
	lessonmgmt_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/usermodadapter"
	lesson_media "github.com/manabie-com/backend/internal/lessonmgmt/modules/media"
	lesson_media_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user"
	user_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	masterClassRepo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	location_queries "github.com/manabie-com/backend/internal/mastermgmt/modules/location/application/queries"
	location_service "github.com/manabie-com/backend/internal/mastermgmt/modules/location/controller"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	notificationRepo "github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	yasuo_repo "github.com/manabie-com/backend/internal/yasuo/repositories"
	yasuo_service "github.com/manabie-com/backend/internal/yasuo/services"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb_v1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	shamir "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	firebaseV4 "firebase.google.com/go/v4"
	"github.com/awa/go-iap/appstore"
	"github.com/dgraph-io/ristretto"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
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
	// checkAppVersionUnary  grpc.UnaryServerInterceptor
	// checkAppVersionStream grpc.StreamServerInterceptor

	cacheWrapper  *caching.RistrettoWrapper
	whiteboardSvc *whiteboard.Service

	fs                  filestore.FileStore
	authInterceptor     *interceptors.Auth
	apiHandlerCollector *metrics.PrometheusCollector

	eurekaDB     *database.DBTrace
	eurekaConn   *grpc.ClientConn
	usermgmtConn *grpc.ClientConn
	shamirConn   *grpc.ClientConn

	courseSvc *services.CourseService

	userSvc         *services.UserService
	userModifierSvc *users.UserModifierService
	userReaderSvc   *users.UserReaderService
	studentSvc      *services.StudentServiceABAC

	internalSvc         *services.InternalService
	internalReaderSvc   *services.InternalReaderService
	internalModifierSvc *services.InternalModifierService
	internalSvcCacher   *services.InternalServiceCacher

	lessonModifierSvc     *lessons.LessonModifierServices
	lessonReaderSvc       *lessons.LessonReaderServices
	lessonReportReaderSvc *lesson_reports.LessonReportReaderService

	masterDataSvc           *services.MasterDataService
	masterDataImporterSvc   *master_data.MasterDataImporterService
	masterDataReaderService *master_data.MasterDataReaderService

	classServiceABAC   *classes.ClassServiceABAC
	classReaderService *classes.ClassReaderService
	classModifierSvc   *classes.ClassModifierService

	cloudConvertSvc *cloudconvert.Service
	pgUserSvc       *services.PostgresUserService
	pgNamespaceSvc  *services.PostgresNamespaceService
}

func (s *server) WithOpencensusViews() []*view.View {
	return []*view.View{
		caching.CacheCounterView,
	}
}

func (s *server) ServerName() string {
	return "bob"
}

func (s *server) GracefulShutdown(context.Context) {}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	dbTrace := rsc.DB()
	unleashClientInstance := rsc.Unleash()
	storageConfig := rsc.Storage()

	eurekaDBTrace := rsc.DBWith("eureka")
	s.eurekaDB = eurekaDBTrace

	s.eurekaConn = rsc.GRPCDial("eureka")
	s.usermgmtConn = rsc.GRPCDial("usermgmt")
	s.shamirConn = rsc.GRPCDial("shamir")

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

	gcpApp, err := gcp.NewApp(context.Background(), "", firebaseProject)
	if err != nil {
		return fmt.Errorf("error init GCP app %s", err)
	}
	tenantManager, err := internal_auth_tenant.NewTenantManagerFromGCP(context.Background(), gcpApp)
	if err != nil {
		return fmt.Errorf("error init tenantManager %s", err)
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,
		MaxCost:     1e5,
		BufferItems: 1 << 6,
		Cost: func(value interface{}) int64 { // need to find a better way to handle
			return 1
		},
	})
	if err != nil {
		return fmt.Errorf("error when init cache %s", err)
	}

	s.cacheWrapper = &caching.RistrettoWrapper{
		RistrettoCacher: cache,
	}

	s.authInterceptor = authInterceptor(&c, zapLogger, dbTrace)

	// // join CheckClientVersions map as a comma delemeted string
	// joinedCheckClientVersions := strings.Join(c.CheckClientVersions, ",")

	// // parse appversion to validate
	// checkAppVersion, err := gl_interceptors.NewCheckAppVersion(joinedCheckClientVersions, "1.3.0", ignoreAuthEndpoint)
	// if err != nil {
	// 	return fmt.Errorf("err init NewCheckAppVersion %s", err)
	// }
	// s.checkAppVersionStream = checkAppVersion.StreamServerInterceptor
	// s.checkAppVersionUnary = checkAppVersion.UnaryServerInterceptor

	iapClient := appstore.New()
	if c.FakeAppleServer != "" {
		iapClient.ProductionURL = c.FakeAppleServer
	}

	// new api handler metrics collector
	s.apiHandlerCollector = metrics.NewMetricCollector()

	// Initialize file storage object.
	var fs filestore.FileStore
	if strings.Contains(storageConfig.Endpoint, "minio") {
		zapLogger.Info(fmt.Sprintf("using minio file storage, secure %v", c.Storage.Secure))

		fs, err = filestore.NewFileStore(filestore.MinIOService, c.Common.ServiceAccountEmail, storageConfig)
		if err != nil {
			return fmt.Errorf("failed to init MinIO file storage %s", err)
		}
	} else {
		fs, err = filestore.NewFileStore(filestore.GoogleCloudStorageService, c.Common.ServiceAccountEmail, storageConfig)
		if err != nil {
			return fmt.Errorf("failed to init Google file storage %s", err)
		}
	}
	s.fs = fs

	// common stuffs
	vCrLog := &log_service.VirtualClassRoomLogService{DB: dbTrace, Repo: &repositories.VirtualClassroomLogRepo{}}
	schoolRepo := &repositories.SchoolRepo{}

	userRepo := &repositories.UserRepoWrapper{
		LocalCacher:    s.cacheWrapper,
		UserRepository: &repositories.UserRepo{},
	}
	lessonMemberRepo := &repositories.LessonMemberRepo{}
	userMgmtModifierService := upb.NewUserModifierServiceClient(s.usermgmtConn)
	userMgmtAuthService := upb.NewAuthServiceClient(s.usermgmtConn)
	httpClient := &http.Client{
		Timeout: time.Second * 2,
	}
	mediaRepo := &repositories.MediaRepo{}
	s.whiteboardSvc = whiteboard.New(&c.Whiteboard)
	classRepo := &repositories.ClassRepo{}
	classMemberRepo := &repositories.ClassMemberRepo{}
	{
		// user related
		studentRepo := &repositories.StudentRepo{
			CreateSchoolFn: schoolRepo.Create,
		}

		topicRepo := &repositories.TopicRepo{}
		userMgmtStudentService := upb.NewStudentServiceClient(s.usermgmtConn)
		studyPlanReaderService := epb.NewStudyPlanReaderServiceClient(s.eurekaConn)
		studentLearningTimeSvc := epb.NewStudentLearningTimeReaderClient(s.eurekaConn)
		studentEventLogModifierSvc := epb.NewStudentEventLogModifierServiceClient(s.eurekaConn)
		s.userReaderSvc = &users.UserReaderService{
			DB:       dbTrace,
			UserRepo: &repositories.UserRepo{},
		}

		studentService := &services.StudentServiceABAC{
			StudentService: services.NewStudentService(
				eurekaDBTrace,
				dbTrace,
				c.Common.Environment,
				rsc.NATS(),
				studentRepo,
				&repositories.StudentEventLogRepo{},
				userRepo,
				&repositories.PresetStudyPlanRepo{},
				&repositories.StudentsLearningObjectivesCompletenessRepo{},
				&repositories.StudentStatRepo{},
				&repositories.LearningObjectiveRepo{},
				&repositories.StudentCommentRepo{},
				&repositories.StudentTopicCompletenessRepo{},
				&repositories.StudentTopicOverdueRepo{},
				topicRepo,
				userMgmtStudentService,
				studyPlanReaderService,
				studentLearningTimeSvc,
				studentEventLogModifierSvc,
			),
		}

		studentService.DB = dbTrace
		studentService.ConfigRepo = &repositories.ConfigRepo{}
		studentService.ActivityLogRepo = &repositories.ActivityLogRepo{}
		studentService.StudentOrderRepo = &repositories.StudentOrderRepo{}
		studentService.AppleUserRepo = &repositories.AppleUserRepo{}

		learningTimeDailyRepo := &repositories.StudentLearningTimeDailyRepo{}
		studentService.StudentLearningTimeDaiyRepo = learningTimeDailyRepo

		classMemberRepo := &repositories.ClassMemberRepo{}
		studentService.ClassMemberRepo = classMemberRepo
		s.studentSvc = studentService

		s.userModifierSvc = &users.UserModifierService{
			ApplicantID:         c.JWTApplicant,
			FirebaseClient:      firebaseClient,
			DB:                  dbTrace,
			UserRepo:            userRepo,
			UserGroupRepo:       &repositories.UserGroupRepo{},
			SchoolAdminRepo:     &repositories.SchoolAdminRepo{},
			TeacherRepo:         &repositories.TeacherRepo{},
			ShamirClient:        shamir.NewTokenReaderServiceClient(s.shamirConn),
			JSM:                 rsc.NATS(),
			TenantManager:       tenantManager,
			UserMgmtModifierSvc: userMgmtModifierService,
			UserMgmtAuthSvc:     userMgmtAuthService,
		}

		s.userSvc = &services.UserService{
			DB:                  dbTrace,
			JSM:                 rsc.NATS(),
			UserRepo:            userRepo,
			StudentRepo:         &repositories.StudentRepo{},
			SchoolAdminRepo:     &repositories.SchoolAdminRepo{},
			ActivityLogRepo:     &repositories.ActivityLogRepo{},
			FirebaseClient:      firebaseClient,
			TeacherRepo:         &repositories.TeacherRepo{},
			AppleUserRepo:       &repositories.AppleUserRepo{},
			UserDeviceTokenRepo: &notificationRepo.UserDeviceTokenRepo{},
		}
	}
	{
		// course related
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
		bookRepo := &repositories.BookRepo{}
		chapterRepo := &repositories.ChapterRepo{}
		lessonMemberRepo := &repositories.LessonMemberRepo{}
		eurekaTopicReaderSvc := epb.NewTopicReaderServiceClient(s.eurekaConn)
		eurekaStudentSubmissionReaderService := epb.NewStudentSubmissionReaderServiceClient(s.eurekaConn)

		s.courseSvc = &services.CourseService{
			EurekaDBTrace:                        eurekaDBTrace,
			DB:                                   dbTrace,
			Env:                                  c.Common.Environment,
			EurekaTopicReaderSvc:                 eurekaTopicReaderSvc,
			EurekaStudentSubmissionReaderService: eurekaStudentSubmissionReaderService,
			QuestionRepo:                         &repositories.QuestionRepo{},
			UserRepo:                             userRepo,
			TopicRepo:                            &repositories.TopicRepo{},
			QuizSetRepo:                          &repositories.QuizSetRepo{},
			QuestionSetsRepo:                     &repositories.QuestionSetRepo{},
			LearningObjectiveRepo:                &repositories.LearningObjectiveRepo{},
			TopicLearningObjectiveRepo:           &repositories.TopicsLearningObjectivesRepo{},
			StudentsLearningObjectivesCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
			StudentEventLogRepo:                        &repositories.StudentEventLogRepo{},
			PresetStudyPlanRepo:                        &repositories.PresetStudyPlanRepo{},
			StudentTopicCompletenessRepo:               &repositories.StudentTopicCompletenessRepo{},
			ChapterRepo:                                chapterRepo,
			CourseRepo:                                 &repositories.CourseRepo{},
			ClassRepo:                                  &repositories.ClassRepo{},
			TeacherRepo:                                &repositories.TeacherRepo{},
			SchoolAdminRepo:                            &repositories.SchoolAdminRepo{},
			ActivityLogRepo:                            &repositories.ActivityLogRepo{},
			LessonRepo:                                 &repositories.LessonRepo{},
			CourseClassRepo:                            &repositories.CourseClassRepo{},
			ClassMemberRepo:                            &repositories.ClassMemberRepo{},
			SchoolRepo:                                 &repositories.SchoolRepo{},
			BookRepo:                                   bookRepo,
			CourseBookRepo:                             &repositories.CourseBookRepo{},
			BookChapterRepo:                            &repositories.BookChapterRepo{},
			BrightCoveService: &yasuo_service.CourseService{
				BrightCoveProfile:    c.Brightcove.Profile,
				BrightcoveExtService: brightcoveExtService,
				BrightcoveService: yasuo_service.BrightcoveService{
					BrightcoveExtService: brightcoveExtService,
				},
			},
			LessonMemberRepo: lessonMemberRepo,
			ConfigRepo:       &repositories.ConfigRepo{},
			Cfg:              &c,
			UnleashClientIns: unleashClientInstance,
		}
	}
	{
		// master data related
		s.masterDataReaderService = &master_data.MasterDataReaderService{
			DB: dbTrace,
			LocationReaderService: &location_service.LocationReaderServices{
				DB:               dbTrace,
				LocationRepo:     &location_repo.LocationRepo{},
				LocationTypeRepo: &location_repo.LocationTypeRepo{},
				GetLocationQueryHandler: location_queries.GetLocationQueryHandler{
					DB:               dbTrace,
					LocationRepo:     &location_repo.LocationRepo{},
					Env:              c.Common.Environment,
					UnleashClientIns: unleashClientInstance,
				},
			},
		}
		s.masterDataSvc = &services.MasterDataService{
			Cfg:                 &c,
			DB:                  dbTrace,
			UserRepo:            userRepo,
			PresetStudyPlanRepo: &repositories.PresetStudyPlanRepo{},
			TopicRepo:           &repositories.TopicRepo{},
			CourseService:       s.courseSvc,
		}
		s.masterDataImporterSvc = &master_data.MasterDataImporterService{
			DB: dbTrace,
			LocationImporterService: location_service.NewLocationManagementGRPCService(
				dbTrace,
				rsc.NATS(),
				&location_repo.LocationRepo{},
				&location_repo.LocationTypeRepo{},
				&location_repo.ImportLogRepo{},
				unleashClientInstance,
				c.Common.Environment,
			),
		}
	}
	{
		// internal related
		s.internalModifierSvc = &services.InternalModifierService{
			EurekaDBTrace: eurekaDBTrace,
			DB:            dbTrace,
			QuizRepo:      &repositories.QuizRepo{},
			StudentsLearningObjectivesCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
			ShuffledQuizSetRepo:                        &repositories.ShuffledQuizSetRepo{},
			StudentRepo:                                &repositories.StudentRepo{},
			StudentLearningTimeDailyRepo:               &repositories.StudentLearningTimeDailyRepo{},
		}
		s.internalSvc = &services.InternalService{
			DB:               dbTrace,
			StudentOrderRepo: &repositories.StudentOrderRepo{},
		}
		s.internalSvcCacher = services.NewInternalServiceCacher(s.cacheWrapper, s.internalSvc)

		s.internalReaderSvc = &services.InternalReaderService{
			EurekaDBTrace:                eurekaDBTrace,
			DB:                           dbTrace,
			BookChapterRepo:              &repositories.BookChapterRepo{},
			TopicRepo:                    &repositories.TopicRepo{},
			TopicsLearningObjectivesRepo: &repositories.TopicsLearningObjectivesRepo{},
			CheckClientVersions:          c.CheckClientVersions,
			CoursesBooksRepo:             &repositories.CourseBookRepo{},
			LearningObjectiveRepo:        &repositories.LearningObjectiveRepo{},
			QuizSetRepo:                  &repositories.QuizSetRepo{},
			StudentsLearningObjectivesCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
			CourseReaderServiceClient:                  epb.NewCourseReaderServiceClient(s.eurekaConn),
		}
	}

	{
		// lesson related
		vCrLog := &log_service.VirtualClassRoomLogService{DB: dbTrace, Repo: &repositories.VirtualClassroomLogRepo{}}
		lessonRepo := &repositories.LessonRepo{}
		s.lessonModifierSvc = lessons.NewLessonModifierServices(
			c,
			dbTrace,
			mediaRepo,
			&repositories.ActivityLogRepo{},
			lessonRepo,
			&repositories.LessonGroupRepo{},
			&repositories.CourseRepo{},
			&repositories.PresetStudyPlanRepo{},
			&yasuo_repo.PresetStudyPlanWeeklyRepo{},
			&yasuo_repo.TopicRepo{},
			&repositories.UserRepo{},
			&repositories.SchoolAdminRepo{},
			&repositories.TeacherRepo{},
			&repositories.StudentRepo{},
			lessonMemberRepo,
			&repositories.LessonPollingRepo{},
			&lessonmgmt_repo.LessonRoomStateRepo{},
			rsc.NATS(),
			vCrLog,
		)

		s.lessonReaderSvc = &lessons.LessonReaderServices{
			DB:                         dbTrace,
			VirtualClassRoomLogService: vCrLog,
			LessonRepo:                 lessonRepo,
			LessonMemberRepo:           lessonMemberRepo,
			UserRepo:                   userRepo,
			SchoolAdminRepo:            &repositories.SchoolAdminRepo{},
			MediaRepo:                  mediaRepo,
			LessonRoomStateRepo:        &lessonmgmt_repo.LessonRoomStateRepo{},
			Env:                        c.Common.Environment,
			UnleashClientIns:           unleashClientInstance,
			CourseClassRepo:            &repositories.CourseClassRepo{},
			ClassRepo:                  &repositories.ClassRepo{},
		}
		s.lessonReportReaderSvc = &lesson_reports.LessonReportReaderService{
			DB:              dbTrace,
			Cfg:             &c,
			UserRepo:        userRepo,
			SchoolAdminRepo: &repositories.SchoolAdminRepo{},
			TeacherRepo:     &repositories.TeacherRepo{},
			ConfigRepo:      &repositories.ConfigRepo{},
			StudentRepo:     &repositories.StudentRepo{},
		}
	}
	{
		// class related
		classSvc := &classes.ClassService{
			Cfg:                   &c,
			ClassCodeLength:       c.ClassCodeLength,
			DB:                    dbTrace,
			UserRepo:              userRepo,
			HTTPClient:            httpClient,
			ClassRepo:             &repositories.ClassRepo{},
			ClassMemberRepo:       &repositories.ClassMemberRepo{},
			ConfigRepo:            &repositories.ConfigRepo{},
			StudentOrderRepo:      &repositories.StudentOrderRepo{},
			SchoolConfigRepo:      &repositories.SchoolConfigRepo{},
			TeacherRepo:           &repositories.TeacherRepo{},
			TopicRepo:             &repositories.TopicRepo{},
			LearningObjectiveRepo: &repositories.LearningObjectiveRepo{},
			QuestionRepo:          &repositories.QuestionRepo{},
			StudentEventLogRepo:   &repositories.StudentEventLogRepo{},
			LessonRepo:            &repositories.LessonRepo{},

			SchoolRepo:             schoolRepo,
			SchoolAdminRepo:        &repositories.SchoolAdminRepo{},
			QuestionSetRepo:        &repositories.QuestionSetRepo{},
			ActivityLogRepo:        &repositories.ActivityLogRepo{},
			CourseClassRepo:        &repositories.CourseClassRepo{},
			YasuoCourseClassRepo:   &yasuo_repo.CourseClassRepo{},
			PresetStudyPlanRepo:    &repositories.PresetStudyPlanRepo{},
			CourseRepo:             &repositories.CourseRepo{},
			MediaRepo:              mediaRepo,
			WhiteboardSvc:          s.whiteboardSvc,
			LessonMemberRepo:       lessonMemberRepo,
			LessonGroupRepo:        &repositories.LessonGroupRepo{},
			LessonModifierServices: s.lessonModifierSvc,
			MasterClassRepo:        &masterClassRepo.ClassRepo{},
			MasterClassMemberRepo:  &masterClassRepo.ClassMemberRepo{},
			JSM:                    rsc.NATS(),
		}
		s.classServiceABAC = &classes.ClassServiceABAC{
			ClassService: classSvc,
		}

		s.classReaderService = &classes.ClassReaderService{
			EurekaDBTrace:            eurekaDBTrace,
			DB:                       dbTrace,
			Env:                      c.Common.Environment,
			ClassRepo:                classRepo,
			LessonMemberRepo:         lessonMemberRepo,
			ClassMemberRepo:          classMemberRepo,
			UserRepo:                 userRepo,
			StudentEventLogRepo:      &repositories.StudentEventLogRepo{},
			UnleashClientIns:         unleashClientInstance,
			MasterClassRepo:          &masterClassRepo.ClassRepo{},
			MasterClassMemberRepo:    &masterClassRepo.ClassMemberRepo{},
			StudentEnrollmentHistory: &repositories.StudentEnrolledHistoryRepo{},
			CourseReaderSvc:          epb.NewCourseReaderServiceClient(s.eurekaConn),
			LessonRepo:               &lessonmgmt_repo.LessonRepo{},
			SchoolHistoryRepo:        &repositories.SchoolHistoryRepo{},
			TaggedUserRepo:           &repositories.TaggedUserRepo{},
			StudyPlanReaderService:   epb.NewStudyPlanReaderServiceClient(s.eurekaConn),
		}
		conversionTaskRepo := &repositories.ConversionTaskRepo{}
		s.cloudConvertSvc = &cloudconvert.Service{
			Host:            c.CloudConvert.Host,
			Token:           c.CloudConvert.Token,
			ProjectID:       c.Common.GoogleCloudProject,
			StorageBucket:   storageConfig.Bucket,
			StorageEndpoint: storageConfig.Endpoint,
			ClientEmail:     c.CloudConvert.ServiceAccountEmail,
			PrivateKey:      c.CloudConvert.ServiceAccountPK,

			Client: &http.Client{
				Timeout: 10 * time.Second,
			},
		}

		classModifierSvc := &classes.ClassModifierService{
			DB:                 dbTrace,
			ConversionTaskRepo: conversionTaskRepo,
			ConversionSvc:      s.cloudConvertSvc,
			OldClassService: &classes.ClassServiceABAC{
				ClassService: classSvc,
			},
			VirtualClassRoomLogService: vCrLog,
			LessonRoomStateRepo:        &repo.LessonRoomStateRepo{},
			LessonMgmtRoomStateRepo:    &lessonmgmt_repo.LessonRoomStateRepo{},
		}
		classModifierSvc.RegisterMetric(s.apiHandlerCollector)
		s.classModifierSvc = classModifierSvc
	}

	privateKey, err := gl_interceptors.PrivateKeyFromString(c.GetPostgresUserPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create postgres get user info private key: %v", err)
	}
	s.pgUserSvc = &services.PostgresUserService{
		DB:               dbTrace,
		OriginalKey:      c.GetPostgresUserKey,
		PrivateKey:       privateKey,
		PostgresUserRepo: &repositories.PostgresUserRepo{},
	}

	s.pgNamespaceSvc = &services.PostgresNamespaceService{
		DB:                    dbTrace,
		OriginalKey:           c.GetPostgresUserKey,
		PrivateKey:            privateKey,
		PostgresNamespaceRepo: &repositories.PostgresNamespaceRepo{},
	}

	return nil
}

func (s *server) WithPrometheusCollectors(rsc *bootstrap.Resources) []prometheus.Collector {
	return s.apiHandlerCollector.Collectors()
}

func (s *server) InitMetricsValue() {
}

func (s *server) SetupGRPC(_ context.Context, grpcserv *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	// order by the order of old bob.go code, for easy code review
	db := rsc.DB()
	unleashClient := rsc.Unleash()
	storageConfig := rsc.Storage()
	pb.RegisterStudentServer(grpcserv, s.studentSvc)
	pb.RegisterCourseServer(grpcserv, &services.CourseServiceABAC{
		CourseService: s.courseSvc,
	})
	pb.RegisterMasterDataServiceServer(grpcserv, s.masterDataSvc)
	pb.RegisterUserServiceServer(grpcserv, s.userSvc)

	pb.RegisterInternalServer(grpcserv, s.internalSvcCacher)

	pb_v1.RegisterInternalReaderServiceServer(grpcserv, s.internalReaderSvc)

	pb_v1.RegisterLessonModifierServiceServer(grpcserv, s.lessonModifierSvc)
	wrapperDBConnection := support.InitWrapperDBConnector(db, db, unleashClient, c.Common.Environment)
	userModule := user.New(grpcserv, db, wrapperDBConnection, c.Common.Environment, unleashClient)
	umAdapter := &usermodadapter.UserModuleAdapter{
		Module: userModule,
	}
	mediaModule := lesson_media.New(db, &lesson_media_infrastructure.MediaRepo{})
	mediaModuleAdapter := &mediaadapter.MediaModuleAdapter{
		Module: mediaModule,
	}
	lessonModule := lesson.NewModuleWriter(grpcserv, wrapperDBConnection, rsc.NATS(), umAdapter, mediaModuleAdapter, c.Common.Environment, unleashClient, nil, nil)
	lessonModuleReader := lesson.NewModuleReader(grpcserv, wrapperDBConnection, c.Common.Environment, unleashClient)

	pb_v1.RegisterLessonManagementServiceServer(grpcserv, &lessons.LessonManagementService{
		CreateLessonV2:    lessonModule.LessonModifierService.CreateLesson,
		UpdateLessonV2:    lessonModule.LessonModifierService.UpdateLesson,
		DeleteLessonV2:    lessonModule.LessonModifierService.DeleteLesson,
		RetrieveLessonsV2: lessonModuleReader.LessonReaderService.RetrieveLessonsV2,
	})

	pb_v1.RegisterMasterDataImporterServiceServer(grpcserv, s.masterDataImporterSvc)

	pb_v1.RegisterMasterDataReaderServiceServer(grpcserv, s.masterDataReaderService)

	pb.RegisterClassServer(grpcserv, s.classServiceABAC)

	pb_v1.RegisterClassReaderServiceServer(grpcserv, s.classReaderService)

	pb_v1.RegisterUserModifierServiceServer(grpcserv, s.userModifierSvc)

	pb_v1.RegisterUserReaderServiceServer(grpcserv, s.userReaderSvc)

	pb_v1.RegisterLessonReaderServiceServer(grpcserv, s.lessonReaderSvc)

	pb_v1.RegisterLessonReportReaderServiceServer(grpcserv, s.lessonReportReaderSvc)

	pb_v1.RegisterLessonReportModifierServiceServer(grpcserv, &lesson_reports.LessonReportModifierService{
		DB:                             db,
		PartnerFormConfigRepo:          &repositories.PartnerFormConfigRepo{},
		LessonRepo:                     &repositories.LessonRepo{},
		LessonReportRepo:               &repositories.LessonReportRepo{},
		LessonReportDetailRepo:         &repositories.LessonReportDetailRepo{},
		LessonMemberRepo:               &repositories.LessonMemberRepo{},
		TeacherRepo:                    &repositories.TeacherRepo{},
		LessonReportApprovalRecordRepo: &repositories.LessonReportApprovalRecordRepo{},
		UpdateLessonSchedulingStatus:   lessonModule.LessonModifierService.UpdateLessonSchedulingStatus,
		Env:                            c.Common.Environment,
		UnleashClientIns:               unleashClient,
	})

	pb_v1.RegisterStudentSubscriptionServiceServer(grpcserv, &lessons.StudentSubscriptionService{
		GetStudentCourseSubscriptionsServiceV2: userModule.StudentSubscriptionGRPCLessonmgmtService.GetStudentCourseSubscriptions,
		RetrieveStudentSubscriptionServiceV2:   userModule.StudentSubscriptionGRPCLessonmgmtService.RetrieveStudentSubscription,
	})

	pb_v1.RegisterInternalModifierServiceServer(grpcserv, s.internalModifierSvc)

	health.RegisterHealthServer(grpcserv, &healthcheck.Service{DB: db.DB.(*pgxpool.Pool)})

	eurekaCourseModifierService := epb.NewCourseModifierServiceClient(s.eurekaConn)
	eurekaChapterReaderService := epb.NewChapterReaderServiceClient(s.eurekaConn)
	eurekaQuizReaderService := epb.NewQuizReaderServiceClient(s.eurekaConn)
	eurekaQuizModifierService := epb.NewQuizModifierServiceClient(s.eurekaConn)
	assignmentModifierService := epb.NewAssignmentModifierServiceClient(s.eurekaConn)
	eurekaBookReaderService := epb.NewBookReaderServiceClient(s.eurekaConn)
	eurekaFlashcardReaderService := epb.NewFlashCardReaderServiceClient(s.eurekaConn)
	eurekaStudyPlanReaderService := epb.NewStudyPlanReaderServiceClient(s.eurekaConn)
	studyPlanReaderService := epb.NewStudyPlanReaderServiceClient(s.eurekaConn)
	eurekaLearningObjectiveModifierService := epb.NewLearningObjectiveModifierServiceClient(s.eurekaConn)
	initV1BOB(grpcserv, s.eurekaDB, rsc.DB(), c.Common.Environment, rsc.NATS(), studyPlanReaderService, eurekaCourseModifierService, assignmentModifierService, eurekaBookReaderService, eurekaChapterReaderService, eurekaFlashcardReaderService, eurekaStudyPlanReaderService, eurekaQuizReaderService, eurekaQuizModifierService, eurekaLearningObjectiveModifierService)

	pb_v1.RegisterClassModifierServiceServer(grpcserv, s.classModifierSvc)

	userMgmtStudentService := upb.NewStudentServiceClient(s.usermgmtConn)
	studentModifierSvc := students.NewStudentModifierServices(db, userMgmtStudentService)
	pb_v1.RegisterStudentModifierServiceServer(grpcserv, studentModifierSvc)

	userReaderService := upb.NewUserReaderServiceClient(s.usermgmtConn)
	studentReaderSvc := students.NewStudentReaderService(db, userReaderService)
	pb_v1.RegisterStudentReaderServiceServer(grpcserv, studentReaderSvc)

	pb_v1.RegisterUploadServiceServer(grpcserv, &uploads.UploadReaderService{FileStore: s.fs, Cfg: *storageConfig})

	pb_v1.RegisterMediaModifierServiceServer(grpcserv, media.NewMediaModifierService(
		&uploads.UploadReaderService{FileStore: s.fs, Cfg: *storageConfig},
	))

	pb_v1.RegisterPostgresUserServiceServer(grpcserv, s.pgUserSvc)

	pb_v1.RegisterPostgresNamespaceServiceServer(grpcserv, s.pgNamespaceSvc)

	return nil
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()
	// DB must have locations table applied AC
	locationRestrictedInterceptor := locationRestrictedInterceptor(rsc.DB())
	customs := []grpc.UnaryServerInterceptor{
		grpc_zap.PayloadUnaryServerInterceptor(rsc.Logger(), func(ctx context.Context, fullMethod string, _ interface{}) bool {
			if !c.Common.Log.LogPayload {
				return false
			}

			// don't need to log internal APIs
			if golibs.InArrayString(fullMethod, ignoreAuthEndpoint) {
				return false
			}

			// RetrieveSchools API may return more than 4000 schools,
			// logging response may be slow, skip for now.
			if fullMethod == "/manabie.bob.SchoolService/RetrieveSchools" {
				return false
			}

			return true
		}),
		// s.checkAppVersionUnary,
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		locationRestrictedInterceptor.UnaryServerInterceptor,
	}
	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	grpcUnary = append(grpcUnary, customs...)

	return grpcUnary
}

func (s *server) WithStreamServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	grpcStream := bootstrap.DefaultStreamServerInterceptor(rsc)
	// grpcStream = append(grpcStream, s.checkAppVersionStream)
	grpcStream = append(grpcStream, s.authInterceptor.StreamServerInterceptor)
	return grpcStream
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	jsm := rsc.NATS()
	zapLogger := rsc.Logger()
	dbTrace := rsc.DB()
	lessonsmgmtDBTrace := rsc.DBWith("lessonmgmt")
	unleashClient := rsc.Unleash()
	mediaRepo := &repositories.MediaRepo{}
	conversionTaskRepo := &repositories.ConversionTaskRepo{}
	wrapperConnection := bob_support.InitWrapperDBConnector(dbTrace, lessonsmgmtDBTrace, unleashClient, c.Common.Environment)
	err := initCloudConvertJobEventsSubscription(jsm, zapLogger, dbTrace, mediaRepo, conversionTaskRepo, s.cloudConvertSvc)
	if err != nil {
		return fmt.Errorf("initCloudConvertJobEventsSubscription: %v", err)
	}

	studentSubscriptionRepo := &repositories.StudentSubscriptionRepo{}
	studentSubscriptionAccessPathRepo := &repositories.StudentSubscriptionAccessPathRepo{}
	userRepoBob := &repositories.UserRepo{}
	lessonAllocationRepo := &lesson_allocation_repo.LessonAllocationRepo{}
	courseLocationScheduleRepo := &cls_repo.CourseLocationScheduleRepo{}
	studentCourseRepo := &user_repo.StudentCourseRepo{}
	err = initStudentCourserSubscription(
		jsm,
		zapLogger,
		wrapperConnection,
		studentSubscriptionRepo,
		studentSubscriptionAccessPathRepo,
		userRepoBob,
		courseLocationScheduleRepo,
		lessonAllocationRepo,
		studentCourseRepo,
		c.Common.Environment,
		unleashClient,
	)
	if err != nil {
		return fmt.Errorf("initStudentCourserSubscription: %v", err)
	}

	// as source database, it will listen to incremental snapshot events which will trigger new captured table
	err = debezium.InitDebeziumIncrementalSnapshot(jsm, zapLogger, dbTrace, c.Common.Name)
	if err != nil {
		return fmt.Errorf("initInternalDebeziumIncrementalSnapshot: %v", err)
	}

	return nil
}
