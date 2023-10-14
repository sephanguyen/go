package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	discountEntities "github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var MapOrgIDUserID = map[string]string{
	"-2147483627": "01GTZYX224982Z1X4MHZQW6DO1",
	"-2147483628": "01GTZYX224982Z1X4MHZQW6DO2",
	"-2147483629": "01GTZYX224982Z1X4MHZQW6DP1",
	"-2147483630": "01GTZYX224982Z1X4MHZQW6DP2",
	"-2147483631": "01GTZYX224982Z1X4MHZQW6DP3",
	"-2147483632": "01GTZYX224982Z1X4MHZQW6DP4",
	"-2147483633": "01GTZYX224982Z1X4MHZQW6DP5",
	"-2147483634": "01GTZYX224982Z1X4MHZQW6DP6",
	"-2147483635": "01GTZYX224982Z1X4MHZQW6DP7",
	"-2147483637": "01GTZYX224982Z1X4MHZQW6DP8",
	"-2147483638": "01GTZYX224982Z1X4MHZQW6DP9",
	"-2147483639": "01GTZYX224982Z1X4MHZQW6DQ1",
	"-2147483640": "01GTZYX224982Z1X4MHZQW6DQ2",
	"-2147483641": "01GTZYX224982Z1X4MHZQW6DQ3",
	"-2147483642": "01GTZYX224982Z1X4MHZQW6DQ4",
	"-2147483643": "01GTZYX224982Z1X4MHZQW6DQ5",
	"-2147483644": "01GTZYX224982Z1X4MHZQW6DQ6",
	"-2147483645": "01GTZYX224982Z1X4MHZQW6DQ7",
	"-2147483646": "01GTZYX224982Z1X4MHZQW6DQ8",
	"-2147483647": "01GTZYX224982Z1X4MHZQW6DQ9",
	"-2147483648": "01GTZYX224982Z1X4MHZQW6DR1",
}

type DiscountEventSubscription struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	DB     database.Ext

	BillingRatioRepo                   IBillingRatioRepoForDiscountEventSubscription
	BillingSchedulePeriodRepo          IBillingSchedulePeriodRepoForDiscountEventSubscription
	BillItemRepo                       IBillItemRepoForDiscountEventSubscription
	OrderItemRepo                      IOrderItemRepoForDiscountEventSubscription
	OrderItemCourseRepo                IOrderItemCourseRepoForDiscountEventSubscription
	OrderService                       IOrderServiceServiceForDiscountEventSubscription
	PackageRepo                        IPackageRepoForDiscountEventSubscription
	PackageCourseRepo                  IPackageCourseRepoForDiscountEventSubscription
	PackageQuantityTypeRepo            IPackageQuantityTypeMappingRepoForDiscountEventSubscription
	ProductRepo                        IProductRepoForDiscountEventSubscription
	ProductPriceRepo                   IProductPriceRepoForDiscountEventSubscription
	TaxRepo                            ITaxRepoForDiscountEventSubscription
	StudentEnrollmentStatusHistoryRepo IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription
}

type IBillingRatioRepoForDiscountEventSubscription interface {
	GetFirstRatioByBillingSchedulePeriodIDAndFromTime(ctx context.Context, db database.QueryExecer, billingSchedulePeriodID string, from time.Time) (entities.BillingRatio, error)
}

type IBillingSchedulePeriodRepoForDiscountEventSubscription interface {
	GetAllBillingPeriodsByBillingScheduleID(ctx context.Context, db database.QueryExecer, billingScheduleID string) ([]entities.BillingSchedulePeriod, error)
}

type IBillItemRepoForDiscountEventSubscription interface {
	GetBillItemByStudentProductIDAndPeriodID(ctx context.Context, db database.QueryExecer, studentProductID string, periodID string) (billItem entities.BillItem, err error)
}

type IOrderItemRepoForDiscountEventSubscription interface {
	GetOrderItemByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (orderItem entities.OrderItem, err error)
}

type IOrderItemCourseRepoForDiscountEventSubscription interface {
	GetMapOrderItemCourseByOrderIDAndPackageID(ctx context.Context, db database.QueryExecer, orderID string, packageID string) (mapOrderItemCourse map[string]entities.OrderItemCourse, err error)
}

type IOrderServiceServiceForDiscountEventSubscription interface {
	CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (res *pb.CreateOrderResponse, err error)
}

