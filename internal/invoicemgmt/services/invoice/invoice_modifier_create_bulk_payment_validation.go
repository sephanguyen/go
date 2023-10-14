package invoicesvc

import (
	"context"
	"fmt"
	"sort"

	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	pfutils "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_utils"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (s *InvoiceModifierService) CreateBulkPaymentValidation(ctx context.Context, req *invoice_pb.CreateBulkPaymentValidationRequest) (*invoice_pb.CreateBulkPaymentValidationResponse, error) {
	response := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful:              false,
		PaymentValidationDetail: nil,
	}

	if err := validateCreateBulkPaymentValidationRequest(req); err != nil {
		return response, status.Error(codes.InvalidArgument, err.Error())
	}

	file, err := s.convertPayload(ctx, req.Payload, req.PaymentMethod)
	if err != nil {
		return response, status.Error(codes.InvalidArgument, err.Error())
	}

	validatedPaymentResult, err := s.validateFile(ctx, file, req.DirectDebitPaymentDate, req.PaymentMethod)
	if err != nil {
		return response, status.Error(codes.InvalidArgument, err.Error())
	}

	response.PaymentValidationDetail = createBulkPaymentResponseDetails(validatedPaymentResult.ValidatedPayments)
	response.ValidationDate = timestamppb.New(*validatedPaymentResult.ValidationDate)
	response.SuccessfulPayments = validatedPaymentResult.SuccessfulPayments
	response.PendingPayments = validatedPaymentResult.PendingPayments
	response.FailedPayments = validatedPaymentResult.FailedPayments
	response.Successful = true

	return response, nil
}

func validateCreateBulkPaymentValidationRequest(req *invoice_pb.CreateBulkPaymentValidationRequest) error {
	switch req.PaymentMethod {
	case invoice_pb.PaymentMethod_DIRECT_DEBIT, invoice_pb.PaymentMethod_CONVENIENCE_STORE:
		break
	default:
		return fmt.Errorf("invalid payment method")
	}

	if req.Payload == nil {
		return fmt.Errorf("file payload is required")
	}

	if req.PaymentMethod == invoice_pb.PaymentMethod_DIRECT_DEBIT {
		if req.DirectDebitPaymentDate == nil {
			return fmt.Errorf("direct debit payment date is required")
		}
	}

	return nil
}

func (s *InvoiceModifierService) convertPayload(ctx context.Context, fileContents []byte, paymentMethod invoice_pb.PaymentMethod) (*pfutils.PaymentFile, error) {
	var fileConverter pfutils.PaymentFileConverter

	// No default case; validation already performed before this step
	switch paymentMethod {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE:
		fileConverter = &pfutils.ConvenienceStoreCSVPaymentFileConverter{}
	case invoice_pb.PaymentMethod_DIRECT_DEBIT:
		fileConverter = &pfutils.DirectDebitTextPaymentFileConverter{}
	}

	enableEncodePaymentRequestFiles, err := s.UnleashClient.IsFeatureEnabled(constant.EnableEncodePaymentRequestFiles, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableEncodePaymentRequestFiles, err))
	}

	// Decode payload in Shift-JIS
	byteContent := fileContents
	if enableEncodePaymentRequestFiles {
		byteContent, err = utils.DecodeByteToShiftJIS(fileContents)
		if err != nil {
			return nil, err
		}
	}

	file, err := fileConverter.ConvertFromBytesToPaymentFile(ctx, byteContent)
	if err != nil {
		return nil, fmt.Errorf("error processing the payload: %v", err)
	}

	return file, nil
}

func (s *InvoiceModifierService) validateFile(ctx context.Context, paymentFile *pfutils.PaymentFile, paymentDate *timestamppb.Timestamp, paymentMethod invoice_pb.PaymentMethod) (*pfutils.PaymentValidationResult, error) {
	var paymentValidator pfutils.PaymentFileValidator

	basePaymentValidator := &pfutils.BasePaymentFileValidator{
		DB:                               s.DB,
		InvoiceRepo:                      s.InvoiceRepo,
		PaymentRepo:                      s.PaymentRepo,
		BulkPaymentValidationsRepo:       s.BulkPaymentValidationsRepo,
		BulkPaymentValidationsDetailRepo: s.BulkPaymentValidationsDetailRepo,
		InvoiceActionLogRepo:             s.InvoiceActionLogRepo,
		UserBasicInfoRepo:                s.UserBasicInfoRepo,
		UnleashClient:                    s.UnleashClient,
		Env:                              s.Env,
	}

	// No default case; validation already performed before this step
	switch paymentMethod {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE:
		paymentValidator = &pfutils.ConvenienceStoreCSVPaymentFileValidator{
			BasePaymentFileValidator: basePaymentValidator,
		}
	case invoice_pb.PaymentMethod_DIRECT_DEBIT:
		paymentValidator = &pfutils.DirectDebitTextPaymentFileValidator{
			BasePaymentFileValidator: basePaymentValidator,
			PaymentDate:              paymentDate.AsTime(),
		}
	}

	return paymentValidator.Validate(ctx, paymentFile)
}

func createBulkPaymentResponseDetails(validatedPayments []*pfutils.ValidatedPayment) []*invoice_pb.ImportPaymentValidationDetail {
	paymentValidationDetails := make([]*invoice_pb.ImportPaymentValidationDetail, 0)

	// Sort by created date desc
	sort.Slice(validatedPayments, func(x, y int) bool {
		return !validatedPayments[x].PaymentCreatedDate.Before(validatedPayments[y].PaymentCreatedDate)
	})

	// Create response details
	for _, validatedPayment := range validatedPayments {
		paymentValidationDetails = append(paymentValidationDetails, &invoice_pb.ImportPaymentValidationDetail{
			Amount:                validatedPayment.Amount,
			InvoiceSequenceNumber: validatedPayment.InvoiceSequenceNumber,
			PaymentSequenceNumber: validatedPayment.PaymentSequenceNumber,
			Result:                validatedPayment.ResultCode,
			StudentId:             validatedPayment.StudentID,
			StudentName:           validatedPayment.StudentName,
			PaymentMethod:         validatedPayment.PaymentMethod,
			InvoiceId:             validatedPayment.InvoiceID,
			PaymentStatus:         validatedPayment.PaymentStatus,
		})
	}

	return paymentValidationDetails
}
