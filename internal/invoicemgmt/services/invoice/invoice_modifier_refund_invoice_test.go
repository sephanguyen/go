package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getValidInvoice() *entities.Invoice {
	validInvoice := &entities.Invoice{
		InvoiceID:          database.Text("invoice-test-id-1"),
		Status:             database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		Total:              database.Numeric(-100),
		OutstandingBalance: database.Numeric(-100),
	}
	return validInvoice
}

func TestInvoiceModifierService_RefundInvoice(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	failedInvoice := &entities.Invoice{
		InvoiceID: database.Text("invoice-test-id-1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
	}

	invoiceWithPositiveTotalAmount := &entities.Invoice{
		InvoiceID: database.Text("invoice-test-id-1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		Total:     database.Numeric(100),
	}

	invoiceWithZeroTotalAmount := &entities.Invoice{
		InvoiceID: database.Text("invoice-test-id-1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		Total:     database.Numeric(0),
	}

	invoiceWithPositiveOutstandingBalance := &entities.Invoice{
		InvoiceID:          database.Text("invoice-test-id-1"),
		Status:             database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		Total:              database.Numeric(-100),
		OutstandingBalance: database.Numeric(100),
	}

	invoiceWithZeroOutstandingBalance := &entities.Invoice{
		InvoiceID:          database.Text("invoice-test-id-1"),
		Status:             database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		Total:              database.Numeric(-100),
		OutstandingBalance: database.Numeric(0),
	}

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
	}

	testError := errors.New("test error")

	testCases := []TestCase{
		{
			name: "Happy Case - Cash Refund Method",
			ctx:  ctx,
			expectedResp: &invoice_pb.RefundInvoiceResponse{
				Successful: true,
			},
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(getValidInvoice(), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "Happy Case - Bank Transfer Refund Method",
			ctx:  ctx,
			expectedResp: &invoice_pb.RefundInvoiceResponse{
				Successful: true,
			},
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_BANK_TRANSFER,
				Amount:       -100,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(getValidInvoice(), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "Validation Error - Invoice ID is empty",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "invoice ID cannot be empty"),
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "Validation Error - Amount is a positive amount",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "amount should be negative value"),
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       100,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "Validation Error - Amount is zero amount",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "amount should be negative value"),
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       0,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "Validation Error - Amount is zero amount",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "amount should be negative value"),
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       0,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Validation Error - Invoice is not ISSUED",
			ctx:  ctx,
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice status should be ISSUED"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(failedInvoice, nil)
			},
		},
		{
			name: "Validation Error - Invoice with positive total amount",
			ctx:  ctx,
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice total should be negative"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceWithPositiveTotalAmount, nil)
			},
		},
		{
			name: "Validation Error - Invoice with zero total amount",
			ctx:  ctx,
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice total should be negative"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceWithZeroTotalAmount, nil)
			},
		},
		{
			name: "Validation Error - Invoice with positive outstanding balance",
			ctx:  ctx,
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice outstanding balance should be negative"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceWithPositiveOutstandingBalance, nil)
			},
		},
		{
			name: "Validation Error - Invoice with zero outstanding balance",
			ctx:  ctx,
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice outstanding balance should be negative"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceWithZeroOutstandingBalance, nil)
			},
		},
		{
			name: "Validation Error - Given amount is not equal to invoice outstanding balance",
			ctx:  ctx,
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -50,
			},
			expectedErr: status.Error(codes.InvalidArgument, "the given amount should be equal to the invoice outstanding balance"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(getValidInvoice(), nil)
			},
		},
		{
			name:        "Internal Error - InvoiceRepo.RetrieveInvoiceByInvoiceID error",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("InvoiceRepo.RetrieveInvoiceByInvoiceID err: %v", testError)),
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name:        "Internal Error - InvoiceRepo.UpdateWithFields error",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, testError.Error()),
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(getValidInvoice(), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "Internal Error - InvoiceActionLogRepo.Create error",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, testError.Error()),
			req: &invoice_pb.RefundInvoiceRequest{
				InvoiceId:    "test-invoice-id",
				RefundMethod: invoice_pb.RefundMethod_REFUND_CASH,
				Amount:       -100,
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(getValidInvoice(), nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.RefundInvoice(testCase.ctx, testCase.req.(*invoice_pb.RefundInvoiceRequest))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockInvoiceActionLogRepo)
		})
	}
}
