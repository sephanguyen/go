package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	exportEntities "github.com/manabie-com/backend/internal/payment/export_entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ConvertBillItemError          = "converting billing item description with error %v"
	CreateBillItemError           = "creating billing item with error %v"
	GenerateBillItemUpdateError   = "generating bill item entities for update without adjustment price"
	SettingNonLatestBillItemError = "setting non latest bill item for student product entities with error %v"
)

type BillItemService struct {
	BillItemRepo interface {
		Create(ctx context.Context, db database.QueryExecer, billItem *entities.BillItem) (pgtype.Int4, error)
		SetNonLatestBillItemByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (err error)
		GetLatestBillItemByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (billItem entities.BillItem, err error)
		GetBillItemByStudentProductIDAndPeriodID(ctx context.Context, db database.QueryExecer, studentProductID string, periodID string) (billItem entities.BillItem, err error)
		UpdateReviewFlagByOrderID(ctx context.Context, db database.QueryExecer, orderID string, isReview bool) (err error)
		VoidBillItemByOrderID(ctx context.Context, db database.QueryExecer, orderID string, status string) (err error)
		UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID(ctx context.Context, db database.QueryExecer, billItemSequenceNumber int32, status string) (orderID string, err error)
		GetRecurringBillItemsForScheduledGenerationOfNextBillItems(ctx context.Context, db database.QueryExecer) ([]*entities.BillItem, error)
		GetBillItemByOrderIDAndPaging(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
			from int64,
			limit int64,
		) (
			billItems []*entities.BillItem,
			err error,
		)
		CountBillItemByOrderID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
		) (
			total int,
			err error,
		)
		GetBillItemByStudentIDAndLocationIDsPaging(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
			locationIDs []string,
			from int64,
			limit int64,
		) (
			billItems []*entities.BillItem,
			err error,
		)
		CountBillItemByStudentIDAndLocationIDs(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
			locationIDs []string,
		) (
			total int,
			err error,
		)
		GetBillItemInfoByOrderIDAndUniqueByProductID(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
		) (
			billItems []*entities.BillItem,
			err error,
		)
		GetAllFirstBillItemDistinctByOrderIDAndUniqueByProductID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
		) (
			billItems []*entities.BillItem,
			err error,
		)
		GetLatestBillItemByStudentProductIDForStudentBilling(
			ctx context.Context,
			db database.QueryExecer,
			studentProductID string,
		) (
			billItems entities.BillItem,
			err error,
		)
		GetExportStudentBilling(ctx context.Context, db database.QueryExecer, locationIDs []string) (billItems []*entities.BillItem, studentIDs []string, err error)
		GetPastBillItemsByStudentProductIDs(ctx context.Context, db database.QueryExecer, studentProductIDs []string, studentID string) ([]*entities.BillItem, error)
		GetPresentAndFutureBillItemsByStudentProductIDs(ctx context.Context, db database.QueryExecer, studentProductIDs []string, studentID string) ([]*entities.BillItem, error)
		GetUpcomingBillingByStudentProductID(
			ctx context.Context, db database.QueryExecer, studentProductID string, studentID string) (
			*entities.BillItem, error)
		GetByOrderIDAndProductIDs(
			ctx context.Context, db database.QueryExecer, orderID string, productIDs []string) (
			[]entities.BillItem, error)
		GetBillItemsByOrderIDAndProductID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
			productID string) (
			billItems []entities.BillItem, err error)
	}
	BillItemCourseRepo interface {
		MultiCreate(ctx context.Context, db database.QueryExecer, course []*entities.BillItemCourse, billItemSequenceNumber int32) error
	}
	BillItemAccountCategoryRepo interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, billItemAccountCategories []*entities.BillItemAccountCategory) (err error)
	}
	MaterialRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, materialID string) (entities.Material, error)
	}
	UserRepo interface {
		GetStudentsByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]entities.User, error)
	}
	StudentRepo interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]*entities.Student, error)
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, studentID string) (entities.Student, error)
	}
	GradeRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Grade, error)
	}
	UpcomingBillItemRepo interface {
		RemoveOldUpcomingBillItem(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
			productID string,

		) (
			billingSchedulePeriodID string,
			billingDate time.Time,
			err error,
		)
		Create(ctx context.Context, db database.QueryExecer, e *entities.UpcomingBillItem) (err error)
		GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
			productID string,
			billingSchedulePeriodID string,
		) (upcomingBillItems []entities.UpcomingBillItem, err error)
		UpdateCurrentUpcomingBillItemStatus(
			ctx context.Context,
			db database.QueryExecer,
			upcomingBillItem entities.UpcomingBillItem,
		) (err error)
	}
	OrderRepo interface {
		GetOrderByIDForUpdate(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
		) (order entities.Order, err error)
	}
	ProductRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, productID string) (entities.Product, error)
	}
	PriceRepo interface {
	}
}

func (s *BillItemService) GetRecurringBillItemsByOrderIDAndProductID(ctx context.Context, db database.QueryExecer, orderID string, productID string) (billItems []entities.BillItem, err error) {
	billItems, err = s.BillItemRepo.GetBillItemsByOrderIDAndProductID(ctx, db, orderID, productID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error occurred while retrieving bill items: %v", err)
	}
	return
}

func (s *BillItemService) CreateNewBillItemForOneTimeBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	discountName string,
) (err error) {
	var (
		billItem *entities.BillItem
	)
	billItem, err = s.generateBillingEntitiesForOneTimeProduct(orderItemData)
	if err != nil {
		err = status.Errorf(codes.Internal, "generating bill item entities with error %v", err.Error())
		return
	}
	return s.generateBillingDescriptionProductAndCreateData(ctx, db, orderItemData, discountName, billItem, nil, nil)
}

func (s *BillItemService) generateBillingDescriptionProductAndCreateData(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	discountName string,
	billItem *entities.BillItem,
	billRatio *entities.BillingRatio,
	periodInfo *entities.BillingSchedulePeriod,
) (err error) {
	switch orderItemData.ProductType {
	case pb.ProductType_PRODUCT_TYPE_MATERIAL:
		return s.generateBillingDescriptionForMaterialAndCreateData(ctx, db, orderItemData, *billItem, billItem.BillDate.Time, discountName, billRatio, periodInfo)
	case pb.ProductType_PRODUCT_TYPE_FEE:
		return s.generateBillingDescriptionForFeeAndCreateData(ctx, db, orderItemData, *billItem, discountName, billRatio, periodInfo)
	case pb.ProductType_PRODUCT_TYPE_PACKAGE:
		return s.generateBillingDescriptionForPackageAndCreateData(ctx, db, orderItemData, *billItem, discountName, billRatio, periodInfo)
	}
	return
}

