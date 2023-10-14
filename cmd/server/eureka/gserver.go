package eureka

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/eureka/configurations"
	eureka_interceptors "github.com/manabie-com/backend/internal/eureka/golibs/interceptors"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	items_bank_repo "github.com/manabie-com/backend/internal/eureka/repositories/items_bank"
	"github.com/manabie-com/backend/internal/eureka/services"
	ib_service "github.com/manabie-com/backend/internal/eureka/services/items_bank"
	lhds_service "github.com/manabie-com/backend/internal/eureka/services/learning_history_data_sync"
	"github.com/manabie-com/backend/internal/eureka/services/question"
	"github.com/manabie-com/backend/internal/eureka/services/studyplans"
	assessment_grpc2 "github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/transport/grpc"
	assessment_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository/postgres"
	grpc2 "github.com/manabie-com/backend/internal/eureka/v2/modules/book/transport/grpc"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/usecase"
	ext_course_repo "github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/external"
	course_repo "github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres"
	course_grpc2 "github.com/manabie-com/backend/internal/eureka/v2/modules/course/transport/grpc"
	course_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/course/usecase"
	item_bank_grpc2 "github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/transport/grpc"
	item_bank_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/usecase"
	study_plan_repo "github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres"
	study_plan_grpc2 "github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/transport/grpc"
	study_plan_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/cerebry"
	utils "github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	learnosity_data "github.com/manabie-com/backend/internal/golibs/learnosity/data"
	learnosity_http "github.com/manabie-com/backend/internal/golibs/learnosity/http"
	"github.com/manabie-com/backend/internal/golibs/mathpix"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/golibs/vision"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	pbv2 "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	vv1 "cloud.google.com/go/vision/apiv1"
	"github.com/gin-gonic/gin"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opencensus.io/plugin/ocgrpc"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithMonitorServicer(s).
		WithHTTP(s).
		Register(s)
	bootstrap.RegisterJob("eureka_regenerate_speeches_audio_link", RunRegenerateSpeechesAudioLink)
}

type server struct {
	bootstrap.DefaultMonitorService[configurations.Config]

	bobConn        *grpc.ClientConn
	yasuoConn      *grpc.ClientConn
	userMgmtConn   *grpc.ClientConn
	masterMgmtConn *grpc.ClientConn

	authInterceptor *interceptors.Auth

	studentLearningTimeReaderSvc *services.StudentLearningTimeReaderService
	assignmentModifierSvc        *services.AssignmentModifierService
	assignmentReaderSvc          *services.AssignmentReaderService
	studentAssignmentWriteSvc    *services.StudentAssignmentWriteABACService
	studentAssignmentReadSvc     *services.StudentAssignmentReaderABACService
	importSvc                    *services.ImportService
	loReaderSvc                  *services.LoReaderService
	courseModifierSvc            *services.CourseModifierService
	studyPlanReaderSvc           pb.StudyPlanReaderServiceServer
	studentSubmissionWriterSvc   *services.StudentSubmissionWriterService
	studentSubmissionReaderSvc   *services.StudentSubmissionReaderService
	studyPlanModifierSvc         pb.StudyPlanModifierServiceServer
	topicReaderSvc               *services.TopicReaderService
	topicModifierSvc             *services.TopicModifierService
	internalModifierSvc          *services.InternalModifierService
	bookReaderSvc                *services.BookReaderService
	bookModifierSvc              pb.BookModifierServiceServer
	quizModifierSvc              *services.QuizModifierService
	quizReaderSvc                *services.QuizReaderService
	chapterModifierSvc           *services.ChapterModifierService
	chapterReaderSvc             *services.ChapterReaderService
	flashcardReaderSvc           *services.FlashCardReaderService
	learningObjectiveModifierSvc *services.LearningObjectiveModifierService
	studentEventLogModifierSvc   pb.StudentEventLogModifierServiceServer
	visionReaderSvc              *services.VisionReaderService
	imageToText                  *services.ImageToText
	flashcardSvc                 sspb.FlashcardServer
	assignmentSvc                *services.AssignmentService
	learningMaterialSvc          *services.LearningMaterialService
	learningObjectiveSvc         *services.LearningObjectiveService
	examLOSvc                    sspb.ExamLOServer
	statisticSvc                 sspb.StatisticsServer
	studyPlanSvc                 sspb.StudyPlanServer
	taskAssignmentSvc            sspb.TaskAssignmentServer
	quizSvc                      sspb.QuizServer
	questionSvc                  *question.Service
	studentSubmissionSvc         sspb.StudentSubmissionServiceServer
	questionTagSvc               sspb.QuestionTagServer
	questionTagTypeSvc           sspb.QuestionTagTypeServer
	learningHistoryDataSyncSvc   sspb.LearningHistoryDataSyncServiceServer
	studyPlanItemReaderSvc       pb.StudyPlanItemReaderServiceServer
	asssessmentSvc               sspb.AssessmentServer
	itemsBankSvc                 sspb.ItemsBankServiceServer
	studentSvc                   pb.StudentServiceServer
	assessmentSessionSvc         pb.AssessmentSessionServiceServer

	// v2
	v2 serverV2
}

