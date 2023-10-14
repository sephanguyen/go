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

type BookRepo struct{}

func (r *BookRepo) FindByID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, enhancers ...QueryEnhancer) (*entities.Book, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.FindByID")
	defer span.End()

	findByBookIDStmtPlt := `SELECT %s FROM books WHERE book_id = $1 AND deleted_at IS NULL`
	b := &entities.Book{}
	fields, values := b.FieldMap()

	for _, e := range enhancers {
		e(&findByBookIDStmtPlt)
	}

	err := db.QueryRow(ctx, fmt.Sprintf(findByBookIDStmtPlt, strings.Join(fields, ", ")), &bookID).Scan(values...)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *BookRepo) FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.FindByIDs")
	defer span.End()

	query := "SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = ANY($1) ORDER BY created_at DESC"
	b := &entities.Book{}
	fields, _ := b.FieldMap()

	books := entities.Books{}

	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &bookIDs).ScanAll(&books)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	mapBooks := map[string]*entities.Book{}
	for _, book := range books {
		mapBooks[book.ID.String] = book
	}

	return mapBooks, nil
}

func countCourseBooks(ctx context.Context, db database.QueryExecer, query string, args []interface{}) (int, error) {
	var count int
	if err := db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(b.book_id) %s", query), args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BookRepo) FindWithFilter(ctx context.Context, db database.QueryExecer, courseID string, limit, offset uint32) ([]*entities.Book, int, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.FindWithFilter")
	defer span.End()

	cb := new(entities.CoursesBooks)
	e := new(entities.Book)
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("FROM %s AS b LEFT JOIN %s cb ON cb.book_id = b.book_id WHERE ($1::text IS NULL OR cb.course_id = $1) ", e.TableName(), cb.TableName())

	args := []interface{}{&courseID}
	count, err := countCourseBooks(ctx, db, query, args)
	if err != nil {
		return nil, count, err
	}

	query += "ORDER BY b.created_at, b.name ASC "

	if limit > 0 {
		query += fmt.Sprintf("LIMIT $%d ", len(args)+1)
		args = append(args, limit)
	}
	if offset > 0 {
		query += fmt.Sprintf("OFFSET $%d ", len(args)+1)
		args = append(args, offset)
	}

	query = fmt.Sprintf("SELECT b.%s %s", strings.Join(fields, ", b."), query)

	books := entities.Books{}
	err = database.Select(ctx, db, query, args...).ScanAll(&books)
	if err != nil {
		return nil, 0, fmt.Errorf("database.Select: %w", err)
	}

	return books, count, nil
}

const bulkUpsertBookStmTpl = `
INSERT INTO %s (%s) 
VALUES %s 
ON CONFLICT ON CONSTRAINT books_pk DO UPDATE 
SET 
	name = excluded.name,
	updated_at = excluded.updated_at`

func (r *BookRepo) Upsert(ctx context.Context, db database.Ext, books []*entities.Book) error {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.Upsert")
	defer span.End()

	now := time.Now()
	for _, book := range books {
		err := multierr.Combine(
			book.CreatedAt.Set(now),
			book.UpdatedAt.Set(now),
		)
		if err != nil {
			return err
		}
	}
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertBookStmTpl, books)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertBook error: %s", err.Error())
	}
	return nil
}

func (r *BookRepo) SoftDelete(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.SoftDelete")
	defer span.End()

	query := "UPDATE books SET deleted_at = now() WHERE book_id = ANY($1) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &bookIDs)
	if err != nil {
		return err
	}

	return nil
}

type BookTreeInfo struct {
	LoID      pgtype.Text
	TopicID   pgtype.Text
	ChapterID pgtype.Text

	LoDisplayOrder      pgtype.Int2
	TopicDisplayOrder   pgtype.Int2
	ChapterDisplayOrder pgtype.Int2
}

