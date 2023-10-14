package openapisvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
	setupCtx     func(ctx context.Context)
	PayloadByte  []byte
}

func TestOpenAPIModifierService_ConvenienceStoreSetPaymentMethod(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := &mock_database.Tx{}
	mockPrefectureRepo := &mock_repositories.MockPrefectureRepo{}
	mockBillingAddressRepo := &mock_repositories.MockBillingAddressRepo{}
	mockStudentPaymentDetailRepo := &mock_repositories.MockStudentPaymentDetailRepo{}
	mockStudentPaymentDetailActionLogRepo := &mock_repositories.MockStudentPaymentDetailActionLogRepo{}
	mockJsm := new(mock_nats.JetStreamManagement)

	s := &OpenAPIModifierService{
		DB:                                mockDB,
		JSM:                               mockJsm,
		PrefectureRepo:                    mockPrefectureRepo,
		BillingAddressRepo:                mockBillingAddressRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		StudentPaymentDetailActionLogRepo: mockStudentPaymentDetailActionLogRepo,
		logger:                            *zap.L().Sugar(),
	}

	billingAddressInfo := &BillingAddressInfo{
		StudentID:    "test-student-id",
		PayerName:    "test-payer-name",
		PostalCode:   "test-postal-code",
		PrefectureID: "test-prefecture-id",
		City:         "test-city",
		Street1:      "test-street1",
		Street2:      "test-street2",
	}

	prefecture := &entities.Prefecture{
		ID:             database.Text("test-id"),
		PrefectureCode: database.Text("prefecture-code"),
	}

	existingStudentPaymentDetail := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("test-student-payment-detail-id"),
		StudentID:              database.Text(billingAddressInfo.StudentID),
		PayerName:              database.Text("test-payer-name-existing"),
	}

	existingBillingAddress := &entities.BillingAddress{
		StudentPaymentDetailID: existingStudentPaymentDetail.StudentPaymentDetailID,
		UserID:                 existingStudentPaymentDetail.StudentID,
		PostalCode:             database.Text("test-postal-code-existing"),
		PrefectureCode:         database.Text("test-prefecture-code-existing"),
		City:                   database.Text("test-city-existing"),
		Street1:                database.Text("test-street1-existing"),
		Street2:                database.Text("test-street2-existing"),
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name:         "success create billing address",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          billingAddressInfo,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockDB, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "success update billing address",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          billingAddressInfo,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockDB, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(existingStudentPaymentDetail, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(existingBillingAddress, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "error on student payment detail find student id",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          billingAddressInfo,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "studentPaymentDetailRepo.FindByStudentID: test error"),
			setup: func(ctx context.Context) {
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockDB, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name:         "error on prefecture find by id",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          billingAddressInfo,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "prefectureRepo.FindByPrefectureID: test error"),
			setup: func(ctx context.Context) {
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name:         "error on billing address find by user id",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          billingAddressInfo,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "billingAddressRepo.FindByUserID: test error"),
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(existingStudentPaymentDetail, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockDB, mock.Anything).Once().Return(prefecture, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name:         "error on student payment detail upsert",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          billingAddressInfo,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "student payment detail repo upsert err: test error"),
			setup: func(ctx context.Context) {
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockDB, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "error on billing address upsert",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          billingAddressInfo,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "billing address repo upsert err: test error"),
			setup: func(ctx context.Context) {
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockDB, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed test missing student id",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &BillingAddressInfo{
				StudentID:    "",
				PayerName:    "test-payer-name",
				PostalCode:   "test-postal-code",
				PrefectureID: "test-prefecture-id",
				City:         "test-city",
				Street1:      "test-street1",
				Street2:      "test-street2",
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed test missing payer name",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &BillingAddressInfo{
				StudentID:    "test-student-id",
				PayerName:    "",
				PostalCode:   "test-postal-code",
				PrefectureID: "test-prefecture-id",
				City:         "test-city",
				Street1:      "test-street1",
				Street2:      "test-street2",
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed test missing postal code",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &BillingAddressInfo{
				StudentID:    "test-student-id",
				PayerName:    "test-payer-name",
				PostalCode:   "",
				PrefectureID: "test-prefecture-id",
				City:         "test-city",
				Street1:      "test-street1",
				Street2:      "test-street2",
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed test missing prefecture id",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &BillingAddressInfo{
				StudentID:    "test-student-id",
				PayerName:    "test-payer-name",
				PostalCode:   "test-post-caode",
				PrefectureID: "",
				City:         "test-city",
				Street1:      "test-street1",
				Street2:      "test-street2",
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed test missing city",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &BillingAddressInfo{
				StudentID:    "test-student-id",
				PayerName:    "test-payer-name",
				PostalCode:   "test-post-caode",
				PrefectureID: "test-prefecture-id",
				City:         "",
				Street1:      "test-street1",
				Street2:      "test-street2",
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.AutoSetConvenienceStore(testCase.ctx, testCase.req.(*BillingAddressInfo))
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockStudentPaymentDetailRepo, mockBillingAddressRepo, mockPrefectureRepo, mockStudentPaymentDetailActionLogRepo)
		})
	}
}