type serverV2 struct {
	bookSvc             pbv2.BookServiceServer
	learningMaterialSvc pbv2.LearningMaterialServiceServer
	assessmentSvc       pbv2.AssessmentServiceServer
	courseSvc           pbv2.CourseServiceServer
	itemBankSvc         pbv2.ItemBankServiceServer
	studyPlanItemSvc    pbv2.StudyPlanItemServiceServer
	studyPlanSvc        pbv2.StudyPlanServiceServer
	cerebrySvc          pbv2.CerebryServiceServer
}

func (*server) ServerName() string {
	return "eureka"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	logger := rsc.Logger()
	rlsForInternalInterceptor := rlsSimulatedInterceptor()

	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	customs := []grpc.UnaryServerInterceptor{
		grpc_zap.PayloadUnaryServerInterceptor(logger, func(ctx context.Context, fullMethod string, _ interface{}) bool {
			if !c.Common.Log.LogPayload {
				return false
			}

			// don't need to log internal APIs
			if utils.InArrayString(fullMethod, ignoreAuthEndpoint) {
				return false
			}

			// RetrieveSchools API may return more than 4000 schools,
			// logging response may be slow, skip for now.
			if fullMethod == "/manabie.bob.SchoolService/RetrieveSchools" {
				return false
			}

			return true
		}),
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
		eureka_interceptors.UpdateUserIDForParent,
		rlsForInternalInterceptor.UnaryServerInterceptor,
		// interceptors_lib.TimeoutUnaryServerInterceptor(c.Common.GRPC.HandlerTimeout), // consider for future.
	}

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

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	logger := rsc.Logger()
	db := rsc.DB()
	jsm := rsc.NATS()
	unleashClient := rsc.Unleash()
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	grpc_zap.ReplaceGrpcLoggerV2(logger)
	s.authInterceptor = authInterceptor(&c, logger, db.DB)

	// grpc client
	s.bobConn = rsc.GRPCDial("bob")
	s.yasuoConn = rsc.GRPCDial("yasuo")
	s.userMgmtConn = rsc.GRPCDial("usermgmt")
	s.masterMgmtConn = rsc.GRPCDial("mastermgmt")

	classService := mpb.NewClassServiceClient(s.masterMgmtConn)

	creds, err := google.FindDefaultCredentials(ctx, vv1.DefaultAuthScopes()...)
	if err != nil {
		return fmt.Errorf("InitDependencies: google.FindDefaultCredentials: %w", err)
	}
	visionFactory, err := vision.NewFactory(ctx, creds)
	if err != nil {
		return fmt.Errorf("InitDependencies: vision.NewVisionFactory: %w", err)
	}

	// Initialize file storage object.
	storageConfig := rsc.Storage()
	var fs filestore.FileStore
	if strings.Contains(storageConfig.Endpoint, "minio") {
		logger.Info(fmt.Sprintf("using minio file storage, secure %v", c.Storage.Secure))

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

	bobCourseReaderClientV1 := bpb.NewCourseReaderServiceClient(s.bobConn)
	bobUserReaderClient := bpb.NewUserReaderServiceClient(s.bobConn)
	bobInternalReaderClient := bpb.NewInternalReaderServiceClient(s.bobConn)
	bobInternalModifierClient := bpb.NewInternalModifierServiceClient(s.bobConn)
	bobStudentReaderClient := bpb.NewStudentReaderServiceClient(s.bobConn)
	bobMediaModifierClient := bpb.NewMediaModifierServiceClient(s.bobConn)
	bobCourseClient := bob_pb.NewCourseClient(s.bobConn)

	yasuoCourseReaderClient := ypb.NewCourseReaderServiceClient(s.yasuoConn)
	yasuoUploadReader := ypb.NewUploadReaderServiceClient(s.yasuoConn)
	yasuoUploadModifier := ypb.NewUploadModifierServiceClient(s.yasuoConn)

	usermgmtUserReaderServiceClient := upb.NewUserReaderServiceClient(s.userMgmtConn)

	// init repo
	mathpixFactory := mathpix.NewFactory(c.Mathpix.AppID, c.Mathpix.AppKey, c.Storage.InsecureSkipVerify)

	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}
	loItemStudyPlanItemRepo := &repositories.LoStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &repositories.AssignmentStudyPlanItemRepo{}
	assignmentRepo := &repositories.AssignmentRepo{}
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	courseStudyPlanRepo := &repositories.CourseStudyPlanRepo{}
	studentRepo := &repositories.StudentRepo{}
	assignStudyPlanTaskRepo := &repositories.AssignStudyPlanTaskRepo{}
	userRepo := &repositories.UserRepo{}

	// init service
	s.topicModifierSvc = services.NewTopicModifierService(db, jsm)
	s.chapterReaderSvc = services.NewChapterReaderService(db)
	s.chapterModifierSvc = services.NewChapterModifierService(db)
	s.flashcardReaderSvc = services.NewFlashCardReaderService(db, c.Common.Environment)
	s.flashcardSvc = services.NewFlashcardService(db)
	s.bookModifierSvc = services.NewBookModifierService(db)
	s.learningMaterialSvc = services.NewLearningMaterialService(db)
	s.examLOSvc = services.NewExamLOService(db)
	s.statisticSvc = services.NewStatisticService(db, bobCourseClient, bobStudentReaderClient)
	s.quizSvc = services.NewQuizService(db, bobMediaModifierClient, yasuoUploadReader, yasuoUploadModifier)
	s.questionTagSvc = services.NewQuestionTagService(db)
	s.questionTagTypeSvc = services.NewQuestionTagTypeService(db)
	s.studyPlanItemReaderSvc = services.NewStudyPlanItemReaderService(db)
	// only use internal backend
	s.internalModifierSvc = &services.InternalModifierService{
		DB:                          db,
		LoStudyPlanItemRepo:         loItemStudyPlanItemRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		StudyPlanRepo:               studyPlanRepo,
		CourseBookRepo:              &repositories.CourseBookRepo{},
		StudentRepo:                 &repositories.StudentRepo{},
		BookRepo:                    &repositories.BookRepo{},
		AssignmentRepo:              &repositories.AssignmentRepo{},
		LearningObjectiveRepo:       &repositories.LearningObjectiveRepo{},
		AssignmentStudyPlanItemRepo: &repositories.AssignmentStudyPlanItemRepo{},
		StudentStudyPlanRepo:        &repositories.StudentStudyPlanRepo{},
	}

	s.importSvc = &services.ImportService{
		DB:                          db,
		StudyPlanItemRepo:           studyPlanItemRepo,
		StudyPlanRepo:               studyPlanRepo,
		StudentRepo:                 studentRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loItemStudyPlanItemRepo,
		StudentStudyPlan:            studentStudyPlanRepo,
		AssignStudyPlanTaskRepo:     assignStudyPlanTaskRepo,
		BookChapterRepo:             &repositories.BookChapterRepo{},
		JSM:                         jsm,
	}

	s.assignmentModifierSvc = &services.AssignmentModifierService{
		DB:                          db,
		JSM:                         jsm,
		StudyPlanRepo:               &repositories.StudyPlanRepo{},
		StudyPlanItemRepo:           studyPlanItemRepo,
		AssignmentRepo:              &repositories.AssignmentRepo{},
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loItemStudyPlanItemRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		StudentRepo:                 studentRepo,
		ClassStudyPlanRepo:          &repositories.ClassStudyPlanRepo{},
		StudentStudyPlanRepo:        studentStudyPlanRepo,
		TopicsAssignmentsRepo:       &repositories.TopicsAssignmentsRepo{},
		TopicRepo:                   &repositories.TopicRepo{},
		BookRepo:                    &repositories.BookRepo{},
		BookChapterRepo:             &repositories.BookChapterRepo{},
		ChapterRepo:                 &repositories.ChapterRepo{},
		CourseBookRepo:              &repositories.CourseBookRepo{},
		LearningObjectiveRepo:       &repositories.LearningObjectiveRepo{},
		BobStudentReaderSvc:         bobStudentReaderClient,
	}

	s.assignmentReaderSvc = &services.AssignmentReaderService{
		DB:                           db,
		Env:                          c.Common.Environment,
		StudentStudyPlanRepo:         studentStudyPlanRepo,
		LoStudyPlanItemRepo:          loItemStudyPlanItemRepo,
		AssignmentStudyPlanItemRepo:  assignmentStudyPlanItemRepo,
		AssignmentRepo:               assignmentRepo,
		StudyPlanItemRepo:            studyPlanItemRepo,
		TopicsAssignmentsRepo:        &repositories.TopicsAssignmentsRepo{},
		TopicRepo:                    &repositories.TopicRepo{},
		TopicsLearningObjectivesRepo: &repositories.TopicsLearningObjectivesRepo{},
	}

	s.studentAssignmentWriteSvc = &services.StudentAssignmentWriteABACService{
		StudentAssignmentWriteService: &services.StudentAssignmentWriteService{
			JSM:                          jsm,
			DB:                           db,
			SubmissionRepo:               &repositories.StudentSubmissionRepo{},
			StudentLatestSubmissionRepo:  &repositories.StudentLatestSubmissionRepo{},
			SubmissionGradeRepo:          &repositories.StudentSubmissionGradeRepo{},
			StudyPlanItemRepo:            studyPlanItemRepo,
			StudentReaderClient:          bobStudentReaderClient,
			StudentLearningTimeDailyRepo: &repositories.StudentLearningTimeDailyRepo{},
			UsermgmtUserReaderService:    usermgmtUserReaderServiceClient,
		},
		DB:             db,
		AssignmentRepo: assignmentRepo,
	}

	s.learningObjectiveModifierSvc = services.NewLearningObjectiveModifierService(db, jsm)
	s.learningObjectiveSvc = services.NewLearningObjectiveService(db, jsm)

	gradeRepo := &repositories.StudentSubmissionGradeRepo{}
	s.studentAssignmentReadSvc = &services.StudentAssignmentReaderABACService{
		StudentAssignmentReaderService: &services.StudentAssignmentReaderService{
			DB:                db,
			SubmissionRepo:    &repositories.StudentSubmissionRepo{},
			StudentRepo:       &repositories.StudentRepo{},
			GradeRepo:         gradeRepo,
			StudyPlanRepo:     studyPlanRepo,
			StudyPlanItemRepo: studyPlanItemRepo,
		},
		DB:            db,
		StudyPlanRepo: &repositories.StudentStudyPlanRepo{},
		GradeRepo:     gradeRepo,
	}
	s.courseModifierSvc = services.NewCourseModifierService(
		db,
		bobInternalModifierClient,
		s.assignmentModifierSvc,
		s.learningObjectiveModifierSvc,
		usermgmtUserReaderServiceClient,
	)
	s.quizReaderSvc = services.NewQuizReaderService(db)
	s.loReaderSvc = services.NewLOReaderService(db, services.NewCourseReaderService(db, bobUserReaderClient, classService))
	s.bookReaderSvc = services.NewBookReaderService(db)

	s.studyPlanReaderSvc = services.NewStudyPlanReaderService(db, bobInternalReaderClient)
	s.studentSubmissionWriterSvc = services.NewStudentSubmissionWriterService(db)
	s.studentSubmissionReaderSvc = services.NewStudentSubmissionReaderService(db)

	s.studyPlanModifierSvc = studyplans.NewStudyPlanModifierService(db, c.Common.Environment, jsm)
	s.studentEventLogModifierSvc = services.NewStudentEventLogModifierService(db, jsm, usermgmtUserReaderServiceClient)

	s.assignmentSvc = &services.AssignmentService{
		DB:                           db,
		TopicRepo:                    &repositories.TopicRepo{},
		GeneralAssignmentRepo:        &repositories.GeneralAssignmentRepo{},
		AssignmentRepo:               assignmentRepo,
		SubmissionRepo:               &repositories.StudentSubmissionRepo{},
		StudentLatestSubmissionRepo:  &repositories.StudentLatestSubmissionRepo{},
		StudentLearningTimeDailyRepo: &repositories.StudentLearningTimeDailyRepo{},
		UsermgmtUserReaderService:    usermgmtUserReaderServiceClient,
	}

	s.topicReaderSvc = &services.TopicReaderService{
		Env:                   c.Common.Environment,
		DB:                    db,
		StudyPlanRepo:         studyPlanRepo,
		StudyPlanItemRepo:     studyPlanItemRepo,
		AssignmentRepo:        assignmentRepo,
		LearningObjectiveRepo: &repositories.LearningObjectiveRepo{},
		BobCourseReaderClient: bobCourseReaderClientV1,
		TopicRepo:             &repositories.TopicRepo{},
		BookRepo:              &repositories.BookRepo{},
	}

	s.quizModifierSvc = services.NewQuizModifierService(db, unleashClient, c.Common.Environment, yasuoCourseReaderClient, bobMediaModifierClient, yasuoUploadReader, yasuoUploadModifier)

	s.studentLearningTimeReaderSvc = &services.StudentLearningTimeReaderService{
		DB:                          db,
		StudentLearningTimeDaiyRepo: &repositories.StudentLearningTimeDailyRepo{},
		UserMgmtService:             usermgmtUserReaderServiceClient,
	}

	s.visionReaderSvc = &services.VisionReaderService{
		VisionFactory: visionFactory,
	}

	s.imageToText = &services.ImageToText{
		MathpixFactory: mathpixFactory,
	}

	s.studyPlanSvc = services.NewStudyPlanService(db, jsm)

	s.taskAssignmentSvc = services.NewTaskAssignmentService(db, bobStudentReaderClient)

	s.questionSvc = question.NewQuestionService(
		db,
		&repositories.QuestionGroupRepo{},
		&repositories.QuizSetRepo{},
		&repositories.QuizRepo{},
		s.quizModifierSvc,
		yasuoUploadReader,
		yasuoUploadModifier,
	)

	s.studentSubmissionSvc = services.NewStudentSubmissionService(db)
	s.learningHistoryDataSyncSvc = lhds_service.NewLearningHistoryDataSyncService(db, yasuoUploadModifier)

	// learnosity's services
	s.asssessmentSvc = services.NewAssessmentService(db, &c.LearnosityConfig)
	s.itemsBankSvc = &ib_service.ItemsBankService{
		DB:                   db,
		LearningMaterialRepo: &repositories.LearningMaterialRepo{},
		ContentBankMediaRepo: &repositories.ContentBankMediaRepo{},
		ItemsBankRepo: &items_bank_repo.ItemsBankRepo{
			LearnosityConfig: &c.LearnosityConfig,
			HTTP:             &learnosity_http.Client{},
			DataAPI:          &learnosity_data.Client{},
		},
		Cfg:       storageConfig,
		FileStore: fs,
	}
	s.studentSvc = &services.StudentsService{
		DB:          db,
		StudentRepo: studentRepo,
		UserRepo:    userRepo,
	}

	s.assessmentSessionSvc = services.NewAssessmentSessionService(db)

	return s.initV2(c, rsc)
}

