package invoicesvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInvoiceModifierService_RetrieveInvoiceRecords(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)

	s := &InvoiceModifierService{
		DB:          mockDB,
		InvoiceRepo: mockInvoiceRepo,
		PaymentRepo: mockPaymentRepo,
	}

	const (
		txClosedError     = "tx is closed"
		dbClosedPoolError = "closed pool"
	)

	// no student invoice test scenario
	studentIDWithNoInvoice := database.Text("student-no-invoice")
	emptyInvoiceRecord := []*entities.Invoice{}

	// student with single invoice no payment
	studentWithSingleInvoiceNoPayment := database.Text("student-01")
	singleStudentInvoiceNoPayment := []*entities.Invoice{
		{
			InvoiceID: database.Text("12"),
			Status:    database.Text("ISSUED"),
			StudentID: studentWithSingleInvoiceNoPayment,
		},
	}
	singleStudentInvoiceNoPayment[0].Total.Set(123333.57)

	// student with single invoice with single payment test scenario
	studentWithSingleInvoiceAndPayment := database.Text("student-02")
	singleStudentInvoiceSinglePayment := []*entities.Invoice{
		{
			InvoiceID: database.Text("13"),
			Status:    database.Text("ISSUED"),
			StudentID: studentWithSingleInvoiceAndPayment,
		},
	}
	singleStudentInvoiceSinglePayment[0].Total.Set(155.00)

	singlePayment := &entities.Payment{
		InvoiceID:      singleStudentInvoiceSinglePayment[0].InvoiceID,
		PaymentDueDate: pgtype.Timestamptz{Time: time.Now()},
		CreatedAt:      pgtype.Timestamptz{Time: time.Now()},
		PaymentStatus:  database.Text("PAYMENT_PENDING"),
	}

	// student with single invoice with multiple payment test scenario
	studentWithSingleInvoiceAndMultiplePayment := database.Text("student-03")
	singleStudentInvoiceMultiplePayment := []*entities.Invoice{
		{
			InvoiceID: database.Text("14"),
			Status:    database.Text("PAID"),
			StudentID: studentWithSingleInvoiceAndMultiplePayment,
		},
	}
	singleStudentInvoiceMultiplePayment[0].Total.Set(55440.58)

	singleInvoicemultiplePayment := []*entities.Payment{
		{
			InvoiceID:      singleStudentInvoiceMultiplePayment[0].InvoiceID,
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			PaymentStatus:  database.Text("PAYMENT_PENDING"),
		},
		{
			InvoiceID:      singleStudentInvoiceMultiplePayment[0].InvoiceID,
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now()},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now()},
			PaymentStatus:  database.Text("PAYMENT_SUCCESSFUL"),
		},
	}

	// student with multiple invoice with no payment test scenario
	studentWithMultipleInvoice := database.Text("student-04")
	multipleStudentInvoice := []*entities.Invoice{
		{
			InvoiceID: database.Text("15"),
			Status:    database.Text("ISSUED"),
			StudentID: studentWithMultipleInvoice,
		},
		{
			InvoiceID: database.Text("16"),
			Status:    database.Text("ISSUED"),
			StudentID: studentWithMultipleInvoice,
		},
	}
	multipleStudentInvoice[0].Total.Set(12000036.87)
	multipleStudentInvoice[1].Total.Set(1500.55)

	// student with multiple invoice with single payment on an invoice test scenario
	studentWithMultipleInvoiceAndPayment := database.Text("student-05")
	multipleStudentInvoiceSinglePayment := []*entities.Invoice{
		{
			InvoiceID: database.Text("17"),
			Status:    database.Text("PAID"),
			StudentID: studentWithMultipleInvoiceAndPayment,
		},
		{
			InvoiceID: database.Text("18"),
			Status:    database.Text("ISSUED"),
			StudentID: studentWithMultipleInvoiceAndPayment,
		},
	}
	multipleStudentInvoiceSinglePayment[0].Total.Set(2500.32)
	multipleStudentInvoiceSinglePayment[1].Total.Set(1000.28)

	multipleInvoiceSinglePayment := &entities.Payment{
		InvoiceID:      multipleStudentInvoiceSinglePayment[0].InvoiceID,
		PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		PaymentStatus:  database.Text("PAYMENT_FAILED"),
	}

	// student with multiple invoice with multiple payment and single payment test scenario
	studentWithMultipleInvoiceAndMultiplePayment := database.Text("student-06")
	multipleStudentInvoiceMultiplePayment := []*entities.Invoice{
		{
			InvoiceID: database.Text("19"),
			Status:    database.Text("INVOICED"),
			StudentID: studentWithMultipleInvoiceAndMultiplePayment,
		},
		{
			InvoiceID: database.Text("20"),
			Status:    database.Text("ISSUED"),
			StudentID: studentWithMultipleInvoiceAndMultiplePayment,
		},
	}
	multipleStudentInvoiceMultiplePayment[0].Total.Set(2222.99)
	multipleStudentInvoiceMultiplePayment[1].Total.Set(77775.33)

	multipleInvoiceMultiplePayment := []*entities.Payment{
		{
			InvoiceID: multipleStudentInvoiceMultiplePayment[0].InvoiceID,
			// latest payment
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now()},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now()},
			PaymentStatus:  database.Text("PAYMENT_SUCCESSFUL"),
		},
		{
			InvoiceID:      multipleStudentInvoiceMultiplePayment[0].InvoiceID,
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			PaymentStatus:  database.Text("PAYMENT_FAILED"),
		},
		{
			InvoiceID:      multipleStudentInvoiceMultiplePayment[1].InvoiceID,
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour)},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour)},
			PaymentStatus:  database.Text("PAYMENT_SUCCESSFUL"),
		},
	}

	// student with multiple invoice with multiple payments each invoice test scenario
	studentWithMultipleInvoiceWithMultiplePaymentEach := database.Text("student-07")
	multipleStudentInvoiceWithMultiplePaymentEach := []*entities.Invoice{
		{
			InvoiceID: database.Text("21"),
			Status:    database.Text("VOID"),
			StudentID: studentWithMultipleInvoiceWithMultiplePaymentEach,
		},
		{
			InvoiceID: database.Text("22"),
			Status:    database.Text("PAID"),
			StudentID: studentWithMultipleInvoiceWithMultiplePaymentEach,
		},
	}
	multipleStudentInvoiceWithMultiplePaymentEach[0].Total.Set(55122.99)
	multipleStudentInvoiceWithMultiplePaymentEach[1].Total.Set(3333.33)

	multipleInvoiceMultiplePaymentEach := []*entities.Payment{
		{
			InvoiceID: multipleStudentInvoiceMultiplePayment[0].InvoiceID,
			// latest payment for first invoice
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now()},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now()},
			PaymentStatus:  database.Text("PAYMENT_FAILED"),
		},
		{
			InvoiceID:      multipleStudentInvoiceMultiplePayment[0].InvoiceID,
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
			PaymentStatus:  database.Text("PAYMENT_PENDING"),
		},
		{
			InvoiceID:      multipleStudentInvoiceMultiplePayment[1].InvoiceID,
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-3 * time.Hour)},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-3 * time.Hour)},
			PaymentStatus:  database.Text("PAYMENT_FAILED"),
		},
		{
			InvoiceID: multipleStudentInvoiceMultiplePayment[1].InvoiceID,
			// latest payment for second invoice
			PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour)},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour)},
			PaymentStatus:  database.Text("PAYMENT_SUCCESSFUL"),
		},
	}

	testcases := []TestCase{
		{
			name:        "no invoice record for a student",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentIDWithNoInvoice.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(emptyInvoiceRecord, nil)
			},
		},
		{
			name:        "happy case retrieve single invoice record with no payment",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithSingleInvoiceNoPayment.String,
			},
			mockInvoiceEntities: singleStudentInvoiceNoPayment,
			expectedResp: &invoice_pb.RetrieveInvoiceRecordsResponse{
				InvoiceRecords: []*invoice_pb.InvoiceRecord{
					{
						InvoiceIdString: singleStudentInvoiceNoPayment[0].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(singleStudentInvoiceNoPayment[0].InvoiceID.Status),
						Total:           123333.57,
						DueDate:         timestamppb.New(pgtype.Timestamptz{Status: pgtype.Null}.Time),
					},
				},
				NextPage: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: constant.PageLimit + 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(singleStudentInvoiceNoPayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "happy case retrieve single invoice record with single payment",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithSingleInvoiceAndPayment.String,
			},
			mockInvoiceEntities: singleStudentInvoiceSinglePayment,
			expectedResp: &invoice_pb.RetrieveInvoiceRecordsResponse{
				InvoiceRecords: []*invoice_pb.InvoiceRecord{
					{
						InvoiceIdString: singleStudentInvoiceSinglePayment[0].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(singleStudentInvoiceSinglePayment[0].InvoiceID.Status),
						Total:           155.00,
						DueDate:         timestamppb.New(singlePayment.PaymentDueDate.Time),
					},
				},
				NextPage: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: constant.PageLimit + 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(singleStudentInvoiceSinglePayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(singlePayment, nil)
			},
		},
		{
			name:        "happy case retrieve single invoice record with multiple payment get latest due date",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithSingleInvoiceAndMultiplePayment.String,
			},
			mockInvoiceEntities: singleStudentInvoiceMultiplePayment,
			expectedResp: &invoice_pb.RetrieveInvoiceRecordsResponse{
				InvoiceRecords: []*invoice_pb.InvoiceRecord{
					{
						InvoiceIdString: singleStudentInvoiceMultiplePayment[0].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(singleStudentInvoiceMultiplePayment[0].InvoiceID.Status),
						Total:           55440.58,
						DueDate:         timestamppb.New(singleInvoicemultiplePayment[1].PaymentDueDate.Time),
					},
				},
				NextPage: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: constant.PageLimit + 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(singleStudentInvoiceMultiplePayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(singleInvoicemultiplePayment[1], nil)
			},
		},
		{
			name:        "happy case retrieve multiple invoices record with no payment",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			mockInvoiceEntities: multipleStudentInvoice,
			expectedResp: &invoice_pb.RetrieveInvoiceRecordsResponse{
				InvoiceRecords: []*invoice_pb.InvoiceRecord{
					{
						InvoiceIdString: multipleStudentInvoice[0].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoice[0].InvoiceID.Status),
						Total:           12000036.87,
						DueDate:         timestamppb.New(pgtype.Timestamptz{Status: pgtype.Null}.Time),
					},
					{
						InvoiceIdString: multipleStudentInvoice[1].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoice[1].InvoiceID.Status),
						Total:           1500.55,
						DueDate:         timestamppb.New(pgtype.Timestamptz{Status: pgtype.Null}.Time),
					},
				},
				NextPage: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: constant.PageLimit + 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(multipleStudentInvoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "happy case retrieve multiple invoices record with single payment",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			mockInvoiceEntities: multipleStudentInvoiceSinglePayment,
			expectedResp: &invoice_pb.RetrieveInvoiceRecordsResponse{
				InvoiceRecords: []*invoice_pb.InvoiceRecord{
					{
						InvoiceIdString: multipleStudentInvoiceSinglePayment[0].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoiceSinglePayment[0].InvoiceID.Status),
						Total:           2500.32,
						DueDate:         timestamppb.New(multipleInvoiceSinglePayment.PaymentDueDate.Time),
					},
					{
						InvoiceIdString: multipleStudentInvoiceSinglePayment[1].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoiceSinglePayment[1].InvoiceID.Status),
						Total:           1000.28,
						DueDate:         timestamppb.New(pgtype.Timestamptz{Status: pgtype.Null}.Time),
					},
				},
				NextPage: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: constant.PageLimit + 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(multipleStudentInvoiceSinglePayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(multipleInvoiceSinglePayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "happy case multiple payments and single payment on multiple invoices",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			mockInvoiceEntities: multipleStudentInvoiceMultiplePayment,
			expectedResp: &invoice_pb.RetrieveInvoiceRecordsResponse{
				InvoiceRecords: []*invoice_pb.InvoiceRecord{
					{
						InvoiceIdString: multipleStudentInvoiceMultiplePayment[0].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoiceMultiplePayment[0].InvoiceID.Status),
						Total:           2222.99,
						DueDate:         timestamppb.New(multipleInvoiceMultiplePayment[0].PaymentDueDate.Time),
					},
					{
						InvoiceIdString: multipleStudentInvoiceMultiplePayment[1].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoiceMultiplePayment[1].InvoiceID.Status),
						Total:           77775.33,
						DueDate:         timestamppb.New(multipleInvoiceMultiplePayment[2].PaymentDueDate.Time),
					},
				},
				NextPage: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: constant.PageLimit + 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(multipleStudentInvoiceMultiplePayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(multipleInvoiceMultiplePayment[0], nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(multipleInvoiceMultiplePayment[2], nil)
			},
		},
		{
			name:        "happy case multiple payments on each multiple invoices",
			ctx:         ctx,
			expectedErr: nil,
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			mockInvoiceEntities: multipleStudentInvoiceWithMultiplePaymentEach,
			expectedResp: &invoice_pb.RetrieveInvoiceRecordsResponse{
				InvoiceRecords: []*invoice_pb.InvoiceRecord{
					{
						InvoiceIdString: multipleStudentInvoiceWithMultiplePaymentEach[0].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoiceWithMultiplePaymentEach[0].InvoiceID.Status),
						Total:           55122.99,
						DueDate:         timestamppb.New(multipleInvoiceMultiplePaymentEach[0].PaymentDueDate.Time),
					},
					{
						InvoiceIdString: multipleStudentInvoiceWithMultiplePaymentEach[1].InvoiceID.String,
						InvoiceStatus:   invoice_pb.InvoiceStatus(multipleStudentInvoiceWithMultiplePaymentEach[1].InvoiceID.Status),
						Total:           3333.33,
						DueDate:         timestamppb.New(multipleInvoiceMultiplePaymentEach[3].PaymentDueDate.Time),
					},
				},
				NextPage: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: constant.PageLimit + 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(multipleStudentInvoiceWithMultiplePaymentEach, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(multipleInvoiceMultiplePaymentEach[0], nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(multipleInvoiceMultiplePaymentEach[3], nil)
			},
		},
		{
			name:        "error retrieve invoice records failed tx closed",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, txClosedError),
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "error retrieve invoice records failed closed db pool",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, dbClosedPoolError),
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
		},
		{
			name:        "error on getting single payment db closed pool",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, dbClosedPoolError),
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithSingleInvoiceAndPayment.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(singleStudentInvoiceSinglePayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
		},
		{
			name:        "error on getting single payment tx closed",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, txClosedError),
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithSingleInvoiceAndPayment.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(singleStudentInvoiceSinglePayment, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "error on getting multiple payment tx closed",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, txClosedError),
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(multipleStudentInvoiceWithMultiplePaymentEach, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(multipleInvoiceMultiplePaymentEach[0], nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "error on getting multiple payment on db pool closed",
			ctx:         ctx,
			expectedErr: status.Error(codes.Internal, dbClosedPoolError),
			req: &invoice_pb.RetrieveInvoiceRecordsRequest{
				Paging: &cpb.Paging{
					Limit: constant.PageLimit,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				StudentId: studentWithMultipleInvoice.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(multipleStudentInvoiceWithMultiplePaymentEach, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(multipleInvoiceMultiplePaymentEach[0], nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.RetrieveInvoiceRecords(testCase.ctx, testCase.req.(*invoice_pb.RetrieveInvoiceRecordsRequest))
			if err != nil {
				fmt.Println(err)
			}
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			}
			if testCase.expectedResp != nil {
				assert.Equal(t, len(testCase.expectedResp.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords), len(resp.InvoiceRecords))
				assert.Equal(t, testCase.expectedResp.(*invoice_pb.RetrieveInvoiceRecordsResponse).NextPage.Limit, resp.NextPage.Limit)
				assert.Equal(t, testCase.expectedResp.(*invoice_pb.RetrieveInvoiceRecordsResponse).NextPage.Offset, resp.NextPage.Offset)

				if len(testCase.expectedResp.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords) > 0 && testCase.mockInvoiceEntities != nil && len(testCase.mockInvoiceEntities) > 0 {
					for i, invoiceRec := range testCase.expectedResp.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords {
						responseTotal := invoiceRec.Total
						requestTotal := testCase.mockInvoiceEntities[i].Total
						getExactValue, err := utils.GetFloat64ExactValueAndDecimalPlaces(requestTotal, "2")
						if err != nil {
							fmt.Println(err)
						}
						assert.Equal(t, getExactValue, responseTotal)
					}
				}

			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo)
		})
	}
}
