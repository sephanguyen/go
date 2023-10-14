package studentbilling

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StudentBilling) RetrieveListOfOrderAssociatedProductOfPackages(
	ctx context.Context,
	req *pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest,
) (
	res *pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse,
	err error,
) {
	var (
		from                                           int64
		limit                                          int64
		total                                          int
		studentProductIDs                              []string
		studentProducts                                []*entities.StudentProduct
		prevPage                                       *cpb.Paging
		nextPage                                       *cpb.Paging
		mapPresentAndFutureStudentProductIDAndBillItem map[string]*entities.BillItem
		mapPastStudentProductIDAndBillItem             map[string]*entities.BillItem
	)
	from, limit, err = utils.PagingToFromAndLimit(req.Paging)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid paging data with error: %s", err.Error())
		return
	}

	orderProducts := make([]*pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct, 0, limit)
	studentProductIDs, studentProducts, total, err = s.StudentProductService.GetStudentAssociatedProductByStudentProductID(ctx, s.DB, req.StudentProductId, from, limit)
	if err != nil {
		return
	}

	if len(studentProducts) != 0 {
		mapPresentAndFutureStudentProductIDAndBillItem, err = s.BillItemService.GetMapPresentAndFutureBillItemInfo(ctx, s.DB, studentProductIDs, studentProducts[0].StudentID.String)
		if err != nil {
			return
		}

		mapPastStudentProductIDAndBillItem, err = s.BillItemService.GetMapPastBillItemInfo(ctx, s.DB, studentProductIDs, studentProducts[0].StudentID.String)
		if err != nil {
			return
		}
	}

	count := int32(from)
	for _, studentProduct := range studentProducts {
		var (
			orderProduct *pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct
			billItem     *entities.BillItem
		)
		count++
		if mapPresentAndFutureStudentProductIDAndBillItem[studentProduct.StudentProductID.String] != nil {
			billItem = mapPresentAndFutureStudentProductIDAndBillItem[studentProduct.StudentProductID.String]
		} else {
			billItem = mapPastStudentProductIDAndBillItem[studentProduct.StudentProductID.String]
		}

		orderProduct, err = utils.ConvertEntityStudentProductAndBillItemToOrderProductAssociatedInStudentBilling(billItem, studentProduct)
		if err != nil {
			return
		}

		if orderProduct.Status.String() != pb.StudentProductStatus_CANCELLED.String() {
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
				upcomingBillingItem, err = s.BillItemService.GetUpcomingBilling(ctx, s.DB, billItem.StudentProductID.String, studentProducts[0].StudentID.String)
				if err != nil {
					return nil, err
				}
				if billItem.BillSchedulePeriodID.Status == pgtype.Present &&
					upcomingBillingItem.BillDate.Time.After(time.Now()) {
					orderProduct.UpcomingBillingDate = timestamppb.New(upcomingBillingItem.BillDate.Time)
					if upcomingBillingItem.DiscountID.Status != pgtype.Present {
						orderProduct.DiscountInfo = &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct_DiscountInfo{}
					}
				}
			}
		}

		orderProduct.Index = count
		orderProducts = append(orderProducts, orderProduct)
	}

	totalAssociatedProductsOfPackage, err := s.PackageService.GetTotalAssociatedPackageWithCourseIDAndPackageID(ctx, s.DB, req.PackageId, req.CourseIds)
	if err != nil {
		return
	}

	prevPage, nextPage, err = utils.ConvertCommonPaging(total, from, limit)
	if err != nil {
		return
	}

	res = &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse{
		NextPage:                         nextPage,
		PreviousPage:                     prevPage,
		TotalItems:                       uint32(total),
		Items:                            orderProducts,
		TotalAssociatedProductsOfPackage: totalAssociatedProductsOfPackage,
	}
	return
}