func (s *server) initV2(c configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()

	bookUsecase := usecase.NewBookUsecase(&postgres.BookRepo{
		DB: db,
	})

	learningMaterialUsecase := usecase.NewLearningMaterialUsecase(&postgres.LearningMaterialRepo{
		DB: db,
	})

	courseUseCase := course_usecase.NewCourseUsecase(
		&ext_course_repo.CourseRepo{CourseClient: mpb.NewMasterDataCourseServiceClient(s.masterMgmtConn)},
		&course_repo.CourseBookRepo{DB: db},
		&course_repo.CourseRepo{DB: db},
	)

	studyPlanUseCase := study_plan_usecase.NewStudyPlanUsecase(db, &study_plan_repo.StudyPlanRepo{})

	s.v2.bookSvc = grpc2.NewBookService(bookUsecase)
	s.v2.learningMaterialSvc = grpc2.NewLearningMaterialGrpcService(learningMaterialUsecase)

	assessmentUsecase := assessment_usecase.NewAssessmentUsecase(db, c.LearnosityConfig,
		&learnosity_http.Client{},
		&learnosity_data.Client{})
	s.v2.assessmentSvc = assessment_grpc2.NewAssessmentService(assessmentUsecase)

	s.v2.courseSvc = course_grpc2.NewCourseService(courseUseCase)

	activityUsecase := item_bank_usecase.NewActivityUsecase(db, c.LearnosityConfig, &learnosity_http.Client{}, &learnosity_data.Client{})
	s.v2.itemBankSvc = item_bank_grpc2.NewItemBankService(activityUsecase)
	s.v2.studyPlanSvc = study_plan_grpc2.NewStudyPlanService(studyPlanUseCase)

	studyPlanItemRepo := &study_plan_repo.StudyPlanItemRepo{DB: db}
	learningMaterialListRepo := &study_plan_repo.LmListRepo{DB: db}
	studyPlanItemUsecase := study_plan_usecase.NewStudyPlanItemUseCase(db, studyPlanItemRepo, learningMaterialListRepo)
	s.v2.studyPlanItemSvc = study_plan_grpc2.NewStudyPlanItemService(studyPlanItemUsecase)

	config := cerebry.Config{
		BaseURL:        cerebry.StagingURL,
		PermanentToken: cerebry.PermanentToken,
	}
	cerebryUsecase := course_usecase.NewCerebryUsecase(config)
	s.v2.cerebrySvc = course_grpc2.NewCerebryService(cerebryUsecase)

	return nil
}

