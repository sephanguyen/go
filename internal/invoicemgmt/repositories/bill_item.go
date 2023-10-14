package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type BillItemRepo struct {
}

func (r *BillItemRepo) FindByID(ctx context.Context, db database.QueryExecer, billItemID int32) (*entities.BillItem, error) {
	ctx, span := interceptors.StartSpan(ctx, "BillItemRepo.FindByID")
	defer span.End()

	billItem := &entities.BillItem{}
	fields, _ := billItem.FieldMap()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE bill_item_sequence_number = $1 AND resource_path = $2", strings.Join(fields, ","), billItem.TableName())

	err := database.Select(ctx, db, query, billItemID, resourcePath).ScanOne(billItem)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Errorf("db.QueryRowEx %v", billItemID).Error())
	}
	return billItem, nil
}

func (r *BillItemRepo) FindByStatuses(ctx context.Context, db database.QueryExecer, billItemStatuses []string) ([]*entities.BillItem, error) {
	ctx, span := interceptors.StartSpan(ctx, "BillItemRepo.FindByStatuses")
	defer span.End()

	e := &entities.BillItem{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE billing_status = ANY($1)", strings.Join(fields, ","), e.TableName())

	var bs pgtype.TextArray
	_ = bs.Set(billItemStatuses)

	rows, err := db.Query(ctx, stmt, bs)
	if err != nil {
		return nil, err
	}

	billItems := []*entities.BillItem{}
	defer rows.Close()
	for rows.Next() {
		billItem := new(entities.BillItem)
		database.AllNullEntity(billItem)

		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, billItem)
	}

	return billItems, nil
}

func (r *BillItemRepo) FindByOrderID(ctx context.Context, db database.QueryExecer, orderID string) ([]*entities.BillItem, error) {
	ctx, span := interceptors.StartSpan(ctx, "BillItemRepo.FindByOrderID")
	defer span.End()

	e := &entities.BillItem{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE order_id = $1", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}

	billItems := []*entities.BillItem{}
	defer rows.Close()
	for rows.Next() {
		billItem := new(entities.BillItem)
		database.AllNullEntity(billItem)

		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, billItem)
	}

	return billItems, nil
}

func (r *BillItemRepo) RetrieveBillItemsByInvoiceReferenceNum(ctx context.Context, db database.QueryExecer, referenceID string) ([]*entities.BillItem, error) {
	ctx, span := interceptors.StartSpan(ctx, "BillItemRepo.RetrieveBillItemsByInvoiceReferenceNum")
	defer span.End()

	e := &entities.BillItem{}
	stmt := fmt.Sprintf(`
		SELECT
			order_id,
			bill_item_sequence_number,
			final_price,
			adjustment_price,
			bill_type,
			billing_status,
			billing_date,
			reference
		FROM %s WHERE reference = $1 AND resource_path =  $2`,
		e.TableName(),
	)

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	rows, err := db.Query(ctx, stmt, &referenceID, &resourcePath)
	if err != nil {
		return nil, err
	}

	billItems := []*entities.BillItem{}
	defer rows.Close()
	for rows.Next() {
		billItem := new(entities.BillItem)
		database.AllNullEntity(billItem)

		err := rows.Scan(
			&billItem.OrderID,
			&billItem.BillItemSequenceNumber,
			&billItem.FinalPrice,
			&billItem.AdjustmentPrice,
			&billItem.BillType,
			&billItem.BillStatus,
			&billItem.BillDate,
			&billItem.Reference,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, billItem)
	}

	return billItems, nil
}

// used for data migration for finding invoice reference and summing up final price only
func (r *BillItemRepo) GetBillItemTotalByStudentAndReference(ctx context.Context, db database.QueryExecer, studentID, invoiceReferenceID string) (pgtype.Numeric, error) {
	var totalFinalPrice pgtype.Numeric
	e := &entities.BillItem{}

	query := fmt.Sprintf("SELECT SUM(final_price) AS total_final_price FROM %s WHERE student_id = $1 AND reference = $2", e.TableName())

	err := db.QueryRow(ctx, query, &studentID, &invoiceReferenceID).Scan(&totalFinalPrice)

	if err != nil {
		return totalFinalPrice, fmt.Errorf("err BillItemRepo GetBillItemTotalByStudentAndReference: %w", err)
	}

	return totalFinalPrice, nil
}

func (r *BillItemRepo) FindInvoiceBillItemMapByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string) ([]*entities.InvoiceBillItemMap, error) {
	ctx, span := interceptors.StartSpan(ctx, "BillItemRepo.FindByStatuses")
	defer span.End()

	stmt := `
		SELECT i.invoice_id, b.bill_item_sequence_number, b.billing_item_description, b.final_price, b.adjustment_price, b.bill_type
		FROM invoice i
		INNER JOIN invoice_bill_item ibi
			ON ibi.invoice_id = i.invoice_id
		INNER JOIN bill_item b
			ON b.bill_item_sequence_number = ibi.bill_item_sequence_number
				AND ibi.resource_path = b.resource_path
		WHERE i.invoice_id = ANY($1)
	`

	var ids pgtype.TextArray
	_ = ids.Set(invoiceIDs)

	rows, err := db.Query(ctx, stmt, ids)
	if err != nil {
		return nil, err
	}

	res := []*entities.InvoiceBillItemMap{}
	defer rows.Close()
	for rows.Next() {
		e := new(entities.InvoiceBillItemMap)

		err := rows.Scan(
			&e.InvoiceID,
			&e.BillItemSequenceNumber,
			&e.BillingItemDescription,
			&e.FinalPrice,
			&e.AdjustmentPrice,
			&e.BillType,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		res = append(res, e)
	}

	return res, nil
}
