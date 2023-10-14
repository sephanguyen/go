package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	discount "github.com/manabie-com/backend/internal/payment/services/domain_service/discount"
	tax "github.com/manabie-com/backend/internal/payment/services/domain_service/tax"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IDiscountServiceForProductPrice interface {
	CalculatorDiscountPrice(
		discount entities.Discount,
		price float32,
		billItem *entities.BillItem,
	) (finalPrice float32, err error)
}
type ITaxServiceForProductPrice interface {
	CalculatorTaxPrice(
		tax entities.Tax,
		price float32,
		billItem *entities.BillItem,
	) (finalPrice float32, err error)
}
type PriceService struct {
	productPriceRepo interface {
		GetByProductIDAndPriceType(ctx context.Context, db database.QueryExecer, productID string, priceType string) ([]entities.ProductPrice, error)
		GetByProductIDAndQuantityAndPriceType(ctx context.Context, db database.QueryExecer, productID string, weight int32, priceType string) (entities.ProductPrice, error)
		GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(ctx context.Context, db database.QueryExecer, productID string, billingSchedulePeriodID string, quantity int32, priceType string) (entities.ProductPrice, error)
		GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx context.Context, db database.QueryExecer, productID string, billingSchedulePeriodID string, priceType string) (entities.ProductPrice, error)
	}
	DiscountService IDiscountServiceForProductPrice
	TaxService      ITaxServiceForProductPrice
}

func (s *PriceService) IsValidPriceForOneTimeBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	var (
		billingData utils.BillingItemData
		billItem    *pb.BillingItem
	)
	billingData = orderItemData.BillItems[0]
	billItem = billingData.BillingItem
	if orderItemData.ProductType != pb.ProductType_PRODUCT_TYPE_PACKAGE {
		return s.validatePriceForOneTimeNonQuantityProduct(ctx, db, orderItemData, billItem)
	}
	return s.validatePriceForOneTimeQuantityProduct(ctx, db, orderItemData, billItem)
}

func (s *PriceService) validatePriceForOneTimeNonQuantityProduct(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	billingItem *pb.BillingItem,
) (err error) {
	var (
		productPrices []entities.ProductPrice
	)
	productPrices, err = s.getPricesOfNoneQuantityProduct(ctx, db, orderItemData.ProductInfo, orderItemData.PriceType)
	if err != nil {
		return
	}
	err = s.checkPriceForNoneQuantityProduct(billingItem, productPrices)
	return
}

func (s *PriceService) getPricesOfNoneQuantityProduct(
	ctx context.Context,
	tx database.QueryExecer,
	productInfo entities.Product,
	priceType string,
) (productPrices []entities.ProductPrice, err error) {
	productPrices, err = s.productPriceRepo.GetByProductIDAndPriceType(ctx, tx, productInfo.ProductID.String, priceType)
	if err != nil {
		err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
			productInfo.ProductID.String,
			err.Error(),
		)
	}
	return
}

func (s *PriceService) checkPriceForNoneQuantityProduct(
	billItem *pb.BillingItem,
	productPrices []entities.ProductPrice,
) (err error) {
	var foundPrice bool
	for _, price := range productPrices {
		foundPrice = utils.IsEqualNumericAndFloat32(price.Price, billItem.Price)
		if foundPrice {
			break
		}
	}
	if !foundPrice {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.ProductPriceNotExistOrUpdated,
			&errdetails.DebugInfo{
				Detail: fmt.Sprintf("Price of product with id %v does not exist in system (or just updated)", billItem.ProductId),
			},
		)
	} else {
		err = validateFinalPrice(billItem)
	}
	return
}

func (s *PriceService) getPricesOfQuantityProduct(
	ctx context.Context,
	tx database.QueryExecer,
	packageInfo utils.PackageInfo,
	priceType string,
) (productPrice entities.ProductPrice, err error) {
	productPrice, err = s.productPriceRepo.GetByProductIDAndQuantityAndPriceType(
		ctx,
		tx,
		packageInfo.Package.PackageID.String,
		packageInfo.Quantity,
		priceType,
	)
	if err != nil {
		err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
			packageInfo.Package.PackageID.String,
			err.Error(),
		)
		return
	}
	return
}

