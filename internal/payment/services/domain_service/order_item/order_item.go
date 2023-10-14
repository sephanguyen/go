package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/google/uuid"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderItemService struct {
	orderItemRepo interface {
		Create(ctx context.Context, db database.QueryExecer, orderItem entities.OrderItem) (err error)
		CountOrderItemsByOrderID(ctx context.Context, db database.QueryExecer, orderID string) (int, error)
		GetOrderItemsByOrderIDWithPaging(
			ctx context.Context, db database.QueryExecer, orderID string, offset int64, limit int64) (
			[]entities.OrderItem, error)
		GetOrderItemsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) ([]entities.OrderItem, error)
		GetOrderItemsByProductIDs(ctx context.Context, db database.QueryExecer, productIDs []string) ([]entities.OrderItem, error)
	}
}

func (s *OrderItemService) CountOrderItemsByOrderID(ctx context.Context, db database.Ext, orderID string) (count int, err error) {
	count, err = s.orderItemRepo.CountOrderItemsByOrderID(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get count order items by order id: %v", err.Error())
		return
	}
	return
}

func (s *OrderItemService) GetOrderItemsByOrderIDWithPaging(ctx context.Context, db database.Ext, orderID string, from int64, limit int64) (orderItems []entities.OrderItem, err error) {
	orderItems, err = s.orderItemRepo.GetOrderItemsByOrderIDWithPaging(ctx, db, orderID, from, limit)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get order items by order id with paging: %v", err.Error())
		return
	}
	return
}

func (s *OrderItemService) CreateOrderItem(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	orderItem entities.OrderItem,
	err error,
) {
	var orderType string

	orderType = orderItemData.Order.OrderType.String
	switch orderType {
	case pb.OrderType_ORDER_TYPE_NEW.String(),
		pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
		pb.OrderType_ORDER_TYPE_RESUME.String():
		orderItem, err = s.createOrderItemForCreate(ctx, db, orderItemData)
	case pb.OrderType_ORDER_TYPE_UPDATE.String(),
		pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
		pb.OrderType_ORDER_TYPE_GRADUATE.String(),
		pb.OrderType_ORDER_TYPE_LOA.String():
		orderItem, err = s.createOrderItemForUpdate(ctx, db, orderItemData)
	default:
		err = status.Errorf(codes.InvalidArgument, fmt.Sprintf("error when creating order item with invalid type %s", orderType))
	}
	return
}

func (s *OrderItemService) GetOrderItemsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) (orderItems []entities.OrderItem, err error) {
	orderItems, err = s.orderItemRepo.GetOrderItemsByOrderIDs(ctx, db, orderIDs)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get order items by order ids: %v", err.Error())
		return
	}
	return
}

func (s *OrderItemService) GetOrderItemsByProductIDs(ctx context.Context, db database.QueryExecer, productIDs []string) (orderItems []entities.OrderItem, err error) {
	orderItems, err = s.orderItemRepo.GetOrderItemsByProductIDs(ctx, db, productIDs)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get order items by product ids: %v", err.Error())
		return
	}
	return
}

func (s *OrderItemService) CreateMultiCustomOrderItem(
	ctx context.Context,
	db database.QueryExecer,
	req *pb.CreateCustomBillingRequest,
	order entities.Order,
) (
	orderItems []entities.OrderItem,
	err error,
) {
	for _, item := range req.CustomBillingItems {
		if len(strings.TrimSpace(item.Name)) == 0 {
			err = status.Errorf(codes.FailedPrecondition, "Missing mandatory data: custom billing item name")
			return
		}
		orderItem := &entities.OrderItem{}
		database.AllNullEntity(orderItem)
		err = multierr.Combine(
			orderItem.OrderID.Set(order.OrderID.String),
			orderItem.OrderItemID.Set(uuid.NewString()),
			orderItem.ProductName.Set(item.Name),
		)
		if err != nil {
			return
		}

		err = s.orderItemRepo.Create(ctx, db, *orderItem)
		if err != nil {
			err = status.Errorf(codes.Internal, "creating custom order item have error: %v", err.Error())
			return
		}
		orderItems = append(orderItems, *orderItem)
	}
	return
}

func (s *OrderItemService) createOrderItemForCreate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	orderItem entities.OrderItem,
	err error,
) {
	if orderItemData.OrderItem.CancellationDate != nil {
		err = orderItem.CancellationDate.Set(orderItemData.OrderItem.CancellationDate.AsTime())
		if err != nil {
			return
		}
	} else {
		err = orderItem.CancellationDate.Set(nil)
		if err != nil {
			return
		}
	}
	err = utils.GroupErrorFunc(
		multierr.Combine(
			orderItem.OrderID.Set(orderItemData.Order.OrderID.String),
			orderItem.ProductID.Set(orderItemData.OrderItem.ProductId),
			orderItem.ProductName.Set(orderItemData.ProductInfo.Name.String),
			orderItem.StudentProductID.Set(orderItemData.StudentProduct.StudentProductID.String),
			orderItem.OrderItemID.Set(uuid.NewString()),
			orderItem.EffectiveDate.Set(nil),
			orderItem.StartDate.Set(nil),
			orderItem.EndDate.Set(nil),
			orderItem.DiscountID.Set(nil),
		),

		checkStartDateAndAddStartDateInOrderItem(orderItemData, &orderItem),
		checkDiscountAndAddDiscountInOrderItem(orderItemData, &orderItem),

		s.orderItemRepo.Create(ctx, db, orderItem),
	)
	return
}

