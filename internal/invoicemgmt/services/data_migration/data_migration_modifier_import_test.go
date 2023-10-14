package services

import (
	"context"
	"fmt"
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

func TestDataMigrationModifierService_Import_Payment(t *testing.T) {
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
	zapLogger := logger.NewZapLogger("debug", true)
	mockTx := &mock_database.Tx{}

	s := &DataMigrationModifierService{
		logger:      *zapLogger.Sugar(),
		DB:          mockDB,
		InvoiceRepo: mockInvoiceRepo,
		PaymentRepo: mockPaymentRepo,
	}

	testInvoices := []*entities.Invoice{
		{
			StudentID: database.Text("test-student"),
			InvoiceID: database.Text("test-invoice-id"),
			Total:     database.Numeric(1000),
			Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
		},
		{
			StudentID: database.Text("test-student2"),
			InvoiceID: database.Text("test-invoice-id2"),
			Total:     database.Numeric(500),
			Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
		},
		{
			StudentID: database.Text("test-student3"),
			InvoiceID: database.Text("test-invoice-id3"),
			Total:     database.Numeric(1000),
			Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
		},
	}

	testSinglePayload := `payment_csv_id,payment_id,invoice_id,payment_method,payment_status,due_date,expiry_date,payment_date,student_id,payment_sequence_number,is_exported,created_at,result_code,amount,reference
	0,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id`

	testSinglePayloadWithSuccessfulPayment := `payment_csv_id,payment_id,invoice_id,payment_method,payment_status,due_date,expiry_date,payment_date,student_id,payment_sequence_number,is_exported,created_at,result_code,amount,reference
	0,,,CASH,PAYMENT_SUCCESSFUL,2009-12-30,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id`

	testMultiPayload := `payment_csv_id,payment_id,invoice_id,payment_method,payment_status,due_date,expiry_date,payment_date,student_id,payment_sequence_number,is_exported,created_at,result_code,amount,reference
	0,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id
	1,,,BANK_TRANSFER,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student2,,true,2009-12-30,,,test-invoice-id2
	2,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student3,,true,2009-12-30,,,test-invoice-id3`

	csvJustheader := `payment_csv_id,payment_id,invoice_id,payment_method,payment_status,due_date,expiry_date,payment_date,student_id,payment_sequence_number,is_exported,created_at,result_code,amount,reference`

	csvInvalidHeaderCount := `payment_csv_id
	%v`

	csvInvalidHeader := `payment_csv_id,payment_id,invoice_id,payment_method,payment_status,due_date,expiry_date,payment_date,student_id,payment_sequence_number,is_exported,created_at,result_code,amount,test
	,,,,,,,,,,,,,,`

	testMultiPayloadRequiredValues := `payment_csv_id,payment_id,invoice_id,payment_method,payment_status,due_date,expiry_date,payment_date,student_id,payment_sequence_number,is_exported,created_at,result_code,amount,reference
	0,,,,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id
	0,,,CASH,,2009-12-30,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id
	0,,,CASH,PAYMENT_FAILED,,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id
	0,,,CASH,PAYMENT_FAILED,2009-12-30,,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id
	0,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,,,true,2009-12-30,,,test-invoice-id
	0,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student,,,2009-12-30,,,test-invoice-id
	0,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student,,true,,,,test-invoice-id
	0,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,`

	testMultiPayloadPaymentStatus := `payment_csv_id,payment_id,invoice_id,payment_method,payment_status,due_date,expiry_date,payment_date,student_id,payment_sequence_number,is_exported,created_at,result_code,amount,reference
	0,,,CASH,PAYMENT_FAILED,2009-12-30,2009-12-31,2009-12-31,test-student,,true,2009-12-30,,,test-invoice-id`

	testcases := []TestCase{
		{
			name: "happy case - payment is successfully created single row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Payload:    []byte(fmt.Sprintf(testSinglePayload)),
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors:     []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, mockDB, mock.Anything).Once().Return(testInvoices[0], nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - payment is successfully created multiple row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Payload:    []byte(fmt.Sprintf(testMultiPayload)),
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors:     []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{},
			},
			setup: func(ctx context.Context) {
				for i := 0; i < 3; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, mockDB, mock.Anything).Once().Return(testInvoices[i], nil)
				}
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < 3; i++ {
					mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - header no values csv",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(csvJustheader)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedErr: status.Error(codes.InvalidArgument, "no data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty CSV file",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    nil,
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "no data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(csvInvalidHeaderCount, "test-1")),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "PAYMENT_ENTITY - csv file invalid format - number of column should be 15"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(csvInvalidHeader)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedErr: status.Error(codes.InvalidArgument, "PAYMENT_ENTITY - csv file invalid format - test column (toLowerCase) should be 'reference'"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - multiple csv values required",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testMultiPayloadRequiredValues)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: payment_method",
					},
					{
						RowNumber: 3,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: payment_status",
					},
					{
						RowNumber: 4,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: due_date",
					},
					{
						RowNumber: 5,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: expiry_date",
					},
					{
						RowNumber: 6,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: student_id",
					},
					{
						RowNumber: 7,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: is_exported",
					},
					{
						RowNumber: 8,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: created_at",
					},
					{
						RowNumber: 9,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: reference",
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - invoice status mismatch to payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testMultiPayloadPaymentStatus)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = Internal desc = invalid invoice status: DRAFT for payment status: PAYMENT_FAILED",
					},
				},
			},
			setup: func(ctx context.Context) {
				testInvoices[0].Status = database.Text(invoice_pb.InvoiceStatus_DRAFT.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, s.DB, mock.Anything).Once().Return(testInvoices[0], nil)
			},
		},
		{
			name: "negative test - invoice status mismatch to payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testMultiPayloadPaymentStatus)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = Internal desc = invalid invoice status: DRAFT for payment status: PAYMENT_FAILED",
					},
				},
			},
			setup: func(ctx context.Context) {
				testInvoices[0].Status = database.Text(invoice_pb.InvoiceStatus_DRAFT.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, s.DB, mock.Anything).Once().Return(testInvoices[0], nil)
			},
		},
		{
			name: "negative test - issued invoice status mismatch to payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testMultiPayloadPaymentStatus)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = Internal desc = ISSUED invoice should have payment status PAYMENT_PENDING but got: PAYMENT_FAILED",
					},
				},
			},
			setup: func(ctx context.Context) {
				testInvoices[0].Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, s.DB, mock.Anything).Once().Return(testInvoices[0], nil)
			},
		},
		{
			name: "negative test - paid invoice status mismatch to payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testMultiPayloadPaymentStatus)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = Internal desc = PAID invoice should have payment status PAYMENT_SUCCESSFUL but got: PAYMENT_FAILED",
					},
				},
			},
			setup: func(ctx context.Context) {
				testInvoices[0].Status = database.Text(invoice_pb.InvoiceStatus_PAID.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, s.DB, mock.Anything).Once().Return(testInvoices[0], nil)
			},
		},
		{
			name: "negative test - refunded invoice status mismatch to payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testMultiPayloadPaymentStatus)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = Internal desc = REFUNDED invoice should have payment status PAYMENT_SUCCESSFUL but got: PAYMENT_FAILED",
					},
				},
			},
			setup: func(ctx context.Context) {
				testInvoices[0].Status = database.Text(invoice_pb.InvoiceStatus_REFUNDED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, s.DB, mock.Anything).Once().Return(testInvoices[0], nil)
			},
		},
		{
			name: "negative test - failed invoice status mismatch to payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testSinglePayloadWithSuccessfulPayment)),
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = Internal desc = FAILED invoice should have payment status PAYMENT_FAILED but got: PAYMENT_SUCCESSFUL",
					},
				},
			},
			setup: func(ctx context.Context) {
				testInvoices[0].Status = database.Text(invoice_pb.InvoiceStatus_FAILED.String())
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceReferenceID", ctx, s.DB, mock.Anything).Once().Return(testInvoices[0], nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ImportDataMigration(testCase.ctx, testCase.req.(*invoice_pb.ImportDataMigrationRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if response == nil {
					fmt.Println(err)
				}

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			if testCase.expectedResp != nil {
				assert.Equal(t, compareImportDataMigrationResponseErr(testCase.expectedResp.(*invoice_pb.ImportDataMigrationResponse), response), true)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo)
		})
	}
}

