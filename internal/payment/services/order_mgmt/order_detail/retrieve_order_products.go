package order_detail

import (
	"context"

	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *OrderDetail) RetrieveProductsOfOrder(
	ctx context.Context,
	req *pb.RetrieveListOfOrderDetailProductsRequest,
) (
	res *pb.RetrieveListOfOrderDetailProductsResponse,
	err error,
) {
	var (
		from      int64
		limit     int64
		orderType string
	)
	from, limit, err = utils.PagingToFromAndLimit(req.Paging)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid paging data with error: %s", err.Error())
		return
	}

	orderType, err = s.OrderService.GetOrderTypeByOrderID(ctx, s.DB, req.OrderId)
	if err != nil {
		return
	}

	orderItemsCount, err := s.OrderItemService.CountOrderItemsByOrderID(ctx, s.DB, req.OrderId)
	if err != nil {
		return
	}

	orderItems, err := s.OrderItemService.GetOrderItemsByOrderIDWithPaging(ctx, s.DB, req.OrderId, from, limit)
	if err != nil {
		return
	}

	mapProductIDWithOrderItem := make(map[string]entities.OrderItem, len(orderItems))
	studentProductIDs := make([]string, 0, len(orderItems))
	productIDs := make([]string, 0, len(orderItems))
	for _, orderItem := range orderItems {
		studentProductIDs = append(studentProductIDs, orderItem.StudentProductID.String)
		productIDs = append(productIDs, orderItem.ProductID.String)
		mapProductIDWithOrderItem[orderItem.ProductID.String] = orderItem
	}

	mapProductIDWithBillItem, err := s.BillItemService.BuildMapBillItemWithProductIDByOrderIDAndProductIDs(ctx, s.DB, req.OrderId, productIDs)
	if err != nil {
		return
	}

	mapStudentProductIDWithStudentProduct := make(map[string]entities.StudentProduct, len(orderItems))
	studentProducts, err := s.StudentProductService.GetStudentProductsByStudentProductIDs(ctx, s.DB, studentProductIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student products by student product ids: %v", err.Error())
		return nil, err
	}
	for _, studentProduct := range studentProducts {
		mapStudentProductIDWithStudentProduct[studentProduct.StudentProductID.String] = studentProduct
	}

	count := int32(from)
	productInfos := make([]*pb.RetrieveListOfOrderDetailProductsResponse_OrderProduct, 0, len(productIDs))
	for _, orderItem := range orderItems {
		count++

		var productInfo *pb.RetrieveListOfOrderDetailProductsResponse_OrderProduct
		if orderItem.StudentProductID.Status != pgtype.Present {
			return nil, status.Error(codes.Internal, "Error when missing student_product_id in order_item")
		}
		if orderItem.ProductID.Status != pgtype.Present {
			return nil, status.Error(codes.Internal, "Error when missing product_id in order_item")
		}

		billItem := mapProductIDWithBillItem[orderItem.ProductID.String]
		studentProduct := mapStudentProductIDWithStudentProduct[orderItem.StudentProductID.String]
		productInfo, err = utils.ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(
			billItem,
			studentProduct,
			orderType,
			orderItem,
		)
		if err != nil {
			return
		}

		productInfo.Index = count
		productInfo.OrderType = pb.OrderType(pb.OrderType_value[orderType])
		productInfos = append(productInfos, productInfo)
	}

	prevPage, nextPage, err := utils.ConvertCommonPaging(orderItemsCount, from, limit)
	if err != nil {
		return nil, err
	}

	return &pb.RetrieveListOfOrderDetailProductsResponse{
		Items:        productInfos,
		NextPage:     nextPage,
		PreviousPage: prevPage,
		TotalItems:   uint32(orderItemsCount),
	}, nil
}
