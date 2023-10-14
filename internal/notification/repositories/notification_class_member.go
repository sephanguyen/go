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
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type NotificationClassMemberRepo struct{}

func (n *NotificationClassMemberRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.NotificationClassMember) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationClassMemberRepo.Upsert")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		e.UpdatedAt.Set(now),
	)

	if e.CreatedAt.Status != pgtype.Present || e.CreatedAt.Time.IsZero() {
		err = multierr.Append(err, e.CreatedAt.Set(now))
	}
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fieldNames := database.GetFieldNames(e)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s as ncm (%s) 
		VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__notification_class_members
		DO UPDATE SET 
			start_at = $3, 
			end_at = $4, 
			updated_at = $6, 
			deleted_at = $9
	`, e.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := database.GetScanFields(e, fieldNames)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return errors.Wrap(err, "r.DB.ExecEx")
	}
	return nil
}

type NotificationClassMemberFilter struct {
	StudentIDs  pgtype.TextArray
	CourseIDs   pgtype.TextArray
	LocationIDs pgtype.TextArray
}

func NewNotificationClassMemberFilter() *NotificationClassMemberFilter {
	f := &NotificationClassMemberFilter{}
	_ = f.StudentIDs.Set(nil)
	_ = f.CourseIDs.Set(nil)
	_ = f.LocationIDs.Set(nil)
	return f
}

func (n *NotificationClassMemberRepo) SoftDeleteByFilter(ctx context.Context, db database.QueryExecer, filter *NotificationClassMemberFilter) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationClassMemberRepo.SoftDeleteByFilter")
	defer span.End()

	query := `
		UPDATE notification_class_members
		SET deleted_at = NOW()
		WHERE ($1::TEXT[] IS NULL OR student_id = ANY($1))
		AND ($2::TEXT[] IS NULL OR course_id = ANY($2))
		AND ($3::TEXT[] IS NULL OR location_id = ANY($3))
		AND deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, filter.StudentIDs, filter.CourseIDs, filter.LocationIDs)

	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (n *NotificationClassMemberRepo) queueUpsert(b *pgx.Batch, item *entities.NotificationClassMember) {
	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	now := time.Now()
	_ = item.UpdatedAt.Set(now)

	if item.CreatedAt.Status != pgtype.Present || item.CreatedAt.Time.IsZero() {
		_ = item.CreatedAt.Set(now)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__notification_class_members
		DO UPDATE SET start_at = $3, end_at = $4, updated_at = $6, deleted_at = $9 
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)
}

func (n *NotificationClassMemberRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.NotificationClassMember) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationClassMemberRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		n.queueUpsert(b, item)
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
