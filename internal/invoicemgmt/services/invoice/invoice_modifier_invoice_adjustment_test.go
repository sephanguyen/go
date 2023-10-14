package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_InvoiceAdjustment(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID                  = "user-id"
		emptyDescriptionValidation = "invoice adjustment detail description is empty"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Mock objects
	mockTx := &mock_database.Tx{}
	mockDB := new(mock_database.Ext)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockInvoiceAdjustmentRepo := new(mock_repositories.MockInvoiceAdjustmentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &InvoiceModifierService{
		DB:                    mockDB,
		InvoiceRepo:           mockInvoiceRepo,
		InvoiceActionLogRepo:  mockInvoiceActionLogRepo,
		InvoiceAdjustmentRepo: mockInvoiceAdjustmentRepo,
	}

	invoiceDraft := &entities.Invoice{
		InvoiceID: database.Text("1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		SubTotal: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		StudentID: database.Text("1"),
	}

	invoiceIssued := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		SubTotal: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		StudentID: database.Text("1"),
	}
	invoiceVoid := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		SubTotal: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		StudentID: database.Text("1"),
	}
	invoiceFailed := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		SubTotal: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		StudentID: database.Text("1"),
	}

	invoicePaid := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_PAID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		SubTotal: pgtype.Numeric{Int: big.NewInt(int64(523.00)),
			Status: pgtype.Present,
		},
		StudentID: database.Text("1"),
	}

	invoiceAdjustment := &entities.InvoiceAdjustment{
		InvoiceAdjustmentID: database.Text("3"),
		Amount: pgtype.Numeric{Int: big.NewInt(int64(100.00)),
			Status: pgtype.Present,
		},
	}
	invoiceAdjustmentTwo := &entities.InvoiceAdjustment{
		InvoiceAdjustmentID: database.Text("4"),
		Amount: pgtype.Numeric{Int: big.NewInt(int64(100.00)),
			Status: pgtype.Present,
		},
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name: "happy case - single create invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 623.00,
				InvoiceTotal:    623.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  100,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedResp: &invoice_pb.UpsertInvoiceAdjustmentsResponse{
				Success: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("UpsertMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - multi create invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 623.00,
				InvoiceTotal:    623.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  50,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
					{
						Description:             "test2",
						Amount:                  50,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedResp: &invoice_pb.UpsertInvoiceAdjustmentsResponse{
				Success: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("UpsertMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - single edit invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 573.00,
				InvoiceTotal:    573.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						InvoiceAdjustmentId:     "1",
						Description:             "test",
						Amount:                  150,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
				},
			},
			expectedResp: &invoice_pb.UpsertInvoiceAdjustmentsResponse{
				Success: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(523)
				invoiceDraft.SubTotal = database.Numeric(523)
				invoiceAdjustment.Amount = database.Numeric(100)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("UpsertMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - multi edit invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 420.00,
				InvoiceTotal:    420.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						InvoiceAdjustmentId:     "1",
						Description:             "test",
						Amount:                  50,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "2",
						Description:             "test2",
						Amount:                  70,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
				},
			},
			expectedResp: &invoice_pb.UpsertInvoiceAdjustmentsResponse{
				Success: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(550)
				invoiceDraft.SubTotal = database.Numeric(550)
				invoiceAdjustment.Amount = database.Numeric(100)
				invoiceAdjustmentTwo.Amount = database.Numeric(150)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustmentTwo, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("UpsertMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - single delete invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 300.00,
				InvoiceTotal:    300.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedResp: &invoice_pb.UpsertInvoiceAdjustmentsResponse{
				Success: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(450)
				invoiceDraft.SubTotal = database.Numeric(450)
				invoiceAdjustment.Amount = database.Numeric(150)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - multi delete invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 150.00,
				InvoiceTotal:    150.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "4",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedResp: &invoice_pb.UpsertInvoiceAdjustmentsResponse{
				Success: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(450)
				invoiceDraft.SubTotal = database.Numeric(450)
				invoiceAdjustment.Amount = database.Numeric(150)
				invoiceAdjustmentTwo.Amount = database.Numeric(150)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustmentTwo, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - create, edit and delete invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 350.00,
				InvoiceTotal:    350.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  50,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "3",
						Description:             "test",
						Amount:                  200,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "4",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedResp: &invoice_pb.UpsertInvoiceAdjustmentsResponse{
				Success: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(500)
				invoiceDraft.SubTotal = database.Numeric(500)
				invoiceAdjustment.Amount = database.Numeric(150)
				invoiceAdjustmentTwo.Amount = database.Numeric(250)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustmentTwo, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("UpsertMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceAdjustmentRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - failed to find invoice id",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 623.00,
				InvoiceTotal:    623.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  100,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", testError)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "negative case - failed to upsert invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 160.00,
				InvoiceTotal:    160.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  100,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error InvoiceAdjustmentRepo UpsertMultiple: %v", testError)),
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(60)
				invoiceDraft.SubTotal = database.Numeric(60)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("UpsertMultiple", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - failed to update fields on invoice",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 450.00,
				InvoiceTotal:    450.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  250,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error InvoiceRepo UpdateWithFields: %v", testError)),
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(200)
				invoiceDraft.SubTotal = database.Numeric(200)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("UpsertMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - failed to soft delete invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 300.00,
				InvoiceTotal:    300.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error InvoiceAdjustmentRepo SoftDeleteByIDs: %v", testError)),
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(200)
				invoiceDraft.SubTotal = database.Numeric(200)
				invoiceAdjustment.Amount = database.Numeric(-100)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - failed to find invoice adjustment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 423.00,
				InvoiceTotal:    423.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error InvoiceAdjustmentRepo FindByID: test error")),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "negative case - failed to create action log",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 800.00,
				InvoiceTotal:    800.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(700)
				invoiceDraft.SubTotal = database.Numeric(700)
				invoiceAdjustment.Amount = database.Numeric(-100)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceAdjustmentRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - failed to proceed with issued invoice status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 423.00,
				InvoiceTotal:    423.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, fmt.Sprintf("error invoice status: %v should be in draft", invoiceIssued.Status.String)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceIssued, nil)
			},
		},
		{
			name: "negative case - failed to proceed with void invoice status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 423.00,
				InvoiceTotal:    423.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, fmt.Sprintf("error invoice status: %v should be in draft", invoiceVoid.Status.String)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceVoid, nil)
			},
		},
		{
			name: "negative case - failed to proceed with failed invoice status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 423.00,
				InvoiceTotal:    423.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, fmt.Sprintf("error invoice status: %v should be in draft", invoiceFailed.Status.String)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceFailed, nil)
			},
		},
		{
			name: "negative case - failed to proceed with paid invoice status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 423.00,
				InvoiceTotal:    423.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{

					{
						InvoiceAdjustmentId:     "3",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, fmt.Sprintf("error invoice status: %v should be in draft", invoicePaid.Status.String)),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoicePaid, nil)
			},
		},
		{
			name: "negative test - empty invoice id in request",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "",
				InvoiceSubTotal: 122.00,
				InvoiceTotal:    122.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  1233,
						InvoiceAdjustmentId:     "1",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice id is required"),
			setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			name: "negative test - no invoice adjustment details",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:                "1",
				InvoiceSubTotal:          123.00,
				InvoiceTotal:             123.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice adjustment detail is empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty description in single create invoice adjustment details",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 122.00,
				InvoiceTotal:    122.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, emptyDescriptionValidation),
			setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			name: "negative test - empty description in single edit invoice adjustment details",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 122.00,
				InvoiceTotal:    122.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "",
						InvoiceAdjustmentId:     "1",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, emptyDescriptionValidation),
			setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			name: "negative test - empty description in multi create invoice adjustment details",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 122.00,
				InvoiceTotal:    122.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  12,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
					{
						Description:             "",
						Amount:                  13,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
					{
						Description:             "test2",
						Amount:                  14,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, emptyDescriptionValidation),
			setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			name: "negative test - empty description in multi edit invoice adjustment details",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 122.00,
				InvoiceTotal:    122.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						InvoiceAdjustmentId:     "1",
						Description:             "test",
						Amount:                  33,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "2",
						Description:             "",
						Amount:                  34,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "3",
						Description:             "test2",
						Amount:                  35,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, emptyDescriptionValidation),
			setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			name: "negative test - total amount not match in invoice adjustment details",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 850.00,
				InvoiceTotal:    123.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						InvoiceAdjustmentId:     "3",
						Description:             "test",
						Amount:                  50,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "4",
						Description:             "test2",
						Amount:                  50,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "expected invoice total amount 850 received 123"),
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(950)
				invoiceDraft.SubTotal = database.Numeric(950)
				invoiceAdjustment.Amount = database.Numeric(-100)
				invoiceAdjustmentTwo.Amount = database.Numeric(300)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustmentTwo, nil)

			},
		},
		{
			name: "negative test - subtotal amount not match in invoice adjustment details",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 123.00,
				InvoiceTotal:    347.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						InvoiceAdjustmentId:     "1",
						Description:             "test",
						Amount:                  23,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
					{
						InvoiceAdjustmentId:     "2",
						Description:             "test2",
						Amount:                  24,
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "expected invoice subtotal amount 347 received 123"),
			setup: func(ctx context.Context) {
				invoiceDraft.Total = database.Numeric(500)
				invoiceDraft.SubTotal = database.Numeric(500)
				invoiceAdjustment.Amount = database.Numeric(-200)
				invoiceAdjustmentTwo.Amount = database.Numeric(400)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceDraft, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustment, nil)
				mockInvoiceAdjustmentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(invoiceAdjustmentTwo, nil)

			},
		},
		{
			name: "negative test - create invoice adjustment with invoice adjustment id on parameter",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.UpsertInvoiceAdjustmentsRequest{
				InvoiceId:       "1",
				InvoiceSubTotal: 555.00,
				InvoiceTotal:    333.00,
				InvoiceAdjustmentDetails: []*invoice_pb.InvoiceAdjustmentDetail{
					{
						Description:             "test",
						Amount:                  322,
						InvoiceAdjustmentId:     "2",
						InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid invoice adjustment id: 2 should be null when creating new record"),
			setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.UpsertInvoiceAdjustments(testCase.ctx, testCase.req.(*invoice_pb.UpsertInvoiceAdjustmentsRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, response)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockInvoiceAdjustmentRepo, mockInvoiceActionLogRepo)
		})
	}
}
