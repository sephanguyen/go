package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mockRepositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_services "github.com/manabie-com/backend/mock/invoicemgmt/services"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func genMultipleInvoiceRequest(count int) *invoice_pb.GenerateInvoicesRequest {
	req := &invoice_pb.GenerateInvoicesRequest{}

	for i := 0; i < count; i++ {
		req.Invoices = append(req.Invoices, &invoice_pb.GenerateInvoiceDetail{
			InvoiceType: invoice_pb.InvoiceType_MANUAL,
			BillItemIds: []int32{1, 2, 3},
			StudentId:   fmt.Sprintf("test-student-id-%d", i),
			SubTotal:    500,
			Total:       500,
		})
	}

	return req
}

func TestInvoiceModifierService_GenerateInvoices(t *testing.T) {
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
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		InternalOrderService: mockOrderServiceClient,
		UnleashClient:        mockUnleashClient,
	}

	invoice1Request := genMultipleInvoiceRequest(1)
	invoice100Request := genMultipleInvoiceRequest(100)
	invalidRequest := genMultipleInvoiceRequest(0)
	requestWithInvalidStudentID := &invoice_pb.GenerateInvoicesRequest{
		Invoices: []*invoice_pb.GenerateInvoiceDetail{
			{
				StudentId: "",
			},
		},
	}
	requestWithInvalidBillItemIDs := &invoice_pb.GenerateInvoicesRequest{
		Invoices: []*invoice_pb.GenerateInvoiceDetail{
			{
				StudentId:   "test-student",
				BillItemIds: []int32{},
			},
		},
	}

	invoiceResponseHappyCase := &invoice_pb.GenerateInvoicesResponse{
		Successful:   true,
		InvoicesData: []*invoice_pb.GenerateInvoicesResponse_InvoicesData{},
	}

	mockBilledBillItem := &entities.BillItem{BillStatus: pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()}, IsReviewed: pgtype.Bool{Bool: true}}
	mockInvoicedBillItem := &entities.BillItem{BillStatus: pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()}}
	testErr := errors.New("test error")

	orderUpdateBillStatusRespErr := &payment_pb.UpdateBillItemStatusResponse{
		Errors: []*payment_pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{
			{
				BillItemSequenceNumber: 1,
				Error:                  pgx.ErrNoRows.Error(),
			},
		},
	}

	testcases := []TestCase{
		{
			name:         "Happy case with 100 invoices",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          invoice100Request,
			expectedResp: invoiceResponseHappyCase,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				for i := 0; i < 100; i++ {
					mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(3).Return(mockBilledBillItem, nil)
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(3).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Times(1).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				}
			},
		},
		{
			name:        "Generate invoice invalid request",
			ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			req:         invalidRequest,
			expectedErr: status.Error(codes.InvalidArgument, "Invoices cannot be empty"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "Invalid Student ID",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  requestWithInvalidStudentID,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: requestWithInvalidStudentID.Invoices[0],
						Error:         status.Error(codes.InvalidArgument, "Student ID cannot be empty").Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "Invalid Bill Item ID",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  requestWithInvalidBillItemIDs,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful:   false,
				InvoicesData: []*invoice_pb.GenerateInvoicesResponse_InvoicesData{},
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: requestWithInvalidStudentID.Invoices[0],
						Error:         status.Error(codes.InvalidArgument, fmt.Sprintf("Bill Items of student %s cannot be empty", requestWithInvalidStudentID.Invoices[0].StudentId)).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "Error on bill item with invalid status",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  invoice1Request,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: invoice1Request.Invoices[0],
						Error:         fmt.Errorf("bill item with ID %d has an invalid status %s", invoice1Request.Invoices[0].BillItemIds[0], mockInvoicedBillItem.BillStatus.String).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(1).Return(mockInvoicedBillItem, nil)
				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Error on fetching bill items",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  invoice1Request,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: invoice1Request.Invoices[0],
						Error:         fmt.Errorf("s.BillItemRepo.FindByID err: %w", pgx.ErrNoRows).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(1).Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Error on creating invoice",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  invoice1Request,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: invoice1Request.Invoices[0],
						Error:         fmt.Errorf("error Invoice Create: %v", testErr).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(3).Return(mockBilledBillItem, nil)
				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgtype.Text{}, testErr)
				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Error on creating invoice bill item",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  invoice1Request,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: invoice1Request.Invoices[0],
						Error:         fmt.Errorf("error Process Bill Items: %v", testErr).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(3).Return(mockBilledBillItem, nil)
				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{}).InvoiceID, nil)
				mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(testErr)
				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Error on requesting update bill item in order management",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  invoice1Request,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: invoice1Request.Invoices[0],
						Error:         fmt.Errorf("error Update When Bill Items Statuses Changed: %v", testErr).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(3).Return(mockBilledBillItem, nil)
				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{}).InvoiceID, nil)
				mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(3).Return(nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Times(1).Return(nil, testErr)
				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Requesting update bill item in order management contains an error with bill item",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  invoice1Request,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: invoice1Request.Invoices[0],
						Error:         fmt.Errorf("error UpdateBillItemStatus: %v", strings.Join([]string{fmt.Sprintf("BillItemSequenceNumber %v with error %v", invoice1Request.Invoices[0].BillItemIds[0], pgx.ErrNoRows)}, ",")).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(3).Return(mockBilledBillItem, nil)
				mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{}).InvoiceID, nil)
				mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(3).Return(nil)
				mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Times(1).Return(orderUpdateBillStatusRespErr, nil)
				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Error on creating invoice due to billing item having Review Required tag",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req:  invoice1Request,
			expectedResp: &invoice_pb.GenerateInvoicesResponse{
				Successful: false,
				Errors: []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
					{
						InvoiceDetail: invoice1Request.Invoices[0],
						Error:         fmt.Errorf("error Process Bill Items: %v", testErr).Error(),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

				mockBilledBillItem.IsReviewed.Set(false)
				mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(1).Return(mockBilledBillItem, nil)
				mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.GenerateInvoices(testCase.ctx, testCase.req.(*invoice_pb.GenerateInvoicesRequest))
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
				mockOrderServiceClient,
				mockDB,
				mockUnleashClient,
			)
		})
	}
}