func (s *PriceService) validatePriceForOneTimeQuantityProduct(
	ctx context.Context,
	tx database.QueryExecer,
	orderItemData utils.OrderItemData,
	billingItem *pb.BillingItem,
) (err error) {
	var (
		productPrice entities.ProductPrice
	)
	if billingItem.Quantity == nil {
		err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, billingItem.ProductId)
		return
	}
	if orderItemData.PackageInfo.Quantity != billingItem.Quantity.Value {
		err = status.Errorf(codes.Internal,
			"inconsistency quantity between order item and bill item of product %v, order item quantity %v vs bill item quantity %v",
			billingItem.ProductId,
			orderItemData.PackageInfo.Quantity,
			billingItem.Quantity.Value,
		)
		return
	}
	productPrice, err = s.getPricesOfQuantityProduct(ctx, tx, orderItemData.PackageInfo, orderItemData.PriceType)
	if err != nil {
		return
	}
	if !utils.IsEqualNumericAndFloat32(productPrice.Price, billingItem.Price) {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.ProductPriceNotExistOrUpdated,
			&errdetails.DebugInfo{
				Detail: fmt.Sprintf("Price of package with id %v and quantity %v does not exist in system (or just updated)", billingItem.ProductId, billingItem.Quantity.Value),
			},
		)
	} else {
		err = validateFinalPrice(billingItem)
	}
	return
}

func (s *PriceService) IsValidAdjustmentPriceForOneTimeBilling(
	oldBillItem entities.BillItem,
	orderItemData utils.OrderItemData) (err error) {
	var (
		billingData utils.BillingItemData
		billItem    *pb.BillingItem
	)
	billingData = orderItemData.BillItems[0]
	billItem = billingData.BillingItem
	return validateAdjustmentPrice(billItem, oldBillItem, 1, 1)
}

func calculateOriginalDiscount(originalPrice float32, discountType string, discountAmountValue float32) (originalDiscount float32) {
	originalDiscount = discountAmountValue
	if discountType == pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String() {
		originalDiscount = (originalPrice * discountAmountValue) / float32(100)
	}
	return
}

func validateAdjustmentPrice(
	billingItem *pb.BillingItem,
	oldBillItem entities.BillItem,
	billingRatioNumerator int32,
	billingRatioDenominator int32,
) error {
	var (
		oldOriginalPrice   float32
		newOriginalPrice   float32
		tmpAdjustmentPrice float32
	)
	if billingItem.AdjustmentPrice == nil {
		return status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, billingItem.ProductId)
	}
	_ = oldBillItem.Price.AssignTo(&oldOriginalPrice)
	if oldBillItem.DiscountAmount.Status == pgtype.Present {
		var (
			discountAmountValue float32
		)
		_ = oldBillItem.DiscountAmountValue.AssignTo(&discountAmountValue)
		oldOriginalDiscount := calculateOriginalDiscount(oldOriginalPrice, oldBillItem.DiscountAmountType.String, discountAmountValue)
		oldOriginalPrice -= oldOriginalDiscount
	}
	newOriginalPrice = (billingItem.Price * float32(billingRatioDenominator)) / float32(billingRatioNumerator)

	if billingItem.DiscountItem != nil {
		newOriginalDiscount := calculateOriginalDiscount(newOriginalPrice, billingItem.DiscountItem.DiscountAmountType.String(), billingItem.DiscountItem.DiscountAmountValue)
		newOriginalPrice -= newOriginalDiscount
	}
	if billingRatioNumerator == 0 || billingRatioDenominator == 0 {
		tmpAdjustmentPrice = 0
	} else {
		tmpAdjustmentPrice = ((newOriginalPrice - oldOriginalPrice) * float32(billingRatioNumerator)) / float32(billingRatioDenominator)
	}
	if !utils.CompareAmountValue(tmpAdjustmentPrice, billingItem.AdjustmentPrice.Value) {
		return status.Errorf(codes.FailedPrecondition, "Incorrect adjustment price for update of product %v actual = %v vs expect = %v",
			billingItem.ProductId,
			billingItem.AdjustmentPrice.Value,
			tmpAdjustmentPrice,
		)
	}
	////Todo: Will add logic subtract tax exclusive
	return nil
}

