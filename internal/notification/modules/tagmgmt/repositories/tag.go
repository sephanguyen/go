package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

var upsertTagQuery = `
INSERT INTO %s (%s) VALUES (%s)
ON CONFLICT ON CONSTRAINT pk__tags
DO UPDATE SET tag_name=EXCLUDED.tag_name, updated_at=EXCLUDED.updated_at, is_archived=EXCLUDED.is_archived
`

// TagRepo repo for tags table
type TagRepo struct{}

func (repo *TagRepo) Upsert(ctx context.Context, db database.QueryExecer, tag *entities.Tag) error {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.Upsert")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		tag.CreatedAt.Set(now),
		tag.UpdatedAt.Set(now),
		tag.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if tag.TagID.String == "" {
		_ = tag.TagID.Set(idutil.ULIDNow())
	}

	tableName := tag.TableName()
	fields := database.GetFieldNames(tag)
	pl := database.GeneratePlaceholders(len(fields))
	fieldValues := database.GetScanFields(tag, fields)

	query := fmt.Sprintf(upsertTagQuery, tableName, strings.Join(fields, ","), pl)

	cmd, err := db.Exec(ctx, query, fieldValues...)
	if err != nil {
		return fmt.Errorf("TagRepo.Upsert: %w", err)
	}

	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("TagRepo.Upsert: Tag is not inserted")
	}

	return nil
}

func (repo *TagRepo) DoesTagNameExist(ctx context.Context, db database.QueryExecer, name pgtype.Text) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.IsTagNameExist")
	defer span.End()
	query := `SELECT COUNT(*) FROM tags WHERE tag_name=$1 AND deleted_at IS NULL`
	row := db.QueryRow(ctx, query, name)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("TagRepo.IsNameExist: %w", err)
	}
	if count >= 1 {
		return true, nil
	}
	return false, nil
}

func (repo *TagRepo) SoftDelete(ctx context.Context, db database.QueryExecer, tagIds pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.SoftDelete")
	defer span.End()

	query := `
		UPDATE tags AS t
		SET deleted_at=now(),updated_at=now()
		WHERE t.tag_id = ANY($1) AND deleted_at IS NULL
	`
	_, err := db.Exec(ctx, query, tagIds)
	if err != nil {
		return err
	}

	return nil
}

