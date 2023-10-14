package paymentsvc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_filestorage "github.com/manabie-com/backend/mock/invoicemgmt/services/filestorage"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type mockPgConnError struct {
	code   string
	errMsg string
}

func (e *mockPgConnError) Error() string {
	return e.errMsg
}

func (e *mockPgConnError) Unwrap() error {
	return &pgconn.PgError{
		Code:    e.code,
		Message: e.errMsg,
	}
}

func TestPaymentModifierService_CreatePaymentRequest_CS(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)

	// Generate Invoice Mocks
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRequestRepo := new(mock_repositories.MockBulkPaymentRequestRepo)
	mockBulkPaymentRequestFileRepo := new(mock_repositories.MockBulkPaymentRequestFileRepo)
	mockBulkPaymentRequestFilePaymentRepo := new(mock_repositories.MockBulkPaymentRequestFilePaymentRepo)
	mockPartnerConvenienceStoreRepo := new(mock_repositories.MockPartnerConvenienceStoreRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockFileStorage := &mock_filestorage.FileStorage{}
	mockPrefectureRepo := new(mock_repositories.MockPrefectureRepo)
	// mock data for PENDING - CS
	mockPaymentIDs := genMockPaymentIDs()

	mockBulkPaymentRequestID := "bulk-payment-request-id-1"
	mockRequestFileID := "bulk-payment-request-file-id-1"
	mockRequestPaymentID := "bulk-payment-request-file-payment-id-1"

	mockPartnerConvenienceStore := &entities.PartnerConvenienceStore{
		PartnerConvenienceStoreID: database.Text(mockBulkPaymentRequestID),
		CompanyName:               database.Text("test-company-name"),
		CompanyTelNumber:          database.Text("123-456-789"),
		PostalCode:                database.Text("123123"),
		ManufacturerCode:          database.Int4(123456),
		CompanyCode:               database.Int4(12345),
		ShopCode:                  database.Text("123123"),
	}

	successfulResp := &invoice_pb.CreatePaymentRequestResponse{
		Successful:           true,
		BulkPaymentRequestId: mockBulkPaymentRequestID,
	}

	testError := errors.New("test error")

	// Init service
	s := &PaymentModifierService{
		DB:                                mockDB,
		InvoiceRepo:                       mockInvoiceRepo,
		PaymentRepo:                       mockPaymentRepo,
		BulkPaymentRequestRepo:            mockBulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:        mockBulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: mockBulkPaymentRequestFilePaymentRepo,
		PartnerConvenienceStoreRepo:       mockPartnerConvenienceStoreRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		UnleashClient:                     mockUnleashClient,
		FileStorage:                       mockFileStorage,
		Env:                               "local",
		PrefectureRepo:                    mockPrefectureRepo,
		TempFileCreator:                   &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"},
	}

	prefectures := genMockPrefectures()

	testcases := []TestCase{
		{
			name: "Convenience Store CSV - happy case",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - happy case with negative total amount",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)
				mockPaymentInvoice[0].Invoice.Total = database.Numeric(-13)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				// Reduce by 1 since there is one negative invoice
				for i := 0; i < len(mockPaymentInvoice)-1; i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - happy case student have multiple billing ID",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockStudentBilling = append(mockStudentBilling, genMockStudentBillingDetails(1)...)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - happy case optional street1 field",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				// Set street1 to empty
				for _, b := range mockStudentBilling {
					_ = b.BillingAddress.Street1.Set("")
				}

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - empty payment IDs",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "payment IDs cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Convenience Store CSV - empty convenience store dates",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				PaymentIds:    mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "convenience store dates cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Convenience Store CSV - empty due_date_from",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  nil,
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "due date from or until cannot be empty"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
		{
			name: "Convenience Store CSV - empty due_date_until",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: nil,
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "due date from or until cannot be empty"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
		{
			name: "Convenience Store CSV - due_date_from is ahead of due_date_until",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 5)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "due_date_from should not be ahead on due_date_until"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
		{
			name: "Convenience Store CSV - invalid payment method",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CASH,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 5)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid payment method"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Convenience Store CSV - error on saving bulk_payment_request entity",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestRepo.Create err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return("", testError)
			},
		},
		{
			name: "Convenience Store CSV - error on partnerConvenienceStoreRepo.FindOne",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.PartnerConvenienceStoreRepo.FindOne err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - returns no row on partnerConvenienceStoreRepo.FindOne",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "Partner has no associated convenience store"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(nil, pgx.ErrNoRows)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - invalid partner CS value",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "The partner CS company name is empty"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(&entities.PartnerConvenienceStore{}, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - error on paymentRepo.FindPaymentInvoiceByIDs",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.PaymentRepo.FindPaymentInvoiceByIDs err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - return payments count is not same with the given payment",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "There are payments that does not exist"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)

				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice[:1], nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - payment entity method is invalid",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "The payment method is not equal to the given payment method parameter"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - payment entity status is invalid",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "The payment status should be PENDING"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_REFUNDED.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - payment and invoice entity is already exported",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "Payment isExported field should be false"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					true,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - error on StudentPaymentDetailRepo.FindStudentBillingByStudentIDs",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.StudentPaymentDetailRepo.FindStudentBillingByStudentIDs err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - error student has no billing details",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "There is a student that does not have billing details"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling[:1], nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - error on invalid student billing details",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "The student postal code is empty"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockStudentBilling[0].BillingAddress.PostalCode = database.Text("")

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - error on BulkPaymentRequestFileRepo.Create",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFileRepo.Create err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return("", testError)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - duplicate error on BulkPaymentRequestFilePaymentRepo.Create",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("Payment with ID %s already exists in a payment request file", mockPaymentIDs[0])),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return("", &mockPgConnError{code: "23505", errMsg: "Duplicate error (SQLSTATE 23505)"})
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - duplicate error on BulkPaymentRequestFilePaymentRepo.Update",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFilePaymentRepo.Create err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)

				mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return("", testError)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - error on PaymentRepo.UpdateIsExportedByPaymentRequestFileID",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.PaymentRepo.UpdateIsExportedByPaymentRequestFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - duplicate error on InvoiceRepo.UpdateIsExportedByPaymentRequestFileID",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.InvoiceRepo.UpdateIsExportedByPaymentRequestFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - error on updating bulk payment request entity error details",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("cannot update the error details of bulk payment request err: %v", status.Error(codes.Internal, "Partner has no associated convenience store"))),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(nil, pgx.ErrNoRows)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.CreatePaymentRequest(testCase.ctx, testCase.req.(*invoice_pb.CreatePaymentRequestRequest))
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
				mockUnleashClient,
				mockFileStorage,
			)
		})
	}

}

