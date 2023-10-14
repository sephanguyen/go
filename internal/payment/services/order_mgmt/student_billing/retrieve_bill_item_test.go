package studentbilling

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudentBilling_RetrieveListOfBillItems(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.IStudentProductForStudentBilling
		billItemService       *mockServices.IBillItemServiceForStudentBilling
		orderService          *mockServices.IOrderServiceForStudentBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: with nil paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfBillItemsRequest{
				StudentId: "1",
			},
			ExpectedErr: fmt.Errorf("invalid paging data with error"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: when get bill description by studentID",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfBillItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				adjustmentPrice := pgtype.Numeric{}
				price := pgtype.Numeric{}
				_ = price.Set(123)
				_ = adjustmentPrice.Set(123)
				billItemService.On("GetBillItemDescriptionByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]utils.BillItemForRetrieveApi{
					{
						BillItemEntity: entities.BillItem{
							AdjustmentPrice: adjustmentPrice,
						},
					},
					{
						BillItemEntity: entities.BillItem{
							FinalPrice: adjustmentPrice,
						},
					},
				}, 0, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: when get order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfBillItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				adjustmentPrice := pgtype.Numeric{}
				price := pgtype.Numeric{}
				_ = price.Set(123)
				_ = adjustmentPrice.Set(123)
				billItemService.On("GetBillItemDescriptionByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]utils.BillItemForRetrieveApi{
					{
						BillItemEntity: entities.BillItem{
							AdjustmentPrice: adjustmentPrice,
						},
					},
					{
						BillItemEntity: entities.BillItem{
							FinalPrice: adjustmentPrice,
						},
					},
				}, 0, nil)
				orderService.On("GetOrderTypeByOrderID", ctx, mock.Anything, mock.Anything).Return("", constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: get paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfBillItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: status.Error(codes.Internal, "Error offset"),
			Setup: func(ctx context.Context) {
				adjustmentPrice := pgtype.Numeric{}
				price := pgtype.Numeric{}
				_ = price.Set(123)
				_ = adjustmentPrice.Set(123)
				billItemService.On("GetBillItemDescriptionByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]utils.BillItemForRetrieveApi{
					{
						BillItemEntity: entities.BillItem{
							AdjustmentPrice: adjustmentPrice,
						},
					},
					{
						BillItemEntity: entities.BillItem{
							FinalPrice: adjustmentPrice,
						},
					},
				}, 0, nil)
				orderService.On("GetOrderTypeByOrderID", ctx, mock.Anything, mock.Anything).Return("", nil)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfBillItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				adjustmentPrice := pgtype.Numeric{}
				price := pgtype.Numeric{}
				_ = price.Set(123)
				_ = adjustmentPrice.Set(123)
				billItemService.On("GetBillItemDescriptionByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]utils.BillItemForRetrieveApi{
					{
						BillItemEntity: entities.BillItem{
							AdjustmentPrice: adjustmentPrice,
							BillDate: pgtype.Timestamptz{
								Time: time.Now().Add(-5 * time.Second),
							},
						},
					},
					{
						BillItemEntity: entities.BillItem{
							FinalPrice: adjustmentPrice,
							BillDate: pgtype.Timestamptz{
								Time: time.Now().Add(5 * time.Second),
							},
						},
					},
				}, 20, nil)
				orderService.On("GetOrderTypeByOrderID", ctx, mock.Anything, mock.Anything).Return("", nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.IStudentProductForStudentBilling)
			billItemService = new(mockServices.IBillItemServiceForStudentBilling)
			orderService = new(mockServices.IOrderServiceForStudentBilling)
			testCase.Setup(testCase.Ctx)
			s := &StudentBilling{
				DB:                    db,
				StudentProductService: studentProductService,
				BillItemService:       billItemService,
				OrderService:          orderService,
			}

			resp, err := s.RetrieveListOfBillItems(testCase.Ctx, testCase.Req.(*pb.RetrieveListOfBillItemsRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				billItemService,
				studentProductService,
				orderService,
			)
		})
	}
}
