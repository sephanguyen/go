package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentPackageRepo struct {
}

func (r *StudentPackageRepo) Insert(ctx context.Context, db database.QueryExecer, studentPackage *entities.StudentPackages) (err error) {
	now := time.Now()
	_ = multierr.Combine(
		studentPackage.CreatedAt.Set(now),
		studentPackage.UpdatedAt.Set(now),
		studentPackage.PackageID.Set(nil),
		studentPackage.DeletedAt.Set(nil),
	)
	_, err = database.Insert(ctx, studentPackage, db.Exec)
	if err != nil {
		err = fmt.Errorf("insert student package have error: %w", err)
	}
	return
}

func (r *StudentPackageRepo) Update(ctx context.Context, db database.QueryExecer, studentPackage *entities.StudentPackages) (err error) {
	var commandTag pgconn.CommandTag
	now := time.Now()
	_ = multierr.Combine(
		studentPackage.UpdatedAt.Set(now),
		studentPackage.IsActive.Set(true),
		studentPackage.PackageID.Set(nil),
	)

	commandTag, err = database.UpdateFields(ctx, studentPackage, db.Exec, "student_package_id", []string{
		"start_at",
		"end_at",
		"properties",
		"is_active",
		"location_ids",
		"updated_at",
		"deleted_at",
	})
	if err != nil {
		missingFields := fmt.Sprintf("start_at: %s,end_at: %s, properties: %s, is_active: %s, location_ids: %s, updated_at: %s, deleted_at: %s",
			string(studentPackage.StartAt.Status),
			string(studentPackage.EndAt.Status),
			string(studentPackage.Properties.Status),
			string(studentPackage.IsActive.Status),
			string(studentPackage.LocationIDs.Status),
			string(studentPackage.UpdatedAt.Status),
			string(studentPackage.DeletedAt.Status))
		err = fmt.Errorf("update student package have error: %w and field status : %s", err, missingFields)
		return
	}

	if commandTag.RowsAffected() == 0 {
		err = fmt.Errorf("update student package have no row affected")
	}
	return
}

func (r *StudentPackageRepo) GetByID(ctx context.Context, db database.QueryExecer, studentPackageID string) (studentPackage entities.StudentPackages, err error) {
	fieldNames, fieldValues := studentPackage.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM %s
				WHERE student_package_id = $1
				FOR NO KEY UPDATE
				`,
		strings.Join(fieldNames, ","),
		studentPackage.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentPackageID)
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}

func (r *StudentPackageRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.StudentPackages) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageRepo.Upsert")
	defer span.End()

	now := time.Now()
	_ = multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)

	var fieldNames []string
	updateCommand := "start_at = $4, end_at = $5, properties = $6, is_active = $7, location_ids = $8 ,updated_at = $10, deleted_at = $11"
	fieldNames = database.GetFieldNamesExcepts(e, []string{"resource_path"})
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__student_packages DO UPDATE
		SET %s`, e.TableName(), strings.Join(fieldNames, ","), placeHolders, updateCommand)
	args := database.GetScanFields(e, fieldNames)
	commandTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error when upsert student package: %v", err)
	}

	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("upsert student package have no row affected")
	}

	return
}

func (r *StudentPackageRepo) GetStudentPackageForUpsert(ctx context.Context, db database.QueryExecer, e *entities.StudentPackages) (
	studentPackageID pgtype.Text,
	err error,
) {
	stmt := fmt.Sprintf(
		`SELECT student_package_id 
				FROM %s
				WHERE student_id = $1 AND package_id = $2 AND '%s' = ANY(location_ids)
				`,
		e.TableName(),
		e.LocationIDs.Elements[0].String,
	)
	row := db.QueryRow(ctx, stmt, e.StudentID.String, e.PackageID.String)
	err = row.Scan(&studentPackageID)
	if err == nil {
		return
	}
	if err.Error() == pgx.ErrNoRows.Error() {
		err = nil
	}
	return
}

func (r *StudentPackageRepo) SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, ids []string, deletedAt time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageRepo.SoftDeleteByIDs")
	defer span.End()

	studentPackage := &entities.StudentPackages{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = $1, updated_at = now() 
                         WHERE student_package_id = ANY($2) 
                           AND deleted_at IS NULL`, studentPackage.TableName())
	_, err := db.Exec(ctx, sql, deletedAt, database.TextArray(ids))
	if err != nil {
		return fmt.Errorf("err db.Exec StudentPackageRepo.SoftDeleteByIDs: %w", err)
	}

	return nil
}

func (r *StudentPackageRepo) UpdateTimeByID(ctx context.Context, db database.QueryExecer, id string, endTime time.Time) (err error) {
	var commandTag pgconn.CommandTag
	now := time.Now()
	studentPackage := &entities.StudentPackages{}
	_ = multierr.Combine(
		studentPackage.UpdatedAt.Set(now),
		studentPackage.ID.Set(id),
		studentPackage.EndAt.Set(endTime),
	)

	commandTag, err = database.UpdateFields(ctx, studentPackage, db.Exec, "student_package_id", []string{
		"end_at",
		"updated_at",
	})
	if err != nil {
		err = fmt.Errorf("update time student package have error: %w", err)
		return
	}

	if commandTag.RowsAffected() == 0 {
		err = fmt.Errorf("update time student package have no row affected")
	}
	return
}

func (r *StudentPackageRepo) CancelByID(ctx context.Context, db database.QueryExecer, id string) (err error) {
	var commandTag pgconn.CommandTag
	now := time.Now()
	studentPackage := &entities.StudentPackages{}
	_ = multierr.Combine(
		studentPackage.UpdatedAt.Set(now),
		studentPackage.ID.Set(id),
		studentPackage.DeletedAt.Set(now),
	)

	commandTag, err = database.UpdateFields(ctx, studentPackage, db.Exec, "student_package_id", []string{
		"updated_at",
		"deleted_at",
	})
	if err != nil {
		err = fmt.Errorf("cancel student package have error: %w", err)
		return
	}

	if commandTag.RowsAffected() == 0 {
		err = fmt.Errorf("cancel student package have no row affected")
	}
	return
}

// GetStudentPackagesForCronjobByDay
// Get student packages with end_at is in X latest days
func (r *StudentPackageRepo) GetStudentPackagesForCronjobByDay(ctx context.Context, db database.QueryExecer, days int) (studentPackages []entities.StudentPackages, err error) {
	entity := &entities.StudentPackages{}
	studentPackageFieldNames, _ := entity.FieldMap()
	stmt := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE 
			end_at::DATE >= current_date - $1::INTEGER AND end_at::DATE <= current_date
			AND deleted_at is null;
		`,
		strings.Join(studentPackageFieldNames, ","),
		entity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, days)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		studentPackage := new(entities.StudentPackages)
		_, fieldValues := studentPackage.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = status.Errorf(codes.Internal, "err when scan student package")
			return
		}
		studentPackages = append(studentPackages, *studentPackage)
	}
	return
}