func TestPaymentModifierService_CreatePaymentRequest_DD(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)

	// Generate Invoice Mocks
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRequestRepo := new(mock_repositories.MockBulkPaymentRequestRepo)
	mockBulkPaymentRequestFileRepo := new(mock_repositories.MockBulkPaymentRequestFileRepo)
	mockBulkPaymentRequestFilePaymentRepo := new(mock_repositories.MockBulkPaymentRequestFilePaymentRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockBankBranchRepo := new(mock_repositories.MockBankBranchRepo)
	mockNewCustomerCodeHistoryRepo := new(mock_repositories.MockNewCustomerCodeHistoryRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockFileStorage := &mock_filestorage.FileStorage{}

	// mock data for PENDING - CS
	mockPaymentIDs := genMockPaymentIDs()

	mockBulkPaymentRequestID := "bulk-payment-request-id-1"
	mockRequestFileID := "bulk-payment-request-file-id-1"
	mockRequestPaymentID := "bulk-payment-request-file-payment-id-1"

	successfulResp := &invoice_pb.CreatePaymentRequestResponse{
		Successful:           true,
		BulkPaymentRequestId: mockBulkPaymentRequestID,
	}
	partnerBankAccountNumber := "1234567"
	testError := errors.New("test error")

	// Init service
	s := &PaymentModifierService{
		DB:                                mockDB,
		InvoiceRepo:                       mockInvoiceRepo,
		PaymentRepo:                       mockPaymentRepo,
		BulkPaymentRequestRepo:            mockBulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:        mockBulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: mockBulkPaymentRequestFilePaymentRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		BankBranchRepo:                    mockBankBranchRepo,
		NewCustomerCodeHistoryRepo:        mockNewCustomerCodeHistoryRepo,
		UnleashClient:                     mockUnleashClient,
		FileStorage:                       mockFileStorage,
		Env:                               "local",
		TempFileCreator:                   &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"},
	}

	testcases := []TestCase{
		{
			name: "Direct Debit TXT - happy case",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - happy case with negative total amount",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)
				mockPaymentInvoice[0].Invoice.Total = database.Numeric(-100)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				// Reduce by 1 since there is one negative invoice
				for i := 0; i < len(mockPaymentInvoice)-1; i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - happy case student with multiple bank account",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)
				mockStudentBankAccount = append(mockStudentBankAccount, genMockStudentBankDetails(1)...)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - happy case with existing customer code",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)
				mockPaymentInvoice[0].Invoice.Total = database.Numeric(-100)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockCC := genMockCustomerCode(len(mockPaymentIDs), "1")
				mockCC[0].NewCustomerCode = database.Text("0")
				mockCC[1].NewCustomerCode = database.Text("0")
				mockCC[1].BankAccountNumber = database.Text(partnerBankAccountNumber)
				mockStudentBankAccount[2].BankAccount.BankAccountNumber = database.Text(partnerBankAccountNumber)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{mockCC[1]}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockCC, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				// Reduce by 1 since there is one negative invoice
				for i := 0; i < len(mockPaymentInvoice)-1; i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)

				// Only update twice since one customer code is already 0
				mockNewCustomerCodeHistoryRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockNewCustomerCodeHistoryRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - empty payment IDs",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "payment IDs cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Direct Debit Txt - empty direct debit dates",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				PaymentIds:    mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "direct debit dates cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Direct Debit Txt - empty due_date",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: nil,
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "due date cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Direct Debit TXT - error on saving bulk_payment_request entity",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestRepo.Create err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return("", testError)
			},
		},
		{
			name: "Convenience Store CSV - error on paymentRepo.FindPaymentInvoiceByIDs",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.PaymentRepo.FindPaymentInvoiceByIDs err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},

		{
			name: "Direct Debit TXT - error on paymentRepo.FindStudentBankDetailsByStudentIDs",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.StudentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - error on BankBranchRepo.FindRelatedBankOfBankBranches",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.BankBranchRepo.FindRelatedBankOfBankBranches err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - error on NewCustomerCodeHistoryRepo.FindByAccountNumbers",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.NewCustomerCodeHistoryRepo.FindByAccountNumbers err :%v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - error on NewCustomerCodeHistoryRepo.FindByStudentIDs",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.NewCustomerCodeHistoryRepo.FindByStudentIDs err :%v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, testError)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - student do not have bank account",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "There is a student that do not have bank account"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount[:1], nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - student bank has no related bank",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "Student bank has no related bank"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					0,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - student bank is invalid",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "The bank code is empty"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap[0].Bank.BankCode = database.Text("")

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - payment entity method is invalid",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "The payment method is not equal to the given payment method parameter"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - payment entity status is invalid",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "The payment status should be PENDING"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_REFUNDED.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - payment entity is already exported",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, "Payment isExported field should be false"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					true,
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - error on BulkPaymentRequestFileRepo.Create",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFileRepo.Create err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return("", testError)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - duplicate error on PaymentRepo.UpdateIsExportedByPaymentRequestFileID",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.PaymentRepo.UpdateIsExportedByPaymentRequestFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - duplicate error on InvoiceRepo.UpdateIsExportedByPaymentRequestFileID",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.InvoiceRepo.UpdateIsExportedByPaymentRequestFileID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Direct Debit TXT - duplicate error on NewCustomerCodeHistoryRepo.Create",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("g.NewCustomerCodeHistoryRepo.Create err %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)

				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.CreatePaymentRequest(testCase.ctx, testCase.req.(*invoice_pb.CreatePaymentRequestRequest))
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
				mockUnleashClient,
				mockFileStorage,
			)
		})
	}
}

