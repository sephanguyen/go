package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_sequence_number "github.com/manabie-com/backend/mock/invoicemgmt/services/sequence_number"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_BulkIssueInvoiceV2_Improved(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkAddPaymentRepo := new(mock_repositories.MockBulkPaymentRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockSeqNumberService := new(mock_sequence_number.ISequenceNumberService)

	s := &InvoiceModifierService{
		DB:                       mockDB,
		InvoiceRepo:              mockInvoiceRepo,
		PaymentRepo:              mockPaymentRepo,
		StudentPaymentDetailRepo: mockStudentPaymentDetailRepo,
		InvoiceActionLogRepo:     mockInvoiceActionLogRepo,
		BulkPaymentRepo:          mockBulkAddPaymentRepo,
		UnleashClient:            mockUnleashClient,
		SequenceNumberService:    mockSeqNumberService,
	}

	concretePaymentSeqNumberService := &seqnumberservice.PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	successfulResp := &invoice_pb.BulkIssueInvoiceResponseV2{
		Success: true,
	}

	negativeTotalDraftInvoice := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			Total:     database.Numeric(-100),
		},
	}

	zeroTotalDraftInvoice := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			Total:     database.Numeric(0),
		},
	}

	mockInvoices := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("1"),
			Total:     database.Numeric(100),
		},
		{
			InvoiceID: database.Text("2"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("2"),
			Total:     database.Numeric(100),
		},
		{
			InvoiceID: database.Text("3"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("3"),
			Total:     database.Numeric(100),
		},
	}

	createdCSPayment := []*entities.Payment{
		{
			InvoiceID:             database.Text("1"),
			PaymentID:             database.Text("payment-id-1"),
			PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			InvoiceID:             database.Text("2"),
			PaymentID:             database.Text("payment-id-2"),
			PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			InvoiceID:             database.Text("3"),
			PaymentID:             database.Text("payment-id-3"),
			PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
	}

	mockStudentCS := []*entities.StudentPaymentDetail{
		{
			StudentPaymentDetailID: database.Text("1"),
			StudentID:              database.Text("1"),
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			StudentPaymentDetailID: database.Text("2"),
			StudentID:              database.Text("2"),
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			StudentPaymentDetailID: database.Text("3"),
			StudentID:              database.Text("3"),
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
	}

	mockStudentDD := []*entities.StudentPaymentDetail{
		{
			StudentPaymentDetailID: database.Text("1"),
			StudentID:              database.Text("1"),
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
		{
			StudentPaymentDetailID: database.Text("2"),
			StudentID:              database.Text("2"),
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
		{
			StudentPaymentDetailID: database.Text("3"),
			StudentID:              database.Text("3"),
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
	}

	mockIssuedInvoice := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-failed-invoice-1"),
			Total:     database.Numeric(100),
		},
	}

	mockVoidInvoice := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-failed-invoice-1"),
			Total:     database.Numeric(100),
		},
	}

	studentWithEmptyPaymentMethodSet := []*entities.StudentPaymentDetail{
		{
			StudentPaymentDetailID: database.Text("123"),
			StudentID:              database.Text("1"),
			PaymentMethod:          database.Text(""),
		},
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name: "happy case auto set payment sequence number - convenience store",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("FindByPaymentIDs", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment, nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - created payments are not equal to the number of invoice",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "there are 2 payments that were not created"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("FindByPaymentIDs", ctx, mockTx, mock.Anything).Once().Return([]*entities.Payment{createdCSPayment[1]}, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - convenience store",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - default payment convenience store",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockStudentCS, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - default payment direct debit",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockStudentDD, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - invoice have ISSUED status",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("error invalid invoice status: %v", invoice_pb.InvoiceStatus_ISSUED.String())),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockIssuedInvoice, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - invoice have VOID status",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("error invalid invoice status: %v", invoice_pb.InvoiceStatus_VOID.String())),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockVoidInvoice, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - there is missing invoice",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3", "4"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "there are 1 invoices that does not exist"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - there is a negative invoice amount",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("error Should have positive total, negative total found")),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(negativeTotalDraftInvoice, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - there is a zero invoice amount",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("error Should have positive total, zero total amount found")),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(zeroTotalDraftInvoice, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - default payment empty payment method set",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.Internal, "bulk issue student: 1 payment method in student payment detail is empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return([]*entities.Invoice{mockInvoices[0]}, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return(studentWithEmptyPaymentMethodSet, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - There are missing student payment detail",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "there are 2 students that does not have student payment detail"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return([]*entities.StudentPaymentDetail{mockStudentCS[1]}, nil)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - InvoiceRepo.InsertInvoiceIDsTempTable error",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.InvoiceRepo.InsertInvoiceIDsTempTable err: %v", testError)),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - InvoiceRepo.FindInvoicesFromInvoiceIDTempTable error",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.InvoiceRepo.FindInvoicesFromInvoiceIDTempTable err: %v", testError)),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(nil, testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - StudentPaymentDetailRepo.FindFromInvoiceIDTempTable",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.StudentPaymentDetailRepo.FindFromInvoiceIDTempTable err: %v", testError)),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return(nil, testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - InvoiceRepo.UpdateStatusFromInvoiceIDTempTable",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.InvoiceRepo.UpdateStatusFromInvoiceIDTempTable err: %v", testError)),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockStudentCS, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - PaymentRepo.CreateMultiple",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.PaymentRepo.CreateMultiple err: %v", testError)),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockStudentCS, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - InvoiceActionLogRepo.CreateMultiple",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.InvoiceActionLogRepo.CreateMultiple err: %v", testError)),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("InsertInvoiceIDsTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceRepo.On("FindInvoicesFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockInvoices, nil)
				mockStudentPaymentDetailRepo.On("FindFromInvoiceIDTempTable", ctx, mockTx).Once().Return(mockStudentCS, nil)
				mockInvoiceRepo.On("UpdateStatusFromInvoiceIDTempTable", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Once().Return(testError)

				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.BulkIssueInvoiceV2(testCase.ctx, testCase.req.(*invoice_pb.BulkIssueInvoiceRequestV2))

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
				mockUnleashClient,
			)
		})
	}
}
