package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type StudentPackageRepo struct {
}

func (s *StudentPackageRepo) CurrentPackage(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.StudentPackage, error) {
	e := &entities.StudentPackage{}
	fields, _ := e.FieldMap()

	sql := fmt.Sprintf(`SELECT %s
		FROM %s
		WHERE student_id = $1 AND NOW() BETWEEN start_at AND end_at AND is_active = TRUE`,
		strings.Join(fields, ","), e.TableName())

	results := entities.StudentPackages{}
	err := database.Select(ctx, db, sql, &studentID).ScanAll(&results)
	if err != nil {
		return nil, fmt.Errorf("err database.Select: %w", err)
	}

	return results, nil
}

func (s *StudentPackageRepo) Insert(ctx context.Context, db database.QueryExecer, e *entities.StudentPackage) error {
	now := time.Now()
	if err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return fmt.Errorf("StudentPackageRepo.Insert: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot create student_packages")
	}

	return nil
}

const studentPackageRepoBulkInsertStmtTpl = `INSERT INTO %s AS sp (%s)
VALUES (%s) ON
CONFLICT (student_package_id) DO
UPDATE
SET
	updated_at = NOW(),
	is_active = TRUE
WHERE sp.is_active = FALSE`

// BulkInsert will also update matched rows to active if given rows are not
func (s *StudentPackageRepo) BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudentPackage) error {
	b := &pgx.Batch{}
	e := &entities.StudentPackage{}
	time := timeutil.Now().UTC()

	for _, item := range items {
		fieldNames, value := item.FieldMap()
		placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10"

		query := fmt.Sprintf(studentPackageRepoBulkInsertStmtTpl, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

		if item.CreatedAt.Status != pgtype.Present && item.UpdatedAt.Status != pgtype.Present {
			b.Queue(query, append(value[:8], time, time)...)
		} else {
			b.Queue(query, value...)
		}
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

func (s *StudentPackageRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.StudentPackage) error {
	now := time.Now()
	_ = e.UpdatedAt.Set(now)

	cmdTag, err := database.Update(ctx, e, db.Exec, "student_package_id")
	if err != nil {
		return fmt.Errorf("StudentPackageRepo.Update: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update student_packages")
	}

	return nil
}

func (s *StudentPackageRepo) Get(ctx context.Context, db database.QueryExecer, studentPackageID pgtype.Text) (*entities.StudentPackage, error) {
	e := &entities.StudentPackage{}
	fields, _ := e.FieldMap()

	sql := fmt.Sprintf(`SELECT %s
			FROM %s
			WHERE student_package_id = $1`,
		strings.Join(fields, ","), e.TableName(),
	)

	err := database.Select(ctx, db, sql, &studentPackageID).ScanOne(e)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return e, nil
}

const studentPackageRepoSoftDeleteStmt = `UPDATE student_packages SET is_active = FALSE
WHERE student_id = $1
AND is_active = TRUE`

// SoftDelete marks rows is_active = false if current active
func (s *StudentPackageRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageRepo.SoftDelete")
	defer span.End()

	_, err := db.Exec(ctx, studentPackageRepoSoftDeleteStmt, &studentID)
	if err != nil {
		return err
	}

	return nil
}

const softDeleteByIDsSoftDeleteByIDs = `UPDATE student_packages SET is_active = FALSE
WHERE student_package_id = ANY($1)
AND is_active = TRUE`

// SoftDeleteByIDs marks student_packages as inactive if currently active
func (s *StudentPackageRepo) SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageRepo.SoftDelete")
	defer span.End()

	_, err := db.Exec(ctx, softDeleteByIDsSoftDeleteByIDs, &ids)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentPackageRepo) GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.StudentPackage, error) {
	e := &entities.StudentPackage{}
	query := fmt.Sprintf(`
		SELECT %s 
		FROM student_packages
		WHERE student_id = ANY($1)
		ORDER BY updated_at; 
	`, strings.Join(database.GetFieldNames(e), ", "))

	rows, err := db.Query(ctx, query, studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentPackages := make([]*entities.StudentPackage, 0)
	for rows.Next() {
		e := &entities.StudentPackage{}
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

func (s *StudentPackageRepo) GetByCourseIDAndLocationIDs(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationIDs pgtype.TextArray) ([]*entities.StudentPackage, error) {
	e := &entities.StudentPackage{}
	query := fmt.Sprintf(`
		SELECT %s 
		FROM student_packages
		WHERE properties->'can_do_quiz' @> '["%s"]'
		AND ((ARRAY_LENGTH($1::TEXT[], 1) IS NULL) OR (location_ids && $1::TEXT[]))
		AND is_active = TRUE
		AND deleted_at IS NULL; 
	`, strings.Join(database.GetFieldNames(e), ", "), courseID.String)

	rows, err := db.Query(ctx, query, locationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentPackages := make([]*entities.StudentPackage, 0)
	for rows.Next() {
		e := &entities.StudentPackage{}
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

func (s *StudentPackageRepo) GetByStudentPackageIDAndStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, studentPackageID pgtype.Text, studentID pgtype.Text, courseID pgtype.Text) (*entities.StudentPackage, error) {
	e := &entities.StudentPackage{}

	sql := fmt.Sprintf(`
		SELECT %s 
		FROM %s
		WHERE properties->'can_do_quiz' @> '["%s"]' 
		and student_package_id = $1
		and student_id = $2
		and deleted_at is null
		AND is_active = TRUE
	`, strings.Join(database.GetFieldNames(e), ", "), e.TableName(), courseID.String)

	err := database.Select(ctx, db, sql, &studentPackageID, &studentID).ScanOne(e)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return e, nil
}
