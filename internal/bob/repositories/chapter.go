package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ChapterRepo struct{}

func countChapter(ctx context.Context, db database.QueryExecer, query string, args []interface{}) (int, error) {
	e := new(entities.Chapter)
	count := 0
	row := db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(chapter_id) FROM %s %s", e.TableName(), query), args...)
	err := row.Scan(&count)
	if err != nil {
		return count, err
	}
	return count, nil
}

// FindWithFilter find chapter by chapteId and chapterName
func (r *ChapterRepo) FindWithFilter(ctx context.Context, db database.QueryExecer, chapterIDs []string, chapterName, subject string, grade int, limit, offset uint32) ([]*entities.Chapter, int, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindWithFilter")
	defer span.End()

	e := new(entities.Chapter)
	fields := database.GetFieldNames(e)

	query := ""
	var args []interface{}
	if grade > 0 {
		query += fmt.Sprintf(" AND grade = $%d", len(args)+1)
		args = append(args, grade)
	}
	if subject != "" && subject != pb.SUBJECT_NONE.String() {
		query += fmt.Sprintf(" AND subject = $%d", len(args)+1)
		args = append(args, subject)
	}

	if len(chapterIDs) > 0 {
		query += fmt.Sprintf(" AND chapter_id = ANY($%d)", len(args)+1)
		args = append(args, chapterIDs)
	}

	if chapterName != "" {
		query += fmt.Sprintf(" AND name = $%d", len(args)+1)
		args = append(args, chapterName)
	}

	if query != "" {
		query = fmt.Sprintf(" WHERE deleted_at IS NULL AND (%s)", query[4:])
	}

	count, err := countChapter(ctx, db, query, args)
	if err != nil {
		return nil, count, err
	}

	query += " ORDER BY display_order ASC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", len(args)+1)
		args = append(args, offset)
	}

	chapters := []*entities.Chapter{}
	rows, err := db.Query(ctx, fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(fields, ","), e.TableName(), query), args...)
	if err != nil {
		return nil, count, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	for rows.Next() {
		chapter := new(entities.Chapter)
		if err := rows.Scan(database.GetScanFields(chapter, fields)...); err != nil {
			return nil, count, errors.Wrap(err, "rows.Scan")
		}
		chapters = append(chapters, chapter)
	}
	if err := rows.Err(); err != nil {
		return nil, count, errors.Wrap(err, "rows.Err")
	}

	return chapters, count, nil
}

