package invoicesvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_BulkIssueInvoice(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDb := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)

	s := &InvoiceModifierService{
		DB:                       mockDb,
		InvoiceRepo:              mockInvoiceRepo,
		PaymentRepo:              mockPaymentRepo,
		StudentPaymentDetailRepo: mockStudentPaymentDetailRepo,
		InvoiceActionLogRepo:     mockInvoiceActionLogRepo,
	}

	failedResp := &invoice_pb.BulkIssueInvoiceResponse{
		Success: false,
	}

	successfulResp := &invoice_pb.BulkIssueInvoiceResponse{
		Success: true,
	}

	issuedInvoice := &entities.Invoice{
		InvoiceID: database.Text("1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(100),
	}

	negativeTotalInvoice := &entities.Invoice{
		InvoiceID: database.Text("1"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Total:     database.Numeric(-100),
	}

	//invoices default payment convenience store
	defaultConvenienceStoreInvoices := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
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
			Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-invoice-3"),
			Total:     database.Numeric(100),
		},
	}

	defaultConvenienceStorePayments := []*entities.Payment{
		{
			PaymentID:             database.Text("2"),
			InvoiceID:             defaultConvenienceStoreInvoices[0].InvoiceID,
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			PaymentID:             database.Text("2"),
			InvoiceID:             defaultConvenienceStoreInvoices[1].InvoiceID,
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			PaymentID:             database.Text("3"),
			InvoiceID:             defaultConvenienceStoreInvoices[2].InvoiceID,
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
	}
	defaultConvenienceStoreStudentDetail := []*entities.StudentPaymentDetail{
		{
			StudentPaymentDetailID: database.Text("1"),
			StudentID:              defaultConvenienceStoreInvoices[0].StudentID,
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			StudentPaymentDetailID: database.Text("2"),
			StudentID:              defaultConvenienceStoreInvoices[1].StudentID,
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
		{
			StudentPaymentDetailID: database.Text("3"),
			StudentID:              defaultConvenienceStoreInvoices[2].StudentID,
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		},
	}

	//invoices default payment direct debit
	defaultDirectDebitInvoices := []*entities.Invoice{
		{
			InvoiceID: database.Text("1"),
			Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
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
			Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			StudentID: database.Text("student-invoice-3"),
			Total:     database.Numeric(100),
		},
	}

	defaultDirectDebitPayments := []*entities.Payment{
		{
			PaymentID:             database.Text("2"),
			InvoiceID:             defaultDirectDebitInvoices[0].InvoiceID,
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
		{
			PaymentID:             database.Text("2"),
			InvoiceID:             defaultDirectDebitInvoices[1].InvoiceID,
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
		{
			PaymentID:             database.Text("3"),
			InvoiceID:             defaultDirectDebitInvoices[2].InvoiceID,
			PaymentSequenceNumber: database.Int4(1),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
	}
	defaultDirectDebitStudentDetail := []*entities.StudentPaymentDetail{
		{
			StudentPaymentDetailID: database.Text("1"),
			StudentID:              defaultDirectDebitInvoices[0].StudentID,
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
		{
			StudentPaymentDetailID: database.Text("2"),
			StudentID:              defaultDirectDebitInvoices[1].StudentID,
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
		{
			StudentPaymentDetailID: database.Text("3"),
			StudentID:              defaultDirectDebitInvoices[2].StudentID,
			PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		},
	}

	testcases := []TestCase{
		{
			name: "happy case - DEFAULT_PAYMENT for CONVENIENCE STORE invoice draft/failed records",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < 3; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(defaultConvenienceStoreInvoices[i], nil)
					mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(defaultConvenienceStoreStudentDetail[i], nil)
					mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(defaultConvenienceStorePayments[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", ctx).Once().Return(nil)
				}
			},
		},
		{
			name: "happy case - DEFAULT_PAYMENT for DIRECT DEBIT invoice draft/failed records",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"1", "2", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				for i := 0; i < 3; i++ {
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(defaultDirectDebitInvoices[i], nil)
					mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockTx, mock.Anything).Once().Return(defaultDirectDebitStudentDetail[i], nil)
					mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(defaultDirectDebitPayments[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", ctx).Once().Return(nil)
				}
			},
		},
		{
			name: "issue invoice failed - invalid payment method",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"9", "3", "1"},
				BulkIssuePaymentMethod: 10,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid 10 BulkIssuePaymentMethod value"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - invalid invoice status",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{issuedInvoice.InvoiceID.String},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "error Wrong Invoice Status"),
			setup: func(ctx context.Context) {
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(issuedInvoice, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "issue invoice failed - invalid convenience store due date empty",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    nil,
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - invalid convenience store expiry date empty",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"3", "2", "1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: nil,
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup:        func(ctx context.Context) {},
		},

		{
			name: "issue invoice failed - invalid direct debit due date empty",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    nil,
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid DueDate value"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - invalid direct debit expiry date empty",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"3", "2", "1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: nil,
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid ExpiryDate value"),
			setup:        func(ctx context.Context) {},
		},

		{
			name: "issue invoice failed - due date < time now",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - direct debit due date < time now",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - expiry date < time now",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(-1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - direct debit expiry date < time now",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(-1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - expiry date < due date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(2 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - direct debit expiry date < due date",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"5", "4", "3"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(2 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup:        func(ctx context.Context) {},
		},

		{
			name: "issue invoice failed - invalid empty direct debit dates on default payment method",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"9", "3", "1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "direct debit dates cannot be empty"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - invalid  empty convenience store dates on convenience payment method",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"9", "3", "1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "convenience store dates cannot be empty"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - invalid  empty convenience store dates on default payment method",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{"9", "3", "1"},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT,
				DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "convenience store dates cannot be empty"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "issue invoice failed - negative total invoice",
			ctx:  ctx,
			req: &invoice_pb.BulkIssueInvoiceRequest{
				InvoiceIds:             []string{negativeTotalInvoice.InvoiceID.String},
				BulkIssuePaymentMethod: invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE,
				ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
					DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
					ExpiryDate: timestamppb.New(time.Now().Add(1 * time.Hour)),
				},
			},
			expectedResp: failedResp,
			expectedErr:  status.Error(codes.InvalidArgument, "error Should have positive total, negative total found"),
			setup: func(ctx context.Context) {
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(negativeTotalInvoice, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.BulkIssueInvoice(testCase.ctx, testCase.req.(*invoice_pb.BulkIssueInvoiceRequest))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t, mockDb, mockInvoiceRepo, mockPaymentRepo, mockStudentPaymentDetailRepo, mockInvoiceActionLogRepo)
		})
	}
}
