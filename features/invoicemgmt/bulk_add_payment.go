package invoicemgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/pkg/errors"
)

func (s *suite) theseInvoicesHasType(ctx context.Context, invoiceType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIds {
		stmt := `UPDATE invoice SET type = $1 WHERE resource_path = $2 AND student_id = $3`

		if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, invoiceType, s.ResourcePath, studentID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("err: %v student: %v on updating invoice type to %v", err, studentID, invoiceType))
		}
	}

	stepState.InvoiceTypes = append(stepState.InvoiceTypes, invoiceType)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bulkAddPaymentForTheseInvoicesWithPaymentMethod(ctx context.Context, signedInUser, paymentMethodStr, dueDate, expiryDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, signedInUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	dueDateTimestamp := getFormattedTimestampDate(dueDate)
	expiryDateTimestamp := getFormattedTimestampDate(expiryDate)

	paymentMethod := invoice_pb.BulkPaymentMethod(invoice_pb.BulkPaymentMethod_value[paymentMethodStr])

	latestPaymentStatuses := make([]invoice_pb.PaymentStatus, 0)
	invoiceTypes := make([]invoice_pb.InvoiceType, 0)

	for _, latestPaymentStatus := range stepState.LatestPaymentStatuses {
		latestPaymentStatuses = append(latestPaymentStatuses, invoice_pb.PaymentStatus(invoice_pb.PaymentStatus_value[latestPaymentStatus]))
	}

	for _, invoiceType := range stepState.InvoiceTypes {
		invoiceTypes = append(invoiceTypes, invoice_pb.InvoiceType(invoice_pb.InvoiceType_value[invoiceType]))
	}

	req := &invoice_pb.BulkAddPaymentRequest{
		InvoiceIds: s.StepState.InvoiceIDs,

		BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
			BulkPaymentMethod:   paymentMethod,
			LatestPaymentStatus: latestPaymentStatuses,
			InvoiceType:         invoiceTypes,
		},
		ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
			DueDate:    dueDateTimestamp,
			ExpiryDate: expiryDateTimestamp,
		},
		DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
			DueDate:    nil,
			ExpiryDate: nil,
		},
	}

	if paymentMethod == invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT {
		req.DirectDebitDates.DueDate = dueDateTimestamp
		req.DirectDebitDates.ExpiryDate = expiryDateTimestamp
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).BulkAddPayment(contextWithToken(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreNoPaymentsForTheseInvoices(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.LatestPaymentStatuses = append(stepState.LatestPaymentStatuses, invoice_pb.PaymentStatus_PAYMENT_NONE.String())

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bulkPaymentRecordIsCreatedSuccessfullyWithPaymentMethod(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	payment := &entities.Payment{}

	query := fmt.Sprintf(`SELECT bulk_payment_id FROM %s WHERE invoice_id = ANY($1) AND resource_path = $2 AND payment_status = $3 GROUP BY bulk_payment_id`, payment.TableName())

	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, query, stepState.InvoiceIDs, stepState.ResourcePath, invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting payments err: %v", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Err() err: %v", err)
	}

	bulkPaymentIDs := make([]string, 0)
	for rows.Next() {
		var bulkPaymentID string
		if err := rows.Scan(&bulkPaymentID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("row.Scan() err: %v", err)
		}
		bulkPaymentIDs = append(bulkPaymentIDs, bulkPaymentID)
	}

	if len(bulkPaymentIDs) != 1 {
		return StepStateToContext(ctx, stepState), errors.New("error expected payments should be group by one bulk payment record")
	}

	var bulkPaymentCount int

	stmt := `SELECT COUNT(*) FROM bulk_payment WHERE bulk_payment_id = $1 AND payment_method = $2 AND resource_path = $3`

	err = s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, bulkPaymentIDs[0], paymentMethod, stepState.ResourcePath).Scan(&bulkPaymentCount)

	if err != nil {
		return nil, fmt.Errorf("error on retrieving bulk payment record: %v", err)
	}

	if bulkPaymentCount != 1 {
		return StepStateToContext(ctx, stepState), errors.New("error on retrieving bulk payment record")
	}

	return StepStateToContext(ctx, stepState), nil
}
