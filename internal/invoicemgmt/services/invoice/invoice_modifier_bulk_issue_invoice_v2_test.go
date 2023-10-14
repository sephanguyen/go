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
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_BulkIssueInvoiceV2(t *testing.T) {
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

	s := &InvoiceModifierService{
		DB:                       mockDB,
		InvoiceRepo:              mockInvoiceRepo,
		PaymentRepo:              mockPaymentRepo,
		StudentPaymentDetailRepo: mockStudentPaymentDetailRepo,
		InvoiceActionLogRepo:     mockInvoiceActionLogRepo,
		BulkPaymentRepo:          mockBulkAddPaymentRepo,
		UnleashClient:            mockUnleashClient,
	}

	successfulResp := &invoice_pb.BulkIssueInvoiceResponseV2{
		Success: true,
	}

	negativeTotalDraftInvoice := &entities.Invoice{
		InvoiceID: database.Text("1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(-100),
	}

	zeroTotalDraftInvoice := &entities.Invoice{
		InvoiceID: database.Text("1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(0),
	}

	mockInvoices := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-failed-invoice-1"),
			Total:     database.Numeric(100),
		},
		{
			InvoiceID: database.Text("2"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-invoice-2"),
			Total:     database.Numeric(100),
		},
		{
			InvoiceID: database.Text("3"),
			Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-invoice-3"),
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

	mockPaidInvoice := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_PAID.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-failed-invoice-1"),
			Total:     database.Numeric(100),
		},
	}

	mockRefundedInvoice := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_REFUNDED.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-failed-invoice-1"),
			Total:     database.Numeric(100),
		},
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

	createdDDPayment := &entities.Payment{
		InvoiceID:             database.Text("2"),
		PaymentID:             database.Text("payment-id-2"),
		PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	studentWithEmptyPaymentMethodSet := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              database.Text("1"),
		PaymentMethod:          database.Text(""),
	}

	studentWithNoPaymentMethodSet := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              database.Text("1"),
	}

	testError := errors.New("test error")

	testcases := []TestCase{
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

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)

				for i := 0; i < 3; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockInvoices[i], nil)
					mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment[i], nil)
					mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - default payment convenience store",
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
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockInvoices[0], nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodCS, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdCSPayment[0], nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - default payment direct debit",
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
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockInvoices[0], nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithStudentPaymentMethodDD, nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(createdDDPayment, nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - empty invoice IDs",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds: []string{},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invoice ids cannot be empty"),
			setup: func(ctx context.Context) {

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

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				for i := 0; i < 1; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockIssuedInvoice[i], nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
				}
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

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				for i := 0; i < 1; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockVoidInvoice[i], nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
				}
			},
		},
		{
			name: "negative case - invoice have PAID status",
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
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("error invalid invoice status: %v", invoice_pb.InvoiceStatus_PAID.String())),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				for i := 0; i < 1; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockPaidInvoice[i], nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
				}
			},
		},
		{
			name: "negative case - invoice have REFUNDED status",
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
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("error invalid invoice status: %v", invoice_pb.InvoiceStatus_REFUNDED.String())),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				for i := 0; i < 1; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockRefundedInvoice[i], nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
				}
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

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				for i := 0; i < 1; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(negativeTotalDraftInvoice, nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
				}
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

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				for i := 0; i < 1; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(zeroTotalDraftInvoice, nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
				}
			},
		},
		{
			name: "negative case - InvoiceRepo.RetrieveInvoiceByInvoiceID failed",
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
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("error InvoiceRepo RetrieveInvoiceByInvoiceID: %v", testError)),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				for i := 0; i < 1; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
					mockTx.On("Rollback", ctx).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test - no convenience store dates",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				InvoiceType:            []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "convenience store dates cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no convenience store due date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no convenience store expiry date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no direct debit store due date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no direct debit store expiry date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - no direct debit store dates",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "direct debit dates cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - direct debit due date < now",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - direct debit expiry date < now",
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
					ExpiryDate: timestamppb.New(time.Now().Add(-1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - convenience store expiry date < now",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(-1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - convenience store due date < now",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - convenience store expiry date < due date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(2 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - direct debit expiry date < due date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequestV2{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(2 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				InvoiceType: []invoice_pb.InvoiceType{invoice_pb.InvoiceType_SCHEDULED},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup: func(ctx context.Context) {
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
			expectedErr: status.Error(codes.Internal, "bulk issue student: student-failed-invoice-1 payment method in student payment detail is empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockInvoices[0], nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithEmptyPaymentMethodSet, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative case - default payment no payment method set",
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
			expectedErr: status.Error(codes.Internal, "bulk issue student: student-failed-invoice-1 payment method in student payment detail is empty"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkAddPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkIssueInvoice, mock.Anything).Once().Return(false, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(mockInvoices[0], nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(studentWithNoPaymentMethodSet, nil)
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
