package payment_detail

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	utils "github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name                string
	ctx                 context.Context
	req                 interface{}
	expectedResp        interface{}
	expectedErr         error
	setup               func(ctx context.Context)
	mockInvoiceEntities []*entities.Invoice
	setupCtx            func(ctx context.Context)
}

func TestEditPaymentDetailService_UpsertStudentPaymentInfo(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	successfulResp := &invoice_pb.UpsertStudentPaymentInfoResponse{
		Successful: true,
	}

	mockDB := new(mock_database.Ext)
	mockTx := &mock_database.Tx{}
	mockPrefectureRepo := &mock_repositories.MockPrefectureRepo{}
	mockBillingAddressRepo := &mock_repositories.MockBillingAddressRepo{}
	mockStudentPaymentDetailRepo := &mock_repositories.MockStudentPaymentDetailRepo{}
	mockBankRepo := &mock_repositories.MockBankRepo{}
	mockBankBranchRepo := &mock_repositories.MockBankBranchRepo{}
	mockBankAccountRepo := &mock_repositories.MockBankAccountRepo{}
	mockStudentPaymentDetailActionLogRepo := &mock_repositories.MockStudentPaymentDetailActionLogRepo{}

	s := &EditPaymentDetailService{
		DB:                                mockDB,
		PrefectureRepo:                    mockPrefectureRepo,
		BillingAddressRepo:                mockBillingAddressRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		BankRepo:                          mockBankRepo,
		BankBranchRepo:                    mockBankBranchRepo,
		BankAccountRepo:                   mockBankAccountRepo,
		Logger:                            zap.NewNop().Sugar(),
		StudentPaymentDetailActionLogRepo: mockStudentPaymentDetailActionLogRepo,
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name: "failed to process the request that missing student id",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "student id to upsert payment info can not be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed to create because request has both billing information and bank account can not be empty",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId:       "example-student-id",
				BillingInfo:     nil,
				BankAccountInfo: nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "both billing information and bank account can not be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed to create because request has empty billing address and empty bank account",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					BillingAddress: nil,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "both billing information and bank account can not be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed to create because request has empty payer name",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName: "",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "payer name can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create because request has empty postal code",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName: "example-payer-name",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "postal code can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create because request has empty prefecture",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName: "example-payer-name",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "prefecture can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create because request has empty city",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName: "example-payer-name",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "city can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create because request has non-existing prefecture code",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName: "example-payer-name",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "non-existing-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "prefecture code does not exist"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create more student payment detail",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "example-payer-name",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "student payment detail already exists, can not create more"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create more student payment detail with empty payer name in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "payer name can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create more billing address",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "example-payer-name",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "billing address already exists, can not create more"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create more billing address with empty postal code in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "example-payer-name",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "postal code can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create a new bank account with empty bank id in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank id can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with empty bank branch id in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank branch id can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with empty bank account number in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create a new bank account with bank account number is not equal to 7 in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "123",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number only can accept 7 digit numbers"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with empty bank account holder in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "example-bank-account-number",
					BankAccountHolder: "",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account holder can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with invalid bank account holder in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "eXAMPLE - ｱ BRANCH - ｢123｣ ()  ﾟ ﾞ . ﾆ",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  utils.InvalidHalfWidthKanaBankHolder,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create a new bank account with empty bank account type in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-number",
					BankAccountType:   3,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account type can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create a new bank account with empty bank account number in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with bank account number is not equal to 7 in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "123",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number only can accept 7 digit numbers"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create a new bank account with empty bank account number in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "example-bank-account-number",
					BankAccountHolder: "",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account holder can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create a new bank account with empty bank account type in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   3,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account type can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with empty bank id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank id can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to create a new bank account with empty bank branch id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank branch id can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with non-existing bank id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "non-existing-bank-branch-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "bank id does not exist"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create a new bank account with non-existing bank branch id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "",
					BankId:            "existing-bank-id",
					BankBranchId:      "non-existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "bank branch id does not exist"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Bank{}, nil)
				mockBankBranchRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update a non-existing student payment id in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "non-existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "student payment detail does not exist, can not update"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update an existing student payment with empty payer name in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "non-existing-student-payment-id",
					PayerName:              "",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "payer name can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update a non-existing billing address id in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "billing address does not exist, can not update"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update an existing billing address with empty postal code in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "postal code can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update an existing billing address with non-existing prefecture code in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "non-existing-prefecture-code",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "prefecture code does not exist"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update an existing bank account with empty bank account number in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update an existing bank account with bank account number is not equal to 7 in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "123",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number only can accept 7 digit numbers"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to update an existing bank account with empty bank account holder in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "example-bank-account-number",
					BankAccountHolder: "",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account holder can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to update an existing bank account with invalid bank account holder in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE BANK - ｱ BRANCH - <123> ()  ﾟ ﾞ . ﾆ",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  utils.InvalidHalfWidthKanaBankHolder,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to update an existing bank account with empty bank account type in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "existing-student-payment-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "non-existing-student-payment-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-number",
					BankAccountType:   3,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account type can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to update existing bank account with empty bank account number in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankBranchId:      "existing-bank-branch-id",
					BankId:            "existing-bank-id",
					BankAccountNumber: "",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to update existing bank account with bank account number is not equal to 7 in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "123",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account number only can accept 7 digit numbers"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to update existing bank account with empty bank account number in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "example-bank-account-number",
					BankAccountHolder: "",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account holder can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update existing bank account with empty bank account type in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   3,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank account type can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update existing bank account with empty bank id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank id can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to update existing bank account with empty bank branch id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "",
					BankAccountNumber: "1234567",
					BankAccountHolder: "example-bank-account-holder",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "bank branch id can not be empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update existing bank account with non-existing bank id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "non-existing-bank-branch-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "bank id does not exist"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "failed to update existing bank account with non-existing bank branch id in request when billing information info is missing",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "non-existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_CHECKING_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "bank branch id does not exist"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Bank{}, nil)
				mockBankBranchRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed to create student payment detail action log",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "example-student-payment-detail-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "example-billing-address-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "err cannot create student payment detail action log: test error"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "create successfully with a request has valid billing information",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "example-payer-name",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "create successfully with a request has valid billing information and valid bank account",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "example-payer-name",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_SAVINGS_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Bank{}, nil)
				mockBankBranchRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BankBranch{}, nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "create successfully with a request has valid billing information and valid bank account, but bank account is unverified",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "example-payer-name",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_SAVINGS_ACCOUNT,
					IsVerified:        false,
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "create successfully with a request has valid billing information and valid bank account, but bank account is unverified",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					PayerName:        "example-payer-name",
					PayerPhoneNumber: "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						PostalCode:     "example-postal-code",
						PrefectureCode: "example-prefecture",
						City:           "example-city",
						Street1:        "example-street-1",
						Street2:        "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_SAVINGS_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Bank{}, nil)
				mockBankBranchRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BankBranch{}, nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name: "update only billing information successfully with a valid request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "example-student-payment-detail-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "example-billing-address-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "update only bank account successfully with a valid request, bank account status is unverified",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_SAVINGS_ACCOUNT,
					IsVerified:        false,
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBankAccountRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BankAccount{}, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "update only bank account successfully with a valid request, bank account status is verified",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_SAVINGS_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Bank{}, nil)
				mockBankBranchRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BankBranch{}, nil)
				mockBankAccountRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BankAccount{}, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "update both billing address and bank account successfully with a valid request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertStudentPaymentInfoRequest{
				StudentId: "example-student-id",
				BillingInfo: &invoice_pb.BillingInformation{
					StudentPaymentDetailId: "example-student-payment-detail-id",
					PayerName:              "example-payer-name",
					PayerPhoneNumber:       "example-payer-phone-number",
					BillingAddress: &invoice_pb.BillingAddress{
						BillingAddressId: "example-billing-address-id",
						PostalCode:       "example-postal-code",
						PrefectureCode:   "example-prefecture",
						City:             "example-city",
						Street1:          "example-street-1",
						Street2:          "example-street-2",
					},
				},
				BankAccountInfo: &invoice_pb.BankAccountInformation{
					BankAccountId:     "existing-bank-account-id",
					BankId:            "existing-bank-id",
					BankBranchId:      "existing-bank-branch-id",
					BankAccountNumber: "1234567",
					BankAccountHolder: "EXAMPLE-BANK-ACCOUNT-HOLDER",
					BankAccountType:   invoice_pb.BankAccountType_SAVINGS_ACCOUNT,
					IsVerified:        true,
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureCode", ctx, mockDB, mock.Anything).Once().Return(&entities.Prefecture{}, nil)
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.StudentPaymentDetail{}, nil)
				mockBillingAddressRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BillingAddress{}, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Bank{}, nil)
				mockBankBranchRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BankBranch{}, nil)
				mockBankAccountRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.BankAccount{}, nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.UpsertStudentPaymentInfo(testCase.ctx, testCase.req.(*invoice_pb.UpsertStudentPaymentInfoRequest))
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockStudentPaymentDetailRepo, mockBillingAddressRepo, mockPrefectureRepo, mockBankRepo, mockBankBranchRepo, mockBankAccountRepo, mockStudentPaymentDetailActionLogRepo)
		})
	}
}

