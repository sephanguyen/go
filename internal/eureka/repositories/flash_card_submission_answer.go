package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type FlashCardSubmissionAnswerRepo struct{}

func (r *FlashCardSubmissionAnswerRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.FlashCardSubmissionAnswer) error {
	ctx, span := interceptors.StartSpan(ctx, "FlashCardSubmissionAnswerRepo.Upsert")
	defer span.End()

	fieldNames := database.GetFieldNames(e)
	scanFields := database.GetScanFields(e, fieldNames)

	stmt := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT ON CONSTRAINT flash_card_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        point = EXCLUDED.point,
        updated_at = EXCLUDED.updated_at;
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

func (r *FlashCardSubmissionAnswerRepo) ListSubmissionAnswers(
	ctx context.Context, db database.QueryExecer, setID pgtype.Text, limit, offset pgtype.Int8,
) ([]*entities.FlashCardSubmissionAnswer, []pgtype.Text, error) {
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

	fcsa := &entities.FlashCardSubmissionAnswer{}
	fields, _ := fcsa.FieldMap()
	stmt := fmt.Sprintf(`
	SELECT %s FROM %s WHERE
		shuffled_quiz_set_id = $1::TEXT
		AND quiz_id = ANY($2::_TEXT);
	`, strings.Join(fields, ","), fcsa.TableName())
	answers := entities.FlashCardSubmissionAnswers{}
	if err := database.Select(ctx, db, stmt, setID, quizIDs).ScanAll(&answers); err != nil {
		return nil, nil, fmt.Errorf("database.Select: %w", err)
	}

	return answers, quizIDs, nil
}
