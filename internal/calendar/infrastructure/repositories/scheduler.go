package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type SchedulerRepo struct{}

func (s *SchedulerRepo) Create(ctx context.Context, db database.QueryExecer, params *dto.CreateSchedulerParams) (id string, err error) {
	ctx, span := interceptors.StartSpan(ctx, "SchedulerRepo.Create")
	defer span.End()
	sch, err := NewScheduler(map[string]interface{}{
		"scheduler_id": params.SchedulerID,
		"start_date":   params.StartDate,
		"end_date":     params.EndDate,
		"frequency":    params.Frequency,
	})
	if err != nil {
		return
	}
	if err = sch.PreInsert(); err != nil {
		return
	}
	fields, args := sch.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, sch.TableName(), strings.Join(fields, ","), placeHolders)
	_, err = db.Exec(ctx, query, args...)
	if err != nil {
		err = fmt.Errorf("db.Exec: %w", err)
		return
	}
	id = params.SchedulerID
	return
}

func (s *SchedulerRepo) CreateMany(ctx context.Context, db database.QueryExecer, params []*dto.CreateSchedulerParamWithIdentity) (map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchedulerRepo.CreateMany")
	defer span.End()
	b := &pgx.Batch{}
	emptyMap := map[string]string{}
	for _, param := range params {
		sch, err := NewScheduler(map[string]interface{}{
			"scheduler_id": param.CreateSchedulerParam.SchedulerID,
			"start_date":   param.CreateSchedulerParam.StartDate,
			"end_date":     param.CreateSchedulerParam.EndDate,
			"frequency":    param.CreateSchedulerParam.Frequency,
		})
		if err != nil {
			return emptyMap, err
		}
		if err = sch.PreInsert(); err != nil {
			return emptyMap, err
		}
		fields, args := sch.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING scheduler_id, '%s' as identity",
			sch.TableName(),
			strings.Join(fields, ","),
			placeHolders,
			param.ID)
		b.Queue(query, args...)
	}
	batchResults := db.SendBatch(ctx, b)
	mapSchedulers := make(map[string]string, b.Len())
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		var schedulerID, ID pgtype.Text
		row := batchResults.QueryRow()
		err := row.Scan(&schedulerID, &ID)
		if err != nil {
			return emptyMap, err
		}
		mapSchedulers[ID.String] = schedulerID.String
	}
	return mapSchedulers, nil
}

func (s *SchedulerRepo) Update(ctx context.Context, db database.QueryExecer, params *dto.UpdateSchedulerParams, updatedFields []string) error {
	ctx, span := interceptors.StartSpan(ctx, "SchedulerRepo.Update")
	defer span.End()
	sch, err := NewScheduler(map[string]interface{}{
		"scheduler_id": params.SchedulerID,
		"end_date":     params.EndDate,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := sch.PreUpdate(); err != nil {
		return fmt.Errorf("got error when preUpdate scheduler dto: %w", err)
	}
	_, err = database.UpdateFields(ctx, sch, db.Exec, "scheduler_id", updatedFields)
	return err
}

func (s *SchedulerRepo) GetByID(ctx context.Context, db database.QueryExecer, schedulerID string) (*dto.Scheduler, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchedulerRepo.GetByID")
	defer span.End()
	scheduler := &Scheduler{}
	fields, values := scheduler.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE scheduler_id = $1 AND deleted_at is null",
		strings.Join(fields, ", "), scheduler.TableName())
	err := db.QueryRow(ctx, query, schedulerID).Scan(values...)
	if err != nil {
		return nil, err
	}
	return &dto.Scheduler{
		SchedulerID: scheduler.SchedulerID.String,
		StartDate:   scheduler.StartDate.Time,
		EndDate:     scheduler.EndDate.Time,
		Frequency:   scheduler.Frequency.String,
	}, nil
}