func TestPaymentModifierService_CreatePaymentRequest_CS_EnableGcloud(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRequestRepo := new(mock_repositories.MockBulkPaymentRequestRepo)
	mockBulkPaymentRequestFileRepo := new(mock_repositories.MockBulkPaymentRequestFileRepo)
	mockPartnerConvenienceStoreRepo := new(mock_repositories.MockPartnerConvenienceStoreRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockPrefectureRepo := new(mock_repositories.MockPrefectureRepo)
	mockBulkPaymentRepo := new(mock_repositories.MockBulkPaymentRepo)
	mockFileStorage := &mock_filestorage.FileStorage{}

	s := &PaymentModifierService{
		DB:                          mockDB,
		InvoiceRepo:                 mockInvoiceRepo,
		PaymentRepo:                 mockPaymentRepo,
		BulkPaymentRequestRepo:      mockBulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:  mockBulkPaymentRequestFileRepo,
		PartnerConvenienceStoreRepo: mockPartnerConvenienceStoreRepo,
		StudentPaymentDetailRepo:    mockStudentPaymentDetailRepo,
		UnleashClient:               mockUnleashClient,
		Env:                         "local",
		PrefectureRepo:              mockPrefectureRepo,
		BulkPaymentRepo:             mockBulkPaymentRepo,
		FileStorage:                 mockFileStorage,
		TempFileCreator:             &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"},
	}

	mockPaymentIDs := genMockPaymentIDs()

	mockBulkPaymentRequestID := "bulk-payment-request-id-99"
	mockRequestFileID := "bulk-payment-request-file-id-99"

	mockPartnerConvenienceStore := genMockConvenienceStore(mockBulkPaymentRequestID)

	prefectures := genMockPrefectures()

	testError := errors.New("test error")

	mockBulkPayment := &entities.BulkPayment{
		BulkPaymentID:     database.Text("123"),
		BulkPaymentStatus: database.Text(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING.String()),
	}

	testcases := []TestCase{
		// other test cases are tested on existing unit test
		{
			name: "happy case - Convenience Store CSV store gcloud",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: &invoice_pb.CreatePaymentRequestResponse{
				Successful:           true,
				BulkPaymentRequestId: mockBulkPaymentRequestID,
			},
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "happy case - Convenience Store CSV store gcloud with payment belongs to bulk payment",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: &invoice_pb.CreatePaymentRequestResponse{
				Successful:           true,
				BulkPaymentRequestId: mockBulkPaymentRequestID,
			},
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentInvoice[0].Payment.BulkPaymentID = mockBulkPayment.BulkPaymentID

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - update invoice exported",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error InvoiceRepo UpdateIsExportedByInvoiceIDs: %v", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - update payment exported",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error PaymentRepo UpdateIsExportedByPaymentIDs: %v", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - update bulk payment status",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error BulkPaymentRepo UpdateBulkPaymentStatusByIDs: %v", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentInvoice[0].Payment.BulkPaymentID = mockBulkPayment.BulkPaymentID

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - uploading csv file on gcloud error",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("objectUploader.DoUploadFile error: UploadFile error: %v with object name: mock.Anything", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - Convenience Store error feature flag",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableGCloudUploadFeatureFlag, testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Times(2).Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.CreatePaymentRequest(testCase.ctx, testCase.req.(*invoice_pb.CreatePaymentRequestRequest))
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
				mockPartnerConvenienceStoreRepo,
				mockUnleashClient,
				mockPrefectureRepo,
				mockStudentPaymentDetailRepo,
				mockFileStorage,
				mockBulkPaymentRepo,
			)
		})
	}
}

