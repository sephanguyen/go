package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/try"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) loginsToBackofficeApp(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, user)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnExistingInvoiceWithInvoiceStatusWithBillItem(ctx context.Context, invoiceStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create an invoice
	var err error
	if ctx, err = s.insertInvoiceIntoInvoicemgmt(ctx, invoiceStatus); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billItemHasPreviousStatus(ctx context.Context, previousStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var status string

	switch previousStatus {
	case "billed":
		status = payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()
	default:
		status = payment_pb.BillingStatus_BILLING_STATUS_PENDING.String()
	}

	// This step assumes there's a bill item already created before this step
	// Insert invoice bill item record; associates invoice with bill item
	_, err := s.createInvoiceBillItem(ctx, status)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasBillingDateToday(ctx context.Context, comparison string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var billingDate time.Time

	switch comparison {
	case "before":
		billingDate = time.Now().UTC().AddDate(0, -5, 0)
	case "after":
		billingDate = time.Now().UTC().AddDate(0, 5, 0)
	case "same":
		billingDate = time.Now().UTC()
	}

	query := "UPDATE bill_item SET billing_date = $1 WHERE bill_item_sequence_number = $2"
	if _, err := s.InvoiceMgmtDB.Exec(ctx, query, billingDate, &stepState.BillItemSequenceNumber); err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsPaymentHistory(ctx context.Context, paymentHistory string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// If none, no payment history to be created for the invoice;
	// Otherwise, create a series of payment records
	if paymentHistory != "none" {
		paymentStatuses := strings.Split(paymentHistory, "-")

		for _, paymentStatus := range paymentStatuses {
			var status invoice_pb.PaymentStatus

			switch paymentStatus {
			case "FAILED":
				status = invoice_pb.PaymentStatus_PAYMENT_FAILED
			case "SUCCESSFUL":
				status = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL
			case "REFUNDED":
				status = invoice_pb.PaymentStatus_PAYMENT_REFUNDED
			default:
				status = invoice_pb.PaymentStatus_PAYMENT_PENDING
			}

			ctx, err := s.createPayment(ctx, invoice_pb.PaymentMethod_DIRECT_DEBIT, status.String(), "", false)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("thereIsPaymentHistory error: %v", err.Error())
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminVoidsAnInvoiceWithRemarks(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.VoidInvoiceRequest{
		InvoiceId: fmt.Sprintf("%v", stepState.InvoiceID),
	}

	if remarks == "any" {
		req.Remarks = remarks
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).VoidInvoice(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceHasInvoiceStatus(ctx context.Context, invoiceStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoice, err := s.getInvoiceByInvoiceID(ctx, stepState.InvoiceID)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invoiceHasInvoiceStatus error: %v", err)
	}

	if invoice.Status.String != invoiceStatus {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v invoice status but got %v for invoice_id %v", invoiceStatus, invoice.Status, s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billItemHasBillItemStatus(ctx context.Context, billItemStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var expectedBillItemStatus string
	switch billItemStatus {
	case "billed":
		expectedBillItemStatus = payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()
	case "invoiced":
		expectedBillItemStatus = payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()
	default:
		expectedBillItemStatus = payment_pb.BillingStatus_BILLING_STATUS_PENDING.String()
	}

	// Wait for Kafka to sync updated bill items with invoicemgmt db
	_, err := s.billItemExistsWithStatusInInvoicemgmtDatabase(ctx, expectedBillItemStatus)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("billItemExistsWithStatusInInvoicemgmtDatabase error: %v", err.Error())
	}

	billItem, err := s.getBillItemByInvoiceID(ctx, stepState.InvoiceID)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("billItemHasBillItemStatus error: %v", err.Error())
	}

	if billItem.BillStatus.String != expectedBillItemStatus {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v bill item status but got %v for invoice_id %v", expectedBillItemStatus, billItem.BillStatus.String, s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billItemExistsWithStatusInInvoicemgmtDatabase(ctx context.Context, billingStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stmt := `SELECT bill_item_sequence_number FROM bill_item WHERE student_id = $1 and billing_status = $2`

	if err := try.Do(func(attempt int) (bool, error) {
		var billItemSequenceNumber int32
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, s.StepState.StudentID, billingStatus)
		err := row.Scan(&billItemSequenceNumber)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if billItemSequenceNumber != 0 {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("bill item sequence number with status %v is not found in invoicemgmt", billingStatus)
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) actionLogRecordIsRecorded(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoiceActionLog, err := s.getLatestInvoiceActionLogByInvoiceID(ctx, stepState.InvoiceID)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("actionLogRecordIsRecorded error: %v", err.Error())
	}

	if invoiceActionLog == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected non-nil invoiceActionLog but got nil for invoice_id %v", s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}
