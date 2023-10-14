package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDataMigrationModifierService_InsertInvoiceData(t *testing.T) {
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
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)
	mockTx := &mock_database.Tx{}

	zapLogger := logger.NewZapLogger("debug", true)
	s := &DataMigrationModifierService{
		DB:           mockDB,
		InvoiceRepo:  mockInvoiceRepo,
		PaymentRepo:  mockPaymentRepo,
		logger:       *zapLogger.Sugar(),
		BillItemRepo: mockBillItemRepo,
	}

	time := time.Now().Format("2006-01-02")

	testError := errors.New("test-error")
	// testStudent := &entities.Student{StudentID: database.Text("test-student-1")}

	testTotalFinalPrice := database.Numeric(1000)

	testcases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{},
			setup: func(ctx context.Context) {
				mockBillItemRepo.On("GetBillItemTotalByStudentAndReference", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testTotalFinalPrice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(database.Text("invoice-id"), nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - missing student ID",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "  ", "MANUAL", "ISSUED", "1000", "1000", time, "", "TRUE", "1", "1"},
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
			name: "negative test - missing invoice type",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "  ", "ISSUED", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: type").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - missing invoice status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "  ", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: status").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - missing invoice sub_total",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "    ", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: sub_total").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - missing invoice total",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", " ", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: total").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - missing created_at",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "1000", " ", "", "TRUE", "1", "1"},
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
			name: "negative test - missing reference1",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "1000", time, "", "TRUE", " ", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "missing mandatory data: reference1").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - sub_total is invalid format",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "text sub_total", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "error parsing string to float64 sub_total: strconv.ParseFloat: parsing \"text sub_total\": invalid syntax").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - total is invalid format",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "text total", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "error parsing string to float64 total: strconv.ParseFloat: parsing \"text total\": invalid syntax").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - created_at is invalid format",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "1000", "invalid date format", "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "error parsing string to date created_at: parsing time \"invalid date format\" as \"2006-01-02\": cannot parse \"invalid date format\" as \"2006\"").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - is_exported is invalid",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "1000", time, "", "Invalid Bool", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "error parsing string to bool is_exported: strconv.ParseBool: parsing \"Invalid Bool\": invalid syntax").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - invoice status is invalid",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "TestStatus", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "invoice status TestStatus is invalid").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - type is invalid",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "TestType", "ISSUED", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "invoice type TestType is invalid").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - PAID invoice has negative total and sub_total",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "PAID", "-1000", "-1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "total or sub_total should not be negative").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - REFUNDED invoice has positive total and sub_total",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "REFUNDED", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.InvalidArgument, "total or sub_total should not be positive").Error(),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - error on finding student bill item",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				{
					RowNumber: 2,
					Error:     status.Error(codes.Internal, "cannot retrieve bill items with student id: 12345 and invoice reference: 1").Error(),
				},
			},
			setup: func(ctx context.Context) {
				mockBillItemRepo.On("GetBillItemTotalByStudentAndReference", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testTotalFinalPrice, testError)
			},
		},
		{
			name: "negative test - error on InvoiceRepo.Create",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			csvLine: [][]string{
				{"1", "", "12345", "MANUAL", "ISSUED", "1000", "1000", time, "", "TRUE", "1", "1"},
			},
			expectedErrorLines: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{},
			expectedErr:        status.Error(codes.Internal, "Data Migration error: test-error when creating a valid invoice with reference: 1"),
			setup: func(ctx context.Context) {
				mockBillItemRepo.On("GetBillItemTotalByStudentAndReference", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testTotalFinalPrice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(database.Text(""), testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			errorLines, err := s.InsertInvoiceDataMigration(testCase.ctx, s.DB, testCase.csvLine)

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

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockBillItemRepo)
		})
	}
}
