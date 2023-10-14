package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *CreateOrderService) OrderItemLoa(ctx context.Context, tx database.QueryExecer, mapKeyWithOrderItemData map[string]utils.OrderItemData) (
	message utils.MessageSyncData,
	elasticData utils.ElasticSearchData,
	err error,
) {
	message.StudentCourseMessage = make(map[string][]*pb.EventSyncStudentPackageCourse)
	message.StudentProducts = []entities.StudentProduct{}

	for key := range mapKeyWithOrderItemData {
		var (
			orderItemEntity entities.OrderItem
			orderItemData   utils.OrderItemData
		)
		orderItemData = mapKeyWithOrderItemData[key]
		err = s.StudentService.ValidateStudentStatusForOrderType(ctx, tx, pb.OrderType_ORDER_TYPE_LOA, orderItemData.StudentInfo, orderItemData.Order.LocationID.String, orderItemData.OrderItem.EffectiveDate.AsTime())
		if err != nil {
			return
		}

		orderItemData.ProductInfo,
			orderItemData.IsOneTimeProduct,
			orderItemData.IsDisableProRatingFlag,
			orderItemData.ProductType,
			orderItemData.ProductSetting,
			err = s.ProductService.VerifiedProductReturnProductInfoAndBillingType(
			ctx,
			tx,
			orderItemData,
		)
		if err != nil {
			return
		}

		err = s.StudentProductService.ValidateProductSettingForLOAOrder(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		orderItemData.StudentProduct, err = s.StudentProductService.MutationStudentProductForLOAOrder(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		message.StudentProducts = append(message.StudentProducts, orderItemData.StudentProduct)

		orderItemEntity, err = s.OrderItemService.CreateOrderItem(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		if orderItemData.ProductInfo.ProductType.String == pb.ProductType_PRODUCT_TYPE_PACKAGE.String() {
			orderItemData.PackageInfo, err = s.PackageService.VerifyPackageDataAndUpsertRelateData(ctx, tx, orderItemData)
			if err != nil {
				return
			}

			var studentPackageEvents []*npb.EventStudentPackage
			studentPackageEvents, err = s.StudentPackageService.MutationStudentPackageForCancelOrder(ctx, tx, orderItemData)
			if err != nil {
				return
			}
			message.StudentPackages = append(message.StudentPackages, studentPackageEvents...)
			message.StudentCourseMessage[key] = append([]*pb.EventSyncStudentPackageCourse{}, orderItemData.PackageInfo.StudentCourseSync...)
		}

		elasticData.OrderItems = append(elasticData.OrderItems, orderItemEntity)
		elasticData.Products = append(elasticData.Products, orderItemData.ProductInfo)

		if len(orderItemData.BillItems) == 0 {
			return
		}
		orderItemData.PriceType, err = s.getPriceTypeForUpdate(ctx, tx, orderItemData)
		if err != nil {
			return
		}
		err = s.BillingService.CreateBillItemForOrderLOA(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		mapKeyWithOrderItemData[key] = orderItemData
	}
	return
}
