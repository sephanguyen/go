package invoicesvc

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_ApprovePayment(t *testing.T) {
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
	mockInvoiceBillItemRepo := new(mock_repositories.MockInvoiceBillItemRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
		PaymentRepo:          mockPaymentRepo,
	}

	// Entity objects
	invoicePaid := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_PAID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceRefunded := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_REFUNDED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceVoid := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceFailed := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceDraft := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceIssued := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	paymentFailed := &entities.Payment{
		PaymentID:         database.Text("1"),
		CreatedAt:         pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:         invoiceIssued.InvoiceID,
		PaymentDate:       pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:     database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		PaymentStatus:     database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
	}

	paymentSuccessful := &entities.Payment{
		PaymentID:             database.Text("1"),
		CreatedAt:             pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:             invoiceIssued.InvoiceID,
		PaymentDate:           pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:        pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()),
		PaymentSequenceNumber: database.Int4(2),
	}

	paymentPending := &entities.Payment{
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

	invoiceBillItemsEmpty := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{})

	invoiceBillItemsSingle := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{
		{
			BillItemSequenceNumber: database.Int4(1),
			PastBillingStatus:      database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
		},
	})

	invoiceBillItemsMultiple := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{
		{
			BillItemSequenceNumber: database.Int4(1),
			PastBillingStatus:      database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
		},
		{
			BillItemSequenceNumber: database.Int4(2),
			PastBillingStatus:      database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
		},
	})

	// Request and response objects
	approvePaymentReq := &invoice_pb.ApproveInvoicePaymentRequest{
		InvoiceId: "1",
		Remarks:   "any",
	}

	approvePaymentReqNoRemarks := &invoice_pb.ApproveInvoicePaymentRequest{
		InvoiceId: "1",
	}

	successfulResp := &invoice_pb.ApproveInvoicePaymentResponse{
		Successful: true,
	}

	failedResp := &invoice_pb.ApproveInvoicePaymentResponse{
		Successful: false,
	}

	testcases := []TestCase{
		{
			name: "negative test - No invoice provided",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ApproveInvoicePaymentRequest{
				Remarks: "any",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "InvoiceId is required"),
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
			name:         "negative test - InvoiceBillItem.FindAllByInvoiceID DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error InvoiceBillItem FindAllByInvoiceID: no rows in result set"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:         "negative test - No invoice bill items",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "No invoice bill items; cannot approve payment"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(&invoiceBillItemsEmpty, nil)
			},
		},
		{
			name:         "negative test - InvoiceRepo Update DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Invoice UpdateWithFields: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
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
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
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
				paymentPending.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case - with single bill item (positive final price) and remarks",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				paymentPending.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case - with single bill item (negative final price) and no remarks",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReqNoRemarks,
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				paymentPending.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case - with multiple bill items and remarks",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          approvePaymentReq,
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				invoiceIssued.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				paymentPending.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(paymentPending, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(&invoiceBillItemsMultiple, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ApproveInvoicePayment(testCase.ctx, testCase.req.(*invoice_pb.ApproveInvoicePaymentRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, response)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockInvoiceBillItemRepo, mockInvoiceActionLogRepo)
		})
	}
}
