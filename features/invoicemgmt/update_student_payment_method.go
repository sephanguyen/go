package invoicemgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) studentBankAccountIsSetToStatus(ctx context.Context, bankAccountStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var isVerified bool
	if bankAccountStatus == "verified" {
		isVerified = true
	}

	stmt := `UPDATE bank_account SET is_verified = $1 WHERE student_id = $2`

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, isVerified, stepState.StudentID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error update bank account verified status: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPaymentDetailHasPaymentMethod(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `UPDATE student_payment_detail SET payment_method = $1 WHERE student_id = $2`

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, paymentMethod, stepState.StudentID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error update student payment detail payment method: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updatesPaymentMethodOfTheStudent(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch paymentMethod {
	case "CONVENIENCE_STORE":
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
	case "DIRECT_DEBIT":
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("payment method %s is not supported", paymentMethod)
	}

	req := &invoice_pb.UpdateStudentPaymentMethodRequest{
		StudentId:              stepState.StudentID,
		StudentPaymentDetailId: stepState.StudentPaymentDetailID,
		PaymentMethod:          constant.PaymentMethodsConvertToEnums[paymentMethod],
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewEditPaymentDetailServiceClient(s.InvoiceMgmtConn).UpdateStudentPaymentMethod(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPaymentMethodIsUpdatedSuccesfullyTo(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	resp, ok := stepState.Response.(*invoice_pb.UpdateStudentPaymentMethodResponse)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update student payment method response is nil")
	}

	if !resp.Successful {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot update student payment method successfully")
	}

	studentPaymentDetailRepo := &repositories.StudentPaymentDetailRepo{}
	studentPaymentDetail, err := studentPaymentDetailRepo.FindByID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentPaymentDetailID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentPaymentDetailRepo.FindByID err: %v", err)
	}

	if studentPaymentDetail.PaymentMethod.String != paymentMethod {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student payment method of student is not updated to: %v", paymentMethod)
	}

	return StepStateToContext(ctx, stepState), nil
}
