package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaxService struct {
	taxRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, taxID string) (entities.Tax, error)
	}
}

func (s *TaxService) IsValidTaxForOneTimeBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	var (
		taxEntity entities.Tax
		item      utils.BillingItemData
	)
	taxEntity, err = s.getTax(ctx, db, orderItemData)
	item = orderItemData.BillItems[0]
	err = s.validateTaxWithBillItem(item, taxEntity)
	return
}

func (s *TaxService) IsValidTaxForRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (err error) {
	var taxEntity entities.Tax
	taxEntity, err = s.getTax(ctx, db, orderItemData)
	for _, item := range orderItemData.BillItems {
		err = s.validateTaxWithBillItem(item, taxEntity)
		if err != nil {
			return
		}
	}
	return
}

func (s *TaxService) IsValidTaxForCustomOrder(
	ctx context.Context,
	db database.QueryExecer,
	customBillingItem *pb.CustomBillingItem,
) (err error) {
	if customBillingItem.TaxItem == nil {
		return nil
	}

	tax, err := s.taxRepo.GetByIDForUpdate(ctx, db, customBillingItem.TaxItem.TaxId)
	if err != nil {
		err = status.Errorf(codes.Internal,
			"Error when retrieving tax id %v with error: %s",
			customBillingItem.TaxItem.TaxId,
			err.Error())
		return
	}

	if tax.TaxCategory.String != customBillingItem.TaxItem.TaxCategory.String() {
		err = status.Errorf(codes.FailedPrecondition,
			"Product with name %v changed tax category from %v to %v",
			customBillingItem.Name,
			customBillingItem.TaxItem.TaxCategory.String(),
			tax.TaxCategory.String)
		return
	}
	if customBillingItem.TaxItem.TaxCategory == pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE {
		err = status.Errorf(codes.FailedPrecondition, "This tax category %v is not supported in this version", pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String())
		return
	}
	if tax.TaxPercentage.Int != int32(customBillingItem.TaxItem.TaxPercentage) {
		err = status.Errorf(codes.FailedPrecondition,
			"Product with name %v change tax percentage from %v to %v",
			customBillingItem.Name,
			customBillingItem.TaxItem.TaxPercentage,
			tax.TaxPercentage.Int)
		return
	}

	taxAmount := float32(float64(customBillingItem.Price*customBillingItem.TaxItem.TaxPercentage) / float64(100+customBillingItem.TaxItem.TaxPercentage))
	if customBillingItem.TaxItem.TaxAmount != taxAmount {
		err = status.Errorf(codes.FailedPrecondition, "Incorrect tax amount actual = %v vs expected = %v", customBillingItem.TaxItem.TaxAmount, taxAmount)
	}
	return
}

func (s *TaxService) getTax(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (taxEntity entities.Tax, err error) {
	if orderItemData.ProductInfo.TaxID.Status != pgtype.Present {
		return
	}

	taxEntity, err = s.taxRepo.GetByIDForUpdate(ctx, db, orderItemData.ProductInfo.TaxID.String)
	if err != nil {
		err = status.Errorf(codes.Internal,
			"getting tax of product %v have error %s",
			orderItemData.ProductInfo.ProductID,
			err.Error())
	}
	return
}

func (s *TaxService) GetTaxByID(ctx context.Context, db database.QueryExecer, taxID string) (taxEntity entities.Tax, err error) {
	taxEntity, err = s.taxRepo.GetByIDForUpdate(ctx, db, taxID)
	if err != nil {
		err = status.Errorf(codes.Internal,
			"getting tax %v have error %s", taxID, err.Error())
	}
	return
}

func (s *TaxService) validateTaxWithBillItem(
	billingData utils.BillingItemData,
	tax entities.Tax,
) (err error) {
	var (
		billItem  *pb.BillingItem
		taxItem   *pb.TaxBillItem
		productID string
	)
	billItem = billingData.BillingItem
	if (billItem.TaxItem == nil && tax.TaxID.Status == pgtype.Present) ||
		(billItem.TaxItem != nil && tax.TaxID.Status != pgtype.Present) {
		err = utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InconsistentTax, nil)
		return
	}

	if billItem.TaxItem == nil {
		return
	}

	taxItem = billingData.BillingItem.TaxItem
	productID = billItem.ProductId
	if tax.TaxCategory.String != taxItem.TaxCategory.String() {
		err = status.Errorf(codes.FailedPrecondition,
			"Product with ID %v change tax category from %v to %v",
			productID,
			taxItem.TaxCategory.String(),
			tax.TaxCategory.String)
		return
	}
	if taxItem.TaxCategory == pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE {
		err = status.Errorf(codes.FailedPrecondition, "This tax category is not supported in this version")
		return
	}
	if tax.TaxPercentage.Int != int32(taxItem.TaxPercentage) {
		err = status.Errorf(codes.FailedPrecondition,
			"Product with ID %v change tax percentage from %v to %v",
			productID,
			taxItem.TaxPercentage,
			tax.TaxPercentage.Int)
		return
	}

	priceAfterDiscount := billItem.Price
	if billItem.DiscountItem != nil {
		priceAfterDiscount -= billItem.DiscountItem.DiscountAmount
	}
	tmpTaxAmount := float32(float64(priceAfterDiscount*billItem.TaxItem.TaxPercentage) / float64(100+billItem.TaxItem.TaxPercentage))
	if !utils.CompareAmountValue(taxItem.TaxAmount, tmpTaxAmount) {
		err = status.Errorf(codes.FailedPrecondition, "Incorrect tax amount actual = %v vs expected = %v", taxItem.TaxAmount, tmpTaxAmount)
	}
	return
}

func (s *TaxService) CalculatorTaxPrice(tax entities.Tax, price float32, billItem *entities.BillItem) (finalPrice float32, err error) {
	if tax.TaxCategory.String == pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String() {
		err = status.Errorf(codes.FailedPrecondition, "This tax category is not supported in this version")
		return
	}
	taxAmount := float32(tax.TaxPercentage.Int) * price / float32(100+tax.TaxPercentage.Int)
	err = multierr.Combine(
		billItem.TaxID.Set(tax.TaxID),
		billItem.TaxPercentage.Set(tax.TaxPercentage),
		billItem.TaxAmount.Set(taxAmount),
		billItem.TaxCategory.Set(tax.TaxCategory),
	)
	if err != nil {
		err = fmt.Errorf(
			"err occurred while assigning tax %s into orderID %s and productID %s : %v",
			tax.TaxID.String,
			billItem.OrderID.String,
			billItem.ProductID.String,
			err,
		)
	}
	finalPrice = price
	return
}

func NewTaxService() *TaxService {
	return &TaxService{
		taxRepo: &repositories.TaxRepo{},
	}
}
