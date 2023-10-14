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
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_sequence_number "github.com/manabie-com/backend/mock/invoicemgmt/services/sequence_number"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPaymentModifierService_AddInvoicePayment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)

	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBankAccountRepo := new(mock_repositories.MockBankAccountRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockSeqNumberService := new(mock_sequence_number.ISequenceNumberService)

	s := &PaymentModifierService{
		DB:                    mockDB,
		InvoiceRepo:           mockInvoiceRepo,
		PaymentRepo:           mockPaymentRepo,
		BankAccountRepo:       mockBankAccountRepo,
		InvoiceActionLogRepo:  mockInvoiceActionLogRepo,
		SequenceNumberService: mockSeqNumberService,
		UnleashClient:         mockUnleashClient,
	}

	concretePaymentSeqNumberService := &seqnumberservice.PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	failedInvoice := &entities.Invoice{
		InvoiceID: database.Text("invoice-test-id-1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
	}

	invoiceWithNegativeTotalAmount := &entities.Invoice{
		InvoiceID: database.Text("invoice-test-id-1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		Total:     database.Numeric(-100),
	}

	invoiceWithZeroTotalAmount := &entities.Invoice{
		InvoiceID: database.Text("invoice-test-id-1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		Total:     database.Numeric(0),
	}

	testInvoice := &entities.Invoice{
		InvoiceID:          database.Text("invoice-test-id-1"),
		Status:             database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		OutstandingBalance: database.Numeric(10),
		Total:              database.Numeric(10),
		StudentID:          database.Text("test-student-id"),
	}

	invalidLatestPayment := &entities.Payment{
		PaymentID:     database.Text("payment-id-1"),
		PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()),
	}

	failedPayment := &entities.Payment{
		PaymentID:     database.Text("payment-id-1"),
		PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
	}

	createdCSPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}

	createdDDPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	unVerifiedBankAccount := &entities.BankAccount{
		StudentID:  database.Text("test-student-id-1"),
		IsVerified: database.Bool(false),
	}

	verifiedBankAccount := &entities.BankAccount{
		StudentID:  database.Text("test-student-id-1"),
		IsVerified: database.Bool(true),
	}

	testError := errors.New("testError")

	testcases := []TestCase{
		{
			name: "Happy Case - Convenience Store - Invoice has no existing payment",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedResp: &invoice_pb.AddInvoicePaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Happy Case - Convenience Store - Invoice has existing failed payment",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedResp: &invoice_pb.AddInvoicePaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(failedPayment, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Happy Case - Direct Debit - Invoice has no existing payment",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedResp: &invoice_pb.AddInvoicePaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(verifiedBankAccount, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdDDPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Validation Error - Empty Invoice ID",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice ID cannot be empty"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Zero amount",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        0,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "amount cannot be less than or equal to 0"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Payment Method Not Allowed",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("payment method %v is not allowed", invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT.String())),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Nil Due Date",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       nil,
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Nil Expiry Date",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    nil,
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Due Date is less than now",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(-1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Expiry Date is less than now",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Due Date is greater than Expiry Date",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(2 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Validation Error - Invoice is not ISSUED",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice status should be ISSUED"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(failedInvoice, nil)
			},
		},
		{
			name: "Validation Error - Invoice total is negative",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice total cannot be less than or equal to 0"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceWithNegativeTotalAmount, nil)
			},
		},
		{
			name: "Validation Error - Invoice total is zero",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice total cannot be less than or equal to 0"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceWithZeroTotalAmount, nil)
			},
		},
		{
			name: "Validation Error - Given amount is not equal to the invoice outstanding balance",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "the given amount should be equal to the invoice outstanding balance"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
			},
		},
		{
			name: "Validation Error - Latest payment of invoice is not in FAILED status",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "latest payment should have FAILED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invalidLatestPayment, nil)
			},
		},
		{
			name: "Validation Error - Student has no bank account",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "student should have verified bank account if the payment method is DIRECT DEBIT"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(failedPayment, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "Validation Error - Payment method is DIRECT DEBIT and the bank account is not verified",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.InvalidArgument, "bank account should be verified if the payment method is DIRECT DEBIT"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(failedPayment, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(unVerifiedBankAccount, nil)
			},
		},
		{
			name: "Internal Error - InvoiceRepo.RetrieveInvoiceByInvoiceID error",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("InvoiceRepo.RetrieveInvoiceByInvoiceID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Internal Error - PaymentRepo.GetLatestPaymentDueDateByInvoiceID error",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("PaymentRepo.GetLatestPaymentDueDateByInvoiceID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Internal Error - BankAccountRepo.FindByStudentID error",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("BankAccountRepo.FindByStudentID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(failedPayment, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Internal Error - PaymentRepo.Create error",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "Internal Error - PaymentRepo.GetLatestPaymentDueDateByInvoiceID inside TX error",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "Internal Error - InvoiceActionLogRepo.Create error",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedErr: status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.AddInvoicePayment(testCase.ctx, testCase.req.(*invoice_pb.AddInvoicePaymentRequest))
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
				mockTx,
				mockInvoiceRepo,
				mockPaymentRepo,
				mockBankAccountRepo,
			)
		})
	}
}

func TestPaymentModifierService_AddInvoicePayment_ManualSetOPaymentSeqNumber(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)

	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBankAccountRepo := new(mock_repositories.MockBankAccountRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockSeqNumberService := new(mock_sequence_number.ISequenceNumberService)

	s := &PaymentModifierService{
		DB:                    mockDB,
		InvoiceRepo:           mockInvoiceRepo,
		PaymentRepo:           mockPaymentRepo,
		BankAccountRepo:       mockBankAccountRepo,
		InvoiceActionLogRepo:  mockInvoiceActionLogRepo,
		SequenceNumberService: mockSeqNumberService,
		UnleashClient:         mockUnleashClient,
	}

	concretePaymentSeqNumberService := &seqnumberservice.PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	testInvoice := &entities.Invoice{
		InvoiceID:          database.Text("invoice-test-id-1"),
		Status:             database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		OutstandingBalance: database.Numeric(10),
		Total:              database.Numeric(10),
		StudentID:          database.Text("test-student-id"),
	}

	failedPayment := &entities.Payment{
		PaymentID:     database.Text("payment-id-1"),
		PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
	}

	createdCSPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}

	createdDDPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	verifiedBankAccount := &entities.BankAccount{
		StudentID:  database.Text("test-student-id-1"),
		IsVerified: database.Bool(true),
	}

	testcases := []TestCase{
		{
			name: "Happy Case - Convenience Store - Invoice has no existing payment",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedResp: &invoice_pb.AddInvoicePaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Happy Case - Convenience Store - Invoice has existing failed payment",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedResp: &invoice_pb.AddInvoicePaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(failedPayment, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Happy Case - Direct Debit - Invoice has no existing payment",
			ctx:  ctx,
			req: &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     "test-invoice-id-1",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        10,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:       "test-remarks",
			},
			expectedResp: &invoice_pb.AddInvoicePaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(verifiedBankAccount, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdDDPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.AddInvoicePayment(testCase.ctx, testCase.req.(*invoice_pb.AddInvoicePaymentRequest))
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
				mockTx,
				mockInvoiceRepo,
				mockPaymentRepo,
				mockBankAccountRepo,
			)
		})
	}
}