func (s *BillItemService) generateBillingEntitiesForOneTimeProduct(
	orderItemData utils.OrderItemData,
) (billItem *entities.BillItem, err error) {
	billItemReq := orderItemData.BillItems[0].BillingItem
	billItem = &entities.BillItem{}
	billItemStatus := pb.BillingStatus_BILLING_STATUS_BILLED
	billType := pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER
	billDate := time.Now()

	err = multierr.Combine(
		billItem.StudentID.Set(orderItemData.Order.StudentID.String),
		billItem.ProductID.Set(billItemReq.ProductId),
		billItem.ProductDescription.Set(orderItemData.ProductInfo.Name.String),
		billItem.ProductPricing.Set(billItemReq.Price),
		billItem.BillSchedulePeriodID.Set(nil),
		billItem.OrderID.Set(orderItemData.Order.OrderID.String),
		billItem.BillDate.Set(billDate),
		billItem.BillStatus.Set(billItemStatus.String()),
		billItem.BillType.Set(billType.String()),
		billItem.BillFrom.Set(nil),
		billItem.BillTo.Set(nil),
		billItem.FinalPrice.Set(billItemReq.FinalPrice),
		billItem.BillItemSequenceNumber.Set(nil),
		billItem.StudentProductID.Set(orderItemData.StudentProduct.StudentProductID.String),
		billItem.BillApprovalStatus.Set(nil),
		billItem.LocationID.Set(orderItemData.Order.LocationID.String),
		billItem.LocationName.Set(orderItemData.LocationName),
		billItem.Price.Set(billItemReq.Price),
		billItem.BillingRatioNumerator.Set(nil),
		billItem.BillingRatioDenominator.Set(nil),
		billItem.IsLatestBillItem.Set(true),
		billItem.PreviousBillItemStatus.Set(nil),
		billItem.PreviousBillItemSequenceNumber.Set(nil),
		billItem.AdjustmentPrice.Set(nil),
		billItem.OldPrice.Set(0),
		billItem.DiscountID.Set(nil),
		billItem.DiscountAmountValue.Set(0),
		billItem.DiscountAmountType.Set(nil),
		billItem.DiscountAmount.Set(0),
		billItem.RawDiscountAmount.Set(0),
		billItem.TaxCategory.Set(nil),
		billItem.TaxPercentage.Set(nil),
		billItem.TaxID.Set(nil),
		billItem.TaxAmount.Set(nil),
	)
	if err != nil {
		return
	}

	if billItemReq.TaxItem != nil {
		err = multierr.Combine(
			billItem.TaxCategory.Set(billItemReq.TaxItem.TaxCategory),
			billItem.TaxPercentage.Set(billItemReq.TaxItem.TaxPercentage),
			billItem.TaxID.Set(billItemReq.TaxItem.TaxId),
			billItem.TaxAmount.Set(billItemReq.TaxItem.TaxAmount),
		)
	}
	if err != nil {
		return
	}

	if billItemReq.DiscountItem != nil {
		roundDiscount := float32(math.Round(float64(billItemReq.DiscountItem.DiscountAmount)))
		finalPriceBeforeDiscount := billItemReq.FinalPrice + billItemReq.DiscountItem.DiscountAmount
		finalPriceAfterRoundDiscount := finalPriceBeforeDiscount - roundDiscount
		err = multierr.Combine(
			billItem.DiscountID.Set(billItemReq.DiscountItem.DiscountId),
			billItem.DiscountAmountValue.Set(billItemReq.DiscountItem.DiscountAmountValue),
			billItem.DiscountAmountType.Set(billItemReq.DiscountItem.DiscountAmountType),
			billItem.DiscountAmount.Set(roundDiscount),
			billItem.RawDiscountAmount.Set(billItemReq.DiscountItem.DiscountAmount),
			billItem.FinalPrice.Set(finalPriceAfterRoundDiscount),
		)
	}
	if err != nil {
		return
	}
	return
}

func (s *BillItemService) generateBillingDescriptionForMaterialAndCreateData(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	billItem entities.BillItem,
	billingDate time.Time,
	discountName string,
	billRatio *entities.BillingRatio,
	periodInfo *entities.BillingSchedulePeriod,
) (
	err error,
) {
	var (
		material               entities.Material
		billItemStatus         string
		billingItemDescription entities.BillingItemDescription
		billType               string
	)
	billItemStatus = billItem.BillStatus.String
	billType = billItem.BillType.String

	materialType := pb.MaterialType_MATERIAL_TYPE_RECURRING.String()

	if orderItemData.IsOneTimeProduct {
		materialType = pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String()
		material, err = s.MaterialRepo.GetByIDForUpdate(ctx, db, orderItemData.ProductInfo.ProductID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "Error retrieving material product with ID %v: %v", orderItemData.ProductInfo.ProductID.String, err.Error())
			return
		}

		if material.CustomBillingDate.Status != pgtype.Null {
			if billingDate.Before(material.CustomBillingDate.Time) && orderItemData.Order.OrderType.String == pb.OrderType_ORDER_TYPE_NEW.String() {
				billItemStatus = pb.BillingStatus_BILLING_STATUS_PENDING.String()
				billType = pb.BillingType_BILLING_TYPE_UPCOMING_BILLING.String()
			}
			billingDate = material.CustomBillingDate.Time
		}
	}

	billingItemDescription = entities.BillingItemDescription{
		ProductID:               orderItemData.ProductInfo.ProductID.String,
		ProductName:             orderItemData.ProductInfo.Name.String,
		ProductType:             orderItemData.ProductInfo.ProductType.String,
		MaterialType:            &materialType,
		DiscountName:            &discountName,
		BillingRatioNumerator:   nil,
		BillingRatioDenominator: nil,
		GradeID:                 orderItemData.StudentInfo.GradeID.String,
		GradeName:               orderItemData.GradeName,
	}

	if billRatio != nil {
		billingItemDescription.BillingRatioNumerator = &billRatio.BillingRatioNumerator.Int
		billingItemDescription.BillingRatioDenominator = &billRatio.BillingRatioDenominator.Int
	}
	if periodInfo != nil {
		billingItemDescription.BillingPeriodName = &periodInfo.Name.String
	}
	err = multierr.Combine(
		billItem.BillingItemDescription.Set(billingItemDescription),
		billItem.BillStatus.Set(billItemStatus),
		billItem.BillType.Set(billType),
		billItem.BillDate.Set(billingDate),
	)
	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, ConvertBillItemError, err.Error())
		return
	}

	_, err = s.BillItemRepo.Create(ctx, db, &billItem)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, CreateBillItemError, err.Error())
	}
	return
}

func (s *BillItemService) generateBillingDescriptionForFeeAndCreateData(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	billItem entities.BillItem,
	discountName string,
	billRatio *entities.BillingRatio,
	periodInfo *entities.BillingSchedulePeriod,
) (
	err error,
) {
	feeType := pb.FeeType_FEE_TYPE_ONE_TIME.String()
	if !orderItemData.IsOneTimeProduct {
		feeType = pb.FeeType_FEE_TYPE_RECURRING.String()
	}
	billingItemDescription := entities.BillingItemDescription{
		ProductID:               orderItemData.ProductInfo.ProductID.String,
		ProductName:             orderItemData.ProductInfo.Name.String,
		ProductType:             orderItemData.ProductInfo.ProductType.String,
		FeeType:                 &feeType,
		DiscountName:            &discountName,
		BillingRatioNumerator:   nil,
		BillingRatioDenominator: nil,
		GradeID:                 orderItemData.StudentInfo.GradeID.String,
		GradeName:               orderItemData.GradeName,
	}

	if billRatio != nil {
		billingItemDescription.BillingRatioNumerator = &billRatio.BillingRatioNumerator.Int
		billingItemDescription.BillingRatioDenominator = &billRatio.BillingRatioDenominator.Int
	}

	if periodInfo != nil {
		billingItemDescription.BillingPeriodName = &periodInfo.Name.String
	}

	err = multierr.Combine(
		billItem.BillingItemDescription.Set(billingItemDescription),
	)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, ConvertBillItemError, err.Error())
	}

	_, err = s.BillItemRepo.Create(ctx, db, &billItem)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, CreateBillItemError, err.Error())
	}

	return
}