func (s *OrderItemService) createOrderItemForUpdate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	orderItem entities.OrderItem,
	err error,
) {
	if orderItemData.OrderItem.CancellationDate != nil {
		err = orderItem.CancellationDate.Set(orderItemData.OrderItem.CancellationDate.AsTime())
		if err != nil {
			return
		}
	} else {
		err = orderItem.CancellationDate.Set(nil)
		if err != nil {
			return
		}
	}

	err = utils.GroupErrorFunc(
		multierr.Combine(
			orderItem.OrderID.Set(orderItemData.Order.OrderID.String),
			orderItem.ProductID.Set(orderItemData.OrderItem.ProductId),
			orderItem.ProductName.Set(orderItemData.ProductInfo.Name.String),
			orderItem.StudentProductID.Set(orderItemData.StudentProduct.StudentProductID.String),
			orderItem.OrderItemID.Set(uuid.NewString()),
			orderItem.EffectiveDate.Set(nil),
			orderItem.StartDate.Set(nil),
			orderItem.EndDate.Set(nil),
			orderItem.DiscountID.Set(nil),
		),

		checkDatesOfOrderItemForUpdate(orderItemData, &orderItem),
		checkDiscountAndAddDiscountInOrderItem(orderItemData, &orderItem),

		s.orderItemRepo.Create(ctx, db, orderItem),
	)

	return
}

func checkDatesOfOrderItemForUpdate(orderItemData utils.OrderItemData, orderItem *entities.OrderItem) (err error) {
	if orderItemData.Order.OrderType.String == pb.OrderType_ORDER_TYPE_LOA.String() {
		err = checkLOADurationInOrderItem(orderItemData, orderItem)
	} else {
		err = checkEffectiveDateAndAddEffectiveDateInOrderItem(orderItemData, orderItem)
	}

	return
}

func checkStartDateAndAddStartDateInOrderItem(orderItemData utils.OrderItemData, orderItem *entities.OrderItem) (err error) {
	if orderItemData.IsOneTimeProduct {
		return
	}

	if orderItemData.OrderItem.StartDate == nil {
		err = status.Errorf(codes.FailedPrecondition, "create recurring order item for create order without start date")
		return
	}
	err = orderItem.StartDate.Set(orderItemData.OrderItem.StartDate.AsTime())
	return
}

func checkEffectiveDateAndAddEffectiveDateInOrderItem(orderItemData utils.OrderItemData, orderItem *entities.OrderItem) (err error) {
	if orderItemData.IsOneTimeProduct {
		return
	}

	if orderItemData.OrderItem.EffectiveDate == nil {
		err = status.Errorf(codes.FailedPrecondition, "create recurring order item for create order without start date")
		return
	}
	err = orderItem.EffectiveDate.Set(orderItemData.OrderItem.EffectiveDate.AsTime())
	return
}

func checkLOADurationInOrderItem(orderItemData utils.OrderItemData, orderItem *entities.OrderItem) (err error) {
	if orderItemData.IsOneTimeProduct {
		return
	}

	if orderItemData.OrderItem.StartDate == nil {
		err = status.Errorf(codes.FailedPrecondition, "missing start date of LOA")
		return
	}

	if orderItemData.OrderItem.EndDate == nil {
		err = status.Errorf(codes.FailedPrecondition, "missing end date of LOA")
		return
	}

	// Subtract 1 day from the duration calculation to account for the LOA's effective date being the following day.
	if orderItemData.Order.LOAStartDate.Time.AddDate(0, 0, -1) != orderItemData.OrderItem.StartDate.AsTime() {
		err = status.Errorf(codes.FailedPrecondition, fmt.Sprintf("start_date of order and order item is inconsistency with product_id=%s", orderItem.ProductID.String))
		return
	}

	if orderItemData.Order.LOAEndDate.Time != orderItemData.OrderItem.EndDate.AsTime() {
		err = status.Errorf(codes.FailedPrecondition, fmt.Sprintf("end_date of order and order item is inconsistency with product_id=%s", orderItem.ProductID.String))
		return
	}

	err = utils.GroupErrorFunc(
		multierr.Combine(
			orderItem.StartDate.Set(orderItemData.OrderItem.StartDate.AsTime()),
			orderItem.EndDate.Set(orderItemData.OrderItem.EndDate.AsTime()),
		),
	)

	return
}

func checkDiscountAndAddDiscountInOrderItem(orderItemData utils.OrderItemData, orderItem *entities.OrderItem) (err error) {
	if orderItemData.OrderItem.DiscountId == nil {
		return
	}

	err = orderItem.DiscountID.Set(orderItemData.OrderItem.DiscountId.Value)
	return
}

func NewOrderItemService() *OrderItemService {
	return &OrderItemService{
		orderItemRepo: &repositories.OrderItemRepo{},
	}
}
