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

	"github.com/google/uuid"
	"go.uber.org/multierr"
)

type StudentPackageOrderRepo struct{}

func (r *StudentPackageOrderRepo) Create(ctx context.Context, db database.QueryExecer, e entities.StudentPackageOrder) (err error) {
	_ = e.ID.Set(uuid.New().String())
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageOrder.Create")
	defer span.End()

	cmdTag, err := database.Insert(ctx, &e, db.Exec)
	if err != nil {
		err = fmt.Errorf("err insert StudentPackageOrder: %w", err)
		return
	}

	if cmdTag.RowsAffected() != 1 {
		err = fmt.Errorf("err insert StudentPackageOrder: %d RowsAffected", cmdTag.RowsAffected())
	}

	return
}

func (r *StudentPackageOrderRepo) ResetCurrentPosition(ctx context.Context, db database.QueryExecer, studentPackageID string) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.ResetCurrentPosition")
	defer span.End()

	studentPackageOrder := &entities.StudentPackageOrder{}
	sql := fmt.Sprintf(`UPDATE %s SET is_current_student_package = false, updated_at = now() 
                         WHERE student_package_id = $1 AND deleted_at IS NULL`, studentPackageOrder.TableName())
	_, err = db.Exec(ctx, sql, studentPackageID)
	if err != nil {
		return fmt.Errorf("err db.Exec studentPackageOrderRepo.ResetCurrentPosition: %w", err)
	}

	return nil
}

func (r *StudentPackageOrderRepo) GetStudentPackageOrdersByStudentPackageID(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageID string,
) (studentPackageOrders []*entities.StudentPackageOrder, err error) {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID")
	defer span.End()

	table := entities.StudentPackageOrder{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE student_package_id = $1 AND deleted_at IS NULL ORDER BY start_at ASC`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentPackageID)
	if err != nil {
		err = fmt.Errorf("error when query studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID: %v", err)
		return nil, err
	}

	defer rows.Close()

	var result []*entities.StudentPackageOrder
	for rows.Next() {
		studentPackageOrder := new(entities.StudentPackageOrder)
		_, fieldValues := studentPackageOrder.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, studentPackageOrder)
	}
	return result, nil
}

func (r *StudentPackageOrderRepo) GetStudentPackageOrderByTimeAndStudentPackageID(ctx context.Context, db database.QueryExecer, studentPackageID string, startTime time.Time) (studentPackageOrder *entities.StudentPackageOrder, err error) {
	studentPackageOrder = &entities.StudentPackageOrder{}
	studentPackageOrderFieldNames, studentPackageOrderFieldValues := studentPackageOrder.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_id = $1 AND start_at <= $2 AND end_at >= $3 AND deleted_at is null
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentPackageOrderFieldNames, ","),
		studentPackageOrder.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentPackageID, startTime, startTime)
	err = row.Scan(studentPackageOrderFieldValues...)
	if err != nil {
		return
	}
	return studentPackageOrder, nil
}

func (r *StudentPackageOrderRepo) SetCurrentStudentPackageByID(ctx context.Context, db database.QueryExecer, id string, isCurrent bool) error {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.SetCurrentStudentPackageByID")
	defer span.End()

	studentPackageOrder := &entities.StudentPackageOrder{}
	sql := fmt.Sprintf(`UPDATE %s SET is_current_student_package = $1, updated_at = now() 
                         WHERE student_package_order_id = $2 AND deleted_at IS NULL`, studentPackageOrder.TableName())
	_, err := db.Exec(ctx, sql, isCurrent, id)
	if err != nil {
		return fmt.Errorf("err db.Exec studentPackageOrderRepo.SetCurrentStudentPackageByID: %w", err)
	}

	return nil
}

func (r *StudentPackageOrderRepo) SoftDeleteByID(ctx context.Context, db database.QueryExecer, id string) error {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.SoftDeleteByID")
	defer span.End()

	studentPackageOrder := &entities.StudentPackageOrder{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now(), updated_at = now() 
                         WHERE student_package_order_id = $1
                           AND deleted_at IS NULL`, studentPackageOrder.TableName())
	_, err := db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("err db.Exec studentPackageOrderRepo.SoftDeleteByID: %w", err)
	}

	return nil
}