func validateAdjustmentPriceForCancelOrder(
	billingItem *pb.BillingItem,
	oldBillItem entities.BillItem,
	billingRatioNumerator int32,
	billingRatioDenominator int32,
) error {
	var (
		oldOriginalPrice   float32
		tmpAdjustmentPrice float32
	)
	if billingItem.AdjustmentPrice == nil {
		return status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, billingItem.ProductId)
	}
	_ = oldBillItem.Price.AssignTo(&oldOriginalPrice)
	if oldBillItem.DiscountAmount.Status == pgtype.Present {
		var (
			discountAmountValue float32
		)
		_ = oldBillItem.DiscountAmountValue.AssignTo(&discountAmountValue)
		oldOriginalDiscount := calculateOriginalDiscount(oldOriginalPrice, oldBillItem.DiscountAmountType.String, discountAmountValue)
		oldOriginalPrice -= oldOriginalDiscount
	}
	if billingRatioNumerator == 0 || billingRatioDenominator == 0 {
		tmpAdjustmentPrice = 0
	} else {
		tmpAdjustmentPrice = -(oldOriginalPrice) * (float32(billingRatioNumerator) / float32(billingRatioDenominator))
	}
	if !utils.CompareAmountValue(tmpAdjustmentPrice, billingItem.AdjustmentPrice.Value) {
		return status.Errorf(codes.FailedPrecondition, "Incorrect adjustment price for cancel of product %v actual = %v vs expect = %v",
			billingItem.ProductId,
			billingItem.AdjustmentPrice.Value,
			tmpAdjustmentPrice,
		)
	}
	////Todo: Will add logic subtract tax exclusive
	return nil
}

func (s *PriceService) IsValidPriceForRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	nonProRatedBillItem []utils.BillingItemData) (proRatedPrice entities.ProductPrice, err error) {
	if orderItemData.ProductType != pb.ProductType_PRODUCT_TYPE_PACKAGE {
		proRatedPrice, err = s.checkPriceRecurringNoneQuantityBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			nonProRatedBillItem,
		)
		return
	}

	proRatedPrice, err = s.checkPriceRecurringQuantityBilling(
		ctx,
		db,
		orderItemData,
		proRatedBillItem,
		ratioOfProRatedBillingItem,
		nonProRatedBillItem,
	)
	return
}

func (s *PriceService) checkPriceRecurringNoneQuantityBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	nonProRatedBillItems []utils.BillingItemData) (proRatedPrice entities.ProductPrice, err error) {
	var (
		firstProductPrice entities.ProductPrice
	)
	if proRatedBillItem.BillingItem != nil {
		firstProductPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, db, proRatedBillItem.BillingItem.ProductId, proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value, orderItemData.PriceType)
		if err != nil {
			err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProRatingPriceOfNoneQuantity, err.Error())
			return
		}
		tmpProductPrice := firstProductPrice
		proRatedPrice = firstProductPrice
		err = s.checkPriceForProRatingBillItem(tmpProductPrice, ratioOfProRatedBillingItem, proRatedBillItem.BillingItem)
		if err != nil {
			return
		}
	}
	for _, data := range nonProRatedBillItems {
		var productPrice entities.ProductPrice
		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, db, data.BillingItem.ProductId, data.BillingItem.BillingSchedulePeriodId.Value, orderItemData.PriceType)
		if err != nil {
			err = status.Errorf(codes.Internal, "getting price of none quantity recurring product have have err %v", err.Error())
			return
		}
		err = s.checkPriceForNormalBillItem(productPrice, data.BillingItem)
		if err != nil {
			return
		}
	}
	return
}

