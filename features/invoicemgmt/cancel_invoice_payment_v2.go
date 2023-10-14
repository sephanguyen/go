package invoicemgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) adminCancelsAnInvoiceWithRemarksUsingV2Endpoint(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &invoice_pb.CancelInvoicePaymentV2Request{
		InvoiceId: stepState.InvoiceID,
	}
	if strings.TrimSpace(remarks) != "" {
		req.Remarks = remarks
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).CancelInvoicePaymentV2(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsPaymentHistoryWithPaymentMethod(ctx context.Context, paymentStatus, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var status invoice_pb.PaymentStatus

	switch paymentStatus {
	case "FAILED":
		status = invoice_pb.PaymentStatus_PAYMENT_FAILED
	case "SUCCESSFUL":
		status = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL
	default:
		status = invoice_pb.PaymentStatus_PAYMENT_PENDING
	}

	getPaymentMethod := invoice_pb.PaymentMethod(invoice_pb.PaymentMethod_value[paymentMethod])

	ctx, err := s.createPayment(ctx, getPaymentMethod, status.String(), "", false)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("thereIsPaymentHistory error: %v", err.Error())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisPaymentHasExportedStatus(ctx context.Context, exportedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var isExported bool

	if exportedStatus == "TRUE" {
		isExported = true
	}
	paymentRepo := &repositories.PaymentRepo{}

	err := paymentRepo.UpdateIsExportedByPaymentIDs(ctx, s.InvoiceMgmtPostgresDBTrace, []string{stepState.PaymentID}, isExported)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) belongsInBulkWithOtherPaymentsWithStatus(ctx context.Context, otherPaymentCount int, paymentStatus, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.PaymentIDs) == 0 {
		stepState.PaymentIDs = append(stepState.PaymentIDs, stepState.PaymentID)
	}

	ctx, err := s.thereAreExistingPayments(ctx, otherPaymentCount, paymentStatus, paymentMethod)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.thesePaymentsBelongsToABulkPayment(ctx, len(stepState.PaymentIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bulkPaymentRecordHasStatus(ctx context.Context, bulkPaymentStatusStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// check bulk payment status
	var bulkPaymentStatusDB string

	stmt := `SELECT bulk_payment_status FROM bulk_payment WHERE bulk_payment_id = $1 AND resource_path = $2`
	row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.BulkPaymentID, stepState.ResourcePath)
	err := row.Scan(&bulkPaymentStatusDB)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error selecting bulk payments with bulk payment id: %v", stepState.BulkPaymentID)
	}

	if bulkPaymentStatusDB != bulkPaymentStatusStr {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error bulk payment status expected: %v got: %v on bulk payment id: %v", bulkPaymentStatusStr, bulkPaymentStatusDB, stepState.BulkPaymentID)
	}

	return StepStateToContext(ctx, stepState), nil
}
