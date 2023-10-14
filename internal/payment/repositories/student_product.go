package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type StudentProductRepo struct{}

// Create creates StudentProduct entity
func (r *StudentProductRepo) Create(ctx context.Context, db database.QueryExecer, e entities.StudentProduct) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentProductRepo.Create")
	defer span.End()

	cmdTag, err := database.InsertExcept(ctx, &e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert StudentProduct: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert StudentProduct: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *StudentProductRepo) GetLatestEndDateStudentProductWithProductIDAndStudentID(ctx context.Context, db database.QueryExecer, studentID, productID string) (studentProducts []*entities.StudentProduct, err error) {
	var (
		rows pgx.Rows
	)
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt :=
		`SELECT %s
		FROM %s
		WHERE product_id = $1 AND student_id = $2 AND ((end_date is not null) OR (end_date is null AND product_status != 'CANCELLED'))
		ORDER BY end_date DESC
		LIMIT 1
		FOR NO KEY UPDATE;`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err = db.Query(ctx, stmt, productID, studentID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmpStudentProduct := new(entities.StudentProduct)
		_, fieldValues := tmpStudentProduct.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return
		}
		studentProducts = append(studentProducts, tmpStudentProduct)
	}
	return
}

// Update StudentProduct entity
func (r *StudentProductRepo) UpdateWithVersionNumber(ctx context.Context, db database.QueryExecer, e entities.StudentProduct, versionNumber int32) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentProductRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}
	if e.StudentProductLabel.Status == pgtype.Undefined {
		_ = e.StudentProductLabel.Set(nil)
	}
	cmdTag, err := database.UpdateFieldsForVersionNumber(ctx, &e, db.Exec, "student_product_id", []string{
		"student_id",
		"product_id",
		"upcoming_billing_date",
		"start_date",
		"end_date",
		"product_status",
		"approval_status",
		"updated_at",
		"deleted_at",
		"location_id",
		"updated_from_student_product_id",
		"updated_to_student_product_id",
		"student_product_label",
	}, versionNumber)
	if err != nil {
		return fmt.Errorf("err update Student Product: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Student Product: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *StudentProductRepo) Update(ctx context.Context, db database.QueryExecer, e entities.StudentProduct) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentProductRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}
	if e.StudentProductLabel.Status == pgtype.Undefined {
		_ = e.StudentProductLabel.Set(nil)
	}
	cmdTag, err := database.UpdateFields(ctx, &e, db.Exec, "student_product_id", []string{
		"student_id",
		"product_id",
		"upcoming_billing_date",
		"start_date",
		"end_date",
		"product_status",
		"approval_status",
		"updated_at",
		"deleted_at",
		"location_id",
		"updated_from_student_product_id",
		"updated_to_student_product_id",
		"student_product_label",
	})
	if err != nil {
		return fmt.Errorf("err update Student Product: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Student Product: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *StudentProductRepo) GetStudentProductForUpdateByStudentProductID(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
) (studentProduct entities.StudentProduct, err error) {
	studentProductFieldNames, studentProductFieldValues := (&studentProduct).FieldMap()

	stmt := fmt.Sprintf(`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = $1 
		`,
		strings.Join(studentProductFieldNames, ","),
		studentProduct.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentProductID)
	err = row.Scan(studentProductFieldValues...)
	if err != nil {
		return entities.StudentProduct{}, err
	}
	return studentProduct, nil
}

func (r *StudentProductRepo) UpdateStatusStudentProductAndResetStudentProductLabel(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
	studentProductStatus string,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentProductRepo.UpdateStatusStudentProductAndResetStudentProductLabel")
	defer span.End()

	now := time.Now()
	e := entities.StudentProduct{}
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}
	_ = e.StudentProductLabel.Set(nil)
	_ = e.StudentProductID.Set(studentProductID)
	_ = e.ProductStatus.Set(studentProductStatus)
	cmdTag, err := database.UpdateFields(ctx, &e, db.Exec, "student_product_id", []string{
		"product_status",
		"updated_at",
		"student_product_label",
	})
	if err != nil {
		return fmt.Errorf("err update Student Product: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Student Product: %d RowsAffected", cmdTag.RowsAffected())
	}
	return
}

func (r *StudentProductRepo) GetStudentProductsByStudentProductLabelForUpdate(
	ctx context.Context,
	db database.QueryExecer,
	studentProductLabels []string,
) (
	studentProducts []*entities.StudentProduct,
	err error,
) {
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE
			student_product_label = ANY($1)
			FOR NO KEY UPDATE;
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)

	rows, err := db.Query(ctx, stmt, studentProductLabels)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentProducts = append(studentProducts, studentProduct)
	}
	return
}