func (s *server) SetupGRPC(ctx context.Context, grpcServer *grpc.Server, _ configurations.Config, rsc *bootstrap.Resources) error {
	pb.RegisterStudentLearningTimeReaderServer(grpcServer, s.studentLearningTimeReaderSvc)
	pb.RegisterAssignmentModifierServiceServer(grpcServer, s.assignmentModifierSvc)
	pb.RegisterAssignmentReaderServiceServer(grpcServer, s.assignmentReaderSvc)
	pb.RegisterStudentAssignmentWriteServiceServer(grpcServer, s.studentAssignmentWriteSvc)
	pb.RegisterStudentAssignmentReaderServiceServer(grpcServer, s.studentAssignmentReadSvc)
	pb.RegisterStudyPlanWriteServiceServer(grpcServer, s.importSvc)
	pb.RegisterCourseReaderServiceServer(grpcServer, s.loReaderSvc)
	pb.RegisterCourseModifierServiceServer(grpcServer, s.courseModifierSvc)
	pb.RegisterStudyPlanReaderServiceServer(grpcServer, s.studyPlanReaderSvc)
	pb.RegisterStudentSubmissionModifierServiceServer(grpcServer, s.studentSubmissionWriterSvc)
	pb.RegisterStudentSubmissionReaderServiceServer(grpcServer, s.studentSubmissionReaderSvc)
	pb.RegisterStudyPlanModifierServiceServer(grpcServer, s.studyPlanModifierSvc)
	pb.RegisterTopicReaderServiceServer(grpcServer, s.topicReaderSvc)
	pb.RegisterTopicModifierServiceServer(grpcServer, s.topicModifierSvc)
	pb.RegisterInternalModifierServiceServer(grpcServer, s.internalModifierSvc)
	pb.RegisterBookReaderServiceServer(grpcServer, s.bookReaderSvc)
	pb.RegisterBookModifierServiceServer(grpcServer, s.bookModifierSvc)
	pb.RegisterChapterReaderServiceServer(grpcServer, s.chapterReaderSvc)
	pb.RegisterQuizModifierServiceServer(grpcServer, s.quizModifierSvc)
	pb.RegisterChapterModifierServiceServer(grpcServer, s.chapterModifierSvc)
	pb.RegisterFlashCardReaderServiceServer(grpcServer, s.flashcardReaderSvc)
	pb.RegisterQuizReaderServiceServer(grpcServer, s.quizReaderSvc)
	pb.RegisterLearningObjectiveModifierServiceServer(grpcServer, s.learningObjectiveModifierSvc)
	pb.RegisterStudentEventLogModifierServiceServer(grpcServer, s.studentEventLogModifierSvc)
	pb.RegisterVisionReaderServiceServer(grpcServer, s.visionReaderSvc)
	pb.RegisterImageToTextServer(grpcServer, s.imageToText)
	pb.RegisterStudyPlanItemReaderServiceServer(grpcServer, s.studyPlanItemReaderSvc)
	sspb.RegisterQuestionServiceServer(grpcServer, s.questionSvc)
	sspb.RegisterFlashcardServer(grpcServer, s.flashcardSvc)
	sspb.RegisterAssignmentServer(grpcServer, s.assignmentSvc)
	sspb.RegisterLearningMaterialServer(grpcServer, s.learningMaterialSvc)
	sspb.RegisterLearningObjectiveServer(grpcServer, s.learningObjectiveSvc)
	sspb.RegisterExamLOServer(grpcServer, s.examLOSvc)
	sspb.RegisterStatisticsServer(grpcServer, s.statisticSvc)
	sspb.RegisterStudyPlanServer(grpcServer, s.studyPlanSvc)
	sspb.RegisterTaskAssignmentServer(grpcServer, s.taskAssignmentSvc)
	sspb.RegisterQuizServer(grpcServer, s.quizSvc)
	sspb.RegisterStudentSubmissionServiceServer(grpcServer, s.studentSubmissionSvc)
	sspb.RegisterQuestionTagServer(grpcServer, s.questionTagSvc)
	sspb.RegisterQuestionTagTypeServer(grpcServer, s.questionTagTypeSvc)
	sspb.RegisterLearningHistoryDataSyncServiceServer(grpcServer, s.learningHistoryDataSyncSvc)
	sspb.RegisterAssessmentServer(grpcServer, s.asssessmentSvc)
	sspb.RegisterItemsBankServiceServer(grpcServer, s.itemsBankSvc)
	pb.RegisterStudentServiceServer(grpcServer, s.studentSvc)
	pb.RegisterAssessmentSessionServiceServer(grpcServer, s.assessmentSessionSvc)

	health.RegisterHealthServer(grpcServer, &healthcheck.Service{
		DB: rsc.DB().DB.(*pgxpool.Pool),
	})

	return s.SetupGRPCV2(ctx, grpcServer)
}

