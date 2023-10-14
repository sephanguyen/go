package paymentsvc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PaymentModifierService) DownloadBulkPaymentValidationsDetail(ctx context.Context, req *invoice_pb.DownloadBulkPaymentValidationsDetailRequest) (*invoice_pb.DownloadBulkPaymentValidationsDetailResponse, error) {
	// validate request for bulk payment validations id
	if req.BulkPaymentValidationsId == "" {
		return nil, status.Error(codes.FailedPrecondition, "invalid empty bulk payment validations id")
	}
	// find bulk payment validations if record exist
	bulkPaymentValidations, err := s.BulkPaymentValidationsRepo.FindByID(ctx, s.DB, req.BulkPaymentValidationsId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	paymentMethod := bulkPaymentValidations.PaymentMethod.String
	// get bulk validation details
	listBulkPaymentValidationsDetails, err := s.BulkPaymentValidationsDetailRepo.RetrieveRecordsByBulkPaymentValidationsID(ctx, s.DB, database.Text(req.BulkPaymentValidationsId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(listBulkPaymentValidationsDetails) == 0 {
		return nil, status.Error(codes.Internal, "error no associated bulk payment validation detail records")
	}
	// compute total success and failed validations to compare total retrieve validation detail records
	totalValidationRecords := bulkPaymentValidations.SuccessfulPayments.Int + bulkPaymentValidations.FailedPayments.Int + bulkPaymentValidations.PendingPayments.Int
	if len(listBulkPaymentValidationsDetails) != int(totalValidationRecords) {
		return nil, status.Error(codes.Internal, fmt.Sprintf("bulk payment validations detail records count not match expected %d got %d on bulk payment validations id %v", totalValidationRecords, len(listBulkPaymentValidationsDetails), req.BulkPaymentValidationsId))
	}

	listImportedPaymentValidationDetail := []*invoice_pb.ImportPaymentValidationDetail{}
	for _, bulkPaymentValidationDetails := range listBulkPaymentValidationsDetails {
		importedPaymentValidationDetails, err := s.generateImportedPaymentValidationDetail(ctx, bulkPaymentValidationDetails, paymentMethod)
		if err != nil {
			return nil, err
		}
		listImportedPaymentValidationDetail = append(listImportedPaymentValidationDetail, importedPaymentValidationDetails)
	}

	return &invoice_pb.DownloadBulkPaymentValidationsDetailResponse{
		PaymentValidationDetail: listImportedPaymentValidationDetail,
		ValidationDate:          timestamppb.New(bulkPaymentValidations.ValidationDate.Time),
	}, nil
}

func (s *PaymentModifierService) generateImportedPaymentValidationDetail(ctx context.Context, bulkPaymentValidationsDetails *entities.BulkPaymentValidationsDetail, paymentMethod string) (*invoice_pb.ImportPaymentValidationDetail, error) {
	// find associated invoice
	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, bulkPaymentValidationsDetails.InvoiceID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	invoiceSequenceNumber := invoice.InvoiceSequenceNumber

	// find associated payment
	payment, err := s.PaymentRepo.FindByPaymentID(ctx, s.DB, bulkPaymentValidationsDetails.PaymentID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// compare payment method on bulk payment validations
	if payment.PaymentMethod.String != paymentMethod {
		return nil, status.Error(codes.Internal, "payment method are not match")
	}

	paymentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(payment.Amount, "2")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err := s.UserBasicInfoRepo.FindByID(ctx, s.DB, invoice.StudentID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	importedPaymentValidationDetail := &invoice_pb.ImportPaymentValidationDetail{
		PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
		Result:                bulkPaymentValidationsDetails.ValidatedResultCode.String,
		Amount:                paymentAmount,
		StudentId:             user.UserID.String,
		StudentName:           user.Name.String,
		PaymentMethod:         constant.PaymentMethodsConvertToEnums[payment.PaymentMethod.String],
		InvoiceSequenceNumber: invoiceSequenceNumber.Int,
		InvoiceId:             invoice.InvoiceID.String,
		PaymentStatus:         payment.PaymentStatus.String,
	}

	return importedPaymentValidationDetail, nil
}
