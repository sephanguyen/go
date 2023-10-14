package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type BillItemRepo struct{}

func (r *BillItemRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BillItem) (billItemSequenceNumber pgtype.Int4, err error) {
	ctx, span := interceptors.StartSpan(ctx, "BillingItemRepo.Create")
	defer span.End()

	now := time.Now()
	if err = multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
		e.Reference.Set(nil),
	); err != nil {
		err = fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
		return
	}
	if e.PreviousBillItemStatus.Status == pgtype.Undefined {
		_ = e.PreviousBillItemStatus.Set(nil)
	}
	if e.PreviousBillItemSequenceNumber.Status == pgtype.Undefined {
		_ = e.PreviousBillItemSequenceNumber.Set(nil)
	}
	if e.AdjustmentPrice.Status == pgtype.Undefined {
		_ = e.PreviousBillItemSequenceNumber.Set(nil)
	}

	err = database.InsertReturningAndExcept(ctx, e, db, []string{"bill_item_sequence_number", "resource_path"}, "bill_item_sequence_number", &billItemSequenceNumber)
	if err != nil {
		err = fmt.Errorf("err insert BillingItem: %w", err)
		return
	}
	_ = e.BillItemSequenceNumber.Set(billItemSequenceNumber)
	return
}

func (r *BillItemRepo) SetNonLatestBillItemByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "BillItemRepo.SetNonLatestBillItemByStudentProductID")
	defer span.End()
	var (
		billItem   entities.BillItem
		commandTag pgconn.CommandTag
	)
	stmt :=
		`
		UPDATE %s
		SET 
			is_latest_bill_item = false,
			updated_at = now()
		WHERE 
			student_product_id = $1;
		`
	stmt = fmt.Sprintf(
		stmt,
		billItem.TableName(),
	)
	commandTag, err = db.Exec(ctx, stmt, studentProductID)
	if err != nil {
		err = fmt.Errorf("err set none latest bill item: %w", err)
		return
	}
	if commandTag.RowsAffected() == 0 {
		err = fmt.Errorf("err set none latest bill item: %d RowsAffected", commandTag.RowsAffected())
	}
	return
}

func (r *BillItemRepo) GetLatestBillItemByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (billItem entities.BillItem, err error) {
	fieldNames, fieldValues := (&billItem).FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM "%s" 
				WHERE student_product_id = $1 AND is_latest_bill_item = true
				FOR NO KEY UPDATE
				`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)

	row := db.QueryRow(ctx, stmt, studentProductID)
	err = row.Scan(fieldValues...)
	return
}

func (r *BillItemRepo) GetBillItemByStudentProductIDAndPeriodID(ctx context.Context, db database.QueryExecer, studentProductID string, periodID string) (billItem entities.BillItem, err error) {
	fieldNames, fieldValues := (&billItem).FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM "%s" 
				WHERE student_product_id = $1 AND is_latest_bill_item = true AND billing_schedule_period_id = $2
				FOR NO KEY UPDATE
				`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)

	row := db.QueryRow(ctx, stmt, studentProductID, periodID)
	err = row.Scan(fieldValues...)
	return
}

func (r *BillItemRepo) UpdateReviewFlagByOrderID(ctx context.Context, db database.QueryExecer, orderID string, isReview bool) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "BillingItemRepo.UpdateBillItemReviewedFlag")
	defer span.End()

	e := &entities.BillItem{}

	stmt := fmt.Sprintf(
		`UPDATE public.%s SET is_reviewed = $1
		WHERE order_id = $2;`,
		e.TableName(),
	)

	cmdTag, err := db.Exec(ctx, stmt, isReview, orderID)
	if err != nil {
		return fmt.Errorf("err update bill item: %w", err)
	}

	if cmdTag.RowsAffected() < 1 {
		return fmt.Errorf("updating review flag for bill item by order id %v have %d RowsAffected", orderID, cmdTag.RowsAffected())
	}

	return
}

func (r *BillItemRepo) VoidBillItemByOrderID(ctx context.Context, db database.QueryExecer, orderID string, status string) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "BillingItemRepo.VoidBillItemByOrderID")
	defer span.End()

	e := &entities.BillItem{}

	stmt := fmt.Sprintf(
		`UPDATE public.%s SET billing_status = $1
		WHERE order_id = $2;`,
		e.TableName(),
	)

	_, err = db.Exec(ctx, stmt, status, orderID)
	if err != nil {
		return fmt.Errorf("err void bill item: %w", err)
	}

	return
}

func (r *BillItemRepo) UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID(
	ctx context.Context,
	db database.QueryExecer,
	billItemSequenceNumber int32,
	status string,
) (
	orderID string,
	err error,
) {
	ctx, span := interceptors.StartSpan(ctx, "BillingItemRepo.UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID")
	defer span.End()

	e := &entities.BillItem{}

	stmt := fmt.Sprintf(
		`UPDATE public.%s SET billing_status = $1
		WHERE bill_item_sequence_number = $2
		RETURNING order_id;`,
		e.TableName(),
	)

	row := db.QueryRow(ctx, stmt, status, billItemSequenceNumber)
	err = row.Scan(&orderID)
	if err != nil {
		return
	}
	return
}