func (r *StudentProductRepo) GetUniqueProductsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) ([]*entities.StudentProduct, error) {
	var studentProducts []*entities.StudentProduct
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1
			AND is_unique = true
			ORDER BY product_id, created_at DESC
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentProducts = append(studentProducts, studentProduct)
	}
	return studentProducts, nil
}

func (r *StudentProductRepo) GetUniqueProductsByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentProduct, error) {
	var studentProducts []*entities.StudentProduct
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = ANY($1)
			AND is_unique = true
			ORDER BY product_id, created_at DESC
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentProducts = append(studentProducts, studentProduct)
	}
	return studentProducts, nil
}

func (r *StudentProductRepo) GetStudentProductByStudentProductID(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
) (studentProduct entities.StudentProduct, err error) {
	studentProductFieldNames, studentProductFieldValues := (&studentProduct).FieldMap()

	stmt := fmt.Sprintf(`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = $1
		`,
		strings.Join(studentProductFieldNames, ","),
		studentProduct.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentProductID)
	err = row.Scan(studentProductFieldValues...)
	if err != nil {
		return entities.StudentProduct{}, err
	}
	return studentProduct, nil
}

func (r *StudentProductRepo) CountStudentProductIDsByStudentIDAndLocationIDs(ctx context.Context, db database.QueryExecer, studentID string, locationIDs []string) (total int, err error) {
	var rows pgx.Rows
	studentProduct := entities.StudentProduct{}
	if len(locationIDs) == 0 {
		stmt := fmt.Sprintf(
			`SELECT student_product_id FROM "%s" WHERE student_id = $1 AND
			root_student_product_id IS NULL AND
			is_associated = false
			`,
			studentProduct.TableName(),
		)
		rows, err = db.Query(ctx, stmt, studentID)
		if err != nil {
			return
		}
	} else {
		stmt := fmt.Sprintf(
			`SELECT student_product_id FROM "%s" WHERE student_id = $1 AND
			location_id = ANY($2) AND
			root_student_product_id IS NULL AND
			is_associated = false
			`,
			studentProduct.TableName(),
		)
		rows, err = db.Query(ctx, stmt, studentID, locationIDs)
		if err != nil {
			return
		}
	}

	defer rows.Close()

	for rows.Next() {
		total++
	}
	return
}

func (r *StudentProductRepo) GetStudentProductIDsByRootStudentProductID(ctx context.Context, db database.QueryExecer, rootStudentProductID string) ([]*entities.StudentProduct, error) {
	var studentProducts []*entities.StudentProduct
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			root_student_product_id = $1 OR
			student_product_id = $1
			ORDER BY created_at
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, rootStudentProductID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		studentProducts = append(studentProducts, studentProduct)
	}
	return studentProducts, nil
}

func (r *StudentProductRepo) GetByStudentIDAndLocationIDsWithPaging(ctx context.Context, db database.QueryExecer, studentID string, locationIDs []string, offset int64, limit int64) (result []*entities.StudentProduct, err error) {
	var rows pgx.Rows
	studentProduct := entities.StudentProduct{}
	fieldNames, _ := studentProduct.FieldMap()
	if len(locationIDs) == 0 {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE student_id = $1 AND
			root_student_product_id IS NULL AND
			is_associated = false
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			strings.Join(fieldNames, ","),
			studentProduct.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, limit, offset)
		if err != nil {
			return nil, err
		}
	} else {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE student_id = $1 AND
			location_id = ANY($2) AND
			root_student_product_id IS NULL AND
			is_associated = false
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
			strings.Join(fieldNames, ","),
			studentProduct.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, locationIDs, limit, offset)
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

