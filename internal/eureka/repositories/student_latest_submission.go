package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// StudentSubmissionRepo works with "student_submissions" table
type StudentLatestSubmissionRepo struct{}

const upsertStmt = `
INSERT INTO %s (%s) VALUES (%s)
ON CONFLICT ON CONSTRAINT student_latest_submissions_old_uk
DO UPDATE SET
	student_submission_id = $1,
	submission_content = $5,
	check_list = $6,
	note = $7,
	student_submission_grade_id = $8,
	status = $9,
	created_at = $10,
	updated_at = $11,
	deleted_at = $12,
	deleted_by = $13,
	editor_id = $14,
  complete_date = $15,
  duration = $16,
  correct_score = $17,
  total_score = $18,
  understanding_level = $19
`

func (r *StudentLatestSubmissionRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.StudentLatestSubmission) error {
	e.UpdatedAt.Set(timeutil.Now())

	fieldNames := database.GetFieldNames(e)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(upsertStmt, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	scanFields := database.GetScanFields(e, fieldNames)
	if _, err := db.Exec(ctx, query, scanFields...); err != nil {
		return err
	}

	return nil
}

func (r *StudentLatestSubmissionRepo) UpsertV2(ctx context.Context, db database.QueryExecer, e *entities.StudentLatestSubmission) error {
	e.UpdatedAt.Set(timeutil.Now())

	fieldNames := database.GetFieldNames(e)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	const upsertStmt = `
	INSERT INTO %s (%s) VALUES (%s)
	ON CONFLICT ON CONSTRAINT student_latest_submissions_pk
	DO UPDATE SET
		student_submission_id = $1,
		submission_content = $5,
		check_list = $6,
		note = $7,
		student_submission_grade_id = $8,
		status = $9,
		created_at = $10,
		updated_at = $11,
		deleted_at = $12,
		deleted_by = $13,
		editor_id = $14,
		complete_date = $15,
		duration = $16,
		correct_score = $17,
		total_score = $18,
		understanding_level = $19
  `
	query := fmt.Sprintf(upsertStmt, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	scanFields := database.GetScanFields(e, fieldNames)
	if _, err := db.Exec(ctx, query, scanFields...); err != nil {
		return err
	}

	return nil
}

func (r *StudentLatestSubmissionRepo) QueueUpsertStudentLatestSubmission(b *pgx.Batch, item *entities.StudentLatestSubmission) {
	fieldNames := database.GetFieldNames(item)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(upsertStmt, item.TableName(), strings.Join(fieldNames, ","), placeHolders)
	scanFields := database.GetScanFields(item, fieldNames)
	b.Queue(query, scanFields...)
}

const bulkUpsertStmt = `
INSERT INTO %s (%s) 
VALUES %s
ON CONFLICT ON CONSTRAINT student_latest_submissions_old_uk
DO UPDATE SET
	student_submission_id = excluded.student_submission_id,
	submission_content = excluded.submission_content,
	check_list = excluded.check_list,
	note = excluded.note,
	student_submission_grade_id = excluded.student_submission_grade_id,
	status = excluded.status,
	created_at = excluded.created_at,
	updated_at = excluded.updated_at,
	deleted_at = excluded.deleted_at,
	deleted_by = excluded.deleted_by,
	editor_id = excluded.editor_id,
  complete_date = excluded.complete_date,
  duration = excluded.duration,
  correct_score = excluded.correct_score,
  total_score = excluded.total_score,
  understanding_level = excluded.understanding_level
`

func (r *StudentLatestSubmissionRepo) BulkUpserts(ctx context.Context, db database.QueryExecer, studentLatestSubmission []*entities.StudentLatestSubmission) error {
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertStmt, studentLatestSubmission)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertStudentLatestSubmission error: %s", err.Error())
	}
	return nil
}

func (r *StudentLatestSubmissionRepo) DeleteByStudyPlanItemID(
	ctx context.Context, db database.QueryExecer,
	studyPlanItemID, deletedBy pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentLatestSubmissionRepo.DeleteByStudyPlanItemID")
	defer span.End()

	a := &entities.StudentLatestSubmission{}
	query := fmt.Sprintf("UPDATE %s SET deleted_at = now(), deleted_by = $1 WHERE study_plan_item_id = $2", a.TableName())
	commandTag, err := db.Exec(ctx, query, &deletedBy, &studyPlanItemID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no raw affected, failed delete study plan item")
	}
	return nil
}
