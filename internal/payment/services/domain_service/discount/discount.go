package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiscountService struct {
	discountRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Discount, error)
		GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Discount, error)
	}
	productDiscountRepo interface {
		GetByProductIDAndDiscountID(ctx context.Context, db database.QueryExecer, productID string, discountID string) (entities.ProductDiscount, error)
	}
	userDiscountTagRepo interface {
		GetDiscountTagByUserIDAndDiscountTagID(ctx context.Context, db database.QueryExecer, userID string, discountTagID string) ([]*entities.UserDiscountTag, error)
		GetAvailableDiscountTagIDsByUserID(ctx context.Context, db database.QueryExecer, userID string) (usersDiscountTagIDs []string, err error)
	}
}

func (s *DiscountService) IsValidDiscountForOneTimeBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	discountName *string,
) (err error) {
	var (
		billingData utils.BillingItemData
		billItem    *pb.BillingItem
		discount    entities.Discount
	)
	billingData = orderItemData.BillItems[0]
	billItem = billingData.BillingItem
	if billItem.DiscountItem == nil {
		return
	}
	timeNow := time.Now()

	discount, err = s.getDiscountAndCheckProductDiscount(ctx, db, orderItemData)
	if err != nil {
		return
	}

	if discount.DiscountTagID.Status == pgtype.Present {
		isValid, errDiscountOrg := s.isValidDiscountOrgLevelDiscount(ctx, db, orderItemData.StudentInfo.StudentID.String, discount.DiscountTagID.String)
		if errDiscountOrg != nil {
			err = fmt.Errorf("error occurred while verify discount org level, err: %s", errDiscountOrg.Error())
			return
		}
		if !isValid {
			err = fmt.Errorf("unavailable/expired product discount org level")
			return
		}
	}
	_ = discount.Name.AssignTo(discountName)
	if discount.AvailableFrom.Status != pgtype.Present ||
		discount.AvailableUntil.Status != pgtype.Present ||
		discount.AvailableFrom.Time.After(timeNow) ||
		discount.AvailableUntil.Time.Before(timeNow) {
		err = utils.StatusErrWithDetail(codes.FailedPrecondition, constant.DiscountIsNotAvailable, nil)
		return
	}

	err = utils.GroupErrorFunc(
		s.checkCorrectDiscountInfoBetweenDiscountEntitiesAndBillItem(billingData, discount),
		s.checkCorrectDiscountAmountBetweenDiscountEntitiesAndBillItem(billingData, discount),
	)

	return
}

func (s *DiscountService) IsValidDiscountForRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	nonProRatedBillItems []utils.BillingItemData,
	discountName *string) (err error) {
	var (
		discountEntity entities.Discount
	)
	discountEntity, err = s.isRecurringValidDurationValid(ctx, db, orderItemData)
	if err != nil {
		return
	}

	if discountEntity.DiscountID.Status != pgtype.Present {
		return
	}

	if discountEntity.DiscountTagID.Status == pgtype.Present {
		isValid, errDiscountOrg := s.isValidDiscountOrgLevelDiscount(ctx, db, orderItemData.StudentInfo.StudentID.String, discountEntity.DiscountTagID.String)
		if errDiscountOrg != nil {
			err = fmt.Errorf("error occurred while verify discount org level, err: %s", errDiscountOrg.Error())
			return
		}
		if !isValid {
			err = fmt.Errorf("unavailable/expired product discount org level")
			return
		}
	}

	err = discountEntity.Name.AssignTo(discountName)
	if err != nil {
		err = status.Errorf(codes.Internal, "assigning discount name have error %v", err.Error())
		return
	}
	err = s.isDiscountOfNormalBillItemsValid(nonProRatedBillItems, discountEntity)
	if err != nil {
		return
	}

	if proRatedBillItem.BillingItem == nil {
		return
	}
	err = s.isDiscountOfProRatingBillItemsValid(proRatedBillItem, discountEntity, ratioOfProRatedBillingItem)

	return
}

