package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/payment/constant"

	"github.com/jackc/pgx/v4"
)

type StudentProductRepo struct{}

func (r *StudentProductRepo) GetActiveStudentProductsByStudentIDAndLocationID(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (result []*entities.StudentProduct, err error) {
	var (
		rows pgx.Rows
		stmt string
	)
	studentProduct := entities.StudentProduct{}
	fieldNames, _ := studentProduct.FieldMap()
	studentProductFieldNamesWithPrefix := sliceutils.Map(fieldNames, func(fieldName string) string {
		return fmt.Sprintf("sp.%s", fieldName)
	})

	if locationID != "" {
		stmt = fmt.Sprintf(`
		SELECT %s
		FROM %s sp
		INNER JOIN product p on p.product_id = sp.product_id
		INNER JOIN product_setting ps on ps.product_id = sp.product_id
		WHERE
			sp.student_id = $1 AND
			sp.location_id = $2 AND
			sp.student_product_label IS DISTINCT FROM 'PAUSED' AND
			sp.product_status <> 'CANCELLED' AND
			((
				sp.start_date <= Now() AND
				sp.end_date > Now()
			) OR
			(
				sp.start_date > Now() AND
				sp.end_date > Now() AND
				sp.updated_from_student_product_id IS NULL
			)) AND
			p.billing_schedule_id IS NOT NULL AND
			ps.is_operation_fee = false
		ORDER BY
			sp.created_at DESC`,
			strings.Join(studentProductFieldNamesWithPrefix, ","),
			studentProduct.TableName(),
		)
		rows, err = db.Query(ctx, stmt, studentID, locationID)
		if err != nil {
			return nil, err
		}
	} else {
		stmt = fmt.Sprintf(`
		SELECT %s
		FROM %s sp
		INNER JOIN product p on p.product_id = sp.product_id
		INNER JOIN product_setting ps on ps.product_id = sp.product_id
		WHERE
			sp.student_id = $1 AND
			sp.student_product_label IS DISTINCT FROM 'PAUSED' AND
			sp.product_status <> 'CANCELLED' AND
			((
				sp.start_date <= Now() AND
				sp.end_date > Now()
			) OR
			(
				sp.start_date > Now() AND
				sp.end_date > Now() AND
				sp.updated_from_student_product_id IS NULL
			)) AND
			p.billing_schedule_id IS NOT NULL AND
			ps.is_operation_fee = false
		ORDER BY
			sp.created_at DESC`,
			strings.Join(studentProductFieldNamesWithPrefix, ","),
			studentProduct.TableName(),
		)
		rows, err = db.Query(ctx, stmt, studentID)
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, studentProduct)
	}
	return result, nil
}

func (r *StudentProductRepo) GetByID(ctx context.Context, db database.QueryExecer, id string) (entities.StudentProduct, error) {
	studentProduct := &entities.StudentProduct{}
	studentProductFieldNames, studentProductFieldValues := studentProduct.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProduct.TableName(),
	)
	row := db.QueryRow(ctx, stmt, id)
	err := row.Scan(studentProductFieldValues...)
	if err != nil {
		return entities.StudentProduct{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *studentProduct, nil
}

func (r *StudentProductRepo) GetByIDs(ctx context.Context, db database.QueryExecer, studentProductIDs []string) ([]entities.StudentProduct, error) {
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = ANY($1)`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentProductIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentProducts := []entities.StudentProduct{}
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentProducts = append(studentProducts, *studentProduct)
	}
	return studentProducts, nil
}
