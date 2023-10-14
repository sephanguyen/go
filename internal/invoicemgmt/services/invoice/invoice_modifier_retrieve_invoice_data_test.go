package invoicesvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_RetrieveInvoiceData(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Mock objects
	mockDB := new(mock_database.Ext)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)

	s := &InvoiceModifierService{
		DB:          mockDB,
		InvoiceRepo: mockInvoiceRepo,
	}

	invoiceMapOnly := []*entities.InvoicePaymentMap{
		{
			Invoice: &entities.Invoice{
				InvoiceID:             database.Text("1"),
				InvoiceSequenceNumber: database.Int4(1),
				Status:                database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
				StudentID:             database.Text("1"),
				SubTotal:              database.Numeric(50.00),
				Total:                 database.Numeric(50.00),
				OutstandingBalance:    database.Numeric(50.00),
				AmountPaid:            database.Numeric(0.00),
				Type:                  database.Text(invoice_pb.InvoiceType_MANUAL.String()),
				CreatedAt:             pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
			UserBasicInfo: &entities.UserBasicInfo{
				Name: database.Text("test"),
			},
		},
	}

	invoicePaymentMapMultipleInvoice := []*entities.InvoicePaymentMap{
		{
			Invoice: &entities.Invoice{
				InvoiceID:             database.Text("1"),
				InvoiceSequenceNumber: database.Int4(1),
				Status:                database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
				StudentID:             database.Text("1"),
				SubTotal:              database.Numeric(50.00),
				Total:                 database.Numeric(50.00),
				OutstandingBalance:    database.Numeric(50.00),
				AmountPaid:            database.Numeric(0.00),
				Type:                  database.Text(invoice_pb.InvoiceType_MANUAL.String()),
				CreatedAt:             pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
			UserBasicInfo: &entities.UserBasicInfo{
				Name: database.Text("test"),
			},
		},
		{
			Invoice: &entities.Invoice{
				InvoiceID:             database.Text("2"),
				InvoiceSequenceNumber: database.Int4(2),
				Status:                database.Text(invoice_pb.InvoiceStatus_VOID.String()),
				StudentID:             database.Text("2"),
				SubTotal:              database.Numeric(510.00),
				Total:                 database.Numeric(510.00),
				OutstandingBalance:    database.Numeric(510.00),
				AmountPaid:            database.Numeric(0.00),
				Type:                  database.Text(invoice_pb.InvoiceType_MANUAL.String()),
				CreatedAt:             pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
			UserBasicInfo: &entities.UserBasicInfo{
				Name: database.Text("test2"),
			},
		},
	}

	invoicePaymentMap := []*entities.InvoicePaymentMap{
		{
			Invoice: &entities.Invoice{
				InvoiceID:             database.Text("1"),
				InvoiceSequenceNumber: database.Int4(1),
				Status:                database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
				StudentID:             database.Text("1"),
				SubTotal:              database.Numeric(50.00),
				Total:                 database.Numeric(50.00),
				OutstandingBalance:    database.Numeric(50.00),
				AmountPaid:            database.Numeric(0.00),
				Type:                  database.Text(invoice_pb.InvoiceType_MANUAL.String()),
				CreatedAt:             pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
			Payment: &entities.Payment{
				PaymentID:             database.Text("1"),
				PaymentSequenceNumber: database.Int4(1),
				IsExported:            database.Bool(false),
				PaymentDueDate:        pgtype.Timestamptz{Time: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC)},
				PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Date(2020, 11, 12, 0, 0, 0, 0, time.UTC)},
				PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			},
			UserBasicInfo: &entities.UserBasicInfo{
				Name: database.Text("test"),
			},
		},
	}

	invoicePaymentMapMultiple := []*entities.InvoicePaymentMap{
		{
			Invoice: &entities.Invoice{
				InvoiceID:             database.Text("1"),
				InvoiceSequenceNumber: database.Int4(1),
				Status:                database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
				StudentID:             database.Text("1"),
				SubTotal:              database.Numeric(50.00),
				Total:                 database.Numeric(50.00),
				OutstandingBalance:    database.Numeric(50.00),
				AmountPaid:            database.Numeric(0.00),
				Type:                  database.Text(invoice_pb.InvoiceType_MANUAL.String()),
				CreatedAt:             pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
			Payment: &entities.Payment{
				PaymentID:             database.Text("1"),
				PaymentSequenceNumber: database.Int4(1),
				IsExported:            database.Bool(false),
				PaymentDueDate:        pgtype.Timestamptz{Time: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC)},
				PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Date(2020, 11, 12, 0, 0, 0, 0, time.UTC)},
				PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
			},
			UserBasicInfo: &entities.UserBasicInfo{
				Name: database.Text("test"),
			},
		},
		{
			Invoice: &entities.Invoice{
				InvoiceID:             database.Text("2"),
				InvoiceSequenceNumber: database.Int4(2),
				Status:                database.Text(invoice_pb.InvoiceStatus_PAID.String()),
				StudentID:             database.Text("2"),
				SubTotal:              database.Numeric(530.00),
				Total:                 database.Numeric(530.00),
				OutstandingBalance:    database.Numeric(0.00),
				AmountPaid:            database.Numeric(530.00),
				Type:                  database.Text(invoice_pb.InvoiceType_MANUAL.String()),
				CreatedAt:             pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
			Payment: &entities.Payment{
				PaymentID:             database.Text("2"),
				PaymentSequenceNumber: database.Int4(2),
				IsExported:            database.Bool(true),
				PaymentDueDate:        pgtype.Timestamptz{Time: time.Date(2020, 10, 11, 0, 0, 0, 0, time.UTC)},
				PaymentExpiryDate:     pgtype.Timestamptz{Time: time.Date(2020, 10, 12, 0, 0, 0, 0, time.UTC)},
				PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
				PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()),
			},
			UserBasicInfo: &entities.UserBasicInfo{
				Name: database.Text("test2"),
			},
		},
	}

	testcases := []TestCase{
		{
			name: "happy case - no filter invoice only",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						StudentName: "test",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoiceMapOnly, nil)
			},
		},
		{
			name: "happy case - no filter multiple invoice only",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						StudentName: "test",
					},
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "2",
							InvoiceSequenceNumber: 2,
							StudentId:             "2",
							SubTotal:              510.00,
							Total:                 510.00,
							OutstandingBalance:    510.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_VOID,
						},
						StudentName: "test2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoicePaymentMapMultipleInvoice, nil)
			},
		},
		{
			name: "happy case - no filter invoice and payment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						InvoiceDataPaymentDetail: &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
							PaymentId:             "1",
							PaymentSequenceNumber: 1,
							IsExported:            false,
							PaymentDueDate:        timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentExpiryDate:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 12, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
							PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING,
						},
						StudentName: "test",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoicePaymentMap, nil)
			},
		},
		{
			name: "happy case - with filter invoice only",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "5.00",
					MaxAmount:        "500.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					InvoiceStatus:    invoice_pb.InvoiceStatus_DRAFT,
				},
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						StudentName: "test",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoiceMapOnly, nil)
			},
		},
		{
			name: "happy case - with filter invoice and payment",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "5.00",
					MaxAmount:        "500.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					InvoiceStatus:    invoice_pb.InvoiceStatus_DRAFT,
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					PaymentMethods:  []invoice_pb.PaymentMethod{invoice_pb.PaymentMethod_CONVENIENCE_STORE, invoice_pb.PaymentMethod_DIRECT_DEBIT},
					DueDateFrom:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					DueDateUntil:    timestamppb.New(pgtype.Timestamptz{Time: time.Date(2023, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					PaymentStatuses: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL},
				},
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							AmountPaid:            0.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						InvoiceDataPaymentDetail: &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
							PaymentId:             "1",
							PaymentSequenceNumber: 1,
							IsExported:            false,
							PaymentDueDate:        timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentExpiryDate:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 12, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
							PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING,
						},
						StudentName: "test",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoicePaymentMap, nil)
			},
		},
		{
			name: "happy case - with filter multiple invoice and payment with student name",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "5.00",
					MaxAmount:        "500.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					InvoiceStatus:    invoice_pb.InvoiceStatus_DRAFT,
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					PaymentMethods:  []invoice_pb.PaymentMethod{invoice_pb.PaymentMethod_CONVENIENCE_STORE, invoice_pb.PaymentMethod_DIRECT_DEBIT},
					DueDateFrom:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					DueDateUntil:    timestamppb.New(pgtype.Timestamptz{Time: time.Date(2023, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					PaymentStatuses: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL},
				},
				StudentName: "test",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							AmountPaid:            0.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						InvoiceDataPaymentDetail: &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
							PaymentId:             "1",
							PaymentSequenceNumber: 1,
							IsExported:            false,
							PaymentDueDate:        timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentExpiryDate:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 12, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
							PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING,
						},
						StudentName: "test",
					},
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "2",
							InvoiceSequenceNumber: 2,
							StudentId:             "2",
							SubTotal:              530.00,
							Total:                 530.00,
							OutstandingBalance:    0.00,
							AmountPaid:            530.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_PAID,
						},
						InvoiceDataPaymentDetail: &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
							PaymentId:             "2",
							PaymentSequenceNumber: 2,
							IsExported:            true,
							PaymentDueDate:        timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 11, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentExpiryDate:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 12, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
							PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
						},
						StudentName: "test2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoicePaymentMapMultiple, nil)
			},
		},
		{
			name: "happy case - with filter is exported set to true",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					IsExported: true,
				},
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							AmountPaid:            0.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						InvoiceDataPaymentDetail: &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
							PaymentId:             "1",
							PaymentSequenceNumber: 1,
							IsExported:            false,
							PaymentDueDate:        timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentExpiryDate:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 12, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
							PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING,
						},
						StudentName: "test",
					},
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "2",
							InvoiceSequenceNumber: 2,
							StudentId:             "2",
							SubTotal:              530.00,
							Total:                 530.00,
							OutstandingBalance:    0.00,
							AmountPaid:            530.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_PAID,
						},
						InvoiceDataPaymentDetail: &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
							PaymentId:             "2",
							PaymentSequenceNumber: 2,
							IsExported:            true,
							PaymentDueDate:        timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 11, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentExpiryDate:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 12, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
							PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
						},
						StudentName: "test2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoicePaymentMapMultiple, nil)
			},
		},
		{
			name: "happy case - with filter is exported set to false",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					IsExported: true,
				},
			},
			expectedResp: &invoice_pb.RetrieveInvoiceDataResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				InvoiceData: []*invoice_pb.InvoiceData{
					{
						InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
							InvoiceId:             "1",
							InvoiceSequenceNumber: 1,
							StudentId:             "1",
							SubTotal:              50.00,
							Total:                 50.00,
							OutstandingBalance:    50.00,
							InvoiceType:           invoice_pb.InvoiceType_MANUAL,
							CreatedAt:             timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
							InvoiceStatus:         invoice_pb.InvoiceStatus_DRAFT,
						},
						InvoiceDataPaymentDetail: &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
							PaymentId:             "1",
							PaymentSequenceNumber: 1,
							IsExported:            false,
							PaymentDueDate:        timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentExpiryDate:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 11, 12, 0, 0, 0, 0, time.UTC)}.Time),
							PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
							PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING,
						},
						StudentName: "test",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(invoicePaymentMap, nil)
			},
		},
		{
			name: "negative case - no rows result set",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative case - err db scan set",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceDataRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "test error"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceData", ctx, mockDB, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, errors.New("test error"))
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.RetrieveInvoiceData(testCase.ctx, testCase.req.(*invoice_pb.RetrieveInvoiceDataRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedResp, response)
				assert.NotNil(t, response)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo)
		})
	}
}