func (s *BillItemService) generateBillingDescriptionForPackageAndCreateData(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	billItem entities.BillItem,
	discountName string,
	billRatio *entities.BillingRatio,
	periodInfo *entities.BillingSchedulePeriod,
) (
	err error,
) {
	var billItemCourses []*entities.BillItemCourse
	currentBillItemData := orderItemData.BillItems[0]
	packageInfo := orderItemData.PackageInfo
	var (
		courseInfos           []*entities.CourseItem
		billingSequenceNumber pgtype.Int4
	)
	if len(currentBillItemData.BillingItem.CourseItems) == 0 {
		d := &errdetails.DebugInfo{Detail: fmt.Sprintf("MissingCourseInfoBillItem for Package: %s", packageInfo.Package.ProductID.String)}
		err = utils.StatusErrWithDetail(codes.FailedPrecondition, constant.MissingCourseInfoBillItem, d)

		return err
	}
	for _, item := range currentBillItemData.BillingItem.CourseItems {
		courseInfoInOrder := packageInfo.MapCourseInfo[item.CourseId]
		if packageInfo.QuantityType == pb.QuantityType_QUANTITY_TYPE_SLOT ||
			packageInfo.QuantityType == pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK {
			if item.Slot == nil {
				err = status.Errorf(codes.FailedPrecondition, "missing slot in course info %v of bill item", item.CourseId)
				return
			}
			if courseInfoInOrder.Slot.Value != item.Slot.Value {
				err = status.Errorf(codes.FailedPrecondition, "inconsistency course info %v between order item and bill item", item.CourseId)
				return
			}
			slotValue := item.Slot.Value
			courseInfos = append(courseInfos, &entities.CourseItem{
				CourseName: item.CourseName,
				CourseID:   item.CourseId,
				Slot:       &slotValue,
			})
			billItemCourse := entities.BillItemCourse{}
			_ = billItemCourse.CourseID.Set(item.CourseId)
			_ = billItemCourse.CourseSlot.Set(slotValue)
			_ = billItemCourse.CourseWeight.Set(nil)
			_ = billItemCourse.CourseName.Set(item.CourseName)
			_ = billItemCourse.CreatedAt.Set(time.Now())
			billItemCourses = append(billItemCourses, &billItemCourse)
		} else {
			if item.Weight == nil {
				err = status.Errorf(codes.FailedPrecondition, "missing weight in course info %v of bill item", item.CourseId)
				return
			}
			if courseInfoInOrder.Weight.Value != item.Weight.Value {
				err = status.Errorf(codes.FailedPrecondition, "inconsistency course info %v between order item and bill item", item.CourseId)
				return
			}
			weightValue := item.Weight.Value
			courseInfos = append(courseInfos, &entities.CourseItem{
				CourseName: item.CourseName,
				CourseID:   item.CourseId,
				Weight:     &weightValue,
			})
			billItemCourse := entities.BillItemCourse{}
			_ = billItemCourse.CourseID.Set(item.CourseId)
			_ = billItemCourse.CourseSlot.Set(nil)
			_ = billItemCourse.CourseWeight.Set(weightValue)
			_ = billItemCourse.CourseName.Set(item.CourseName)
			_ = billItemCourse.CreatedAt.Set(time.Now())
			billItemCourses = append(billItemCourses, &billItemCourse)
		}
	}
	packageType := orderItemData.PackageInfo.Package.PackageType.String
	quantityType := orderItemData.PackageInfo.QuantityType.String()
	billingItemDescription := entities.BillingItemDescription{
		ProductID:               orderItemData.ProductInfo.ProductID.String,
		ProductName:             orderItemData.ProductInfo.Name.String,
		ProductType:             orderItemData.ProductInfo.ProductType.String,
		PackageType:             &packageType,
		QuantityType:            &quantityType,
		DiscountName:            &discountName,
		BillingRatioNumerator:   nil,
		BillingRatioDenominator: nil,
		CourseItems:             courseInfos,
		GradeID:                 orderItemData.StudentInfo.GradeID.String,
		GradeName:               orderItemData.GradeName,
	}

	if billRatio != nil {
		billingItemDescription.BillingRatioNumerator = &billRatio.BillingRatioNumerator.Int
		billingItemDescription.BillingRatioDenominator = &billRatio.BillingRatioDenominator.Int
	}

	if periodInfo != nil {
		billingItemDescription.BillingPeriodName = &periodInfo.Name.String
	}

	err = multierr.Combine(
		billItem.BillingItemDescription.Set(billingItemDescription),
	)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, ConvertBillItemError, err.Error())
		return
	}

	billingSequenceNumber, err = s.BillItemRepo.Create(ctx, db, &billItem)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, CreateBillItemError, err.Error())
		return
	}

	err = s.BillItemCourseRepo.MultiCreate(ctx, db, billItemCourses, billingSequenceNumber.Int)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, "creating billing item course with error %v", err.Error())
	}
	return
}

func (s *BillItemService) CreateUpdateBillItemForOneTimeBilling(
	ctx context.Context,
	db database.QueryExecer,
	oldBillItem entities.BillItem,
	orderItemData utils.OrderItemData,
	discountName string,
) (err error) {
	var (
		billItem               *entities.BillItem
		billItemReq            *pb.BillingItem
		oldPrice               float32
		oldBillItemDescription *entities.BillingItemDescription
	)

	billItem, err = s.generateBillingEntitiesForOneTimeProduct(orderItemData)
	if err != nil {
		err = status.Errorf(codes.Internal, "generating bill item entities with error %v", err.Error())
		return
	}

	billItemReq = orderItemData.BillItems[0].BillingItem
	if billItemReq.AdjustmentPrice == nil {
		err = status.Errorf(codes.Internal, GenerateBillItemUpdateError)
		return
	}

	err = multierr.Combine(
		billItem.BillStatus.Set(oldBillItem.BillStatus.String),
		oldBillItem.Price.AssignTo(&oldPrice),
		billItem.AdjustmentPrice.Set(billItemReq.AdjustmentPrice.Value),
		billItem.OldPrice.Set(oldPrice),
		billItem.BillType.Set(pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
	)

	if err != nil {
		err = status.Errorf(codes.Internal, "assigning old price and adjustment price for bill item entities with error %v", err.Error())
		return
	}
	oldBillItemDescription, err = oldBillItem.GetBillingItemDescription()
	if err != nil {
		return
	}
	orderItemData.GradeName = oldBillItemDescription.GradeName

	err = s.BillItemRepo.SetNonLatestBillItemByStudentProductID(ctx, db, billItem.StudentProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, SettingNonLatestBillItemError, err.Error())
		return
	}

	switch orderItemData.ProductType {
	case pb.ProductType_PRODUCT_TYPE_MATERIAL:
		return s.generateBillingDescriptionForMaterialAndCreateData(ctx, db, orderItemData, *billItem, billItem.BillDate.Time, discountName, nil, nil)
	case pb.ProductType_PRODUCT_TYPE_FEE:
		return s.generateBillingDescriptionForFeeAndCreateData(ctx, db, orderItemData, *billItem, discountName, nil, nil)
	case pb.ProductType_PRODUCT_TYPE_PACKAGE:
		return s.generateBillingDescriptionForPackageAndCreateData(ctx, db, orderItemData, *billItem, discountName, nil, nil)
	}
	return
}

func (s *BillItemService) CreateCancelBillItemForOneTimeBilling(
	ctx context.Context,
	db database.QueryExecer,
	oldBillItem entities.BillItem,
) (err error) {
	var (
		billItem      entities.BillItem
		oldPrice      float32
		oldFinalPrice float32
	)
	billItem = oldBillItem
	err = multierr.Combine(
		oldBillItem.OldPrice.AssignTo(&oldPrice),
		oldBillItem.FinalPrice.AssignTo(&oldFinalPrice),
		billItem.OldPrice.Set(oldPrice),
		billItem.AdjustmentPrice.Set(-oldFinalPrice),
		billItem.FinalPrice.Set(oldFinalPrice),
		billItem.BillType.Set(pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
		billItem.BillStatus.Set(oldBillItem.BillStatus.String),
	)
	if err != nil {
		err = status.Errorf(codes.Internal, "assigning old price to bill item with error %v", err.Error())
		return
	}

	err = s.BillItemRepo.SetNonLatestBillItemByStudentProductID(ctx, db, billItem.StudentProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, SettingNonLatestBillItemError, err.Error())
		return
	}

	_, err = s.BillItemRepo.Create(ctx, db, &billItem)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, CreateBillItemError, err.Error())
	}

	return
}