func (s *server) SetupGRPCV2(_ context.Context, grpcServer *grpc.Server) error {
	pbv2.RegisterBookServiceServer(grpcServer, s.v2.bookSvc)
	pbv2.RegisterLearningMaterialServiceServer(grpcServer, s.v2.learningMaterialSvc)
	pbv2.RegisterAssessmentServiceServer(grpcServer, s.v2.assessmentSvc)
	pbv2.RegisterCourseServiceServer(grpcServer, s.v2.courseSvc)
	pbv2.RegisterItemBankServiceServer(grpcServer, s.v2.itemBankSvc)
	pbv2.RegisterStudyPlanItemServiceServer(grpcServer, s.v2.studyPlanItemSvc)
	pbv2.RegisterStudyPlanServiceServer(grpcServer, s.v2.studyPlanSvc)
	pbv2.RegisterCerebryServiceServer(grpcServer, s.v2.cerebrySvc)

	return nil
}

func (s *server) GracefulShutdown(_ context.Context) {}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	mux := gateway.NewServeMux(
		gateway.WithOutgoingHeaderMatcher(clients.IsHeaderAllowed),
		gateway.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			authHeader := request.Header.Get("Authorization")
			pkgHeader := request.Header.Get("pkg")
			versionHeader := request.Header.Get("version")

			// add defaults for pkg header
			if pkgHeader == "" {
				pkgHeader = "com.manabie.liz"
			}

			// add defaults for version header
			if versionHeader == "" {
				versionHeader = "1.0.0"
			}

			md := metadata.Pairs(
				"token", authHeader,
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

	err := setupGrpcGateway(mux, rsc.GetGRPCPort(c.Common.Name))
	if err != nil {
		return fmt.Errorf("error setupGrpcGateway %s", err)
	}

	superGroup := r.Group("/syllabus/api/v1")
	{
		superGroup.Group("/proxy/*{grpc_gateway}").Any("", gin.WrapH(mux))
	}
	return nil
}

func setupGrpcGateway(mux *gateway.ServeMux, port string) error {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(&tracer.B3Handler{ClientHandler: &ocgrpc.ClientHandler{}}),
	}

	serviceMap := map[string]func(context.Context, *gateway.ServeMux, string, []grpc.DialOption) error{
		"ItemsBankService": sspb.RegisterItemsBankServiceHandlerFromEndpoint,
	}

	for _, registerFunc := range serviceMap {
		err := registerFunc(context.Background(), mux, fmt.Sprintf("localhost%s", port), dialOpts)
		if err != nil {
			return err
		}
	}

	return nil
}
