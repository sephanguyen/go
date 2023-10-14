package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type ActivityLogRepo struct{}

func (r *ActivityLogRepo) Get(ctx context.Context, db database.Ext, ids []string) ([]*bob_entities.ActivityLog, error) {
	var p []*bob_entities.ActivityLog
	e := &bob_entities.ActivityLog{}
	fieldNames := database.GetFieldNames(e)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE activity_log_id = ANY($1)", strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, stmt, ids).ScanOne(e)
	if err != nil {
		return nil, fmt.Errorf("ActivityLogRepo.Get:%w", err)
	}
	return p, nil
}

func (r *ActivityLogRepo) FindByUserIDAndType(ctx context.Context, db database.Ext, userID, actionType string) (*bob_entities.ActivityLog, error) {
	e := &bob_entities.ActivityLog{}
	fieldNames := database.GetFieldNames(e)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND action_type = $2", strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, stmt, userID, actionType).ScanOne(e)
	if err != nil {
		return nil, fmt.Errorf("ActivityLogRepo.FindByUserIDAndType:%w", err)
	}
	return e, nil
}

func (r *ActivityLogRepo) FindByFilter(ctx context.Context, db database.Ext, systemSchoolID int32, nonSystemSchoolIds []int32) ([]*bob_entities.ActivityLog, error) {
	var p bob_entities.ActivityLogs
	query := "SELECT activity_log_id, user_id, action_type, created_at, updated_at, payload, deleted_at FROM activity_logs WHERE payload -> 'system_school_id' = $1 AND payload ->> 'non_system_school_id' =array_to_json($2::int[])::text "
	err := database.Select(ctx, db, query, systemSchoolID, database.Int4Array(nonSystemSchoolIds)).ScanAll(&p)
	if err != nil {
		return nil, fmt.Errorf("ActivityLogRepo.FindByFilter: %w", err)
	}
	return p, nil
}

func (r *ActivityLogRepo) BulkCreate(ctx context.Context, db database.Ext, logs []*bob_entities.ActivityLog) error {
	for _, log := range logs {
		if err := r.CreateV2(ctx, db, log); err != nil {
			return err
		}
	}
	return nil
}

func (r *ActivityLogRepo) CreateV2(ctx context.Context, db database.Ext, log *bob_entities.ActivityLog) error {
	ctx, span := interceptors.StartSpan(ctx, "ActivityLogRepo.CreateV2")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		log.CreatedAt.Set(now),
		log.UpdatedAt.Set(now),
		log.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	if log.ID.String == "" {
		err = log.ID.Set(idutil.ULIDNow())
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}
	}
	cmdTag, err := database.Insert(ctx, log, db.Exec)
	if err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot insert new ActivityLog")
	}

	return nil
}

func (r *ActivityLogRepo) Upsert(ctx context.Context, db database.QueryExecer, ens []*bob_entities.ActivityLog) error {
	ctx, span := interceptors.StartSpan(ctx, "ActivityLogRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *bob_entities.ActivityLog) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO activity_logs (%s) VALUES (%s)", strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range ens {
		t.CreatedAt.Set(now)
		t.UpdatedAt.Set(now)

		queue(b, t)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(ens); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("book chapter not inserted")
		}
	}
	return nil
}
