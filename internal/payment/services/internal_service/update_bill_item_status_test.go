package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/internal_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateBillItemStatus(t *testing.T) {
	t.Parallel()
	const (
		ctxUserID                           = "user-id"
		invalidBillingItemsParameters       = "invalid billing items parameters"
		defaultBillItemSequenceNumber int32 = 1
	)
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	mockDB := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockBillItemService := new(mockServices.IBillItemServiceForInternalService)
	mockOrderService := new(mockServices.IOrderServiceForInternalService)
	s := &InternalService{
		DB:              mockDB,
		billItemService: mockBillItemService,
		orderService:    mockOrderService,
	}

	testcases := []utils.TestCase{
		{
			Name:        "No Billing items status to update",
			Ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "billing items cannot be empty"),
			Req:         &pb.UpdateBillItemStatusRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: have error when update bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			Req: &pb.UpdateBillItemStatusRequest{
				UpdateBillItems: []*pb.UpdateBillItemStatusRequest_UpdateBillItem{
					{
						BillItemSequenceNumber: defaultBillItemSequenceNumber,
						BillingStatusTo:        pb.BillingStatus_BILLING_STATUS_WAITING_APPROVAL,
					},
				},
			},
			ExpectedResp: &pb.UpdateBillItemStatusResponse{
				Errors: []*pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{{
					BillItemSequenceNumber: defaultBillItemSequenceNumber,
					Error:                  constant.ErrDefault.Error(),
				},
				},
			},
			Setup: func(ctx context.Context) {
				mockBillItemService.On("UpdateBillItemStatusAndReturnOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return("1", constant.ErrDefault)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockDB.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
		{
			Name: "Happy case: have error when update order item",
			Ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			Req: &pb.UpdateBillItemStatusRequest{
				UpdateBillItems: []*pb.UpdateBillItemStatusRequest_UpdateBillItem{
					{
						BillItemSequenceNumber: defaultBillItemSequenceNumber,
						BillingStatusTo:        pb.BillingStatus_BILLING_STATUS_INVOICED,
					},
				},
			},
			ExpectedResp: &pb.UpdateBillItemStatusResponse{
				Errors: []*pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{{
					BillItemSequenceNumber: defaultBillItemSequenceNumber,
					Error:                  constant.ErrDefault.Error(),
				},
				},
			},
			Setup: func(ctx context.Context) {
				mockBillItemService.On("UpdateBillItemStatusAndReturnOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return("1", nil)
				mockOrderService.On("UpdateOrderStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockDB.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
		{
			Name: "Happy case: without error",
			Ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			Req: &pb.UpdateBillItemStatusRequest{
				UpdateBillItems: []*pb.UpdateBillItemStatusRequest_UpdateBillItem{
					{
						BillItemSequenceNumber: defaultBillItemSequenceNumber,
						BillingStatusTo:        pb.BillingStatus_BILLING_STATUS_INVOICED,
					},
				},
			},
			ExpectedResp: &pb.UpdateBillItemStatusResponse{
				Errors: []*pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{},
			},
			Setup: func(ctx context.Context) {
				mockBillItemService.On("UpdateBillItemStatusAndReturnOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return("1", nil)
				mockOrderService.On("UpdateOrderStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockDB.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.UpdateBillItemStatus(testCase.Ctx, testCase.Req.(*pb.UpdateBillItemStatusRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.UpdateBillItemStatusResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.BillItemSequenceNumber, expectedResp.Errors[i].BillItemSequenceNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockBillItemService, mockOrderService)
		})
	}
}
