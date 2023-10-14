package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentPackageAccessPathRepo struct {
}

func (r *StudentPackageAccessPathRepo) Insert(ctx context.Context, db database.QueryExecer, studentPackageAccessPath *entities.StudentPackageAccessPath) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageAccessPathRepo.Upsert")
	defer span.End()

	now := time.Now()

	err = multierr.Combine(
		studentPackageAccessPath.CreatedAt.Set(now),
		studentPackageAccessPath.UpdatedAt.Set(now),
		studentPackageAccessPath.DeletedAt.Set(nil),
		studentPackageAccessPath.AccessPath.Set(nil),
	)

	if err != nil {
		err = fmt.Errorf("multierr.Err: %w", err)
		return
	}

	fieldNames := database.GetFieldNamesExcepts(studentPackageAccessPath, []string{"resource_path"})
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)`, studentPackageAccessPath.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := database.GetScanFields(studentPackageAccessPath, fieldNames)
	_, err = db.Exec(ctx, query, args...)
	if err != nil {
		err = fmt.Errorf("error when insert student package access path: %v", err)
	}
	return
}

func (r *StudentPackageAccessPathRepo) Update(ctx context.Context, db database.QueryExecer, studentPackageAccessPath *entities.StudentPackageAccessPath) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageAccessPathRepo.Upsert")
	defer span.End()

	now := time.Now()

	err = multierr.Combine(
		studentPackageAccessPath.CreatedAt.Set(now),
		studentPackageAccessPath.UpdatedAt.Set(now),
		studentPackageAccessPath.DeletedAt.Set(nil),
		studentPackageAccessPath.AccessPath.Set(nil),
	)

	if err != nil {
		err = fmt.Errorf("multierr.Err: %w", err)
		return
	}
	fieldNames := database.GetFieldNamesExcepts(studentPackageAccessPath, []string{"resource_path"})
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT student_package_access_path_pk DO NOTHING`, studentPackageAccessPath.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := database.GetScanFields(studentPackageAccessPath, fieldNames)
	_, err = db.Exec(ctx, query, args...)
	if err != nil {
		err = fmt.Errorf("error when update student package access path: %v", err)
		return
	}
	return
}

func (r *StudentPackageAccessPathRepo) DeleteMulti(ctx context.Context, db database.QueryExecer, studentPackageAccessPaths []entities.StudentPackageAccessPath) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageAccessPathRepo.DeleteMulti")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.StudentPackageAccessPath) {
		query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = NOW()
		WHERE location_id = $1 AND student_package_id = $2 AND student_id =$3 AND course_id=$4`, u.TableName())
		b.Queue(
			query,
			u.LocationID.String,
			u.StudentPackageID.String,
			u.StudentID.String,
			u.CourseID.String,
		)
	}

	b := &pgx.Batch{}

	for i := 0; i < len(studentPackageAccessPaths); i++ {
		queueFn(b, &studentPackageAccessPaths[i])
	}

	batchResults := db.SendBatch(ctx, b)

	for i := 0; i < len(studentPackageAccessPaths); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("student package access path not inserted")
		}
	}

	defer batchResults.Close()

	return
}

func (r *StudentPackageAccessPathRepo) InsertMulti(ctx context.Context, db database.QueryExecer, studentPackageAccessPaths []entities.StudentPackageAccessPath) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageAccessPathRepo.InsertMulti")
	defer span.End()
	now := time.Now()

	queueFn := func(b *pgx.Batch, u *entities.StudentPackageAccessPath) {
		err = multierr.Combine(
			u.CreatedAt.Set(now),
			u.UpdatedAt.Set(now),
			u.DeletedAt.Set(nil),
			u.AccessPath.Set(nil),
		)

		if err != nil {
			err = fmt.Errorf("multierr.Err: %w", err)
			return
		}

		fieldNames := database.GetFieldNamesExcepts(u, []string{"resource_path"})
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)`, u.TableName(), strings.Join(fieldNames, ","), placeHolders)
		args := database.GetScanFields(u, fieldNames)

		b.Queue(
			query,
			args...,
		)
	}

	b := &pgx.Batch{}

	for i := 0; i < len(studentPackageAccessPaths); i++ {
		queueFn(b, &studentPackageAccessPaths[i])
	}

	batchResults := db.SendBatch(ctx, b)

	for i := 0; i < len(studentPackageAccessPaths); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("student package access path not inserted")
		}
	}

	defer batchResults.Close()

	return
}

