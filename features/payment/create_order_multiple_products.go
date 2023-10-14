package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const MultipleOneTimeProducts = "Multiple One Time Products"

func (s *suite) prepareDataForCreatingOrderMultipleOneTimeProducts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		feeIDs      []string
		materialIDs []string
		packageIDs  []string
		courseIDs   []string
		req         pb.CreateOrderRequest
		err         error
	)

	taxID,
		discountIDs,
		locationID,
		feeIDs,
		materialIDs,
		packageIDs,
		courseIDs,
		userID,
		_,
		err = s.insertAllDataForInsertOrderMultipleOneTimeProducts(ctx, MultipleOneTimeProducts)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order multiple one time products"

	orderItems := make([]*pb.OrderItem, 0, len(feeIDs)+len(materialIDs)+len(packageIDs))
	billingItems := make([]*pb.BillingItem, 0, len(feeIDs)+len(materialIDs)+len(packageIDs))

	courseItems := []*pb.CourseItem{
		{
			CourseId:   courseIDs[0],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[0]),
			Weight:     &wrapperspb.Int32Value{Value: 1},
		},
		{
			CourseId:   courseIDs[1],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[1]),
			Weight:     &wrapperspb.Int32Value{Value: 2},
		},
		{
			CourseId:   courseIDs[2],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[2]),
			Weight:     &wrapperspb.Int32Value{Value: 3},
		},
	}
	startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: feeIDs[0],
			StartDate: startDate,
		},
		&pb.OrderItem{
			ProductId:  feeIDs[1],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId:  feeIDs[2],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId: materialIDs[0],
			StartDate: startDate,
		},
		&pb.OrderItem{
			ProductId:  materialIDs[1],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId:  materialIDs[2],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[0],
			CourseItems: courseItems,
			StartDate:   startDate,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[1],
			CourseItems: courseItems,
			StartDate:   startDate,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[2],
			CourseItems: courseItems,
			StartDate:   startDate,
		},
	)

	billingItems = append(billingItems,
		&pb.BillingItem{
			ProductId: feeIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId: feeIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[0],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 20,
				DiscountAmount:      20,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     80,
			},
			FinalPrice: PriceOrder - 20,
		},
		&pb.BillingItem{
			ProductId: feeIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 20,
				DiscountAmount:      PriceOrder * 20 / 100,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     66.666664,
			},
			FinalPrice: PriceOrder - PriceOrder*20/100,
		},
		&pb.BillingItem{
			ProductId: materialIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId: materialIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[0],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 20,
				DiscountAmount:      20,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     80,
			},
			FinalPrice: PriceOrder - 20,
		},
		&pb.BillingItem{
			ProductId: materialIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 20,
				DiscountAmount:      PriceOrder * 20 / 100,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     66.666664,
			},
			FinalPrice: PriceOrder - PriceOrder*20/100,
		},
		&pb.BillingItem{
			ProductId: packageIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			CourseItems: courseItems,
			FinalPrice:  PriceOrder,
		},
		&pb.BillingItem{
			ProductId: packageIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		},
		&pb.BillingItem{
			ProductId: packageIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		},
	)

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertAllDataForInsertOrderMultipleOneTimeProducts(ctx context.Context, name string) (
	taxID string,
	discountIDs []string,
	locationID string,
	feeIDs []string,
	materialIDs []string,
	packageIDs []string,
	courseIDs []string,
	userID string,
	gradeID string,
	err error,
) {
	gradeID, err = mockdata.InsertOneGrade(ctx, s.FatimaDBTrace)
	if err != nil {
		return
	}

	taxID, err = mockdata.InsertOneTax(ctx, s.FatimaDBTrace, name)
	if err != nil {
		return
	}

	discountIDs, err = mockdata.InsertOneDiscountAmount(ctx, s.FatimaDBTrace, name)
	if err != nil {
		return
	}

	locationID = constants.ManabieOrgLocation

	userID, err = mockdata.InsertOneUser(ctx, s.FatimaDBTrace, gradeID)
	if err != nil {
		return
	}

	materialIDs, err = s.insertMaterial(ctx, taxID)
	if err != nil {
		return
	}

	feeIDs, err = mockdata.InsertFee(ctx, s.FatimaDBTrace, taxID)
	if err != nil {
		return
	}

	packageIDs, err = s.insertPackage(ctx, taxID)
	if err != nil {
		return
	}

	err = mockdata.InsertPackageTypeQuantityTypeMapping(ctx, s.FatimaDBTrace)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"package_quantity_type_mapping_pk\"") {
			return
		}
	}

	for _, productIDs := range [][]string{feeIDs, materialIDs, packageIDs} {
		err = s.insertProductLocation(ctx, locationID, productIDs)
		if err != nil {
			return
		}
	}

	for _, productIDs := range [][]string{feeIDs, materialIDs} {
		err = s.insertProductPrice(ctx, productIDs, pb.ProductPriceType_DEFAULT_PRICE.String())
		if err != nil {
			return
		}
	}

	err = s.insertProductPriceForPackage(ctx, packageIDs)
	if err != nil {
		return
	}

	for _, productIDs := range [][]string{feeIDs, materialIDs, packageIDs} {
		err = s.insertProductGrade(ctx, gradeID, productIDs)
		if err != nil {
			return
		}
	}

	courseIDs, err = s.insertCourses(ctx)
	if err != nil {
		return
	}

	err = s.insertPackageCourses(ctx, packageIDs, courseIDs)
	if err != nil {
		return
	}

	err = s.insertProductDiscount(ctx, materialIDs, discountIDs)
	if err != nil {
		return
	}

	err = s.insertProductDiscount(ctx, feeIDs, discountIDs)
	if err != nil {
		return
	}

	err = s.insertProductDiscount(ctx, packageIDs, discountIDs)
	if err != nil {
		return
	}

	return
}