func (r *ChapterRepo) FindByBookID(ctx context.Context, db database.QueryExecer, bookID string) ([]*entities.Chapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindByBookID")
	defer span.End()

	e := new(entities.Chapter)
	fields := database.GetFieldNames(e)

	bc := new(entities.BookChapter)
	bcFields := database.GetFieldNames(bc)

	query := "SELECT c.%s, bc.%s FROM %s c JOIN %s bc ON bc.chapter_id = c.chapter_id WHERE c.deleted_at IS NULL AND bc.deleted_at IS NULL AND bc.book_id = $1 ORDER BY c.display_order ASC"

	rows, err := db.Query(ctx, fmt.Sprintf(query, strings.Join(fields, ", c."), strings.Join(bcFields, ", bc."), e.TableName(), bc.TableName()), &bookID)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	chapters := []*entities.Chapter{}
	for rows.Next() {
		chapter := new(entities.Chapter)
		args := database.GetScanFields(chapter, fields)

		bookChapter := new(entities.BookChapter)
		bcArgs := database.GetScanFields(bookChapter, bcFields)

		if err := rows.Scan(append(args, bcArgs...)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		chapters = append(chapters, chapter)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return chapters, nil
}

func (r *ChapterRepo) FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindByIDs")
	defer span.End()

	e := new(entities.Chapter)
	fields := database.GetFieldNames(e)

	pgChapterIDs := database.TextArray(chapterIDs)
	chapters := entities.Chapters{}

	err := database.Select(ctx, db, fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND chapter_id = ANY($1)", strings.Join(fields, ","), e.TableName()), &pgChapterIDs).ScanAll(&chapters)
	if err != nil {
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

func (r *ChapterRepo) Upsert(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities.Chapter) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT chapters_pk DO UPDATE
		SET name = $2, country = $3, subject = $4, grade = $5, display_order = $6, school_id = $7, updated_at = $8, deleted_at = $10`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		t.CreatedAt.Set(now)
		t.UpdatedAt.Set(now)

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("chapters not inserted")
		}
	}
	return nil
}

func (r *ChapterRepo) SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs []string) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.SoftDelete")
	defer span.End()
	e := &entities.Chapter{}
	query := fmt.Sprintf("UPDATE %s SET deleted_at = now() WHERE chapter_id = ANY($1::TEXT[]) AND deleted_at IS NULL", e.TableName())
	cmdTag, err := db.Exec(ctx, query, &chapterIDs)
	if err != nil {
		return 0, err
	}

	return int(cmdTag.RowsAffected()), nil
}

type EnSchoolID struct {
	SchoolID int32 `sql:"school_id"`
}

func (rcv *EnSchoolID) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"school_id"}
	values = []interface{}{&rcv.SchoolID}
	return
}

func (*EnSchoolID) TableName() string {
	return ""
}

type EnSchoolIDs []*EnSchoolID

func (u *EnSchoolIDs) Add() database.Entity {
	e := &EnSchoolID{}
	*u = append(*u, e)

	return e
}

func (r *ChapterRepo) FindSchoolIDsOnChapters(ctx context.Context, db database.QueryExecer, chapterIDs []string) ([]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.FindSchoolIDsOnCourses")
	defer span.End()

	query := "SELECT school_id FROM chapters WHERE deleted_at IS NULL AND chapter_id = ANY($1)"
	pgIDs := database.TextArray(chapterIDs)

	schoolIDs := EnSchoolIDs{}
	err := database.Select(ctx, db, query, &pgIDs).ScanAll(&schoolIDs)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	result := []int32{}
	for _, v := range schoolIDs {
		result = append(result, v.SchoolID)
	}

	return result, nil
}

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

func (r *ChapterRepo) DuplicateChapters(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) ([]*entities.CopiedChapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.DuplicateChapters")
	defer span.End()
	e := &entities.Chapter{}
	bookFieldNames := database.GetFieldNames(e)
	selectFields := golibs.Replace(bookFieldNames, []string{"chapter_id", "copied_from"}, []string{"uuid_generate_v4()", "chapter_id"})
	query := fmt.Sprintf(`
		INSERT INTO 
			%s (%s)
		SELECT
			%s
		FROM
			%s
		WHERE
			chapter_id = ANY($1)
		RETURNING
			chapter_id, copied_from
	`, e.TableName(), strings.Join(bookFieldNames, ", "), strings.Join(selectFields, ", "), e.TableName())

	copiedChapters := entities.CopiedChapters{}
	err := database.Select(ctx, db, query, &chapterIDs).ScanAll(&copiedChapters)
	if err != nil {
		return nil, err
	}
	return copiedChapters, nil
}

func (r *ChapterRepo) UpsertWithoutDisplayOrderWhenUpdate(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error {
	ctx, span := interceptors.StartSpan(ctx, "ChapterRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities.Chapter) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT chapters_pk DO UPDATE
		SET name = $2, country = $3, subject = $4, grade = $5, school_id = $7, updated_at = $8, deleted_at = $10`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		t.CreatedAt.Set(now)
		t.UpdatedAt.Set(now)

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("chapters not inserted")
		}
	}
	return nil
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

func (r *ChapterRepo) UpdateCurrentTopicDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedTopicDisplayOrder pgtype.Int4, chapterID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.UpdateCurrentChapter")
	defer span.End()

	updateCurrentTopicDisplayOrderStmtPlt := `UPDATE chapters SET current_topic_display_order = current_topic_display_order + $1::int, updated_at = now() WHERE chapter_id = $2::text`
	_, err := db.Exec(ctx, updateCurrentTopicDisplayOrderStmtPlt, totalGeneratedTopicDisplayOrder, chapterID)
	if err != nil {
		return err
	}

	return nil
}