func TestValidateBankHolder(t *testing.T) {
	t.Parallel()

	validNumber := `0123456789`
	validCapitalCharacters := `ABCDEFGHIJKLMNOPQRSTUVWXYZ`
	validHalfWidthKana := `ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ`
	validSymbols := `ﾞﾟ()｢｣/-.\ `

	testcases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "valid number",
			input:       validNumber,
			expectedErr: nil,
		},
		{
			name:        "valid capital alphabet",
			input:       validCapitalCharacters,
			expectedErr: nil,
		},
		{
			name:        "valid half-width kana",
			input:       validHalfWidthKana,
			expectedErr: nil,
		},
		{
			name:        "valid symbols",
			input:       validSymbols,
			expectedErr: nil,
		},
		{
			name:        "valid combination",
			input:       validNumber + validCapitalCharacters + validHalfWidthKana + validSymbols,
			expectedErr: nil,
		},
		{
			name:        "valid input",
			input:       "EXAMPLE BANK - ｱ BRANCH - ｢123｣ ()  ﾟ ﾞ . ﾆ",
			expectedErr: nil,
		},
		{
			name:        "invalid lowercase alphabet",
			input:       "abcdefghijklmnopqrstuvwxyz",
			expectedErr: utils.InvalidHalfWidthKanaBankHolder,
		},
		{
			name:        "invalid small kana",
			input:       "ｧｨｩｪｫｯｬｭｮ",
			expectedErr: utils.InvalidHalfWidthKanaBankHolder,
		},
		{
			name:        "invalid symbols",
			input:       `~!@#$%^&*+=[]{}|;:'",<>./?`,
			expectedErr: utils.InvalidHalfWidthKanaBankHolder,
		},
		{
			name:        "invalid input",
			input:       "eXAMPLE BANK - ｱ BRANCH - ｢123｣ ()  ﾟ ﾞ . ﾆ",
			expectedErr: utils.InvalidHalfWidthKanaBankHolder,
		},
		{
			name:        "invalid input",
			input:       "EXAMPLE BANK - ｱ BRANCH - <123> ()  ﾟ ﾞ . ﾆ",
			expectedErr: utils.InvalidHalfWidthKanaBankHolder,
		},
		{
			name:        "invalid input",
			input:       "EXAMPLE BANK - ｧ BRANCH - ｢123｣ ()  ﾟ ﾞ . ﾆ",
			expectedErr: utils.InvalidHalfWidthKanaBankHolder,
		},
		{
			name:        "valid input",
			input:       "EXAMPLE BANK - ｱ BRANCH - ｢123｣ []  ﾟ ﾞ . ﾆ",
			expectedErr: utils.InvalidHalfWidthKanaBankHolder,
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {

			err := utils.ValidateBankHolder(testCase.input)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, utils.InvalidHalfWidthKanaBankHolder, err)
			}
		})
	}
}