type IPackageRepoForDiscountEventSubscription interface {
	GetByID(ctx context.Context, db database.QueryExecer, packageID string) (entities.Package, error)
}

type IPackageCourseRepoForDiscountEventSubscription interface {
	GetByPackageIDAndCourseID(ctx context.Context, db database.QueryExecer, packageID string, courseID string) (entities.PackageCourse, error)
}

type IPackageQuantityTypeMappingRepoForDiscountEventSubscription interface {
	GetByPackageTypeForUpdate(ctx context.Context, db database.QueryExecer, packageType string) (quantityType pb.QuantityType, err error)
}

type IProductRepoForDiscountEventSubscription interface {
	GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Product, error)
}

type IProductPriceRepoForDiscountEventSubscription interface {
	GetByProductIDAndPriceType(ctx context.Context, db database.QueryExecer, productID, priceType string) ([]entities.ProductPrice, error)
	GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx context.Context, db database.QueryExecer, productID string, billingSchedulePeriodID string, priceType string) (entities.ProductPrice, error)
	GetByProductIDAndQuantityAndPriceType(ctx context.Context, db database.QueryExecer, productID string, weight int32, priceType string) (entities.ProductPrice, error)
}

type IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription interface {
	GetListEnrolledStatusByStudentIDAndTime(ctx context.Context, db database.QueryExecer, StudentID string, time2 time.Time) ([]*entities.StudentEnrollmentStatusHistory, error)
}

type ITaxRepoForDiscountEventSubscription interface {
	GetByIDForUpdate(ctx context.Context, db database.QueryExecer, taxID string) (entities.Tax, error)
}

func (s *DiscountEventSubscription) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
			nats.Bind(constants.StreamUpdateStudentProduct, constants.DurableUpdateStudentProductCreated),
			nats.DeliverSubject(constants.DeliverUpdateStudentProductCreated),
		},
	}

	_, err := s.JSM.QueueSubscribe(constants.SubjectUpdateStudentProductCreated,
		constants.QueueUpdateStudentProductCreated, option, s.HandleEventUpdateProductDiscount)
	if err != nil {
		return fmt.Errorf("updateStudentProduct.Subscribe: %w", err)
	}

	return nil
}