func (s *suite) prepareDataForCreatingOrderMultipleRecurringProducts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                        true,
		InsertDiscount:                   true,
		InsertStudent:                    true,
		InsertProductPrice:               true,
		InsertProductLocation:            true,
		InsertLocation:                   false,
		InsertProductGrade:               true,
		InsertFee:                        true,
		InsertMaterial:                   true,
		InsertBillingSchedule:            true,
		InsertBillingScheduleArchived:    false,
		IsTaxExclusive:                   false,
		InsertDiscountNotAvailable:       false,
		InsertProductOutOfTime:           false,
		InsertPackageCourses:             true,
		InsertPackageCourseScheduleBased: true,
		InsertProductDiscount:            true,
		BillingScheduleStartDate:         time.Now(),
	}

	req := pb.CreateOrderRequest{}
	var err error

	feeIDs, materialIDs, packageIDs, data, err := s.insertDataForInsertOrderWithMultipleRecurringProducts(ctx, s.FatimaDBTrace, defaultOptionPrepareData, constants.ManabieOrgLocation, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: feeIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
		&pb.OrderItem{
			ProductId: materialIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
		&pb.OrderItem{
			ProductId: packageIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               feeIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId:               materialIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId:               packageIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 2),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               feeIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId:               materialIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId:               packageIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 2),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order multiple recurring products"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertDataForInsertOrderWithMultipleRecurringProducts(
	ctx context.Context,
	fatimaDBTrace *database.DBTrace,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
	locationID string,
	gradeID string,
) (feeIDs []string, materialIDs []string, packageIDs []string, data mockdata.DataForRecurringProduct, err error) {
	name := "recurring product test"
	if len(gradeID) == 0 {
		gradeID, err = mockdata.InsertOneGrade(ctx, s.FatimaDBTrace)
		if err != nil {
			return
		}
	}

	data.BillingScheduleID, err = mockdata.InsertBillingScheduleForRecurringProduct(ctx, fatimaDBTrace, options.InsertBillingScheduleArchived)
	if err != nil {
		return
	}

	billingSchedule, billingScheduleErr := mockdata.GetBillingSchedule(ctx, fatimaDBTrace, data.BillingScheduleID)
	if billingScheduleErr != nil {
		err = billingScheduleErr
		return
	}

	err = mockdata.InsertBillingSchedulePeriodForRecurringProduct(ctx, fatimaDBTrace, billingSchedule, options.BillingScheduleStartDate)
	if err != nil {
		return
	}

	err = mockdata.InsertBillingRatioForRecurringProduct(ctx, fatimaDBTrace, billingSchedule)
	if err != nil {
		return
	}

	data.BillingSchedulePeriods, err = mockdata.GetBillingPeriodBySchedule(ctx, fatimaDBTrace, billingSchedule)
	if err != nil {
		return
	}

	if options.IsTaxExclusive {
		data.TaxID, err = mockdata.InsertOneTaxExclusive(ctx, fatimaDBTrace, name)
		if err != nil {
			return
		}
	} else {
		data.TaxID, err = mockdata.InsertOneTax(ctx, fatimaDBTrace, name)
		if err != nil {
			return
		}
	}

	data.DiscountIDs, err = mockdata.InsertDiscountForRecurringProduct(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	if options.InsertDiscountNotAvailable {
		data.DiscountIDs, err = mockdata.InsertOneDiscountAmountNotAvailable(ctx, fatimaDBTrace, name)
		if err != nil {
			return
		}
	}

	if len(locationID) == 0 {
		data.LocationID, err = mockdata.InsertOneLocation(ctx, fatimaDBTrace)
		if err != nil {
			return
		}
	} else {
		data.LocationID = locationID
	}

	data.UserID, err = mockdata.InsertOneUser(ctx, fatimaDBTrace, gradeID)
	if err != nil {
		return
	}

	materialIDs, err = mockdata.InsertRecurringMaterials(ctx, fatimaDBTrace, data.TaxID, data.BillingScheduleID)
	if err != nil {
		return
	}

	feeIDs, err = mockdata.InsertRecurringFees(ctx, fatimaDBTrace, data.TaxID, data.BillingScheduleID)
	if err != nil {
		return
	}

	err = mockdata.InsertPackageTypeQuantityTypeMapping(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	packageIDs, data.CourseIDs, err = mockdata.InsertPackageCourses(ctx, fatimaDBTrace, data.TaxID, data.BillingScheduleID, options.InsertPackageCourseScheduleBased, options.ArePackageCoursesMandatory)
	if err != nil {
		return
	}

	data.PackageCourses, err = mockdata.GetPackageCourseByPackageIDs(ctx, fatimaDBTrace, packageIDs)
	if err != nil {
		return
	}

	err = mockdata.InsertProductPriceForQtyPackage(ctx, fatimaDBTrace, packageIDs, data.BillingSchedulePeriods)
	if err != nil {
		return
	}

	for _, productIDs := range [][]string{feeIDs, materialIDs, packageIDs} {
		err = mockdata.InsertProductPriceForRecurringProducts(ctx, fatimaDBTrace, productIDs, data.BillingSchedulePeriods, PriceOrder, pb.ProductPriceType_DEFAULT_PRICE.String())
		if err != nil {
			return
		}
	}

	for _, productIDs := range [][]string{feeIDs, materialIDs, packageIDs} {
		err = mockdata.InsertProductGrade(ctx, fatimaDBTrace, gradeID, productIDs)
		if err != nil {
			return
		}
	}

	for _, productIDs := range [][]string{feeIDs, materialIDs, packageIDs} {
		err = mockdata.InsertProductLocation(ctx, fatimaDBTrace, data.LocationID, productIDs)
		if err != nil {
			return
		}
	}

	err = s.insertProductDiscount(ctx, materialIDs, data.DiscountIDs)
	if err != nil {
		return
	}

	err = s.insertProductDiscount(ctx, feeIDs, data.DiscountIDs)
	if err != nil {
		return
	}

	err = s.insertProductDiscount(ctx, packageIDs, data.DiscountIDs)
	if err != nil {
		return
	}

	return
}

func (s *suite) prepareDataForCreatingOrderMultipleOneTimeAndRecurringProducts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		feeIDs      []string
		materialIDs []string
		packageIDs  []string
		courseIDs   []string
		gradeID     string
		req         pb.CreateOrderRequest
		err         error
	)
	taxID,
		discountIDs,
		locationID,
		feeIDs,
		materialIDs,
		packageIDs,
		courseIDs,
		userID,
		gradeID,
		err = s.insertAllDataForInsertOrderMultipleOneTimeProducts(ctx, MultipleOneTimeProducts)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                        true,
		InsertDiscount:                   true,
		InsertStudent:                    true,
		InsertProductPrice:               true,
		InsertProductLocation:            true,
		InsertLocation:                   false,
		InsertProductGrade:               true,
		InsertFee:                        true,
		InsertMaterial:                   true,
		InsertBillingSchedule:            true,
		InsertBillingScheduleArchived:    false,
		IsTaxExclusive:                   false,
		InsertDiscountNotAvailable:       false,
		InsertProductOutOfTime:           false,
		InsertPackageCourses:             true,
		InsertPackageCourseScheduleBased: true,
		InsertProductDiscount:            true,
		BillingScheduleStartDate:         time.Now(),
	}

	recurringFeeIDs, recurringMaterialIDs, recurringPackageIDs, data, err := s.insertDataForInsertOrderWithMultipleRecurringProducts(ctx, s.FatimaDBTrace, defaultOptionPrepareData, locationID, gradeID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order multiple one time products"

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	courseItems := []*pb.CourseItem{
		{
			CourseId:   courseIDs[0],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[0]),
			Weight:     &wrapperspb.Int32Value{Value: 1},
		},
		{
			CourseId:   courseIDs[1],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[1]),
			Weight:     &wrapperspb.Int32Value{Value: 2},
		},
		{
			CourseId:   courseIDs[2],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[2]),
			Weight:     &wrapperspb.Int32Value{Value: 3},
		},
	}

	startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: feeIDs[0],
			StartDate: startDate,
		},
		&pb.OrderItem{
			ProductId:  feeIDs[1],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId:  feeIDs[2],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId: recurringFeeIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
		&pb.OrderItem{
			ProductId: materialIDs[0],
			StartDate: startDate,
		},
		&pb.OrderItem{
			ProductId:  materialIDs[1],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId:  materialIDs[2],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId: recurringMaterialIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
		&pb.OrderItem{
			ProductId:   packageIDs[0],
			CourseItems: courseItems,
			StartDate:   startDate,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[1],
			CourseItems: courseItems,
			StartDate:   startDate,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[2],
			CourseItems: courseItems,
			StartDate:   startDate,
		},
		&pb.OrderItem{
			ProductId: recurringPackageIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)

	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId: feeIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId: feeIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[0],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 20,
				DiscountAmount:      20,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     80,
			},
			FinalPrice: PriceOrder - 20,
		},
		&pb.BillingItem{
			ProductId: feeIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 20,
				DiscountAmount:      PriceOrder * 20 / 100,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     66.666664,
			},
			FinalPrice: PriceOrder - PriceOrder*20/100,
		},
		&pb.BillingItem{
			ProductId:               recurringFeeIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId: materialIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId: materialIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[0],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 20,
				DiscountAmount:      20,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     80,
			},
			FinalPrice: PriceOrder - 20,
		},
		&pb.BillingItem{
			ProductId: materialIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 20,
				DiscountAmount:      PriceOrder * 20 / 100,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     66.666664,
			},
			FinalPrice: PriceOrder - PriceOrder*20/100,
		},
		&pb.BillingItem{
			ProductId:               recurringMaterialIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId: packageIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			CourseItems: courseItems,
			FinalPrice:  PriceOrder,
		},
		&pb.BillingItem{
			ProductId: packageIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		},
		&pb.BillingItem{
			ProductId: packageIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		},
		&pb.BillingItem{
			ProductId:               recurringPackageIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 2),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)

	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               recurringFeeIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId:               recurringMaterialIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: PriceOrder,
		},
		&pb.BillingItem{
			ProductId:               recurringPackageIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 2),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order multiple one time and recurring products"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) orderOfMultipleOneTimeAndRecurringProductsIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	err := s.checkCreatedOrderDetailsAndActionLogs(ctx, pb.OrderType_ORDER_TYPE_NEW)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*pb.CreateOrderRequest)
	res := stepState.Response.(*pb.CreateOrderResponse)

	orderItems, err := s.getOrderItems(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	foundOrderItem := countOrderItemForMultipleOneTimeAndRecurringProducts(orderItems, req.OrderItems)
	if foundOrderItem < len(req.OrderItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("create miss orderItem")
	}

	billingItems, err := s.getBillItems(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	foundBillItem := countBillItemForMultipleOneTimeAndRecurringProducts(billingItems, req.BillingItems, pb.BillingStatus_BILLING_STATUS_BILLED, pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER, req.LocationId)
	if foundBillItem < len(req.BillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing billing item")
	}

	foundUpcomingBillItem := countBillItemForMultipleOneTimeAndRecurringProducts(billingItems, req.UpcomingBillingItems, pb.BillingStatus_BILLING_STATUS_PENDING, pb.BillingType_BILLING_TYPE_UPCOMING_BILLING, req.LocationId)
	if foundUpcomingBillItem < len(req.UpcomingBillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing upcoming billing item")
	}

	return StepStateToContext(ctx, stepState), nil
}

func countOrderItemForMultipleOneTimeAndRecurringProducts(dbOrderItems []entities.OrderItem, orderItems []*pb.OrderItem) int {
	foundOrderItem := 0
	for _, item := range orderItems {
		for _, dbItem := range dbOrderItems {
			if item.ProductId == dbItem.ProductID.String &&
				((dbItem.DiscountID.Status == pgtype.Null) ||
					(dbItem.DiscountID.Status == pgtype.Present && dbItem.DiscountID.String == item.DiscountId.Value)) &&
				((dbItem.StartDate.Status == pgtype.Null) ||
					(dbItem.StartDate.Status == pgtype.Present && dbItem.StartDate.Time.Equal(item.StartDate.AsTime()))) {
				foundOrderItem++
			}
		}
	}
	return foundOrderItem
}

func countBillItemForMultipleOneTimeAndRecurringProducts(billItems []entities.BillItem, billingItems []*pb.BillingItem, billingStatus pb.BillingStatus, billingType pb.BillingType, locationID string) int {
	foundBillItem := 0
	for _, item := range billingItems {
		for _, dbItem := range billItems {
			if item.ProductId == dbItem.ProductID.String &&
				dbItem.BillStatus.String == billingStatus.String() &&
				dbItem.BillType.String == billingType.String() &&
				IsEqualNumericAndFloat32(dbItem.FinalPrice, item.FinalPrice) &&
				dbItem.LocationID.String == locationID {
				foundBillItem++
			}
		}
	}
	return foundBillItem
}
