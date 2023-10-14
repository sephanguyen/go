package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type InfoNotificationAccessPathRepo struct {
}

func (r *InfoNotificationAccessPathRepo) Upsert(ctx context.Context, db database.QueryExecer, notiLocation *entities.InfoNotificationAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationAccessPathRepo.Upsert")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		notiLocation.CreatedAt.Set(now),
		notiLocation.UpdatedAt.Set(now),
		notiLocation.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fields := database.GetFieldNames(notiLocation)
	values := database.GetScanFields(notiLocation, fields)
	pl := database.GeneratePlaceholders(len(fields))
	tableName := notiLocation.TableName()

	query := fmt.Sprintf(`INSERT INTO %s as inap (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__info_notifications_access_paths 
		DO UPDATE SET 
			notification_id = EXCLUDED.notification_id,
			location_id = EXCLUDED.location_id,
			access_path = EXCLUDED.access_path,
			updated_at = EXCLUDED.updated_at,
			deleted_at = NULL;
		`, tableName, strings.Join(fields, ","), pl)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not upsert notification access path")
	}

	return nil
}

func (r *InfoNotificationAccessPathRepo) queueUpsert(b *pgx.Batch, item *entities.InfoNotificationAccessPath) error {
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
		INSERT INTO %s as inap (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__info_notifications_access_paths
		DO UPDATE SET
			notification_id = EXCLUDED.notification_id,
			location_id = EXCLUDED.location_id,
			access_path = EXCLUDED.access_path,
			updated_at = EXCLUDED.updated_at,
			deleted_at = NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (r *InfoNotificationAccessPathRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.InfoNotificationAccessPaths) error {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationAccessPathRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		err := r.queueUpsert(b, item)
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

func (r *InfoNotificationAccessPathRepo) GetByNotificationIDAndNotInLocationIDs(ctx context.Context, db database.QueryExecer, notificationID string, locationIDs []string) (entities.InfoNotificationAccessPaths, error) {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationAccessPathRepo.GetByNotificationIDAndNotInLocationIDs")
	defer span.End()

	fields := strings.Join(database.GetFieldNames(&entities.InfoNotificationAccessPath{}), ",")
	ents := entities.InfoNotificationAccessPaths{}
	err := database.Select(ctx, db, fmt.Sprintf(`
		SELECT %s
		FROM info_notifications_access_paths inap
		WHERE inap.notification_id = $1
			AND NOT (inap.location_id = ANY($2))
			AND deleted_at IS NULL
	`, fields), database.Text(notificationID), database.TextArray(locationIDs)).ScanAll(&ents)
	if err != nil {
		return nil, fmt.Errorf("database.Select %w", err)
	}

	return ents, nil
}

type SoftDeleteNotificationAccessPathFilter struct {
	NotificationIDs pgtype.TextArray
	LocationIDs     pgtype.TextArray
}

func NewSoftDeleteNotificationAccessPathFilter() *SoftDeleteNotificationAccessPathFilter {
	f := &SoftDeleteNotificationAccessPathFilter{}
	_ = f.NotificationIDs.Set(nil)
	_ = f.LocationIDs.Set(nil)
	return f
}

func (r *InfoNotificationAccessPathRepo) SoftDelete(ctx context.Context, db database.QueryExecer, filter *SoftDeleteNotificationAccessPathFilter) error {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationAccessPathRepo.SoftDelete")
	defer span.End()

	if filter.NotificationIDs.Status != pgtype.Present {
		return fmt.Errorf("cannot delete notification access path without notification_id")
	}

	query := `
		UPDATE info_notifications_access_paths
		SET deleted_at = NOW()
		WHERE ($1::TEXT[] IS NULL OR notification_id = ANY($1)) 
			AND ($2::TEXT[] IS NULL OR location_id = ANY($2))
			AND deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, filter.NotificationIDs, filter.LocationIDs)

	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (r *InfoNotificationAccessPathRepo) GetLocationIDsByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationAccessPathRepo.GetLocationIDsByNotificationID")
	defer span.End()

	query := `
		SELECT location_id
		FROM info_notifications_access_paths
		WHERE notification_id = $1 AND deleted_at IS NULL;
	`
	rows, err := db.Query(ctx, query, notificationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0)
	for rows.Next() {
		var locationID string
		err = rows.Scan(&locationID)
		if err != nil {
			return nil, err
		}
		result = append(result, locationID)
	}

	return result, nil
}
