package paymentsvc

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	pfutils "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	gocsv "github.com/gocarina/gocsv"
	"github.com/ianlopshire/go-fixedwidth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPaymentModifierService_CreateBulkPaymentValidation(t *testing.T) {

	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLog := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkValidationsRepo := new(mock_repositories.MockBulkPaymentValidationsRepo)
	mockBulkValidationsDetailRepo := new(mock_repositories.MockBulkPaymentValidationsDetailRepo)
	mockUserBasicInfoRepo := new(mock_repositories.MockUserBasicInfoRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	responseDDSimple := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful: true,
		PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
			{
				Amount:                1000,
				InvoiceSequenceNumber: 1,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
				PaymentSequenceNumber: 1,
				Result:                "D-R0",
				StudentId:             "1",
				StudentName:           "Name-test",
				InvoiceId:             "1",
				PaymentStatus:         "PAYMENT_SUCCESSFUL",
			},
		},
	}

	responseDDMult := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful: true,
		PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
			{
				Amount:                1000,
				InvoiceSequenceNumber: 2,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
				PaymentSequenceNumber: 2,
				Result:                "D-R0-2",
				StudentId:             "2",
				StudentName:           "Name-test",
				InvoiceId:             "2",
				PaymentStatus:         "PAYMENT_FAILED",
			},
			{
				Amount:                1000,
				InvoiceSequenceNumber: 1,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
				PaymentSequenceNumber: 1,
				Result:                "D-R0",
				StudentId:             "1",
				StudentName:           "Name-test",
				InvoiceId:             "1",
				PaymentStatus:         "PAYMENT_SUCCESSFUL",
			},
		},
	}

	responseCCSimple := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful: true,
		PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
			{
				Amount:                1000,
				InvoiceSequenceNumber: 1,
				PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				PaymentSequenceNumber: 1,
				Result:                "C-R0",
				StudentId:             "1",
				StudentName:           "Name-test",
				InvoiceId:             "1",
				PaymentStatus:         "PAYMENT_SUCCESSFUL",
			},
		},
	}

	responseDDAmountMismatched := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful: true,
		PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
			{
				Amount:                1000,
				InvoiceSequenceNumber: 1,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
				PaymentSequenceNumber: 1,
				Result:                "D-R0-1",
				StudentId:             "1",
				StudentName:           "Name-test",
				InvoiceId:             "1",
				PaymentStatus:         "PAYMENT_FAILED",
			},
		},
	}

	responseDDAmountMismatchedAndInvoiceNotIssued := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful: true,
		PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
			{
				Amount:                1000,
				InvoiceSequenceNumber: 1,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
				PaymentSequenceNumber: 1,
				Result:                "D-R0-3",
				StudentId:             "1",
				StudentName:           "Name-test",
				InvoiceId:             "1",
				PaymentStatus:         "PAYMENT_FAILED",
			},
		},
	}

	responseDDVoidInvoiceFailedPayment := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful: true,
		PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
			{
				Amount:                1000,
				InvoiceSequenceNumber: 1,
				PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
				PaymentSequenceNumber: 1,
				Result:                "D-R0-2",
				StudentId:             "1",
				StudentName:           "Name-test",
				InvoiceId:             "1",
				PaymentStatus:         "PAYMENT_FAILED",
			},
		},
	}

	responseCCFailedInvoiceFailedPayment := &invoice_pb.CreateBulkPaymentValidationResponse{
		Successful: true,
		PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
			{
				Amount:                1000,
				InvoiceSequenceNumber: 1,
				PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				PaymentSequenceNumber: 1,
				Result:                "C-R0-2",
				StudentId:             "1",
				StudentName:           "Name-test",
				InvoiceId:             "1",
				PaymentStatus:         "PAYMENT_FAILED",
			},
		},
	}

	payment := &entities.Payment{
		PaymentID:             database.Text("1"),
		PaymentSequenceNumber: database.Int4(1),
	}
	paymentTwo := &entities.Payment{
		PaymentID:             database.Text("2"),
		PaymentSequenceNumber: database.Int4(2),
	}

	user := &entities.UserBasicInfo{
		UserID: database.Text("1"),
		Name:   database.Text("Name-test"),
	}

	invoice := &entities.Invoice{
		InvoiceID:             database.Text("1"),
		InvoiceSequenceNumber: database.Int4(1),
		Total:                 database.Numeric(1000),
		Status:                database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		StudentID:             database.Text("1"),
	}
	invoiceTwo := &entities.Invoice{
		InvoiceID:             database.Text("2"),
		InvoiceSequenceNumber: database.Int4(2),
		Total:                 database.Numeric(1000),
		Status:                database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		StudentID:             database.Text("2"),
	}

	s := &PaymentModifierService{
		DB:                               mockDB,
		InvoiceRepo:                      mockInvoiceRepo,
		PaymentRepo:                      mockPaymentRepo,
		InvoiceActionLogRepo:             mockInvoiceActionLog,
		BulkPaymentValidationsRepo:       mockBulkValidationsRepo,
		BulkPaymentValidationsDetailRepo: mockBulkValidationsDetailRepo,
		UserBasicInfoRepo:                mockUserBasicInfoRepo,
		UnleashClient:                    mockUnleashClient,
	}

	testcases := []TestCase{
		{
			name: "happy case - direct debit",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 1000),
			},
			expectedResp: responseDDSimple,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(user, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - direct debit multiple",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(2, "0", 1000),
			},
			expectedResp: responseDDMult,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())

				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String())
				paymentTwo.PaymentMethod = database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String())

				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(2).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Once().Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoice, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(user, nil)

				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Once().Return(paymentTwo, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(invoiceTwo, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(user, nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - convenience store",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Payload:                createFileForConvenienceStore(),
			},
			expectedResp: responseCCSimple,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())

				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(user, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name: "happy case - amount mismatched",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 500),
			},
			expectedResp: responseDDAmountMismatched,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				payment := &entities.Payment{
					PaymentID:             database.Text("1"),
					PaymentSequenceNumber: database.Int4(1),
					PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
				}

				invoice := &entities.Invoice{
					InvoiceID:             database.Text("1"),
					InvoiceSequenceNumber: database.Int4(1),
					Total:                 database.Numeric(1000),
					Status:                database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
					StudentID:             database.Text("1"),
				}
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(user, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - amount mismatched and invoice not issued",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 500),
			},
			expectedResp: responseDDAmountMismatchedAndInvoiceNotIssued,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				payment := &entities.Payment{
					PaymentID:             database.Text("1"),
					PaymentSequenceNumber: database.Int4(1),
					PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
				}

				invoice := &entities.Invoice{
					InvoiceID:             database.Text("1"),
					InvoiceSequenceNumber: database.Int4(1),
					Total:                 database.Numeric(1000),
					Status:                database.Text(invoice_pb.InvoiceStatus_VOID.String()),
					StudentID:             database.Text("1"),
				}
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(user, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - void invoice and failed payment",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 1000),
			},
			expectedResp: responseDDVoidInvoiceFailedPayment,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				payment := &entities.Payment{
					PaymentID:             database.Text("1"),
					PaymentSequenceNumber: database.Int4(1),
					PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
				}

				invoice := &entities.Invoice{
					InvoiceID:             database.Text("1"),
					InvoiceSequenceNumber: database.Int4(1),
					Total:                 database.Numeric(1000),
					Status:                database.Text(invoice_pb.InvoiceStatus_VOID.String()),
					StudentID:             database.Text("1"),
				}
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(user, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - failed invoice and failed payment and payment method is convenience store",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Payload:                createFileForConvenienceStore(),
			},
			expectedResp: responseCCFailedInvoiceFailedPayment,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				payment := &entities.Payment{
					PaymentID:             database.Text("1"),
					PaymentSequenceNumber: database.Int4(1),
					PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				}

				invoice := &entities.Invoice{
					InvoiceID:             database.Text("1"),
					InvoiceSequenceNumber: database.Int4(1),
					Total:                 database.Numeric(1000),
					Status:                database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
					StudentID:             database.Text("1"),
				}
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(user, nil)
				mockInvoiceActionLog.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - no payload",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Payload:                nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "file payload is required"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - invalid payment method",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_CASH,
				Payload:                nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid payment method"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - no payment date for direct debit payment method",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: nil,
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "direct debit payment date is required"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - invalid header category for direct debit payment file",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createInvalidHeaderCategoryForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "error processing the payload: invalid header record's code category: 8"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - invalid trailer category for direct debit payment file",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createInvalidTrailerCategoryForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "error processing the payload: invalid trailer data's code category: 3"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - invalid data category for direct debit payment file",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createInvalidDataCategoryForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "error processing the payload: invalid data record's data category: 9"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - invalid end record category for direct debit payment file",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createInvalidEndRecordCategoryForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "error processing the payload: invalid end record's code category: 4"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - invalid convenience store payment file",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Payload:                createInvalidFileForConvenienceStore(),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "file validation failed: invalid CONVENIENCE_STORE result code at line 1: 022"),
			setup: func(ctx context.Context) {
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "invalid invoice paid and payment successful scenario on convenience store",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Payload:       createFileForConvenienceStore(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "file validation failed: invalid invoice paid status and payment successful status on payment method: CONVENIENCE_STORE"),
			setup: func(ctx context.Context) {
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_PAID.String())
				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "invalid invoice failed and payment failed with existing result code scenario on convenience store",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Payload:       createFileForConvenienceStore(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "file validation failed: invalid invoice failed status and payment failed status with existing result code: Sample-result-code on payment method: CONVENIENCE_STORE"),
			setup: func(ctx context.Context) {
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_FAILED.String())
				payment.ResultCode = database.Text("Sample-result-code")
				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "invalid invoice paid and payment successful scenario on direct debit store",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "file validation failed: invalid invoice paid status and payment successful status on payment method: DIRECT_DEBIT"),
			setup: func(ctx context.Context) {
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_PAID.String())
				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "invalid invoice failed and payment failed with existing result code scenario on on direct debit store",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "file validation failed: invalid invoice failed status and payment failed status with existing result code: Sample-result-code on payment method: DIRECT_DEBIT"),
			setup: func(ctx context.Context) {
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_FAILED.String())
				payment.ResultCode = database.Text("Sample-result-code")
				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "DIRECT_DEBIT payment file contains a CONVENIENCE_STORE payment",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Payload:                createFileForDirectDebit(1, "0", 1000),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "processing DIRECT_DEBIT payment file but contains a record for CONVENIENCE_STORE in line 1"),
			setup: func(ctx context.Context) {
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_FAILED.String())
				payment.ResultCode = database.Text("Sample-result-code")
				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
		{
			name: "CONVENIENCE_STORE payment file contains a DIRECT_DEBIT payment",
			ctx:  ctx,
			req: &invoice_pb.CreateBulkPaymentValidationRequest{
				DirectDebitPaymentDate: timestamppb.Now(),
				PaymentMethod:          invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Payload:                createFileForConvenienceStore(),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "processing CONVENIENCE_STORE payment file but contains a record for DIRECT_DEBIT in line 1"),
			setup: func(ctx context.Context) {
				payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String())
				invoice.Status = database.Text(invoice_pb.InvoiceStatus_FAILED.String())
				payment.ResultCode = database.Text("Sample-result-code")
				payment.PaymentMethod = database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String())
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(payment, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(invoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.CreateBulkPaymentValidation(testCase.ctx, testCase.req.(*invoice_pb.CreateBulkPaymentValidationRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if testCase.expectedResp != nil {
					expectedResp := testCase.expectedResp.(*invoice_pb.CreateBulkPaymentValidationResponse)

					assert.Equal(t, len(expectedResp.PaymentValidationDetail), len(response.PaymentValidationDetail))
					for i, r := range expectedResp.PaymentValidationDetail {
						assert.Equal(t, r.InvoiceSequenceNumber, response.PaymentValidationDetail[i].InvoiceSequenceNumber)
						assert.Equal(t, r.Amount, response.PaymentValidationDetail[i].Amount)
						assert.Equal(t, r.PaymentMethod, response.PaymentValidationDetail[i].PaymentMethod)
						assert.Equal(t, r.Result, response.PaymentValidationDetail[i].Result)
						assert.Equal(t, r.StudentId, response.PaymentValidationDetail[i].StudentId)
						assert.Equal(t, r.StudentName, response.PaymentValidationDetail[i].StudentName)
						assert.Equal(t, r.PaymentSequenceNumber, response.PaymentValidationDetail[i].PaymentSequenceNumber)
						assert.Equal(t, r.InvoiceId, response.PaymentValidationDetail[i].InvoiceId)
						assert.Equal(t, r.PaymentStatus, response.PaymentValidationDetail[i].PaymentStatus)
					}
				}
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}
		})

		mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockInvoiceActionLog, mockBulkValidationsRepo, mockBulkValidationsDetailRepo, mockUserBasicInfoRepo, mockTx, mockUnleashClient)
	}
}