func (s *DiscountEventSubscription) HandleEventUpdateProductDiscount(ctx context.Context, data []byte) (res bool, err error) {
	userInfo := golibs.UserInfoFromCtx(ctx)
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			ResourcePath: userInfo.ResourcePath,
			UserID:       MapOrgIDUserID[userInfo.ResourcePath],
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	ctx = interceptors.ContextWithUserID(ctx, MapOrgIDUserID[userInfo.ResourcePath])
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var updateProductDiscountInfo *discountEntities.UpdateProductDiscount
	if err := json.Unmarshal(data, &updateProductDiscountInfo); err != nil {
		return false, err
	}

	s.Logger.Info(
		fmt.Sprintf("Discount automation HandleEventUpdateProductDiscount: Initializing Update Order for Student ID: %v, Student Product ID: %v, Discount ID: %v with type %v",
			updateProductDiscountInfo.StudentID,
			updateProductDiscountInfo.StudentProductID,
			updateProductDiscountInfo.DiscountID,
			updateProductDiscountInfo.DiscountType,
		))

	err = s.CreateUpdateOrderRequest(ctx, updateProductDiscountInfo)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *DiscountEventSubscription) CreateUpdateOrderRequest(ctx context.Context, updateProductDiscountInfo *discountEntities.UpdateProductDiscount) (err error) {
	orderItems := []*pb.OrderItem{}
	billItems := []*pb.BillingItem{}
	upcomingBillItems := []*pb.BillingItem{}

	product, err := s.ProductRepo.GetByID(ctx, s.DB, updateProductDiscountInfo.ProductID)
	if err != nil {
		err = fmt.Errorf("fail to retrieve product with ID %v with err %v", updateProductDiscountInfo.ProductID, err)
		return
	}

	generatedOrderItems, courseItems, quantityType, err := s.GenerateOrderItemForUpdateProductDiscount(ctx, *updateProductDiscountInfo, product)
	if err != nil {
		return
	}
	orderItems = append(orderItems, generatedOrderItems)

	generatedAdjustmentBillItems, generatedUpcomingAdjustmentBillItem, err := s.GenerateBillItemsForUpdateProductDiscount(ctx, *updateProductDiscountInfo, product, courseItems, quantityType)
	if err != nil {
		return
	}

	billItems = append(billItems, generatedAdjustmentBillItems...)
	upcomingBillItems = append(upcomingBillItems, generatedUpcomingAdjustmentBillItem...)

	if len(billItems)+len(upcomingBillItems) == 0 {
		err = fmt.Errorf("no bill item generated for Student ID: %v, Student Product ID: %v, Discount ID: %v with type %v",
			updateProductDiscountInfo.StudentID,
			updateProductDiscountInfo.StudentProductID,
			updateProductDiscountInfo.DiscountID,
			updateProductDiscountInfo.DiscountType,
		)

		s.Logger.Debug(fmt.Sprintf("Discount automation HandleEventUpdateProductDiscount error: %v", err))
		return err
	}

	req := pb.CreateOrderRequest{
		StudentId:            updateProductDiscountInfo.StudentID,
		LocationId:           updateProductDiscountInfo.LocationID,
		OrderType:            pb.OrderType_ORDER_TYPE_UPDATE,
		OrderItems:           orderItems,
		BillingItems:         billItems,
		UpcomingBillingItems: upcomingBillItems,
	}

	resp, err := s.OrderService.CreateOrder(ctx, &req)
	if err != nil {
		s.Logger.Debug(
			fmt.Sprintf("Discount automation HandleEventUpdateProductDiscount: Create order error: %v, Student ID: %v, Student Product ID: %v, Discount ID: %v with type %v",
				err,
				updateProductDiscountInfo.StudentID,
				updateProductDiscountInfo.StudentProductID,
				updateProductDiscountInfo.DiscountID,
				updateProductDiscountInfo.DiscountType,
			))
		s.Logger.Debug(fmt.Sprintf("Discount automation HandleEventUpdateProductDiscount failed order items for student product %v: %v", updateProductDiscountInfo.StudentProductID, req.OrderItems))
		s.Logger.Debug(fmt.Sprintf("Discount automation HandleEventUpdateProductDiscount failed bill items for student product %v: %v", updateProductDiscountInfo.StudentProductID, req.BillingItems))
		s.Logger.Debug(fmt.Sprintf("Discount automation HandleEventUpdateProductDiscount failed upcoming bill item for student product %v: %v", updateProductDiscountInfo.StudentProductID, req.UpcomingBillingItems))

		return err
	}

	s.Logger.Info(
		fmt.Sprintf("Discount automation HandleEventUpdateProductDiscount: Update Order for Student ID: %v with Student Product ID: %v, successfully created, Order ID: %v",
			updateProductDiscountInfo.StudentID,
			updateProductDiscountInfo.StudentProductID,
			resp.OrderId,
		))

	return
}