func (r *BillItemRepo) GetRecurringBillItemsForScheduledGenerationOfNextBillItems(
	ctx context.Context,
	db database.QueryExecer,
) ([]*entities.BillItem, error) {
	billItem := entities.BillItem{}
	fieldNames, _ := billItem.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT DISTINCT ON (order_id, product_id) %s FROM "%s" WHERE billing_schedule_period_id IS NOT NULL
		AND is_latest_bill_item = true 
		ORDER  BY order_id, product_id, billing_date DESC NULLS LAST`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*entities.BillItem
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, billItem)
	}
	return result, nil
}

func (r *BillItemRepo) GetBillItemByOrderIDAndPaging(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	from int64,
	limit int64,
) (
	billItems []*entities.BillItem,
	err error,
) {
	var rows pgx.Rows
	table := entities.BillItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)

	rows, err = db.Query(ctx, stmt, orderID, limit, from)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	billItems = make([]*entities.BillItem, 0, limit)
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billItems = append(billItems, billItem)
	}
	return
}

func (r *BillItemRepo) CountBillItemByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (
	total int,
	err error,
) {
	table := entities.BillItem{}
	stmt := fmt.Sprintf(
		`SELECT bill_item_sequence_number FROM "%s" WHERE order_id = $1`,
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, orderID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		total++
	}
	return
}

func (r *BillItemRepo) GetBillItemByStudentIDAndLocationIDsPaging(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationIDs []string,
	from int64,
	limit int64,
) (
	billItems []*entities.BillItem,
	err error,
) {
	var rows pgx.Rows
	table := entities.BillItem{}
	fieldNames, _ := table.FieldMap()

	if len(locationIDs) == 0 {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE student_id = $1 
			ORDER BY billing_date DESC LIMIT $2 OFFSET $3
			`,
			strings.Join(fieldNames, ","),
			table.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, limit, from)
		if err != nil {
			return nil, err
		}
	} else {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE student_id = $1 AND
			location_id = ANY($2)
			ORDER BY billing_date DESC LIMIT $3 OFFSET $4
			`,
			strings.Join(fieldNames, ","),
			table.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, locationIDs, limit, from)
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()

	billItems = make([]*entities.BillItem, 0, limit)
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billItems = append(billItems, billItem)
	}
	return
}

func (r *BillItemRepo) CountBillItemByStudentIDAndLocationIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationIDs []string,
) (
	total int,
	err error,
) {
	table := entities.BillItem{}
	var rows pgx.Rows
	if len(locationIDs) == 0 {
		stmt := fmt.Sprintf(
			`SELECT bill_item_sequence_number FROM "%s" WHERE student_id = $1`,
			table.TableName(),
		)
		rows, err = db.Query(ctx, stmt, studentID)
		if err != nil {
			return
		}
	} else {
		stmt := fmt.Sprintf(
			`SELECT bill_item_sequence_number FROM "%s" WHERE student_id = $1 AND location_id = ANY($2)`,
			table.TableName(),
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

func (r *BillItemRepo) GetBillItemInfoByOrderIDAndUniqueByProductID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (
	billItems []*entities.BillItem,
	err error,
) {
	var rows pgx.Rows
	table := entities.BillItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT DISTINCT ON (product_id) %s FROM "%s" WHERE order_id = $1
		ORDER BY product_id
		`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)

	rows, err = db.Query(ctx, stmt, orderID)
	if err != nil {
		return
	}

	defer rows.Close()

	billItems = []*entities.BillItem{}
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = fmt.Errorf(constant.RowScanError, err)
			return
		}
		billItems = append(billItems, billItem)
	}
	return
}

func (r *BillItemRepo) GetAllFirstBillItemDistinctByOrderIDAndUniqueByProductID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (
	billItems []*entities.BillItem,
	err error,
) {
	var rows pgx.Rows
	table := entities.BillItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT DISTINCT ON (product_id) %s FROM "%s" WHERE order_id = $1
		ORDER BY product_id, billing_from
		`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)

	rows, err = db.Query(ctx, stmt, orderID)
	if err != nil {
		return
	}

	defer rows.Close()

	billItems = []*entities.BillItem{}
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = fmt.Errorf(constant.RowScanError, err)
			return
		}
		billItems = append(billItems, billItem)
	}
	return
}

func (r *BillItemRepo) GetLatestBillItemByStudentProductIDForStudentBilling(ctx context.Context, db database.QueryExecer, studentProductID string) (billItem entities.BillItem, err error) {
	fieldNames, fieldValues := (&billItem).FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM "%s" 
				WHERE student_product_id = $1 AND is_latest_bill_item = true
				ORDER BY billing_from
				LIMIT 1
				`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)

	row := db.QueryRow(ctx, stmt, studentProductID)
	err = row.Scan(fieldValues...)
	return
}

