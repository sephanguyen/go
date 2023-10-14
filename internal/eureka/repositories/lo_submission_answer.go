package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LOSubmissionAnswerRepo struct{}

type LOSubmissionAnswerFilter struct {
	SubmissionID      pgtype.Text
	ShuffledQuizSetID pgtype.Text
}

func (r *LOSubmissionAnswerRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.LOSubmissionAnswer) error {
	ctx, span := interceptors.StartSpan(ctx, "LOSubmissionAnswerRepo.Upsert")
	defer span.End()

	fieldNames := database.GetFieldNames(e)
	scanFields := database.GetScanFields(e, fieldNames)

	stmt := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT ON CONSTRAINT lo_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        point = EXCLUDED.point,
        updated_at = EXCLUDED.updated_at,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer;
	`

	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(stmt, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	cmdTag, err := db.Exec(ctx, query, scanFields...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no row affected")
	}

	return nil
}

func (r *LOSubmissionAnswerRepo) ListSubmissionAnswers(
	ctx context.Context, db database.QueryExecer, setID pgtype.Text, limit, offset pgtype.Int8,
) ([]*entities.LOSubmissionAnswer, []pgtype.Text, error) {
	sqs := &entities.ShuffledQuizSet{}
	getExternalQuizIDsStmt := fmt.Sprintf(`
	SELECT quiz_id FROM
		%s sqs, UNNEST(sqs.quiz_external_ids) WITH ORDINALITY AS quiz_id
	WHERE sqs.shuffled_quiz_set_id = $1::TEXT
	ORDER BY ORDINALITY
	LIMIT $2 OFFSET $3;
	`, sqs.TableName())

	rows, err := db.Query(ctx, getExternalQuizIDsStmt, setID, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	quizIDs := make([]pgtype.Text, 0, limit.Int)
	for rows.Next() {
		quizID := pgtype.Text{}
		if err := rows.Scan(&quizID); err != nil {
			return nil, nil, fmt.Errorf("rows.Scan: %w", err)
		}
		quizIDs = append(quizIDs, quizID)
	}

	lsa := &entities.LOSubmissionAnswer{}
	fields, _ := lsa.FieldMap()
	stmt := fmt.Sprintf(`
	SELECT %s FROM %s WHERE
		shuffled_quiz_set_id = $1::TEXT
		AND quiz_id = ANY($2::_TEXT);
	`, strings.Join(fields, ","), lsa.TableName())
	answers := entities.LOSubmissionAnswers{}
	if err := database.Select(ctx, db, stmt, setID, quizIDs).ScanAll(&answers); err != nil {
		return nil, nil, fmt.Errorf("database.Select: %w", err)
	}

	return answers, quizIDs, nil
}

func (r *LOSubmissionAnswerRepo) List(ctx context.Context, db database.QueryExecer, filter *LOSubmissionAnswerFilter) ([]*entities.LOSubmissionAnswer, error) {
	ctx, span := interceptors.StartSpan(ctx, "LOSubmissionAnswerRepo.List")
	defer span.End()

	args := []interface{}{
		&filter.SubmissionID,
		&filter.ShuffledQuizSetID,
	}

	stmt := `
	SELECT %s
	FROM %s
	WHERE 
		deleted_at IS NULL
		AND ($1::TEXT IS NULL OR submission_id = $1)
		AND ($2::TEXT IS NULL OR shuffled_quiz_set_id = $2)
	`

	lo := &entities.LOSubmissionAnswer{}
	los := entities.LOSubmissionAnswers{}
	listStmt := fmt.Sprintf(stmt, strings.Join(database.GetFieldNames(lo), ","), lo.TableName())

	if err := database.Select(ctx, db, listStmt, args...).ScanAll(&los); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return los, nil
}

func (r *LOSubmissionAnswerRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.LOSubmissionAnswers) error {
	ctx, span := interceptors.StartSpan(ctx, "LOSubmissionAnswerRepo.BulkUpsert")
	defer span.End()

	stmt := `
    INSERT INTO %s (%s) VALUES %s
    ON CONFLICT ON CONSTRAINT lo_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        point = EXCLUDED.point,
        updated_at = EXCLUDED.updated_at,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer;
	`

	err := dbeureka.BulkUpsert(ctx, db, stmt, items)
	if err != nil {
		return fmt.Errorf("database.BulkUpsert error: %s", err.Error())
	}

	return nil
}