func (s *DiscountEventSubscription) GenerateOrderItemForUpdateProductDiscount(
	ctx context.Context,
	req discountEntities.UpdateProductDiscount,
	product entities.Product,
) (
	orderItem *pb.OrderItem,
	courseItems []*pb.CourseItem,
	quantityType pb.QuantityType,
	err error,
) {
	var (
		oldOrderItem entities.OrderItem
	)

	oldOrderItem, err = s.OrderItemRepo.GetOrderItemByStudentProductID(ctx, s.DB, req.StudentProductID)
	if err != nil {
		err = fmt.Errorf("fail to retrieve order items of student product with ID %v with err %v", req.StudentProductID, err)
		return
	}

	orderItem = &pb.OrderItem{
		ProductId: req.ProductID,
		StudentProductId: &wrapperspb.StringValue{
			Value: req.StudentProductID,
		},
		EffectiveDate: &timestamppb.Timestamp{Seconds: req.EffectiveDate.Unix()},
	}

	if req.DiscountID != "" {
		orderItem.DiscountId = &wrapperspb.StringValue{
			Value: req.DiscountID,
		}
	}

	if product.ProductType.String == pb.ProductType_PRODUCT_TYPE_PACKAGE.String() {
		var (
			courseMap      map[string]entities.OrderItemCourse
			productPackage entities.Package
		)
		courseItems = []*pb.CourseItem{}

		courseMap, err = s.OrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, s.DB, oldOrderItem.OrderID.String, product.ProductID.String)
		if err != nil {
			err = fmt.Errorf("fail to create course map of product with ID %v with err %v", req.ProductID, err)
			return
		}

		productPackage, err = s.PackageRepo.GetByID(ctx, s.DB, product.ProductID.String)
		if err != nil {
			err = fmt.Errorf("fail to retrieve package with ID %v with err %v", req.ProductID, err)
			return
		}

		quantityType, err = s.PackageQuantityTypeRepo.GetByPackageTypeForUpdate(ctx, s.DB, productPackage.PackageType.String)
		if err != nil {
			err = fmt.Errorf("fail to retrieve quantity type of package with ID %v with err %v", req.ProductID, err)
			return
		}

		for courseID, orderItemCourse := range courseMap {
			courseItem := &pb.CourseItem{
				CourseId:   courseID,
				CourseName: orderItemCourse.CourseName.String,
			}

			if quantityType.String() == pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String() {
				courseItem.Slot = wrapperspb.Int32(orderItemCourse.CourseSlot.Int)
			} else {
				var packageCourse entities.PackageCourse
				packageCourse, err = s.PackageCourseRepo.GetByPackageIDAndCourseID(ctx, s.DB, product.ProductID.String, courseID)
				if err != nil {
					err = fmt.Errorf("fail to retrieve package course weight of package ID %v and course ID %v with err %v", product.ProductID.String, courseID, err)
					return
				}
				courseItem.Weight = wrapperspb.Int32(packageCourse.CourseWeight.Int)
			}
			courseItems = append(courseItems, courseItem)
		}

		orderItem.CourseItems = courseItems
	} else {
		quantityType = pb.QuantityType_QUANTITY_TYPE_NONE
	}

	return
}

func (s *DiscountEventSubscription) GenerateBillItemsForUpdateProductDiscount(
	ctx context.Context,
	updateProductDiscountInfo discountEntities.UpdateProductDiscount,
	product entities.Product,
	courseItems []*pb.CourseItem,
	quantityType pb.QuantityType,
) (
	billItems []*pb.BillingItem,
	upcomingBillItems []*pb.BillingItem,
	err error,
) {
	var (
		enrollmentStatusList []*entities.StudentEnrollmentStatusHistory
		isOrgEnrolled        bool
	)
	billItems = []*pb.BillingItem{}
	upcomingBillItems = []*pb.BillingItem{}

	billingPeriods, err := s.BillingSchedulePeriodRepo.GetAllBillingPeriodsByBillingScheduleID(ctx, s.DB, product.BillingScheduleID.String)
	if err != nil {
		err = fmt.Errorf("fail to retrieve billing periods for billing schedule ID %v with err %v", product.BillingScheduleID.String, err)
		return
	}

	enrollmentStatusList, err = s.StudentEnrollmentStatusHistoryRepo.GetListEnrolledStatusByStudentIDAndTime(ctx, s.DB, updateProductDiscountInfo.StudentID, time.Now())
	if err == nil && len(enrollmentStatusList) > 0 {
		isOrgEnrolled = true
	}

	orderDate := time.Now()
	for _, billingPeriod := range billingPeriods {
		var (
			oldBillItem entities.BillItem
			billItem    *pb.BillingItem
		)

		if billingPeriod.EndDate.Time.Before(updateProductDiscountInfo.EffectiveDate) ||
			billingPeriod.StartDate.Time.After(updateProductDiscountInfo.StudentProductEndDate) {
			continue
		}

		oldBillItem, billErr := s.BillItemRepo.GetBillItemByStudentProductIDAndPeriodID(ctx, s.DB, updateProductDiscountInfo.StudentProductID, billingPeriod.BillingSchedulePeriodID.String)
		if billErr != nil {
			continue
		}

		billItem, err = s.InitializeBillItemData(ctx, updateProductDiscountInfo, oldBillItem, courseItems, product, billingPeriod, quantityType, isOrgEnrolled)
		if err != nil {
			err = fmt.Errorf("fail to initialize bill item for product %v and period %v with err %v", updateProductDiscountInfo.ProductID, billingPeriod.BillingSchedulePeriodID, err)
			return
		}

		if billingPeriod.BillingDate.Time.After(orderDate) {
			upcomingBillItems = append(upcomingBillItems, billItem)
		} else {
			billItems = append(billItems, billItem)
		}
	}

	return
}

