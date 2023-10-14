package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *CreateOrderService) OrderItemCreate(ctx context.Context, tx database.QueryExecer, mapKeyWithOrderItemData map[string]utils.OrderItemData) (
	message utils.MessageSyncData,
	elasticData utils.ElasticSearchData,
	err error,
) {
	var associatedProducts []*pb.ProductAssociation
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
			orderItemData.GradeName,
			orderItemData.ProductSetting,
			err = s.ProductService.VerifiedProductWithStudentInfoReturnProductInfoAndBillingType(
			ctx,
			tx,
			orderItemData,
		)
		if err != nil {
			return
		}

		orderItemData.IsEnrolledInLocation, err = s.StudentService.IsEnrolledInLocation(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		err = s.StudentProductService.ValidateProductSettingForCreateOrder(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		orderItemData.StudentProduct, err = s.StudentProductService.CreateStudentProduct(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		message.StudentProducts = append(message.StudentProducts, orderItemData.StudentProduct)

		if orderItemData.ProductInfo.ProductType.String == pb.ProductType_PRODUCT_TYPE_PACKAGE.String() {
			orderItemData.PackageInfo, err = s.PackageService.VerifyPackageDataAndUpsertRelateData(ctx, tx, orderItemData)
			if err != nil {
				return
			}
			if len(orderItemData.OrderItem.ProductAssociations) > 0 {
				associatedProducts = append(associatedProducts, orderItemData.OrderItem.ProductAssociations...)
			}
			var studentPackageEvents []*npb.EventStudentPackage
			studentPackageEvents, err = s.StudentPackageService.MutationStudentPackageForCreateOrder(ctx, tx, orderItemData)
			if err != nil {
				return
			}
			message.StudentPackages = append(message.StudentPackages, studentPackageEvents...)
			message.StudentCourseMessage[key] = append([]*pb.EventSyncStudentPackageCourse{}, orderItemData.PackageInfo.StudentCourseSync...)
		}

		elasticData.OrderItems = append(elasticData.OrderItems, orderItemEntity)
		elasticData.Products = append(elasticData.Products, orderItemData.ProductInfo)

		orderItemEntity, err = s.OrderItemService.CreateOrderItem(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		orderItemData.PriceType, err = s.getPriceTypeForNew(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		err = s.BillingService.CreateBillItemForOrderCreate(ctx, tx, orderItemData)
		if err != nil {
			return
		}

		mapKeyWithOrderItemData[key] = orderItemData
		err = s.StudentProductService.CreateAssociatedStudentProductByAssociatedStudentProductID(ctx, tx, orderItemData)
		if err != nil {
			return
		}
	}

	err = s.StudentProductService.CreateAssociatedStudentProduct(ctx, tx, associatedProducts, mapKeyWithOrderItemData)
	return
}

// getPriceTypeForNew ...
func (s *CreateOrderService) getPriceTypeForNew(ctx context.Context, tx database.QueryExecer, orderItemData utils.OrderItemData) (
	priceType string,
	err error,
) {
	var isEnrolledInOrg bool
	priceType = pb.ProductPriceType_DEFAULT_PRICE.String()
	hasEnrolledPriceByProduct, err := s.hasEnrolledPriceByProductID(ctx, tx, orderItemData.StudentProduct.ProductID.String)
	if err != nil {
		return
	}
	if !hasEnrolledPriceByProduct {
		return
	}
	isEnrolledInOrg, err = s.StudentService.IsEnrolledInOrg(ctx, tx, orderItemData)
	if err != nil {
		return
	}
	if isEnrolledInOrg && hasEnrolledPriceByProduct {
		priceType = pb.ProductPriceType_ENROLLED_PRICE.String()
	}
	return
}