func (s *BillItemService) GetOldBillItemForUpdateOneTimeBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (billItem entities.BillItem, err error) {
	billItem, err = s.BillItemRepo.GetLatestBillItemByStudentProductID(ctx, db, orderItemData.StudentProduct.StudentProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting latest bill item by student product id with err %v", err.Error())
	}
	return
}

func (s *BillItemService) CreateNewBillItemForRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	proRatedPrice entities.ProductPrice,
	ratioOfProRatedBillingItem entities.BillingRatio,
	nonProRatedBillItems []utils.BillingItemData,
	mapPeriodInfoWithID map[string]entities.BillingSchedulePeriod,
	discountName string,
) (err error) {
	var upcomingBillItemEntity *entities.BillItem
	if proRatedBillItem.BillingItem != nil {
		var proRatedBillItemEntity *entities.BillItem
		proRatingPeriod := mapPeriodInfoWithID[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]

		proRatedBillItemEntity, err = s.generateBillingEntitiesForProRatingRecurringProduct(orderItemData, proRatedBillItem, proRatingPeriod, ratioOfProRatedBillingItem, proRatedPrice)
		if err != nil {
			return
		}

		err = s.generateBillingDescriptionProductAndCreateData(ctx, db, orderItemData, discountName, proRatedBillItemEntity, &ratioOfProRatedBillingItem, &proRatingPeriod)
		if err != nil {
			return
		}
		if proRatedBillItem.IsUpcoming {
			upcomingBillItemEntity = proRatedBillItemEntity
		}
	}

	for _, data := range nonProRatedBillItems {
		var (
			billItemEntity *entities.BillItem
		)
		period := mapPeriodInfoWithID[data.BillingItem.BillingSchedulePeriodId.Value]
		billItemEntity, err = s.generateBillingEntitiesForNormalRecurringProduct(orderItemData, data, period)
		if err != nil {
			return
		}

		err = s.generateBillingDescriptionProductAndCreateData(ctx, db, orderItemData, discountName, billItemEntity, nil, &period)
		if err != nil {
			return
		}

		if data.IsUpcoming {
			upcomingBillItemEntity = billItemEntity
		}
	}
	if upcomingBillItemEntity != nil {
		err = s.createUpcomingBillItem(ctx, db, upcomingBillItemEntity)
	}
	return
}