func (s *PriceService) checkUpdatePriceRecurringNoneQuantityBilling(
	ctx context.Context,
	db database.QueryExecer,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapOldBillingItem map[string]entities.BillItem,
	orderItemData utils.OrderItemData,
) (proRatedPrice entities.ProductPrice, err error) {
	if proRatedBillItem.BillingItem != nil {
		var (
			productPrice        entities.ProductPrice
			oldProRatedBillItem entities.BillItem
		)
		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, db, proRatedBillItem.BillingItem.ProductId, proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value, orderItemData.PriceType)
		if err != nil {
			err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProRatingPriceOfNoneQuantity, err.Error())
			return
		}
		proRatedPrice = productPrice
		err = s.checkPriceForProRatingBillItem(productPrice, ratioOfProRatedBillingItem, proRatedBillItem.BillingItem)
		if err != nil {
			return
		}

		oldProRatedBillItem = mapOldBillingItem[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]
		err = validateAdjustmentPrice(
			proRatedBillItem.BillingItem,
			oldProRatedBillItem,
			ratioOfProRatedBillingItem.BillingRatioNumerator.Int,
			ratioOfProRatedBillingItem.BillingRatioDenominator.Int,
		)
		if err != nil {
			return
		}
	}

	for _, data := range normalBillItem {
		var (
			productPrice entities.ProductPrice
			oldBillItem  entities.BillItem
		)
		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, db, data.BillingItem.ProductId, data.BillingItem.BillingSchedulePeriodId.Value, orderItemData.PriceType)
		if err != nil {
			err = status.Errorf(codes.Internal, "getting price of none quantity recurring product have have err %v", err.Error())
			return
		}
		err = s.checkPriceForNormalBillItem(productPrice, data.BillingItem)
		if err != nil {
			return
		}

		oldBillItem = mapOldBillingItem[data.BillingItem.BillingSchedulePeriodId.Value]
		err = validateAdjustmentPrice(
			data.BillingItem,
			oldBillItem,
			1,
			1,
		)
		if err != nil {
			return
		}
	}
	return
}

func (s *PriceService) checkPriceForNormalBillItem(productPrice entities.ProductPrice, billItem *pb.BillingItem) (err error) {
	if !utils.IsEqualNumericAndFloat32(productPrice.Price, billItem.Price) {
		var floatProductPrice float32
		_ = productPrice.Price.AssignTo(&floatProductPrice)
		err = status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice, billItem.ProductId, billItem.Price, floatProductPrice)
	} else {
		err = validateFinalPrice(billItem)
	}
	return
}

func (s *PriceService) checkPriceForProRatingBillItem(productPrice entities.ProductPrice, ratio entities.BillingRatio, billItem *pb.BillingItem) (err error) {
	var (
		productPriceConverted  float32
		productPriceAfterRatio float32
	)

	productPriceConverted = utils.ConvertNumericToFloat32(productPrice.Price)
	productPriceAfterRatio = (productPriceConverted * float32(ratio.BillingRatioNumerator.Int)) / float32(ratio.BillingRatioDenominator.Int)
	if productPriceAfterRatio != billItem.Price {
		err = status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice,
			billItem.ProductId,
			billItem.Price,
			productPriceAfterRatio,
		)
	} else {
		err = validateFinalPrice(billItem)
	}
	return
}

