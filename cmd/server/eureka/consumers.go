package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/eureka/subscriptions"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
)

func runAllConsumersNatsJS(
	ctx context.Context,
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	c *configurations.Config,
	courseClassService *services.CourseClassService,
	courseStudentService *services.CourseStudentService,
	loStudentService *services.StudentService,
	studyPlanWriterSvc *services.ImportService,
	assignStudyPlanTaskService *services.AssignStudyPlanTaskModifierService,
	classStudentService *services.ClassStudentService,
	masterMgmtClassService *services.MasterMgmtClassService,
) error {
	jprepCourseClass := &subscriptions.JprepCourseClass{
		Logger:             logger,
		CourseClassService: courseClassService,
		JSM:                jsm,
	}
	if err := jprepCourseClass.Subscribe(ctx); err != nil {
		return fmt.Errorf("jprepSyncCourseClass.Subscribe: %w", err)
	}

	studentPackageEventSub := &subscriptions.StudentPackageEvent{
		Logger:                     logger,
		StudentPackageEventService: courseStudentService,
		JSM:                        jsm,
	}
	if err := studentPackageEventSub.Subscribe(ctx); err != nil {
		return fmt.Errorf("studentPackageEventSub.Subscribe: %w", err)
	}

	if err := studentPackageEventSub.SubscribeV2(ctx); err != nil {
		return fmt.Errorf("studentPackageEventSub.SubscribeV2: %w", err)
	}

	loEventSub := &subscriptions.LearningObjectivesCreatedHandler{
		JSM:             jsm,
		Logger:          logger,
		StudyPlanWriter: studyPlanWriterSvc,
	}
	if err := loEventSub.Subscribe(); err != nil {
		return fmt.Errorf("loEventSub.Subscribe: %w", err)
	}

	importStudyPlanItemsEventSub := &subscriptions.StudyPlanItemsImportedHandler{
		JSM:             jsm,
		Logger:          logger,
		StudyPlanWriter: studyPlanWriterSvc,
	}
	if err := importStudyPlanItemsEventSub.Subscribe(); err != nil {
		return fmt.Errorf("importStudyPlanItemsEventSub.Subscribe: %w", err)
	}

	asmEventSub := &subscriptions.AssignmentsCreatedHandler{
		JSM:             jsm,
		Logger:          logger,
		StudyPlanWriter: studyPlanWriterSvc,
	}
	if err := asmEventSub.Subscribe(); err != nil {
		return fmt.Errorf("asmEventSub.Subscribe: %w", err)
	}

	classEventSub := &subscriptions.ClassEvent{
		Logger:            logger,
		ClassEventService: classStudentService,
		JSM:               jsm,
	}
	if err := classEventSub.Subscribe(ctx); err != nil {
		return fmt.Errorf("classEventSub.Subscribe: %w", err)
	}

	masterMgmtClassSub := &subscriptions.MasterMgmtClassEvent{
		Logger:                      logger,
		MasterMgmtClassEventService: masterMgmtClassService,
		JSM:                         jsm,
	}
	if err := masterMgmtClassSub.Subscribe(ctx); err != nil {
		return fmt.Errorf("masterMgmtClassEventSub.Subscribe: %w", err)
	}

	return nil
}

var startTime, durableName string

func init() {
	bootstrap.RegisterJob("eureka_all_consumers", RunConsumers).StringVar(&startTime, "startTime", "", "RFC3999 format to start reading messages stream")
	bootstrap.RegisterJob("eureka_jprep_sync_course_student", RunJPREPSyncCourseStudentConsumer).StringVar(&startTime, "startTime", "", "RFC3999 format to start reading messages stream").
		StringVar(&durableName, "durableName", "", "Set durable name for consumer")
}

