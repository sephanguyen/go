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

func (s *PaymentModifierService) RetrieveStudentPaymentMethod(ctx context.Context, req *invoice_pb.RetrieveStudentPaymentMethodRequest) (*invoice_pb.RetrieveStudentPaymentMethodResponse, error) {
	if strings.TrimSpace(req.StudentId) == "" {
		return nil, status.Error(codes.FailedPrecondition, "student id cannot be empty")
	}

	// selecting default payment method of student
	existingStudentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, s.DB, req.StudentId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail FindByStudentID: %v on student id: %v", err, req.StudentId))
	}

	response := &invoice_pb.RetrieveStudentPaymentMethodResponse{
		Successful:    true,
		PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
		StudentId:     req.StudentId,
	}

	if existingStudentPaymentDetail != nil && strings.TrimSpace(existingStudentPaymentDetail.PaymentMethod.String) != "" {
		response.PaymentMethod = constant.PaymentMethodsConvertToEnums[existingStudentPaymentDetail.PaymentMethod.String]
	}

	return response, nil
}