func (s *DiscountEventSubscription) InitializeBillItemData(
	ctx context.Context,
	updateProductDiscountInfo discountEntities.UpdateProductDiscount,
	oldBillItem entities.BillItem,
	courseItems []*pb.CourseItem,
	product entities.Product,
	billingPeriod entities.BillingSchedulePeriod,
	quantityType pb.QuantityType,
	isOrgEnrolled bool,
) (
	billItem *pb.BillingItem,
	err error,
) {
	var (
		price                   float32
		originalPrice           float32
		discountAmt             float32
		taxAmt                  float32
		tax                     entities.Tax
		productPriceDefault     entities.ProductPrice
		productPriceEnrolled    entities.ProductPrice
		productPrice            entities.ProductPrice
		billingRatioNumerator   int32
		billingRatioDenominator int32
		billingRatio            entities.BillingRatio
		quantity                int32
	)

	billItem = &pb.BillingItem{
		ProductId: updateProductDiscountInfo.ProductID,
		StudentProductId: &wrapperspb.StringValue{
			Value: updateProductDiscountInfo.StudentProductID,
		},
		BillingSchedulePeriodId: &wrapperspb.StringValue{
			Value: oldBillItem.BillSchedulePeriodID.String,
		},
	}

	if product.ProductType.String == pb.ProductType_PRODUCT_TYPE_PACKAGE.String() {
		quantity = getQuantityFromCourseItems(courseItems, quantityType)
		productPriceDefault, err = s.ProductPriceRepo.GetByProductIDAndQuantityAndPriceType(ctx, s.DB, updateProductDiscountInfo.ProductID, quantity, pb.ProductPriceType_DEFAULT_PRICE.String())
		if err != nil {
			err = fmt.Errorf("fail to retrieve default price of product %v for quantity %v with err %v", updateProductDiscountInfo.ProductID, quantity, err)
			return
		}
		productPrice = productPriceDefault

		if isOrgEnrolled {
			productPriceEnrolled, err = s.ProductPriceRepo.GetByProductIDAndQuantityAndPriceType(ctx, s.DB, updateProductDiscountInfo.ProductID, quantity, pb.ProductPriceType_ENROLLED_PRICE.String())
			if err == nil {
				productPrice = productPriceEnrolled
			} else {
				s.Logger.Debug(fmt.Sprintf("student %v is org enrolled but product %v has no available enrolled price in system, using default product price instead", updateProductDiscountInfo.StudentID, updateProductDiscountInfo.ProductID))
			}
		}

		billItem.CourseItems = courseItems
		billItem.Quantity = &wrapperspb.Int32Value{
			Value: quantity,
		}
	} else {
		productPriceDefault, err = s.ProductPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, s.DB, updateProductDiscountInfo.ProductID, oldBillItem.BillSchedulePeriodID.String, pb.ProductPriceType_DEFAULT_PRICE.String())
		if err != nil {
			err = fmt.Errorf("fail to retrieve price of product %v for period %v with err %v", updateProductDiscountInfo.ProductID, oldBillItem.BillSchedulePeriodID.String, err)
			return
		}
		productPrice = productPriceDefault

		if isOrgEnrolled {
			productPriceEnrolled, err = s.ProductPriceRepo.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, s.DB, updateProductDiscountInfo.ProductID, oldBillItem.BillSchedulePeriodID.String, pb.ProductPriceType_ENROLLED_PRICE.String())
			if err == nil {
				productPrice = productPriceEnrolled
			} else {
				s.Logger.Debug(fmt.Sprintf("student %v is org enrolled but product %v has no available enrolled price in system, using default product price instead", updateProductDiscountInfo.StudentID, updateProductDiscountInfo.ProductID))
			}
		}
	}

	originalPrice = utils.ConvertNumericToFloat32(productPrice.Price)

	if !product.DisableProRatingFlag.Bool && billingPeriod.StartDate.Time.Before(updateProductDiscountInfo.EffectiveDate) && billingPeriod.EndDate.Time.After(updateProductDiscountInfo.EffectiveDate) {
		billingRatio, err = s.BillingRatioRepo.GetFirstRatioByBillingSchedulePeriodIDAndFromTime(ctx, s.DB, billingPeriod.BillingSchedulePeriodID.String, updateProductDiscountInfo.EffectiveDate)
		if err != nil {
			err = fmt.Errorf("fail to retrieve billing ratio of billing schedule period %v for time %v with err %v", oldBillItem.BillSchedulePeriodID.String, updateProductDiscountInfo.EffectiveDate, err)
			return
		}

		billingRatioNumerator = billingRatio.BillingRatioNumerator.Int
		billingRatioDenominator = billingRatio.BillingRatioDenominator.Int
		price = (originalPrice * float32(billingRatioNumerator)) / float32(billingRatioDenominator)
	} else {
		billingRatioNumerator = 1
		billingRatioDenominator = 1
		price = originalPrice
	}

	discountAmt = 0
	if updateProductDiscountInfo.DiscountID != "" {
		discountAmt = getPercentDiscountValue(price, updateProductDiscountInfo.DiscountAmountValue)
		billItem.DiscountItem = &pb.DiscountBillItem{
			DiscountId:          updateProductDiscountInfo.DiscountID,
			DiscountType:        updateProductDiscountInfo.DiscountType,
			DiscountAmountType:  updateProductDiscountInfo.DiscountAmountType,
			DiscountAmountValue: updateProductDiscountInfo.DiscountAmountValue,
			DiscountAmount:      discountAmt,
		}
	}

	tax, err = s.TaxRepo.GetByIDForUpdate(ctx, s.DB, oldBillItem.TaxID.String)
	if err != nil {
		err = fmt.Errorf("fail to retrieve tax from old billing item for student product %v and period %v with err %v", oldBillItem.StudentProductID, oldBillItem.BillSchedulePeriodID, err)
		return
	}

	if tax.TaxID.String != "" {
		taxAmt = getInclusivePercentTax(getPercentDiscountedPrice(price, updateProductDiscountInfo.DiscountAmountValue), float32(tax.TaxPercentage.Int))
		billItem.TaxItem = &pb.TaxBillItem{
			TaxId:         tax.TaxID.String,
			TaxCategory:   pb.TaxCategory(pb.TaxCategory_value[tax.TaxCategory.String]),
			TaxPercentage: float32(tax.TaxPercentage.Int),
			TaxAmount:     taxAmt,
		}
	}

	billItem.Price = price
	billItem.FinalPrice = price - discountAmt

	oldOriginalPrice := utils.ConvertNumericToFloat32(oldBillItem.Price)
	if oldBillItem.DiscountAmount.Status == pgtype.Present {
		var (
			discountAmountValue float32
		)
		_ = oldBillItem.DiscountAmountValue.AssignTo(&discountAmountValue)
		oldOriginalDiscount := calculateOriginalDiscount(oldOriginalPrice, oldBillItem.DiscountAmountType.String, discountAmountValue)
		oldOriginalPrice -= oldOriginalDiscount
	}

	newOriginalPrice := getPercentDiscountedPrice(originalPrice, updateProductDiscountInfo.DiscountAmountValue)
	billItem.AdjustmentPrice = &wrapperspb.FloatValue{
		Value: getAdjustmentPrice(
			newOriginalPrice,
			oldOriginalPrice,
			billingRatioNumerator,
			billingRatioDenominator),
	}

	return
}

