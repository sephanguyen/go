package paymentsvc

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

func TestPaymentModifierService_BulkCancelPayment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)

	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBulkPaymentRepo := new(mock_repositories.MockBulkPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &PaymentModifierService{
		DB:                   mockDB,
		PaymentRepo:          mockPaymentRepo,
		BulkPaymentRepo:      mockBulkPaymentRepo,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
	}

	bulkPaymentID := "test-bulk-payment-id-1"

	mockBulkPayment1 := &entities.BulkPayment{
		BulkPaymentID:     database.Text(bulkPaymentID),
		BulkPaymentStatus: database.Text(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING.String()),
	}

	mockBulkPayment2 := &entities.BulkPayment{
		BulkPaymentID:     database.Text(bulkPaymentID),
		BulkPaymentStatus: database.Text(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_CANCELLED.String()),
	}

	mockBulkPayment3 := &entities.BulkPayment{
		BulkPaymentID:     database.Text(bulkPaymentID),
		BulkPaymentStatus: database.Text(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_EXPORTED.String()),
	}

	mockPayments := []*entities.Payment{
		{
			PaymentID:     database.Text("payment-id-1"),
			InvoiceID:     database.Text("invoice-id-1"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			IsExported:    database.Bool(false),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			IsExported:    database.Bool(false),
		},
	}

	mockPaymentsWithNonPending := []*entities.Payment{
		{
			PaymentID:     database.Text("payment-id-1"),
			InvoiceID:     database.Text("invoice-id-1"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			IsExported:    database.Bool(false),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
			IsExported:    database.Bool(false),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()),
			IsExported:    database.Bool(false),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_REFUNDED.String()),
			IsExported:    database.Bool(false),
		},
	}

	mockPaymentsWithAllFailedStatus := []*entities.Payment{
		{
			PaymentID:     database.Text("payment-id-1"),
			InvoiceID:     database.Text("invoice-id-1"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
			IsExported:    database.Bool(false),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
			IsExported:    database.Bool(false),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
			IsExported:    database.Bool(false),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
			IsExported:    database.Bool(false),
		},
	}

	mockPaymentsWithExported := []*entities.Payment{
		{
			PaymentID:     database.Text("payment-id-1"),
			InvoiceID:     database.Text("invoice-id-1"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			IsExported:    database.Bool(true),
		},
		{
			PaymentID:     database.Text("payment-id-2"),
			InvoiceID:     database.Text("invoice-id-2"),
			PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			IsExported:    database.Bool(false),
		},
	}

	testError := errors.New("test-error")

	testcases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPayments, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateStatusAndAmountByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				for i := 0; i < len(mockPayments); i++ {
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - one payment has FAILED status",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentsWithNonPending, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateStatusAndAmountByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				for i := 0; i < len(mockPayments)-1; i++ {
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - all payment has FAILED status",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentsWithAllFailedStatus, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - empty bulk_payment_id",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: "  ",
			},
			expectedErr: status.Error(codes.InvalidArgument, "bulk_payment_id cannot be empty"),
			setup: func(ctx context.Context) {

			},
		},
		{
			name: "negative case - bulk payment is already CANCELLED",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "bulk payment is not in PENDING status"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment2, nil)
			},
		},
		{
			name: "negative case - bulk payment is already EXPORTED",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "bulk payment is not in PENDING status"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment3, nil)
			},
		},
		{
			name: "negative case - one payment is already exported",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "at least one payment is already exported"),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentsWithExported, nil)
			},
		},
		{
			name: "negative case - error on BulkPaymentRepo.FindByBulkPaymentID",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("BulkPaymentRepo.FindByBulkPaymentID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "negative case - error on PaymentRepo.FindAllByBulkPaymentID",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("PaymentRepo.FindAllByBulkPaymentID err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPayments, testError)
			},
		},
		{
			name: "negative case - error on BulkPaymentRepo.UpdateBulkPaymentStatusByIDs",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("BulkPaymentRepo.UpdateBulkPaymentStatusByIDs err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPayments, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - error on PaymentRepo.UpdateStatusAndAmountByPaymentIDs",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("PaymentRepo.UpdateStatusAndAmountByPaymentIDs err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPayments, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateStatusAndAmountByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - error on utils.CreateActionLogV2",
			ctx:  ctx,
			req: &invoice_pb.BulkCancelPaymentRequest{
				BulkPaymentId: bulkPaymentID,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("utils.CreateActionLogV2 err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBulkPaymentRepo.On("FindByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPayment1, nil)
				mockPaymentRepo.On("FindAllByBulkPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPayments, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkPaymentRepo.On("UpdateBulkPaymentStatusByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("UpdateStatusAndAmountByPaymentIDs", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.BulkCancelPayment(testCase.ctx, testCase.req.(*invoice_pb.BulkCancelPaymentRequest))
			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if response == nil {
					fmt.Println(err)
				}

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTx, mockPaymentRepo, mockBulkPaymentRepo, mockInvoiceActionLogRepo)
		})
	}

}