func (s *PriceService) checkPriceRecurringQuantityBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	nonProRatedBillItems []utils.BillingItemData) (proRatedPrice entities.ProductPrice, err error) {
	var (
		firstProductPrice entities.ProductPrice
	)
	if proRatedBillItem.BillingItem != nil {
		err = s.checkQuantityOfOrderItemAndBillItem(orderItemData, proRatedBillItem)
		if err != nil {
			return
		}

		firstProductPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(
			ctx,
			db,
			proRatedBillItem.BillingItem.ProductId,
			proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value,
			orderItemData.PackageInfo.Quantity,
			orderItemData.PriceType,
		)
		if err != nil {
			err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProRatingPriceOfNoneQuantity, err.Error())
			return
		}

		tmpPrice := firstProductPrice
		proRatedPrice = firstProductPrice

		err = s.checkPriceForProRatingBillItem(tmpPrice, ratioOfProRatedBillingItem, proRatedBillItem.BillingItem)
		if err != nil {
			return
		}
	}
	for _, data := range nonProRatedBillItems {
		var productPrice entities.ProductPrice
		err = s.checkQuantityOfOrderItemAndBillItem(orderItemData, data)
		if err != nil {
			return
		}

		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(
			ctx,
			db,
			data.BillingItem.ProductId,
			data.BillingItem.BillingSchedulePeriodId.Value,
			orderItemData.PackageInfo.Quantity,
			orderItemData.PriceType,
		)

		if err != nil {
			err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProRatingPriceOfNoneQuantity, err.Error())
			return
		}

		err = s.checkPriceForNormalBillItem(productPrice, data.BillingItem)
		if err != nil {
			return
		}
	}
	return
}

func (s *PriceService) checkUpdatePriceRecurringQuantityBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapOldBillingItem map[string]entities.BillItem,
) (proRatedPRice entities.ProductPrice, err error) {
	if proRatedBillItem.BillingItem != nil {
		var (
			productPrice        entities.ProductPrice
			oldProRatedBillItem entities.BillItem
		)
		err = s.checkQuantityOfOrderItemAndBillItem(orderItemData, proRatedBillItem)
		if err != nil {
			return
		}

		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(
			ctx,
			db,
			proRatedBillItem.BillingItem.ProductId,
			proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value,
			orderItemData.PackageInfo.Quantity,
			orderItemData.PriceType,
		)
		if err != nil {
			err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProRatingPriceOfNoneQuantity, err.Error())
			return
		}
		proRatedPRice = productPrice
		err = s.checkPriceForProRatingBillItem(productPrice, ratioOfProRatedBillingItem, proRatedBillItem.BillingItem)
		if err != nil {
			return
		}

		oldProRatedBillItem = mapOldBillingItem[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]
		err = validateAdjustmentPrice(
			proRatedBillItem.BillingItem,
			oldProRatedBillItem,
			ratioOfProRatedBillingItem.BillingRatioNumerator.Int,
			ratioOfProRatedBillingItem.BillingRatioDenominator.Int,
		)
		if err != nil {
			return
		}
	}
	for _, data := range normalBillItem {
		var (
			productPrice        entities.ProductPrice
			oldProRatedBillItem entities.BillItem
		)
		err = s.checkQuantityOfOrderItemAndBillItem(orderItemData, data)
		if err != nil {
			return
		}

		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(
			ctx,
			db,
			data.BillingItem.ProductId,
			data.BillingItem.BillingSchedulePeriodId.Value,
			orderItemData.PackageInfo.Quantity,
			orderItemData.PriceType,
		)

		if err != nil {
			err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProRatingPriceOfNoneQuantity, err.Error())
			return
		}

		err = s.checkPriceForNormalBillItem(productPrice, data.BillingItem)
		if err != nil {
			return
		}

		oldProRatedBillItem = mapOldBillingItem[data.BillingItem.BillingSchedulePeriodId.Value]
		err = validateAdjustmentPrice(
			data.BillingItem,
			oldProRatedBillItem,
			1,
			1,
		)
		if err != nil {
			return
		}
	}
	return
}

