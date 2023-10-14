package lessonmgmt

import (
	"context"
	"fmt"
	"sync"
	"time"

	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	elastic_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/elasticsearch"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	user_repo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

const (
	limit         = 200
	layout string = "2006-01-02"
)

type SyncJob struct {
	wg      *sync.WaitGroup
	queue   chan Message
	service Service
	input   *SyncDate
	zLogger *zap.SugaredLogger
}
type SyncDate struct {
	fromDate, toDate string
}
type Message struct {
	queue int
	ctx   context.Context
}

type Service struct {
	db            database.Ext
	searchIndexer *application.LessonSearchIndexer
}

func NewSyncJob(ctx context.Context,
	wg *sync.WaitGroup,
	queue chan Message,
	cfg *configurations.Config,
) *SyncJob {
	zapLogger := logger.NewZapLogger("debug", cfg.Common.Environment == "local")
	syncJob := &SyncJob{
		wg:      wg,
		queue:   queue,
		zLogger: zapLogger.Sugar(),
	}

	searchClient, err := elastic.NewSearchFactory(
		zapLogger,
		cfg.ElasticSearch.Addresses,
		cfg.ElasticSearch.Username,
		cfg.ElasticSearch.Password, "", "")
	if err != nil {
		syncJob.zLogger.Fatal("unable to connect elasticsearch", zap.Error(err))
	}

	// We should have saved and called dbcancel to clean up resources
	// after the job finishes, but that'll be a TODO
	dbPool, _, err := database.NewPool(ctx, zapLogger, cfg.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}

	dbTrace := &database.DBTrace{DB: dbPool}
	searchIndexer := &application.LessonSearchIndexer{
		Logger: zapLogger,
		DB:     dbTrace,
		SearchRepo: &elastic_repo.SearchRepo{
			SearchFactory: searchClient,
		},
		LessonRepo:  &repo.LessonRepo{},
		StudentRepo: &user_repo.StudentRepo{},
		UserRepo:    &bob_repo.UserRepo{},
	}

	syncJob.service = Service{
		db:            dbTrace,
		searchIndexer: searchIndexer,
	}
	return syncJob
}

func (s *SyncJob) Publish(ctx context.Context) error {
	query := "SELECT COUNT(*) FROM lessons WHERE resource_path = $1"
	args := []interface{}{
		golibs.ResourcePathFromCtx(ctx),
	}
	if s.input != nil {
		args = append(args, s.input.fromDate, s.input.toDate)
		query += " and cast(updated_at as date) between $2 and $3"
	}
	var totalLesson int
	if err := s.service.db.QueryRow(ctx, query, args...).Scan(&totalLesson); err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("row.Scan: %w", err)
	}
	if s.input != nil {
		s.zLogger.Infof("total the sync lessons to elasticsearch from %s to %s: %d", s.input.fromDate, s.input.toDate, totalLesson)
	}

	numOffset := totalLesson / limit
	if remaining := totalLesson % limit; remaining != 0 {
		numOffset += 1
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

func (s *SyncJob) ResetLessonIndex(c *configurations.Config) error {
	zapLogger := logger.NewZapLogger("debug", c.Common.Environment == "local")

	zapLogger.Info(`release "lesson" index exist. Delete lesson index`)
	if err := s.service.searchIndexer.SearchRepo.DeleteLessonIndex(); err != nil {
		zapLogger.Error("unable delete lesson index", zap.Error(err))
		return err
	}
	zapLogger.Info(`"lesson" index deleted!`)

	zapLogger.Info(`release "lesson" index does not exist. Creating it now`)
	if err := s.service.searchIndexer.SearchRepo.CreateLessonIndex(); err != nil {
		zapLogger.Error("unable to create lesson index", zap.Error(err))
		return err
	}

	zapLogger.Info(`"lesson" index created!`)
	return nil
}

func (s *SyncJob) Start(ctx context.Context) error {
	orgQuery := "select organization_id from organizations"
	organizations, err := s.service.db.Query(ctx, orgQuery)
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
		s.Publish(ctxOrg)
	}
	return nil
}

func (s *SyncJob) Subscribe() {
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

func (s *SyncJob) ExecuteJob(ctx context.Context, offSet int) error {
	defer s.wg.Done()
	query := "SELECT lesson_id FROM lessons WHERE resource_path = $1 "
	args := []interface{}{
		golibs.ResourcePathFromCtx(ctx),
	}
	paramsNum := 1
	if s.input != nil {
		query += fmt.Sprintf(" and cast(updated_at as date) between $%d and $%d ", paramsNum+1, paramsNum+2)
		paramsNum += 2
		args = append(args, s.input.fromDate, s.input.toDate)
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramsNum+1, paramsNum+2)
	args = append(args, limit, offSet)

	rows, err := s.service.db.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	lessonIDs := []string{}
	for rows.Next() {
		var lessonID string
		if err := rows.Scan(&lessonID); err != nil {
			return fmt.Errorf("row.Scan: %w", err)
		}
		lessonIDs = append(lessonIDs, lessonID)
	}
	if rows.Err() != nil {
		return fmt.Errorf("row.Err: %w", err)
	}
	err = s.service.searchIndexer.SyncLessonIndex(ctx, lessonIDs)
	if err != nil {
		return fmt.Errorf("fail SyncLessonIndex in org: %s, offset: %d, limit: %d: %s", golibs.ResourcePathFromCtx(ctx), offSet, limit, err)
	}
	s.zLogger.Infof("the total of sync lessons success in org: %s, offset: %d, limit: %d: %d", golibs.ResourcePathFromCtx(ctx), offSet, limit, len(lessonIDs))
	return nil
}

var (
	// used for manual sync Lesson data in db to ElasticSearch
	// for job "sync_lesson_data_to_elasticsearch_by_date"
	fromDate string
	toDate   string
)

func init() {
	bootstrap.RegisterJob("sync_lesson_data_to_elasticsearch_by_date", syncLessonsToElasticSearchByDate).
		Desc("job sync lesson data to elasticsearch by date").
		StringVar(&fromDate, "fromDate", "", "from date on the updated field in the lesson table which be desired").
		StringVar(&toDate, "toDate", "", "to date on the updated field in the lesson table which be desired")
}

func syncLessonsToElasticSearchByDate(ctx context.Context, cfg configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	zLogger := zapLogger.Sugar()
	zLogger.Infof("start sync lessons to elasticsearch from %s to %s", fromDate, toDate)

	if err := checkSyncDate(fromDate, toDate); err != nil {
		return fmt.Errorf("check sync date: %s", err)
	}
	var wg sync.WaitGroup
	msgChan := make(chan Message, 5)

	syncJob := NewSyncJob(ctx, &wg, msgChan, &cfg)
	syncJob.input = &SyncDate{fromDate: fromDate, toDate: toDate}
	syncJob.zLogger = zLogger

	numberWorker := 5
	for i := 0; i < numberWorker; i++ {
		go syncJob.Subscribe()
	}
	if err := syncJob.Start(ctx); err != nil {
		return fmt.Errorf("error syncJob.Start %s", err)
	}
	wg.Wait()
	zLogger.Infof("complete sync lessons to elasticsearch from %s to %s", fromDate, toDate)
	return nil
}

func checkSyncDate(fromDate, toDate string) (err error) {
	if fromDate == "" || toDate == "" {
		return fmt.Errorf("from_date and to_date is required")
	}
	start, err := time.Parse(layout, fromDate)
	if err != nil {
		return err
	}
	end, err := time.Parse(layout, toDate)
	if err != nil {
		return err
	}

	if end.Sub(start).Hours() < 0 {
		return fmt.Errorf("Start date: %s must come before End date: %s", fromDate, toDate)
	}

	return nil
}
