package paymentsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	generator "github.com/manabie-com/backend/internal/invoicemgmt/services/payment/payment_request_generator"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) CreatePaymentRequest(ctx context.Context, req *invoice_pb.CreatePaymentRequestRequest) (*invoice_pb.CreatePaymentRequestResponse, error) {
	requestDateTime, err := utils.GetTimeInLocation(time.Now(), utils.CountryJp)
	if err != nil {
		return nil, err
	}

	// validate the request
	err = s.validateGeneratePaymentRequest(req)
	if err != nil {
		return nil, err
	}

	// The creation of bulk payment request is outside the transaction so that we can still track the request when error occurred.
	// When error occurred during request, this entity should be updated with the error details
	bulkPaymentRequest, err := generateBulkPaymentRequestEntity(req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	bulkPaymentRequestID, err := s.BulkPaymentRequestRepo.Create(ctx, s.DB, bulkPaymentRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestRepo.Create err: %v", err))
	}

	err = s.processPaymentRequestFile(ctx, bulkPaymentRequestID, req, requestDateTime)
	if err != nil {
		// Save the error details to bulk payment request
		bulkPaymentRequest.ErrorDetails = pgtype.Text{String: err.Error(), Status: pgtype.Present}
		updateErr := s.BulkPaymentRequestRepo.Update(ctx, s.DB, bulkPaymentRequest)
		if updateErr != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("cannot update the error details of bulk payment request err: %v", err))
		}

		return nil, err
	}

	return &invoice_pb.CreatePaymentRequestResponse{
		Successful:           true,
		BulkPaymentRequestId: bulkPaymentRequestID,
	}, nil
}

func (s *PaymentModifierService) processPaymentRequestFile(
	ctx context.Context,
	bulkPaymentRequestID string,
	req *invoice_pb.CreatePaymentRequestRequest,
	requestDateTime time.Time,
) error {
	var paymentRequestGenerator generator.PaymentRequestGenerator

	basePaymentGenerator := &generator.BasePaymentRequestGenerator{
		DB:                                s.DB,
		Logger:                            s.logger,
		UnleashClient:                     s.UnleashClient,
		Env:                               s.Env,
		PaymentRepo:                       s.PaymentRepo,
		InvoiceRepo:                       s.InvoiceRepo,
		BulkPaymentRequestFileRepo:        s.BulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: s.BulkPaymentRequestFilePaymentRepo,
		PartnerConvenienceStoreRepo:       s.PartnerConvenienceStoreRepo,
		StudentPaymentDetailRepo:          s.StudentPaymentDetailRepo,
		BankBranchRepo:                    s.BankBranchRepo,
		NewCustomerCodeHistoryRepo:        s.NewCustomerCodeHistoryRepo,
		PrefectureRepo:                    s.PrefectureRepo,
		BulkPaymentRepo:                   s.BulkPaymentRepo,
		Validator:                         utils.PaymentRequestValidator{},
		FileStorage:                       s.FileStorage,
		TempFileCreator:                   s.TempFileCreator,
	}

	switch req.PaymentMethod.String() {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
		useKECFeedbackPh1, err := s.UnleashClient.IsFeatureEnabled(constant.EnableKECFeedbackPh1, s.Env)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableKECFeedbackPh1, err))
		}

		enableOptionalValidationInPaymentRequest, err := s.UnleashClient.IsFeatureEnabled(constant.EnableOptionalValidationInPaymentRequest, s.Env)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableOptionalValidationInPaymentRequest, err))
		}

		paymentRequestGenerator = &generator.CScsvPaymentRequestGenerator{
			BasePaymentRequestGenerator:              basePaymentGenerator,
			Req:                                      req,
			BulkPaymentRequestID:                     bulkPaymentRequestID,
			UseKECFeedbackPh1:                        useKECFeedbackPh1,
			RequestDate:                              requestDateTime,
			EnableOptionalValidationInPaymentRequest: enableOptionalValidationInPaymentRequest,
		}
	case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
		paymentRequestGenerator = &generator.DDtxtPaymentRequestGenerator{
			BasePaymentRequestGenerator: basePaymentGenerator,
			Req:                         req,
			BulkPaymentRequestID:        bulkPaymentRequestID,
		}
	}

	useGCloudUploadFeature, err := s.UnleashClient.IsFeatureEnabled(constant.EnableGCloudUploadFeatureFlag, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableGCloudUploadFeatureFlag, err))
	}

	err = paymentRequestGenerator.ValidateDataV2(ctx)
	if err != nil {
		return err
	}

	err = paymentRequestGenerator.PlanPaymentAndFileAssociation(ctx)
	if err != nil {
		return err
	}

	switch useGCloudUploadFeature {
	case true:
		err = paymentRequestGenerator.SaveAndUploadPaymentFileV2(ctx)
		if err != nil {
			return err
		}

	default:
		err = paymentRequestGenerator.SaveAndUploadPaymentFile(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateBulkPaymentRequestEntity(data *invoice_pb.CreatePaymentRequestRequest) (*entities.BulkPaymentRequest, error) {
	e := new(entities.BulkPaymentRequest)
	database.AllNullEntity(e)

	if err := multierr.Combine(
		e.PaymentMethod.Set(data.PaymentMethod.String()),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}

func (s *PaymentModifierService) validateGeneratePaymentRequest(req *invoice_pb.CreatePaymentRequestRequest) error {
	if len(req.PaymentIds) == 0 {
		return status.Error(codes.InvalidArgument, "payment IDs cannot be empty")
	}

	switch req.PaymentMethod.String() {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
		if req.ConvenienceStoreDates == nil {
			return status.Error(codes.InvalidArgument, "convenience store dates cannot be empty")
		}

		useKECFeedbackPh1, err := s.UnleashClient.IsFeatureEnabled(constant.EnableKECFeedbackPh1, s.Env)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableKECFeedbackPh1, err))
		}

		isThereEmptyDueDates := req.ConvenienceStoreDates.DueDateFrom == nil || req.ConvenienceStoreDates.DueDateUntil == nil
		if isThereEmptyDueDates {
			if useKECFeedbackPh1 {
				return nil
			}

			return status.Error(codes.InvalidArgument, "due date from or until cannot be empty")
		}

		// if due_date_from is greater than due_date_until, return error
		if req.ConvenienceStoreDates.DueDateFrom.AsTime().After(req.ConvenienceStoreDates.DueDateUntil.AsTime()) {
			return status.Error(codes.InvalidArgument, "due_date_from should not be ahead on due_date_until")
		}

	case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
		if req.DirectDebitDates == nil {
			return status.Error(codes.InvalidArgument, "direct debit dates cannot be empty")
		}

		if req.DirectDebitDates.DueDate == nil {
			return status.Error(codes.InvalidArgument, "due date cannot be empty")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid payment method")
	}

	return nil
}
