package invoicemgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) adminAddsPaymentToInvoice(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &invoice_pb.AddInvoicePaymentRequest{
		InvoiceId: stepState.InvoiceID,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setsPaymentMethodToInAddPaymentRequest(ctx context.Context, paymentMethodStr string) (context.Context, error) {
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

func (s *suite) setsDueDateToAndExpiryDateToInAddPaymentRequest(ctx context.Context, dueDate, expiryDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.AddInvoicePaymentRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.AddInvoicePaymentRequest got %T", request)
	}

	dueDateTimestamp := getFormattedTimestampDate(dueDate)
	expiryDateTimestamp := getFormattedTimestampDate(expiryDate)

	request.DueDate = dueDateTimestamp
	request.ExpiryDate = expiryDateTimestamp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setsAmountSameWithInvoiceOutstandingBalance(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.AddInvoicePaymentRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.AddInvoicePaymentRequest got %T", request)
	}

	invoice, err := s.getInvoiceFromStepState(ctx)
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

func (s *suite) adminSubmitsTheAddPaymentFormWithRemarks(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, ok := stepState.Request.(*invoice_pb.AddInvoicePaymentRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the request should be type *invoice_pb.AddInvoicePaymentRequest got %T", request)
	}

	request.Remarks = remarks

	stepState.Response, stepState.ResponseErr = invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).AddInvoicePayment(contextWithToken(ctx), request)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentBankAccountIsNotVerified(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	bankAccount := &entities.BankAccount{
		StudentID:  database.Text(stepState.StudentID),
		IsVerified: database.Bool(false),
	}

	_, err := database.UpdateFields(ctx, bankAccount, s.InvoiceMgmtDB.Exec, "student_id", []string{"is_verified"})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bank account of student err: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentHasPaymentAndBankAccountDetail(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.createStudentBankAccount(ctx, []string{stepState.StudentID}, stepState.BankBranchID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
