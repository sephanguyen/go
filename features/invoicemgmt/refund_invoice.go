package invoicemgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) adminRefundsAnInvoice(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &invoice_pb.RefundInvoiceRequest{
		InvoiceId: stepState.InvoiceID,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setsRefundMethodInRefundMethodRequest(ctx context.Context, refundMethodStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.RefundInvoiceRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.RefundInvoiceRequest got %T", request)
	}

	refundMethod, err := getRefundMethodFromStr(refundMethodStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	request.RefundMethod = refundMethod

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setsAmountSameWithInvoiceOutstandingBalanceInRefundMethodRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.RefundInvoiceRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.RefundInvoiceRequest got %T", request)
	}

	invoice, err := s.getInvoiceByInvoiceID(ctx, stepState.InvoiceID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	request.Amount = exactOutstandingBalance

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminSubmitsTheRefundInvoiceFormWithRemarks(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.RefundInvoiceRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.RefundInvoiceRequest got %T", request)
	}

	request.Remarks = remarks

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).RefundInvoice(contextWithToken(ctx), request)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setsAmountToInRefundInvoiceRequest(ctx context.Context, amount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.RefundInvoiceRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.RefundInvoiceRequest got %T", request)
	}

	request.Amount = float64(amount)

	return StepStateToContext(ctx, stepState), nil
}