func (s *BillItemService) CreateUpdateBillItemForRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	proRatedPrice entities.ProductPrice,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapPeriodInfo map[string]entities.BillingSchedulePeriod,
	mapOldBillingItem map[string]entities.BillItem,
	discountName string,
) (err error) {
	var (
		upcomingBillingItemEntity *entities.BillItem
	)
	if proRatedBillItem.BillingItem != nil {
		var (
			proRatedBillItemEntity *entities.BillItem
			oldBillItem            entities.BillItem
			oldPrice               float32
			oldBillItemDescription *entities.BillingItemDescription
		)
		oldBillItem = mapOldBillingItem[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]
		proRatingPeriod := mapPeriodInfo[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]
		proRatedBillItemEntity, err = s.generateBillingEntitiesForProRatingRecurringProduct(orderItemData, proRatedBillItem, proRatingPeriod, ratioOfProRatedBillingItem, proRatedPrice)
		if err != nil {
			return
		}
		if proRatedBillItem.BillingItem.AdjustmentPrice == nil {
			err = status.Errorf(codes.Internal, GenerateBillItemUpdateError)
			return
		}

		err = multierr.Combine(
			oldBillItem.Price.AssignTo(&oldPrice),
			proRatedBillItemEntity.AdjustmentPrice.Set(proRatedBillItem.BillingItem.AdjustmentPrice.Value),
			proRatedBillItemEntity.OldPrice.Set(oldPrice),
			proRatedBillItemEntity.BillType.Set(pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
		)
		if err != nil {
			return
		}
		oldBillItemDescription, err = oldBillItem.GetBillingItemDescription()
		if err != nil {
			return
		}
		orderItemData.GradeName = oldBillItemDescription.GradeName
		err = s.generateBillingDescriptionProductAndCreateData(ctx, db, orderItemData, discountName, proRatedBillItemEntity, &ratioOfProRatedBillingItem, &proRatingPeriod)
		if err != nil {
			return
		}
		if proRatedBillItem.IsUpcoming {
			upcomingBillingItemEntity = proRatedBillItemEntity
		}
	}

	for _, data := range normalBillItem {
		var (
			billItemEntity *entities.BillItem
			oldBillItem    entities.BillItem
			oldPrice       float32
		)
		oldBillItem = mapOldBillingItem[data.BillingItem.BillingSchedulePeriodId.Value]
		period := mapPeriodInfo[data.BillingItem.BillingSchedulePeriodId.Value]
		billItemEntity, err = s.generateBillingEntitiesForNormalRecurringProduct(orderItemData, data, period)
		if err != nil {
			return
		}

		err = multierr.Combine(
			oldBillItem.Price.AssignTo(&oldPrice),
			billItemEntity.AdjustmentPrice.Set(data.BillingItem.AdjustmentPrice.Value),
			billItemEntity.OldPrice.Set(oldPrice),
			billItemEntity.BillType.Set(pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
		)

		if err != nil {
			return
		}

		err = s.generateBillingDescriptionProductAndCreateData(ctx, db, orderItemData, discountName, billItemEntity, nil, &period)
		if err != nil {
			return
		}
		if data.IsUpcoming {
			upcomingBillingItemEntity = billItemEntity
		}
	}
	if upcomingBillingItemEntity != nil {
		err = s.createUpcomingBillItem(ctx, db, upcomingBillingItemEntity)
	}
	return
}

func (s *BillItemService) createUpcomingBillItem(
	ctx context.Context,
	db database.QueryExecer,
	upcomingBillingItemEntity *entities.BillItem,
) (err error) {
	var upcomingBillingItem entities.UpcomingBillItem
	err = multierr.Combine(
		upcomingBillingItem.BillItemSequenceNumber.Set(nil),
		upcomingBillingItem.OrderID.Set(upcomingBillingItemEntity.OrderID.String),
		upcomingBillingItem.ProductID.Set(upcomingBillingItemEntity.ProductID.String),
		upcomingBillingItem.StudentProductID.Set(upcomingBillingItemEntity.StudentProductID.String),
		upcomingBillingItem.DiscountID.Set(upcomingBillingItemEntity.DiscountID.String),
		upcomingBillingItem.TaxID.Set(upcomingBillingItemEntity.TaxID.String),
		upcomingBillingItem.BillingSchedulePeriodID.Set(upcomingBillingItemEntity.BillSchedulePeriodID.String),
		upcomingBillingItem.BillingDate.Set(upcomingBillingItemEntity.BillDate.Time),
		upcomingBillingItem.IsGenerated.Set(false),
		upcomingBillingItem.ProductDescription.Set(upcomingBillingItemEntity.ProductDescription.String),
		upcomingBillingItem.ExecuteNote.Set(nil),
	)
	if err != nil {
		err = fmt.Errorf("error while assigning upcoming bill item: %v", err.Error())
		return
	}
	err = s.UpcomingBillItemRepo.Create(ctx, db, &upcomingBillingItem)
	if err != nil {
		err = fmt.Errorf("error while creating upcoming bill item: %v", err.Error())
		return
	}
	return
}

func (s *BillItemService) CreateCancelBillItemForRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapPeriodInfo map[string]entities.BillingSchedulePeriod,
	mapOldBillingItem map[string]entities.BillItem,
) (err error) {
	if orderItemData.OrderItem.StudentProductId.Value == orderItemData.StudentProduct.ProductID.String {
		err = s.BillItemRepo.SetNonLatestBillItemByStudentProductID(ctx, db, orderItemData.OrderItem.StudentProductId.Value)
		if err != nil {
			err = status.Errorf(codes.Internal, SettingNonLatestBillItemError, err.Error())
			return
		}
	}
	if proRatedBillItem.BillingItem != nil {
		var (
			proRatedBillItemEntity    *entities.BillItem
			oldBillItem               entities.BillItem
			proRatedPrice             entities.ProductPrice
			oldPrice                  float32
			oldBillingItemDescription *entities.BillingItemDescription
			discountName              string
		)
		_ = proRatedPrice.Price.Set(0)
		oldBillItem = mapOldBillingItem[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]
		_, _, err = s.UpcomingBillItemRepo.RemoveOldUpcomingBillItem(ctx, db, oldBillItem.OrderID.String, oldBillItem.ProductID.String)
		if err != nil {
			return
		}

		proRatingPeriod := mapPeriodInfo[proRatedBillItem.BillingItem.BillingSchedulePeriodId.Value]
		proRatedBillItemEntity, err = s.generateBillingEntitiesForProRatingRecurringProduct(orderItemData, proRatedBillItem, proRatingPeriod, ratioOfProRatedBillingItem, proRatedPrice)
		if err != nil {
			return
		}

		if proRatedBillItem.BillingItem.AdjustmentPrice == nil {
			err = status.Errorf(codes.Internal, GenerateBillItemUpdateError)
			return
		}

		err = multierr.Combine(
			oldBillItem.Price.AssignTo(&oldPrice),
			proRatedBillItemEntity.AdjustmentPrice.Set(proRatedBillItem.BillingItem.AdjustmentPrice.Value),
			proRatedBillItemEntity.OldPrice.Set(oldPrice),
			proRatedBillItemEntity.BillType.Set(pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
		)

		if err != nil {
			return
		}
		oldBillingItemDescription, err = oldBillItem.GetBillingItemDescription()
		if err != nil {
			return
		}
		if oldBillingItemDescription.DiscountName != nil {
			discountName = *oldBillingItemDescription.DiscountName
		}
		orderItemData.GradeName = oldBillingItemDescription.GradeName
		err = s.generateBillingDescriptionProductAndCreateData(ctx, db, orderItemData, discountName, proRatedBillItemEntity, &ratioOfProRatedBillingItem, &proRatingPeriod)
		if err != nil {
			return
		}
	}

	for _, data := range normalBillItem {
		var (
			billItemEntity            *entities.BillItem
			oldBillItem               entities.BillItem
			oldPrice                  float32
			oldBillingItemDescription *entities.BillingItemDescription
			discountName              string
		)
		period := mapPeriodInfo[data.BillingItem.BillingSchedulePeriodId.Value]
		oldBillItem = mapOldBillingItem[data.BillingItem.BillingSchedulePeriodId.Value]
		billItemEntity, err = s.generateBillingEntitiesForNormalRecurringProduct(orderItemData, data, period)
		if err != nil {
			return
		}

		err = multierr.Combine(
			oldBillItem.Price.AssignTo(&oldPrice),
			billItemEntity.AdjustmentPrice.Set(data.BillingItem.AdjustmentPrice.Value),
			billItemEntity.OldPrice.Set(oldPrice),
			billItemEntity.BillType.Set(pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
		)

		if err != nil {
			return
		}
		oldBillingItemDescription, err = oldBillItem.GetBillingItemDescription()
		if err != nil {
			return
		}
		if oldBillingItemDescription.DiscountName != nil {
			discountName = *oldBillingItemDescription.DiscountName
		}
		err = s.generateBillingDescriptionProductAndCreateData(ctx, db, orderItemData, discountName, billItemEntity, nil, &period)
		if err != nil {
			return
		}
	}
	return
}

func (s *BillItemService) generateBillingEntitiesForNormalRecurringProduct(
	orderItemData utils.OrderItemData,
	billItemData utils.BillingItemData,
	periodInfo entities.BillingSchedulePeriod,
) (billItem *entities.BillItem, err error) {
	var billItemsReq *pb.BillingItem

	billItemsReq = billItemData.BillingItem
	billItem = &entities.BillItem{}
	billItemStatus := pb.BillingStatus_BILLING_STATUS_BILLED
	billType := pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER
	billDate := periodInfo.BillingDate.Time
	if periodInfo.BillingDate.Time.Before(time.Now()) {
		billDate = time.Now()
	}
	if billItemData.IsUpcoming {
		billType = pb.BillingType_BILLING_TYPE_UPCOMING_BILLING
		billItemStatus = pb.BillingStatus_BILLING_STATUS_PENDING
	}

	err = multierr.Combine(
		billItem.StudentID.Set(orderItemData.Order.StudentID.String),
		billItem.ProductID.Set(billItemsReq.ProductId),
		billItem.ProductDescription.Set(orderItemData.ProductInfo.Name.String),
		billItem.ProductPricing.Set(billItemsReq.Price),
		billItem.BillSchedulePeriodID.Set(billItemsReq.BillingSchedulePeriodId.Value),
		billItem.OrderID.Set(orderItemData.Order.OrderID.String),
		billItem.BillDate.Set(billDate),
		billItem.BillStatus.Set(billItemStatus.String()),
		billItem.BillType.Set(billType.String()),
		billItem.BillFrom.Set(periodInfo.StartDate.Time),
		billItem.BillTo.Set(periodInfo.EndDate.Time),
		billItem.FinalPrice.Set(billItemsReq.FinalPrice),
		billItem.BillItemSequenceNumber.Set(nil),
		billItem.StudentProductID.Set(orderItemData.StudentProduct.StudentProductID.String),
		billItem.BillApprovalStatus.Set(nil),
		billItem.LocationID.Set(orderItemData.Order.LocationID.String),
		billItem.LocationName.Set(orderItemData.LocationName),
		billItem.Price.Set(billItemsReq.Price),
		billItem.BillingRatioNumerator.Set(nil),
		billItem.BillingRatioDenominator.Set(nil),
		billItem.IsLatestBillItem.Set(true),
		billItem.PreviousBillItemStatus.Set(nil),
		billItem.PreviousBillItemSequenceNumber.Set(nil),
		billItem.AdjustmentPrice.Set(nil),
		billItem.OldPrice.Set(0),
		billItem.RawDiscountAmount.Set(0),
		billItem.DiscountID.Set(nil),
		billItem.DiscountAmountValue.Set(0),
		billItem.DiscountAmountType.Set(nil),
		billItem.DiscountAmount.Set(0),
		billItem.TaxCategory.Set(nil),
		billItem.TaxPercentage.Set(nil),
		billItem.TaxID.Set(nil),
		billItem.TaxAmount.Set(nil),
	)
	if err != nil {
		return
	}

	if billItemsReq.TaxItem != nil {
		err = multierr.Combine(
			billItem.TaxCategory.Set(billItemsReq.TaxItem.TaxCategory),
			billItem.TaxPercentage.Set(billItemsReq.TaxItem.TaxPercentage),
			billItem.TaxID.Set(billItemsReq.TaxItem.TaxId),
			billItem.TaxAmount.Set(billItemsReq.TaxItem.TaxAmount),
		)
	}
	if err != nil {
		return
	}

	if billItemsReq.DiscountItem != nil {
		roundDiscount := float32(math.Round(float64(billItemsReq.DiscountItem.DiscountAmount)))
		finalPriceBeforeDiscount := billItemsReq.FinalPrice + billItemsReq.DiscountItem.DiscountAmount
		finalPriceAfterRoundDiscount := finalPriceBeforeDiscount - roundDiscount
		err = multierr.Combine(
			billItem.DiscountID.Set(billItemsReq.DiscountItem.DiscountId),
			billItem.DiscountAmountValue.Set(billItemsReq.DiscountItem.DiscountAmountValue),
			billItem.DiscountAmountType.Set(billItemsReq.DiscountItem.DiscountAmountType),
			billItem.DiscountAmount.Set(roundDiscount),
			billItem.FinalPrice.Set(finalPriceAfterRoundDiscount),
			billItem.RawDiscountAmount.Set(billItemsReq.DiscountItem.DiscountAmount),
		)
	}
	return
}

func (s *BillItemService) generateBillingEntitiesForProRatingRecurringProduct(
	orderItemData utils.OrderItemData,
	billItemData utils.BillingItemData,
	periodInfo entities.BillingSchedulePeriod,
	ratio entities.BillingRatio,
	productPrice entities.ProductPrice,
) (billItem *entities.BillItem, err error) {
	var (
		billItemsReq *pb.BillingItem
		price        float32
	)

	_ = productPrice.Price.AssignTo(&price)
	billItemsReq = billItemData.BillingItem
	billItem = &entities.BillItem{}
	billItemStatus := pb.BillingStatus_BILLING_STATUS_BILLED
	billType := pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER
	billDate := periodInfo.BillingDate.Time
	if periodInfo.BillingDate.Time.Before(time.Now()) {
		billDate = time.Now()
	}
	if billItemData.IsUpcoming {
		billType = pb.BillingType_BILLING_TYPE_UPCOMING_BILLING
		billItemStatus = pb.BillingStatus_BILLING_STATUS_PENDING
	}

	err = multierr.Combine(
		billItem.StudentID.Set(orderItemData.Order.StudentID.String),
		billItem.ProductID.Set(billItemsReq.ProductId),
		billItem.ProductDescription.Set(orderItemData.ProductInfo.Name.String),
		billItem.ProductPricing.Set(billItemsReq.Price),
		billItem.BillSchedulePeriodID.Set(billItemsReq.BillingSchedulePeriodId.Value),
		billItem.OrderID.Set(orderItemData.Order.OrderID.String),
		billItem.BillDate.Set(billDate),
		billItem.BillStatus.Set(billItemStatus.String()),
		billItem.BillType.Set(billType.String()),
		billItem.BillFrom.Set(orderItemData.OrderItem.StartDate.AsTime()),
		billItem.BillTo.Set(periodInfo.EndDate.Time),
		billItem.FinalPrice.Set(billItemsReq.FinalPrice),
		billItem.BillItemSequenceNumber.Set(nil),
		billItem.StudentProductID.Set(orderItemData.StudentProduct.StudentProductID.String),
		billItem.BillApprovalStatus.Set(nil),
		billItem.LocationID.Set(orderItemData.Order.LocationID.String),
		billItem.LocationName.Set(orderItemData.LocationName),
		billItem.Price.Set(price),
		billItem.BillingRatioNumerator.Set(ratio.BillingRatioNumerator.Int),
		billItem.BillingRatioDenominator.Set(ratio.BillingRatioDenominator.Int),
		billItem.IsLatestBillItem.Set(true),
		billItem.PreviousBillItemStatus.Set(nil),
		billItem.PreviousBillItemSequenceNumber.Set(nil),
		billItem.AdjustmentPrice.Set(nil),
		billItem.OldPrice.Set(0),
		billItem.TaxCategory.Set(nil),
		billItem.TaxPercentage.Set(nil),
		billItem.TaxID.Set(nil),
		billItem.TaxAmount.Set(nil),
		billItem.DiscountID.Set(nil),
		billItem.DiscountAmountValue.Set(0),
		billItem.DiscountAmountType.Set(nil),
		billItem.DiscountAmount.Set(0),
		billItem.RawDiscountAmount.Set(0),
	)
	if err != nil {
		return
	}
	if billItemsReq.TaxItem != nil {
		err = multierr.Combine(
			billItem.TaxCategory.Set(billItemsReq.TaxItem.TaxCategory),
			billItem.TaxPercentage.Set(billItemsReq.TaxItem.TaxPercentage),
			billItem.TaxID.Set(billItemsReq.TaxItem.TaxId),
			billItem.TaxAmount.Set(billItemsReq.TaxItem.TaxAmount),
		)
	}
	if err != nil {
		return
	}
	if billItemsReq.DiscountItem != nil {
		roundDiscount := float32(math.Round(float64(billItemsReq.DiscountItem.DiscountAmount)))
		finalPriceBeforeDiscount := billItemsReq.FinalPrice + billItemsReq.DiscountItem.DiscountAmount
		finalPriceAfterRoundDiscount := finalPriceBeforeDiscount - roundDiscount
		err = multierr.Combine(
			billItem.DiscountID.Set(billItemsReq.DiscountItem.DiscountId),
			billItem.DiscountAmountValue.Set(billItemsReq.DiscountItem.DiscountAmountValue),
			billItem.DiscountAmountType.Set(billItemsReq.DiscountItem.DiscountAmountType),
			billItem.DiscountAmount.Set(roundDiscount),
			billItem.FinalPrice.Set(finalPriceAfterRoundDiscount),
			billItem.RawDiscountAmount.Set(billItemsReq.DiscountItem.DiscountAmount),
		)
	}
	return
}

func (s *BillItemService) GetMapOldBillingItemForRecurringBilling(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	mapPeriodInfo map[string]entities.BillingSchedulePeriod,
) (mapOldBillingItem map[string]entities.BillItem, err error) {
	mapOldBillingItem = make(map[string]entities.BillItem, len(mapPeriodInfo))
	for periodID := range mapPeriodInfo {
		var billItem entities.BillItem
		billItem, err = s.BillItemRepo.GetBillItemByStudentProductIDAndPeriodID(ctx, db, orderItemData.OrderItem.StudentProductId.Value, periodID)
		if err != nil {
			err = status.Errorf(codes.Internal, "getting old bill item by student product id and period id with err %v", err.Error())
			return
		}
		mapOldBillingItem[periodID] = billItem
	}
	return
}

func (s *BillItemService) CreateCustomBillItem(
	ctx context.Context,
	db database.QueryExecer,
	customBillingItem *pb.CustomBillingItem,
	order entities.Order,
	locationName string,
) (err error) {
	var billingSequenceNumber pgtype.Int4
	billItem := &entities.BillItem{}
	database.AllNullEntity(billItem)
	billItemStatus := pb.BillingStatus_BILLING_STATUS_BILLED
	billDate := time.Now()
	billingItemDescription := entities.BillingItemDescription{
		ProductName: customBillingItem.Name,
	}
	err = multierr.Combine(
		billItem.BillItemSequenceNumber.Set(nil),
		billItem.OrderID.Set(order.OrderID.String),
		billItem.StudentID.Set(order.StudentID.String),
		billItem.BillType.Set(pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String()),
		billItem.BillStatus.Set(billItemStatus),
		billItem.BillDate.Set(billDate),
		billItem.BillFrom.Set(nil),
		billItem.BillTo.Set(nil),
		billItem.ProductDescription.Set(customBillingItem.Name),
		billItem.BillingItemDescription.Set(billingItemDescription),
		billItem.FinalPrice.Set(customBillingItem.Price),
		billItem.BillApprovalStatus.Set(nil),
		billItem.LocationID.Set(order.LocationID.String),
		billItem.LocationName.Set(locationName),
		billItem.TaxCategory.Set(nil),
		billItem.TaxPercentage.Set(nil),
		billItem.TaxID.Set(nil),
		billItem.TaxAmount.Set(nil),
	)
	if err != nil {
		return
	}
	if customBillingItem.TaxItem != nil {
		err = multierr.Combine(
			billItem.TaxCategory.Set(customBillingItem.TaxItem.TaxCategory),
			billItem.TaxPercentage.Set(customBillingItem.TaxItem.TaxPercentage),
			billItem.TaxID.Set(customBillingItem.TaxItem.TaxId),
			billItem.TaxAmount.Set(customBillingItem.TaxItem.TaxAmount),
		)
	}
	if err != nil {
		return
	}
	billingSequenceNumber, err = s.BillItemRepo.Create(ctx, db, billItem)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating bill item for custom order have error %v", err.Error())
	}

	if len(customBillingItem.AccountCategoryIds) == 0 {
		return
	}

	billItemAccountCategories := make([]*entities.BillItemAccountCategory, 0, len(customBillingItem.AccountCategoryIds))
	for _, id := range customBillingItem.AccountCategoryIds {
		billItemAccountCategories = append(billItemAccountCategories, &entities.BillItemAccountCategory{
			BillItemSequenceNumber: billingSequenceNumber,
			AccountCategoryID:      pgtype.Text{Status: pgtype.Present, String: id},
		})
	}
	err = s.BillItemAccountCategoryRepo.CreateMultiple(ctx, db, billItemAccountCategories)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating multiple bill item account category for custom order have error %v", err.Error())
	}
	return
}

func (s *BillItemService) UpdateReviewFlagForBillItem(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	isReviewFlag bool,
) (err error) {
	err = s.BillItemRepo.UpdateReviewFlagByOrderID(ctx, db, orderID, isReviewFlag)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating review flag in billing item service have error: %v", err.Error())
	}
	return
}

