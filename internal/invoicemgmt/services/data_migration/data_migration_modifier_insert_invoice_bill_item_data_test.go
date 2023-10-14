package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDataMigrationModifierService_InsertInvoiceBillItemDataMigration(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Mock objects
	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)

	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockStudentRepo := new(mock_repositories.MockStudentRepo)
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)
	mockInvoiceBillItemRepo := new(mock_repositories.MockInvoiceBillItemRepo)

	zapLogger := logger.NewZapLogger("debug", true)
	s := &DataMigrationModifierService{
		DB:                  mockDB,
		InvoiceRepo:         mockInvoiceRepo,
		PaymentRepo:         mockPaymentRepo,
		BillItemRepo:        mockBillItemRepo,
		InvoiceBillItemRepo: mockInvoiceBillItemRepo,
		logger:              *zapLogger.Sugar(),
		StudentRepo:         mockStudentRepo,
	}

	student1 := "test-student-id-1"
	invoices := []*entities.Invoice{
		{
			InvoiceID:           database.Text("test-invoice-id"),
			Type:                database.Text(invoice_pb.InvoiceType_MANUAL.String()),
			Status:              database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
			StudentID:           database.Text(student1),
			SubTotal:            database.Numeric(100),
			Total:               database.Numeric(100),
			InvoiceReferenceID:  database.Text("1"),
			InvoiceReferenceID2: database.Text("1"),
		},
	}

	billItemsOfStudent1 := []*entities.BillItem{
		{
			BillItemSequenceNumber: database.Int4(1),
			StudentID:              database.Text(student1),
			BillStatus:             database.Text(payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()),
			FinalPrice:             database.Numeric(20),
		},
		{
			BillItemSequenceNumber: database.Int4(2),
			StudentID:              database.Text(student1),
			BillStatus:             database.Text(payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()),
			FinalPrice:             database.Numeric(30),
		},
		{
			BillItemSequenceNumber: database.Int4(3),
			StudentID:              database.Text(student1),
			BillStatus:             database.Text(payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()),
			FinalPrice:             database.Numeric(50),
		},
	}

	testError := errors.New("test-error")

	testcases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)

				mockInvoiceRepo.On("RetrievedMigratedInvoices", ctx, mockTx, mock.Anything).Once().Return(invoices, nil)
				mockBillItemRepo.On("RetrieveBillItemsByInvoiceReferenceNum", ctx, mockTx, mock.Anything).Once().Return(billItemsOfStudent1, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.InvoiceBillItems{}, nil)
				for i := 0; i < len(billItemsOfStudent1); i++ {
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - there are no migrated invoices",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)

				mockInvoiceRepo.On("RetrievedMigratedInvoices", ctx, mockTx, mock.Anything).Once().Return([]*entities.Invoice{}, nil)

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - there are bill items already mapped to an invoice",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)

				mockInvoiceRepo.On("RetrievedMigratedInvoices", ctx, mockTx, mock.Anything).Once().Return(invoices, nil)
				mockBillItemRepo.On("RetrieveBillItemsByInvoiceReferenceNum", ctx, mockTx, mock.Anything).Once().Return(billItemsOfStudent1, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.InvoiceBillItems{
					{
						InvoiceID:              database.Text("test-invoice-id-1"),
						BillItemSequenceNumber: database.Int4(1),
					},
				}, nil)
				for i := 0; i < len(billItemsOfStudent1)-1; i++ {
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test - error on InvoiceRepo.RetrievedMigratedInvoices",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			expectedErr: fmt.Errorf("InvoiceRepo.RetrievedMigratedInvoices err: %v", testError),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)

				mockInvoiceRepo.On("RetrievedMigratedInvoices", ctx, mockTx, mock.Anything).Once().Return(nil, testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test - error on BillItemRepo.RetrieveBillItemsByInvoiceReferenceNum",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			expectedErr: fmt.Errorf("error BillItemRepo.RetrieveBillItemsByInvoiceReferenceNum err: %v reference_id: %v", testError, "1"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)

				mockInvoiceRepo.On("RetrievedMigratedInvoices", ctx, mockTx, mock.Anything).Once().Return(invoices, nil)
				mockBillItemRepo.On("RetrieveBillItemsByInvoiceReferenceNum", ctx, mockTx, mock.Anything).Once().Return(nil, testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test - error on InvoiceBillItemRepo.FindAllByInvoiceID",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			expectedErr: fmt.Errorf("error InvoiceBillItemRepo.FindAllByInvoiceID err: %v reference_id: %v", testError, "1"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)

				mockInvoiceRepo.On("RetrievedMigratedInvoices", ctx, mockTx, mock.Anything).Once().Return(invoices, nil)
				mockBillItemRepo.On("RetrieveBillItemsByInvoiceReferenceNum", ctx, mockTx, mock.Anything).Once().Return(billItemsOfStudent1, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name:        "negative test - error on InvoiceBillItemRepo.Create",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			expectedErr: fmt.Errorf("error InvoiceBillItemRepo.Create err: %v reference_id: %v", testError, "1"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)

				mockInvoiceRepo.On("RetrievedMigratedInvoices", ctx, mockTx, mock.Anything).Once().Return(invoices, nil)
				mockBillItemRepo.On("RetrieveBillItemsByInvoiceReferenceNum", ctx, mockTx, mock.Anything).Once().Return(billItemsOfStudent1, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.InvoiceBillItems{}, nil)
				mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.InsertInvoiceBillItemDataMigration(testCase.ctx)

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTx, mockInvoiceRepo, mockPaymentRepo, mockStudentRepo, mockBillItemRepo, mockInvoiceBillItemRepo)
		})
	}
}
