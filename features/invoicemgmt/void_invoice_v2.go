package invoicemgmt

import (
	"context"
	"fmt"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	pgx "github.com/jackc/pgx/v4"
)

func (s *suite) latestPaymentRecordHasPaymentStatusAndAmountZero(ctx context.Context, paymentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	payment, err := s.getLatestPaymentHistoryRecordByInvoiceID(ctx, stepState.InvoiceID)

	// It's possible to have no payment records
	if err != nil && err != pgx.ErrNoRows {
		return StepStateToContext(ctx, stepState), fmt.Errorf("latestPaymentRecordHasPaymentStatus error: %v", err.Error())
	}

	// "none" expects no payment record
	if payment != nil && paymentStatus == "none" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected no payment record but got one for invoice_id %v", s.StepState.InvoiceID)
	}

	// if status is not "none", payment record is expected
	if payment == nil && paymentStatus != "none" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment record but got none for invoice_id %v", s.StepState.InvoiceID)
	}

	var expectedPaymentStatus invoice_pb.PaymentStatus

	switch paymentStatus {
	case "SUCCESSFUL":
		expectedPaymentStatus = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL
	case "FAILED":
		expectedPaymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
	default:
		expectedPaymentStatus = invoice_pb.PaymentStatus_PAYMENT_PENDING
	}

	if paymentStatus != "none" && payment.PaymentStatus.String != expectedPaymentStatus.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v payment status but got %v for invoice_id %v", expectedPaymentStatus.String(), payment.PaymentStatus.String, s.StepState.InvoiceID)
	}

	// payment amount should be zero
	if payment != nil && payment.Amount.Int.Int64() != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment amount to be zero but got %d", payment.Amount.Int.Int64())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminVoidsAnInvoiceWithRemarksUsingV2Endpoint(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.VoidInvoiceRequestV2{
		InvoiceId: stepState.InvoiceID,
	}

	if remarks == "any" {
		req.Remarks = remarks
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).VoidInvoiceV2(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