func (r *BookRepo) RetrieveBookTreeByBookID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text) ([]*BookTreeInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.RetrieveBookTreeByBookID")
	defer span.End()
	ce := &entities.Chapter{}
	te := &entities.Topic{}
	lme := &entities.LearningMaterial{}
	query := fmt.Sprintf(`
	SELECT lm.learning_material_id, tp.topic_id, ct.chapter_id
	FROM %s as ct
	JOIN %s AS tp
	USING (chapter_id) 
	LEFT OUTER JOIN %s AS lm
	USING (topic_id) 
	WHERE book_id = $1 AND ct.deleted_at IS NULL AND tp.deleted_at IS NULL AND lm.deleted_at IS NULL 
	`,
		ce.TableName(),
		te.TableName(),
		lme.TableName(),
	)
	var result []*BookTreeInfo
	rows, err := db.Query(ctx, query, &bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e BookTreeInfo
		if err := rows.Scan(&e.LoID, &e.TopicID, &e.ChapterID); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, nil
}

func (r *BookRepo) RetrieveBookTreeByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*BookTreeInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.RetrieveBookTreeByTopicIDs")
	defer span.End()
	ce := &entities.Chapter{}
	te := &entities.Topic{}
	loe := &entities.LearningObjective{}
	query := fmt.Sprintf(`
	SELECT lo.lo_id, tp.topic_id, ct.chapter_id, lo.display_order as lo_display_order, tp.display_order as topic_display_order, ct.display_order as chapter_display_order
	FROM %s as tp
	JOIN %s as ct
	USING (chapter_id) 
	LEFT OUTER JOIN (
		SELECT * FROM %s
		WHERE deleted_at IS NULL
	) AS lo
	USING (topic_id) 
	WHERE topic_id = ANY($1) AND ct.deleted_at IS NULL AND tp.deleted_at IS NULL
	`,
		te.TableName(),
		ce.TableName(),
		loe.TableName(),
	)
	var result []*BookTreeInfo
	rows, err := db.Query(ctx, query, &topicIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e BookTreeInfo
		if err := rows.Scan(&e.LoID, &e.TopicID, &e.ChapterID, &e.LoDisplayOrder, &e.TopicDisplayOrder, &e.ChapterDisplayOrder); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, nil
}

type ListBooksArgs struct {
	BookIDs pgtype.TextArray
	Limit   uint32

	// used for pagination
	Offset                pgtype.Timestamptz
	BookID                pgtype.Text
	StudentStudyPlanBooks pgtype.TextArray
}

func (r *BookRepo) ListBooks(ctx context.Context, db database.QueryExecer, args *ListBooksArgs) ([]*entities.Book, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.FindBooks")
	defer span.End()

	b := &entities.Book{}
	fields, _ := b.FieldMap()
	selectFields := strings.Join(fields, ", ")
	query := fmt.Sprintf(`
	SELECT %s
		FROM books
			WHERE ((deleted_at IS NULL AND ($1::TEXT[] IS NULL OR book_id = ANY($1::TEXT[]))) OR (book_id = ANY($5::TEXT[])))
		AND (($2::TIMESTAMPTZ IS NULL AND $3::TEXT IS NULL) OR ((created_at, book_id) < ($2, $3)))
	ORDER BY created_at DESC , book_id DESC
	LIMIT $4;
`, selectFields)

	books := entities.Books{}
	if err := database.Select(ctx, db, query, &args.BookIDs, &args.Offset, &args.BookID, &args.Limit, &args.StudentStudyPlanBooks).ScanAll(&books); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return books, nil
}

func (r *BookRepo) DuplicateBook(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, bookName pgtype.Text) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.DuplicateBook")
	defer span.End()

	var now pgtype.Timestamptz
	if err := now.Set(time.Now().UTC()); err != nil {
		return "", err
	}

	e := &entities.Book{}
	bookFieldNames := database.GetFieldNames(e)
	selectFields := golibs.Replace(
		bookFieldNames,
		[]string{"book_id", "name", "copied_from", "created_at", "updated_at"},
		[]string{"uuid_generate_v4()", "$1", "book_id", "$2", "$2"},
	)

	var newBookID string
	query := fmt.Sprintf(`
		INSERT INTO 
			%s (%s)
		SELECT
			%s
		FROM
			%s
		WHERE
			book_id = $3
		RETURNING
			book_id
	`, e.TableName(), strings.Join(bookFieldNames, ", "), strings.Join(selectFields, ", "), e.TableName())

	if err := db.QueryRow(ctx, query, &bookName, &now, &bookID).Scan(&newBookID); err != nil {
		return "", err
	}
	return newBookID, nil
}

func (r *BookRepo) UpdateCurrentChapterDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedChapterDisplayOrder pgtype.Int4, bookID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "BookRepo.UpdateCurrentChapter")
	defer span.End()

	updateCurrentChapterDisplayOrderStmtPlt := `UPDATE books SET current_chapter_display_order = current_chapter_display_order + $1::INT WHERE book_id = $2::TEXT AND deleted_at IS NULL`
	_, err := db.Exec(ctx, updateCurrentChapterDisplayOrderStmtPlt, totalGeneratedChapterDisplayOrder, bookID)
	if err != nil {
		return err
	}

	return nil
}

const retrieveAdHocBookByCourseIDAndStudentIDStmtTpl = `SELECT
	sp.%s
FROM books AS b
JOIN study_plans AS sp ON sp.book_id = b.book_id
JOIN student_study_plans AS ssp ON sp.study_plan_id = ssp.study_plan_id
WHERE b.book_type = 'BOOK_TYPE_ADHOC'::TEXT 
	AND sp.course_id = $1::TEXT
	AND sp.study_plan_type = 'STUDY_PLAN_TYPE_INDIVIDUAL'::TEXT
	AND sp.master_study_plan_id IS NULL
	AND ssp.student_id = $2::TEXT
	AND sp.deleted_at IS NULL
	AND ssp.deleted_at IS NULL
	AND b.deleted_at IS NULL
`

func (r *BookRepo) RetrieveAdHocBookByCourseIDAndStudentID(ctx context.Context, db database.QueryExecer, courseID, studentID pgtype.Text) (*entities.Book, error) {
	book := &entities.Book{}
	fieldNames := database.GetFieldNames(book)

	query := fmt.Sprintf(retrieveAdHocBookByCourseIDAndStudentIDStmtTpl, strings.Join(fieldNames, ", b."))
	err := database.Select(ctx, db, query, &courseID, &studentID).ScanOne(book)
	if err != nil {
		return nil, err
	}
	return book, nil
}
