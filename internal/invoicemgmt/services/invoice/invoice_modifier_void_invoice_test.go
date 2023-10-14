package invoicesvc

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
	mock_services "github.com/manabie-com/backend/mock/invoicemgmt/services"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_VoidInvoice(t *testing.T) {
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
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)
	mockOrderServiceClient := new(mock_services.OrderService)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		InternalOrderService: mockOrderServiceClient,
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

	invoiceDraft := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceIssued := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	payment := &entities.Payment{
		PaymentID:         database.Text("1"),
		CreatedAt:         pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:         invoiceIssued.InvoiceID,
		PaymentDate:       pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:     database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		PaymentStatus:     database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
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

	billItem := &entities.BillItem{
		BillItemSequenceNumber: database.Int4(1),
	}

	updateBillItemStatusResp := &payment_pb.UpdateBillItemStatusResponse{
		Errors: []*payment_pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{},
	}

	// Request and response objects
	updateBillItemStatusRespErr := &payment_pb.UpdateBillItemStatusResponse{
		Errors: []*payment_pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{
			{
				BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
				Error:                  "unable to update billing item status: err 1",
			},
			{
				BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
				Error:                  "unable to update billing item status: err 2",
			},
		},
	}

	voidInvoiceReq := &invoice_pb.VoidInvoiceRequest{
		InvoiceId: "1",
		Remarks:   "any",
	}

	updateBillItemStatusErr := fmt.Errorf("error updating billing items status")

	failedResp := &invoice_pb.VoidInvoiceResponse{
		Successful: false,
	}

	testcases := []TestCase{
		{
			name: "negative test - No invoice provided",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.VoidInvoiceRequest{
				Remarks: "any",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invoiceID is required"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "negative test - Invoice not found",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Invoice RetrieveInvoiceByInvoiceID: no rows in result set"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:         "negative test - InvoiceRepo.RetrieveInvoiceByInvoiceID DB error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Invoice RetrieveInvoiceByInvoiceID: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:         "negative test -  Invalid invoice status: VOID",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in DRAFT, ISSUED, or FAILED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceVoid, nil)
			},
		},
		{
			name:         "negative test -  Invalid invoice status: PAID",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in DRAFT, ISSUED, or FAILED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoicePaid, nil)
			},
		},
		{
			name:         "negative test -  Invalid invoice status: REFUNDED",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "Invoice should be in DRAFT, ISSUED, or FAILED status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceRefunded, nil)
			},
		},
		{
			name:         "negative test - PaymentRepo.GetLatestPaymentDueDateByInvoiceID DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Payment GetLatestPaymentDueDateByInvoiceID: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name:         "negative test - PaymentRepo.Update DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Payment Update: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name:         "negative test - InvoiceRepo.Update DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Invoice Update: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name:         "negative test - InvoiceActionLogRepo.Create DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "negative test - InvoiceBillItemRepo.FindAllByInvoiceID DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Invoice Bill Item FindAllByInvoiceID: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "negative test - BillItemRepo.FindByID DB Error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error Bill Item FindByID: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "negative test - UpdateBillItemStatus failed",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error UpdateBillItemStatus: error updating billing items status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Once().Return(billItem, nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, updateBillItemStatusErr)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "negative test - UpdateBillItemStatus partially failed with bill items",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "error UpdateBillItemStatus: BillItemSequenceNumber 1 with error unable to update billing item status: err 1,BillItemSequenceNumber 1 with error unable to update billing item status: err 2"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Once().Return(billItem, nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(updateBillItemStatusRespErr, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case with single invoice bill item",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Once().Return(billItem, nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(updateBillItemStatusResp, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case with multiple invoice bill items",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&invoiceBillItemsMultiple, nil)
				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(2).Return(billItem, nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(updateBillItemStatusResp, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "happy case without invoice bill item",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          voidInvoiceReq,
			expectedResp: failedResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&invoiceBillItemsEmpty, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.VoidInvoice(testCase.ctx, testCase.req.(*invoice_pb.VoidInvoiceRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, response)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockBillItemRepo, mockInvoiceBillItemRepo, mockInvoiceActionLogRepo)
		})
	}
}
