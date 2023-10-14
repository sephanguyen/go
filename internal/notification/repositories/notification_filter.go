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

type NotificationLocationFilterRepo struct{}

func (repo *NotificationLocationFilterRepo) queueUpsert(b *pgx.Batch, item *entities.NotificationLocationFilter) error {
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
		INSERT INTO %s as nlf (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk_notification_location_filter 
		DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			deleted_at = NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *NotificationLocationFilterRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.NotificationLocationFilters) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationLocationFilterRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		err := repo.queueUpsert(b, item)
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

func (repo *NotificationLocationFilterRepo) SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationLocationFilterRepo.SoftDeleteByNotificationID")
	defer span.End()

	query := `
		UPDATE notification_location_filter
		SET deleted_at = NOW()
		WHERE notification_id = $1
		AND deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, notificationID)

	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (repo *NotificationLocationFilterRepo) GetNotificationIDsByLocationIDs(ctx context.Context, db database.QueryExecer, notificationIDs, locationIDs pgtype.TextArray) ([]string, error) {
	query := `
		SELECT notification_id 
		FROM notification_location_filter nlf
		WHERE ($1::TEXT[] IS NULL OR nlf.notification_id = ANY($1::TEXT[]))
			AND nlf.location_id = ANY($2)
			AND nlf.deleted_at IS NULL
		`

	notificationIDsResult := make([]string, 0)
	rows, err := db.Query(ctx, query, notificationIDs, locationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		notificationID := &pgtype.Text{}
		f := []interface{}{notificationID}
		err := rows.Scan(f...)
		if err != nil {
			return nil, err
		}
		notificationIDsResult = append(notificationIDsResult, notificationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notificationIDsResult, nil
}

type NotificationCourseFilterRepo struct{}

func (repo *NotificationCourseFilterRepo) queueUpsert(b *pgx.Batch, item *entities.NotificationCourseFilter) error {
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
		INSERT INTO %s as ncf (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk_notification_course_filter 
		DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			deleted_at = NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *NotificationCourseFilterRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.NotificationCourseFilters) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationCourseFilterRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		err := repo.queueUpsert(b, item)
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

func (repo *NotificationCourseFilterRepo) SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationCourseFilterRepo.SoftDeleteByNotificationID")
	defer span.End()

	query := `
		UPDATE notification_course_filter
		SET deleted_at = NOW()
		WHERE notification_id = $1
		AND deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, notificationID)

	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (repo *NotificationCourseFilterRepo) GetNotificationIDsByCourseIDs(ctx context.Context, db database.QueryExecer, notificationIDs, courseIDs pgtype.TextArray) ([]string, error) {
	query := `
		SELECT notification_id 
		FROM notification_course_filter ncf
		WHERE ($1::TEXT[] IS NULL OR ncf.notification_id = ANY($1::TEXT[]))
			AND ncf.course_id = ANY($2)
			AND ncf.deleted_at IS NULL
		`

	notificationIDsResult := make([]string, 0)
	rows, err := db.Query(ctx, query, notificationIDs, courseIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		notificationID := &pgtype.Text{}
		f := []interface{}{notificationID}
		err := rows.Scan(f...)
		if err != nil {
			return nil, err
		}
		notificationIDsResult = append(notificationIDsResult, notificationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notificationIDsResult, nil
}

type NotificationClassFilterRepo struct{}

func (repo *NotificationClassFilterRepo) queueUpsert(b *pgx.Batch, item *entities.NotificationClassFilter) error {
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
		INSERT INTO %s as ncf (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk_notification_class_filter 
		DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			deleted_at = NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *NotificationClassFilterRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.NotificationClassFilters) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationClassFilterRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		err := repo.queueUpsert(b, item)
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

func (repo *NotificationClassFilterRepo) SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationClassFilterRepo.SoftDeleteByNotificationID")
	defer span.End()

	query := `
		UPDATE notification_class_filter
		SET deleted_at = NOW()
		WHERE notification_id = $1
		AND deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, notificationID)

	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (repo *NotificationClassFilterRepo) GetNotificationIDsByClassIDs(ctx context.Context, db database.QueryExecer, notificationIDs, classIDs pgtype.TextArray) ([]string, error) {
	query := `
		SELECT notification_id 
		FROM notification_class_filter ncf
		WHERE ($1::TEXT[] IS NULL OR ncf.notification_id = ANY($1::TEXT[]))
			AND ncf.class_id = ANY($2)
			AND ncf.deleted_at IS NULL
		`

	notificationIDsResult := make([]string, 0)
	rows, err := db.Query(ctx, query, notificationIDs, classIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		notificationID := &pgtype.Text{}
		f := []interface{}{notificationID}
		err := rows.Scan(f...)
		if err != nil {
			return nil, err
		}
		notificationIDsResult = append(notificationIDsResult, notificationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notificationIDsResult, nil
}
