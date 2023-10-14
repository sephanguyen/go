package invoicesvc

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
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mockRepositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_services "github.com/manabie-com/backend/mock/invoicemgmt/services"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func generateCreateInvoiceFromOrderRequest(count int) *invoice_pb.CreateInvoiceFromOrderRequest {

	req := &invoice_pb.CreateInvoiceFromOrderRequest{}
	for i := 0; i < count; i++ {
		req.OrderDetails = append(req.OrderDetails, &invoice_pb.OrderDetail{
			OrderId: fmt.Sprintf("test-order-id-%d", i),
		})
	}

	return req
}

func TestInvoiceModifierService_CreateInvoiceFromOrder(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockTx := &mockDb.Tx{}
	mockDB := &mockDb.Ext{}
	mockInvoiceRepo := new(mockRepositories.MockInvoiceRepo)
	mockInvoiceBillItemRepo := new(mockRepositories.MockInvoiceBillItemRepo)
	mockBillItemRepo := new(mockRepositories.MockBillItemRepo)
	mockOrderServiceClient := new(mock_services.OrderService)
	mockInvoiceScheduleRepo := new(mockRepositories.MockInvoiceScheduleRepo)
	mockOrderRepo := new(mockRepositories.MockOrderRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		OrderService:         mockOrderServiceClient,
		InternalOrderService: mockOrderServiceClient,
		InvoiceScheduleRepo:  mockInvoiceScheduleRepo,
		OrderRepo:            mockOrderRepo,
		UnleashClient:        mockUnleashClient,
	}

	request := generateCreateInvoiceFromOrderRequest(3)
	testErr := errors.New("test error")

	testcases := []TestCase{
		{
			name: "Happy case",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				for i, r := range request.OrderDetails {
					mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
						OrderID:     database.Text(r.OrderId),
						OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
						IsReviewed:  database.Bool(true),
						StudentID:   database.Text(fmt.Sprintf("test-student-id-%d", i)),
					}, nil)
				}

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				for _, r := range request.OrderDetails {
					mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC()),
						},
					}, nil)
				}

				for i := 0; i < len(request.OrderDetails); i++ {
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{
						InvoiceID: database.Text(fmt.Sprintf("test-invoice-id-%d", i)),
					}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "Happy case - student have multiple order",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				for _, r := range request.OrderDetails {
					mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
						OrderID:     database.Text(r.OrderId),
						OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
						IsReviewed:  database.Bool(true),
						StudentID:   database.Text(fmt.Sprintf("test-student-id-1")),
					}, nil)
				}

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				for _, r := range request.OrderDetails {
					mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC()),
						},
						{
							OrderID:         database.Text(r.OrderId),
							BillStatus:      database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:        database.Timestamptz(time.Now().UTC()),
							AdjustmentPrice: database.Numeric(float32(10)),
							BillType:        database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
						},
					}, nil)
				}

				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{
					InvoiceID: database.Text(fmt.Sprintf("test-invoice-id")),
				}).InvoiceID, nil)
				mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(6).Return(nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Happy case - one billing date is greater than the upcoming invoice date",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				for i, r := range request.OrderDetails {
					mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
						OrderID:     database.Text(r.OrderId),
						OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
						IsReviewed:  database.Bool(true),
						StudentID:   database.Text(fmt.Sprintf("test-student-id-%d", i)),
					}, nil)
				}

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				for _, r := range request.OrderDetails {
					mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC()),
						},
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC().AddDate(0, 0, 20)),
						},
					}, nil)
				}

				for i := 0; i < len(request.OrderDetails); i++ {
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{
						InvoiceID: database.Text(fmt.Sprintf("test-invoice-id-%d", i)),
					}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name:        "Empty OrderDetails error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         &invoice_pb.CreateInvoiceFromOrderRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "the OrderDetails should not be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Empty OrderID error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.CreateInvoiceFromOrderRequest{
				OrderDetails: []*invoice_pb.OrderDetail{{OrderId: ""}},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("the order ID cannot be empty").Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name:        "OrderRepo.FindByOrderID error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.OrderRepo.FindByOrderID err: %v", testErr)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(nil, testErr)
			},
		},
		{
			name:        "One order has invalid status error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.InvalidArgument, "order status should be SUBMITTED"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_VOIDED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)
			},
		},
		{
			name:        "One order has review required tag error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.InvalidArgument, "order should not contain review required tag"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(false),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)
			},
		},
		{
			name:        "InvoiceScheduleRepo.GetCurrentEarliestInvoiceSchedule error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         generateCreateInvoiceFromOrderRequest(1),
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.InvoiceScheduleRepo.GetCurrentEarliestInvoiceSchedule err: %v", testErr)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(nil, testErr)
			},
		},
		{
			name:        "BillItemRepo.FindByOrderID error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         generateCreateInvoiceFromOrderRequest(1),
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BillItemRepo.FindByOrderID err: %v", testErr)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return(nil, testErr)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "no bill items found in order error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         generateCreateInvoiceFromOrderRequest(1),
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("order with ID %s has no associated billing item", "test-order-id-1")),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{}, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "bill item has invalid status error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  generateCreateInvoiceFromOrderRequest(1),
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("the bill item %v of order with ID %s has invalid status %s",
				1, "test-order-id-1", payment_pb.BillingStatus_BILLING_STATUS_INVOICED)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
					{
						BillItemSequenceNumber: database.Int4(1),
						BillStatus:             database.Text(payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()),
						OrderID:                database.Text("test-order-id-1"),
					},
				}, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "createInvoice error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         generateCreateInvoiceFromOrderRequest(1),
			expectedErr: status.Error(codes.Internal, fmt.Errorf("error Invoice Create: %v", testErr).Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
					{
						OrderID:    database.Text("test-order-id-1"),
						BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
						BillDate:   database.Timestamptz(time.Now().UTC()),
					},
				}, nil)

				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(database.Text(""), testErr)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "Billing with ADJUSTMENT_BILLING type has no present adjustment_price error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         generateCreateInvoiceFromOrderRequest(1),
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("The bill item %d with type BILLING_TYPE_ADJUSTMENT_BILLING has no present adjustment price", 1)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
					{
						OrderID:                database.Text("test-order-id-1"),
						BillStatus:             database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
						BillDate:               database.Timestamptz(time.Now().UTC()),
						BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
						BillItemSequenceNumber: database.Int4(1),
					},
				}, nil)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "Billing with present adjustment price has no BILLING_TYPE_ADJUSTMENT_BILLING billing type error",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         generateCreateInvoiceFromOrderRequest(1),
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("The bill item %d has present adjustment price but has no BILLING_TYPE_ADJUSTMENT_BILLING type", 1)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
				mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
					OrderID:     database.Text("test-order-id-1"),
					OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
					IsReviewed:  database.Bool(true),
					StudentID:   database.Text("test-student-id-1"),
				}, nil)

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
					{
						OrderID:                database.Text("test-order-id-1"),
						BillStatus:             database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
						BillDate:               database.Timestamptz(time.Now().UTC()),
						BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String()),
						BillItemSequenceNumber: database.Int4(1),
						AdjustmentPrice:        database.Numeric(10),
					},
				}, nil)

				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.CreateInvoiceFromOrder(testCase.ctx, testCase.req.(*invoice_pb.CreateInvoiceFromOrderRequest))
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert := assert.New(t)
				assert.NoError(err)
			}

			mock.AssertExpectationsForObjects(
				t,
				mockInvoiceBillItemRepo,
				mockInvoiceRepo,
				mockBillItemRepo,
				mockInvoiceScheduleRepo,
				mockOrderRepo,
				mockOrderServiceClient,
				mockDB,
				mockUnleashClient,
			)
		})
	}
}

