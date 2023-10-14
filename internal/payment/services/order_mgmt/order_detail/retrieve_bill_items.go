package order_detail

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

func (s *OrderDetail) RetrieveBillItemsOfOrder(
	ctx context.Context,
	req *pb.RetrieveBillingOfOrderDetailsRequest,
) (
	res *pb.RetrieveBillingOfOrderDetailsResponse,
	err error,
) {
	var (
		from                int64
		limit               int64
		total               int
		billingDescriptions []utils.BillItemForRetrieveApi
		billingItems        []*pb.RetrieveBillingOfOrderDetailsResponse_OrderDetails
		prevPage            *cpb.Paging
		nextPage            *cpb.Paging
	)
	from, limit, err = utils.PagingToFromAndLimit(req.Paging)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid paging data with error: %s", err.Error())
		return
	}
	billingItems = make([]*pb.RetrieveBillingOfOrderDetailsResponse_OrderDetails, 0, limit)
	billingDescriptions, total, err = s.BillItemService.GetBillItemDescriptionsByOrderIDWithPaging(ctx, s.DB, req.OrderId, from, limit)
	if err != nil {
		return
	}
	count := int32(from)
	for _, description := range billingDescriptions {
		var floatAmount float32
		count++
		if description.BillItemEntity.AdjustmentPrice.Status == pgtype.Present {
			_ = description.BillItemEntity.AdjustmentPrice.AssignTo(&floatAmount)
		} else {
			_ = description.BillItemEntity.FinalPrice.AssignTo(&floatAmount)
		}
		billingItems = append(billingItems, &pb.RetrieveBillingOfOrderDetailsResponse_OrderDetails{
			Index:                  count,
			OrderId:                description.BillItemEntity.OrderID.String,
			BillItemSequenceNumber: description.BillItemEntity.BillItemSequenceNumber.Int,
			BillingStatus:          pb.BillingStatus(pb.BillingStatus_value[description.BillItemEntity.BillStatus.String]),
			BillingDate:            timestamppb.New(description.BillItemEntity.BillDate.Time),
			Amount:                 floatAmount,
			BillItemDescription:    description.BillItemDescription,
			ProductDescription:     wrapperspb.String(description.BillItemEntity.ProductDescription.String),
			BillingType:            pb.BillingType(pb.BillingType_value[description.BillItemEntity.BillType.String]),
		})
	}

	prevPage, nextPage, err = utils.ConvertCommonPaging(total, from, limit)
	if err != nil {
		return
	}

	res = &pb.RetrieveBillingOfOrderDetailsResponse{
		NextPage:     nextPage,
		PreviousPage: prevPage,
		TotalItems:   uint32(total),
		Items:        billingItems,
	}

	return
}
