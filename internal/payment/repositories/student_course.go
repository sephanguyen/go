package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgconn"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentCourseRepo struct{}

func (r *StudentCourseRepo) GetStudentCoursesByStudentPackageIDForUpdate(
	ctx context.Context,
	tx database.QueryExecer,
	studentPackageID string,
) (studentCourseEntities []entities.StudentCourse, err error) {
	studentCourse := &entities.StudentCourse{}
	studentCourseFieldNames, studentCourseFieldValues := studentCourse.FieldMap()
	stmt := fmt.Sprintf(`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_id = $1
		FOR NO KEY UPDATE
		`,
		strings.Join(studentCourseFieldNames, ","),
		studentCourse.TableName(),
	)
	rows, err := tx.Query(ctx, stmt, studentPackageID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(studentCourseFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		studentCourseEntities = append(studentCourseEntities, *studentCourse)
	}
	return
}

func (r *StudentCourseRepo) UpsertStudentCourseData(
	ctx context.Context,
	tx database.QueryExecer,
	studentCourseEntities []entities.StudentCourse,
) (
	err error,
) {
	ctx, span := interceptors.StartSpan(ctx, "StudentCourseRepo.Upsert")
	defer span.End()
	now := time.Now()
	for _, e := range studentCourseEntities {
		err = multierr.Combine(
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)

		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		var fieldNames []string
		updateCommand := "student_start_date = $5, student_end_date = $6, course_slot = $7, course_slot_per_week = $8, weight = $9 ,updated_at = $11, deleted_at = $12"
		fieldNames = database.GetFieldNamesExcepts(&e, []string{"resource_path"})
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT student_course_pk DO UPDATE
		SET %s`, e.TableName(), strings.Join(fieldNames, ","), placeHolders, updateCommand)
		args := database.GetScanFields(&e, fieldNames)
		commandTag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("error when upsert student course: %v", err)
		}

		if commandTag.RowsAffected() != 1 {
			return fmt.Errorf("error when upsert student course in payment")
		}
	}
	return
}

func (r *StudentCourseRepo) SoftDeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, studentPackageIDs []string, deletedAt time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentCourseRepo.SoftDeleteByStudentPackageIDs")
	defer span.End()

	studentCourse := &entities.StudentCourse{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = $1, updated_at = now() 
                         WHERE student_package_id = ANY($2) 
                           AND deleted_at IS NULL`, studentCourse.TableName())
	_, err := db.Exec(ctx, sql, deletedAt, database.TextArray(studentPackageIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec StudentCourseRepo.SoftDeleteByStudentPackageIDs: %w", err)
	}

	return nil
}

func (r *StudentCourseRepo) VoidStudentCoursesByStudentPackageID(
	ctx context.Context,
	tx database.QueryExecer,
	studentEndDate time.Time,
	studentPackageID string,
) (
	err error,
) {
	ctx, span := interceptors.StartSpan(ctx, "StudentCourseRepo.VoidStudentCoursesByStudentPackageID")
	defer span.End()
	e := &entities.StudentCourse{}
	query := fmt.Sprintf(`
	UPDATE %s SET student_end_date = $1, updated_at = now() 
	WHERE student_package_id = $2 AND deleted_at IS NOT NULL`, e.TableName())
	commandTag, err := tx.Exec(ctx, query, studentEndDate, studentPackageID)
	if err != nil {
		return fmt.Errorf("error when void student course: %v", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("error when void student courses have no row affected")
	}
	return
}

func (r *StudentCourseRepo) GetStudentCoursesByStudentPackageIDsForUpdate(ctx context.Context, db database.QueryExecer, studentPackageIDs []string) (studentCourses []entities.StudentCourse, err error) {
	entity := &entities.StudentCourse{}
	studentCourseFieldNames, _ := entity.FieldMap()
	stmt := fmt.Sprintf(`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_id = ANY($1::text[]) AND deleted_at is null
		FOR NO KEY UPDATE
		`,
		strings.Join(studentCourseFieldNames, ","),
		entity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentPackageIDs)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		studentCourse := new(entities.StudentCourse)
		_, fieldValues := studentCourse.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = status.Errorf(codes.Internal, "Err while scan student course")
			return
		}
		studentCourses = append(studentCourses, *studentCourse)
	}
	return
}

func (r *StudentCourseRepo) UpsertStudentCourse(
	ctx context.Context,
	tx database.QueryExecer,
	studentCourse entities.StudentCourse,
) (
	err error,
) {
	ctx, span := interceptors.StartSpan(ctx, "StudentCourseRepo.Upsert")
	defer span.End()
	now := time.Now()
	err = multierr.Combine(
		studentCourse.CreatedAt.Set(now),
		studentCourse.UpdatedAt.Set(now),
	)

	if err != nil {
		return fmt.Errorf("multierr.Err: %w", err)
	}

	var fieldNames []string
	updateCommand := "student_start_date = $5, student_end_date = $6, course_slot = $7, course_slot_per_week = $8, weight = $9 ,updated_at = $11, deleted_at = $12"
	fieldNames = database.GetFieldNamesExcepts(&studentCourse, []string{"resource_path"})
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT student_course_pk DO UPDATE
		SET %s`, studentCourse.TableName(), strings.Join(fieldNames, ","), placeHolders, updateCommand)
	args := database.GetScanFields(&studentCourse, fieldNames)
	commandTag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error when upsert student course: %v", err)
	}

	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("error when upsert student course with no affected row")
	}
	return
}

func (r *StudentCourseRepo) UpdateTimeByID(ctx context.Context, db database.QueryExecer, id string, courseID string, endTime time.Time) (err error) {
	var commandTag pgconn.CommandTag
	now := time.Now()
	studentCourse := &entities.StudentCourse{}

	query := fmt.Sprintf(`
		UPDATE %s
		SET student_end_date = $1, updated_at = $2
		WHERE student_package_id= $3 and course_id= $4`, studentCourse.TableName())

	commandTag, err = db.Exec(ctx, query, endTime, now, id, courseID)
	if err != nil {
		err = fmt.Errorf("update time student course have error: %w", err)
		return
	}

	if commandTag.RowsAffected() == 0 {
		err = fmt.Errorf("update time student course have no row affected")
	}
	return
}

func (r *StudentCourseRepo) CancelByStudentPackageIDAndCourseID(ctx context.Context, db database.QueryExecer, studentPackageID string, courseID string) (err error) {
	var commandTag pgconn.CommandTag
	now := time.Now()
	studentCourse := &entities.StudentCourse{}

	query := fmt.Sprintf(`
		UPDATE %s
		SET updated_at = $1, deleted_at = $2
		WHERE student_package_id= $3 and course_id= $4`, studentCourse.TableName())

	commandTag, err = db.Exec(ctx, query, now, now, studentPackageID, courseID)
	if err != nil {
		err = fmt.Errorf("cancel student course have error: %w", err)
		return
	}

	if commandTag.RowsAffected() == 0 {
		err = fmt.Errorf("cancel student course have no row affected")
	}
	return
}

func (r *StudentCourseRepo) GetByStudentIDAndCourseIDAndLocationID(ctx context.Context, db database.QueryExecer, studentID, courseID, locationID string) (studentCourse entities.StudentCourse, err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentCourseRepo.GetByStudentIDAndCourseIDAndLocationID")
	defer span.End()

	fieldNames, fieldValues := studentCourse.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM %s
				WHERE student_id = $1 AND course_id = $2 AND location_id = $3 AND deleted_at IS NULL
				FOR NO KEY UPDATE
				`,
		strings.Join(fieldNames, ","),
		studentCourse.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, courseID, locationID)
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf("err db.Exec StudentCourseRepo.GetByStudentIDAndCourseIDAndLocationID: %w", err)
		return
	}
	return
}