func TestDataMigrationModifierService_Import_Invoice(t *testing.T) {
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
	zapLogger := logger.NewZapLogger("debug", true)
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)
	mockTx := &mock_database.Tx{}

	s := &DataMigrationModifierService{
		logger:       *zapLogger.Sugar(),
		DB:           mockDB,
		InvoiceRepo:  mockInvoiceRepo,
		PaymentRepo:  mockPaymentRepo,
		BillItemRepo: mockBillItemRepo,
	}

	testSingleInvoicePayload := `invoice_csv_id,invoice_id,student_id,type,status,sub_total,total,created_at,invoice_sequence_number,is_exported,reference1,reference2
	1,,test-student-1,MANUAL,ISSUED,1000,1000,2009-12-30,,TRUE,1,1`

	testMultiInvoicePayload := `invoice_csv_id,invoice_id,student_id,type,status,sub_total,total,created_at,invoice_sequence_number,is_exported,reference1,reference2
	1,,test-student-1,MANUAL,ISSUED,1000,1000,2009-12-30,,TRUE,1,1
	2,,test-student-2,MANUAL,ISSUED,1000,1000,2009-12-30,,TRUE,1,1
	3,,test-student-3,MANUAL,ISSUED,1000,1000,2009-12-30,,TRUE,1,1`

	testInvoices := []*entities.Invoice{
		{
			StudentID: database.Text("test-student"),
			InvoiceID: database.Text("test-invoice-id"),
			Total:     database.Numeric(1000),
			SubTotal:  database.Numeric(1000),
		},
		{
			StudentID: database.Text("test-student2"),
			InvoiceID: database.Text("test-invoice-id2"),
			Total:     database.Numeric(500),
			SubTotal:  database.Numeric(500),
		},
		{
			StudentID: database.Text("test-student3"),
			InvoiceID: database.Text("test-invoice-id3"),
			Total:     database.Numeric(1000),
			SubTotal:  database.Numeric(1000),
		},
	}

	invoiceCsvJustheader := `invoice_csv_id,invoice_id,student_id,type,status,sub_total,total,created_at,invoice_sequence_number,is_exported,reference1,reference2`

	invoiceCsvInvalidHeaderCount := `invoice_csv_id
	%v`

	invoiceCsvInvalidHeader := `invoice_csv_id,invoice_id,student_id,type,status,sub_total,total,created_at,invoice_sequence_number,is_exported,reference1,reference252
	,,,,,,,,,,,`

	testInvoiceMultiPayloadRequiredValues := `invoice_csv_id,invoice_id,student_id,type,status,sub_total,total,created_at,invoice_sequence_number,is_exported,reference1,reference2
	1,,,MANUAL,ISSUED,1000,1000,2009-12-30,,TRUE,1,1
	2,,test-student-2,,ISSUED,1000,1000,2009-12-30,,TRUE,1,1
	3,,test-student-3,MANUAL,,1000,1000,2009-12-30,,TRUE,1,1
	4,,test-student-3,MANUAL,ISSUED,,1000,2009-12-30,,TRUE,1,1
	5,,test-student-3,MANUAL,ISSUED,1000,,2009-12-30,,TRUE,1,1
	6,,test-student-3,MANUAL,ISSUED,1000,1000,,,TRUE,1,1
	6,,test-student-3,MANUAL,ISSUED,1000,1000,2009-12-30,,,1,1
	6,,test-student-3,MANUAL,ISSUED,1000,1000,2009-12-30,,TRUE,,1`

	testTotalFinalPrice := database.Numeric(1000)

	testcases := []TestCase{
		{
			name: "happy case - invoice is successfully created single row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
				Payload:    []byte(fmt.Sprintf(testSingleInvoicePayload)),
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
				Errors:     []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{},
			},
			setup: func(ctx context.Context) {
				mockBillItemRepo.On("GetBillItemTotalByStudentAndReference", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testTotalFinalPrice, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(database.Text("invoice-id"), nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - invoice is successfully created multiple row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
				Payload:    []byte(fmt.Sprintf(testMultiInvoicePayload)),
			},
			setup: func(ctx context.Context) {
				for i := 0; i < 3; i++ {
					mockBillItemRepo.On("GetBillItemTotalByStudentAndReference", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testTotalFinalPrice, nil)
				}
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < 3; i++ {
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(database.Text(testInvoices[i].InvoiceID.String), nil)
				}
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - header no values invoice csv",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(invoiceCsvJustheader)),
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
			},
			expectedErr: status.Error(codes.InvalidArgument, "no data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty invoice CSV file",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    nil,
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "no data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid invoice header count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(invoiceCsvInvalidHeaderCount, "test-1")),
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "INVOICE_ENTITY - csv file invalid format - number of column should be 12"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid invoice header",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(invoiceCsvInvalidHeader)),
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
			},
			expectedErr: status.Error(codes.InvalidArgument, "INVOICE_ENTITY - csv file invalid format - reference252 column (toLowerCase) should be 'reference2'"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - multiple invoice csv values required",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportDataMigrationRequest{
				Payload:    []byte(fmt.Sprintf(testInvoiceMultiPayloadRequiredValues)),
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
			},
			expectedResp: &invoice_pb.ImportDataMigrationResponse{
				EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
				Errors: []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
					{
						RowNumber: 2,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: student_id",
					},
					{
						RowNumber: 3,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: type",
					},
					{
						RowNumber: 4,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: status",
					},
					{
						RowNumber: 5,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: sub_total",
					},
					{
						RowNumber: 6,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: total",
					},
					{
						RowNumber: 7,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: created_at",
					},
					{
						RowNumber: 8,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: is_exported",
					},
					{
						RowNumber: 9,
						Error:     "rpc error: code = InvalidArgument desc = missing mandatory data: reference1",
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ImportDataMigration(testCase.ctx, testCase.req.(*invoice_pb.ImportDataMigrationRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if response == nil {
					fmt.Println(err)
				}

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			if testCase.expectedResp != nil {
				assert.Equal(t, compareImportDataMigrationResponseErr(testCase.expectedResp.(*invoice_pb.ImportDataMigrationResponse), response), true)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockBillItemRepo)
		})
	}
}

func compareImportDataMigrationResponseErr(expectedResp *invoice_pb.ImportDataMigrationResponse, actualResp *invoice_pb.ImportDataMigrationResponse) bool {
	if len(expectedResp.Errors) != len(actualResp.Errors) {
		log.Fatalf("Errors length: expected %v but got %v\n", len(expectedResp.Errors), len(actualResp.Errors))
		return false
	}
	for i := 0; i < len(expectedResp.Errors); i++ {
		if expectedResp.Errors[i].RowNumber != actualResp.Errors[i].RowNumber {
			log.Fatalf("RowNumber: expected %v but got %v at line %v\n", expectedResp.Errors[i].RowNumber, actualResp.Errors[i].RowNumber, i+1)
			return false
		}

		if expectedResp.Errors[i].Error != actualResp.Errors[i].Error {
			log.Fatalf("Error: expected %v but got %v at line %v\n", expectedResp.Errors[i].Error, actualResp.Errors[i].Error, i+1)
			return false
		}
	}

	return true
}