func TestInvoiceModifierService_GenerateInvoices_ReviewOrderDisabled(t *testing.T) {
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
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		InternalOrderService: mockOrderServiceClient,
		UnleashClient:        mockUnleashClient,
	}

	invoice100Request := genMultipleInvoiceRequest(100)

	invoiceResponseHappyCase := &invoice_pb.GenerateInvoicesResponse{
		Successful:   true,
		InvoicesData: []*invoice_pb.GenerateInvoicesResponse_InvoicesData{},
	}

	testError := errors.New("test-error")

	mockBilledBillItem := &entities.BillItem{BillStatus: pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()}, IsReviewed: pgtype.Bool{Bool: true}}

	mockBilledBillItemReviewRequired := &entities.BillItem{BillStatus: pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()}, IsReviewed: pgtype.Bool{Bool: false}}

	testcases := []TestCase{
		{
			name:         "Happy case with 100 invoices",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          invoice100Request,
			expectedResp: invoiceResponseHappyCase,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)

				for i := 0; i < 100; i++ {
					mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(3).Return(mockBilledBillItem, nil)
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(3).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Times(1).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				}
			},
		},
		{
			name:         "Happy case with 100 invoices that have review required tag bill item",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          invoice100Request,
			expectedResp: invoiceResponseHappyCase,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)

				for i := 0; i < 100; i++ {
					mockBillItemRepo.On("FindByID", ctx, mockTx, mock.Anything).Times(3).Return(mockBilledBillItemReviewRequired, nil)
					mockInvoiceRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", ctx, mockTx, mock.Anything).Times(3).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", ctx, mock.Anything).Times(1).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				}
			},
		},
		{
			name:         "negative case - error on IsFeatureEnabled",
			ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			req:          invoice100Request,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.UnleashClient.IsFeatureEnabled err: %v", testError)),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.GenerateInvoices(testCase.ctx, testCase.req.(*invoice_pb.GenerateInvoicesRequest))
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
				mockOrderServiceClient,
				mockDB,
				mockUnleashClient,
			)
		})
	}
}