func TestPaymentModifierService_CreatePaymentRequest_DD_EnableGcloud(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRequestRepo := new(mock_repositories.MockBulkPaymentRequestRepo)
	mockBulkPaymentRequestFileRepo := new(mock_repositories.MockBulkPaymentRequestFileRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockBankBranchRepo := new(mock_repositories.MockBankBranchRepo)
	mockNewCustomerCodeHistoryRepo := new(mock_repositories.MockNewCustomerCodeHistoryRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockFileStorage := &mock_filestorage.FileStorage{}
	mockBulkPaymentRepo := new(mock_repositories.MockBulkPaymentRepo)

	mockPaymentIDs := genMockPaymentIDs()

	s := &PaymentModifierService{
		DB:                         mockDB,
		InvoiceRepo:                mockInvoiceRepo,
		PaymentRepo:                mockPaymentRepo,
		BulkPaymentRequestRepo:     mockBulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo: mockBulkPaymentRequestFileRepo,
		StudentPaymentDetailRepo:   mockStudentPaymentDetailRepo,
		BankBranchRepo:             mockBankBranchRepo,
		NewCustomerCodeHistoryRepo: mockNewCustomerCodeHistoryRepo,
		UnleashClient:              mockUnleashClient,
		FileStorage:                mockFileStorage,
		TempFileCreator:            &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"},
		BulkPaymentRepo:            mockBulkPaymentRepo,
	}

	mockBulkPaymentRequestID := "bulk-payment-request-id-100"
	mockRequestFileID := "bulk-payment-request-file-id-100"
	testError := errors.New("test error")

	successfulResp := &invoice_pb.CreatePaymentRequestResponse{
		Successful:           true,
		BulkPaymentRequestId: mockBulkPaymentRequestID,
	}

	mockBulkPayment := &entities.BulkPayment{
		BulkPaymentID:     database.Text("123"),
		BulkPaymentStatus: database.Text(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING.String()),
	}

	testcases := []TestCase{
		{
			name: "Direct Debit TXT - happy case",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "Direct Debit TXT - happy case with payment belongs to bulk payment",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentInvoice[0].Payment.BulkPaymentID = mockBulkPayment.BulkPaymentID

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - uploading csv file on gcloud error",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("objectUploader.DoUploadFile error: UploadFile error: %v with object name: mock.Anything", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - update invoice exported",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error InvoiceRepo UpdateIsExportedByInvoiceIDs: %v", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentInvoice[0].Payment.BulkPaymentID = mockBulkPayment.BulkPaymentID

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - update payment exported",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error PaymentRepo UpdateIsExportedByPaymentIDs: %v", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentInvoice[0].Payment.BulkPaymentID = mockBulkPayment.BulkPaymentID

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - update bulk payment status",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error BulkPaymentRepo UpdateBulkPaymentStatusByIDs: %v", testError)),
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
					false,
				)

				mockStudentBankAccount := genMockStudentBankDetails(
					len(mockPaymentIDs),
				)

				mockRelatedBankMap := genMockBankRelationMap(
					len(mockPaymentIDs),
				)

				mockPaymentInvoice[0].Payment.BulkPaymentID = mockBulkPayment.BulkPaymentID

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBankDetailsByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBankAccount, nil)
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", ctx, mockDB, mock.Anything).Once().Return(mockRelatedBankMap, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByAccountNumbers", ctx, mockDB, mock.Anything).Once().Return([]*entities.NewCustomerCodeHistory{}, nil)
				mockNewCustomerCodeHistoryRepo.On("FindByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(nil, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockNewCustomerCodeHistoryRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name: "negative case - Direct Debit error feature flag",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DirectDebitDates: &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{
					DueDate: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableGCloudUploadFeatureFlag, testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, testError)
				mockBulkPaymentRequestRepo.On("Update", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.CreatePaymentRequest(testCase.ctx, testCase.req.(*invoice_pb.CreatePaymentRequestRequest))
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
				mockStudentPaymentDetailRepo,
				mockBankBranchRepo,
				mockNewCustomerCodeHistoryRepo,
				mockUnleashClient,
				mockFileStorage,
			)
		})
	}
}

