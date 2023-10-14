package payment_detail

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_UpdateStudentPaymentMethod(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := &mock_database.Tx{}
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockBankAccountRepo := new(mock_repositories.MockBankAccountRepo)
	mockBillingAddressRepo := new(mock_repositories.MockBillingAddressRepo)
	mockStudentPaymentDetailActionLogRepo := &mock_repositories.MockStudentPaymentDetailActionLogRepo{}

	s := &EditPaymentDetailService{
		DB:                                mockDB,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		BillingAddressRepo:                mockBillingAddressRepo,
		BankAccountRepo:                   mockBankAccountRepo,
		StudentPaymentDetailActionLogRepo: mockStudentPaymentDetailActionLogRepo,
	}

	studentPaymentDetailID := "student-payment-detail-id"
	studentID := "student-id-1"

	bankAccount := &entities.BankAccount{
		StudentPaymentDetailID: database.Text(studentPaymentDetailID),
		IsVerified:             pgtype.Bool{Bool: true},
	}

	billingAddress := &entities.BillingAddress{
		StudentPaymentDetailID: database.Text(studentPaymentDetailID),
		UserID:                 database.Text(studentPaymentDetailID),
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name: "Happy case update student payment detail payment method convenience store",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.UpdateStudentPaymentMethodResponse{
				Successful: true,
			},
			setupCtx: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(getTestStudentPaymentDetail(studentPaymentDetailID, studentID), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(billingAddress, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(bankAccount, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Happy case - no changes were made",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.UpdateStudentPaymentMethodResponse{
				Successful: true,
			},
			setupCtx: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(getTestStudentPaymentDetail(studentPaymentDetailID, studentID), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(billingAddress, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(bankAccount, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Error on finding student payment detail record",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail FindByID: %v", testError)),
			setupCtx: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Error on finding student billing address record",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error Billing Address FindByUserID: %v", testError)),
			setupCtx: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(getTestStudentPaymentDetail(studentPaymentDetailID, studentID), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Error on finding bank account of student record",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error BankAccount FindByStudentID: %v", testError)),
			setupCtx: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(getTestStudentPaymentDetail(studentPaymentDetailID, studentID), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(billingAddress, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Error on updating student detail record",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail Upsert: %v", testError)),
			setupCtx: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(getTestStudentPaymentDetail(studentPaymentDetailID, studentID), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(billingAddress, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(bankAccount, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "Error on creating action log",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("err cannot create student payment detail action log: %v", testError)),
			setupCtx: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(getTestStudentPaymentDetail(studentPaymentDetailID, studentID), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(billingAddress, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(bankAccount, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "Error on updating bank account not verified",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              studentID,
				StudentPaymentDetailId: studentPaymentDetailID,
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			expectedErr: status.Error(codes.Internal, "error existing bank account should be verified"),
			setupCtx: func(ctx context.Context) {
				bankAccount.IsVerified = pgtype.Bool{Bool: false}
				mockStudentPaymentDetailRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(getTestStudentPaymentDetail(studentPaymentDetailID, studentID), nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(billingAddress, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(bankAccount, nil)
			},
		},
		{
			name: "Invalid student id empty",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "student id cannot be empty"),
			setupCtx: func(ctx context.Context) {
			},
		},
		{
			name: "Invalid student payment detail id empty",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              "123",
				StudentPaymentDetailId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "student payment detail id cannot be empty"),
			setupCtx: func(ctx context.Context) {
			},
		},
		{
			name: "Invalid student payment method cash",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              "123",
				StudentPaymentDetailId: "523",
				PaymentMethod:          invoice_pb.PaymentMethod_CASH,
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("invalid PaymentMethod value: %v", invoice_pb.PaymentMethod_CASH)),
			setupCtx: func(ctx context.Context) {
			},
		},
		{
			name: "Invalid student payment method bank transfer",
			ctx:  ctx,
			req: &invoice_pb.UpdateStudentPaymentMethodRequest{
				StudentId:              "123",
				StudentPaymentDetailId: "523",
				PaymentMethod:          invoice_pb.PaymentMethod_BANK_TRANSFER,
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("invalid PaymentMethod value: %v", invoice_pb.PaymentMethod_BANK_TRANSFER)),
			setupCtx: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupCtx(testCase.ctx)
			resp, err := s.UpdateStudentPaymentMethod(testCase.ctx, testCase.req.(*invoice_pb.UpdateStudentPaymentMethodRequest))
			log.Println(err)

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockTx,
				mockStudentPaymentDetailRepo,
				mockBillingAddressRepo,
				mockBankAccountRepo,
				mockStudentPaymentDetailActionLogRepo,
			)
		})
	}

}

func getTestStudentPaymentDetail(studentPaymentDetailID, studentID string) *entities.StudentPaymentDetail {
	return &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text(studentPaymentDetailID),
		StudentID:              database.Text(studentID),
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}
}
