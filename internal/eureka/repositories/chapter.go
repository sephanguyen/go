package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type ChapterRepo struct{}

type ListChaptersArgs struct {
	ChapterIDs pgtype.TextArray
	Limit      uint32

	// used for pagination
	Offset    pgtype.Int4
	ChapterID pgtype.Text
}

func (r *ChapterRepo) ListChapters(ctx context.Context, db database.QueryExecer, args *ListChaptersArgs) ([]*entities.Chapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindChapters")
	defer span.End()
	b := &entities.Chapter{}
	fields, _ := b.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE
			deleted_at IS NULL
			AND ($1::text[] IS NULL OR chapter_id = ANY($1))
			AND (($2::int IS NULL AND $3::text IS NULL) OR ((display_order, chapter_id) > ($2, $3)))
		ORDER BY
			display_order ASC, chapter_id ASC
		LIMIT $4`, strings.Join(fields, ", "), b.TableName())

	chapters := entities.Chapters{}
	if err := database.Select(ctx, db, query, &args.ChapterIDs, &args.Offset, &args.ChapterID, &args.Limit).ScanAll(&chapters); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return chapters, nil
}

func (r *ChapterRepo) FindByID(ctx context.Context, db database.QueryExecer, chapterID pgtype.Text, enhancers ...QueryEnhancer) (*entities.Chapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "Chapter.FindByID")
	defer span.End()

	findByChapterIDStmtPlt := `SELECT %s FROM chapters WHERE chapter_id = $1 AND deleted_at IS NULL`
	c := &entities.Chapter{}
	fields, values := c.FieldMap()

	for _, e := range enhancers {
		e(&findByChapterIDStmtPlt)
	}

	err := db.QueryRow(ctx, fmt.Sprintf(findByChapterIDStmtPlt, strings.Join(fields, ", ")), &chapterID).Scan(values...)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *ChapterRepo) FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindByIDs")
	defer span.End()

	e := new(entities.Chapter)
	fields := database.GetFieldNames(e)

	pgChapterIDs := database.TextArray(chapterIDs)
	chapters := entities.Chapters{}

	if err := database.Select(ctx, db, fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND chapter_id = ANY($1)", strings.Join(fields, ","), e.TableName()), &pgChapterIDs).ScanAll(&chapters); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	mChapters := make(map[string]*entities.Chapter)
	if len(chapters) > 0 {
		for _, v := range chapters {
			mChapters[v.ID.String] = v
		}
	}

	return mChapters, nil
}

func (r *ChapterRepo) UpdateCurrentTopicDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedTopicDisplayOrder pgtype.Int4, chapterID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.UpdateCurrentChapter")
	defer span.End()

	query := `
		UPDATE chapters SET
			current_topic_display_order = current_topic_display_order + $1::int,
			updated_at = now()
		WHERE
			chapter_id = $2::text
		`
	_, err := db.Exec(ctx, query, totalGeneratedTopicDisplayOrder, chapterID)
	if err != nil {
		return err
	}
	return nil
}

const bulkUpsertChapterStmTpl = `
INSERT INTO %s (%s) 
VALUES %s 
ON CONFLICT ON CONSTRAINT chapters_pk DO UPDATE 
SET
	name = excluded.name,
	country = excluded.country,
	subject = excluded.subject,
	grade = excluded.grade,
	display_order = excluded.display_order,
	school_id = excluded.school_id,
	updated_at = excluded.updated_at,
	deleted_at = excluded.deleted_at
`

func (r *ChapterRepo) Upsert(ctx context.Context, db database.QueryExecer, chapters []*entities.Chapter) error {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.Upsert")
	defer span.End()
	now := time.Now()

	for _, chapter := range chapters {
		if err := multierr.Combine(
			chapter.CreatedAt.Set(now),
			chapter.UpdatedAt.Set(now),
		); err != nil {
			return fmt.Errorf("failed to set timestamp: %w", err)
		}
	}
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertChapterStmTpl, chapters)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertChapter error: %s", err.Error())
	}
	return nil
}

const bulkUpsertChapterWithoutDisPlayOrderStmTpl = `
INSERT INTO %s (%s) VALUES %s
ON CONFLICT ON CONSTRAINT chapters_pk DO UPDATE 
SET
	name = excluded.name,
	country = excluded.country,
	subject = excluded.subject,
	grade = excluded.grade,
	school_id = excluded.school_id,
	updated_at = excluded.updated_at,
	deleted_at = excluded.deleted_at
`

func (r *ChapterRepo) UpsertWithoutDisplayOrderWhenUpdate(ctx context.Context, db database.QueryExecer, chapters []*entities.Chapter) error {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.Upsert")
	defer span.End()
	now := time.Now()

	for _, chapter := range chapters {
		if err := multierr.Combine(
			chapter.CreatedAt.Set(now),
			chapter.UpdatedAt.Set(now),
		); err != nil {
			return err
		}
	}
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertChapterWithoutDisPlayOrderStmTpl, chapters)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertChapter error: %s", err.Error())
	}
	return nil
}

func (r *ChapterRepo) SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs []string) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.SoftDelete")
	defer span.End()
	e := &entities.Chapter{}
	query := fmt.Sprintf(`
		UPDATE %s SET
			deleted_at = now()
		WHERE
			chapter_id = ANY($1::TEXT[]) AND
			deleted_at IS NULL
		`,
		e.TableName(),
	)
	cmdTag, err := db.Exec(ctx, query, database.TextArray(chapterIDs))
	if err != nil {
		return 0, err
	}

	return int(cmdTag.RowsAffected()), nil
}

func (r *ChapterRepo) DuplicateChapters(ctx context.Context, db database.QueryExecer, bookID string, chapterIDs []string) ([]*entities.CopiedChapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.DuplicateChapters")
	defer span.End()
	e := &entities.Chapter{}
	chapterFieldNames := database.GetFieldNames(e)
	selectFields := golibs.Replace(chapterFieldNames, []string{"chapter_id", "copied_from", "book_id"}, []string{"uuid_generate_v4()", "chapter_id", "$1"})
	query := fmt.Sprintf(`
		INSERT INTO 
			%s (%s)
		SELECT
			%s
		FROM
			%s
		WHERE
			chapter_id = ANY($2)
		RETURNING
			chapter_id, copied_from
	`, e.TableName(), strings.Join(chapterFieldNames, ", "), strings.Join(selectFields, ", "), e.TableName())

	copiedChapters := entities.CopiedChapters{}
	err := database.Select(ctx, db, query, &bookID, &chapterIDs).ScanAll(&copiedChapters)
	if err != nil {
		return nil, err
	}
	return copiedChapters, nil
}

func (r *ChapterRepo) FindByBookID(ctx context.Context, db database.QueryExecer, bookID string) (map[string]*entities.Chapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindByBookID")
	defer span.End()

	e := new(entities.Chapter)
	fields := database.GetFieldNames(e)

	chapters := entities.Chapters{}
	if err := database.Select(ctx, db, fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = $1::TEXT", strings.Join(fields, ","), e.TableName()), &bookID).ScanAll(&chapters); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	mChapters := make(map[string]*entities.Chapter)
	for _, v := range chapters {
		mChapters[v.ID.String] = v
	}

	return mChapters, nil
}

func (r *ChapterRepo) FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.Chapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindByBookIDs")
	defer span.End()
	b := &entities.Chapter{}
	fields, _ := b.FieldMap()
	chapters := entities.Chapters{}

	query := fmt.Sprintf(`
		SELECT %s FROM %s 
		WHERE ($1::TEXT[] IS NULL OR book_id = ANY($1::TEXT[])) 
		ORDER BY display_order ASC`,
		strings.Join(fields, ", "), b.TableName())

	err := database.Select(ctx, db, query, &bookIDs).ScanAll(&chapters)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return chapters, nil
}