func TestPaymentModifierService_CreatePaymentRequest_CS_OptionalDueDates(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)

	// Generate Invoice Mocks
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRequestRepo := new(mock_repositories.MockBulkPaymentRequestRepo)
	mockBulkPaymentRequestFileRepo := new(mock_repositories.MockBulkPaymentRequestFileRepo)
	mockBulkPaymentRequestFilePaymentRepo := new(mock_repositories.MockBulkPaymentRequestFilePaymentRepo)
	mockPartnerConvenienceStoreRepo := new(mock_repositories.MockPartnerConvenienceStoreRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockFileStorage := &mock_filestorage.FileStorage{}
	mockPrefectureRepo := new(mock_repositories.MockPrefectureRepo)
	// mock data for PENDING - CS
	mockPaymentIDs := genMockPaymentIDs()

	mockBulkPaymentRequestID := "bulk-payment-request-id-1"
	mockRequestFileID := "bulk-payment-request-file-id-1"
	mockRequestPaymentID := "bulk-payment-request-file-payment-id-1"

	mockPartnerConvenienceStore := &entities.PartnerConvenienceStore{
		PartnerConvenienceStoreID: database.Text(mockBulkPaymentRequestID),
		CompanyName:               database.Text("test-company-name"),
		CompanyTelNumber:          database.Text("123-456-789"),
		PostalCode:                database.Text("123123"),
		ManufacturerCode:          database.Int4(123456),
		CompanyCode:               database.Int4(12345),
		ShopCode:                  database.Text("123123"),
	}

	successfulResp := &invoice_pb.CreatePaymentRequestResponse{
		Successful:           true,
		BulkPaymentRequestId: mockBulkPaymentRequestID,
	}

	// Init service
	s := &PaymentModifierService{
		DB:                                mockDB,
		InvoiceRepo:                       mockInvoiceRepo,
		PaymentRepo:                       mockPaymentRepo,
		BulkPaymentRequestRepo:            mockBulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:        mockBulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: mockBulkPaymentRequestFilePaymentRepo,
		PartnerConvenienceStoreRepo:       mockPartnerConvenienceStoreRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		UnleashClient:                     mockUnleashClient,
		FileStorage:                       mockFileStorage,
		Env:                               "local",
		PrefectureRepo:                    mockPrefectureRepo,
		TempFileCreator:                   &utils.TempFileCreator{TempDirPattern: "invoicemgmt-unit-test"},
	}

	prefectures := genMockPrefectures()

	// other test cases are covered on other unit tests already
	testcases := []TestCase{
		{
			name: "Convenience Store CSV - happy case - optional convenience store due dates",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  nil,
					DueDateUntil: nil,
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - happy case with due dates",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 5)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - empty due_date_from",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  nil,
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - empty due_date_until",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 1)),
					DueDateUntil: nil,
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)

				for i := 0; i < len(mockPaymentInvoice); i++ {
					mockBulkPaymentRequestFilePaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestPaymentID, nil)
				}

				mockPaymentRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByPaymentRequestFileID", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Convenience Store CSV - due_date_from is greater the due_date_until",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  timestamppb.New(time.Now().AddDate(0, 0, 5)),
					DueDateUntil: timestamppb.New(time.Now().AddDate(0, 0, 1)),
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedErr: status.Error(codes.InvalidArgument, "due_date_from should not be ahead on due_date_until"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "happy case - Convenience Store CSV store gcloud with kec feature flag",
			ctx:  ctx,
			req: &invoice_pb.CreatePaymentRequestRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{
					DueDateFrom:  nil,
					DueDateUntil: nil,
				},
				PaymentIds: mockPaymentIDs,
			},
			expectedResp: &invoice_pb.CreatePaymentRequestResponse{
				Successful:           true,
				BulkPaymentRequestId: mockBulkPaymentRequestID,
			},
			setup: func(ctx context.Context) {
				mockPaymentInvoice := genMockPaymentInvoice(
					mockPaymentIDs,
					invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
					false,
				)

				mockStudentBilling := genMockStudentBillingDetails(
					len(mockPaymentIDs),
				)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)

				mockBulkPaymentRequestRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentRequestID, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableGCloudUploadFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableOptionalValidationInPaymentRequest, mock.Anything).Once().Return(true, nil)

				mockPartnerConvenienceStoreRepo.On("FindOne", ctx, mockDB).Once().Return(mockPartnerConvenienceStore, nil)
				mockPrefectureRepo.On("FindAll", ctx, mockDB).Once().Return(prefectures, nil)
				mockPaymentRepo.On("FindPaymentInvoiceByIDs", ctx, mockDB, mock.Anything).Once().Return(mockPaymentInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindStudentBillingByStudentIDs", ctx, mockDB, mock.Anything).Once().Return(mockStudentBilling, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockFileStorage.On("FormatObjectName", mock.Anything).Once().Return(mock.Anything)
				mockFileStorage.On("GetDownloadURL", mock.Anything).Once().Return(mock.Anything)
				mockBulkPaymentRequestFileRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mockRequestFileID, nil)
				mockPaymentRepo.On("UpdateIsExportedByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateIsExportedByInvoiceIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockFileStorage.On("UploadFile", ctx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)

			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.CreatePaymentRequest(testCase.ctx, testCase.req.(*invoice_pb.CreatePaymentRequestRequest))
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
				mockUnleashClient,
				mockFileStorage,
			)
		})
	}

}

