package studentbilling

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudentBilling_RetrieveListOfOrderItems(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.IStudentProductForStudentBilling
		billItemService       *mockServices.IBillItemServiceForStudentBilling
		orderService          *mockServices.IOrderServiceForStudentBilling
		locationService       *mockServices.ILocationServiceForStudentBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: with nil paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderItemsRequest{
				StudentId: "1",
			},
			ExpectedErr: fmt.Errorf("invalid paging data with error"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: get order id by student id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrdersByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.Order{{}}, 0, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: get bill item id by student id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrdersByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.Order{{}}, 0, nil)
				billItemService.On("GetBillItemInfoByOrderIDAndUniqueByProductID",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.BillItem{{}}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: getting paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: status.Error(codes.Internal, "Error offset"),
			Setup: func(ctx context.Context) {
				orderService.On("GetOrdersByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.Order{{}}, 0, nil)
				billItemService.On("GetBillItemInfoByOrderIDAndUniqueByProductID",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.BillItem{{}}, nil)
			},
		},
		{
			Name: "Fail case: when get location",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrdersByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.Order{{}}, 20, nil)
				billItemService.On("GetBillItemInfoByOrderIDAndUniqueByProductID",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.BillItem{}, nil)
				locationService.On("GetLocationInfoByID",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return(&pb.LocationInfo{}, constant.ErrDefault)
			},
		},
		{
			Name: "happy case when withdraw have product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrdersByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.Order{{}}, 20, nil)
				billItemService.On("GetBillItemInfoByOrderIDAndUniqueByProductID",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.BillItem{{}}, nil)
			},
		},
		{
			Name: "happy case when withdraw have product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderItemsRequest{
				StudentId: "1",
				Paging:    &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrdersByStudentIDAndLocationIDs",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.Order{{}}, 20, nil)
				billItemService.On("GetBillItemInfoByOrderIDAndUniqueByProductID",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return([]*entities.BillItem{}, nil)
				locationService.On("GetLocationInfoByID",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return(&pb.LocationInfo{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.IStudentProductForStudentBilling)
			billItemService = new(mockServices.IBillItemServiceForStudentBilling)
			orderService = new(mockServices.IOrderServiceForStudentBilling)
			locationService = new(mockServices.ILocationServiceForStudentBilling)
			testCase.Setup(testCase.Ctx)
			s := &StudentBilling{
				DB:                    db,
				StudentProductService: studentProductService,
				BillItemService:       billItemService,
				OrderService:          orderService,
				LocationService:       locationService,
			}

			resp, err := s.RetrieveListOfOrderItems(testCase.Ctx, testCase.Req.(*pb.RetrieveListOfOrderItemsRequest))
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
				locationService,
			)
		})
	}
}
