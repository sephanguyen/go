package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type StudentPackageAccessPathRepo struct{}

const studentPackageAccessPathRepoBulkUpsertStmtTpl = `INSERT INTO %s AS sp (%s) VALUES (%s)
ON CONFLICT (student_package_id, course_id, student_id, location_id)
DO UPDATE SET
	deleted_at = NULL,
	updated_at = NOW()`

func (r *StudentPackageAccessPathRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudentPackageAccessPath) error {
	b := &pgx.Batch{}
	e := &entities.StudentPackageAccessPath{}
	currentTime := timeutil.Now().UTC()

	for _, item := range items {
		if err := multierr.Combine(
			item.UpdatedAt.Set(currentTime),
			item.CreatedAt.Set(currentTime),
			item.DeletedAt.Set(nil),
		); err != nil {
			return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set DeletedAt.Set: %w", err)
		}
		fieldNames, value := item.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(studentPackageAccessPathRepoBulkUpsertStmtTpl, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

		b.Queue(query, value...)
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

const studentPackageAccessPathRepoDeleteByStudentPackageIDsStmtTpl = `
	UPDATE %s 
	SET deleted_at = NOW() 
	WHERE student_package_id = ANY($1) 
	AND deleted_at IS NULL`

func (r *StudentPackageAccessPathRepo) DeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, spIDs pgtype.TextArray) error {
	spap := &entities.StudentPackageAccessPath{}
	query := fmt.Sprintf(studentPackageAccessPathRepoDeleteByStudentPackageIDsStmtTpl, spap.TableName())
	_, err := db.Exec(ctx, query, &spIDs)
	if err != nil {
		return err
	}
	return nil
}

const studentPackageAccessPathRepoDeleteByStudentIDsStmtTpl = `
	UPDATE %s 
	SET deleted_at = NOW() 
	WHERE student_id = ANY($1) 
	AND deleted_at IS NULL`

func (r *StudentPackageAccessPathRepo) DeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error {
	spap := &entities.StudentPackageAccessPath{}
	query := fmt.Sprintf(studentPackageAccessPathRepoDeleteByStudentIDsStmtTpl, spap.TableName())
	_, err := db.Exec(ctx, query, &studentIDs)
	if err != nil {
		return err
	}
	return nil
}

func (r *StudentPackageAccessPathRepo) GetByCourseIDAndLocationIDs(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationIDs pgtype.TextArray) ([]*entities.StudentPackageAccessPath, error) {
	e := &entities.StudentPackageAccessPath{}
	query := fmt.Sprintf(`
		SELECT DISTINCT %s 
		FROM student_package_access_path
		WHERE course_id = $1::TEXT
		AND ((ARRAY_LENGTH($2::TEXT[], 1) IS NULL) OR (location_id = ANY($2::TEXT[])))
		AND deleted_at IS NULL; 
	`, strings.Join(database.GetFieldNames(e), ", "))

	rows, err := db.Query(ctx, query, courseID, locationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentPackages := make([]*entities.StudentPackageAccessPath, 0)
	for rows.Next() {
		e := &entities.StudentPackageAccessPath{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		studentPackages = append(studentPackages, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return studentPackages, nil
}
