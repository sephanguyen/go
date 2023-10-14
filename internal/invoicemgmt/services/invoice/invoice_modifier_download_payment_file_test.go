package invoicesvc

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	downloader "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_downloader"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_filestorage "github.com/manabie-com/backend/mock/invoicemgmt/services/filestorage"
	mock_utils "github.com/manabie-com/backend/mock/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_DownloadPaymentFile_CS(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)

	// Generate Invoice Mocks
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRequestRepo := new(mock_repositories.MockBulkPaymentRequestRepo)
	mockBulkPaymentRequestFileRepo := new(mock_repositories.MockBulkPaymentRequestFileRepo)
	mockBulkPaymentRequestFilePaymentRepo := new(mock_repositories.MockBulkPaymentRequestFilePaymentRepo)
	mockPartnerConvenienceStoreRepo := new(mock_repositories.MockPartnerConvenienceStoreRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockPrefectureRepo := new(mock_repositories.MockPrefectureRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockFileStorage := new(mock_filestorage.FileStorage)
	mockTempFileCreator := new(mock_utils.ITempFileCreator)

	mockPaymentRequestID := "test-request-id-1"
	mockPaymentRequest := &entities.BulkPaymentRequest{
		BulkPaymentRequestID: database.Text(mockPaymentRequestID),
		PaymentMethod:        database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}

	mockPaymentRequestFileID := "test-request-file-id-1"
	mockBulkPaymentRequestFile := &entities.BulkPaymentRequestFile{
		BulkPaymentRequestFileID: database.Text(mockPaymentRequestFileID),
		BulkPaymentRequestID:     database.Text(mockPaymentRequestID),
		FileName:                 database.Text("mycsvfile.csv"),
	}

	mockPartnerCsID := "partner-convenience-store-id-1"
	mockPartnerConvenienceStore := &entities.PartnerConvenienceStore{
		PartnerConvenienceStoreID: database.Text(mockPartnerCsID),
		ManufacturerCode:          database.Int4(123456),
		CompanyCode:               database.Int4(12345),
		ShopCode:                  database.Text("shop-code-1"),
		CompanyName:               database.Text("company-name-1"),
		CompanyTelNumber:          database.Text("company-tel-number-1"),
		PostalCode:                database.Text("postal-code-1"),
		Address1:                  database.Text("address-1"),
		Address2:                  database.Text("address-2"),
	}

	mockPaymentIDs := []string{}
	for i := 0; i < 3; i++ {
		mockPaymentIDs = append(mockPaymentIDs, fmt.Sprintf("mock-payment-id-%d", i))
	}

	mockFilePaymentInvoiceMap, mockFilePaymentData, studentBillingMap := genMockFilePaymentRequestDataForCC(
		mockPaymentIDs,
		mockPaymentRequestFileID,
		invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
		invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
		true,
	)

	testError := errors.New("test error")

	zapLogger := logger.NewZapLogger("debug", true)
	// Init service
	s := &InvoiceModifierService{
		DB:                                mockDB,
		logger:                            *zapLogger.Sugar(),
		InvoiceRepo:                       mockInvoiceRepo,
		PaymentRepo:                       mockPaymentRepo,
		BulkPaymentRequestRepo:            mockBulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:        mockBulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: mockBulkPaymentRequestFilePaymentRepo,
		PartnerConvenienceStoreRepo:       mockPartnerConvenienceStoreRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		PrefectureRepo:                    mockPrefectureRepo,
		UnleashClient:                     mockUnleashClient,
		Env:                               "local",
		FileStorage:                       mockFileStorage,
		TempFileCreator:                   mockTempFileCreator,
	}
	prefectures := []*entities.Prefecture{
		{
			PrefectureCode: database.Text("test-code-1"),
			Name:           database.Text("test-name-1"),
			ID:             database.Text("test-id-1"),
		},
		{
			PrefectureCode: database.Text("test-code-2"),
			Name:           database.Text("test-name-2"),
			ID:             database.Text("test-id-2"),
		},
		{
			PrefectureCode: database.Text("test-code-3"),
			Name:           database.Text("test-name-3"),
			ID:             database.Text("test-id-3"),
		},
	}

	prefectureCodeWithNameMapped := map[string]string{
		"test-code-1": "test-name-1",
		"test-code-2": "test-name-2",
		"test-code-3": "test-name-3",
	}

	expectedCSVData, _ := downloader.GenCSVData(mockPartnerConvenienceStore, mockFilePaymentData, prefectureCodeWithNameMapped, false)
	expectedCSVByte, _ := downloader.GenCSVBytesFromData(expectedCSVData)
	successfulCSVResponse := &invoice_pb.DownloadPaymentFileResponse{
		Successful: true,
		Data:       expectedCSVByte,
		FileType:   invoice_pb.FileType_CSV,
	}

	testcases := []TestCase{
		{
			name: "Happy case - generate CSV bytes",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulCSVResponse,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(studentBillingMap, nil)
			},
		},
		{
			name: "The payment request file ID parameter is empty",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "payment_request_file_id should not be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "BulkPaymentRequestFileRepo.FindByPaymentFileID returns error",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFileRepo.FindByID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "BulkPaymentRequestRepo.FindByPaymentRequestID returns error",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestRepo.FindByPaymentRequestID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Partner CS Manufacturer code is invalid",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The manufacturer_code of the partner CS should be 6 digits"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				invalidCS := &entities.PartnerConvenienceStore{
					PartnerConvenienceStoreID: database.Text(mockPartnerCsID),
					ManufacturerCode:          database.Int4(1),
					CompanyCode:               database.Int4(12345),
					ShopCode:                  database.Text("shop-code-1"),
					CompanyName:               database.Text("company-name-1"),
					CompanyTelNumber:          database.Text("company-tel-number-1"),
					PostalCode:                database.Text("postal-code-1"),
					Address1:                  database.Text("address-1"),
					Address2:                  database.Text("address-2"),
				}

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(invalidCS, nil)
			},
		},
		{
			name: "Partner CS company code is invalid",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The company_code of the partner CS should be 5 digits"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				invalidCS := &entities.PartnerConvenienceStore{
					PartnerConvenienceStoreID: database.Text(mockPartnerCsID),
					ManufacturerCode:          database.Int4(123456),
					CompanyCode:               database.Int4(1),
					ShopCode:                  database.Text("shop-code-1"),
					CompanyName:               database.Text("company-name-1"),
					CompanyTelNumber:          database.Text("company-tel-number-1"),
					PostalCode:                database.Text("postal-code-1"),
					Address1:                  database.Text("address-1"),
					Address2:                  database.Text("address-2"),
				}

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(invalidCS, nil)
			},
		},
		{
			name: "Partner has no partner CS",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "Partner has no associated convenience store"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "PartnerConvenienceStoreRepo.FindOne returns error",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("d.PartnerConvenienceStoreRepo.FindOne err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(nil, testError)
			},
		},
		{
			name: "BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID returns error",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("d.BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "The returned list of filePaymentInvoiceMap is empty",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "There is no associated payment in the request file"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return([]*entities.FilePaymentInvoiceMap{}, nil)
			},
		},
		{
			name: "The associated payment has invalid payment method",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "the payment method is not equal to the given payment method parameter"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				invalid, _, _ := genMockFilePaymentRequestDataForCC(
					[]string{"payment-id-1"},
					mockPaymentRequestFileID,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					true,
				)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(invalid, nil)
			},
		},
		{
			name: "The associated payment is not yet exported",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "payment isExported field should be true"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				invalid, _, _ := genMockFilePaymentRequestDataForCC(
					[]string{"payment-id-1"},
					mockPaymentRequestFileID,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(invalid, nil)
			},
		},
		{
			name: "The length of invoice total amount exceeds the requirement",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The invoice total length exceeds the limit"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				invalid, _, _ := genMockFilePaymentRequestDataForCC(
					[]string{"payment-id-1"},
					mockPaymentRequestFileID,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					true,
				)
				invalid[0].Invoice.Total.Int = big.NewInt(int64(1231231231231231231))

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(invalid, nil)
			},
		},
		{
			name: "No student billing info records exists",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulCSVResponse,
			expectedErr:  status.Error(codes.Internal, "No student billing detail records"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.StudentBillingDetailsMap{}, nil)
			},
		},
		{
			name: "No associated student billing info records exists for a student",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulCSVResponse,
			expectedErr:  status.Error(codes.Internal, "There is a student that does not have billing details"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				_, _, studentBillingMap := genMockFilePaymentRequestDataForCC(
					[]string{"payment-id-1"},
					mockPaymentRequestFileID,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					true,
				)
				studentBillingMap[0].StudentPaymentDetail.StudentID = database.Text("test-not-exist")
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(studentBillingMap, nil)
			},
		},
		{
			name: "Invalid Student Billing Address Prefecture code not match on prefecture table",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "student student-id-0 with billing details prefecture code test000 that does not match prefecture records"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				studentBillingMap[0].BillingAddress.PrefectureCode = database.Text("test000")
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(studentBillingMap, nil)
			},
		},
		{
			name: "Happy case - generate CSV bytes with GCloud Upload turned on",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulCSVResponse,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				// mock the creation of temp CSV file
				objectName := "mycsvfile.csv"
				tempFileCreator := &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"}
				tempFile, err := tempFileCreator.CreateTempFile(objectName)
				if err != nil {
					t.Error(err)
				}

				// add the expected data to the file
				csvWriter := csv.NewWriter(tempFile.File)
				err = csvWriter.WriteAll(expectedCSVData)
				if err != nil {
					t.Error(err)
				}

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockTempFileCreator.On("CreateTempFile", objectName).Once().Return(tempFile, err)
				mockFileStorage.On("DownloadFile", ctx, mock.Anything).Once().Return(nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
			},
		},
		{
			name: "Error on BulkPaymentRequestFileRepo.FindByPaymentFileID second call",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("d.BulkPaymentRequestFileRepo.FindByPaymentFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Error on TempFileCreator.CreateTempFile",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				objectName := "mycsvfile.csv"

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockTempFileCreator.On("CreateTempFile", objectName).Once().Return(nil, testError)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
			},
		},
		{
			name: "Error on FileStorage.DownloadFile",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("d.FileStorage.DownloadFile err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				objectName := "mycsvfile.csv"
				tempFileCreator := &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"}
				tempFile, err := tempFileCreator.CreateTempFile(objectName)
				if err != nil {
					t.Error(err)
				}

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockTempFileCreator.On("CreateTempFile", objectName).Once().Return(tempFile, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("DownloadFile", ctx, mock.Anything).Once().Return(testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.DownloadPaymentFile(testCase.ctx, testCase.req.(*invoice_pb.DownloadPaymentFileRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockInvoiceRepo,
				mockPaymentRepo,
				mockBulkPaymentRequestRepo,
				mockBulkPaymentRequestFileRepo,
				mockBulkPaymentRequestFilePaymentRepo,
				mockPartnerConvenienceStoreRepo,
				mockStudentPaymentDetailRepo,
				mockPrefectureRepo,
			)
		})
	}
}

func TestInvoiceModifierService_DownloadPaymentFile_DD(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	// Generate Invoice Mocks
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRequestRepo := new(mock_repositories.MockBulkPaymentRequestRepo)
	mockBulkPaymentRequestFileRepo := new(mock_repositories.MockBulkPaymentRequestFileRepo)
	mockBulkPaymentRequestFilePaymentRepo := new(mock_repositories.MockBulkPaymentRequestFilePaymentRepo)
	mockNewCustomerCodeRepo := new(mock_repositories.MockNewCustomerCodeHistoryRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockBankBranchRepo := new(mock_repositories.MockBankBranchRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockFileStorage := new(mock_filestorage.FileStorage)
	mockTempFileCreator := new(mock_utils.ITempFileCreator)

	mockPaymentRequestID := "test-request-id-1"
	mockPaymentRequest := &entities.BulkPaymentRequest{
		BulkPaymentRequestID: database.Text(mockPaymentRequestID),
		PaymentMethod:        database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	mockPaymentRequestFileID := "test-request-file-id-1"
	mockBulkPaymentRequestFile := &entities.BulkPaymentRequestFile{
		BulkPaymentRequestFileID: database.Text(mockPaymentRequestFileID),
		BulkPaymentRequestID:     database.Text(mockPaymentRequestID),
		FileName:                 database.Text("mytxtfile.txt"),
	}

	mockPartnerBankID := "partner-bank-id-1"
	mockPartnerBank := &entities.PartnerBank{
		PartnerBankID:    database.Text(mockPartnerBankID),
		ConsignorCode:    database.Text("1234"),
		ConsignorName:    database.Text("consigner-name"),
		BankNumber:       database.Text("1234"),
		BankName:         database.Text("bank-name-1"),
		BankBranchNumber: database.Text("123"),
		BankBranchName:   database.Text("bank-branch-name-1"),
		DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
		AccountNumber:    database.Text("1234"),
	}

	mockPaymentIDs := []string{}
	for i := 0; i < 3; i++ {
		mockPaymentIDs = append(mockPaymentIDs, fmt.Sprintf("mock-payment-id-%d", i))
	}
	mockFilePaymentInvoiceMap := genMockFilePaymentInvoiceMap(
		mockPaymentIDs,
		mockPaymentRequestFileID,
		invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
		invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
		true,
	)

	mockFilePaymentDataMap :=
		genMockFilePaymentRequestDataForDD(
			mockPaymentIDs,
			mockPaymentRequestFileID,
			invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
			invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
			true,
			mockPartnerBank,
		)

	expectedData, err := downloader.GenTxtBankContent(mockFilePaymentDataMap)
	if err != nil {
		t.Errorf("Error on generating expected txt content %v", err)
	}

	successfulResponse := &invoice_pb.DownloadPaymentFileResponse{
		Successful: true,
		Data:       expectedData,
		FileType:   invoice_pb.FileType_TXT,
	}

	testError := errors.New("test error")

	zapLogger := logger.NewZapLogger("debug", true)
	// Init service
	s := &InvoiceModifierService{
		DB:                                mockDB,
		logger:                            *zapLogger.Sugar(),
		InvoiceRepo:                       mockInvoiceRepo,
		PaymentRepo:                       mockPaymentRepo,
		BulkPaymentRequestRepo:            mockBulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:        mockBulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: mockBulkPaymentRequestFilePaymentRepo,
		NewCustomerCodeHistoryRepo:        mockNewCustomerCodeRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		BankBranchRepo:                    mockBankBranchRepo,
		UnleashClient:                     mockUnleashClient,
		FileStorage:                       mockFileStorage,
		TempFileCreator:                   mockTempFileCreator,
	}

	testcases := []TestCase{
		{
			name: "Happy case - generate txt bytes",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulResponse,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				// all have new customer code history
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Happy case - no customer code, student has updated bank account info",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulResponse,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentIDs); i++ {
					mockNewCustomerCodeRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "The payment request file ID parameter is empty",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "payment_request_file_id should not be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "BulkPaymentRequestFileRepo.FindByPaymentFileID returns error",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFileRepo.FindByID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "BulkPaymentRequestRepo.FindByPaymentRequestID returns error",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestRepo.FindByPaymentRequestID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID returns error",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("d.BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "No Student Payment Detail records",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("d.BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Empty Student Payment Detail ID",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "There is no student payment detail"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)

				mockStudentBankAccount[0].StudentPaymentDetail.StudentPaymentDetailID = database.Text("")

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Empty records for student related bank map",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("d.BankBranchRepo.FindRelatedBankOfBankBranches err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Student bank has no related bank",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "Student bank has no related bank"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					0,
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "No bank account number existing in new customer code",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulResponse,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory[0].BankAccountNumber = database.Text("421")
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockNewCustomerCodeRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "No Payment request file",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "s.BulkPaymentRequestRepo.FindByPaymentRequestID err: no rows in result set"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "The associated payment has invalid payment method",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("the payment method is not equal to the given payment method parameter")),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				invalid := genMockFilePaymentInvoiceMap(
					[]string{"payment-id-1"},
					mockPaymentRequestFileID,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					true,
				)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(invalid, nil)
			},
		},
		{
			name: "The associated payment is not yet exported",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "payment isExported field should be true"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				invalid := genMockFilePaymentInvoiceMap(
					[]string{"payment-id-1"},
					mockPaymentRequestFileID,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(invalid, nil)
			},
		},
		{
			name: "The length of invoice total amount exceeds the requirement",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The invoice total length exceeds the limit"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				invalid := genMockFilePaymentInvoiceMap(
					[]string{"payment-id-1"},
					mockPaymentRequestFileID,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					true,
				)
				invalid[0].Invoice.Total.Int = big.NewInt(int64(1231231231231231231))

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(invalid, nil)
			},
		},
		{
			name: "The length of bank number exceeds the requirement",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The partner bank number length exceeds the limit. Please check the default partner bank."),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap[0].PartnerBank.BankNumber = database.Text("1234567")

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "The length of bank branch number exceeds the requirement",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The partner bank branch number length exceeds the limit. Please check the default partner bank."),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap[0].PartnerBank.BankBranchNumber = database.Text("1234567")

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "The length of bank account number exceeds the requirement",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The partner bank account number can only accept 7 digit numbers."),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap[0].PartnerBank.AccountNumber = database.Text("12345678")

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "The length of consignor code exceeds the requirement",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedErr: status.Error(codes.Internal, "The partner bank consignor code length exceeds the limit. Please check the default partner bank."),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)
				mockNewCustomerCodeHistory := genMockCustomerCodeHistory(
					len(mockPaymentIDs),
				)
				mockRelatedBankMap[0].PartnerBank.ConsignorCode = database.Text("123456789011")

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("FindPaymentInvoiceByRequestFileID", ctx, mockDB, mock.Anything).Once().Return(mockFilePaymentInvoiceMap, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockNewCustomerCodeHistory, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Happy case - upload TXT bytes with GCloud Upload turned on",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: successfulResponse,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)

				// mock the creation of temp TXT file
				objectName := "mytxtfile.txt"
				tempFileCreator := &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"}
				tempFile, err := tempFileCreator.CreateTempFile(objectName)
				if err != nil {
					t.Error(err)
				}

				// add the expected data to the file
				_, err = tempFile.File.Write(expectedData)
				if err != nil {
					t.Error(err)
				}

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockTempFileCreator.On("CreateTempFile", objectName).Once().Return(tempFile, err)
				mockFileStorage.On("DownloadFile", ctx, mock.Anything).Once().Return(nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
			},
		},
		{
			name: "Error on TempFileCreator.CreateTempFile",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)

				// mock the creation of temp TXT file
				objectName := "mytxtfile.txt"

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockTempFileCreator.On("CreateTempFile", objectName).Once().Return(nil, testError)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
			},
		},
		{
			name: "Error on FileStorage.DownloadFile",
			ctx:  ctx,
			req: &invoice_pb.DownloadPaymentFileRequest{
				PaymentRequestFileId: mockPaymentRequestFileID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("d.FileStorage.DownloadFile err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)

				// mock the creation of temp TXT file
				objectName := "mytxtfile.txt"
				tempFileCreator := &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"}
				tempFile, err := tempFileCreator.CreateTempFile(objectName)
				if err != nil {
					t.Error(err)
				}

				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockBulkPaymentRequestRepo.On("FindByPaymentRequestID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentRequest, nil)
				mockBulkPaymentRequestFileRepo.On("FindByPaymentFileID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestFile, nil)
				mockTempFileCreator.On("CreateTempFile", objectName).Once().Return(tempFile, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("DownloadFile", ctx, mock.Anything).Once().Return(testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.DownloadPaymentFile(testCase.ctx, testCase.req.(*invoice_pb.DownloadPaymentFileRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockInvoiceRepo,
				mockPaymentRepo,
				mockBulkPaymentRequestRepo,
				mockBulkPaymentRequestFileRepo,
				mockBulkPaymentRequestFilePaymentRepo,
				mockNewCustomerCodeRepo,
				mockStudentPaymentDetailRepo,
				mockBankBranchRepo,
			)
		})
	}
}

func genMockFilePaymentInvoiceMap(paymentIDs []string, requestFileID string, paymentStatus string, paymentMethod string, isExported bool) []*entities.FilePaymentInvoiceMap {
	count := len(paymentIDs)
	mockFilePaymentInvoiceMap := make([]*entities.FilePaymentInvoiceMap, count)

	for i := 0; i < count; i++ {
		invoiceID := fmt.Sprintf("invoice-id-%d", i)
		studentID := fmt.Sprintf("student-id-%d", i)

		payment := &entities.Payment{
			PaymentID:         database.Text(fmt.Sprintf("payment-id-%d", i)),
			PaymentStatus:     database.Text(paymentStatus),
			PaymentMethod:     database.Text(paymentMethod),
			IsExported:        pgtype.Bool{Bool: isExported, Status: pgtype.Present},
			InvoiceID:         database.Text(invoiceID),
			PaymentDueDate:    database.Timestamptz(time.Now().UTC()),
			PaymentExpiryDate: database.Timestamptz(time.Now().UTC()),
			StudentID:         database.Text(studentID),
		}

		invoice := &entities.Invoice{
			InvoiceID:  database.Text(invoiceID),
			IsExported: pgtype.Bool{Bool: isExported, Status: pgtype.Present},
			Total: pgtype.Numeric{
				Int:    big.NewInt(100),
				Status: pgtype.Present,
			},
			StudentID: database.Text(studentID),
		}

		mockFilePaymentInvoiceMap[i] = &entities.FilePaymentInvoiceMap{
			Invoice: invoice,
			Payment: payment,
			BulkPaymentRequestFilePayment: &entities.BulkPaymentRequestFilePayment{
				BulkPaymentRequestFileID:        database.Text(requestFileID),
				BulkPaymentRequestFilePaymentID: database.Text(fmt.Sprintf("payment-request-file-payment-id-%d", i)),
			},
		}
	}

	return mockFilePaymentInvoiceMap
}

func genMockFilePaymentRequestDataForDD(paymentIDs []string, requestFileID string, paymentStatus string, paymentMethod string, isExported bool, partnerBank *entities.PartnerBank) []*downloader.FilePaymentDataMap {
	count := len(paymentIDs)
	mockFilePaymentInvoiceMap := make([]*entities.FilePaymentInvoiceMap, count)
	mockFilePaymentDataMap := make([]*downloader.FilePaymentDataMap, count)

	studentBankDetailsMap := make([]*entities.StudentBankDetailsMap, count)
	studentRelatedBankMap := make([]*entities.BankRelationMap, count)
	newCustomerCodeHistoryMap := make([]*entities.NewCustomerCodeHistory, count)
	bankAccountNumber := "1234567"

	for i := 0; i < count; i++ {
		invoiceID := fmt.Sprintf("invoice-id-%d", i)
		studentID := fmt.Sprintf("student-id-%d", i)
		studentPaymentDetailID := fmt.Sprintf("student-payment-detail-id-%d", i)
		bankBranchID := fmt.Sprintf("bank-branch-id-%d", i)
		bankID := fmt.Sprintf("bank-id-%d", i)
		partnerBankID := fmt.Sprintf("partner-bank-id-%d", i)

		payment := &entities.Payment{
			PaymentID:         database.Text(fmt.Sprintf("payment-id-%d", i)),
			PaymentStatus:     database.Text(paymentStatus),
			PaymentMethod:     database.Text(paymentMethod),
			IsExported:        pgtype.Bool{Bool: isExported, Status: pgtype.Present},
			InvoiceID:         database.Text(invoiceID),
			PaymentDueDate:    database.Timestamptz(time.Now().UTC()),
			PaymentExpiryDate: database.Timestamptz(time.Now().UTC()),
			StudentID:         database.Text(studentID),
		}

		invoice := &entities.Invoice{
			InvoiceID:  database.Text(invoiceID),
			IsExported: pgtype.Bool{Bool: isExported, Status: pgtype.Present},
			Total: pgtype.Numeric{
				Int:    big.NewInt(100),
				Status: pgtype.Present,
			},
			StudentID: database.Text(studentID),
		}
		studentBankDetails := &entities.StudentBankDetailsMap{
			StudentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(studentPaymentDetailID),
				StudentID:              database.Text(studentID),
				PayerName:              database.Text("test-payer-name"),
				PayerPhoneNumber:       database.Text("123-4567-890"),
			},
			BankAccount: &entities.BankAccount{
				StudentPaymentDetailID: database.Text(studentPaymentDetailID),
				BankAccountID:          database.Text(fmt.Sprintf("bank-account-id-%d", i)),
				BankAccountNumber:      database.Text(bankAccountNumber),
				BankAccountHolder:      database.Text("test-bank-account-holder"),
				BankAccountType:        database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:             database.Bool(true),
				BankBranchID:           database.Text(bankBranchID),
			},
		}
		studentRelatedBank := &entities.BankRelationMap{
			BankBranch: &entities.BankBranch{
				BankBranchID:   database.Text(bankBranchID),
				BankID:         database.Text(bankID),
				BankBranchCode: database.Text("123"),
				BankBranchName: database.Text("test-bank-branch-name"),
			},
			Bank: &entities.Bank{
				BankID:   database.Text(bankID),
				BankCode: database.Text("1234"),
				BankName: database.Text("test-bank-name"),
			},
			PartnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(partnerBankID),
				ConsignorCode:    database.Text("123123"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-partner-bank-name"),
				BankBranchNumber: database.Text("456"),
				BankBranchName:   database.Text("test-partner-bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(bankAccountNumber),
			},
		}

		newCustomerCodeHistory := &entities.NewCustomerCodeHistory{
			NewCustomerCodeHistoryID: database.Text(fmt.Sprintf("new-cc-id-%d", i)),
			NewCustomerCode:          database.Text("1"),
			StudentID:                database.Text(studentID),
			BankAccountNumber:        database.Text("123"),
		}

		studentBankDetailsMap[i] = studentBankDetails
		studentRelatedBankMap[i] = studentRelatedBank
		newCustomerCodeHistoryMap[i] = newCustomerCodeHistory

		mockFilePaymentDataMap[i] = &downloader.FilePaymentDataMap{
			Payment:                payment,
			Invoice:                invoice,
			StudentBankDetails:     studentBankDetails,
			StudentRelatedBank:     studentRelatedBank,
			NewCustomerCodeHistory: newCustomerCodeHistory,
		}

		mockFilePaymentInvoiceMap[i] = &entities.FilePaymentInvoiceMap{
			Invoice: invoice,
			Payment: payment,
			BulkPaymentRequestFilePayment: &entities.BulkPaymentRequestFilePayment{
				BulkPaymentRequestFileID:        database.Text(requestFileID),
				BulkPaymentRequestFilePaymentID: database.Text(fmt.Sprintf("payment-request-file-payment-id-%d", i)),
			},
		}
	}

	return mockFilePaymentDataMap
}

func genMockCustomerCodeHistory(count int) []*entities.NewCustomerCodeHistory {
	newCustomerCodeHistoryMap := make([]*entities.NewCustomerCodeHistory, count)

	for i := 0; i < count; i++ {
		newCustomerCodeHistory := &entities.NewCustomerCodeHistory{
			NewCustomerCodeHistoryID: database.Text(fmt.Sprintf("new-cc-id-%d", i)),
			NewCustomerCode:          database.Text("1"),
			StudentID:                database.Text(fmt.Sprintf("student-id-%d", i)),
			BankAccountNumber:        database.Text("1234567"),
		}
		newCustomerCodeHistoryMap[i] = newCustomerCodeHistory

	}

	return newCustomerCodeHistoryMap
}

func genMockFilePaymentRequestDataForCC(paymentIDs []string, requestFileID string, paymentStatus string, paymentMethod string, isExported bool) ([]*entities.FilePaymentInvoiceMap, []*downloader.FilePaymentDataMap, []*entities.StudentBillingDetailsMap) {
	count := len(paymentIDs)
	mockFilePaymentInvoiceMap := make([]*entities.FilePaymentInvoiceMap, count)
	mockFilePaymentDataMap := make([]*downloader.FilePaymentDataMap, count)
	studentBillingMap := make([]*entities.StudentBillingDetailsMap, count)

	for i := 0; i < count; i++ {
		paymentID := fmt.Sprintf("payment-id-%d", i)
		invoiceID := fmt.Sprintf("invoice-id-%d", i)
		studentID := fmt.Sprintf("student-id-%d", i)
		studentPaymentDetailID := fmt.Sprintf("student-payment-detail-id-%d", i)
		billingAddressID := fmt.Sprintf("billing-address-id-%d", i)
		requestFilePaymentID := fmt.Sprintf("payment-request-file-payment-id-%d", i)

		payment := &entities.Payment{
			PaymentID:         database.Text(paymentID),
			PaymentStatus:     database.Text(paymentStatus),
			PaymentMethod:     database.Text(paymentMethod),
			IsExported:        pgtype.Bool{Bool: isExported, Status: pgtype.Present},
			InvoiceID:         database.Text(invoiceID),
			PaymentDueDate:    database.Timestamptz(time.Now().UTC()),
			PaymentExpiryDate: database.Timestamptz(time.Now().UTC()),
			StudentID:         database.Text(studentID),
		}

		invoice := &entities.Invoice{
			InvoiceID:  database.Text(invoiceID),
			IsExported: pgtype.Bool{Bool: isExported, Status: pgtype.Present},
			Total: pgtype.Numeric{
				Int:    big.NewInt(100),
				Status: pgtype.Present,
			},
			StudentID: database.Text(studentID),
		}

		studentBillingDetailsMap := &entities.StudentBillingDetailsMap{
			StudentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(studentPaymentDetailID),
				StudentID:              database.Text(studentID),
				PayerName:              database.Text(fmt.Sprintf("payer-name-%d", i)),
				PayerPhoneNumber:       database.Text(fmt.Sprintf("payer-phone-num-%d", i)),
				PaymentMethod:          database.Text(paymentMethod),
			},
			BillingAddress: &entities.BillingAddress{
				BillingAddressID:       database.Text(billingAddressID),
				UserID:                 database.Text(studentID),
				StudentPaymentDetailID: database.Text(studentPaymentDetailID),
				PostalCode:             database.Text(fmt.Sprintf("postal-code-%d", i)),
				PrefectureCode:         database.Text(fmt.Sprintf("test-code-%d", i+1)),
				City:                   database.Text(fmt.Sprintf("city-name-%d", i)),
				Street1:                database.Text(fmt.Sprintf("street1-name-%d", i)),
				Street2:                database.Text(fmt.Sprintf("street2-name-%d", i)),
			},
		}
		studentBillingMap[i] = studentBillingDetailsMap
		mockFilePaymentDataMap[i] = &downloader.FilePaymentDataMap{
			Payment:            payment,
			Invoice:            invoice,
			StudentBillingInfo: studentBillingDetailsMap,
		}

		mockFilePaymentInvoiceMap[i] = &entities.FilePaymentInvoiceMap{
			Invoice: invoice,
			Payment: payment,
			BulkPaymentRequestFilePayment: &entities.BulkPaymentRequestFilePayment{
				BulkPaymentRequestFileID:        database.Text(requestFileID),
				BulkPaymentRequestFilePaymentID: database.Text(requestFilePaymentID),
			},
		}
	}

	return mockFilePaymentInvoiceMap, mockFilePaymentDataMap, studentBillingMap
}