func genMockConvenienceStore(mockBulkPaymentRequestID string) *entities.PartnerConvenienceStore {
	mockPartnerConvenienceStore := &entities.PartnerConvenienceStore{
		PartnerConvenienceStoreID: database.Text(mockBulkPaymentRequestID),
		CompanyName:               database.Text("test-company-name"),
		CompanyTelNumber:          database.Text("123-456-789"),
		PostalCode:                database.Text("123123"),
		ManufacturerCode:          database.Int4(123456),
		CompanyCode:               database.Int4(12345),
		ShopCode:                  database.Text("123123"),
	}

	return mockPartnerConvenienceStore
}

func genMockPrefectures() []*entities.Prefecture {
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

	return prefectures
}

func genMockStudentBillingDetails(count int) []*entities.StudentBillingDetailsMap {
	mockData := make([]*entities.StudentBillingDetailsMap, count)

	for i := 0; i < count; i++ {
		studentID := fmt.Sprintf("student-id-%d", i)
		studentPaymentDetailID := fmt.Sprintf("student-payment-detail-id-%d", i)
		billingAddressiD := fmt.Sprintf("billing-address-id-%d", i)

		mockData[i] = &entities.StudentBillingDetailsMap{
			StudentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(studentPaymentDetailID),
				StudentID:              database.Text(studentID),
				PayerName:              database.Text("test-payer-name"),
				PayerPhoneNumber:       database.Text("123-4567-890"),
				PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			},
			BillingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(billingAddressiD),
				UserID:           database.Text(studentID),
				PostalCode:       database.Text("23213"),
				PrefectureCode:   database.Text(fmt.Sprintf("test-code-%d", i+1)),
				City:             database.Text("test-city"),
				Street1:          database.Text("test-street1"),
				Street2:          database.Text("test-street2"),
			},
		}

	}

	return mockData
}