func createFileForConvenienceStore() []byte {
	transferredDate, _ := strconv.Atoi(fmt.Sprintf("%v%02d%02d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))
	receiveDate, _ := strconv.Atoi(fmt.Sprintf("%v%02d%02d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))
	dataRecord := []*pfutils.ConvenienceStoreFileDataRecord{
		{
			Amount:          1000,
			Category:        "02",
			CodeForUser2:    "1",
			TransferredDate: transferredDate,
			DateOfReceipt:   receiveDate,
		},
	}

	fileBytes, _ := gocsv.MarshalBytes(&dataRecord)

	return fileBytes
}

func createInvalidFileForConvenienceStore() []byte {
	transferredDate, _ := strconv.Atoi(fmt.Sprintf("%v%02d%02d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))
	receiveDate, _ := strconv.Atoi(fmt.Sprintf("%v%02d%02d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))
	dataRecord := []*pfutils.ConvenienceStoreFileDataRecord{
		{
			Amount:          1000,
			Category:        "022",
			CodeForUser2:    "1",
			TransferredDate: transferredDate,
			DateOfReceipt:   receiveDate,
		},
	}

	fileBytes, _ := gocsv.MarshalBytes(&dataRecord)

	return fileBytes
}

func createFileForDirectDebit(dataRecordCnt int, resultCode string, amount int) []byte {
	headerRecordBytes := generateDirectDebitHeader()

	dataRecordsBytes := generateDirectDebitData(dataRecordCnt, resultCode, amount)

	trailerRecordBytes := generateDirectDebitTrailerRecord()

	endRecord := pfutils.DirectDebitFileEndRecord{
		DataCategory: pfutils.DataTypeEndRecord,
	}

	endRecordBytes, _ := fixedwidth.Marshal(endRecord)

	fileContentsStr := fmt.Sprintf("%v\n", string(headerRecordBytes))

	for _, dataBytes := range dataRecordsBytes {
		fileContentsStr += fmt.Sprintf("%v\n", string(dataBytes))
	}

	fileContentsStr += fmt.Sprintf("%v\n%v", string(trailerRecordBytes), string(endRecordBytes))

	return []byte(fileContentsStr)
}

func createInvalidHeaderCategoryForDirectDebit(dataRecordCnt int, resultCode string, amount int) []byte {
	headerRecord := pfutils.DirectDebitFileHeaderRecord{
		DataCategory: pfutils.DataTypeTrailerRecord,
	}

	headerRecordBytes, _ := fixedwidth.Marshal(headerRecord)

	dataRecordsBytes := generateDirectDebitData(dataRecordCnt, resultCode, amount)

	trailerRecordBytes := generateDirectDebitTrailerRecord()

	endRecordBytes := generateDirectDebitEndRecord()

	fileContentsStr := fmt.Sprintf("%v\n", string(headerRecordBytes))

	for _, dataBytes := range dataRecordsBytes {
		fileContentsStr += fmt.Sprintf("%v\n", string(dataBytes))
	}

	fileContentsStr += fmt.Sprintf("%v\n%v", string(trailerRecordBytes), string(endRecordBytes))

	return []byte(fileContentsStr)
}

func createInvalidTrailerCategoryForDirectDebit(dataRecordCnt int, resultCode string, amount int) []byte {
	headerRecordBytes := generateDirectDebitHeader()
	dataRecordsBytes := generateDirectDebitData(dataRecordCnt, resultCode, amount)

	trailerRecord := pfutils.DirectDebitFileTrailerRecord{
		DataCategory:      3,
		TransferredAmount: 1000,
		TransferredNumber: 1,
	}

	trailerRecordBytes, _ := fixedwidth.Marshal(trailerRecord)

	endRecordBytes := generateDirectDebitEndRecord()

	fileContentsStr := fmt.Sprintf("%v\n", string(headerRecordBytes))

	for _, dataBytes := range dataRecordsBytes {
		fileContentsStr += fmt.Sprintf("%v\n", string(dataBytes))
	}

	fileContentsStr += fmt.Sprintf("%v\n%v", string(trailerRecordBytes), string(endRecordBytes))

	return []byte(fileContentsStr)
}

func createInvalidDataCategoryForDirectDebit(dataRecordCnt int, resultCode string, amount int) []byte {
	headerRecordBytes := generateDirectDebitHeader()

	var dataRecordsBytes [][]byte
	for i := 1; i <= dataRecordCnt; i++ {
		dataRecord := pfutils.DirectDebitFileDataRecord{
			DataCategory:   9,
			CustomerNumber: fmt.Sprintf("%v", i),
			DepositAmount:  amount,
			ResultCode:     resultCode,
		}

		dataRecordBytesData, _ := fixedwidth.Marshal(dataRecord)

		dataRecordsBytes = append(dataRecordsBytes, dataRecordBytesData)
	}

	trailerRecordBytes := generateDirectDebitTrailerRecord()

	endRecordBytes := generateDirectDebitEndRecord()

	fileContentsStr := fmt.Sprintf("%v\n", string(headerRecordBytes))

	for _, dataBytes := range dataRecordsBytes {
		fileContentsStr += fmt.Sprintf("%v\n", string(dataBytes))
	}

	fileContentsStr += fmt.Sprintf("%v\n%v", string(trailerRecordBytes), string(endRecordBytes))

	return []byte(fileContentsStr)
}

func createInvalidEndRecordCategoryForDirectDebit(dataRecordCnt int, resultCode string, amount int) []byte {
	headerRecordBytes := generateDirectDebitHeader()

	dataRecordsBytes := generateDirectDebitData(dataRecordCnt, resultCode, amount)

	trailerRecordBytes := generateDirectDebitTrailerRecord()

	endRecord := pfutils.DirectDebitFileEndRecord{
		DataCategory: 4,
	}

	endRecordBytes, _ := fixedwidth.Marshal(endRecord)

	fileContentsStr := fmt.Sprintf("%v\n", string(headerRecordBytes))

	for _, dataBytes := range dataRecordsBytes {
		fileContentsStr += fmt.Sprintf("%v\n", string(dataBytes))
	}

	fileContentsStr += fmt.Sprintf("%v\n%v", string(trailerRecordBytes), string(endRecordBytes))

	return []byte(fileContentsStr)
}

func generateDirectDebitHeader() []byte {
	headerRecord := pfutils.DirectDebitFileHeaderRecord{
		DataCategory: pfutils.DataTypeHeaderRecord,
	}
	headerRecordBytes, _ := fixedwidth.Marshal(headerRecord)

	return headerRecordBytes
}

func generateDirectDebitData(dataRecordCnt int, resultCode string, amount int) [][]byte {
	var dataRecordsBytes [][]byte
	for i := 1; i <= dataRecordCnt; i++ {
		dataRecord := pfutils.DirectDebitFileDataRecord{
			DataCategory:   pfutils.DataTypeDataRecord,
			CustomerNumber: fmt.Sprintf("%v", i),
			DepositAmount:  amount,
			ResultCode:     resultCode,
		}

		dataRecordBytesData, _ := fixedwidth.Marshal(dataRecord)

		dataRecordsBytes = append(dataRecordsBytes, dataRecordBytesData)
	}

	return dataRecordsBytes
}

func generateDirectDebitTrailerRecord() []byte {
	trailerRecord := pfutils.DirectDebitFileTrailerRecord{
		DataCategory:      pfutils.DataTypeTrailerRecord,
		TransferredAmount: 1000,
		TransferredNumber: 1,
	}

	trailerRecordBytes, _ := fixedwidth.Marshal(trailerRecord)

	return trailerRecordBytes
}

func generateDirectDebitEndRecord() []byte {
	endRecord := pfutils.DirectDebitFileEndRecord{
		DataCategory: pfutils.DataTypeEndRecord,
	}

	endRecordBytes, _ := fixedwidth.Marshal(endRecord)

	return endRecordBytes
}
