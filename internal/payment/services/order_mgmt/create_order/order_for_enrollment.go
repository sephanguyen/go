package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *CreateOrderService) OrderItemEnrollment(ctx context.Context, tx database.QueryExecer, mapKeyWithOrderItemData map[string]utils.OrderItemData) (
	message utils.MessageSyncData,
	elasticData utils.ElasticSearchData,
	err error,
) {
	var (
		associatedProducts []*pb.ProductAssociation
		isCheckEnrollment  bool
	)
	message.StudentCourseMessage = make(map[string][]*pb.EventSyncStudentPackageCourse)
	message.StudentProducts = []entities.StudentProduct{}

	for key := range mapKeyWithOrderItemData {
		var (
			orderItemEntity entities.OrderItem
			orderItemData   utils.OrderItemData
		)
		orderItemData = mapKeyWithOrderItemData[key]
		if !isCheckEnrollment {
			err = s.StudentService.ValidateStudentStatusForOrderType(ctx, tx, pb.OrderType_ORDER_TYPE_ENROLLMENT, orderItemData.StudentInfo, orderItemData.Order.LocationID.String, orderItemData.OrderItem.EffectiveDate.AsTime())
			if err != nil {
				return
			}
			isCheckEnrollment = true
		}
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

		orderItemData.PriceType, err = s.getPriceTypeForEnrollment(ctx, tx, orderItemData)
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

func (s *CreateOrderService) getPriceTypeForEnrollment(ctx context.Context, tx database.QueryExecer, orderItemData utils.OrderItemData) (
	priceType string,
	err error,
) {
	priceType = pb.ProductPriceType_DEFAULT_PRICE.String()
	hasEnrolledPriceByProduct, err := s.hasEnrolledPriceByProductID(ctx, tx, orderItemData.StudentProduct.ProductID.String)
	if err != nil {
		return
	}
	if !hasEnrolledPriceByProduct {
		return
	}
	if hasEnrolledPriceByProduct {
		priceType = pb.ProductPriceType_ENROLLED_PRICE.String()
	}
	return
}
