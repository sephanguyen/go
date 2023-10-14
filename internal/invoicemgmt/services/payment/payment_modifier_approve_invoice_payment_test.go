package paymentsvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPaymentModifierService_ApprovePayment(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Mock objects
	mockTx := &mock_database.Tx{}
	mockDB := new(mock_database.Ext)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &PaymentModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
		PaymentRepo:          mockPaymentRepo,
	}

	// Entity objects
	invoicePaid := &entities.Invoice{
		InvoiceID: database.Text("1237"),
		Status:    database.Text(invoice_pb.InvoiceStatus_PAID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(100),
	}

	invoiceRefunded := &entities.Invoice{
		InvoiceID: database.Text("1238"),
		Status:    database.Text(invoice_pb.InvoiceStatus_REFUNDED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(100),
	}

	invoiceVoid := &entities.Invoice{
		InvoiceID: database.Text("1236"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(100),
	}

	invoiceFailed := &entities.Invoice{
		InvoiceID: database.Text("1234"),
		Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(100),
	}

	invoiceDraft := &entities.Invoice{
		InvoiceID: database.Text("1235"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(100),
	}

	invoiceIssued := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(100),
	}

	paymentFailed := &entities.Payment{
		PaymentID:         database.Text("1"),
		CreatedAt:         pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:         invoiceIssued.InvoiceID,
		PaymentDate:       pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:     database.Text(invoice_pb.PaymentMethod_BANK_TRANSFER.String()),
		PaymentStatus:     database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
	}

	paymentSuccessful := &entities.Payment{
		PaymentID:             database.Text("1"),
		CreatedAt:             pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:             invoiceIssued.InvoiceID,
		PaymentDate:           pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:        pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_BANK_TRANSFER.String()),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()),
		PaymentSequenceNumber: database.Int4(2),
	}

	paymentPendingCash := &entities.Payment{
		PaymentID:             database.Text("1"),
		CreatedAt:             pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:             invoiceIssued.InvoiceID,
		PaymentDate:           pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:        pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CASH.String()),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(4),
	}

	paymentPendingBankTransfer := &entities.Payment{
		PaymentID:             database.Text("2"),
		CreatedAt:             pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:             invoiceIssued.InvoiceID,
		PaymentDate:           pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:        pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_BANK_TRANSFER.String()),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(5),
	}

	paymentPendingCC := &entities.Payment{
		PaymentID:             database.Text("1"),
		CreatedAt:             pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:             invoiceIssued.InvoiceID,
		PaymentDate:           pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:        pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(4),
	}

	// invoiceBillItemsEmpty := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{})

	// invoiceBillItemsSingle := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{
	// 	{
	// 		BillItemSequenceNumber: database.Int4(1),
	// 		PastBillingStatus:      database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
	// 	},
	// })

	// billItemNegativeFinalPrice := &entities.BillItem{
	// 	BillItemSequenceNumber: database.Int4(1),
	// 	FinalPrice:             database.Numeric(-1),
	// }

	// Request and response objects
	approvePaymentReq := &invoice_pb.ApproveInvoicePaymentV2Request{
		InvoiceId:   "1",
		Remarks:     "any",
		PaymentDate: timestamppb.Now(),
	}

	approvePaymentReqNoRemarks := &invoice_pb.ApproveInvoicePaymentV2Request{
		InvoiceId:   "1",
		Remarks:     "",
		PaymentDate: timestamppb.Now(),
	}

	// approvePaymentReqNoRemarks := &invoice_pb.ApproveInvoicePaymentRequest{
	// 	InvoiceId: "1",
	// }

	successfulResp := &invoice_pb.ApproveInvoicePaymentV2Response{
		Successful: true,
	}

	failedResp := &invoice_pb.ApproveInvoicePaymentV2Response{
		Successful: false,
	}

	testcases := []TestCase{
		{
			name:         "happy case - approve payment with cash payment method",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingCash, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case - approve payment with bank transfer payment method",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingBankTransfer, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case - approve payment with cash payment method no remarks",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReqNoRemarks,
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				paymentPendingBankTransfer.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingBankTransfer, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case - approve payment with bank transfer payment method no remarks",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				paymentPendingBankTransfer.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingBankTransfer, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - No invoice provided",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ApproveInvoicePaymentV2Request{
				Remarks:     "any",
				PaymentDate: timestamppb.Now(),
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invoice id cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Payment Date is empty",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ApproveInvoicePaymentV2Request{
				InvoiceId: "1",
				Remarks:   "any",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "payment date cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "negative test - Invoice not found",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Invoice RetrieveInvoiceByInvoiceID: no rows in result set"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:         "negative test - Invalid invoice status: DRAFT",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
			},
		},
		{
			name:         "negative test - Invalid invoice status: FAILED",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceFailed, nil)
			},
		},
		{
			name:         "negative test - Invalid invoice status: VOID",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceVoid, nil)
			},
		},
		{
			name:         "negative test - Invalid invoice status: PAID",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoicePaid, nil)
			},
		},
		{
			name:         "negative test - Invalid invoice status: REFUNDED",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceRefunded, nil)
			},
		},
		{
			name:         "negative test - PaymentRepo.GetLatestPaymentDueDateByInvoiceID DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Payment GetLatestPaymentDueDateByInvoiceID: tx is closed"),
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:         "negative test - PaymentRepo.GetLatestPaymentDueDateByInvoiceID: No rows",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Payment GetLatestPaymentDueDateByInvoiceID: no rows in result set"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:         "negative test - Invalid payment status: FAILED",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Payment should be in PENDING status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentFailed, nil)
			},
		},
		{
			name:         "negative test - Invalid payment status: SUCCESSFUL",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Payment should be in PENDING status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentSuccessful, nil)
			},
		},
		{
			name:         "negative test - InvoiceRepo Update DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Invoice UpdateWithFields: tx is closed"),
			setup: func(ctx context.Context) {
				paymentPendingCash.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingCash, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "negative test - InvoiceRepo Update DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("payment method %v is not allowed only cash and bank transfer accepted", invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())),
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				paymentPendingCash.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingCC, nil)
			},
		},
		{
			name:         "negative test - PaymentRepo UpdateWithFields DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Payment UpdateWithFields: tx is closed"),
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingCash, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "negative test - InvoiceActionLog Create DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				paymentPendingCash.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPendingCash, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ApproveInvoicePaymentV2(testCase.ctx, testCase.req.(*invoice_pb.ApproveInvoicePaymentV2Request))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, response)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockInvoiceActionLogRepo)
		})
	}
}
