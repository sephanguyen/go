package invoicemgmt

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
)

func (s *suite) adminCancelTheBulkPayment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.BulkCancelPaymentRequest{
		BulkPaymentId: stepState.BulkPaymentID,
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).BulkCancelPayment(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bulkPaymentRecordStatusIsUpdatedTo(ctx context.Context, expectedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	bulkPayment := &entities.BulkPayment{}
	query := fmt.Sprintf("SELECT bulk_payment_id, bulk_payment_status FROM %s WHERE bulk_payment_id = $1 and resource_path = $2", bulkPayment.TableName())

	err := database.Select(ctx, s.InvoiceMgmtDBTrace, query, stepState.BulkPaymentID, stepState.ResourcePath).ScanOne(bulkPayment)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if bulkPayment.BulkPaymentStatus.String != expectedStatus {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting bulk payment status to be %v got %v", expectedStatus, bulkPayment.BulkPaymentStatus.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eachPaymentsHasPaymentStatus(ctx context.Context, expectedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, paymentID := range stepState.PaymentIDs {
		ctx, err := assertPaymentStatusByPaymentID(ctx, s.InvoiceMgmtDBTrace, paymentID, expectedStatus)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisBulkPaymentHasStatus(ctx context.Context, updateStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `UPDATE bulk_payment SET bulk_payment_status = $1 WHERE resource_path = $2 AND bulk_payment_id = $3`

	if _, err := s.InvoiceMgmtDBTrace.Exec(ctx, stmt, updateStatus, s.ResourcePath, stepState.BulkPaymentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) onlyPendingPaymentsWereUpdatedToPaymentStatus(ctx context.Context, expectedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Check if pending payments are updated to the status
	pendingPaymentIDs := stepState.PaymentStatusIDsMap[invoice_pb.PaymentStatus_PAYMENT_PENDING.String()]
	for _, paymentID := range pendingPaymentIDs {
		ctx, err := assertPaymentStatusByPaymentID(ctx, s.InvoiceMgmtDBTrace, paymentID, expectedStatus)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	// Check if non-pending payments are not updated
	for _, paymentStatus := range []string{invoice_pb.PaymentStatus_PAYMENT_FAILED.String(), invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(), invoice_pb.PaymentStatus_PAYMENT_REFUNDED.String()} {
		paymentIDs := stepState.PaymentStatusIDsMap[paymentStatus]
		for _, paymentID := range paymentIDs {
			ctx, err := assertPaymentStatusByPaymentID(ctx, s.InvoiceMgmtDBTrace, paymentID, paymentStatus)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) onlyPendingPaymentsInvoiceAreRecordedWithActionLogType(ctx context.Context, expectedActionLogType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// check if pending payments generated action log
	pendingPaymentIDs := stepState.PaymentStatusIDsMap[invoice_pb.PaymentStatus_PAYMENT_PENDING.String()]
	for _, paymentID := range pendingPaymentIDs {
		payment := &entities.Payment{}
		query := fmt.Sprintf("SELECT payment_id, payment_status, invoice_id FROM %s WHERE payment_id = $1 and resource_path = $2", payment.TableName())

		err := database.Select(ctx, s.InvoiceMgmtDBTrace, query, paymentID, stepState.ResourcePath).ScanOne(payment)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		actionLog, err := s.getLatestInvoiceActionLogByInvoiceID(ctx, payment.InvoiceID.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("actionLogRecordIsRecordedWithActionLogType error: %v", err.Error())
		}
		if expectedActionLogType != actionLog.Action.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v action log type but got %v for invoice_id %v", expectedActionLogType, actionLog.Action.String, s.StepState.InvoiceID)
		}
	}

	// Check if non-pending payments are not generated with action log
	for _, paymentStatus := range []string{invoice_pb.PaymentStatus_PAYMENT_FAILED.String(), invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(), invoice_pb.PaymentStatus_PAYMENT_REFUNDED.String()} {
		paymentIDs := stepState.PaymentStatusIDsMap[paymentStatus]
		for _, paymentID := range paymentIDs {
			payment := &entities.Payment{}
			query := fmt.Sprintf("SELECT payment_id, payment_status, invoice_id FROM %s WHERE payment_id = $1 and resource_path = $2", payment.TableName())

			err := database.Select(ctx, s.InvoiceMgmtDBTrace, query, paymentID, stepState.ResourcePath).ScanOne(payment)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			_, err = s.getLatestInvoiceActionLogByInvoiceID(ctx, payment.InvoiceID.String)
			if !errors.Is(err, pgx.ErrNoRows) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expecting no action log for invoice %v got one", payment.InvoiceID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noInvoiceActionLogWithActionLogTypeRecordedForEachInvoice(ctx context.Context, actionLogType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, invoiceID := range stepState.InvoiceIDs {
		e := &entities.InvoiceActionLog{}
		fields, _ := e.FieldMap()

		query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_id = $1 and action = $2 ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), e.TableName())
		err := database.Select(ctx, s.InvoiceMgmtDBTrace, query, invoiceID, actionLogType).ScanOne(e)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting no action log %s recorded got 1", actionLogType)
	}

	return StepStateToContext(ctx, stepState), nil
}

func assertPaymentStatusByPaymentID(ctx context.Context, db database.QueryExecer, paymentID string, expectedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	payment := &entities.Payment{}
	query := fmt.Sprintf("SELECT payment_id, payment_status, invoice_id FROM %s WHERE payment_id = $1 and resource_path = $2", payment.TableName())

	err := database.Select(ctx, db, query, paymentID, stepState.ResourcePath).ScanOne(payment)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if payment.PaymentStatus.String != expectedStatus {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting payment status to be %v got %v", expectedStatus, payment.PaymentStatus.String)
	}

	return StepStateToContext(ctx, stepState), nil
}