func (s *DiscountService) isRecurringValidDurationValid(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	discountEntity entities.Discount,
	err error,
) {
	var (
		countDiscountTime int32
	)
	if orderItemData.OrderItem.DiscountId == nil {
		for _, item := range orderItemData.BillItems {
			if item.BillingItem.DiscountItem != nil {
				err = status.Errorf(codes.FailedPrecondition, "This bill item should not have a discount because order item does not have a discount id")
				return
			}
		}
		return
	}

	discountEntity, err = s.getDiscountAndCheckProductDiscount(ctx, db, orderItemData)
	if err != nil {
		return
	}

	if discountEntity.RecurringValidDuration.Status != pgtype.Present {
		return
	}

	for _, item := range orderItemData.BillItems {
		if item.BillingItem.DiscountItem != nil {
			countDiscountTime++
		}
	}

	if countDiscountTime > discountEntity.RecurringValidDuration.Int {
		err = status.Errorf(codes.FailedPrecondition, "Maximum discount is reached")
	}
	return
}

func (s *DiscountService) isDiscountOfNormalBillItemsValid(
	normalBillItems []utils.BillingItemData,
	discountEntities entities.Discount,
) (err error) {
	for _, item := range normalBillItems {
		if item.BillingItem.DiscountItem == nil {
			continue
		}

		if item.BillingItem.DiscountItem.DiscountId != discountEntities.DiscountID.String {
			err = utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InconsistentDiscountID,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf(
						constant.InconsistentDiscountIDDebugMsg,
						item.BillingItem.DiscountItem.DiscountId,
						discountEntities.DiscountID.String,
					),
				},
			)
			return
		}

		err = utils.GroupErrorFunc(
			s.checkCorrectDiscountInfoBetweenDiscountEntitiesAndBillItem(item, discountEntities),
			s.checkCorrectDiscountAmountBetweenDiscountEntitiesAndBillItem(item, discountEntities),
		)
		if err != nil {
			return
		}
	}
	return
}

func (s *DiscountService) checkCorrectDiscountInfoBetweenDiscountEntitiesAndBillItem(
	billItem utils.BillingItemData,
	discountEntities entities.Discount,
) (err error) {
	if discountEntities.AvailableFrom.Status != pgtype.Present ||
		discountEntities.AvailableUntil.Status != pgtype.Present ||
		discountEntities.AvailableFrom.Time.After(time.Now()) ||
		discountEntities.AvailableUntil.Time.Before(time.Now()) {
		err = utils.StatusErrWithDetail(codes.FailedPrecondition, constant.DiscountIsNotAvailable, nil)
		return
	}
	if discountEntities.DiscountType.String != billItem.BillingItem.DiscountItem.DiscountType.String() {
		err = status.Errorf(codes.FailedPrecondition,
			"Product with id %v change discount type from %s to %s",
			billItem.BillingItem.ProductId,
			billItem.BillingItem.DiscountItem.DiscountType,
			discountEntities.DiscountType.String)
		return
	}
	if discountEntities.DiscountAmountType.String != billItem.BillingItem.DiscountItem.DiscountAmountType.String() {
		err = status.Errorf(codes.FailedPrecondition,
			"Product with id %v change discount amount type from %s to %s",
			billItem.BillingItem.ProductId,
			billItem.BillingItem.DiscountItem.DiscountAmountType.String(),
			discountEntities.DiscountAmountType.String)
		return
	}
	if !utils.IsEqualNumericAndFloat32(discountEntities.DiscountAmountValue, billItem.BillingItem.DiscountItem.DiscountAmountValue) {
		err = status.Errorf(codes.FailedPrecondition,
			"Product with id %v change discount amount value from %s to %s",
			billItem.BillingItem.ProductId,
			discountEntities.DiscountAmountValue.Int,
			fmt.Sprintf("%v", billItem.BillingItem.DiscountItem.DiscountAmountValue))
	}
	return
}

func (s *DiscountService) checkCorrectDiscountAmountBetweenDiscountEntitiesAndBillItem(
	billItem utils.BillingItemData,
	discountEntities entities.Discount,
) (err error) {
	var tmpDiscountAmount float32
	if discountEntities.DiscountAmountType.String == pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String() {
		tmpDiscountAmount = billItem.BillingItem.Price * billItem.BillingItem.DiscountItem.DiscountAmountValue / float32(100)
		if !utils.CompareAmountValue(billItem.BillingItem.DiscountItem.DiscountAmount, tmpDiscountAmount) {
			err = utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, billItem.BillingItem.DiscountItem.DiscountAmount, tmpDiscountAmount)},
			)
		}
	} else if !utils.IsEqualNumericAndFloat32(discountEntities.DiscountAmountValue, billItem.BillingItem.DiscountItem.DiscountAmount) {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.DiscountAmountsAreNotEqual,
			&errdetails.DebugInfo{
				Detail: fmt.Sprintf(
					constant.DiscountAmountsAreNotEqualDebugMsg,
					billItem.BillingItem.DiscountItem.DiscountAmount,
					discountEntities.DiscountAmountValue.Int.String(),
				),
			})
	}
	return
}

