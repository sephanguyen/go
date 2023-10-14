package paymentdetail

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_common "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/common"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EditPaymentDetailService) UpdateStudentPaymentMethod(ctx context.Context, request *invoice_pb.UpdateStudentPaymentMethodRequest) (*invoice_pb.UpdateStudentPaymentMethodResponse, error) {
	// validate request
	if err := validateStudentPaymentMethodRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check if student payment detail exists, it will cover checking student also has a foreign key
	studentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByID(ctx, s.DB, request.StudentPaymentDetailId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail FindByID: %v", err))
	}

	// check if billing address exists to cover the selection of convenience store payment
	_, err = s.BillingAddressRepo.FindByUserID(ctx, s.DB, request.StudentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error Billing Address FindByUserID: %v", err))
	}

	// check if bank account exist and should be verified as selecting payment method should be verified for 2 payment method options
	// We don't manage multiple bank account at this time, amend this logic when we do
	existingBankAccount, err := s.BankAccountRepo.FindByStudentID(ctx, s.DB, request.StudentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error BankAccount FindByStudentID: %v", err))
	}

	if !existingBankAccount.IsVerified.Bool {
		return nil, status.Error(codes.Internal, "error existing bank account should be verified")
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		oldPaymentMethod := studentPaymentDetail.PaymentMethod.String

		if oldPaymentMethod == request.PaymentMethod.String() {
			return nil
		}

		// update the student payment detail payment method
		studentPaymentDetail.PaymentMethod = database.Text(request.PaymentMethod.String())
		if err := s.StudentPaymentDetailRepo.Upsert(ctx, tx, studentPaymentDetail); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail Upsert: %v", err))
		}

		return s.saveUpdatePaymentMethodActionLog(ctx, tx, oldPaymentMethod, studentPaymentDetail)
	}); err != nil {
		return nil, err
	}

	return &invoice_pb.UpdateStudentPaymentMethodResponse{
		Successful: true,
	}, nil
}

func (s *EditPaymentDetailService) saveUpdatePaymentMethodActionLog(ctx context.Context, db database.QueryExecer, oldPaymentMethod string, studentPaymentDetail *entities.StudentPaymentDetail) error {
	studentPaymentActionLogDetailType := StudentPaymentActionDetailLogType{
		&PreviousDataStudentActionDetailLog{
			PaymentMethod: oldPaymentMethod,
		},
		&NewDataStudentActionDetailLog{
			PaymentMethod: studentPaymentDetail.PaymentMethod.String,
		},
	}

	studentPaymentDetailActionLog, err := generateStudentPaymentDetailActionLog(ctx, &studentPaymentDetailActionLogData{
		actionDetailInfo:       database.JSONB(studentPaymentActionLogDetailType),
		action:                 invoice_common.StudentPaymentDetailAction_UPDATED_PAYMENT_METHOD.String(),
		StudentPaymentDetailID: studentPaymentDetail.StudentPaymentDetailID.String,
	})
	if err != nil {
		return err
	}

	if err := s.StudentPaymentDetailActionLogRepo.Create(ctx, db, studentPaymentDetailActionLog); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("err cannot create student payment detail action log: %v", err))
	}

	return nil
}

// validating student request from client
func validateStudentPaymentMethodRequest(req *invoice_pb.UpdateStudentPaymentMethodRequest) error {
	if strings.TrimSpace(req.StudentId) == "" {
		return errors.New("student id cannot be empty")
	}

	if strings.TrimSpace(req.StudentPaymentDetailId) == "" {
		return errors.New("student payment detail id cannot be empty")
	}

	// Only Convenience Store and Direct Debit payment method
	if !constant.StudentPaymentMethods[req.PaymentMethod.String()] {
		return fmt.Errorf("invalid PaymentMethod value: %s", req.PaymentMethod.String())
	}

	return nil
}
