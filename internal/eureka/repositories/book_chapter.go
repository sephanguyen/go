package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type BookChapterRepo struct{}

func (r *BookChapterRepo) FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string][]*entities.BookChapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.FindByBookIDs")
	defer span.End()

	query := "SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = ANY($1) ORDER BY created_at DESC"
	b := &entities.BookChapter{}
	fields, _ := b.FieldMap()

	books := entities.BookChapters{}

	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &bookIDs).ScanAll(&books)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	mapBooks := map[string][]*entities.BookChapter{}
	for _, book := range books {
		mapBooks[book.BookID.String] = append(mapBooks[book.BookID.String], book)
	}

	return mapBooks, nil
}

const bulkUpsertBookChapterStmTpl = `
INSERT INTO %s (%s) 
VALUES %s 
ON CONFLICT ON CONSTRAINT books_chapters_pk DO UPDATE 
SET 
	updated_at = excluded.updated_at, 
	deleted_at = excluded.deleted_at`

func (r *BookChapterRepo) Upsert(ctx context.Context, db database.Ext, bookChapters []*entities.BookChapter) error {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.Upsert")
	defer span.End()

	now := time.Now()
	for _, bookChapter := range bookChapters {
		if err := multierr.Combine(
			bookChapter.CreatedAt.Set(now),
			bookChapter.UpdatedAt.Set(now),
		); err != nil {
			return err
		}
	}
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertBookChapterStmTpl, bookChapters)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertBookChapter error: %s", err.Error())
	}
	return nil
}

func (r *BookChapterRepo) SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs, bookIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.SoftDelete")
	defer span.End()

	query := "UPDATE books_chapters SET deleted_at = now() WHERE book_id = ANY($1) AND chapter_id = ANY($2) AND deleted_at IS NULL"
	if _, err := db.Exec(ctx, query, &bookIDs, &chapterIDs); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *BookChapterRepo) RetrieveContentStructuresByLOs(
	ctx context.Context,
	db database.QueryExecer,
	loIDs pgtype.TextArray,
) (map[string]entities.ContentStructure, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.RetrieveContentStructuresByLOs")
	defer span.End()

	query := `
		SELECT bc.book_id, sub.chapter_id, sub.topic_id, sub.lo_id
		FROM books_chapters bc
		INNER JOIN (
			SELECT tlo.lo_id , t.topic_id , c.chapter_id
			FROM topics_learning_objectives tlo
			INNER JOIN topics t ON tlo.topic_id = t.topic_id
			INNER JOIN chapters c ON c.chapter_id = t.chapter_id
			WHERE tlo.lo_id = ANY($1::_TEXT)
		) AS sub ON sub.chapter_id = bc.chapter_id
	`
	rows, err := db.Query(ctx, query, &loIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %v", err)
	}
	defer rows.Close()

	ret := make(map[string]entities.ContentStructure)
	for rows.Next() {
		cs := entities.ContentStructure{}
		var loID string
		if err := rows.Scan(&cs.BookID, &cs.ChapterID, &cs.TopicID, &loID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		ret[loID] = cs
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %v", err)
	}

	return ret, nil
}

func (r *BookChapterRepo) RetrieveContentStructuresByTopics(
	ctx context.Context,
	db database.QueryExecer,
	topicIDs pgtype.TextArray,
) (map[string][]entities.ContentStructure, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.RetrieveContentStructuresByTopics")
	defer span.End()

	query := `
		SELECT bc.book_id, sub.chapter_id, sub.topic_id
		FROM books_chapters bc
		INNER JOIN (
			SELECT t.topic_id , c.chapter_id
			FROM topics t
			INNER JOIN chapters c ON c.chapter_id = t.chapter_id
			WHERE t.topic_id = ANY($1::TEXT[])
		) AS sub ON sub.chapter_id = bc.chapter_id
	`
	rows, err := db.Query(ctx, query, &topicIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %v", err)
	}
	defer rows.Close()

	ret := make(map[string][]entities.ContentStructure)
	for rows.Next() {
		cs := entities.ContentStructure{}
		if err := rows.Scan(&cs.BookID, &cs.ChapterID, &cs.TopicID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		ret[cs.TopicID] = append(ret[cs.TopicID], cs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %v", err)
	}

	return ret, nil
}

func (r *BookChapterRepo) SoftDeleteByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.SoftDeleteByChapterIDs")
	defer span.End()
	e := &entities.BookChapter{}
	query := fmt.Sprintf(`
		UPDATE %s SET deleted_at = now(), updated_at = now()
		WHERE chapter_id = ANY($1) AND deleted_at IS NULL
	`, e.TableName())
	_, err = db.Exec(ctx, query, database.TextArray(chapterIDs))
	return
}

func (r *BookChapterRepo) FindByBookIDsV2(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.BookChapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.FindByBookIDsV2")
	defer span.End()
	b := &entities.BookChapter{}
	fields, _ := b.FieldMap()
	bookChapters := entities.BookChapters{}

	query := fmt.Sprintf(`
		SELECT bc.%s FROM %s bc INNER JOIN chapters c on c.chapter_id = bc.chapter_id 
		WHERE bc.deleted_at IS NULL AND ($1::TEXT[] IS NULL OR bc.book_id = ANY($1::TEXT[])) 
		ORDER BY c.display_order ASC`,
		strings.Join(fields, ", bc."), b.TableName())

	err := database.Select(ctx, db, query, &bookIDs).ScanAll(&bookChapters)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return bookChapters, nil
}
