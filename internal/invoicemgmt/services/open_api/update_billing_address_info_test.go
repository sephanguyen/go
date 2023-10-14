package openapisvc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOpenAPIModifierService_AutoUpdateBillingAddressInfoAndPaymentDetail(t *testing.T) {
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
	mockBankAccountRepo := &mock_repositories.MockBankAccountRepo{}
	mockJsm := new(mock_nats.JetStreamManagement)

	zapLogger := logger.NewZapLogger("debug", true)
	s := &OpenAPIModifierService{
		DB:                                mockDB,
		JSM:                               mockJsm,
		PrefectureRepo:                    mockPrefectureRepo,
		BillingAddressRepo:                mockBillingAddressRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		StudentPaymentDetailActionLogRepo: mockStudentPaymentDetailActionLogRepo,
		BankAccountRepo:                   mockBankAccountRepo,
		logger:                            *zapLogger.Sugar(),
	}

	prefectureID := "test-prefecture-id-1"
	studentPaymentDetailID := "test-student-payment-detail-id-1"
	studentID := "test-student-id-1"

	existingBillingAddress := generateBillingAddress(studentPaymentDetailID, studentID)
	existingPaymentDetail := generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())

	// param for updated billing address
	info1 := &UpdateBillingAddressEventInfo{
		StudentID: studentID,
		PayerName: existingPaymentDetail.PayerName.String,
		UserAddress: &upb.UserAddress{
			PostalCode:   "updated-postal-code",
			Prefecture:   "updated-prefecture-id",
			City:         "updated-test-city",
			FirstStreet:  "updated-test-street1",
			SecondStreet: "updated-test-street2",
		},
	}

	// param for updated payer name
	info2 := &UpdateBillingAddressEventInfo{
		StudentID: studentID,
		PayerName: "updated-test-payer-name-existing",
		UserAddress: &upb.UserAddress{
			PostalCode:   existingBillingAddress.PostalCode.String,
			Prefecture:   prefectureID,
			City:         existingBillingAddress.City.String,
			FirstStreet:  existingBillingAddress.Street1.String,
			SecondStreet: existingBillingAddress.Street2.String,
		},
	}

	// param for no updates
	info3 := &UpdateBillingAddressEventInfo{
		StudentID: studentID,
		PayerName: existingPaymentDetail.PayerName.String,
		UserAddress: &upb.UserAddress{
			PostalCode:   existingBillingAddress.PostalCode.String,
			Prefecture:   prefectureID,
			City:         existingBillingAddress.City.String,
			FirstStreet:  existingBillingAddress.Street1.String,
			SecondStreet: existingBillingAddress.Street2.String,
		},
	}

	// param for updated billing address
	info4 := &UpdateBillingAddressEventInfo{
		StudentID: studentID,
		PayerName: "updated-payer-name",
		UserAddress: &upb.UserAddress{
			PostalCode:   "updated-postal-code",
			Prefecture:   "updated-prefecture-id",
			City:         "updated-test-city",
			FirstStreet:  "updated-test-street1",
			SecondStreet: "updated-test-street2",
		},
	}

	// param for updated street 1
	info5 := &UpdateBillingAddressEventInfo{
		StudentID: studentID,
		PayerName: existingPaymentDetail.PayerName.String,
		UserAddress: &upb.UserAddress{
			PostalCode:   existingBillingAddress.PostalCode.String,
			Prefecture:   prefectureID,
			City:         existingBillingAddress.City.String,
			FirstStreet:  "",
			SecondStreet: existingBillingAddress.Street2.String,
		},
	}

	prefecture := &entities.Prefecture{
		ID:             database.Text(prefectureID),
		PrefectureCode: database.Text(existingBillingAddress.PrefectureCode.String),
	}

	updatedPrefecture := &entities.Prefecture{
		ID:             database.Text("updated-prefecture-id"),
		PrefectureCode: database.Text("updated-prefecture-code"),
	}

	testError := errors.New("test-error")

	testcases := []TestCase{
		{
			name:        "happy case - student has existing billing address and there are changes in home address",
			ctx:         ctx,
			req:         info1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(updatedPrefecture, nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has existing billing address and there are updates in payer name",
			ctx:         ctx,
			req:         info2,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has existing billing address, the payment method was previously empty, there are changes in home address and student has no existing bank account",
			ctx:         ctx,
			req:         info1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, ""), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(updatedPrefecture, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has existing billing address, the payment method was previously empty, there are changes in home address and student bank account is verified",
			ctx:         ctx,
			req:         info1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, ""), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(updatedPrefecture, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(&entities.BankAccount{IsVerified: database.Bool(true)}, nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has no changes",
			ctx:         ctx,
			req:         info3,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(prefecture, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has no billing and payment details but has UserAddress on event message",
			ctx:         ctx,
			req:         info3,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(prefecture, nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has no billing and has payment details but has UserAddress on event message and update payer name",
			ctx:         ctx,
			req:         info2,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - student has billing details but UserAddress is null",
			ctx:  ctx,
			req: &UpdateBillingAddressEventInfo{
				StudentID: studentID,
				PayerName: existingPaymentDetail.PayerName.String,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - student has billing details and one important field is empty",
			ctx:  ctx,
			req: &UpdateBillingAddressEventInfo{
				StudentID: studentID,
				PayerName: existingPaymentDetail.PayerName.String,
				UserAddress: &upb.UserAddress{
					PostalCode:   existingBillingAddress.PostalCode.String,
					Prefecture:   prefectureID,
					City:         "",
					FirstStreet:  existingBillingAddress.Street1.String,
					SecondStreet: existingBillingAddress.Street2.String,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - student has no existing billing address and payment detail and UserAddress value was not provided",
			ctx:  ctx,
			req: &UpdateBillingAddressEventInfo{
				StudentID:   "test",
				PayerName:   "test",
				UserAddress: nil,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - student has no existing billing address and has payment detail and UserAddress value are incomplete",
			ctx:  ctx,
			req: &UpdateBillingAddressEventInfo{
				StudentID: "test",
				PayerName: "test",
				UserAddress: &upb.UserAddress{
					PostalCode:   existingBillingAddress.PostalCode.String,
					Prefecture:   "",
					City:         existingBillingAddress.City.String,
					FirstStreet:  existingBillingAddress.Street1.String,
					SecondStreet: existingBillingAddress.Street2.String,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - there are empty fields when creating billing address and payment detail",
			ctx:  ctx,
			req: &UpdateBillingAddressEventInfo{
				StudentID: studentID,
				PayerName: existingPaymentDetail.PayerName.String,
				UserAddress: &upb.UserAddress{
					PostalCode:   existingBillingAddress.PostalCode.String,
					Prefecture:   "",
					City:         existingBillingAddress.City.String,
					FirstStreet:  existingBillingAddress.Street1.String,
					SecondStreet: existingBillingAddress.Street2.String,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has existing billing address and street1 updated to empty",
			ctx:         ctx,
			req:         info5,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(updatedPrefecture, nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "happy case - student has existing billing address and there are changes in home address and payer name",
			ctx:         ctx,
			req:         info4,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(updatedPrefecture, nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - student ID is empty",
			ctx:  ctx,
			req: &UpdateBillingAddressEventInfo{
				StudentID: "",
				PayerName: "test",
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative case - payer name is empty",
			ctx:  ctx,
			req: &UpdateBillingAddressEventInfo{
				StudentID: "test",
				PayerName: "",
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "negative case - error on StudentPaymentDetailRepo.FindByStudentID",
			ctx:         ctx,
			req:         info1,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("StudentPaymentDetailRepo.FindByStudentID: %v", testError)),
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name:        "negative case - error on BillingAddressRepo.FindByUserID",
			ctx:         ctx,
			req:         info1,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("BillingAddressRepo.FindByUserID: %v", testError)),
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name:        "negative case - BillingAddressRepo.Upsert error",
			ctx:         ctx,
			req:         info1,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("BillingAddressRepo.Upsert: %v", testError)),
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(prefecture, nil)
				mockBillingAddressRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative case - StudentPaymentDetailRepo.Upsert error",
			ctx:         ctx,
			req:         info2,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("StudentPaymentDetailRepo.Upsert: %v", testError)),
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(prefecture, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative case - PrefectureRepo.FindByPrefectureID error",
			ctx:         ctx,
			req:         info2,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("PrefectureRepo.FindByPrefectureID: %v", testError)),
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative case - PrefectureRepo.FindByPrefectureID error",
			ctx:         ctx,
			req:         info2,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("PrefectureRepo.FindByPrefectureID: %v", testError)),
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(generateStudentPaymentDetail(studentPaymentDetailID, studentID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative case - PrefectureRepo.FindByPrefectureID error",
			ctx:         ctx,
			req:         info2,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("PrefectureRepo.FindByPrefectureID: %v", testError)),
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPrefectureRepo.On("FindByPrefectureID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative case - unexpected null payment detail",
			ctx:         ctx,
			req:         info2,
			expectedErr: status.Error(codes.Internal, "unexpected null student payment detail"),
			setup: func(ctx context.Context) {
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(generateBillingAddress(studentPaymentDetailID, studentID), nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.AutoUpdateBillingAddressInfoAndPaymentDetail(testCase.ctx, testCase.req.(*UpdateBillingAddressEventInfo))
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTx, mockStudentPaymentDetailRepo, mockBillingAddressRepo, mockPrefectureRepo, mockStudentPaymentDetailActionLogRepo)
		})
	}
}

func generateBillingAddress(studentPaymentDetailID, studentID string) *entities.BillingAddress {
	return &entities.BillingAddress{
		StudentPaymentDetailID: database.Text(studentPaymentDetailID),
		UserID:                 database.Text(studentID),
		PostalCode:             database.Text("test-postal-code-existing"),
		PrefectureCode:         database.Text("test-prefecture-code-existing"),
		City:                   database.Text("test-city-existing"),
		Street1:                database.Text("test-street1-existing"),
		Street2:                database.Text("test-street2-existing"),
	}
}

func generateStudentPaymentDetail(studentPaymentDetailID, studentID string, paymentMethod string) *entities.StudentPaymentDetail {
	return &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text(studentPaymentDetailID),
		StudentID:              database.Text(studentID),
		PayerName:              database.Text("test-payer-name-existing"),
		PaymentMethod:          database.Text(paymentMethod),
	}
}
