package invoicesvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_IssueInvoice(t *testing.T) {

	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDb := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLog := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &InvoiceModifierService{
		DB:                   mockDb,
		InvoiceRepo:          mockInvoiceRepo,
		PaymentRepo:          mockPaymentRepo,
		InvoiceActionLogRepo: mockInvoiceActionLog,
	}

	invoiceIssued := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
	}

	invoiceIssuedInvalidType := &entities.Invoice{
		InvoiceID: database.Text("456"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text("Invalid-Type"),
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
		PaymentSequenceNumber: database.Int4(1),
	}

	failedResp := &invoice_pb.IssueInvoiceResponse{
		Successful: false,
	}

	successfulResp := &invoice_pb.IssueInvoiceResponse{
		Successful: true,
	}

	testcases := []TestCase{
		{
			name: "happy case - payment method: CONVENIENCE_STORE",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentSuccessful, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - remarks not provided",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedErr:  nil,
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentSuccessful, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - payment method: CASH",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentSuccessful, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - payment method: BANK_TRANSFER",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_BANK_TRANSFER,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentSuccessful, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - begin failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "db.Begin: 0B000"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, fmt.Errorf(pgerrcode.InvalidTransactionInitiation))
			},
		},
		{
			name: "negative test - commit failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "commit unexpectedly resulted in rollback"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxCommitRollback)
				mockTx.On("Commit", ctx).Once().Return(pgx.ErrTxCommitRollback)
				mockTx.On("Rollback", ctx).Once().Return(pgx.ErrTxCommitRollback)
			},
		},
		{
			name: "negative test - create payment failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - issue invoice failed, changes rolled back",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - RetrieveInvoiceByInvoiceID failed, changes rolled back",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "negative test - GetLatestPaymentDueDateByInvoiceID failed, changes rolled back",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - payment Create failed, changes rolled back",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "rpc error: code = Internal desc = tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(paymentSuccessful, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "issue manual invoice failed - invalid payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   4,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid PaymentMethod value: NO_DEFAULT_PAYMENT"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
			},
		},
		{
			name: "issue scheduled invoice failed - invalid direct debit payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_DIRECT_DEBIT,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid PaymentMethod value: DIRECT_DEBIT"),
			setup: func(ctx context.Context) {
				invoiceIssued.Type = database.Text(invoice_pb.InvoiceType_SCHEDULED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssued, nil)
			},
		},
		{
			name: "issue scheduled invoice failed - invalid invoice type",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "456",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.Internal, "invalid InvoiceType value: Invalid-Type"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceIssuedInvalidType, fmt.Errorf("invalid InvoiceType value: Invalid-Type"))
			},
		},
		{
			name: "issue invoice failed - invalid due date",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         nil,
				ExpiryDate:      timestamppb.Now(),
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - invalid expiry date",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: "1",
				PaymentMethod:   invoice_pb.PaymentMethod_CASH,
				DueDate:         timestamppb.Now(),
				ExpiryDate:      nil,
				Remarks:         "",
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup:        func(ctx context.Context) {},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.IssueInvoice(testCase.ctx, testCase.req.(*invoice_pb.IssueInvoiceRequest))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t, mockDb, mockInvoiceRepo, mockPaymentRepo)
		})
	}

}