func (r *StudentProductRepo) GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.StudentProduct, error) {
	var studentProducts []entities.StudentProduct
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = ANY($1)
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func (r *StudentProductRepo) GetStudentProductAssociatedByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductIDs []string) (result []*entities.StudentProduct, err error) {
	var rows pgx.Rows
	studentProduct := entities.StudentProduct{}
	fieldNames, _ := studentProduct.FieldMap()

	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE student_product_id = ANY($1) AND
			is_associated = true`,
		strings.Join(fieldNames, ","),
		studentProduct.TableName(),
	)

	rows, err = db.Query(ctx, stmt, studentProductIDs)
	if err != nil {
		return nil, err
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

func (r *StudentProductRepo) GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.StudentProduct, error) {
	studentProduct := &entities.StudentProduct{}
	productFieldNames, productFieldValues := studentProduct.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productFieldNames, ","),
		studentProduct.TableName(),
	)
	row := db.QueryRow(ctx, stmt, entitiesID)
	err := row.Scan(productFieldValues...)
	if err != nil {
		return entities.StudentProduct{}, err
	}
	return *studentProduct, nil
}

func (r *StudentProductRepo) GetActiveRecurringProductsOfStudentInLocation(ctx context.Context, db database.QueryExecer, studentID string, locationID string, ignoreStudentProductID []string) (result []entities.StudentProduct, err error) {
	studentProduct := entities.StudentProduct{}
	var rows pgx.Rows
	fieldNames, _ := studentProduct.FieldMap()
	studentProductFieldNamesWithPrefix := sliceutils.Map(fieldNames, func(fieldName string) string {
		return fmt.Sprintf("sp.%s", fieldName)
	})

	if len(ignoreStudentProductID) != 0 {
		stmt :=
			`
		SELECT %s
		FROM %s sp
		INNER JOIN product p on p.product_id = sp.product_id
		INNER JOIN product_setting ps on ps.product_id = sp.product_id
		WHERE
			sp.student_id = $1 AND
			sp.location_id = $2 AND
			sp.student_product_label <> 'PAUSED' AND
			sp.product_status <> 'CANCELLED' AND
			sp.student_product_id <> ANY($3) AND
			sp.end_date > Now() AND
			p.billing_schedule_id is not null AND
			ps.is_operation_fee = false
		ORDER BY
			sp.created_at DESC`

		stmt = fmt.Sprintf(
			stmt,
			strings.Join(studentProductFieldNamesWithPrefix, ","),
			studentProduct.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, locationID, ignoreStudentProductID)
		if err != nil {
			return nil, err
		}
	} else {
		stmt :=
			`
		SELECT %s
		FROM %s sp
		INNER JOIN product p on p.product_id = sp.product_id
		INNER JOIN product_setting ps on ps.product_id = sp.product_id
		WHERE
			sp.student_id = $1 AND
			sp.location_id = $2 AND
			sp.student_product_label <> 'PAUSED' AND
			sp.product_status <> 'CANCELLED' AND
			sp.end_date > Now() AND
			p.billing_schedule_id is not null AND
			ps.is_operation_fee = false
		ORDER BY
			sp.created_at DESC`

		stmt = fmt.Sprintf(
			stmt,
			strings.Join(studentProductFieldNamesWithPrefix, ","),
			studentProduct.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, locationID)
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
		result = append(result, *studentProduct)
	}
	return result, nil
}

func (r *StudentProductRepo) GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (result []string, err error) {
	studentProduct := entities.StudentProduct{}

	stmt :=
		`
		SELECT updated_to_student_product_id
		FROM %s sp
		INNER JOIN product p on p.product_id = sp.product_id
		WHERE
			sp.student_id = $1 AND
			sp.location_id = $2 AND
			sp.student_product_label = 'UPDATE_SCHEDULED' AND
			sp.end_date > Now() AND
			p.billing_schedule_id is not null 
		ORDER BY
			sp.created_at DESC`

	stmt = fmt.Sprintf(
		stmt,
		studentProduct.TableName(),
	)

	rows, err := db.Query(ctx, stmt, studentID, locationID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var updateToStudentProduct string
		err := rows.Scan(&updateToStudentProduct)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, updateToStudentProduct)
	}
	return result, nil
}

func (r *StudentProductRepo) GetActiveOperationFeeOfStudent(ctx context.Context, db database.QueryExecer, studentID string) (result []entities.StudentProduct, err error) {
	studentProduct := entities.StudentProduct{}
	var rows pgx.Rows
	fieldNames, _ := studentProduct.FieldMap()
	studentProductFieldNamesWithPrefix := sliceutils.Map(fieldNames, func(fieldName string) string {
		return fmt.Sprintf("sp.%s", fieldName)
	})
	stmt :=
		`
		SELECT %s
		FROM %s sp
		INNER JOIN product p on p.product_id = sp.product_id
		INNER JOIN product_setting ps on sp.product_id = ps.product_id
		WHERE
			sp.student_id = $1 AND
			sp.student_product_label <> 'PAUSED' AND
			sp.end_date > Now() AND
			sp.product_status <> 'CANCELLED' AND
			p.billing_schedule_id is not null AND
			ps.is_operation_fee = true
		ORDER BY
			sp.created_at DESC`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNamesWithPrefix, ","),
		studentProduct.TableName(),
	)

	rows, err = db.Query(ctx, stmt, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, *studentProduct)
	}
	return result, nil
}
