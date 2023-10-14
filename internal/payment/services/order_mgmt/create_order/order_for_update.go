package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
)

func (s *CreateOrderService) OrderItemUpdate(ctx context.Context, tx database.QueryExecer, mapKeyWithOrderItemData map[string]utils.OrderItemData) (
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

		orderItemData.StudentProduct, orderItemData.RootStudentProduct, err = s.StudentProductService.MutationStudentProductForUpdateOrder(ctx, tx, orderItemData)
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
			studentPackageEvents, err = s.StudentPackageService.MutationStudentPackageForUpdateOrder(ctx, tx, orderItemData)
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
		err = s.BillingService.CreateBillItemForOrderUpdate(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		mapKeyWithOrderItemData[key] = orderItemData
	}
	return
}

func (s *CreateOrderService) getPriceTypeForUpdate(ctx context.Context, tx database.QueryExecer, orderItemData utils.OrderItemData) (
	priceType string,
	err error,
) {
	var (
		isEnrolledInOrg    bool
		rootStudentProduct entities.StudentProduct
	)
	priceType = pb.ProductPriceType_DEFAULT_PRICE.String()
	hasEnrolledPriceByProduct, err := s.hasEnrolledPriceByProductID(ctx, tx, orderItemData.StudentProduct.ProductID.String)
	if err != nil {
		return
	}
	if !hasEnrolledPriceByProduct {
		return
	}
	time := orderItemData.StudentProduct.CreatedAt.Time
	if orderItemData.StudentProduct.RootStudentProductID.Status == pgtype.Present {
		rootStudentProductID := orderItemData.StudentProduct.RootStudentProductID.String
		rootStudentProduct, err = s.StudentProductService.GetStudentProductByStudentProductID(ctx, tx, rootStudentProductID)
		if err != nil {
			return
		}
		time = rootStudentProduct.CreatedAt.Time
	}

	isEnrolledInOrg, err = s.StudentService.CheckIsEnrolledInOrgByStudentIDAndTime(ctx, tx, orderItemData.StudentInfo.StudentID.String, time)
	if err != nil {
		return
	}
	if isEnrolledInOrg && hasEnrolledPriceByProduct {
		priceType = pb.ProductPriceType_ENROLLED_PRICE.String()
	}
	return
}
