package calendar

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	cld_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	cld_repo "github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

var resourcePath string
var userID string
var limit int
var limitPerJob int

type JobFillScheduler struct {
	wg      *sync.WaitGroup
	queue   chan JobFillSchedulerMessage
	service JobFillSchedulerService
	zLogger *zap.SugaredLogger
}

type JobFillSchedulerMessage struct {
	queue int
	ctx   context.Context
}

type JobFillSchedulerService struct {
	lessonDB      database.Ext
	calendarDB    database.Ext
	lessonRepo    lesson_repo.LessonRepo
	schedulerRepo cld_repo.SchedulerRepo
}

func NewJobFillScheduler(
	wg *sync.WaitGroup,
	queue chan JobFillSchedulerMessage,
	cfg *configurations.Config,
	rsc *bootstrap.Resources,
) *JobFillScheduler {
	zapLogger := logger.NewZapLogger("debug", cfg.Common.Environment == "local")
	lessonmgmtDBTrace := rsc.DBWith("lessonmgmt")
	calendarDBTrace := rsc.DBWith("calendar")

	syncJob := &JobFillScheduler{
		wg:      wg,
		queue:   queue,
		zLogger: zapLogger.Sugar(),
		service: JobFillSchedulerService{
			lessonDB:      lessonmgmtDBTrace,
			calendarDB:    calendarDBTrace,
			lessonRepo:    lesson_repo.LessonRepo{},
			schedulerRepo: cld_repo.SchedulerRepo{},
		},
	}

	return syncJob
}

func (s *JobFillScheduler) Publish(ctx context.Context) error {
	query := "SELECT COUNT(*) FROM lessons WHERE scheduler_id is null and deleted_at is null and resource_path = $1"
	args := []interface{}{
		golibs.ResourcePathFromCtx(ctx),
	}

	var totalLesson int
	if err := s.service.lessonDB.QueryRow(ctx, query, args...).Scan(&totalLesson); err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("row.Scan: %w", err)
	}
	if totalLesson > limitPerJob {
		totalLesson = limitPerJob
	}

	s.zLogger.Infof("========== Total lessons will be migrated: %d ==========", totalLesson)

	numOffset := totalLesson / limit
	if remaining := totalLesson % limit; remaining != 0 {
		numOffset++
	}
	offset := 0
	for i := 0; i < numOffset; i++ {
		s.wg.Add(1)
		msg := JobFillSchedulerMessage{
			queue: offset,
			ctx:   ctx,
		}
		s.queue <- msg
		offset += limit
	}
	return nil
}

func (s *JobFillScheduler) Start(ctx context.Context) error {
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			ResourcePath: resourcePath,
			UserID:       userID,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)

	return s.Publish(ctx)
}

func (s *JobFillScheduler) Subscribe() {
	for {
		msg, ok := <-s.queue
		if !ok {
			break
		}

		if err := s.ExecuteJob(msg.ctx, msg.queue); err != nil {
			s.zLogger.Errorf("ExecuteJob: %w", err)
		}
	}
}

func (s *JobFillScheduler) ExecuteJob(ctx context.Context, offSet int) error {
	defer s.wg.Done()

	lessons, err := s.service.lessonRepo.GetLessonsWithSchedulerNull(ctx, s.service.lessonDB, limit, offSet)
	if err != nil {
		return fmt.Errorf("get lessons with scheduler null failed: %w", err)
	}

	createSchedulersParams := sliceutils.Map(lessons, func(l *lesson_repo.Lesson) *cld_dto.CreateSchedulerParamWithIdentity {
		return &cld_dto.CreateSchedulerParamWithIdentity{
			ID: l.LessonID.String,
			CreateSchedulerParam: cld_dto.CreateSchedulerParams{
				SchedulerID: idutil.ULIDNow(),
				StartDate:   l.StartTime.Time,
				EndDate:     l.StartTime.Time,
				Frequency:   string(constants.FrequencyOnce),
			},
		}
	})

	schedulersMap, err := s.service.schedulerRepo.CreateMany(ctx, s.service.calendarDB, createSchedulersParams)

	if err != nil {
		return fmt.Errorf("create schedulers failed: %w", err)
	}

	err = s.service.lessonRepo.FillSchedulerToLessons(ctx, s.service.lessonDB, schedulersMap)

	if err != nil {
		return fmt.Errorf("fill scheduler to lesson fail: %w", err)
	}

	return nil
}

func init() {
	bootstrap.RegisterJob("fill_scheduler_to_lessons", jobMigrateFillSchedulerToLessons).
		Desc("migrate fill scheduler to lessons").
		StringVar(&resourcePath, "resourcePath", "", "orgId of partner").
		StringVar(&userID, "userID", "", "userID of school admin").
		IntVar(&limit, "limit", 1000, "limit lessons per time").
		IntVar(&limitPerJob, "limitPerJob", 1000000, "limit lessons per job")
	// limitPerJob should be divisible by limit
}

func jobMigrateFillSchedulerToLessons(ctx context.Context, cfg configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	zLogger := zapLogger.Sugar()

	var wg sync.WaitGroup
	msgChan := make(chan JobFillSchedulerMessage, 5)

	job := NewJobFillScheduler(&wg, msgChan, &cfg, rsc)
	job.zLogger = zLogger

	numberWorker := 5
	for i := 0; i < numberWorker; i++ {
		go job.Subscribe()
	}
	if err := job.Start(ctx); err != nil {
		return fmt.Errorf("error syncJob.Start %s", err)
	}
	wg.Wait()
	zLogger.Infof("========== Done to fill scheduler to lessons on partner %s ==========", resourcePath)
	return nil
}
