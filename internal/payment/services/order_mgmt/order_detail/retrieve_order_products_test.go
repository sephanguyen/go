package order_detail

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

func TestRetrieveOrderProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.IStudentProductForOrderDetail
		billItemService       *mockServices.IBillItemServiceForOrderDetail
		orderItemService      *mockServices.IOrderItemServiceForOrderDetail
		orderService          *mockServices.IOrderServiceForOrderDetail
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: with nil paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: fmt.Errorf("invalid paging data with error"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when get order type by order id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return("", constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when count order items by order id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(0, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get order items by order id with paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				orderItemService.On("GetOrderItemsByOrderIDWithPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when build map bill item with product id by order id and product ids",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				orderItemService.On("GetOrderItemsByOrderIDWithPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{}, nil)
				billItemService.On("BuildMapBillItemWithProductIDByOrderIDAndProductIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student products by student product ids",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				orderItemService.On("GetOrderItemsByOrderIDWithPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{}, nil)
				billItemService.On("BuildMapBillItemWithProductIDByOrderIDAndProductIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				studentProductService.On("GetStudentProductsByStudentProductIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when missing student_product_id in order_item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: status.Error(codes.Internal, "Error when missing student_product_id in order_item"),
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				orderItemService.On("GetOrderItemsByOrderIDWithPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						OrderItemID: pgtype.Text{
							String: constant.OrderItemID,
						},
						DiscountID: pgtype.Text{
							String: constant.DiscountID,
						},
						StartDate: pgtype.Timestamptz{
							Time: time.Now(),
						},
						StudentProductID: pgtype.Text{
							String: constant.StudentProductID,
							Status: pgtype.Null,
						},
						ProductName: pgtype.Text{
							String: constant.ProductName,
						},
					},
				}, nil)
				billItemService.On("BuildMapBillItemWithProductIDByOrderIDAndProductIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				studentProductService.On("GetStudentProductsByStudentProductIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
		{
			Name: "Fail case: Error when missing product_id in order_item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: status.Error(codes.Internal, "Error when missing product_id in order_item"),
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				orderItemService.On("GetOrderItemsByOrderIDWithPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Null,
						},
						OrderItemID: pgtype.Text{
							String: constant.OrderItemID,
						},
						DiscountID: pgtype.Text{
							String: constant.DiscountID,
						},
						StartDate: pgtype.Timestamptz{
							Time: time.Now(),
						},
						StudentProductID: pgtype.Text{
							String: constant.StudentProductID,
							Status: pgtype.Present,
						},
						ProductName: pgtype.Text{
							String: constant.ProductName,
						},
					},
				}, nil)
				billItemService.On("BuildMapBillItemWithProductIDByOrderIDAndProductIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				studentProductService.On("GetStudentProductsByStudentProductIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
		{
			Name: "Fail case: Error when convert common paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 20},
				},
			},
			ExpectedErr: status.Error(codes.Internal, "Error offset"),
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				orderItemService.On("GetOrderItemsByOrderIDWithPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
						OrderItemID: pgtype.Text{
							String: constant.OrderItemID,
						},
						DiscountID: pgtype.Text{
							String: constant.DiscountID,
						},
						StartDate: pgtype.Timestamptz{
							Time: time.Now(),
						},
						StudentProductID: pgtype.Text{
							String: constant.StudentProductID,
							Status: pgtype.Present,
						},
						ProductName: pgtype.Text{
							String: constant.ProductName,
						},
					},
				}, nil)
				billItemService.On("BuildMapBillItemWithProductIDByOrderIDAndProductIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				studentProductService.On("GetStudentProductsByStudentProductIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderDetailProductsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderTypeByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(pb.OrderType_ORDER_TYPE_NEW.String(), nil)
				orderItemService.On("CountOrderItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				orderItemService.On("GetOrderItemsByOrderIDWithPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
						OrderItemID: pgtype.Text{
							String: constant.OrderItemID,
						},
						DiscountID: pgtype.Text{
							String: constant.DiscountID,
						},
						StartDate: pgtype.Timestamptz{
							Time: time.Now(),
						},
						StudentProductID: pgtype.Text{
							String: constant.StudentProductID,
							Status: pgtype.Present,
						},
						ProductName: pgtype.Text{
							String: constant.ProductName,
						},
					},
				}, nil)
				billItemService.On("BuildMapBillItemWithProductIDByOrderIDAndProductIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				studentProductService.On("GetStudentProductsByStudentProductIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.IStudentProductForOrderDetail)
			billItemService = new(mockServices.IBillItemServiceForOrderDetail)
			orderService = new(mockServices.IOrderServiceForOrderDetail)
			orderItemService = new(mockServices.IOrderItemServiceForOrderDetail)
			testCase.Setup(testCase.Ctx)
			s := &OrderDetail{
				DB:                    db,
				StudentProductService: studentProductService,
				BillItemService:       billItemService,
				OrderService:          orderService,
				OrderItemService:      orderItemService,
			}

			resp, err := s.RetrieveProductsOfOrder(testCase.Ctx, testCase.Req.(*pb.RetrieveListOfOrderDetailProductsRequest))

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
				orderItemService,
				billItemService,
			)
		})
	}
}
