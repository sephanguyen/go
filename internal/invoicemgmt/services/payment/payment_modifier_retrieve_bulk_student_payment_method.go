package paymentsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) RetrieveBulkStudentPaymentMethod(ctx context.Context, req *invoice_pb.RetrieveBulkStudentPaymentMethodRequest) (*invoice_pb.RetrieveBulkStudentPaymentMethodResponse, error) {
	if len(req.StudentIds) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "error request student ids cannot be empty")
	}

	studentPaymentMethods := make([]*invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod, len(req.StudentIds))
	for i, studentID := range req.StudentIds {
		// using loop to compare if a single student id has a existing student payment detail
		existingStudentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, s.DB, studentID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail FindByStudentID: %v on student id: %v", err, studentID))
		}
		studentPaymentMethod := &invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod{
			StudentId:     studentID,
			PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
		}

		if existingStudentPaymentDetail != nil && strings.TrimSpace(existingStudentPaymentDetail.PaymentMethod.String) != "" {
			ok := constant.StudentPaymentMethods[existingStudentPaymentDetail.PaymentMethod.String]
			if !ok {
				return nil, status.Error(codes.Internal, fmt.Sprintf("error invalid student default payment method: %v", existingStudentPaymentDetail.PaymentMethod.String))
			}
			studentPaymentMethod.PaymentMethod = constant.PaymentMethodsConvertToEnums[existingStudentPaymentDetail.PaymentMethod.String]
		}
		studentPaymentMethods[i] = studentPaymentMethod
	}
	return &invoice_pb.RetrieveBulkStudentPaymentMethodResponse{
		Successful:            true,
		StudentPaymentMethods: studentPaymentMethods,
	}, nil
}
