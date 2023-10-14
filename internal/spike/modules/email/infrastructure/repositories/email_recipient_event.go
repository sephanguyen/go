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

type EmailRecipientEventRepo struct{}

func (*EmailRecipientEventRepo) GetMapEventsByEventsAndEmailRecipientIDs(ctx context.Context, db database.QueryExecer, events, emailRecipientIDs []string) (map[string]*model.EmailRecipientEvent, error) {
	ctx, span := interceptors.StartSpan(ctx, "GetMapEventsByEventsAndEmailRecipientIDs")
	defer span.End()

	emailRecipientEventEnt := &model.EmailRecipientEvent{}
	fields := strings.Join(database.GetFieldNames(emailRecipientEventEnt), ", ere.")

	query := fmt.Sprintf(`
		SELECT %s
		FROM email_recipient_events ere
		WHERE ere.event = ANY($1::TEXT[]) 
			AND ere.email_recipient_id = ANY($2::TEXT[])
			AND ere.deleted_at IS NULL;
	`, fields)

	rows, err := db.Query(ctx, query, database.TextArray(events), database.TextArray(emailRecipientIDs))
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	mapEventByEventAndEmailRecipientID := make(map[string]*model.EmailRecipientEvent)
	for rows.Next() {
		emailRecipientEvent := &model.EmailRecipientEvent{}
		scanField := database.GetScanFields(emailRecipientEvent, database.GetFieldNames(emailRecipientEvent))
		err = rows.Scan(scanField...)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%s-%s", emailRecipientEvent.EmailRecipientID.String, emailRecipientEvent.Event.String)
		mapEventByEventAndEmailRecipientID[key] = emailRecipientEvent
	}

	return mapEventByEventAndEmailRecipientID, nil
}

func (repo *EmailRecipientEventRepo) queueInsert(b *pgx.Batch, emailRecipientEvent *model.EmailRecipientEvent) error {
	now := time.Now()
	err := multierr.Combine(
		emailRecipientEvent.CreatedAt.Set(now),
		emailRecipientEvent.UpdatedAt.Set(now),
		emailRecipientEvent.DeletedAt.Set(nil),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fieldNames := database.GetFieldNames(emailRecipientEvent)
	values := database.GetScanFields(emailRecipientEvent, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := emailRecipientEvent.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s AS ere (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__email_recipient_events 
		DO UPDATE SET 
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at
		WHERE ere.deleted_at IS NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *EmailRecipientEventRepo) BulkInsertEmailRecipientEventRepo(ctx context.Context, db database.QueryExecer, emailRecipientEvents model.EmailRecipientEvents) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkInsertEmailRecipientEventRepo")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range emailRecipientEvents {
		err := repo.queueInsert(b, item)
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
