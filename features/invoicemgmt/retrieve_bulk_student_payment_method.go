package invoicemgmt

import (
	"context"
	"fmt"
	"strings"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) thereAreExistingStudentsWithDefaultPaymentMethod(ctx context.Context, countStudents int, defaultPaymentMethods string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultPaymentMethodSlice := strings.Split(defaultPaymentMethods, "-")

	for i := 0; i < countStudents; i++ {
		switch defaultPaymentMethodSlice[i] {
		case "NO_DEFAULT_PAYMENT":
			ctx, err := s.anExistingStudentWithNoStudentPaymentMethod(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		default:
			ctx, err := s.anExistingStudentWithDefaultPaymentMethod(ctx, defaultPaymentMethodSlice[i])
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theRetrieveBulkStudentPaymentMethodEndpointIsCalledForThisStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
		StudentIds: stepState.StudentIds,
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewPaymentServiceClient(s.InvoiceMgmtConn).RetrieveBulkStudentPaymentMethod(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentMethodsForTheseStudentsAreRetrieveSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp, ok := stepState.Response.(*invoice_pb.RetrieveBulkStudentPaymentMethodResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("RetrieveStudentPaymentMethodResponse is nil")
	}

	for _, studentPaymentMethods := range resp.StudentPaymentMethods {
		expectedPaymentMethod := stepState.StudentPaymentMethodMap[studentPaymentMethods.StudentId]
		if studentPaymentMethods.PaymentMethod.String() != expectedPaymentMethod {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on retrieving student payment method expected student %v with payment method %v but got payment method %v", studentPaymentMethods.StudentId, expectedPaymentMethod, studentPaymentMethods.PaymentMethod.String())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
