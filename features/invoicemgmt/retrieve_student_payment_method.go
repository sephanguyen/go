package invoicemgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) anExistingStudentWithDefaultPaymentMethod(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.anExistingStudentWithBillingOrBankAccountInfo(ctx, "billing address and bank account")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// update to given payment method
	ctx, err = s.studentPaymentDetailHasPaymentMethod(ctx, paymentMethod)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentPaymentMethodMap[stepState.StudentID] = paymentMethod

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theRetrieveStudentPaymentMethodEndpointIsCalledForThisStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.RetrieveStudentPaymentMethodRequest{
		StudentId: stepState.StudentID,
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).RetrieveStudentPaymentMethod(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentMethodForThisStudentIsRetrieveSuccessfully(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp, ok := stepState.Response.(*invoice_pb.RetrieveStudentPaymentMethodResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("RetrieveStudentPaymentMethodResponse is nil")
	}

	if resp.PaymentMethod != constant.PaymentMethodsConvertToEnums[paymentMethod] || resp.StudentId != stepState.StudentID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on retrieving student payment method expected student %v with payment method %v but have: student %v with payment method %v", stepState.StudentID, paymentMethod, resp.StudentId, resp.PaymentMethod)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anExistingStudentWithNoStudentPaymentMethod(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.anExistingStudentWithBillingOrBankAccountInfo(ctx, "no billing and bank account info")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentPaymentMethodMap[stepState.StudentID] = invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT.String()

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) emptyPaymentMethodForThisStudentIsretrieveSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp, ok := stepState.Response.(*invoice_pb.RetrieveStudentPaymentMethodResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("RetrieveStudentPaymentMethodResponse is nil")
	}

	if resp.PaymentMethod != invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT || resp.StudentId != stepState.StudentID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on retrieving student payment method expected student %v with  empty payment method but have: student %v with payment method %v", stepState.StudentID, resp.StudentId, resp.PaymentMethod)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aNonExistingStudentRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentID = idutil.ULIDNow()
	stepState.StudentIds = append(stepState.StudentIds, stepState.StudentID)

	return StepStateToContext(ctx, stepState), nil
}
