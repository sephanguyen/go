package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	lentities "github.com/manabie-com/backend/internal/tom/domain/lesson"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ConversationLessonRepo struct {
}

func (r *ConversationLessonRepo) UpdateLatestStartTime(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, startTime pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.FindAndUpdateLatestCallID")
	defer span.End()
	updateStmt := `
UPDATE conversation_lesson
SET 
    latest_start_time = $2 WHERE lesson_id = $1 AND deleted_at IS NULL`
	_, err := db.Exec(ctx, updateStmt, &lessonID, &startTime)
	return err
}

func (r *ConversationLessonRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*lentities.ConversationLesson) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.BulkUpsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *lentities.ConversationLesson) {
		fieldNames := database.GetFieldNames(e)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT conversation_lesson_pk 
		DO UPDATE SET updated_at = $4`,
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, c := range conversations {
		queueFn(b, c)
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

func (r *ConversationLessonRepo) FindByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray, includeSoftDeleted bool) ([]*lentities.ConversationLesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationLessonRepo.FindByLessonID")
	defer span.End()

	c := new(lentities.ConversationLesson)
	fields := database.GetFieldNames(c)

	selectStmt := fmt.Sprintf("SELECT %s FROM conversation_lesson WHERE lesson_id = ANY($1)", strings.Join(fields, ","))
	if !includeSoftDeleted {
		selectStmt += " and deleted_at IS NULL"
	}

	rows, err := db.Query(ctx, selectStmt, &lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	ret := []*lentities.ConversationLesson{}
	for rows.Next() {
		e := &lentities.ConversationLesson{}
		err := rows.Scan(database.GetScanFields(e, fields)...)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		ret = append(ret, e)
	}
	return ret, nil
}

func (r *ConversationLessonRepo) FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*lentities.ConversationLesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationLessonRepo.FindByLessonID")
	defer span.End()

	c := new(lentities.ConversationLesson)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = $1", strings.Join(fields, ","), c.TableName())

	row := db.QueryRow(ctx, selectStmt, &lessonID)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}

	return c, nil
}

func (r *ConversationLessonRepo) BulkUpdateResourcePath(ctx context.Context, db database.QueryExecer, lessons []string, resourcePath string) error {
	stmt := `
update conversation_lesson cl set resource_path = $1
where cl.lesson_id = ANY($2)
and (cl.resource_path is null or length(cl.resource_path)=0)`
	_, err := db.Exec(ctx, stmt, database.Text(resourcePath), database.TextArray(lessons))
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}