func (s *BillItemService) VoidBillItemByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (err error) {
	err = s.BillItemRepo.VoidBillItemByOrderID(ctx, db, orderID, pb.BillingStatus_BILLING_STATUS_CANCELLED.String())
	if err != nil {
		err = status.Errorf(codes.Internal, "voiding bill item have error: %v", err)
	}
	return
}

func (s *BillItemService) UpdateBillItemStatusAndReturnOrderID(
	ctx context.Context,
	db database.QueryExecer,
	billItemSequenceNumber int32,
	billItemStatus string,
) (orderID string, err error) {
	if billItemSequenceNumber < 0 {
		err = status.Errorf(codes.InvalidArgument, "invalid bill item sequence number")
		return
	}
	orderID, err = s.BillItemRepo.UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID(ctx, db, billItemSequenceNumber, billItemStatus)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating bill item status by bill item sequence number %v have error: %v", billItemSequenceNumber, err)
	}
	return
}

func (s *BillItemService) GetRecurringBillItemsForScheduledGenerationOfNextBillItems(
	ctx context.Context,
	db database.QueryExecer,
) (
	billItems []*entities.BillItem,
	err error,
) {
	billItems, err = s.BillItemRepo.GetRecurringBillItemsForScheduledGenerationOfNextBillItems(ctx, db)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error retrieving recurring bill items: %v", err)
	}
	return
}