func TestInvoiceModifierService_CreateInvoiceFromOrder_ReviewOrderDisabled(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockTx := &mockDb.Tx{}
	mockDB := &mockDb.Ext{}
	mockInvoiceRepo := new(mockRepositories.MockInvoiceRepo)
	mockInvoiceBillItemRepo := new(mockRepositories.MockInvoiceBillItemRepo)
	mockBillItemRepo := new(mockRepositories.MockBillItemRepo)
	mockOrderServiceClient := new(mock_services.OrderService)
	mockInvoiceScheduleRepo := new(mockRepositories.MockInvoiceScheduleRepo)
	mockOrderRepo := new(mockRepositories.MockOrderRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		OrderService:         mockOrderServiceClient,
		InternalOrderService: mockOrderServiceClient,
		InvoiceScheduleRepo:  mockInvoiceScheduleRepo,
		OrderRepo:            mockOrderRepo,
		UnleashClient:        mockUnleashClient,
	}

	testError := errors.New("test-error")

	request := generateCreateInvoiceFromOrderRequest(3)

	testcases := []TestCase{
		{
			name: "Happy case - with order not reviewed",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)

				for i, r := range request.OrderDetails {
					mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
						OrderID:     database.Text(r.OrderId),
						OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
						IsReviewed:  database.Bool(false),
						StudentID:   database.Text(fmt.Sprintf("test-student-id-%d", i)),
					}, nil)
				}

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				for _, r := range request.OrderDetails {
					mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC()),
						},
					}, nil)
				}

				for i := 0; i < len(request.OrderDetails); i++ {
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{
						InvoiceID: database.Text(fmt.Sprintf("test-invoice-id-%d", i)),
					}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "Happy case - student have multiple order that is not reviewed",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)
				for _, r := range request.OrderDetails {
					mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
						OrderID:     database.Text(r.OrderId),
						OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
						IsReviewed:  database.Bool(false),
						StudentID:   database.Text(fmt.Sprintf("test-student-id-1")),
					}, nil)
				}

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				for _, r := range request.OrderDetails {
					mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC()),
						},
						{
							OrderID:         database.Text(r.OrderId),
							BillStatus:      database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:        database.Timestamptz(time.Now().UTC()),
							AdjustmentPrice: database.Numeric(float32(10)),
							BillType:        database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
						},
					}, nil)
				}

				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{
					InvoiceID: database.Text(fmt.Sprintf("test-invoice-id")),
				}).InvoiceID, nil)
				mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(6).Return(nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Happy case - one billing date is greater than the upcoming invoice date",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)

				for i, r := range request.OrderDetails {
					mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
						OrderID:     database.Text(r.OrderId),
						OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
						IsReviewed:  database.Bool(true),
						StudentID:   database.Text(fmt.Sprintf("test-student-id-%d", i)),
					}, nil)
				}

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				for _, r := range request.OrderDetails {
					mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC()),
						},
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC().AddDate(0, 0, 20)),
						},
					}, nil)
				}

				for i := 0; i < len(request.OrderDetails); i++ {
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{
						InvoiceID: database.Text(fmt.Sprintf("test-invoice-id-%d", i)),
					}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "Happy case - order is reviewed",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)

				for i, r := range request.OrderDetails {
					mockOrderRepo.On("FindByOrderID", ctx, mockDB, mock.Anything).Once().Return(&entities.Order{
						OrderID:     database.Text(r.OrderId),
						OrderStatus: database.Text(payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
						IsReviewed:  database.Bool(true),
						StudentID:   database.Text(fmt.Sprintf("test-student-id-%d", i)),
					}, nil)
				}

				mockInvoiceScheduleRepo.On("GetCurrentEarliestInvoiceSchedule", ctx, mockDB, mock.Anything).Once().Return(&entities.InvoiceSchedule{
					InvoiceDate: database.Timestamptz(time.Now().UTC().AddDate(0, 0, 10)),
				}, nil)

				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				for _, r := range request.OrderDetails {
					mockBillItemRepo.On("FindByOrderID", ctx, mockTx, mock.Anything).Once().Return([]*entities.BillItem{
						{
							OrderID:    database.Text(r.OrderId),
							BillStatus: database.Text(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
							BillDate:   database.Timestamptz(time.Now().UTC()),
						},
					}, nil)
				}

				for i := 0; i < len(request.OrderDetails); i++ {
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{
						InvoiceID: database.Text(fmt.Sprintf("test-invoice-id-%d", i)),
					}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Once().Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name:        "negative case - error on IsFeatureEnabled",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         request,
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.UnleashClient.IsFeatureEnabled err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.CreateInvoiceFromOrder(testCase.ctx, testCase.req.(*invoice_pb.CreateInvoiceFromOrderRequest))
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert := assert.New(t)
				assert.NoError(err)
			}

			mock.AssertExpectationsForObjects(
				t,
				mockInvoiceBillItemRepo,
				mockInvoiceRepo,
				mockBillItemRepo,
				mockInvoiceScheduleRepo,
				mockOrderRepo,
				mockOrderServiceClient,
				mockDB,
				mockUnleashClient,
			)
		})
	}
}
