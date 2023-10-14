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

type ExamLOSubmissionScoreRepo struct{}

type ExamLOSubmissionScoreFilter struct {
	SubmissionID      pgtype.Text
	ShuffledQuizSetID pgtype.Text
}

const listExamLOSubmissionScoreStmtTpl = `
SELECT %s
FROM %s
WHERE 
	deleted_at IS NULL
	AND ($1::TEXT IS NULL OR submission_id = $1)
	AND ($2::TEXT IS NULL OR shuffled_quiz_set_id = $2)
`

func (r *ExamLOSubmissionScoreRepo) List(ctx context.Context, db database.QueryExecer, filter *ExamLOSubmissionScoreFilter) ([]*entities.ExamLOSubmissionScore, error) {
	args := []interface{}{
		&filter.SubmissionID,
		&filter.ShuffledQuizSetID,
	}

	e := &entities.ExamLOSubmissionScore{}
	es := entities.ExamLOSubmissionScores{}
	listStmt := fmt.Sprintf(listExamLOSubmissionScoreStmtTpl, strings.Join(database.GetFieldNames(e), ","), e.TableName())

	if err := database.Select(ctx, db, listStmt, args...).ScanAll(&es); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return es, nil
}

func (r *ExamLOSubmissionScoreRepo) Delete(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionAnswerRepo.DeleteSubmissionsScoreBySubmissionId")
	defer span.End()

	e := entities.ExamLOSubmissionScore{}
	query := fmt.Sprintf(deleteSubmissionQuery, e.TableName())
	cmdTag, err := db.Exec(ctx, query, submissionID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (r *ExamLOSubmissionScoreRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmissionScore) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionScoreRepo.Upsert")
	defer span.End()

	fieldNames := database.GetFieldNames(e)
	scanFields := database.GetScanFields(e, fieldNames)

	stmt := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT ON CONSTRAINT exam_lo_submission_score_pk DO UPDATE SET
        teacher_id = EXCLUDED.teacher_id,
        teacher_comment = EXCLUDED.teacher_comment,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        updated_at = EXCLUDED.updated_at,
        point = EXCLUDED.point,
        shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id;
	`

	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(stmt, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	cmdTag, err := db.Exec(ctx, query, scanFields...)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}

	return int(cmdTag.RowsAffected()), nil
}
