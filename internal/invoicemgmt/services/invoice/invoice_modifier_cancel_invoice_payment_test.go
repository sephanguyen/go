package invoicesvc

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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInvoiceModifierService_CancelInvoicePayment(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockTx := &mock_database.Tx{}
	mockDB := &mockDb.Ext{}
	mockInvoiceRepo := new(mockRepositories.MockInvoiceRepo)
	mockBillItemRepo := new(mockRepositories.MockBillItemRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		BillItemRepo:         mockBillItemRepo,
		PaymentRepo:          mockPaymentRepo,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
	}

	cancelPaymentRequestHappyCase := &invoice_pb.CancelInvoicePaymentRequest{
		Remarks:   "Sample Remarks",
		InvoiceId: "1",
	}

	cancelPaymentResponseHappyCase := &invoice_pb.CancelInvoicePaymentResponse{
		Successful: true,
	}

	issuedInvoice := &entities.Invoice{
		InvoiceID: database.Text("1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	voidInvoice := &entities.Invoice{
		InvoiceID: database.Text("2"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	draftInvoice := &entities.Invoice{
		InvoiceID: database.Text("2"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
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

	happyCaseSetup := func(ctx context.Context) {
		mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
		mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
		mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
		mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(payment, nil)
		mockPaymentRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
		mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
		mockTx.On("Commit", ctx).Once().Return(nil)
	}

	negativeCaseSetup := func(ctx context.Context, p *entities.Payment) {
		mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(issuedInvoice, nil)
		mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
		mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
		mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(p, nil)
		mockTx.On("Rollback", ctx).Once().Return(nil)
	}

	retrieveInvoiceByIDSetup := func(ctx context.Context) {
		mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
	}

	request := &invoice_pb.CancelInvoicePaymentRequest{
		InvoiceId: issuedInvoice.InvoiceID.String,
		Remarks:   "sample remark",
	}

	failedResponse := &invoice_pb.CancelInvoicePaymentResponse{
		Successful: false,
	}

	expectedErr := status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status")

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
			req: &invoice_pb.CancelInvoicePaymentRequest{
				Remarks:   "",
				InvoiceId: "1",
			},
			expectedResp: cancelPaymentResponseHappyCase,
			setup:        happyCaseSetup,
		},
		{
			name: "negative test: No invoice provided",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.CancelInvoicePaymentRequest{
				Remarks:   "sample remark",
				InvoiceId: issuedInvoice.InvoiceID.String,
			},
			expectedResp: failedResponse,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", pgx.ErrNoRows.Error())),
			setup:        retrieveInvoiceByIDSetup,
		},
		{
			name:         "negative test: Invoice not found",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          request,
			expectedResp: failedResponse,
			expectedErr:  status.Error(codes.Internal, "error Invoice RetrieveInvoiceByInvoiceID: no rows in result set"),
			setup:        retrieveInvoiceByIDSetup,
		},
		{
			name:         "negative test: Invalid invoice draft status",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          request,
			expectedResp: failedResponse,
			expectedErr:  expectedErr,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(draftInvoice, nil)
			},
		},
		{
			name:         "negative test: Invalid invoice void status",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          request,
			expectedResp: failedResponse,
			expectedErr:  expectedErr,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(voidInvoice, nil)
			},
		},
		{
			name:         "negative test: RetrieveInvoiceByInvoiceID error",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          request,
			expectedResp: failedResponse,
			expectedErr:  status.Error(codes.Internal, "error Invoice RetrieveInvoiceByInvoiceID: tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:         "negative test: Payment ID is nil",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          request,
			expectedResp: failedResponse,
			expectedErr:  fmt.Errorf("error Payment: Payment is nil"),
			setup: func(ctx context.Context) {
				negativeCaseSetup(ctx, nil)
			},
		},
		{
			name:         "negative test: Payment ID is not pending",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          request,
			expectedResp: failedResponse,
			expectedErr:  fmt.Errorf("error Payment: Payment should be pending"),
			setup: func(ctx context.Context) {
				negativeCaseSetup(ctx, notPendingPayment)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.CancelInvoicePayment(testCase.ctx, testCase.req.(*invoice_pb.CancelInvoicePaymentRequest))
			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockBillItemRepo, mockPaymentRepo, mockInvoiceActionLogRepo)
		})
	}
}
