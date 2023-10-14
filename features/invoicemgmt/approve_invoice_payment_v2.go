package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) thereIsAnExistingInvoiceWithStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.thereArePreexistingNumberOfExistingInvoicesWithStatus(ctx, "1", status)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminAlreadyRequestedPaymentWithAmountSameOnInvoiceOutstandingBalance(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.AddInvoicePaymentRequest{
		InvoiceId:  stepState.InvoiceID,
		DueDate:    getFormattedTimestampDate("TODAY"),
		ExpiryDate: getFormattedTimestampDate("TODAY+1"),
	}

	invoice, err := s.getInvoiceFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req.Amount = exactOutstandingBalance

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisPaymentHasPaymentMethod(ctx context.Context, paymentMethodStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.AddInvoicePaymentRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.AddInvoicePaymentRequest got %T", request)
	}

	paymentMethod, err := getPaymentMethodFromStr(paymentMethodStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	request.PaymentMethod = paymentMethod

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addedTheRequestedPaymentOnTheInvoice(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.AddInvoicePaymentRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.AddInvoicePaymentRequest got %T", request)
	}

	_, err := invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).AddInvoicePayment(contextWithToken(ctx), request)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminSetsTheApprovePaymentFormWithPaymentDate(ctx context.Context, paymentDateStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := &invoice_pb.ApproveInvoicePaymentV2Request{
		InvoiceId:   stepState.InvoiceID,
		PaymentDate: getFormattedTimestampDate(paymentDateStr),
	}

	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminSubmitsTheApprovePaymentFormWithRemarksUsingV2Endpoint(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.ApproveInvoicePaymentV2Request)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.ApproveInvoicePaymentV2Request got %T", request)
	}

	if strings.TrimSpace(remarks) == "any" {
		request.Remarks = remarks
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).ApproveInvoicePaymentV2(contextWithToken(ctx), request)

	// sets the current invoice to nil to refetch the updated amount paid
	stepState.CurrentInvoice = nil

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceAmountPaidIsEqualToPaymentAmount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	invoice, err := s.getInvoiceFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	payment, err := s.getLatestInvoicePaymentFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	invoiceAmountPaid, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.AmountPaid, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	paymentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(payment.Amount, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if invoiceAmountPaid != paymentAmount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invoice amount paid: %v is not equal to payment amount: %v", invoice.AmountPaid, payment.Amount)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) invoiceHasZeroOutstandingBalance(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	invoice, err := s.getInvoiceFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if invoice.OutstandingBalance.Int.Int64() != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invoice has outstanding balance: %v instead of zero", invoice.OutstandingBalance)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) latestPaymentRecordHasReceiptDateToday(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	payment, err := s.getLatestInvoicePaymentFromStepState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("latestPaymentRecordHasPaymentStatus error: %v", err.Error())
	}

	err = compareReceiptDateWhenBulkValidateProcess(payment, time.Now())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
