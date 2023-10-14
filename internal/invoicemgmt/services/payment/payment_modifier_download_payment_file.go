package paymentsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	downloader "github.com/manabie-com/backend/internal/invoicemgmt/services/payment/payment_file_downloader"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) DownloadPaymentFile(ctx context.Context, req *invoice_pb.DownloadPaymentFileRequest) (*invoice_pb.DownloadPaymentFileResponse, error) {
	err := validateDownloadPaymentFileRequest(req)
	if err != nil {
		return nil, err
	}

	// Fetch the payment request file from DB
	paymentRequestFile, err := s.BulkPaymentRequestFileRepo.FindByPaymentFileID(ctx, s.DB, req.PaymentRequestFileId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFileRepo.FindByID err: %v", err))
	}

	// Fetch the payment request from DB
	paymentRequest, err := s.BulkPaymentRequestRepo.FindByPaymentRequestID(ctx, s.DB, paymentRequestFile.BulkPaymentRequestID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestRepo.FindByPaymentRequestID err: %v", err))
	}

	var (
		paymentFileDownloader downloader.PaymentFileDownloader
		fileType              invoice_pb.FileType
	)

	basePaymentFileDownloader := &downloader.BasePaymentFileDownloader{
		DB:                                s.DB,
		Logger:                            s.logger,
		BulkPaymentRequestFilePaymentRepo: s.BulkPaymentRequestFilePaymentRepo,
		BulkPaymentRequestFileRepo:        s.BulkPaymentRequestFileRepo,
		PartnerConvenienceStoreRepo:       s.PartnerConvenienceStoreRepo,
		PartnerBankRepo:                   s.PartnerBankRepo,
		StudentPaymentDetailRepo:          s.StudentPaymentDetailRepo,
		BankBranchRepo:                    s.BankBranchRepo,
		NewCustomerCodeHistoryRepo:        s.NewCustomerCodeHistoryRepo,
		PrefectureRepo:                    s.PrefectureRepo,
		FileStorage:                       s.FileStorage,
		TempFileCreator:                   s.TempFileCreator,
		Validator:                         &utils.PaymentRequestValidator{},
	}

	switch paymentRequest.PaymentMethod.String {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
		useKECFeedbackPh1, err := s.UnleashClient.IsFeatureEnabled(constant.EnableKECFeedbackPh1, s.Env)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableKECFeedbackPh1, err))
		}

		enableOptionalValidationInPaymentRequest, err := s.UnleashClient.IsFeatureEnabled(constant.EnableOptionalValidationInPaymentRequest, s.Env)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableOptionalValidationInPaymentRequest, err))
		}

		fileType = invoice_pb.FileType_CSV
		paymentFileDownloader = &downloader.ConvenienceStoreCSVPaymentFileDownloader{
			BasePaymentFileDownloader:                basePaymentFileDownloader,
			PaymentFileID:                            paymentRequestFile.BulkPaymentRequestFileID.String,
			UseKECFeedbackPh1:                        useKECFeedbackPh1,
			EnableOptionalValidationInPaymentRequest: enableOptionalValidationInPaymentRequest,
		}

	case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
		fileType = invoice_pb.FileType_TXT
		paymentFileDownloader = &downloader.DirectDebitTXTPaymentFileDownloader{
			BasePaymentFileDownloader: basePaymentFileDownloader,
			PaymentFileID:             paymentRequestFile.BulkPaymentRequestFileID.String,
		}
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("payment method %s is invalid", paymentRequest.PaymentMethod.String))
	}

	useGCloudUploadFeature, err := s.UnleashClient.IsFeatureEnabled(constant.EnableGCloudUploadFeatureFlag, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.UnleashClient.IsFeatureEnabled err: %v", err))
	}

	var byteContent []byte
	switch {
	case useGCloudUploadFeature:
		byteContent, err = paymentFileDownloader.GetByteContentV2(ctx)
		if err != nil {
			return nil, err
		}
	default:
		err = paymentFileDownloader.ValidateData(ctx)
		if err != nil {
			return nil, err
		}

		byteContent, err = paymentFileDownloader.GetByteContent(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &invoice_pb.DownloadPaymentFileResponse{
		Successful: true,
		Data:       byteContent,
		FileType:   fileType,
	}, nil
}

func validateDownloadPaymentFileRequest(req *invoice_pb.DownloadPaymentFileRequest) error {
	if strings.TrimSpace(req.PaymentRequestFileId) == "" {
		return status.Error(codes.InvalidArgument, "payment_request_file_id should not be empty")
	}
	return nil
}
