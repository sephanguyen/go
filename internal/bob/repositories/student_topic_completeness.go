package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentTopicCompletenessRepo struct{}

func (r *StudentTopicCompletenessRepo) Upsert(ctx context.Context, db database.Ext, topics []*entities.StudentTopicCompleteness) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentTopicCompletenessRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities.StudentTopicCompleteness) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT students_topics_completeness_pk DO UPDATE SET total_finished_los = $3, updated_at = $5, is_completed = $6", t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range topics {
		t.CreatedAt.Set(now)
		t.UpdatedAt.Set(now)

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(topics); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("student's completed topic not inserted")
		}
	}
	return nil
}

func (r *StudentTopicCompletenessRepo) RetrieveByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, topicID *pgtype.Text) ([]*entities.StudentTopicCompleteness, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentTopicCompletenessRepo.RetrieveByStudentID")
	defer span.End()

	e := &entities.StudentTopicCompleteness{}
	topicE := &entities.Topic{}
	fields := database.GetFieldNames(e)
	args := []interface{}{&studentID}

	query := fmt.Sprintf("SELECT DISTINCT stc.%s FROM %s stc LEFT JOIN %s as t ON t.topic_id = stc.topic_id WHERE stc.student_id = $1 AND t.deleted_at IS NULL", strings.Join(fields, ", stc."), e.TableName(), topicE.TableName())
	if topicID != nil {
		args = append(args, topicID)
		query += " AND topic_id = $2"
	}
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var ss []*entities.StudentTopicCompleteness
	for rows.Next() {
		s := new(entities.StudentTopicCompleteness)
		if err := rows.Scan(database.GetScanFields(s, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		ss = append(ss, s)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return ss, nil
}

func (r *StudentTopicCompletenessRepo) RetrieveCompletedByStudentIDWeeklies(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*entities.StudentTopicCompleteness, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentTopicCompletenessRepo.RetrieveCompletedByStudentIDWeeklies")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentTopicCompleteness{})
	args := []interface{}{&studentID}
	query := fmt.Sprintf("SELECT DISTINCT stc.%s FROM students_topics_completeness stc LEFT JOIN topics t ON t.topic_id = stc.topic_id WHERE stc.student_id = $1 AND stc.is_completed=true AND t.deleted_at IS NULL", strings.Join(fields, ", stc."))

	if from != nil {
		args = append(args, from)
		query += fmt.Sprintf(" AND stc.updated_at >= $%d", len(args))
	}
	if to != nil {
		args = append(args, to)
		query += fmt.Sprintf(" AND stc.updated_at <= $%d", len(args))
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var ss []*entities.StudentTopicCompleteness
	for rows.Next() {
		s := new(entities.StudentTopicCompleteness)
		if err := rows.Scan(database.GetScanFields(s, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		ss = append(ss, s)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return ss, nil
}

func (r *StudentTopicCompletenessRepo) RetrieveByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, topicIDs *pgtype.TextArray) (map[string][]*entities.StudentTopicCompleteness, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentTopicCompletenessRepo.RetrieveByStudentID")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentTopicCompleteness{})
	args := []interface{}{&studentIDs}
	query := fmt.Sprintf("SELECT DISTINCT stc.%s FROM students_topics_completeness stc LEFT JOIN topics t ON t.topic_id = stc.topic_id WHERE stc.student_id = ANY($1) AND t.deleted_at IS NULL", strings.Join(fields, ", stc."))
	if topicIDs != nil {
		args = append(args, topicIDs)
		query += " AND stc.topic_id = ANY($2)"
	}
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	ret := make(map[string][]*entities.StudentTopicCompleteness)
	for rows.Next() {
		s := new(entities.StudentTopicCompleteness)
		if err := rows.Scan(database.GetScanFields(s, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		ret[s.StudentID.String] = append(ret[s.StudentID.String], s)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return ret, nil
}
