package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type SchoolHistoryRepo struct{}

func (r *SchoolHistoryRepo) SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error {
	sql := `UPDATE school_history SET deleted_at = now() WHERE student_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &studentIDs)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (r *SchoolHistoryRepo) Upsert(ctx context.Context, db database.QueryExecer, schoolHistories []*entity.SchoolHistory) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolHistoryRepo.Upsert")
	defer span.End()

	batch := &pgx.Batch{}
	if err := r.queueUpsert(ctx, batch, schoolHistories); err != nil {
		return fmt.Errorf("queueUpsert error: %w", err)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

func (r *SchoolHistoryRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, schoolHistories []*entity.SchoolHistory) error {
	queue := func(b *pgx.Batch, schoolHistory *entity.SchoolHistory) {
		fieldNames := database.GetFieldNames(schoolHistory)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT school_history__pk 
			DO UPDATE SET school_course_id = EXCLUDED.school_course_id, start_date = EXCLUDED.start_date, end_date = EXCLUDED.end_date, updated_at = now(), deleted_at = NULL`,
			schoolHistory.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		b.Queue(stmt, database.GetScanFields(schoolHistory, fieldNames)...)
	}

	now := time.Now()
	for _, schoolHistory := range schoolHistories {
		if schoolHistory.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := schoolHistory.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			schoolHistory.CreatedAt.Set(now),
			schoolHistory.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, schoolHistory)
	}

	return nil
}

func (r *SchoolHistoryRepo) GetByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entity.SchoolHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolHistoryRepo.GetByStudentID")
	defer span.End()

	schoolHistory := &entity.SchoolHistory{}
	fields := database.GetFieldNames(schoolHistory)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), schoolHistory.TableName())

	rows, err := db.Query(ctx, stmt, &studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schoolHistories := make([]*entity.SchoolHistory, 0)
	for rows.Next() {
		schoolHistory := &entity.SchoolHistory{}
		if err := rows.Scan(database.GetScanFields(schoolHistory, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		schoolHistories = append(schoolHistories, schoolHistory)
	}

	return schoolHistories, nil
}

func (r *SchoolHistoryRepo) GetSchoolHistoriesByGradeIDAndStudentID(ctx context.Context, db database.QueryExecer, gradeID pgtype.Text, studentID pgtype.Text, isCurrent pgtype.Bool) ([]*entity.SchoolHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolHistoryRepo.GetSchoolHistoriesByGradeID")
	defer span.End()

	schoolHistory := &entity.SchoolHistory{}
	fields := database.GetFieldNames(schoolHistory)

	stmt := fmt.Sprintf(`SELECT t1.%s FROM %s t1
    INNER JOIN school_info t2 on t1.school_id = t2.school_id
    INNER JOIN school_level t3 on t2.school_level_id = t3.school_level_id
    INNER JOIN school_level_grade t4 on t3.school_level_id = t4.school_level_id
    INNER JOIN grade t5 on t4.grade_id = t5.grade_id
         WHERE t5.grade_id = $1 AND t1.student_id = $2 AND t1.is_current = $3 AND t1.deleted_at IS NULL`, strings.Join(fields, ", t1."), schoolHistory.TableName())

	rows, err := db.Query(ctx, stmt, &gradeID, &studentID, &isCurrent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schoolHistories := make([]*entity.SchoolHistory, 0)
	for rows.Next() {
		schoolHistory := &entity.SchoolHistory{}
		if err := rows.Scan(database.GetScanFields(schoolHistory, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		schoolHistories = append(schoolHistories, schoolHistory)
	}

	return schoolHistories, nil
}

func (r *SchoolHistoryRepo) GetCurrentSchoolByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entity.SchoolHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolHistoryRepo.GetByStudentID")
	defer span.End()

	schoolHistory := &entity.SchoolHistory{}
	fields := database.GetFieldNames(schoolHistory)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND deleted_at IS NULL AND is_current = true", strings.Join(fields, ","), schoolHistory.TableName())

	rows, err := db.Query(ctx, stmt, &studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schoolHistories := make([]*entity.SchoolHistory, 0)
	for rows.Next() {
		schoolHistory := &entity.SchoolHistory{}
		if err := rows.Scan(database.GetScanFields(schoolHistory, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		schoolHistories = append(schoolHistories, schoolHistory)
	}

	return schoolHistories, nil
}

func (r *SchoolHistoryRepo) SetCurrentSchoolByStudentIDAndSchoolID(ctx context.Context, db database.QueryExecer, schoolID pgtype.Text, studentID pgtype.Text) error {
	sql := `UPDATE school_history SET is_current = true WHERE school_id = $1 AND student_id = $2 AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &schoolID, &studentID)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (r *SchoolHistoryRepo) RemoveCurrentSchoolByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) error {
	sql := `UPDATE school_history SET is_current = false , deleted_at = now() WHERE student_id = $1 AND deleted_at IS NULL AND is_current = true`
	_, err := db.Exec(ctx, sql, &studentID)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (r *SchoolHistoryRepo) UnsetCurrentSchoolByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) error {
	sql := `UPDATE school_history SET is_current = false WHERE student_id = $1 AND deleted_at IS NULL AND is_current = true`
	_, err := db.Exec(ctx, sql, &studentID)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (r *SchoolHistoryRepo) SetCurrentSchool(ctx context.Context, db database.QueryExecer, organizationID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "UserAccessPathRepo.Delete")
	defer span.End()

	query := `UPDATE school_history AS s SET is_current = true FROM school_history t1
        INNER JOIN students t6 on t1.student_id = t6.student_id
  		INNER JOIN school_info t2 on t1.school_id = t2.school_id
  		INNER JOIN school_level t3 on t2.school_level_id = t3.school_level_id
  		INNER JOIN school_level_grade t4 on t3.school_level_id = t4.school_level_id
  		INNER JOIN grade t5 on t4.grade_id = t5.grade_id
		WHERE t1.deleted_at IS NULL AND t1.is_current = false AND t1.resource_path = $1`
	_, err := db.Exec(ctx, query, &organizationID)
	if err != nil {
		return err
	}

	return nil
}