// RunConsumers runs Eureka
func RunConsumers(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	dbTrace := rsc.DB()

	pe, tp, err := interceptors.InitTelemetry(&c.Common, "eureka", 1)
	if err != nil {
		zapLogger.Fatal("interceptors.InitTelemetry", zap.Error(err))
	}
	defer tp.Shutdown(ctx)

	go interceptors.StartMetricHandler("/metrics", ":8888", pe)

	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}
	loItemStudyPlanItemRepo := &repositories.LoStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &repositories.AssignmentStudyPlanItemRepo{}
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	courseStudyPlanRepo := &repositories.CourseStudyPlanRepo{}
	bookChapterRepo := &repositories.BookChapterRepo{}
	studentRepo := &repositories.StudentRepo{}

	courseStudentService := &services.CourseStudentService{
		DB:                          dbTrace,
		CourseStudentRepo:           &repositories.CourseStudentRepo{},
		CourseStudentAccessPathRepo: &repositories.CourseStudentAccessPathRepo{},
		StudentStudyPlanRepo:        studentStudyPlanRepo,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		StudentRepo:                 studentRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		StudentStudyPlan:            studentStudyPlanRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loItemStudyPlanItemRepo,
		JSM:                         rsc.NATS(),
		Logger:                      zapLogger,
	}
	courseClassService := &services.CourseClassService{
		DB:              dbTrace,
		CourseClassRepo: &repositories.CourseClassRepo{},
	}
	classStudentService := &services.ClassStudentService{
		DB:               dbTrace,
		ClassStudentRepo: &repositories.ClassStudentRepo{},
	}
	loStudentService := &services.StudentService{
		DB:                  dbTrace,
		LoStudyPlanItemRepo: &repositories.LoStudyPlanItemRepo{},
	}
	masterMgmtClassService := &services.MasterMgmtClassService{
		DB:                         dbTrace,
		CourseClassRepo:            &repositories.CourseClassRepo{},
		MasterMgmtClassStudentRepo: &repositories.MasterMgmtClassStudentRepo{},
	}

	assignStudyPlanTaskService := &services.AssignStudyPlanTaskModifierService{
		DB:                          dbTrace,
		StudyPlanRepo:               studyPlanRepo,
		StudentRepo:                 studentRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		StudentStudyPlan:            studentStudyPlanRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loItemStudyPlanItemRepo,
		AssignStudyPlanTaskRepo:     &repositories.AssignStudyPlanTaskRepo{},
	}

	studyPlanWriterSvc := &services.ImportService{
		DB:                          dbTrace,
		Env:                         c.Common.Environment,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loItemStudyPlanItemRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		BookChapterRepo:             bookChapterRepo,
		ImportStudyPlanTaskRepo:     &repositories.ImportStudyPlanTaskRepo{},
		MasterStudyPlanRepo:         &repositories.MasterStudyPlanRepo{},
		IndividualStudyPlanRepo:     &repositories.IndividualStudyPlan{},
	}

	err = runAllConsumersNatsJS(ctx, rsc.NATS(), zapLogger, &c, courseClassService, courseStudentService, loStudentService, studyPlanWriterSvc, assignStudyPlanTaskService, classStudentService, masterMgmtClassService)
	if err != nil {
		zapLogger.Fatal("runAllConsumersNatsJS: ", zap.Error(err))
	}
	<-ctx.Done()
	zapLogger.Warn("shutting down consumer...")
	return nil
}

// RunJPREPSyncCourseStudentConsumer runs Eureka JPREP sync course student consumer.
func RunJPREPSyncCourseStudentConsumer(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	dbTrace := rsc.DB()

	pe, tp, err := interceptors.InitTelemetry(&c.Common, "eureka", 1)
	if err != nil {
		zapLogger.Fatal("interceptors.InitTelemetry", zap.Error(err))
	}
	defer tp.Shutdown(ctx)

	go interceptors.StartMetricHandler("/metrics", ":8888", pe)

	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}
	loItemStudyPlanItemRepo := &repositories.LoStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &repositories.AssignmentStudyPlanItemRepo{}
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	courseStudyPlanRepo := &repositories.CourseStudyPlanRepo{}
	studentRepo := &repositories.StudentRepo{}

	if err := view.Register(
		nats.JetstreamProcessedMessagesView,
		nats.JetstreamProcessedMessagesLatencyView,
	); err != nil {
		zapLogger.Panic("Failed to register ocgrpc server views", zap.Error(err))
	}

	courseStudentService := &services.CourseStudentService{
		DB:                          dbTrace,
		JSM:                         rsc.NATS(),
		CourseStudentRepo:           &repositories.CourseStudentRepo{},
		CourseStudentAccessPathRepo: &repositories.CourseStudentAccessPathRepo{},
		StudentStudyPlanRepo:        studentStudyPlanRepo,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		StudentRepo:                 studentRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		StudentStudyPlan:            studentStudyPlanRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loItemStudyPlanItemRepo,
	}

	sub := &subscriptions.JprepCourseStudent{
		JSM:                  rsc.NATS(),
		Logger:               zapLogger,
		CourseStudentService: courseStudentService,
	}
	if err := sub.Subscribe(ctx); err != nil {
		zapLogger.Fatal("failed to connect to Subscribe", zap.Error(err))
	}
	<-ctx.Done()
	zapLogger.Warn("shutting down consumer...")
	return nil
}
