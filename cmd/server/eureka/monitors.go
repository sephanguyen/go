package eureka

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	monitor_repo "github.com/manabie-com/backend/internal/eureka/repositories/monitors"
	services "github.com/manabie-com/backend/internal/eureka/services/monitoring"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("eureka_monitors", RunMonitors)
}

//nolint:gocritic
func RunMonitors(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	// graceful shutdown handling
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zapLogger := rsc.Logger()
	szapLogger := zapLogger.Sugar()

	schoolName := c.SchoolInformation.SchoolName
	schoolID := c.SchoolInformation.SchoolID

	szapLogger.Infof("school name ----: %s", schoolName)
	szapLogger.Infof("school id ----: %s", schoolID)
	// TODO: check schoolID and schoolName, since eureka don't have table `organizations` so skip it.
	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, schoolID)
	dbTrace := rsc.DB()

	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}
	courseStudentRepo := &repositories.CourseStudentRepo{}
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studyPlanMonitorRepo := &monitor_repo.StudyPlanMonitorRepo{}
	learningObjectiveRepo := &repositories.LearningObjectiveRepo{}
	assignmentRepo := &repositories.AssignmentRepo{}
	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	loStudyPlanItemRepo := &repositories.LoStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &repositories.AssignmentStudyPlanItemRepo{}
	httpClient := http.Client{Timeout: time.Duration(10) * time.Second}

	alertClient := &alert.SlackImpl{
		WebHookURL: c.SyllabusSlackWebHook,
		HTTPClient: httpClient,
	}
	studyPlanMonitorService := &services.StudyPlanMonitorService{
		DB:                          dbTrace,
		Logger:                      *zapLogger,
		Cfg:                         &c,
		Alert:                       alertClient,
		StudentStudyPlanRepo:        studentStudyPlanRepo,
		CourseStudentRepo:           courseStudentRepo,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanMonitorRepo:        studyPlanMonitorRepo,
		AssignmentRepo:              assignmentRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		LearningObjectiveRepo:       learningObjectiveRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
	}
	cronJob := cron.New()

	// TODO: cancel if the job still running
	// TODO: add time -> DONE
	_, err := cronJob.AddFunc(genCronTimeFormat(c.SyllabusTimeMonitor.CourseStudentUpserted), func() {
		zapLogger.Info(`starting to validate UpsertStudentCourse...`)
		err := studyPlanMonitorService.UpsertStudentCourse(ctx, c.SyllabusTimeMonitor.CourseStudentUpserted)
		if err != nil {
			zapLogger.Error(`error when validate UpsertStudentCourse`, zap.Error(err))
		} else {
			zapLogger.Info(`Validate UpsertStudentCourse completed!`)
		}
	})
	if err != nil {
		zapLogger.Error(`error when add func validate UpsertStudentCourse`, zap.Error(err))
	}

	_, err = cronJob.AddFunc(genCronTimeFormat(c.SyllabusTimeMonitor.CourseStudentUpserted), func() {
		zapLogger.Info(`starting to validate UpsertLearningItems...`)
		err := studyPlanMonitorService.UpsertLearningItems(ctx, c.SyllabusTimeMonitor.LearningItemUpserted, schoolID)
		if err != nil {
			zapLogger.Error(`error when validate UpsertLearningItems`, zap.Error(err))
		} else {
			zapLogger.Info(`Validate UpsertLearningItems completed!`)
		}
	})
	if err != nil {
		zapLogger.Error(`error when add func validate UpsertLearningItems`, zap.Error(err))
	}

	cronJob.Start()

	// graceful shutdown

	chn := make(chan os.Signal, 1)
	signal.Notify(chn, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-chn
		zapLogger.Warn("shutting down monitoring server...")
		cronJob.Stop()
	}()
	for {
		time.Sleep(time.Second)
	}
}

// alway minutes
// genCronTimeFormat this format run every t minutes
func genCronTimeFormat(t int) (cronTimeFormat string) {
	cronTimeFormat = fmt.Sprintf("*/%d * * * *", t)
	return
}
