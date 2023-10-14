package ordermgmt

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateOrderReviewFlag_UpdateOrderReviewedFlag(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db              *mockDb.Ext
		tx              *mockDb.Tx
		orderService    *mockServices.IOrderServiceForUpdateOrderReviewFlag
		billItemService *mockServices.IBillItemServiceForUpdateOrderReviewFlag
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when update review flag for order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.UpdateOrderReviewedFlagRequest{
				OrderId:    "1",
				IsReviewed: true,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("UpdateOrderReview", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when update review flag for bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.UpdateOrderReviewedFlagRequest{
				OrderId:    "1",
				IsReviewed: true,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("UpdateOrderReview", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemService.On("UpdateReviewFlagForBillItem", ctx, tx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.UpdateOrderReviewedFlagRequest{
				OrderId:    "1",
				IsReviewed: true,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("UpdateOrderReview", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemService.On("UpdateReviewFlagForBillItem", ctx, tx, mock.Anything, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			orderService = new(mockServices.IOrderServiceForUpdateOrderReviewFlag)
			billItemService = new(mockServices.IBillItemServiceForUpdateOrderReviewFlag)
			testCase.Setup(testCase.Ctx)
			s := &UpdateOrderReviewFlag{
				DB:              db,
				OrderService:    orderService,
				BillItemService: billItemService,
			}

			resp, err := s.UpdateOrderReviewedFlag(testCase.Ctx, testCase.Req.(*pb.UpdateOrderReviewedFlagRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				assert.Equal(t, resp.Successful, true)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				orderService,
				billItemService,
			)
		})
	}
}