func genMockStudentBankDetails(count int) []*entities.StudentBankDetailsMap {
	mockData := make([]*entities.StudentBankDetailsMap, count)

	for i := 0; i < count; i++ {
		studentID := fmt.Sprintf("student-id-%d", i)
		studentPaymentDetailID := fmt.Sprintf("student-payment-detail-id-%d", i)
		bankAccountID := fmt.Sprintf("bank-account-id-%d", i)
		bankBranchID := fmt.Sprintf("bank-branch-id-%d", i)

		mockData[i] = &entities.StudentBankDetailsMap{
			StudentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(studentPaymentDetailID),
				StudentID:              database.Text(studentID),
				PayerName:              database.Text("test-payer-name"),
				PayerPhoneNumber:       database.Text("123-4567-890"),
				PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
			},
			BankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(bankAccountID),
				BankAccountNumber: database.Text("1234567"),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:        database.Bool(true),
				BankBranchID:      database.Text(bankBranchID),
			},
		}

	}

	return mockData
}

func genMockBankRelationMap(count int) []*entities.BankRelationMap {
	mockData := make([]*entities.BankRelationMap, count)

	for i := 0; i < count; i++ {
		bankBranchID := fmt.Sprintf("bank-branch-id-%d", i)
		bankID := fmt.Sprintf("bank-id-%d", i)
		partnerBankID := fmt.Sprintf("partner-bank-id-%d", i)

		mockData[i] = &entities.BankRelationMap{
			BankBranch: &entities.BankBranch{
				BankBranchID:   database.Text(bankBranchID),
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
				AccountNumber:    database.Text("1234567"),
			},
		}
	}

	return mockData
}

