package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/zeus/entities"

	"github.com/jackc/pgx/v4"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
)

type ActivityLogRepo struct{}

func (r *ActivityLogRepo) Create(ctx context.Context, db database.QueryExecer, en *entities.ActivityLog) error {
	now := time.Now()
	err := multierr.Combine(
		en.CreatedAt.Set(now),
		en.UpdatedAt.Set(now),
		en.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	if en.ID.String == "" {
		err = en.ID.Set(ksuid.New().String())
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}
	}

	cmdTag, err := database.Insert(ctx, en, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new ActivityLog")
	}

	return nil
}

func (r *ActivityLogRepo) CreateBulk(ctx context.Context, db database.QueryExecer, logs []*entities.ActivityLog) error {
	now := time.Now()
	queueFn := func(b *pgx.Batch, log *entities.ActivityLog) {
		fieldNames := database.GetFieldNames(log)
		placeHodlers := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", log.TableName(), strings.Join(fieldNames, ","), placeHodlers)
		b.Queue(query, database.GetScanFields(log, fieldNames)...)
	}

	b := &pgx.Batch{}
	for _, l := range logs {
		if err := multierr.Combine(
			l.ID.Set(ksuid.New().String()),
			l.CreatedAt.Set(now),
			l.UpdatedAt.Set(now),
			l.DeletedAt.Set(nil),
		); err != nil {
			return err
		}

		queueFn(b, l)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(logs); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}

		if ct.RowsAffected() != 1 {
			return fmt.Errorf("activity-log is not inserted")
		}
	}

	return nil
}
