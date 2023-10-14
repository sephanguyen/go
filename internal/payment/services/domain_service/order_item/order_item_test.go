package service

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
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	FailCaseCreateOrderItemError = "Fail case: Error when create order item"
)

func TestOrderItemService_checkStartDateAndAddStartDateInOrderItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when order item is recurring product without start date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "create recurring order item for create order without start date"),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: false, OrderItem: &pb.OrderItem{StartDate: nil}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: Order item is one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: true},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: false, OrderItem: &pb.OrderItem{StartDate: timestamppb.Now()}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			orderItem := testCase.Req.([]interface{})[1].(*entities.OrderItem)
			err := checkStartDateAndAddStartDateInOrderItem(orderItemData, orderItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderItemService_checkEffectiveDateAndAddEffectiveDateInOrderItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when order item is recurring product without effective date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "create recurring order item for create order without start date"),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: false, OrderItem: &pb.OrderItem{EffectiveDate: nil}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: Order item is one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: true},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: false, OrderItem: &pb.OrderItem{EffectiveDate: timestamppb.Now()}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			orderItem := testCase.Req.([]interface{})[1].(*entities.OrderItem)
			err := checkEffectiveDateAndAddEffectiveDateInOrderItem(orderItemData, orderItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderItemService_checkLOADurationInOrderItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	now := time.Now()
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when missing start date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing start date of LOA"),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: false,
					OrderItem: &pb.OrderItem{
						StartDate: nil,
					}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when missing end_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing end date of LOA"),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: false, OrderItem: &pb.OrderItem{
					StartDate: timestamppb.Now(),
				}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when start_date of order and order item is inconsistency",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, fmt.Sprintf("start_date of order and order item is inconsistency with product_id")),
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{
						LOAEndDate: pgtype.Timestamptz{Time: now.AddDate(0, 0, 10)},
						LOAStartDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 0, 1),
							Status: pgtype.Present},
					},
					IsOneTimeProduct: false,
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when end_date of order and order item is inconsistency",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, fmt.Sprintf("end_date of order and order item is inconsistency with product_id")),
			Req: []interface{}{
				utils.OrderItemData{
					Timezone: 7,
					Order: entities.Order{
						LOAStartDate: pgtype.Timestamptz{
							Time:   now.UTC().AddDate(0, 0, 1),
							Status: pgtype.Present,
						},
						LOAEndDate: pgtype.Timestamptz{
							Time:   now.UTC().AddDate(0, 0, 10),
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now.AddDate(0, 0, 9)),
					}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: Order item is one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{IsOneTimeProduct: true},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{
					Timezone: 7,
					Order: entities.Order{
						LOAStartDate: pgtype.Timestamptz{
							Time:   now.UTC().AddDate(0, 0, 1),
							Status: pgtype.Present,
						},
						LOAEndDate: pgtype.Timestamptz{
							Time:   now.UTC().AddDate(0, 0, 10),
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now.AddDate(0, 0, 10)),
					}},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			orderItem := testCase.Req.([]interface{})[1].(*entities.OrderItem)
			err := checkLOADurationInOrderItem(orderItemData, orderItem)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderItemService_checkDiscountAndAddDiscountInOrderItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	testcases := []utils.TestCase{
		{
			Name: "Happy case: Discount is nil",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: true,
					OrderItem:        &pb.OrderItem{DiscountId: nil},
				},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: true,
					OrderItem:        &pb.OrderItem{DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID}},
				},
				&entities.OrderItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			orderItem := testCase.Req.([]interface{})[1].(*entities.OrderItem)
			err := checkDiscountAndAddDiscountInOrderItem(orderItemData, orderItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderItemService_createOrderItemForUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseCreateOrderItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderID: pgtype.Text{String: constant.OrderID},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:  constant.StudentName,
				LocationName: constant.LocationName,
				OrderItem:    &pb.OrderItem{EffectiveDate: timestamppb.Now()},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderID: pgtype.Text{String: constant.OrderID},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:  constant.StudentName,
				LocationName: constant.LocationName,
				OrderItem:    &pb.OrderItem{EffectiveDate: timestamppb.Now()},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderItemService{
				orderItemRepo: orderItemRepo,
			}
			req := testCase.Req.(utils.OrderItemData)
			_, err := s.createOrderItemForUpdate(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderItemRepo)
		})
	}
}

func TestOrderItemService_createOrderItemForCreate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseCreateOrderItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderID: pgtype.Text{String: constant.OrderID},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:      constant.StudentName,
				LocationName:     constant.LocationName,
				OrderItem:        &pb.OrderItem{EffectiveDate: timestamppb.Now(), ProductId: constant.ProductID},
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderID: pgtype.Text{String: constant.OrderID},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:      constant.StudentName,
				LocationName:     constant.LocationName,
				OrderItem:        &pb.OrderItem{EffectiveDate: timestamppb.Now(), ProductId: constant.ProductID},
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderItemService{
				orderItemRepo: orderItemRepo,
			}
			req := testCase.Req.(utils.OrderItemData)
			_, err := s.createOrderItemForCreate(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderItemRepo)
		})
	}
}

