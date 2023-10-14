package utils

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

type TestCase struct {
	name                string
	ctx                 context.Context
	req                 interface{}
	expectedResp        interface{}
	expectedErr         error
	setup               func(ctx context.Context)
	mockInvoiceEntities []*entities.Invoice
}

const userID = "user-id"

func TestInvoiceModifierService_CreateActionLogV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	invoiceID := "test-invoice-id"
	actionComment := "test-action-comment"

	testcases := []TestCase{
		{
			name:        "happy case for issue invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:     invoiceID,
				Action:        invoice_pb.InvoiceAction_INVOICE_ISSUED,
				ActionComment: actionComment,
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_INVOICE_ISSUED),
					actionLog.ActionDetail.Set(""),
					actionLog.UserID.Set(userID),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for void invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_INVOICE_VOIDED,
				PaymentSequenceNumber: 1123,
				ActionComment:         actionComment,
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_INVOICE_VOIDED),
					actionLog.ActionDetail.Set(""),
					actionLog.UserID.Set(userID),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for invoice adjusted",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:     invoiceID,
				Action:        invoice_pb.InvoiceAction_INVOICE_ADJUSTED,
				ActionComment: actionComment,
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_INVOICE_ADJUSTED),
					actionLog.ActionDetail.Set(""),
					actionLog.UserID.Set(userID),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for paid invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_INVOICE_PAID,
				PaymentSequenceNumber: 4,
				ActionComment:         actionComment,
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_INVOICE_PAID),
					actionLog.ActionDetail.Set(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()),
					actionLog.UserID.Set(userID),
					actionLog.PaymentSequenceNumber.Set(4),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for approved payment",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_PAYMENT_APPROVED,
				PaymentSequenceNumber: 4,
				ActionComment:         actionComment,
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_PAYMENT_APPROVED),
					actionLog.ActionDetail.Set(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()),
					actionLog.UserID.Set(userID),
					actionLog.PaymentSequenceNumber.Set(4),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for failed/cancel invoice",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_INVOICE_FAILED,
				PaymentSequenceNumber: 5,
				ActionComment:         actionComment,
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_INVOICE_FAILED),
					actionLog.ActionDetail.Set(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
					actionLog.UserID.Set(userID),
					actionLog.PaymentSequenceNumber.Set(5),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for cancel payment",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_PAYMENT_CANCELLED,
				PaymentSequenceNumber: 5,
				ActionComment:         actionComment,
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_PAYMENT_CANCELLED),
					actionLog.ActionDetail.Set(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
					actionLog.UserID.Set(userID),
					actionLog.PaymentSequenceNumber.Set(5),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for invoice refunded",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_INVOICE_REFUNDED,
				PaymentSequenceNumber: 6,
				ActionComment:         actionComment,
				PaymentMethod:         invoice_pb.PaymentMethod_CASH.String(),
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_INVOICE_REFUNDED),
					actionLog.ActionDetail.Set(invoice_pb.PaymentMethod_CASH.String()),
					actionLog.UserID.Set(userID),
					actionLog.PaymentSequenceNumber.Set(6),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "happy case for payment added",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: nil,
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_PAYMENT_ADDED,
				PaymentSequenceNumber: 6,
				ActionComment:         actionComment,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
			},
			setup: func(ctx context.Context) {
				actionLog := new(entities.InvoiceActionLog)
				database.AllNullEntity(actionLog)

				if err := multierr.Combine(
					actionLog.InvoiceID.Set(invoiceID),
					actionLog.ActionComment.Set(actionComment),
					actionLog.Action.Set(invoice_pb.InvoiceAction_PAYMENT_ADDED),
					actionLog.ActionDetail.Set(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
					actionLog.UserID.Set(userID),
					actionLog.PaymentSequenceNumber.Set(6),
				); err != nil {
					t.Error(err)
				}

				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, actionLog).Once().Return(nil)
			},
		},
		{
			name:        "Failed to create action log invalid invoice id",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: errors.New("invalid invoice id"),
			req: &InvoiceActionLogDetails{
				InvoiceID:     "",
				Action:        invoice_pb.InvoiceAction_INVOICE_ISSUED,
				ActionComment: actionComment,
			},
			setup: func(ctx context.Context) {
				// Do nothing, no setup.
			},
		},
		{
			name:        "Failed to create action log invalid payment sequence number",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: errors.New("invalid payment sequence number"),
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                invoice_pb.InvoiceAction_PAYMENT_APPROVED,
				PaymentSequenceNumber: 0,
				PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
				ActionComment:         actionComment,
			},
			setup: func(ctx context.Context) {
				// Do nothing, no setup.
			},
		},
		{
			name:        "Failed to create action log closed db pool",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: puddle.ErrClosedPool,
			req: &InvoiceActionLogDetails{
				InvoiceID:     invoiceID,
				Action:        invoice_pb.InvoiceAction_INVOICE_ISSUED,
				ActionComment: actionComment,
			},
			setup: func(ctx context.Context) {
				mockInvoiceActionLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(puddle.ErrClosedPool)
			},
		},
		{
			name:        "Invalid Action Detail",
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			expectedErr: errors.New("invalid invoice action detail"),
			req: &InvoiceActionLogDetails{
				InvoiceID:             invoiceID,
				Action:                99,
				PaymentSequenceNumber: 1127,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
				ActionComment:         actionComment,
			},
			setup: func(ctx context.Context) {
				// Do nothing, no setup.
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := CreateActionLogV2(testCase.ctx, mockDB, testCase.req.(*InvoiceActionLogDetails), mockInvoiceActionLogRepo)
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
