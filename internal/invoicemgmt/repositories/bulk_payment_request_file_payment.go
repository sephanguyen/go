package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"go.uber.org/multierr"
)

type BulkPaymentRequestFilePaymentRepo struct {
}

func (r *BulkPaymentRequestFilePaymentRepo) FindByPaymentID(ctx context.Context, db database.QueryExecer, paymentID string) (*entities.BulkPaymentRequestFilePayment, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestFilePaymentRepo.FindByPaymentID")
	defer span.End()

	e := &entities.BulkPaymentRequestFilePayment{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE payment_id = $1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, paymentID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *BulkPaymentRequestFilePaymentRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFilePayment) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestFilePaymentRepo.Create")
	defer span.End()

	id := idutil.ULIDNow()

	now := time.Now()
	if err := multierr.Combine(
		e.BulkPaymentRequestFilePaymentID.Set(id),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return "", fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return "", fmt.Errorf("err insert BulkPaymentRequestFilePaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return "", fmt.Errorf("err insert BulkPaymentRequestFilePaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return id, nil
}

func (r *BulkPaymentRequestFilePaymentRepo) FindPaymentInvoiceByRequestFileID(ctx context.Context, db database.QueryExecer, id string) ([]*entities.FilePaymentInvoiceMap, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID")
	defer span.End()

	query := `
		SELECT
			fp.bulk_payment_request_file_payment_id,
			fp.bulk_payment_request_file_id,
			fp.payment_id,
			p.payment_id,
			p.invoice_id,
			p.payment_method,
			p.payment_due_date,
			p.payment_expiry_date,
			p.payment_status,
			p.payment_sequence_number,
			p.student_id,
			p.is_exported,
			i.invoice_id,
			i.type,
			i.status,
			i.student_id,
			i.sub_total,
			i.total,
			i.is_exported
		FROM bulk_payment_request_file_payment fp
		INNER JOIN payment p
			ON fp.payment_id = p.payment_id
		INNER JOIN invoice i
			ON p.invoice_id = i.invoice_id
		WHERE fp.bulk_payment_request_file_id = $1
	`

	rows, err := db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filePaymentInvoices := []*entities.FilePaymentInvoiceMap{}
	for rows.Next() {
		filePaymentInvoice := new(entities.FilePaymentInvoiceMap)
		filePaymentInvoice.Payment = &entities.Payment{}
		filePaymentInvoice.Invoice = &entities.Invoice{}
		filePaymentInvoice.BulkPaymentRequestFilePayment = &entities.BulkPaymentRequestFilePayment{}

		database.AllNullEntity(filePaymentInvoice.Invoice)
		database.AllNullEntity(filePaymentInvoice.Payment)
		database.AllNullEntity(filePaymentInvoice.BulkPaymentRequestFilePayment)

		err := rows.Scan(
			&filePaymentInvoice.BulkPaymentRequestFilePayment.BulkPaymentRequestFilePaymentID,
			&filePaymentInvoice.BulkPaymentRequestFilePayment.BulkPaymentRequestFileID,
			&filePaymentInvoice.BulkPaymentRequestFilePayment.PaymentID,
			&filePaymentInvoice.Payment.PaymentID,
			&filePaymentInvoice.Payment.InvoiceID,
			&filePaymentInvoice.Payment.PaymentMethod,
			&filePaymentInvoice.Payment.PaymentDueDate,
			&filePaymentInvoice.Payment.PaymentExpiryDate,
			&filePaymentInvoice.Payment.PaymentStatus,
			&filePaymentInvoice.Payment.PaymentSequenceNumber,
			&filePaymentInvoice.Payment.StudentID,
			&filePaymentInvoice.Payment.IsExported,
			&filePaymentInvoice.Invoice.InvoiceID,
			&filePaymentInvoice.Invoice.Type,
			&filePaymentInvoice.Invoice.Status,
			&filePaymentInvoice.Invoice.StudentID,
			&filePaymentInvoice.Invoice.SubTotal,
			&filePaymentInvoice.Invoice.Total,
			&filePaymentInvoice.Invoice.IsExported,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		filePaymentInvoices = append(filePaymentInvoices, filePaymentInvoice)
	}

	return filePaymentInvoices, nil
}