func (s *DiscountService) checkCorrectDiscountAmountBetweenDiscountEntitiesAndProRatingBillItem(
	billItem utils.BillingItemData,
	discountEntities entities.Discount,
	ratioOfProRatedBillingItem entities.BillingRatio,
) (err error) {
	var tmpDiscountAmount float32
	if ratioOfProRatedBillingItem.BillingRatioNumerator.Int == 0 {
		if billItem.BillingItem.DiscountItem.DiscountAmount != 0 {
			err = utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, billItem.BillingItem.DiscountItem.DiscountAmount, tmpDiscountAmount)},
			)
		}
		return
	}
	if discountEntities.DiscountAmountType.String == pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String() {
		tmpDiscountAmount = billItem.BillingItem.Price * billItem.BillingItem.DiscountItem.DiscountAmountValue / float32(100)
		if !utils.CompareAmountValue(billItem.BillingItem.DiscountItem.DiscountAmount, tmpDiscountAmount) {
			err = utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, billItem.BillingItem.DiscountItem.DiscountAmount, tmpDiscountAmount)},
			)
		}
		return
	}

	var discountValue float32
	_ = discountEntities.DiscountAmountValue.AssignTo(&discountValue)
	tmpDiscountAmount = discountValue * float32(ratioOfProRatedBillingItem.BillingRatioNumerator.Int) / float32(ratioOfProRatedBillingItem.BillingRatioDenominator.Int)
	if !utils.CompareAmountValue(billItem.BillingItem.DiscountItem.DiscountAmount, tmpDiscountAmount) {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.DiscountAmountsAreNotEqual,
			&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, billItem.BillingItem.DiscountItem.DiscountAmount, tmpDiscountAmount)},
		)
	}
	return
}

func (s *DiscountService) getDiscountAndCheckProductDiscount(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (discount entities.Discount, err error) {
	discount, err = s.discountRepo.GetByIDForUpdate(ctx, db, orderItemData.OrderItem.DiscountId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal,
			"Error when get discount of product %v with error %s",
			orderItemData.ProductInfo.ProductID.String,
			err.Error())
		return
	}

	if discount.DiscountTagID.Status == pgtype.Present {
		_, err = s.userDiscountTagRepo.GetDiscountTagByUserIDAndDiscountTagID(
			ctx,
			db,
			orderItemData.StudentInfo.StudentID.String,
			discount.DiscountTagID.String,
		)
	} else {
		_, err = s.productDiscountRepo.GetByProductIDAndDiscountID(
			ctx,
			db,
			orderItemData.ProductInfo.ProductID.String,
			orderItemData.OrderItem.DiscountId.Value,
		)
	}

	if err != nil {
		err = status.Errorf(codes.Internal,
			"Product %v and discount %v have non-association",
			orderItemData.ProductInfo.ProductID.String,
			orderItemData.OrderItem.DiscountId.Value,
		)
		return
	}

	return
}

func (s *DiscountService) isDiscountOfProRatingBillItemsValid(
	proRatingBillItem utils.BillingItemData,
	discountEntities entities.Discount,
	ratioOfProRatedBillingItem entities.BillingRatio,
) (err error) {
	if proRatingBillItem.BillingItem.DiscountItem.DiscountId != discountEntities.DiscountID.String {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.InconsistentDiscountID,
			&errdetails.DebugInfo{
				Detail: fmt.Sprintf(
					constant.InconsistentDiscountIDDebugMsg,
					proRatingBillItem.BillingItem.DiscountItem.DiscountId,
					discountEntities.DiscountID.String,
				),
			},
		)
		return
	}

	err = utils.GroupErrorFunc(
		s.checkCorrectDiscountInfoBetweenDiscountEntitiesAndBillItem(proRatingBillItem, discountEntities),
		s.checkCorrectDiscountAmountBetweenDiscountEntitiesAndProRatingBillItem(proRatingBillItem, discountEntities, ratioOfProRatedBillingItem),
	)
	return
}