func getAdjustmentPrice(newPrice float32, oldPrice float32, numerator int32, denominator int32) float32 {
	return ((newPrice - oldPrice) * float32(numerator)) / float32(denominator)
}

func getInclusivePercentTax(priceAfterDiscount float32, taxPercent float32) float32 {
	return float32(float64(priceAfterDiscount*taxPercent) / float64(100+taxPercent))
}

func getPercentDiscountedPrice(priceOrder float32, percentDiscount float32) float32 {
	return priceOrder - getPercentDiscountValue(priceOrder, percentDiscount)
}

func getPercentDiscountValue(priceOrder float32, percentDiscount float32) float32 {
	return priceOrder * (percentDiscount / 100)
}

func calculateOriginalDiscount(originalPrice float32, discountType string, discountAmountValue float32) (originalDiscount float32) {
	originalDiscount = discountAmountValue
	if discountType == pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String() {
		originalDiscount = (originalPrice * discountAmountValue) / float32(100)
	}
	return
}

func getQuantityFromCourseItems(courseItems []*pb.CourseItem, quantityType pb.QuantityType) (quantity int32) {
	quantity = 0
	for _, course := range courseItems {
		if quantityType.String() == pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String() {
			quantity += course.GetSlot().Value
		} else {
			quantity += course.GetWeight().Value
		}
	}
	return
}
