package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type SystemNotificationContentRepo struct{}

func (*SystemNotificationContentRepo) FindBySystemNotificationIDs(ctx context.Context, db database.QueryExecer, systemNotificationIDs []string) (model.SystemNotificationContents, error) {
	ctx, span := interceptors.StartSpan(ctx, "FindBySystemNotificationIDs")
	defer span.End()
	e := &model.SystemNotificationContent{}
	query := fmt.Sprintf(`
		SELECT %s
		FROM system_notification_contents snc
		WHERE snc.system_notification_id = ANY($1::TEXT[])
		AND snc.deleted_at IS NULL;
	`, strings.Join(database.GetFieldNames(e), ","))

	systemNotificationContents := model.SystemNotificationContents{}
	err := database.Select(ctx, db, query, database.TextArray(systemNotificationIDs)).ScanAll(&systemNotificationContents)
	if err != nil {
		return nil, fmt.Errorf("failed ScanAll: %+v", err)
	}

	return systemNotificationContents, nil
}

func (*SystemNotificationContentRepo) SoftDeleteBySystemNotificationID(ctx context.Context, db database.QueryExecer, systemNotificationID string) error {
	ctx, span := interceptors.StartSpan(ctx, "SoftDeleteBySystemNotificationID")
	defer span.End()

	query := `
		UPDATE system_notification_contents
		SET deleted_at = now()
		WHERE system_notification_id = $1 AND deleted_at IS NULL;
	`
	_, err := db.Exec(ctx, query, database.Text(systemNotificationID))
	if err != nil {
		return fmt.Errorf("failed exec: %+v", err)
	}

	return nil
}

func (r *SystemNotificationContentRepo) BulkInsertSystemNotificationContents(ctx context.Context, db database.QueryExecer, systemNotificationContents model.SystemNotificationContents) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkInsertSystemNotificationContents")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range systemNotificationContents {
		err := r.queueInsert(b, item)
		if err != nil {
			return fmt.Errorf("repo.queueUpsert: %w", err)
		}
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (*SystemNotificationContentRepo) queueInsert(b *pgx.Batch, item *model.SystemNotificationContent) error {
	now := time.Now()
	err := multierr.Combine(
		item.CreatedAt.Set(now),
		item.UpdatedAt.Set(now),
		item.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT system_notification_contents_system_notification_content_id_pk
		DO NOTHING;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}
