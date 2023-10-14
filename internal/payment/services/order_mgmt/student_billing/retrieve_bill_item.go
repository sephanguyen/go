package studentbilling

import (
	"context"

	"github.com/manabie-com/backend/internal/payment/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *StudentBilling) RetrieveListOfBillItems(
	ctx context.Context,
	req *pb.RetrieveListOfBillItemsRequest,
) (
	res *pb.RetrieveListOfBillItemsResponse,
	err error,
) {
	var (
		from                    int64
		limit                   int64
		total                   int
		billingDescriptions     []utils.BillItemForRetrieveApi
		billingItems            []*pb.RetrieveListOfBillItemsResponse_BillItems
		prevPage                *cpb.Paging
		nextPage                *cpb.Paging
		mapOrderIDWithOrderType map[string]string
	)
	from, limit, err = utils.PagingToFromAndLimit(req.Paging)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid paging data with error: %s", err.Error())
		return
	}
	billingItems = make([]*pb.RetrieveListOfBillItemsResponse_BillItems, 0, limit)
	mapOrderIDWithOrderType = make(map[string]string, limit)
	billingDescriptions, total, err = s.BillItemService.GetBillItemDescriptionByStudentIDAndLocationIDs(ctx, s.DB, req.StudentId, req.LocationIds, from, limit)
	if err != nil {
		return
	}
	count := int32(from)
	for i, description := range billingDescriptions {
		var (
			floatAmount     float32
			adjustmentPrice float32
			orderType       string
			ok              bool
			// billingDate     *timestamppb.Timestamp
		)
		count++
		_ = description.BillItemEntity.FinalPrice.AssignTo(&floatAmount)
		if description.BillItemEntity.AdjustmentPrice.Status == pgtype.Present {
			_ = description.BillItemEntity.AdjustmentPrice.AssignTo(&adjustmentPrice)
		}
		orderType, ok = mapOrderIDWithOrderType[description.BillItemEntity.OrderID.String]
		if !ok {
			orderType, err = s.OrderService.GetOrderTypeByOrderID(ctx, s.DB, description.BillItemEntity.OrderID.String)
			if err != nil {
				return
			}
			mapOrderIDWithOrderType[description.BillItemEntity.OrderID.String] = orderType
		}

		// if description.BillItemEntity.BillDate.Time.Before(time.Now()) {
		// 	billingDate = timestamppb.New(description.BillItemEntity.CreatedAt.Time)
		// } else {
		// 	billingDate = timestamppb.New(description.BillItemEntity.BillDate.Time)
		// }

		billingItems = append(billingItems, &pb.RetrieveListOfBillItemsResponse_BillItems{
			Index:               count,
			OrderId:             description.BillItemEntity.OrderID.String,
			BillingStatus:       pb.BillingStatus(pb.BillingStatus_value[description.BillItemEntity.BillStatus.String]),
			BillingDate:         timestamppb.New(description.BillItemEntity.BillDate.Time),
			Amount:              floatAmount,
			BillItemDescription: billingDescriptions[i].BillItemDescription,
			BillingNo:           description.BillItemEntity.BillItemSequenceNumber.Int,
			BillingType:         pb.BillingType(pb.BillingType_value[description.BillItemEntity.BillType.String]),
			AdjustmentPrice:     wrapperspb.Float(adjustmentPrice),
			LocationInfo: &pb.LocationInfo{
				LocationId:   description.BillItemEntity.LocationID.String,
				LocationName: description.BillItemEntity.LocationName.String,
			},
			OrderType: pb.OrderType(pb.OrderType_value[orderType]),
		})
	}

	prevPage, nextPage, err = utils.ConvertCommonPaging(total, from, limit)
	if err != nil {
		return
	}

	res = &pb.RetrieveListOfBillItemsResponse{
		NextPage:     nextPage,
		PreviousPage: prevPage,
		TotalItems:   uint32(total),
		Items:        billingItems,
	}
	return
}