func genMockPaymentInvoice(paymentIDs []string, paymentStatus string, paymentMethod string, isExported bool) []*entities.PaymentInvoiceMap {
	count := len(paymentIDs)

	mockData := make([]*entities.PaymentInvoiceMap, count)

	for i := 0; i < count; i++ {
		paymentID := fmt.Sprintf("payment-id-%d", i)
		invoiceID := fmt.Sprintf("invoice-id-%d", i)
		studentID := fmt.Sprintf("student-id-%d", i)

		mockData[i] = &entities.PaymentInvoiceMap{
			Payment: &entities.Payment{
				PaymentID:     database.Text(paymentID),
				PaymentStatus: database.Text(paymentStatus),
				PaymentMethod: database.Text(paymentMethod),
				IsExported:    pgtype.Bool{Bool: isExported, Status: pgtype.Present},
				InvoiceID:     database.Text(invoiceID),
				StudentID:     database.Text(studentID),
			},
			Invoice: &entities.Invoice{
				InvoiceID:  database.Text(invoiceID),
				IsExported: pgtype.Bool{Bool: isExported, Status: pgtype.Present},
				StudentID:  database.Text(studentID),
				Total:      database.Numeric(100),
			},
		}

	}

	return mockData
}

func genMockPaymentBankAndInvoiceData(paymentIDs []string, paymentStatus string, paymentMethod string, isExported bool) ([]*entities.PaymentBankMap, []*entities.Invoice) {
	count := len(paymentIDs)

	mockPendingCSPayments := make([]*entities.PaymentBankMap, count)
	mockPendingCSPaymentInvoices := make([]*entities.Invoice, count)

	for i := 0; i < count; i++ {
		paymentID := fmt.Sprintf("payment-id-%d", i)
		invoiceID := fmt.Sprintf("invoice-id-%d", i)

		paymentBank := &entities.PaymentBankMap{
			Payment: &entities.Payment{
				PaymentID:     database.Text(paymentID),
				PaymentStatus: database.Text(paymentStatus),
				PaymentMethod: database.Text(paymentMethod),
				IsExported:    pgtype.Bool{Bool: isExported, Status: pgtype.Present},
				InvoiceID:     database.Text(invoiceID),
			},
			Bank: &entities.Bank{
				BankID:   database.Text("bank-id-%d"),
				BankName: database.Text("Bank-A"),
			},
		}
		invoice := &entities.Invoice{
			InvoiceID:  database.Text(invoiceID),
			IsExported: pgtype.Bool{Bool: isExported, Status: pgtype.Present},
		}

		mockPendingCSPayments[i] = paymentBank
		mockPendingCSPaymentInvoices[i] = invoice

	}

	return mockPendingCSPayments, mockPendingCSPaymentInvoices
}

func genMockCustomerCode(count int, customerCode string) []*entities.NewCustomerCodeHistory {
	mockData := make([]*entities.NewCustomerCodeHistory, count)

	for i := 0; i < count; i++ {
		studentID := fmt.Sprintf("student-id-%d", i)
		ccID := fmt.Sprintf("cc-di-%d", i)

		mockData[i] = &entities.NewCustomerCodeHistory{
			NewCustomerCodeHistoryID: database.Text(ccID),
			BankAccountNumber:        database.Text(fmt.Sprintf("%d", i)),
			StudentID:                database.Text(studentID),
		}
	}

	return mockData
}

func genMockPaymentIDs() []string {
	mockPaymentIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		mockPaymentIDs[i] = fmt.Sprintf("payment-id-%d", i)
	}

	return mockPaymentIDs
}
