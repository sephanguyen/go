package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type EmailRecipientRepo struct{}

func (repo *EmailRecipientRepo) queueUpsert(b *pgx.Batch, emailRecipient *model.EmailRecipient) error {
	now := time.Now()
	err := multierr.Combine(
		emailRecipient.CreatedAt.Set(now),
		emailRecipient.UpdatedAt.Set(now),
		emailRecipient.DeletedAt.Set(nil),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fieldNames := database.GetFieldNames(emailRecipient)
	values := database.GetScanFields(emailRecipient, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := emailRecipient.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s AS er (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__email_recipients 
		DO UPDATE SET 
			recipient_address = EXCLUDED.recipient_address,
			updated_at = EXCLUDED.updated_at
		WHERE er.deleted_at IS NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *EmailRecipientRepo) BulkUpsertEmailRecipients(ctx context.Context, db database.QueryExecer, emailRecipients model.EmailRecipients) error {
	ctx, span := interceptors.StartSpan(ctx, "EmailRecipientRepo.BulkUpsertEmailRecipients")
	defer span.End()
	b := &pgx.Batch{}
	for _, item := range emailRecipients {
		err := repo.queueUpsert(b, item)
		if err != nil {
			return fmt.Errorf("repo.queueForceUpsert: %w", err)
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

func (repo *EmailRecipientRepo) GetEmailRecipientsByEmailID(ctx context.Context, db database.QueryExecer, emailID string) (model.EmailRecipients, error) {
	ctx, span := interceptors.StartSpan(ctx, "GetEmailRecipientsByEmailID")
	defer span.End()
	emailRecipientEnt := &model.EmailRecipient{}
	fields := strings.Join(database.GetFieldNames(emailRecipientEnt), ", er.")
	query := fmt.Sprintf(`
		SELECT %s
		FROM email_recipients er
		WHERE er.email_id = $1
			AND er.deleted_at IS NULL;
	`, fields)

	emailRecipients := model.EmailRecipients{}
	err := database.Select(ctx, db, query, emailID).ScanAll(&emailRecipients)
	if err != nil {
		return nil, err
	}

	return emailRecipients, nil
}
