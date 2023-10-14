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

type ExamLOSubmissionAnswerRepo struct{}

type ExamLOSubmissionAnswerFilter struct {
	SubmissionID      pgtype.Text
	ShuffledQuizSetID pgtype.Text
}

const listExamLOSubmissionAnswerStmtTpl = `
SELECT %s
FROM %s elsa
INNER JOIN (
	SELECT quiz.id as quiz_id, quiz.idx as display_order FROM
	%s sqs, UNNEST(sqs.quiz_external_ids) WITH ORDINALITY AS quiz(id, idx)
	WHERE shuffled_quiz_set_id = $2
	AND deleted_at IS NULL
) sqs ON elsa.quiz_id = sqs.quiz_id
WHERE 
	elsa.deleted_at IS NULL
	AND ($1::TEXT IS NULL OR submission_id = $1)
	AND ($2::TEXT IS NULL OR shuffled_quiz_set_id = $2)
ORDER BY sqs.display_order
`

func (r *ExamLOSubmissionAnswerRepo) List(ctx context.Context, db database.QueryExecer, filter *ExamLOSubmissionAnswerFilter) ([]*entities.ExamLOSubmissionAnswer, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionAnswerRepo.List")
	defer span.End()

	args := []interface{}{
		&filter.SubmissionID,
		&filter.ShuffledQuizSetID,
	}

	e := &entities.ExamLOSubmissionAnswer{}
	es := entities.ExamLOSubmissionAnswers{}
	sqs := &entities.ShuffledQuizSet{}

	listStmt := fmt.Sprintf(listExamLOSubmissionAnswerStmtTpl, database.SerializeFields(e, "elsa"), e.TableName(), sqs.TableName())

	if err := database.Select(ctx, db, listStmt, args...).ScanAll(&es); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return es, nil
}

func (r *ExamLOSubmissionAnswerRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmissionAnswer) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionAnswerRepo.Upsert")
	defer span.End()

	fieldNames := database.GetFieldNames(e)
	scanFields := database.GetScanFields(e, fieldNames)

	stmt := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT ON CONSTRAINT exam_lo_submission_answer_pk DO UPDATE SET
        study_plan_id = EXCLUDED.study_plan_id,
        learning_material_id = EXCLUDED.learning_material_id,
        shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id,
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        point = EXCLUDED.point,
        created_at = EXCLUDED.created_at,
        updated_at = EXCLUDED.updated_at,
        deleted_at = EXCLUDED.deleted_at;
	`

	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(stmt, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	cmdTag, err := db.Exec(ctx, query, scanFields...)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}

	return int(cmdTag.RowsAffected()), nil
}

func (r *ExamLOSubmissionAnswerRepo) Delete(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionAnswerRepo.DeleteSubmissionAnswerBySubmissionId")
	defer span.End()

	e := entities.ExamLOSubmissionAnswer{}
	query := fmt.Sprintf(deleteSubmissionQuery, e.TableName())
	cmdTag, err := db.Exec(ctx, query, submissionID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (r *ExamLOSubmissionAnswerRepo) UpdateAcceptedQuizPointsByQuizID(ctx context.Context, db database.QueryExecer, quizID pgtype.Text, newPoint pgtype.Int4) error {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionAnswerRepo.UpdateQuizPointByQuizID")
	defer span.End()

	e := entities.ExamLOSubmissionAnswer{}
	stmt := fmt.Sprintf(`UPDATE %s SET point = $1, updated_at = now() WHERE quiz_id = $2 AND is_accepted = 't'`, e.TableName())

	_, err := db.Exec(ctx, stmt, newPoint, quizID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}