func (s *BillItemService) GetBillItemDescriptionsByOrderIDWithPaging(
	ctx context.Context,
	db database.Ext,
	orderID string,
	from int64,
	limit int64,
) (
	billingDescriptions []utils.BillItemForRetrieveApi,
	total int,
	err error,
) {
	var billItems []*entities.BillItem
	total, err = s.BillItemRepo.CountBillItemByOrderID(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting count of bill item by order id with error: %v", err)
		return
	}
	billItems = make([]*entities.BillItem, 0, limit)
	billingDescriptions = make([]utils.BillItemForRetrieveApi, 0, limit)
	billItems, err = s.BillItemRepo.GetBillItemByOrderIDAndPaging(ctx, db, orderID, from, limit)
	for i, item := range billItems {
		billingDescription := utils.BillItemForRetrieveApi{
			BillItemEntity: *billItems[i],
		}
		billingDescription.BillItemDescription, err = utils.ConvertBillItemEntityToBillItemDescription(*item)
		if err != nil {
			err = status.Errorf(codes.Internal, "converting billItemEntity to billItemDescription have error: %v", err)
			return
		}
		billingDescriptions = append(billingDescriptions, billingDescription)
	}
	return
}

func (s *BillItemService) GetBillItemDescriptionByStudentIDAndLocationIDs(
	ctx context.Context,
	db database.Ext,
	studentID string,
	locationIDs []string,
	from int64,
	limit int64,
) (
	billingDescriptions []utils.BillItemForRetrieveApi,
	total int,
	err error,
) {
	var billItems []*entities.BillItem
	total, err = s.BillItemRepo.CountBillItemByStudentIDAndLocationIDs(ctx, db, studentID, locationIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting count of bill item by order id with error: %v", err)
		return
	}
	billItems = make([]*entities.BillItem, 0, limit)
	billingDescriptions = make([]utils.BillItemForRetrieveApi, 0, limit)
	billItems, err = s.BillItemRepo.GetBillItemByStudentIDAndLocationIDsPaging(ctx, db, studentID, locationIDs, from, limit)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting bill item with paging with error: %v", err)
		return
	}
	for i, item := range billItems {
		billingDescription := utils.BillItemForRetrieveApi{
			BillItemEntity: *billItems[i],
		}
		billingDescription.BillItemDescription, err = utils.ConvertBillItemEntityToBillItemDescription(*item)
		if err != nil {
			err = status.Errorf(codes.Internal, "converting billItemEntity to billItemDescription have error: %v", err)
			return
		}
		billingDescriptions = append(billingDescriptions, billingDescription)
	}
	return
}

func (s *BillItemService) GetBillItemInfoByOrderIDAndUniqueByProductID(
	ctx context.Context,
	db database.Ext,
	orderID string,
) (
	billItem []*entities.BillItem,
	err error,
) {
	billItem, err = s.BillItemRepo.GetBillItemInfoByOrderIDAndUniqueByProductID(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting bill item info by order id and unique product id have error: %v", err)
	}
	return
}

func (s *BillItemService) GetFirstBillItemsByOrderIDAndProductID(
	ctx context.Context,
	db database.Ext,
	orderID string,
	from int64,
	limit int64,
) (
	billItems []*entities.BillItem,
	total int,
	err error,
) {
	var allBillItem []*entities.BillItem
	billItems = make([]*entities.BillItem, 0, limit)
	allBillItem, err = s.BillItemRepo.GetAllFirstBillItemDistinctByOrderIDAndUniqueByProductID(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting all first bill item distinct by order id and product id have error: %v", err)
		return
	}
	fromInt := int(from) - 1
	limitInt := int(limit)
	lastIndex := fromInt + limitInt
	total = len(allBillItem)
	for i := range allBillItem {
		if i < fromInt {
			continue
		}
		if i > lastIndex {
			break
		}
		billItems = append(billItems, allBillItem[i])
	}
	return
}

