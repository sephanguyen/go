package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// InfoNotificationTagRepo repo for info_notifications_tags table
type InfoNotificationTagRepo struct{}

func (repo *InfoNotificationTagRepo) GetByNotificationIDs(ctx context.Context, db database.QueryExecer, notificationIDs pgtype.TextArray) (map[string]entities.InfoNotificationsTags, error) {
	e := &entities.InfoNotificationTag{}
	fields := "ifnt." + strings.Join(database.GetFieldNames(e), ", ifnt.")

	query := fmt.Sprintf(`
		SELECT %s 
		FROM info_notifications_tags ifnt 
			INNER JOIN info_notifications ifn ON ifnt.notification_id = ifn.notification_id 
			INNER JOIN tags t ON ifnt.tag_id = t.tag_id 
		WHERE ifn.notification_id = ANY($1) 
			AND ifnt.deleted_at IS NULL
			AND ifn.deleted_at IS NULL
			AND t.deleted_at IS NULL
			AND t.is_archived = FALSE
	`, fields)

	res := make(map[string]entities.InfoNotificationsTags)
	rows, err := db.Query(ctx, query, notificationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.InfoNotificationTag{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		res[e.NotificationID.String] = append(res[e.NotificationID.String], e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *InfoNotificationTagRepo) GetNotificationIDsByTagIDs(ctx context.Context, db database.QueryExecer, tagIDs pgtype.TextArray) ([]string, error) {
	query := `
		SELECT notification_id 
		FROM info_notifications_tags ifnt
		JOIN tags t ON t.tag_id = ifnt.tag_id
		WHERE ifnt.tag_id = ANY($1)
			AND ifnt.deleted_at IS NULL
			AND t.is_archived = FALSE;
		`

	res := make([]string, 0)
	rows, err := db.Query(ctx, query, tagIDs)
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
		res = append(res, notificationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *InfoNotificationTagRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, ents []*entities.InfoNotificationTag) error {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationTagRepo.BulkInsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, ifnt := range ents {
		repo.queueUpsert(b, ifnt)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		cmd, err := result.Exec()
		if err != nil || cmd.RowsAffected() != 1 {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (repo *InfoNotificationTagRepo) queueUpsert(b *pgx.Batch, item *entities.InfoNotificationTag) {
	fields := database.GetFieldNames(item)
	pl := database.GeneratePlaceholders(len(fields))
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__notifications_tags
		DO UPDATE SET created_at = $4, updated_at = $5, deleted_at = NULL`,
		item.TableName(), strings.Join(fields, ","), pl)
	b.Queue(query, database.GetScanFields(item, fields)...)
}

type SoftDeleteNotificationTagFilter struct {
	NotificationTagIDs pgtype.TextArray
	NotificationIDs    pgtype.TextArray
	TagIDs             pgtype.TextArray
}

func NewSoftDeleteNotificationTagFilter() *SoftDeleteNotificationTagFilter {
	f := &SoftDeleteNotificationTagFilter{}
	_ = f.NotificationTagIDs.Set(nil)
	_ = f.NotificationIDs.Set(nil)
	_ = f.TagIDs.Set(nil)
	return f
}

func (repo *InfoNotificationTagRepo) SoftDelete(ctx context.Context, db database.QueryExecer, filter *SoftDeleteNotificationTagFilter) error {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationTagRepo.SoftDeleteByTagID")
	defer span.End()

	if filter.NotificationIDs.Status != pgtype.Present && filter.TagIDs.Status != pgtype.Present && filter.NotificationTagIDs.Status != pgtype.Present {
		return fmt.Errorf("cannot soft delete with nil filter")
	}

	query := `
		UPDATE info_notifications_tags AS infnt
		SET deleted_at=now(),updated_at=now()
		WHERE ($1::TEXT[] IS NULL OR infnt.notification_tag_id = ANY($1)) 
			AND ($2::TEXT[] IS NULL OR infnt.notification_id = ANY($2)) 
			AND ($3::TEXT[] IS NULL OR infnt.tag_id = ANY($2)) 
			AND deleted_at IS NULL
	`
	_, err := db.Exec(ctx, query, filter.NotificationTagIDs, filter.NotificationIDs, filter.TagIDs)
	if err != nil {
		return err
	}

	return nil
}