func TestOrderItemService_CreateMultiCustomOrderItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when missing mandatory data (custom billing item name)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Missing mandatory data: custom billing item name"),
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					StudentId:    constant.StudentID,
					LocationId:   constant.LocationID,
					OrderComment: constant.OrderComment,
					CustomBillingItems: []*pb.CustomBillingItem{
						{
							Name:  "",
							Price: constant.DefaultPrice,
						},
					},
					OrderType: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
				},
				entities.Order{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseCreateOrderItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating custom order item have error: %v", constant.ErrDefault),
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					StudentId:    constant.StudentID,
					LocationId:   constant.LocationID,
					OrderComment: constant.OrderComment,
					CustomBillingItems: []*pb.CustomBillingItem{
						{
							Name:  constant.ProductName,
							Price: constant.DefaultPrice,
						},
					},
					OrderType: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
				},
				entities.Order{},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", ctx, db, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					StudentId:    constant.StudentID,
					LocationId:   constant.LocationID,
					OrderComment: constant.OrderComment,
					CustomBillingItems: []*pb.CustomBillingItem{
						{
							Name:  constant.ProductName,
							Price: constant.DefaultPrice,
						},
					},
					OrderType: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
				},
				entities.Order{},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", ctx, db, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderItemService{
				orderItemRepo: orderItemRepo,
			}
			req := testCase.Req.([]interface{})[0].(*pb.CreateCustomBillingRequest)
			order := testCase.Req.([]interface{})[1].(entities.Order)
			orderItems, err := s.CreateMultiCustomOrderItem(testCase.Ctx, db, req, order)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, len(req.CustomBillingItems), len(orderItems))
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderItemRepo)
		})
	}
}

func TestOrderItemService_CreateOrderItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when order type is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.InvalidArgument, "error when creating order item with invalid type"),
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderType: pgtype.Text{
						String: "",
					},
					OrderID: pgtype.Text{String: constant.OrderID},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:      constant.StudentName,
				LocationName:     constant.LocationName,
				OrderItem:        &pb.OrderItem{EffectiveDate: timestamppb.Now(), ProductId: constant.ProductID},
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when create order items for create",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
					OrderID: pgtype.Text{String: constant.OrderID},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:      constant.StudentName,
				LocationName:     constant.LocationName,
				OrderItem:        &pb.OrderItem{EffectiveDate: timestamppb.Now(), ProductId: constant.ProductID},
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: Create order items for create successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
					OrderID: pgtype.Text{String: constant.OrderID},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:      constant.StudentName,
				LocationName:     constant.LocationName,
				OrderItem:        &pb.OrderItem{EffectiveDate: timestamppb.Now(), ProductId: constant.ProductID},
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when create order items for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderID:   pgtype.Text{String: constant.OrderID},
					OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:  constant.StudentName,
				LocationName: constant.LocationName,
				OrderItem:    &pb.OrderItem{EffectiveDate: timestamppb.Now()},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: Create order items for update successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				Order: entities.Order{
					OrderID:   pgtype.Text{String: constant.OrderID},
					OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{String: constant.StudentID},
				},
				PackageInfo: utils.PackageInfo{},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{String: constant.StudentProductID},
				},
				StudentName:  constant.StudentName,
				LocationName: constant.LocationName,
				OrderItem:    &pb.OrderItem{EffectiveDate: timestamppb.Now()},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderItemService{
				orderItemRepo: orderItemRepo,
			}
			req := testCase.Req.(utils.OrderItemData)
			_, err := s.CreateOrderItem(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderItemRepo)
		})
	}
}

func TestOrderItemService_GetOrderItemsByOrderIDWithPaging(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)
	const offset int64 = 1
	const limit int64 = 1
	expectedResp := []entities.OrderItem{
		{
			OrderID: pgtype.Text{
				String: constant.OrderID,
				Status: pgtype.Present,
			},
			ProductID: pgtype.Text{
				String: constant.ProductID,
				Status: pgtype.Present,
			},
			OrderItemID: pgtype.Text{
				String: "1",
				Status: pgtype.Present,
			},
		},
	}
	testcases := []utils.TestCase{
		{
			Name:         "Success case: Happy case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:          nil,
			ExpectedErr:  nil,
			ExpectedResp: expectedResp,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemsByOrderIDWithPaging", ctx, mock.Anything, constant.OrderID, offset, limit).
					Return(expectedResp, nil)
			},
		},
		{
			Name:        "Failed case: Error when get order items by order id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         &entities.OrderItem{},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemsByOrderIDWithPaging", ctx, mock.Anything, constant.OrderID, offset, limit).
					Return(nil, constant.ErrDefault)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderItemService{
				orderItemRepo: orderItemRepo,
			}
			_, err := s.GetOrderItemsByOrderIDWithPaging(testCase.Ctx, db, constant.OrderID, offset, limit)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, expectedResp)
			}

		})
	}
}

func TestOrderItemService_GetOrderItemsByOrderIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when getting order items by order ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         []string{"1", "2"},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemsByOrderIDs", ctx, mock.Anything, mock.Anything).Return([]entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Success case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         []string{"1", "2"},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemsByOrderIDs", ctx, mock.Anything, mock.Anything).Return([]entities.OrderItem{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderItemService{
				orderItemRepo: orderItemRepo,
			}
			req := testCase.Req.([]string)
			_, err := s.GetOrderItemsByOrderIDs(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderItemRepo)
		})
	}
}

func TestOrderItemService_GetOrderItemsByProductIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when getting order items by product ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         []string{"1", "2"},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemsByProductIDs", ctx, mock.Anything, mock.Anything).Return([]entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Success case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         []string{"1", "2"},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemsByProductIDs", ctx, mock.Anything, mock.Anything).Return([]entities.OrderItem{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderItemService{
				orderItemRepo: orderItemRepo,
			}
			req := testCase.Req.([]string)
			_, err := s.GetOrderItemsByProductIDs(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderItemRepo)
		})
	}
}
