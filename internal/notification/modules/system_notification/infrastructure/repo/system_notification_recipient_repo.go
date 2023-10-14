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

type SystemNotificationRecipientRepo struct{}

func (r *SystemNotificationRecipientRepo) BulkInsertSystemNotificationRecipients(ctx context.Context, db database.QueryExecer, systemNotificationRecipients model.SystemNotificationRecipients) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkInsertSystemNotificationRecipients.BulkInsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range systemNotificationRecipients {
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

func (*SystemNotificationRecipientRepo) queueInsert(b *pgx.Batch, item *model.SystemNotificationRecipient) error {
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
		ON CONFLICT ON CONSTRAINT pk__system_notification_recipients
		DO NOTHING;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (*SystemNotificationRecipientRepo) SoftDeleteBySystemNotificationID(ctx context.Context, db database.QueryExecer, systemNotificationID string) error {
	ctx, span := interceptors.StartSpan(ctx, "DeleteBySystemNotificationID")
	defer span.End()

	query := `
		UPDATE system_notification_recipients
		SET deleted_at = now()
		WHERE system_notification_id = $1 AND deleted_at IS NULL;
	`
	_, err := db.Exec(ctx, query, database.Text(systemNotificationID))
	if err != nil {
		return fmt.Errorf("failed exec: %+v", err)
	}

	return nil
}