func (s *PriceService) checkQuantityOfOrderItemAndBillItem(
	orderItemData utils.OrderItemData,
	billItem utils.BillingItemData,
) (err error) {
	if billItem.BillingItem.Quantity == nil {
		err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, billItem.BillingItem.ProductId)
		return
	}

	if orderItemData.PackageInfo.Quantity != billItem.BillingItem.Quantity.Value {
		err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, billItem.BillingItem.ProductId)
	}
	return
}

func (s *PriceService) IsValidPriceForUpdateRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapOldBillingItem map[string]entities.BillItem,
) (proRatedPRice entities.ProductPrice, err error) {
	if orderItemData.ProductType != pb.ProductType_PRODUCT_TYPE_PACKAGE {
		return s.checkUpdatePriceRecurringNoneQuantityBilling(
			ctx,
			db,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapOldBillingItem,
			orderItemData,
		)
	}

	return s.checkUpdatePriceRecurringQuantityBilling(
		ctx,
		db,
		orderItemData,
		proRatedBillItem,
		ratioOfProRatedBillingItem,
		normalBillItem,
		mapOldBillingItem,
	)
}

func (s *PriceService) IsValidPriceForCancelRecurringBilling(
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapOldBillingItem map[string]entities.BillItem,
	mapPeriodInfo map[string]entities.BillingSchedulePeriod,
) (err error) {
	if proRatedBillItem.BillingItem != nil {
		var (
			oldProRatedBillItem entities.BillItem
			oldPrice            float32
		)
		oldProRatedBillItem = mapOldBillingItem[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]
		_ = oldProRatedBillItem.Price.AssignTo(&oldPrice)
		err = utils.GroupErrorFunc(
			validateAdjustmentPriceForCancelOrder(
				proRatedBillItem.BillingItem,
				oldProRatedBillItem,
				ratioOfProRatedBillingItem.BillingRatioNumerator.Int,
				ratioOfProRatedBillingItem.BillingRatioDenominator.Int,
			),
		)
		if err != nil {
			err = status.Errorf(codes.FailedPrecondition, "error in proRated %v", err.Error())
			return
		}
	}

	for _, data := range normalBillItem {
		var (
			oldBillItem entities.BillItem
			oldPrice    float32
		)
		periodData := mapPeriodInfo[data.BillingItem.BillingSchedulePeriodId.Value]
		if orderItemData.IsDisableProRatingFlag &&
			orderItemData.OrderItem.EffectiveDate.AsTime().After(periodData.StartDate.Time) &&
			!orderItemData.OrderItem.EffectiveDate.AsTime().After(periodData.EndDate.Time) {
			if data.BillingItem.AdjustmentPrice.Value != 0 {
				err = status.Errorf(codes.FailedPrecondition, "adjustment price of this period %v should be 0 with start date: %v, end date: %v, effective date: %v, adjustment price: %v",
					data.BillingItem.BillingSchedulePeriodId.Value,
					periodData.StartDate.Time.String(),
					periodData.EndDate.Time.String(),
					orderItemData.OrderItem.EffectiveDate.AsTime(),
					data.BillingItem.AdjustmentPrice.Value,
				)
				return
			}
		} else {
			oldBillItem = mapOldBillingItem[data.BillingItem.BillingSchedulePeriodId.Value]
			_ = oldBillItem.Price.AssignTo(&oldPrice)
			err = utils.GroupErrorFunc(
				validateAdjustmentPriceForCancelOrder(
					data.BillingItem,
					oldBillItem,
					1,
					1,
				),
			)
			if err != nil {
				err = status.Errorf(codes.FailedPrecondition, "error in normal %v period id %v", err.Error(),
					data.BillingItem.BillingSchedulePeriodId.Value,
				)
				return
			}
		}
	}
	return
}

func validateFinalPrice(billingItem *pb.BillingItem) (err error) {
	var tmpFinalPrice float32
	tmpFinalPrice = billingItem.Price
	if billingItem.DiscountItem != nil {
		tmpFinalPrice -= billingItem.DiscountItem.DiscountAmount
	}
	if !utils.CompareAmountValue(billingItem.FinalPrice, tmpFinalPrice) {
		err = status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, billingItem.ProductId, billingItem.FinalPrice, tmpFinalPrice)
	}
	////Todo: Will add logic subtract tax exclusive
	return
}

