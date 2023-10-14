package paymentsvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPaymentModifierService_CancelInvoicePaymentV2(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockTx := &mock_database.Tx{}
	mockDB := &mockDb.Ext{}
	mockInvoiceRepo := new(mockRepositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkPaymentRepo := new(mock_repositories.MockBulkPaymentRepo)
	s := &PaymentModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		PaymentRepo:          mockPaymentRepo,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
		BulkPaymentRepo:      mockBulkPaymentRepo,
	}

	cancelPaymentRequestHappyCase := &invoice_pb.CancelInvoicePaymentV2Request{
		Remarks:   "Sample Remarks",
		InvoiceId: "1",
	}

	cancelPaymentResponseHappyCase := &invoice_pb.CancelInvoicePaymentV2Response{
		Successful: true,
	}

	issuedInvoice := &entities.Invoice{
		InvoiceID: database.Text("1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	payment := &entities.Payment{
		PaymentID:         database.Text("1"),
		CreatedAt:         pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:         issuedInvoice.InvoiceID,
		PaymentDate:       pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:     database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		PaymentStatus:     database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
	}

	happyCaseSetup := func(ctx context.Context) {
		mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
		mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
		mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
		mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
		mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
		mockTx.On("Commit", ctx).Once().Return(nil)
	}

	retrieveInvoiceByIDSetup := func(ctx context.Context) {
		mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
	}

	request := &invoice_pb.CancelInvoicePaymentV2Request{
		InvoiceId: issuedInvoice.InvoiceID.String,
		Remarks:   "sample remark",
	}

	draftInvoice := &entities.Invoice{
		InvoiceID: database.Text("2"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	voidInvoice := &entities.Invoice{
		InvoiceID: database.Text("2"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	negativeCaseSetup := func(ctx context.Context, p *entities.Payment) {
		mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
		mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
		mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(p, nil)
		mockTx.On("Rollback", ctx).Once().Return(nil)
	}

	notPendingPayment := &entities.Payment{
		PaymentID:         database.Text("1"),
		CreatedAt:         pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:         issuedInvoice.InvoiceID,
		PaymentDate:       pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:     database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		PaymentStatus:     database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
	}

	ddPaymentExported := &entities.Payment{
		PaymentID:         database.Text("1"),
		CreatedAt:         pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:         issuedInvoice.InvoiceID,
		PaymentDate:       pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:     database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		PaymentStatus:     database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		IsExported:        pgtype.Bool{Bool: true},
	}

	paymentBelongInBulk := &entities.Payment{
		PaymentID:         database.Text("1"),
		CreatedAt:         pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		InvoiceID:         issuedInvoice.InvoiceID,
		PaymentDate:       pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentDueDate:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentExpiryDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentMethod:     database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		PaymentStatus:     database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		BulkPaymentID:     database.Text("bulk-payment-1"),
	}

	expectedErrInvoiceStatus := status.Error(codes.InvalidArgument, "invoice should be in ISSUED status")

	testcases := []TestCase{
		{
			name:         "Happy case",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          cancelPaymentRequestHappyCase,
			expectedResp: cancelPaymentResponseHappyCase,
			setup:        happyCaseSetup,
		},
		{
			name: "Happy case: empty remarks",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.CancelInvoicePaymentV2Request{
				Remarks:   "",
				InvoiceId: "1",
			},
			expectedResp: cancelPaymentResponseHappyCase,
			setup:        happyCaseSetup,
		},
		{
			name:         "Happy case payment belong in a bulk no update in bulk record",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          cancelPaymentRequestHappyCase,
			expectedResp: cancelPaymentResponseHappyCase,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentBelongInBulk, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CountOtherPaymentsByBulkPaymentIDNotInStatus", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(1, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:         "Happy case payment belong in a bulk update in bulk record",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          cancelPaymentRequestHappyCase,
			expectedResp: cancelPaymentResponseHappyCase,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentBelongInBulk, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CountOtherPaymentsByBulkPaymentIDNotInStatus", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(0, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test: No invoice provided",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", pgx.ErrNoRows.Error())),
			setup:       retrieveInvoiceByIDSetup,
		},
		{
			name:        "negative test: Invoice not found",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, "error Invoice RetrieveInvoiceByInvoiceID: no rows in result set"),
			setup:       retrieveInvoiceByIDSetup,
		},
		{
			name:        "negative test: Invalid invoice draft status",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: expectedErrInvoiceStatus,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(draftInvoice, nil)
			},
		},
		{
			name:        "negative test: Invalid invoice void status",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: expectedErrInvoiceStatus,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(voidInvoice, nil)
			},
		},
		{
			name:        "negative test: RetrieveInvoiceByInvoiceID error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, "error Invoice RetrieveInvoiceByInvoiceID: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "negative test: Payment ID is nil",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, "error Payment: Payment is nil"),
			setup: func(ctx context.Context) {
				negativeCaseSetup(ctx, nil)
			},
		},
		{
			name:        "negative test: Payment ID is not pending",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, "error Payment: Payment status should be pending"),
			setup: func(ctx context.Context) {
				negativeCaseSetup(ctx, notPendingPayment)
			},
		},
		{
			name:        "negative test: update with fields payment status error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, "error Payment UpdateWithFields: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test: create action log error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: fmt.Errorf("tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test: payment direct debit method already exported",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, "error Payment: Payment method direct debit should not be exported"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(ddPaymentExported, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test: payment belong in a bulk with count other payments by bulk error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         cancelPaymentRequestHappyCase,
			expectedErr: status.Error(codes.Internal, "PaymentRepo.CountOtherPaymentsByBulkPaymentIDNotInStatus err: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentBelongInBulk, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CountOtherPaymentsByBulkPaymentIDNotInStatus", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(0, pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test: payment belong in a bulk no update in bulk record error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         cancelPaymentRequestHappyCase,
			expectedErr: status.Error(codes.Internal, "error BulkPaymentRepo UpdateBulkPaymentStatusByIDs: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentBelongInBulk, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CountOtherPaymentsByBulkPaymentIDNotInStatus", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(0, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.CancelInvoicePaymentV2(testCase.ctx, testCase.req.(*invoice_pb.CancelInvoicePaymentV2Request))
			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockInvoiceActionLogRepo)
		})
	}
}
