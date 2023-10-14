package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"

	"go.uber.org/zap"
)

type ConfigJob struct {
	TotalJobConcurrency int
	Limit               int
}

type LabelExecutor string

type Message struct {
	queue int
	ctx   context.Context
}

type Job struct {
	limit    int
	db       database.Ext
	cfg      *configurations.Config
	zLogger  *zap.SugaredLogger
	wg       *sync.WaitGroup
	queue    chan Message
	Executor Executor

	// cleanup cleans up resources, such as database connections
	// it should be called after executing job
	cleanup func()
}

func getExecutor(db database.Ext, logger *zap.Logger, label LabelExecutor) Executor {
	switch label {
	case LESSON_REPORT_EXECUTOR:
		return InitLessonReportExecutor(db, logger)
	case LESSON_MEMBERS_EXECUTOR:
		return InitLessonMembersExecutor(db, logger)
	case LESSON_TEACHERS_EXECUTOR:
		return InitLessonTeacherExecutor(db, logger)
	case LESSON_STUDENT_SUBSCRIPTIONS_EXECUTOR:
		return InitLessonStudentSubscriptionExecutor(db, logger)
	}
	return nil
}

func InitJob(ctx context.Context, cfg *configurations.Config, cf *ConfigJob, label LabelExecutor) *Job {
	zapLogger := logger.NewZapLogger(cfg.Common.Log.ApplicationLevel, cfg.Common.Environment == "local")
	zLogger := zapLogger.Sugar()
	var wg sync.WaitGroup
	msgChan := make(chan Message, cf.TotalJobConcurrency)

	lessonPool, dbcancel, err := database.NewPool(ctx, zapLogger, cfg.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}

	dbTrace := &database.DBTrace{DB: lessonPool}
	limit := cf.Limit
	if limit == 0 {
		limit = 200
	}
	Job := &Job{
		cfg:      cfg,
		zLogger:  zLogger,
		db:       dbTrace,
		wg:       &wg,
		queue:    msgChan,
		limit:    limit,
		Executor: getExecutor(dbTrace, zapLogger, label),
		cleanup: func() {
			if err := dbcancel(); err != nil {
				zapLogger.Error("dbcancel() failed", zap.Error(err))
			}
		},
	}

	for i := 0; i < cf.TotalJobConcurrency; i++ {
		go Job.subscribe()
	}
	return Job
}

func (s *Job) publish(ctx context.Context) error {
	if err := s.Executor.PreExecute(ctx); err != nil {
		return fmt.Errorf("error syncJob.PreExecute %s", err)
	}
	total, err := s.Executor.GetTotal(ctx)
	if total < 1 {
		return nil
	}
	if err != nil {
		s.zLogger.Errorf("publish: %w", err)
	}
	resourceID := golibs.ResourcePathFromCtx(ctx)
	limit := s.limit

	numOffset := total / limit
	s.zLogger.Infof("Get total: %d of %s", total, resourceID)

	if remaining := total % limit; remaining != 0 {
		numOffset++
	}
	offset := 0
	for i := 0; i < numOffset; i++ {
		s.wg.Add(1)
		msg := Message{
			queue: offset,
			ctx:   ctx,
		}
		s.queue <- msg
		offset += limit
	}
	return nil
}

func (s *Job) subscribe() {
	for {
		msg, ok := <-s.queue
		if !ok {
			break
		}
		limit := s.limit

		fExecute := func() error {
			defer s.wg.Done()
			return s.Executor.ExecuteJob(msg.ctx, limit, msg.queue)
		}

		if err := fExecute(); err != nil {
			s.zLogger.Errorf("ExecuteJob: %w", err)
		}
	}
}

func (s *Job) start(ctx context.Context) error {
	orgQuery := "select organization_id from organizations"
	organizations, err := s.db.Query(ctx, orgQuery)
	if err != nil {
		return fmt.Errorf("failed to get organization:%w", err)
	}
	defer organizations.Close()
	organizationIDs := []string{}
	for organizations.Next() {
		var organizationID string
		err := organizations.Scan(&organizationID)
		if err != nil {
			return fmt.Errorf("failed to scan organization:%w", err)
		}
		organizationIDs = append(organizationIDs, organizationID)
	}
	if err := organizations.Err(); err != nil {
		return err
	}
	for _, org := range organizationIDs {
		ctxOrg := auth.InjectFakeJwtToken(ctx, org)
		err = s.publish(ctxOrg)
		if err != nil {
			return fmt.Errorf("failed to publish:%w", err)
		}
	}
	return nil
}

func (s *Job) Run(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		return fmt.Errorf("error syncJob.Run %s", err)
	}
	s.wg.Wait()
	if s.cleanup != nil {
		s.cleanup()
	}
	s.zLogger.Infof("complete sync job")
	return nil
}