func (s *BillItemService) GetLatestBillItemByStudentProductIDForStudentBilling(
	ctx context.Context,
	db database.Ext,
	studentProductID string,
) (
	billItem entities.BillItem,
	err error,
) {
	billItem, err = s.BillItemRepo.GetLatestBillItemByStudentProductIDForStudentBilling(ctx, db, studentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting two latest bill item by student product id have error: %v", err)
	}
	return
}

func (s *BillItemService) GetExportStudentBilling(
	ctx context.Context, db database.QueryExecer, locationIDs []string) (
	exportDatas []*exportEntities.StudentBillingExport, err error) {
	billItems, studentIDs, err := s.BillItemRepo.GetExportStudentBilling(ctx, db, locationIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get billing item for export student billing: %v", err.Error())
		return
	}

	users, err := s.UserRepo.GetStudentsByIDs(ctx, db, studentIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get user for export student billing: %v", err.Error())
		return
	}

	mapStudentIDAndUser := make(map[string]entities.User, len(users))

	for _, user := range users {
		mapStudentIDAndUser[user.UserID.String] = user
	}

	for _, billItem := range billItems {
		billingItemDescription, err := billItem.GetBillingItemDescription()
		if err != nil {
			return nil, err
		}
		billingItemName := billingItemDescription.ProductName
		billingAmount := billItem.FinalPrice
		taxAmount := utils.ConvertNumericToFloat32(billItem.TaxAmount)

		if billItem.BillType.String == pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String() {
			billingItemName = "[Adjustment] " + billingItemName
			billingAmount = billItem.AdjustmentPrice
			taxAmount = (utils.ConvertNumericToFloat32(billItem.AdjustmentPrice) * float32(billItem.TaxPercentage.Int)) / float32(100+billItem.TaxPercentage.Int)
		}

		if billingItemDescription.BillingPeriodName != nil {
			billingItemName = billingItemName + " - " + *billingItemDescription.BillingPeriodName
		}

		if billingItemDescription.BillingRatioNumerator != nil && billingItemDescription.BillingRatioDenominator != nil {
			billingItemName = fmt.Sprintf("%s (Billing ratio: %v/%v)", billingItemName, *billingItemDescription.BillingRatioNumerator, *billingItemDescription.BillingRatioDenominator)
		}

		exportData := &exportEntities.StudentBillingExport{
			StudentName:     mapStudentIDAndUser[billItem.StudentID.String].Name.String,
			StudentID:       billItem.StudentID.String,
			Grade:           billingItemDescription.GradeName,
			Location:        billItem.LocationName.String,
			CreatedDate:     billItem.CreatedAt.Time,
			Status:          billItem.BillStatus.String,
			BillingItemName: billingItemName,
		}

		if billingItemDescription.CourseItems != nil {
			var courses []string
			for _, course := range billingItemDescription.CourseItems {
				var courseConvert string
				if course.Slot != nil {
					courseConvert = course.CourseName + "(" + strconv.Itoa(int(*course.Slot)) + "/wk)"
				} else {
					courseConvert = course.CourseName
				}

				courses = append(courses, courseConvert)
			}
			mapCourses := strings.Join(courses, ", ")
			exportData.Courses = mapCourses
		}

		if billingItemDescription.DiscountName != nil {
			exportData.DiscountName = *billingItemDescription.DiscountName
		}

		if billItem.DiscountAmount.Status == pgtype.Present {
			exportData.DiscountAmount = utils.ConvertNumericToFloat32(billItem.DiscountAmount)
		}

		if billItem.TaxAmount.Status == pgtype.Present {
			exportData.TaxAmount = taxAmount
		}

		if billItem.FinalPrice.Status == pgtype.Present {
			exportData.BillingAmount = utils.ConvertNumericToFloat32(billingAmount)
		}

		exportDatas = append(exportDatas, exportData)

	}

	return
}

func (s *BillItemService) BuildMapBillItemWithProductIDByOrderIDAndProductIDs(
	ctx context.Context, db database.QueryExecer, orderID string, productIDs []string) (
	mapProductIDAndBillItem map[string]entities.BillItem, err error) {
	billItems, err := s.BillItemRepo.GetByOrderIDAndProductIDs(ctx, db, orderID, productIDs)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get bill items by order ID and product IDs: %v", err.Error())
		return
	}
	mapProductIDAndBillItem = make(map[string]entities.BillItem, len(billItems))
	for _, billItem := range billItems {
		mapProductIDAndBillItem[billItem.ProductID.String] = billItem
	}
	return
}

func (s *BillItemService) GetMapPresentAndFutureBillItemInfo(ctx context.Context, db database.QueryExecer, studentProductIDs []string, studentID string) (mapStudentProductIDAndBillItem map[string]*entities.BillItem, err error) {
	billItems, err := s.BillItemRepo.GetPresentAndFutureBillItemsByStudentProductIDs(ctx, db, studentProductIDs, studentID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error get bill items for map data: %v", err.Error())
	}

	mapStudentProductIDAndBillItem = make(map[string]*entities.BillItem, len(billItems))
	for _, billItem := range billItems {
		if _, ok := mapStudentProductIDAndBillItem[billItem.StudentProductID.String]; ok {
			continue
		}
		mapStudentProductIDAndBillItem[billItem.StudentProductID.String] = billItem
	}
	return
}

func (s *BillItemService) GetMapPastBillItemInfo(ctx context.Context, db database.QueryExecer, studentProductIDs []string, studentID string) (mapStudentProductIDAndBillItem map[string]*entities.BillItem, err error) {
	billItems, err := s.BillItemRepo.GetPastBillItemsByStudentProductIDs(ctx, db, studentProductIDs, studentID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error get bill items for map data: %v", err.Error())
	}

	mapStudentProductIDAndBillItem = make(map[string]*entities.BillItem, len(billItems))
	for _, billItem := range billItems {
		if _, ok := mapStudentProductIDAndBillItem[billItem.StudentProductID.String]; ok {
			continue
		}
		mapStudentProductIDAndBillItem[billItem.StudentProductID.String] = billItem
	}
	return
}

func (s *BillItemService) GetUpcomingBilling(ctx context.Context, db database.QueryExecer, studentProductID string, studentID string) (upcomingBillingItem *entities.BillItem, err error) {
	billItem, err := s.BillItemRepo.GetUpcomingBillingByStudentProductID(ctx, db, studentProductID, studentID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error get upcoming billing by student product ID: %v", err.Error())
	}
	upcomingBillingItem = billItem
	return
}

func (s *BillItemService) CreateUpcomingBillItems(
	ctx context.Context,
	db database.QueryExecer,
	billItem *entities.BillItem,
) (err error) {
	billItemSequenceNumber, createErr := s.BillItemRepo.Create(ctx, db, billItem)
	if createErr != nil {
		err = createErr
		return
	}
	billItem.BillItemSequenceNumber = billItemSequenceNumber
	return
}

func NewBillItemService() *BillItemService {
	return &BillItemService{
		BillItemRepo:                &repositories.BillItemRepo{},
		BillItemCourseRepo:          &repositories.BillItemCourseRepo{},
		MaterialRepo:                &repositories.MaterialRepo{},
		UserRepo:                    &repositories.UserRepo{},
		StudentRepo:                 &repositories.StudentRepo{},
		GradeRepo:                   &repositories.GradeRepo{},
		BillItemAccountCategoryRepo: &repositories.BillItemAccountCategoryRepo{},
		UpcomingBillItemRepo:        &repositories.UpcomingBillItemRepo{},
		OrderRepo:                   &repositories.OrderRepo{},
		ProductRepo:                 &repositories.ProductRepo{},
		PriceRepo:                   &repositories.ProductPriceRepo{},
	}
}
