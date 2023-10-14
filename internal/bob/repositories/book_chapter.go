package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type BookChapterRepo struct{}

func (rcv *BookChapterRepo) SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs, bookIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.SoftDelete")
	defer span.End()

	query := "UPDATE books_chapters SET deleted_at = now(), updated_at = now() WHERE book_id = ANY($1) AND chapter_id = ANY($2) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &bookIDs, &chapterIDs)
	if err != nil {
		return err
	}

	return nil
}

func (rcv *BookChapterRepo) SoftDeleteByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.SoftDeleteByChapterIDs")
	defer span.End()
	e := &entities_bob.BookChapter{}
	query := fmt.Sprintf(`
	UPDATE %s 
	SET deleted_at = now(), updated_at = now() 
	WHERE chapter_id = ANY($1) AND deleted_at IS NULL`, e.TableName())
	_, err := db.Exec(ctx, query, &chapterIDs)
	if err != nil {
		return err
	}

	return nil
}

func (r *BookChapterRepo) Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.BookChapter) error {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities_bob.BookChapter) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT books_chapters_pk DO UPDATE SET updated_at = $3, deleted_at = $5", t.TableName(), strings.Join(fieldNames, ","), placeHolders)
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
			return fmt.Errorf("book chapter not inserted")
		}
	}
	return nil
}

func (r *BookChapterRepo) FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string][]*entities_bob.BookChapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.FindByBookIDs")
	defer span.End()

	query := "SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = ANY($1) ORDER BY created_at DESC"
	b := &entities_bob.BookChapter{}
	fields, _ := b.FieldMap()

	books := entities_bob.BookChapters{}

	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &bookIDs).ScanAll(&books)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	mapBooks := map[string][]*entities_bob.BookChapter{}
	for _, book := range books {
		mapBooks[book.BookID.String] = append(mapBooks[book.BookID.String], book)
	}

	return mapBooks, nil
}

type ContentStructure entities.ContentStructure

func (r *BookChapterRepo) RetrieveContentStructuresByLOs(
	ctx context.Context,
	db database.QueryExecer,
	loIDs pgtype.TextArray,
) (map[string][]ContentStructure, error) {
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
			WHERE tlo.lo_id = ANY($1)
		) AS sub ON sub.chapter_id = bc.chapter_id
	`
	rows, err := db.Query(ctx, query, &loIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %v", err)
	}
	defer rows.Close()

	ret := make(map[string][]ContentStructure)
	for rows.Next() {
		cs := ContentStructure{}
		var loID string
		if err := rows.Scan(&cs.BookID, &cs.ChapterID, &cs.TopicID, &loID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		ret[loID] = append(ret[loID], cs)
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
) (map[string][]ContentStructure, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.RetrieveContentStructuresByTopics")
	defer span.End()

	query := `
		SELECT bc.book_id, sub.chapter_id, sub.topic_id
		FROM books_chapters bc
		INNER JOIN (
			SELECT t.topic_id , c.chapter_id
			FROM topics t
			INNER JOIN chapters c ON c.chapter_id = t.chapter_id
			WHERE t.topic_id = ANY($1)
		) AS sub ON sub.chapter_id = bc.chapter_id
	`
	rows, err := db.Query(ctx, query, &topicIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %v", err)
	}
	defer rows.Close()

	ret := make(map[string][]ContentStructure)
	for rows.Next() {
		cs := ContentStructure{}
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

func (r *BookChapterRepo) FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) ([]*entities_bob.BookChapter, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookChapterRepo.FindByIDs")
	defer span.End()

	e := new(entities_bob.BookChapter)
	fields := database.GetFieldNames(e)

	bookChapters := entities_bob.BookChapters{}

	err := database.Select(ctx, db, fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND chapter_id = ANY($1)", strings.Join(fields, ","), e.TableName()), &chapterIDs).ScanAll(&bookChapters)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return bookChapters, nil
}
