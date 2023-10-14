package paymentsvc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_sequence_number "github.com/manabie-com/backend/mock/invoicemgmt/services/sequence_number"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPaymentModifierService_BulkkAddPayment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const (
		ctxUserID = "user-id"
	)

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)

	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkAddPaymentRepo := new(mock_repositories.MockBulkPaymentRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockSeqNumberService := new(mock_sequence_number.ISequenceNumberService)

	s := &PaymentModifierService{
		DB:                       mockDB,
		InvoiceRepo:              mockInvoiceRepo,
		PaymentRepo:              mockPaymentRepo,
		StudentPaymentDetailRepo: mockStudentPaymentDetailRepo,
		InvoiceActionLogRepo:     mockInvoiceActionLogRepo,
		BulkPaymentRepo:          mockBulkAddPaymentRepo,
		SequenceNumberService:    mockSeqNumberService,
		UnleashClient:            mockUnleashClient,
	}

	concretePaymentSeqNumberService := &seqnumberservice.PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	testInvoice := &entities.Invoice{
		InvoiceID:          database.Text("invoice-test-id-1"),
		Status:             database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		OutstandingBalance: database.Numeric(10),
		Total:              database.Numeric(10),
		StudentID:          database.Text("test-student-id"),
	}

	studentWithStudentPaymentMethodCS := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              database.Text("1"),
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}
	studentWithStudentPaymentMethodDD := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("1234"),
		StudentID:              database.Text("12"),
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	createdCSPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}

	createdDDPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	studentWithNoPaymentMethod := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              database.Text("1"),
		PaymentMethod:          database.Text(""),
	}

	testError := errors.New("testError")

	testcases := []TestCase{
		{
			name: "happy case - default payment convenience store",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: &invoice_pb.BulkAddPaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodCS, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - convenience store payment method",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: &invoice_pb.BulkAddPaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - default payment direct debit",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: &invoice_pb.BulkAddPaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodDD, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdDDPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - no invoice ids",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice ids cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - invalid payment method",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: 5,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid 5 BulkPaymentMethod value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no convenience store dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: 1,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "convenience store dates cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no due date convenience store dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    nil,
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no expiry date convenience store dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: nil,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - direct debit dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "direct debit dates cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty due date direct debit dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    nil,
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty expiry date direct debit dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: nil,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - due date < now convenience store dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - expiry date < now convenience store dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(-1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - expiry date < due date convenience store dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(2 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - due date direct < now debit dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - expiry date direct < now debit dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(-1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - expiry date direct < due date debit dates",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(2 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty latest payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod: invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "latest payment status cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - invalid pending latest payment status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL, invoice_pb.PaymentStatus_PAYMENT_REFUNDED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "latest payment status value should only have no payment or failed"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty invoice type",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice type cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - invalid invoice type",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{5},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invoice type value should only have manual and scheduled"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - error create on bulk payment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "error BulkPaymentRepo Create: testError"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - error retrieving invoice",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "error InvoiceRepo RetrieveInvoiceByInvoiceID: testError"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - error finding student payment detail",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "error StudentPaymentDetailRepo FindByStudentID: testError"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - invoice status is not issued",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "error invalid invoice status: DRAFT"),
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_DRAFT.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - error create payment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "error PaymentRepo Create: testError"),
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodCS, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - error get latest payment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "error PaymentRepo GetLatestPaymentDueDateByInvoiceID: testError"),
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodCS, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - error action log create",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "error InvoiceActionLogRepo Create: testError"),
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodCS, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - student empty default payment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.Internal, "bulk add student: test-student-id payment method in student payment detail is empty"),
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithNoPaymentMethod, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.BulkAddPayment(testCase.ctx, testCase.req.(*invoice_pb.BulkAddPaymentRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockTx,
				mockInvoiceRepo,
				mockPaymentRepo,
				mockStudentPaymentDetailRepo,
				mockBulkAddPaymentRepo,
			)
		})
	}
}

func TestPaymentModifierService_BulkkAddPayment_ManualSetOfPaymentSeqNumber(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const (
		ctxUserID = "user-id"
	)

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)

	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkAddPaymentRepo := new(mock_repositories.MockBulkPaymentRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockSeqNumberService := new(mock_sequence_number.ISequenceNumberService)

	s := &PaymentModifierService{
		DB:                       mockDB,
		InvoiceRepo:              mockInvoiceRepo,
		PaymentRepo:              mockPaymentRepo,
		StudentPaymentDetailRepo: mockStudentPaymentDetailRepo,
		InvoiceActionLogRepo:     mockInvoiceActionLogRepo,
		BulkPaymentRepo:          mockBulkAddPaymentRepo,
		SequenceNumberService:    mockSeqNumberService,
		UnleashClient:            mockUnleashClient,
	}

	concretePaymentSeqNumberService := &seqnumberservice.PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	testInvoice := &entities.Invoice{
		InvoiceID:          database.Text("invoice-test-id-1"),
		Status:             database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		OutstandingBalance: database.Numeric(10),
		Total:              database.Numeric(10),
		StudentID:          database.Text("test-student-id"),
	}

	studentWithStudentPaymentMethodCS := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              database.Text("1"),
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}
	studentWithStudentPaymentMethodDD := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("1234"),
		StudentID:              database.Text("12"),
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	createdCSPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}

	createdDDPayment := &entities.Payment{
		InvoiceID:             database.Text("invoice-test-id-1"),
		PaymentID:             database.Text("payment-id-1"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	testcases := []TestCase{
		{
			name: "happy case - default payment convenience store",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: &invoice_pb.BulkAddPaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodCS, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - convenience store payment method",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: &invoice_pb.BulkAddPaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - default payment direct debit",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.BulkAddPaymentRequest{
				InvoiceIds: []string{"1"},
				BulkAddPaymentDetails: &invoice_pb.BulkAddPaymentRequest_BulkAddPaymentDetails{
					BulkPaymentMethod:   invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT,
					LatestPaymentStatus: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_NONE, invoice_pb.PaymentStatus_PAYMENT_FAILED},
					InvoiceType:         []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
				},
				ConvenienceStoreDates: &invoice_pb.BulkAddConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkAddDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: &invoice_pb.BulkAddPaymentResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				testInvoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(testInvoice, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodDD, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdDDPayment, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.BulkAddPayment(testCase.ctx, testCase.req.(*invoice_pb.BulkAddPaymentRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockTx,
				mockInvoiceRepo,
				mockPaymentRepo,
				mockStudentPaymentDetailRepo,
				mockBulkAddPaymentRepo,
			)
		})
	}
}