func (r *StudentPackageOrderRepo) RevertByID(ctx context.Context, db database.QueryExecer, id string) error {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.RevertByID")
	defer span.End()

	studentPackageOrder := &entities.StudentPackageOrder{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = NULL, updated_at = now() 
                         WHERE student_package_order_id = $1`, studentPackageOrder.TableName())
	_, err := db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("err db.Exec studentPackageOrderRepo.RevertByID: %w", err)
	}

	return nil
}

func (r *StudentPackageOrderRepo) Update(ctx context.Context, db database.QueryExecer, e entities.StudentPackageOrder) error {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}
	cmdTag, err := database.UpdateFields(ctx, &e, db.Exec, "student_package_order_id", []string{
		"user_id",
		"order_id",
		"course_id",
		"start_at",
		"end_at",
		"student_package_object",
		"student_package_id",
		"is_current_student_package",
		"updated_at",
		"deleted_at",
	})
	if err != nil {
		return fmt.Errorf("err update Student Package Order: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Student Package Order: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *StudentPackageOrderRepo) Upsert(
	ctx context.Context,
	tx database.QueryExecer,
	studentPackageOrder entities.StudentPackageOrder,
) (
	err error,
) {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.Upsert")
	defer span.End()
	now := time.Now()
	err = multierr.Combine(
		studentPackageOrder.CreatedAt.Set(now),
		studentPackageOrder.UpdatedAt.Set(now),
	)

	if err != nil {
		return fmt.Errorf("multierr.Err: %w", err)
	}

	var fieldNames []string
	updateCommand := "start_at = $5, end_at = $6, student_package_object = $7, is_current_student_package = $9, updated_at = $11, from_student_package_order_id = $13"
	fieldNames = database.GetFieldNamesExcepts(&studentPackageOrder, []string{"resource_path"})
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT student_package_order_id__pk DO UPDATE
		SET %s`, studentPackageOrder.TableName(), strings.Join(fieldNames, ","), placeHolders, updateCommand)
	args := database.GetScanFields(&studentPackageOrder, fieldNames)
	commandTag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error when upsert student package order studentPackageOrderRepo.Upsert: %v", err)
	}

	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("error when upsert student package order with no affected row")
	}
	return
}

func (r *StudentPackageOrderRepo) GetByStudentPackageIDAndOrderID(ctx context.Context, db database.QueryExecer, studentPackageID, orderID string) (studentPackageOrder *entities.StudentPackageOrder, err error) {
	studentPackageOrder = &entities.StudentPackageOrder{}
	studentPackageOrderFieldNames, studentPackageOrderFieldValues := studentPackageOrder.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_id = $1 AND order_id = $2 AND deleted_at is null
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentPackageOrderFieldNames, ","),
		studentPackageOrder.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentPackageID, orderID)
	err = row.Scan(studentPackageOrderFieldValues...)
	if err != nil {
		err = fmt.Errorf("row.Scan studentPackageOrderRepo.GetByStudentPackageIDAndOrderID: %w", err)
		return
	}
	return studentPackageOrder, nil
}

func (r *StudentPackageOrderRepo) GetByStudentPackageOrderID(ctx context.Context, db database.QueryExecer, studentPackageOrderID string) (studentPackageOrder *entities.StudentPackageOrder, err error) {
	studentPackageOrder = &entities.StudentPackageOrder{}
	studentPackageOrderFieldNames, studentPackageOrderFieldValues := studentPackageOrder.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_order_id = $1 AND deleted_at is null
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentPackageOrderFieldNames, ","),
		studentPackageOrder.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentPackageOrderID)
	err = row.Scan(studentPackageOrderFieldValues...)
	if err != nil {
		err = fmt.Errorf("row.Scan studentPackageOrderRepo.GetByStudentPackageOrderID: %w", err)
		return
	}
	return studentPackageOrder, nil
}

func (r *StudentPackageOrderRepo) UpdateExecuteError(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrder entities.StudentPackageOrder,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.UpdateExecuteError")
	defer span.End()
	sql := fmt.Sprintf(`UPDATE %s SET executed_error = $1, updated_at = now() 
						WHERE student_package_order_id = $2;`, studentPackageOrder.TableName())
	_, err = db.Exec(ctx, sql,
		studentPackageOrder.ExecutedError.String,
		studentPackageOrder.ID.String,
	)

	if err != nil {
		err = fmt.Errorf("err db.Exec studentPackageOrderRepo.UpdateExecuteError: %w", err)
	}
	return
}

func (r *StudentPackageOrderRepo) UpdateExecuteStatus(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrder entities.StudentPackageOrder,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "studentPackageOrderRepo.UpdateExecuteStatus")
	defer span.End()
	sql := fmt.Sprintf(`UPDATE %s SET is_executed_by_cronjob = true, updated_at = now() 
						WHERE student_package_order_id = $1;`, studentPackageOrder.TableName())
	_, err = db.Exec(ctx, sql,
		studentPackageOrder.ID.String,
	)

	if err != nil {
		err = fmt.Errorf("err db.Exec studentPackageOrderRepo.UpdateExecuteStatus: %w", err)
	}
	return
}
