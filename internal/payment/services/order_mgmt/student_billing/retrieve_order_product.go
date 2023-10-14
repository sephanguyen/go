package studentbilling

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StudentBilling) RetrieveListOfOrderProducts(
	ctx context.Context,
	req *pb.RetrieveListOfOrderProductsRequest,
) (
	res *pb.RetrieveListOfOrderProductsResponse,
	err error,
) {
	var (
		from              int64
		limit             int64
		total             int
		studentProductIDs []string
		studentProducts   []*entities.StudentProduct
		orderProducts     []*pb.RetrieveListOfOrderProductsResponse_OrderProduct
		prevPage          *cpb.Paging
		nextPage          *cpb.Paging
	)
	from, limit, err = utils.PagingToFromAndLimit(req.Paging)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid paging data with error: %s", err.Error())
		return
	}

	studentProducts = make([]*entities.StudentProduct, 0, limit)
	orderProducts = make([]*pb.RetrieveListOfOrderProductsResponse_OrderProduct, 0, limit)
	studentProductIDs, studentProducts, total, err = s.StudentProductService.GetStudentProductByStudentIDAndLocationIDs(ctx, s.DB, req.StudentId, req.LocationIds, from, limit)
	if err != nil {
		return
	}

	mapPresentAndFutureStudentProductIDAndBillItem, err := s.BillItemService.GetMapPresentAndFutureBillItemInfo(ctx, s.DB, studentProductIDs, req.StudentId)
	if err != nil {
		return
	}

	mapPastStudentProductIDAndBillItem, err := s.BillItemService.GetMapPastBillItemInfo(ctx, s.DB, studentProductIDs, req.StudentId)
	if err != nil {
		return
	}

	count := int32(from)
	for _, studentProduct := range studentProducts {
		var (
			orderProduct *pb.RetrieveListOfOrderProductsResponse_OrderProduct
			billItem     *entities.BillItem
		)
		count++
		if mapPresentAndFutureStudentProductIDAndBillItem[studentProduct.StudentProductID.String] != nil {
			billItem = mapPresentAndFutureStudentProductIDAndBillItem[studentProduct.StudentProductID.String]
		} else {
			billItem = mapPastStudentProductIDAndBillItem[studentProduct.StudentProductID.String]
		}

		orderProduct, err = utils.ConvertEntityStudentProductAndBillItemToOrderProductInStudentBilling(billItem, studentProduct)
		if err != nil {
			return
		}

		productSetting, err := s.ProductService.GetProductSettingByProductID(ctx, s.DB, orderProduct.ProductId)
		if err != nil {
			return nil, err
		}
		orderProduct.IsOperationFee = productSetting.IsOperationFee.Bool

		if orderProduct.Status.String() != pb.StudentProductStatus_CANCELLED.String() && orderProduct.StudentProductLabel.String() != pb.StudentProductLabel_PAUSED.String() {
			if orderProduct.MaterialType == pb.MaterialType_MATERIAL_TYPE_ONE_TIME {
				material, err := s.MaterialService.GetMaterialByID(ctx, s.DB, billItem.ProductID.String)
				if err != nil {
					return nil, err
				}

				if time.Now().Before(material.CustomBillingDate.Time) {
					orderProduct.UpcomingBillingDate = timestamppb.New(material.CustomBillingDate.Time)
				}

			} else {
				var upcomingBillingItem *entities.BillItem
				upcomingBillingItem, err = s.BillItemService.GetUpcomingBilling(ctx, s.DB, billItem.StudentProductID.String, req.StudentId)
				if err != nil {
					return nil, err
				}
				if billItem.BillSchedulePeriodID.Status == pgtype.Present &&
					upcomingBillingItem.BillDate.Time.After(time.Now()) {
					orderProduct.UpcomingBillingDate = timestamppb.New(upcomingBillingItem.BillDate.Time)
					if upcomingBillingItem.DiscountID.Status != pgtype.Present {
						orderProduct.DiscountInfo = &pb.RetrieveListOfOrderProductsResponse_OrderProduct_DiscountInfo{}
					}
				}
			}
		}

		orderProduct.Index = count
		orderProducts = append(orderProducts, orderProduct)
	}

	prevPage, nextPage, err = utils.ConvertCommonPaging(total, from, limit)
	if err != nil {
		return
	}

	res = &pb.RetrieveListOfOrderProductsResponse{
		NextPage:     nextPage,
		PreviousPage: prevPage,
		TotalItems:   uint32(total),
		Items:        orderProducts,
	}
	return
}