func (r *BillItemRepo) GetBillingItemsThatNeedToBeBilled(
	ctx context.Context, db database.QueryExecer) (
	billItems []*entities.BillItem, err error) {
	billItem := entities.BillItem{}
	fieldNames, _ := billItem.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE billing_date::DATE <= NOW()::DATE
		AND billing_status = 'BILLING_STATUS_PENDING'
		ORDER BY student_product_id, billing_from DESC`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billItems = append(billItems, billItem)
	}
	return billItems, nil
}

func (r *BillItemRepo) GetExportStudentBilling(ctx context.Context, db database.QueryExecer, locationIDs []string) (billItems []*entities.BillItem, studentIDs []string, err error) {
	ctx, span := interceptors.StartSpan(ctx, "BillingItemRepo.GetExportStudentBilling")
	var rows pgx.Rows
	defer span.End()
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	table := entities.BillItem{}
	fieldNames, _ := table.FieldMap()

	if len(locationIDs) == 0 {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE resource_path = $1`,
			strings.Join(fieldNames, ","),
			table.TableName(),
		)

		rows, err = db.Query(ctx, stmt, resourcePath)
		if err != nil {
			return nil, nil, err
		}
	} else {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE resource_path = $1 AND location_id = ANY($2)`,
			strings.Join(fieldNames, ","),
			table.TableName(),
		)

		rows, err = db.Query(ctx, stmt, resourcePath, locationIDs)
		if err != nil {
			return nil, nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, nil, fmt.Errorf(constant.RowScanError, err)
		}
		billItems = append(billItems, billItem)
		studentIDs = append(studentIDs, billItem.StudentID.String)
	}
	return
}

func (r *BillItemRepo) GetByOrderIDAndProductIDs(ctx context.Context, db database.QueryExecer, orderID string, productIDs []string) ([]entities.BillItem, error) {
	var billItems []entities.BillItem
	billItemFieldNames, _ := (&entities.BillItem{}).FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = ANY($1) AND order_id = $2
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		(&entities.BillItem{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt, productIDs, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billItems = append(billItems, *billItem)
	}
	return billItems, nil
}

func (r *BillItemRepo) GetPresentAndFutureBillItemsByStudentProductIDs(ctx context.Context, db database.QueryExecer, studentProductIDs []string, studentID string) ([]*entities.BillItem, error) {
	billItem := entities.BillItem{}
	fieldNames, _ := billItem.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE student_product_id = ANY($1) AND
		(
            (billing_to IS NULL AND billing_from IS NULL) OR
            (NOW() BETWEEN billing_from AND billing_to) OR 
			(billing_from > NOW())
        )
		AND is_latest_bill_item = true 
		AND student_id = $2
		ORDER BY student_product_id, billing_from`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)

	rows, err := db.Query(ctx, stmt, studentProductIDs, studentID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*entities.BillItem
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, billItem)
	}
	return result, nil
}

func (r *BillItemRepo) GetPastBillItemsByStudentProductIDs(ctx context.Context, db database.QueryExecer, studentProductIDs []string, studentID string) ([]*entities.BillItem, error) {
	billItem := entities.BillItem{}
	fieldNames, _ := billItem.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE student_product_id = ANY($1)
		AND billing_to < NOW()
		AND is_latest_bill_item = true 
		AND student_id = $2
		ORDER BY student_product_id, billing_from DESC`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)

	rows, err := db.Query(ctx, stmt, studentProductIDs, studentID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*entities.BillItem
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, billItem)
	}
	return result, nil
}

func (r *BillItemRepo) GetUpcomingBillingByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string, studentID string) (*entities.BillItem, error) {
	billItem := &entities.BillItem{}
	fieldNames, fieldValues := billItem.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE student_product_id = $1 
		AND is_latest_bill_item = true
		AND student_id = $2
		ORDER BY billing_from DESC LIMIT 1`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentProductID, studentID)
	err := row.Scan(fieldValues...)
	if err != nil {
		return nil, err
	}

	return billItem, nil
}

func (r *BillItemRepo) GetPresentBillingByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string, studentID string) ([]*entities.BillItem, error) {
	billItem := &entities.BillItem{}
	fieldNames, _ := billItem.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE student_product_id = $1 
		AND (
            (billing_to IS NULL AND billing_from IS NULL) OR
            (NOW() BETWEEN billing_from AND billing_to)
        )
		AND is_latest_bill_item = true
		AND student_id = $2
		ORDER BY student_product_id, billing_from`,
		strings.Join(fieldNames, ","),
		billItem.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentProductID, studentID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*entities.BillItem
	for rows.Next() {
		billItem := new(entities.BillItem)
		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, billItem)
	}
	return result, nil
}

func (r *BillItemRepo) GetBillItemsByOrderIDAndProductID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	productID string,
) (billItems []entities.BillItem, err error) {
	e := entities.BillItem{}
	fieldNames, _ := e.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM "%s" 
				WHERE
				order_id = '%s' AND product_id = '%s'
				ORDER BY billing_from DESC;`,
		strings.Join(fieldNames, ","),
		e.TableName(),
		orderID,
		productID,
	)

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var billItem entities.BillItem
		_, fieldValues := billItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billItems = append(billItems, billItem)
	}
	return
}