func (repo *TagRepo) FindByID(ctx context.Context, db database.QueryExecer, tagID pgtype.Text) (*entities.Tag, error) {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindByID")
	defer span.End()

	tag := &entities.Tag{}
	fields := database.GetFieldNames(tag)
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE tag_id=$1 AND deleted_at IS NULL
	`, strings.Join(fields, ","), tag.TableName())

	err := database.Select(ctx, db, query, tagID).ScanOne(tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (repo *TagRepo) CheckTagIDsExist(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.CheckTagIDsExist")
	defer span.End()

	query := `
		SELECT CASE WHEN COUNT(*) = array_length($1::text[], 1)
			THEN TRUE
			ELSE FALSE
			END as result
		FROM tags
		WHERE tags.tag_id  = ANY($1::text[]) AND tags.deleted_at IS NULL AND tags.is_archived = FALSE
	`
	row := db.QueryRow(ctx, query, ids)
	var msg bool
	if err := row.Scan(&msg); err != nil {
		return false, fmt.Errorf("TagRepo.CheckTagIDsExist: %w", err)
	}
	switch msg {
	case true:
		return true, nil
	case false:
		return false, nil
	default:
		return false, fmt.Errorf("TagRepo.CheckTagIDsExist: query return unknown error")
	}
}

type FindTagFilter struct {
	Keyword    pgtype.Text
	Limit      pgtype.Int8
	Offset     pgtype.Int8
	WithCount  pgtype.Bool
	IsArchived pgtype.Bool
}

func NewFindTagFilter() FindTagFilter {
	f := FindTagFilter{}
	_ = f.Keyword.Set(nil)
	_ = f.Limit.Set(nil)
	_ = f.Offset.Set(nil)
	_ = f.WithCount.Set(false)
	_ = f.IsArchived.Set(false)
	return f
}

func (f *FindTagFilter) Validate() error {
	if f.Keyword.Status == pgtype.Null &&
		f.Limit.Status == pgtype.Null &&
		f.Offset.Status == pgtype.Null &&
		f.WithCount.Status == pgtype.Null &&
		f.IsArchived.Status == pgtype.Null {
		return fmt.Errorf("FindTagFilter all field is null")
	}
	return nil
}

func (repo *TagRepo) FindByFilter(ctx context.Context, db database.QueryExecer, filter FindTagFilter) (entities.Tags, uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindByFilter")
	defer span.End()

	if err := filter.Validate(); err != nil {
		return nil, 0, err
	}
	tags := entities.Tags{}
	e := &entities.Tag{}
	tableName := e.TableName()
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf(`
		SELECT %s
		FROM %s
	`, strings.Join(fields, ","), tableName)

	conditionStmt := fmt.Sprintf(`
		WHERE ($1::TEXT IS NULL OR tag_name LIKE CONCAT('%%',$1::TEXT,'%%'))
		AND deleted_at IS NULL
		AND ($2::BOOL IS NULL OR is_archived = $2)
	`)

	orderStmt := `
		ORDER BY updated_at DESC
	`

	pagingStmt := `
		LIMIT $3
		OFFSET $4
	`

	queryStmt := selectStmt + conditionStmt + orderStmt + pagingStmt

	rows, err := db.Query(ctx, queryStmt, filter.Keyword, filter.IsArchived, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	for rows.Next() {
		ent := &entities.Tag{}
		if err = rows.Scan(database.GetScanFields(ent, fields)...); err != nil {
			return nil, 0, err
		}
		tags = append(tags, ent)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	var total uint32
	if filter.WithCount.Bool {
		countStmt := fmt.Sprintf(`
			SELECT COUNT(*) as total
			FROM %s
		`, tableName)

		queryStmt = countStmt + conditionStmt
		err = db.QueryRow(ctx, queryStmt, filter.Keyword, filter.IsArchived).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	return tags, total, nil
}

func (repo *TagRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.Tag) error {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		repo.queueUpsert(b, item)
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

func (repo *TagRepo) queueUpsert(b *pgx.Batch, item *entities.Tag) {
	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(upsertTagQuery, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)
}

func (repo *TagRepo) FindDuplicateTagNames(ctx context.Context, db database.QueryExecer, tags []*entities.Tag) (map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindDuplicateTagNames")
	defer span.End()

	query := `
		SELECT tag_id, tag_name
		FROM tags
		WHERE tag_id NOT IN (
			SELECT tag_id
			FROM tags
			WHERE tag_name = ANY ($1::TEXT[]) AND tag_id = ANY($2::TEXT[]) AND deleted_at IS NULL
		) AND tag_name = ANY ($1::TEXT[]) AND deleted_at IS NULL;
	`
	tagNames := []string{}
	tagIDs := []string{}
	for _, tag := range tags {
		tagNames = append(tagNames, tag.TagName.String)
		tagIDs = append(tagIDs, tag.TagID.String)
	}
	rows, err := db.Query(ctx, query, database.TextArray(tagNames), database.TextArray(tagIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	duplicatedRecords := make(map[string]string)
	for rows.Next() {
		var (
			tagName string
			tagID   string
		)
		err = rows.Scan(&tagID, &tagName)
		if err != nil {
			return nil, err
		}
		duplicatedRecords[tagID] = tagName
	}
	return duplicatedRecords, nil
}

func (repo *TagRepo) FindTagIDsNotExist(ctx context.Context, db database.QueryExecer, tagIDs pgtype.TextArray) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindTagIDsNotExist")
	defer span.End()

	query := `
		WITH tmp1 AS (
			SELECT unnest($1::TEXT[]) AS tag_id
		)
		SELECT tag_id
		FROM tmp1
		WHERE tag_id NOT IN (
			SELECT tag_id
			FROM tags
			WHERE tag_id = ANY ($1::TEXT[]) AND deleted_at IS NULL
		);
	`
	rows, err := db.Query(ctx, query, tagIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notExistTagIDs := []string{}
	for rows.Next() {
		var tagID string
		err = rows.Scan(&tagID)
		if err != nil {
			return nil, err
		}
		notExistTagIDs = append(notExistTagIDs, tagID)
	}
	return notExistTagIDs, nil
}
