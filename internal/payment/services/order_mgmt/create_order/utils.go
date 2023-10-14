package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapOrderItemData(
	req *pb.CreateOrderRequest,
	order entities.Order,
	studentInfo entities.Student,
	locationName string,
	studentName string,
) (
	mapKeyWithOrderItemData map[string]utils.OrderItemData,
	typeOfOrder *utils.OrderType,
	err error,
) {
	mapKeyWithOrderItemData = make(map[string]utils.OrderItemData, len(req.OrderItems))
	for i, item := range req.OrderItems {
		var key string
		key, err = utils.GetKeyMapFromOrderItemV2(item, req.OrderType)
		if err != nil {
			return
		}

		if _, ok := mapKeyWithOrderItemData[key]; ok {
			err = status.Errorf(codes.FailedPrecondition, "Product with id (or \"product id\"_\"package associated\") %v was duplicated in your order request", key)
			return
		}
		mapKeyWithOrderItemData[key] = utils.OrderItemData{
			Order:        order,
			StudentInfo:  studentInfo,
			LocationName: locationName,
			StudentName:  studentName,
			OrderItem:    req.OrderItems[i],
			Timezone:     req.Timezone,
		}
	}

	for i, item := range req.BillingItems {
		var key string
		key, err = utils.GetKeyMapFromBillItemV2(item, req.OrderType)
		if err != nil {
			return
		}

		orderItem, ok := mapKeyWithOrderItemData[key]
		if !ok {
			err = status.Errorf(codes.FailedPrecondition, "Product with id (or \"product id\"_\"package associated\") %v existed in bill item but It didn't exist in order item of your order request", key)
			return
		}
		if typeOfOrder == nil {
			orderItemType := convertOrderItemType(req.OrderType, item)
			typeOfOrder = &orderItemType
		}
		tmpBillingItemData := utils.BillingItemData{
			BillingItem: req.BillingItems[i],
			IsUpcoming:  false,
		}
		orderItem.BillItems = append(orderItem.BillItems, tmpBillingItemData)
		mapKeyWithOrderItemData[key] = orderItem
	}

	for i, item := range req.UpcomingBillingItems {
		var key string
		key, err = utils.GetKeyMapFromBillItemV2(item, req.OrderType)
		if err != nil {
			return
		}

		orderItem, ok := mapKeyWithOrderItemData[key]
		if !ok {
			err = status.Errorf(codes.FailedPrecondition, "Product with id (or \"product id\"_\"package associated\") %v existed in bill item but It didn't exist in order item of your order request", key)
			return
		}
		if typeOfOrder == nil {
			orderItemType := convertOrderItemType(req.OrderType, item)
			typeOfOrder = &orderItemType
		}
		tmpBillingItemData := utils.BillingItemData{
			BillingItem: req.UpcomingBillingItems[i],
			IsUpcoming:  true,
		}
		orderItem.BillItems = append(orderItem.BillItems, tmpBillingItemData)
		mapKeyWithOrderItemData[key] = orderItem
	}
	if typeOfOrder == nil {
		orderItemType := convertOrderItemType(req.OrderType, nil)
		typeOfOrder = &orderItemType
	}
	return
}

func convertOrderItemType(orderType pb.OrderType, billItem *pb.BillingItem) (typeOfOrder utils.OrderType) {
	switch orderType {
	case pb.OrderType_ORDER_TYPE_NEW:
		typeOfOrder = utils.OrderCreate
	case pb.OrderType_ORDER_TYPE_UPDATE:
		if billItem.IsCancelBillItem != nil && billItem.IsCancelBillItem.Value {
			typeOfOrder = utils.OrderCancel
		} else {
			typeOfOrder = utils.OrderUpdate
		}
	case pb.OrderType_ORDER_TYPE_ENROLLMENT:
		typeOfOrder = utils.OrderEnrollment
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL:
		typeOfOrder = utils.OrderWithdraw
	case pb.OrderType_ORDER_TYPE_GRADUATE:
		typeOfOrder = utils.OrderGraduate
	case pb.OrderType_ORDER_TYPE_LOA:
		typeOfOrder = utils.OrderLOA
	case pb.OrderType_ORDER_TYPE_RESUME:
		typeOfOrder = utils.OrderResume
	}
	return
}

func (s *CreateOrderService) hasEnrolledPriceByProductID(ctx context.Context, tx database.QueryExecer, productID string) (
	hasEnrolledPrice bool,
	err error,
) {
	productPrices, err := s.ProductPriceService.GetProductPricesByProductIDAndPriceType(ctx, tx, productID, pb.ProductPriceType_ENROLLED_PRICE.String())
	if err != nil {
		return
	}
	if len(productPrices) != 0 {
		hasEnrolledPrice = true
	}
	return
}
