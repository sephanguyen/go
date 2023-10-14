package services

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDataMigrationModifierService_InsertPaymentData(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Mock objects
	mockDB := new(mock_database.Ext)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockTx := &mock_database.Tx{}

	testError := errors.New("test error")

	testInvoice := &entities.Invoice{
		StudentID: database.Text(""),
		InvoiceID: database.Text("1"),
		Total:     database.Numeric(1000),
		Status:    database.Text(invoice_pb.InvoiceStatus_PAID.String()),
	}
	zapLogger := logger.NewZapLogger("debug", true)
	s := &DataMigrationModifierService{
		DB:          mockDB,
		InvoiceRepo: mockInvoiceRepo,
		PaymentRepo: mockPaymentRepo,
		logger:      *zapLogger.Sugar(),
	}

	testcases := []TestCase{
		{
			name: "happy case - payment is successfully created",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{
					"", "", "", "CASH", "PAYMENT_SUCCESSFUL", "2009-12-30", "2009-12-31", "2009-12-31", "test-student-exist", "", "true", "2009-12-30", "", "", "test-reference-id",
				},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{},
			setup: func(ctx context.Context) {
				testInvoice.StudentID = database.Text("test-student-exist")
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - no payment method",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: payment_method").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: payment_status").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment due date",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_FAILED", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: due_date").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment expiry date",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_FAILED", "2009-12-31", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: expiry_date").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment student id",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_FAILED", "2009-12-30", "2009-12-31", "", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: student_id").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment is exported",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_FAILED", "2009-12-30", "2009-12-31", "", "test-student", "", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: is_exported").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment created at",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_FAILED", "2009-12-30", "2009-12-31", "", "test-student", "", "true", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: created_at").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment invoice reference",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_FAILED", "2009-12-30", "2009-12-31", "", "test-student", "", "true", "2009-12-30", "", "", ""},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: reference").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no payment date if payment status is successful",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_SUCCESSFUL", "2009-12-30", "2009-12-31", "", "test-student", "", "true", "2009-12-30", "", "", "test-reference-id"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "payment invoice reference: test-reference-id with successful status should have a payment date").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - payment due date should not be greater than payment expiry date",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_SUCCESSFUL", "2009-12-31", "2009-12-30", "2009-12-31", "test-student", "", "true", "2009-12-30", "", "", "test-reference-id"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "invalid payment due date: 2009-12-31 00:00:00 +0900 JST must be before expiry date: 2009-12-30 00:00:00 +0900 JST on invoice reference: test-reference-id").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - error mismatch student id",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_SUCCESSFUL", "2009-12-30", "2009-12-31", "2009-12-31", "test-student", "", "true", "2009-12-30", "", "", "test-reference-id"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.Internal, "error student id: test-student mismatch on invoice student id: mismatch-student").Error(),
				},
			},
			setup: func(ctx context.Context) {
				testInvoice.StudentID = database.Text("mismatch-student")
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
			},
		},
		{
			name: "negative test - error retrieve invoice",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"", "", "", "CASH", "PAYMENT_SUCCESSFUL", "2009-12-30", "2009-12-31", "2009-12-31", "test-student", "", "true", "2009-12-30", "", "", "test-reference-id"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.Internal, "error retrieving invoice with reference: test-reference-id").Error(),
				},
			},
			setup: func(ctx context.Context) {
				testInvoice.StudentID = database.Text("test-student-exist")
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "negative test - payment cannot create successfully",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{
					"", "", "", "CASH", "PAYMENT_SUCCESSFUL", "2009-12-30", "2009-12-31", "2009-12-31", "test-student-exist", "", "true", "2009-12-30", "", "", "test-reference-id",
				},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{},
			expectedErr:        status.Error(codes.Internal, "Data Migration error: test error when creating a valid payment invoice with reference: test-reference-id"),
			setup: func(ctx context.Context) {
				testInvoice.StudentID = database.Text("test-student-exist")
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, mockDB, mock.Anything).Once().Return(testInvoice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			errorLines, err := s.InsertPaymentDataMigration(testCase.ctx, s.DB, testCase.csvLine)

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			if len(testCase.expectedErrorLines) == 0 {
				assert.Equal(t, testCase.expectedErrorLines, errorLines)
			} else {
				assert.Equal(t, compareImportDataMigrationErr(testCase.expectedErrorLines, errorLines), true)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo)
		})
	}
}

func compareImportDataMigrationErr(expectedErr []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError, actualErr []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError) bool {
	if len(expectedErr) != len(actualErr) {
		log.Fatalf("Errors length: expected %v but got %v\n", len(expectedErr), len(actualErr))
		return false
	}

	for i := 0; i < len(expectedErr); i++ {
		if expectedErr[i].RowNumber != actualErr[i].RowNumber {
			log.Fatalf("RowNumber: expected %v but got %v at line %v\n", expectedErr[i].RowNumber, actualErr[i].RowNumber, i+1)
			return false
		}

		if expectedErr[i].Error != actualErr[i].Error {
			log.Fatalf("Error: expected %v but got %v at line %v\n", expectedErr[i].Error, actualErr[i].Error, i+1)
			return false
		}
	}

	return true
}