func (s *DiscountService) isValidDiscountOrgLevelDiscount(ctx context.Context, db database.QueryExecer, userID, discountTagID string) (isValid bool, err error) {
	isValid = false
	usersDiscountTagIDs, err := s.userDiscountTagRepo.GetAvailableDiscountTagIDsByUserID(ctx, db, userID)
	if err != nil {
		err = fmt.Errorf("error while getting available user's discount tag: %s", err.Error())
		return
	}
	for id := range usersDiscountTagIDs {
		if usersDiscountTagIDs[id] == discountTagID {
			isValid = true
			return
		}
	}
	return
}

func (s *DiscountService) GetDiscountsByDiscountIDs(ctx context.Context, db database.Ext, discountIDs []string) (discounts []entities.Discount, err error) {
	discounts, err = s.discountRepo.GetByIDs(ctx, db, discountIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get discounts by ids: %v", err.Error())
	}
	return
}

func (s *DiscountService) VerifyDiscountForGenerateUpcomingBillItem(ctx context.Context, db database.QueryExecer, billItems []entities.BillItem) (discount entities.Discount, err error) {
	latestBillItem := billItems[0]
	if latestBillItem.DiscountID.Status != pgtype.Present {
		return
	}

	discount, err = s.discountRepo.GetByIDForUpdate(ctx, db, latestBillItem.DiscountID.String)
	if err != nil {
		err = fmt.Errorf("error when get discount of product %v with error %s",
			latestBillItem.DiscountID.String,
			err.Error())
		return
	}
	if discount.RecurringValidDuration.Status == pgtype.Present {
		countDiscountTime := int32(1)
		for index := range billItems {
			if billItems[index].DiscountID == latestBillItem.DiscountID {
				countDiscountTime++
			}
		}
		if countDiscountTime > discount.RecurringValidDuration.Int {
			discount = entities.Discount{}
		}
	}
	if discount.AvailableFrom.Status != pgtype.Present ||
		discount.AvailableUntil.Status != pgtype.Present ||
		discount.AvailableFrom.Time.After(time.Now()) ||
		discount.AvailableUntil.Time.Before(time.Now()) {
		discount = entities.Discount{}
	}
	return
}

func (s *DiscountService) CalculatorDiscountPrice(
	discount entities.Discount,
	price float32,
	billItem *entities.BillItem,
) (finalPrice float32, err error) {
	var tmpPrice float32
	tmpPrice = price
	if discount.DiscountAmountType.String == pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String() {
		discountAmount := price * utils.ConvertNumericToFloat32(discount.DiscountAmountValue) / float32(100)
		tmpPrice -= discountAmount
		err = multierr.Combine(
			billItem.DiscountID.Set(discount.DiscountID),
			billItem.DiscountAmountValue.Set(utils.ConvertNumericToFloat32(discount.DiscountAmountValue)),
			billItem.DiscountAmountType.Set(discount.DiscountAmountType),
			billItem.DiscountAmount.Set(discountAmount),
			billItem.RawDiscountAmount.Set(discountAmount),
		)
		if err != nil {
			err = fmt.Errorf(
				"err occurred while assigning discount %s into orderID %s and productID %s : %v",
				discount.DiscountID.String,
				billItem.OrderID.String,
				billItem.ProductID.String,
				err,
			)
		}
	} else {
		tmpPrice -= utils.ConvertNumericToFloat32(discount.DiscountAmountValue)
		err = multierr.Combine(
			billItem.DiscountID.Set(discount.DiscountID),
			billItem.DiscountAmountValue.Set(utils.ConvertNumericToFloat32(discount.DiscountAmountValue)),
			billItem.DiscountAmountType.Set(discount.DiscountAmountType),
			billItem.DiscountAmount.Set(utils.ConvertNumericToFloat32(discount.DiscountAmountValue)),
			billItem.RawDiscountAmount.Set(utils.ConvertNumericToFloat32(discount.DiscountAmountValue)),
		)
		if err != nil {
			err = fmt.Errorf(
				"err occurred while assigning discount %s into orderID %s and productID %s : %v",
				discount.DiscountID.String,
				billItem.OrderID.String,
				billItem.ProductID.String,
				err,
			)
		}
	}
	finalPrice = tmpPrice
	return
}

func NewDiscountService() *DiscountService {
	return &DiscountService{
		discountRepo:        &repositories.DiscountRepo{},
		productDiscountRepo: &repositories.ProductDiscountRepo{},
		userDiscountTagRepo: &repositories.UserDiscountTagRepo{},
	}
}
