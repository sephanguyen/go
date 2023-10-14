package studentbilling

import (
	"context"

	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StudentBilling) RetrieveListOfOrderItems(
	ctx context.Context,
	req *pb.RetrieveListOfOrderItemsRequest,
) (
	res *pb.RetrieveListOfOrderItemsResponse,
	err error,
) {
	var (
		from       int64
		limit      int64
		total      int
		orders     []*entities.Order
		orderItems []*pb.RetrieveListOfOrderItemsResponse_OrderItems
		prevPage   *cpb.Paging
		nextPage   *cpb.Paging
	)
	from, limit, err = utils.PagingToFromAndLimit(req.Paging)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid paging data with error: %s", err.Error())
		return
	}
	orders = make([]*entities.Order, 0, limit)
	orderItems = make([]*pb.RetrieveListOfOrderItemsResponse_OrderItems, 0, limit)
	orders, total, err = s.OrderService.GetOrdersByStudentIDAndLocationIDs(ctx, s.DB, req.StudentId, req.LocationIds, from, limit)
	if err != nil {
		return
	}
	count := int32(from)
	for _, order := range orders {
		var (
			orderItem *pb.RetrieveListOfOrderItemsResponse_OrderItems
			billItems []*entities.BillItem
		)
		count++
		orderItem = &pb.RetrieveListOfOrderItemsResponse_OrderItems{
			Index:       count,
			OrderNo:     order.OrderSequenceNumber.Int,
			OrderId:     order.OrderID.String,
			OrderStatus: pb.OrderStatus(pb.OrderStatus_value[order.OrderStatus.String]),
			OrderType:   pb.OrderType(pb.OrderType_value[order.OrderType.String]),
			CreateDate:  timestamppb.New(order.CreatedAt.Time),
		}

		billItems, err = s.BillItemService.GetBillItemInfoByOrderIDAndUniqueByProductID(ctx, s.DB, order.OrderID.String)
		if err != nil {
			return
		}
		if len(billItems) == 0 {
			orderItem.LocationInfo, err = s.LocationService.GetLocationInfoByID(ctx, s.DB, order.LocationID.String)
			if err != nil {
				return
			}
		} else {
			orderItem.ProductDetails, orderItem.LocationInfo, err = utils.ConvertEntityBillItemToProtoProductInfoAndLocationInfo(billItems)
			if err != nil {
				return
			}
		}

		orderItems = append(orderItems, orderItem)
	}

	prevPage, nextPage, err = utils.ConvertCommonPaging(total, from, limit)
	if err != nil {
		return
	}

	res = &pb.RetrieveListOfOrderItemsResponse{
		NextPage:     nextPage,
		PreviousPage: prevPage,
		TotalItems:   uint32(total),
		Items:        orderItems,
	}
	return
}
