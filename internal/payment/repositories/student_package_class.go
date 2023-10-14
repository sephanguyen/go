package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type StudentPackageClassRepo struct {
}

func (r *StudentPackageClassRepo) Upsert(ctx context.Context, db database.QueryExecer, studentPackageClass *entities.StudentPackageClass) (err error) {
	var fieldNames []string
	updateCommand := "created_at = $6,updated_at = $7, deleted_at = $8"
	fieldNames = database.GetFieldNames(studentPackageClass)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT student_package_class_pk DO UPDATE
		SET %s`, studentPackageClass.TableName(), strings.Join(fieldNames, ","), placeHolders, updateCommand)
	args := database.GetScanFields(studentPackageClass, fieldNames)
	commandTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error when upsert student package class: %v", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("error when upsert student package class in payment")
	}

	return
}

func (r *StudentPackageClassRepo) Delete(ctx context.Context, db database.QueryExecer, studentPackageClass *entities.StudentPackageClass) (err error) {
	var commandTag pgconn.CommandTag
	query := fmt.Sprintf(`UPDATE %s 
SET updated_at = $1, deleted_at = $2
WHERE student_package_id = $3 AND student_id =$4 AND location_id = $5 AND course_id = $6 AND class_id = $7`, studentPackageClass.TableName())

	commandTag, err = db.Exec(ctx, query,
		studentPackageClass.UpdatedAt,
		studentPackageClass.DeletedAt,
		studentPackageClass.StudentPackageID,
		studentPackageClass.StudentID,
		studentPackageClass.LocationID,
		studentPackageClass.CourseID,
		studentPackageClass.ClassID,
	)
	if err != nil {
		err = fmt.Errorf("delete student package class have error: %w", err)
		return
	}

	if commandTag.RowsAffected() == 0 {
		err = fmt.Errorf("delete student package class have no row affected")
	}
	return
}

func (r *StudentPackageClassRepo) GetByStudentPackageID(ctx context.Context, db database.QueryExecer, studentPackageID string) (studentPackageClass entities.StudentPackageClass, err error) {
	fieldNames, fieldValues := studentPackageClass.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM %s
				WHERE student_package_id = $1 and deleted_at is null
				FOR NO KEY UPDATE
				`,
		strings.Join(fieldNames, ","),
		studentPackageClass.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentPackageID)
	err = row.Scan(fieldValues...)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			err = nil
			return
		}
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}
