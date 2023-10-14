package service

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *CreateOrderService) OrderItemCancel(ctx context.Context, tx database.QueryExecer, mapKeyWithOrderItemData map[string]utils.OrderItemData) (
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

		orderItemData.OrderItem.CancellationDate = timestamppb.New(time.Now())
		orderItemData.StudentProduct, orderItemData.RootStudentProduct, err = s.StudentProductService.MutationStudentProductForCancelOrder(ctx, tx, orderItemData)
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

		orderItemData.PriceType, err = s.getPriceTypeForUpdate(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		err = s.BillingService.CreateBillItemForOrderCancel(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		mapKeyWithOrderItemData[key] = orderItemData
		err = s.StudentProductService.CreateAssociatedStudentProductByAssociatedStudentProductID(ctx, tx, orderItemData)
		if err != nil {
			return
		}
	}
	return
}
