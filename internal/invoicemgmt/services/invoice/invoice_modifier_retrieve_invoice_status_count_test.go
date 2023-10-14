package invoicesvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
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

func TestInvoiceModifierService_RetrieveInvoiceStatusCount(t *testing.T) {
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

	statusCountMap := make(map[string]int32)
	emptyStatusCountMap := make(map[string]int32)
	statusCountMap[invoice_pb.InvoiceStatus_DRAFT.String()] = 10
	statusCountMap[invoice_pb.InvoiceStatus_ISSUED.String()] = 20
	statusCountMap[invoice_pb.InvoiceStatus_PAID.String()] = 30
	statusCountMap[invoice_pb.InvoiceStatus_VOID.String()] = 40
	statusCountMap[invoice_pb.InvoiceStatus_REFUNDED.String()] = 50

	testcases := []TestCase{
		{
			name: "happy case - no filter invoice only for status count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  &invoice_pb.RetrieveInvoiceStatusCountRequest{},
			expectedResp: &invoice_pb.RetrieveInvoiceStatusCountResponse{
				TotalItems: 150,
				InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
					TotalPaid:     30,
					TotalDraft:    10,
					TotalIssued:   20,
					TotalVoid:     40,
					TotalRefunded: 50,
				},
			},
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "happy case - with draft filter invoice and payment with student name for status count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
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
				},
				StudentName: "test",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceStatusCountResponse{
				TotalItems: 10,
				InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
					TotalPaid:     0,
					TotalDraft:    10,
					TotalIssued:   0,
					TotalVoid:     0,
					TotalRefunded: 0,
				},
			},
			setup: func(ctx context.Context) {
				statusCountMap[invoice_pb.InvoiceStatus_ISSUED.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_PAID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_VOID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_REFUNDED.String()] = 0
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "happy case - with issued filter invoice and payment with student name for status count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "58.00",
					MaxAmount:        "200.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					InvoiceStatus:    invoice_pb.InvoiceStatus_ISSUED,
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					PaymentMethods:  []invoice_pb.PaymentMethod{invoice_pb.PaymentMethod_CONVENIENCE_STORE, invoice_pb.PaymentMethod_DIRECT_DEBIT},
					DueDateFrom:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					DueDateUntil:    timestamppb.New(pgtype.Timestamptz{Time: time.Date(2023, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					PaymentStatuses: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_PENDING},
				},
				StudentName: "test2",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceStatusCountResponse{
				TotalItems: 50,
				InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
					TotalPaid:     0,
					TotalDraft:    0,
					TotalIssued:   50,
					TotalVoid:     0,
					TotalRefunded: 0,
				},
			},
			setup: func(ctx context.Context) {
				statusCountMap[invoice_pb.InvoiceStatus_ISSUED.String()] = 50
				statusCountMap[invoice_pb.InvoiceStatus_PAID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_VOID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_REFUNDED.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_DRAFT.String()] = 0
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "happy case - with paid filter invoice and payment with student name for status count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "56.00",
					MaxAmount:        "20.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					InvoiceStatus:    invoice_pb.InvoiceStatus_PAID,
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					PaymentMethods:  []invoice_pb.PaymentMethod{invoice_pb.PaymentMethod_CONVENIENCE_STORE, invoice_pb.PaymentMethod_DIRECT_DEBIT},
					DueDateFrom:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					DueDateUntil:    timestamppb.New(pgtype.Timestamptz{Time: time.Date(2023, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					PaymentStatuses: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL},
				},
				StudentName: "test2",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceStatusCountResponse{
				TotalItems: 330,
				InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
					TotalPaid:     330,
					TotalDraft:    0,
					TotalIssued:   0,
					TotalVoid:     0,
					TotalRefunded: 0,
				},
			},
			setup: func(ctx context.Context) {
				statusCountMap[invoice_pb.InvoiceStatus_ISSUED.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_PAID.String()] = 330
				statusCountMap[invoice_pb.InvoiceStatus_VOID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_REFUNDED.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_DRAFT.String()] = 0
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "happy case - with void filter invoice and payment with student name for status count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "51.00",
					MaxAmount:        "50.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					InvoiceStatus:    invoice_pb.InvoiceStatus_VOID,
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					PaymentMethods:  []invoice_pb.PaymentMethod{invoice_pb.PaymentMethod_CONVENIENCE_STORE, invoice_pb.PaymentMethod_DIRECT_DEBIT},
					DueDateFrom:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					DueDateUntil:    timestamppb.New(pgtype.Timestamptz{Time: time.Date(2023, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					PaymentStatuses: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_FAILED},
				},
				StudentName: "test2",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceStatusCountResponse{
				TotalItems: 32,
				InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
					TotalPaid:     0,
					TotalDraft:    0,
					TotalIssued:   0,
					TotalVoid:     32,
					TotalRefunded: 0,
				},
			},
			setup: func(ctx context.Context) {
				statusCountMap[invoice_pb.InvoiceStatus_ISSUED.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_PAID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_VOID.String()] = 32
				statusCountMap[invoice_pb.InvoiceStatus_REFUNDED.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_DRAFT.String()] = 0
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "happy case - with refunded filter invoice and payment with student name for status count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "55.00",
					MaxAmount:        "503.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					InvoiceStatus:    invoice_pb.InvoiceStatus_REFUNDED,
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					PaymentMethods:  []invoice_pb.PaymentMethod{invoice_pb.PaymentMethod_CONVENIENCE_STORE, invoice_pb.PaymentMethod_DIRECT_DEBIT},
					DueDateFrom:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					DueDateUntil:    timestamppb.New(pgtype.Timestamptz{Time: time.Date(2023, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					PaymentStatuses: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_FAILED},
				},
				StudentName: "test2",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceStatusCountResponse{
				TotalItems: 99,
				InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
					TotalPaid:     0,
					TotalDraft:    0,
					TotalIssued:   0,
					TotalVoid:     0,
					TotalRefunded: 99,
				},
			},
			setup: func(ctx context.Context) {
				statusCountMap[invoice_pb.InvoiceStatus_ISSUED.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_PAID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_VOID.String()] = 0
				statusCountMap[invoice_pb.InvoiceStatus_REFUNDED.String()] = 99
				statusCountMap[invoice_pb.InvoiceStatus_DRAFT.String()] = 0
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "happy case - with no status filter invoice and payment with student name for status count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "55.00",
					MaxAmount:        "503.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
				},
				PaymentFilter: &invoice_pb.InvoiceDataForPaymentFilter{
					PaymentMethods:  []invoice_pb.PaymentMethod{invoice_pb.PaymentMethod_CONVENIENCE_STORE, invoice_pb.PaymentMethod_DIRECT_DEBIT},
					DueDateFrom:     timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					DueDateUntil:    timestamppb.New(pgtype.Timestamptz{Time: time.Date(2023, 11, 11, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC)}.Time),
					ExpiryDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					PaymentStatuses: []invoice_pb.PaymentStatus{invoice_pb.PaymentStatus_PAYMENT_PENDING, invoice_pb.PaymentStatus_PAYMENT_FAILED},
				},
				StudentName: "test2",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceStatusCountResponse{
				TotalItems: 25,
				InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
					TotalPaid:     5,
					TotalDraft:    5,
					TotalIssued:   5,
					TotalVoid:     5,
					TotalRefunded: 5,
				},
			},
			setup: func(ctx context.Context) {
				statusCountMap[invoice_pb.InvoiceStatus_ISSUED.String()] = 5
				statusCountMap[invoice_pb.InvoiceStatus_PAID.String()] = 5
				statusCountMap[invoice_pb.InvoiceStatus_VOID.String()] = 5
				statusCountMap[invoice_pb.InvoiceStatus_REFUNDED.String()] = 5
				statusCountMap[invoice_pb.InvoiceStatus_DRAFT.String()] = 5
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "negative case - invalid invoice status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "55.00",
					MaxAmount:        "503.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
				},
			},
			expectedErr: status.Error(codes.Internal, "invalid invoice status: test"),
			setup: func(ctx context.Context) {
				statusCountMap["test"] = 5
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(statusCountMap, nil)
			},
		},
		{
			name: "negative case - err db scan set",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.RetrieveInvoiceStatusCountRequest{
				InvoiceFilter: &invoice_pb.InvoiceDataForInvoiceFilter{
					InvoiceTypes:     []invoice_pb.InvoiceType{invoice_pb.InvoiceType_MANUAL, invoice_pb.InvoiceType_SCHEDULED},
					MinAmount:        "55.00",
					MaxAmount:        "503.00",
					CreatedDateFrom:  timestamppb.New(pgtype.Timestamptz{Time: time.Date(2019, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
					CreatedDateUntil: timestamppb.New(pgtype.Timestamptz{Time: time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC)}.Time),
				},
			},
			expectedErr: status.Error(codes.Internal, "test error"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceStatusCount", ctx, mockDB, mock.Anything).Once().Return(emptyStatusCountMap, errors.New("test error"))
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.RetrieveInvoiceStatusCount(testCase.ctx, testCase.req.(*invoice_pb.RetrieveInvoiceStatusCountRequest))

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
