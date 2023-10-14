package invoicemgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) adminApprovesPaymentWithRemarks(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.ApproveInvoicePaymentRequest{
		InvoiceId:   fmt.Sprintf("%v", stepState.InvoiceID),
		PaymentDate: timestamppb.Now(),
	}

	if remarks == "any" {
		req.Remarks = remarks
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).ApproveInvoicePayment(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminApprovesPaymentWithRemarksWithoutPaymentDate(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.ApproveInvoicePaymentRequest{
		InvoiceId: fmt.Sprintf("%v", stepState.InvoiceID),
	}

	if remarks == "any" {
		req.Remarks = remarks
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).ApproveInvoicePayment(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) actionLogRecordIsRecordedWithActionLogType(ctx context.Context, expectedActionLogType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	actionLog, err := s.getLatestActionLogByInvoiceIDAndAction(ctx, stepState.InvoiceID, expectedActionLogType)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("actionLogRecordIsRecordedWithActionLogType error: %v", err.Error())
	}

	if expectedActionLogType != actionLog.Action.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v action log type but got %v for invoice_id %v", expectedActionLogType, actionLog.Action.String, s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billItemHasFinalPriceValue(ctx context.Context, finalPriceValue float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.updateBillItemFinalPriceValueByStudentID(ctx, stepState.StudentID, finalPriceValue)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billItemHasAdjustmentPriceValue(ctx context.Context, adjustmentPriceValue float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.updateBillItemAdjustmentPriceValueByStudentID(ctx, stepState.StudentID, adjustmentPriceValue)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceOutstandingBalanceSetTo(ctx context.Context, outstandingBalance float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoice, err := s.getInvoiceFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("getInvoiceFromStepState err: %v", err)
	}

	exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if exactOutstandingBalance != outstandingBalance {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting outstanding_balance to be %v got %v", outstandingBalance, exactOutstandingBalance)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceAmountPaidSetTo(ctx context.Context, amountPaid float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoice, err := s.getInvoiceFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("getInvoiceFromStepState err: %v", err)
	}

	exactAmountPaid, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.AmountPaid, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if exactAmountPaid != amountPaid {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting amount_paid to be %v got %v", amountPaid, exactAmountPaid)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceAmountRefundedSetTo(ctx context.Context, amountRefunded float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoice, err := s.getInvoiceFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("getInvoiceFromStepState err: %v", err)
	}

	exactAmountRefunded, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.AmountRefunded, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if exactAmountRefunded != amountRefunded {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting amount_refunded to be %v got %v", amountRefunded, exactAmountRefunded)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getInvoiceFromStepState(ctx context.Context) (*entities.Invoice, error) {
	stepState := StepStateFromContext(ctx)

	currentInvoice := stepState.CurrentInvoice
	if currentInvoice != nil {
		return currentInvoice, nil
	}

	invoice, err := s.getInvoiceByInvoiceID(ctx, stepState.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoiceHasInvoiceStatus error: %v", err)
	}
	stepState.CurrentInvoice = invoice

	return stepState.CurrentInvoice, nil
}
