package invoicesvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const userID = "user-id"

func TestInvoiceModifierService_CreateActionLog(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
	}

	testcases := []TestCase{
		{
			name:        "happy case for issue invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             "1",
				Action:                invoice_pb.InvoiceAction_INVOICE_ISSUED,
				PaymentSequenceNumber: 1123,
				PaymentMethod:         "Convenience Store",
				ActionComment:         "Sample Comment1",
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case for void invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             "2",
				Action:                invoice_pb.InvoiceAction_INVOICE_VOIDED,
				PaymentSequenceNumber: 1123,
				ActionComment:         "Sample Comment2",
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case for edit credit note",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:     "3",
				Action:        invoice_pb.InvoiceAction_EDIT_CREDIT_NOTE,
				ActionComment: "Sample Comment3",
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case for paid invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             "4",
				Action:                invoice_pb.InvoiceAction_INVOICE_PAID,
				PaymentSequenceNumber: 4,
				ActionComment:         "Sample Comment4",
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case for failed/cancel invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             "4",
				Action:                invoice_pb.InvoiceAction_INVOICE_FAILED,
				PaymentSequenceNumber: 5,
				ActionComment:         "Sample Comment5",
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case for remove credit note",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             "4",
				Action:                invoice_pb.InvoiceAction_REMOVE_CREDIT_NOTE,
				PaymentSequenceNumber: 6611,
				ActionComment:         "Sample Comment remove credit note",
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "Failed to create action log invalid invoice id",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: status.Error(codes.InvalidArgument, "invalid invoice id"),
			req: &InvoiceActionLogDetails{
				InvoiceID:             "",
				Action:                invoice_pb.InvoiceAction_INVOICE_ISSUED,
				PaymentSequenceNumber: 1126,
				PaymentMethod:         "Convenience Store",
				ActionComment:         "Sample Comment6",
			},
			setup: func(ctx context.Context) {
				// Do nothing, no setup.
			},
		},
		{
			name:        "Failed to create action log invalid payment sequence number",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: status.Error(codes.InvalidArgument, "invalid payment sequence number"),
			req: &InvoiceActionLogDetails{
				InvoiceID:             "7",
				Action:                invoice_pb.InvoiceAction_INVOICE_ISSUED,
				PaymentSequenceNumber: 0,
				PaymentMethod:         "Convenience Store",
				ActionComment:         "Sample Comment7",
			},
			setup: func(ctx context.Context) {
				// Do nothing, no setup.
			},
		},
		{
			name:        "Failed to create action log closed db pool",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: status.Error(codes.Internal, "closed pool"),
			req: &InvoiceActionLogDetails{
				InvoiceID:             "1",
				Action:                invoice_pb.InvoiceAction_INVOICE_ISSUED,
				PaymentSequenceNumber: 1127,
				PaymentMethod:         "Direct Debit",
				ActionComment:         "Sample Comment8",
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(puddle.ErrClosedPool)
			},
		},
		{
			name:        "Invalid Action Detail",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: status.Error(codes.InvalidArgument, "invalid invoice action detail"),
			req: &InvoiceActionLogDetails{
				InvoiceID:             "1",
				Action:                99,
				PaymentSequenceNumber: 1127,
				PaymentMethod:         "Direct Debit",
				ActionComment:         "Sample Comment99",
			},
			setup: func(ctx context.Context) {
				// Do nothing, no setup.
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.createActionLog(testCase.ctx, mockDB, testCase.req.(*InvoiceActionLogDetails))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceActionLogRepo)
		})
	}
}