func (s *PriceService) CalculatorBillItemPrice(
	ctx context.Context,
	db database.QueryExecer,
	billItem *entities.BillItem,
	upcomingBillItem entities.UpcomingBillItem,
	tax entities.Tax,
	discount entities.Discount,
	priceType string,
	billItemDescription *entities.BillingItemDescription,
	billingSchedulePeriod entities.BillingSchedulePeriod,
) (err error) {
	var (
		productPrice entities.ProductPrice
		quantity     int32
	)
	if billItemDescription.ProductType == pb.ProductType_PRODUCT_TYPE_PACKAGE.String() {
		switch *billItemDescription.QuantityType {
		case pb.QuantityType_QUANTITY_TYPE_SLOT.String(), pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String():
			for _, item := range billItemDescription.CourseItems {
				quantity += *item.Slot
			}
		case pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String():
			for _, item := range billItemDescription.CourseItems {
				quantity += *item.Weight
			}
		}
		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(
			ctx,
			db,
			upcomingBillItem.ProductID.String,
			billingSchedulePeriod.BillingSchedulePeriodID.String,
			quantity,
			priceType,
		)
	} else {
		productPrice, err = s.productPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(
			ctx,
			db,
			upcomingBillItem.ProductID.String,
			billingSchedulePeriod.BillingSchedulePeriodID.String,
			priceType,
		)
	}

	if err != nil {
		err = fmt.Errorf("error when get product price of product ID %s, product type %s, billing schedule period ID %s, price type %s, quantity %d, with error %v", upcomingBillItem.ProductID.String, billItemDescription.ProductType, billingSchedulePeriod.BillingScheduleID.String, priceType, quantity, err)
		return
	}
	finalPrice := utils.ConvertNumericToFloat32(productPrice.Price)
	if discount.DiscountID.Status == pgtype.Present {
		finalPrice, err = s.DiscountService.CalculatorDiscountPrice(discount, finalPrice, billItem)
		if err != nil {
			err = fmt.Errorf("error while calculate discount id %s in order id %s with err: %v", discount.DiscountID.String, billItem.OrderID.String, err)
			return
		}
	}
	if tax.TaxID.Status == pgtype.Present {
		finalPrice, err = s.TaxService.CalculatorTaxPrice(tax, finalPrice, billItem)
		if err != nil {
			err = fmt.Errorf("error while calculate tax id %s in order id %s with err: %v", tax.TaxID.String, billItem.OrderID.String, err)
			return
		}
	}
	err = multierr.Combine(
		billItem.Price.Set(utils.ConvertNumericToFloat32(productPrice.Price)),
		billItem.FinalPrice.Set(finalPrice),
		billItem.ProductPricing.Set(utils.ConvertNumericToFloat32(productPrice.Price)),
	)
	if err != nil {
		err = fmt.Errorf(
			"err occurred while assigning final price into orderID %s and productID %s : %v",
			billItem.OrderID.String,
			billItem.ProductID.String,
			err,
		)
	}
	return
}

func (s *PriceService) GetProductPricesByProductIDAndPriceType(
	ctx context.Context,
	db database.QueryExecer,
	productID string,
	priceType string,
) (productPrices []entities.ProductPrice, err error) {
	productPrices, err = s.productPriceRepo.GetByProductIDAndPriceType(ctx, db, productID, priceType)
	if err != nil {
		err = status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
			productID,
			err.Error(),
		)
	}
	return
}

func NewPriceService() *PriceService {
	return &PriceService{
		productPriceRepo: &repositories.ProductPriceRepo{},
		DiscountService:  discount.NewDiscountService(),
		TaxService:       tax.NewTaxService(),
	}
}