func (r *StudentPackageAccessPathRepo) GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentIDs []string,
) (
	mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
	err error,
) {
	mapStudentCourseWithStudentPackageAccessPath = make(map[string]entities.StudentPackageAccessPath)
	table := &entities.StudentPackageAccessPath{}
	fieldNames, _ := table.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = ANY($1) AND deleted_at IS NULL
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmpStudentPackageAccessPath := &entities.StudentPackageAccessPath{}
		_, fieldValues := tmpStudentPackageAccessPath.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = fmt.Errorf("row.Scan: %w", err)
			return
		}
		key := fmt.Sprintf("%v_%v", tmpStudentPackageAccessPath.StudentID.String, tmpStudentPackageAccessPath.CourseID.String)
		mapStudentCourseWithStudentPackageAccessPath[key] = *tmpStudentPackageAccessPath
	}
	return
}

func (r *StudentPackageAccessPathRepo) SoftDeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, studentPackageIDs []string, deletedAt time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageAccessPathRepo.SoftDeleteByStudentPackageIDs")
	defer span.End()

	studentPackageAccessPath := &entities.StudentPackageAccessPath{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = $1, updated_at = now() 
                         WHERE student_package_id = ANY($2) 
                           AND deleted_at IS NULL`, studentPackageAccessPath.TableName())
	_, err := db.Exec(ctx, sql, deletedAt, database.TextArray(studentPackageIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec StudentPackageAccessPathRepo.SoftDeleteByStudentPackageIDs: %w", err)
	}

	return nil
}

func (r *StudentPackageAccessPathRepo) CheckExistStudentPackageAccessPath(
	ctx context.Context,
	db database.QueryExecer,
	studentID,
	courseID string,
) (
	err error,
) {
	var studentPackageAccessPath entities.StudentPackageAccessPath
	fieldNames, fieldValues := studentPackageAccessPath.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1 AND course_id = $2
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		studentPackageAccessPath.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, courseID)
	err = row.Scan(fieldValues...)
	if err == nil {
		err = status.Errorf(codes.FailedPrecondition, "duplicate student course id")
		return
	}
	if err.Error() == pgx.ErrNoRows.Error() {
		err = nil
		return
	}
	err = status.Errorf(codes.Internal, "get student package access path have error %v", err.Error())
	return
}

func (r *StudentPackageAccessPathRepo) RevertByStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, studentID, courseID string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageAccessPathRepo.RevertByStudentIDAndCourseID")
	defer span.End()

	studentPackageAccessPath := &entities.StudentPackageAccessPath{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = NULL, updated_at = now() 
                         WHERE student_id = $1 AND course_id = $2`, studentPackageAccessPath.TableName())
	cmd, err := db.Exec(ctx, sql, studentID, courseID)
	if err != nil {
		return fmt.Errorf("err db.Exec StudentPackageAccessPathRepo.RevertByStudentIDAndCourseID: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected StudentPackageAccessPathRepo.RevertByStudentIDAndCourseID")
	}

	return nil
}

func (r *StudentPackageAccessPathRepo) GetByStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, studentID, courseID string) (entities.StudentPackageAccessPath, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageAccessPathRepo.GetByStudentIDAndCourseID")
	defer span.End()

	studentPackageAccessPath := &entities.StudentPackageAccessPath{}
	studentFieldNames, studentFieldValues := studentPackageAccessPath.FieldMap()
	stmt := `
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1 AND course_id = $2 AND deleted_at is NULL
		FOR NO KEY UPDATE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentFieldNames, ","),
		studentPackageAccessPath.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, courseID)
	err := row.Scan(studentFieldValues...)
	if err != nil {
		return entities.StudentPackageAccessPath{}, fmt.Errorf("row.Scan StudentPackageAccessPathRepo.GetByStudentIDAndCourseID: %w", err)
	}
	return *studentPackageAccessPath, nil
}
