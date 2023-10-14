package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForCreateOrderWithdrawScheduleBasePackage(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                        true,
		InsertDiscount:                   true,
		InsertStudent:                    true,
		InsertProductPrice:               false,
		InsertProductLocation:            true,
		InsertLocation:                   false,
		InsertProductGrade:               true,
		InsertFee:                        false,
		InsertMaterial:                   false,
		InsertBillingSchedule:            true,
		InsertBillingScheduleArchived:    false,
		IsTaxExclusive:                   false,
		InsertDiscountNotAvailable:       false,
		InsertProductOutOfTime:           false,
		InsertPackageCourses:             true,
		InsertPackageCourseScheduleBased: false,
		InsertProductDiscount:            true,
		BillingScheduleStartDate:         time.Now(),
	}

	var (
		insertOrderReq   pb.CreateOrderRequest
		withdrawOrderReq pb.CreateOrderRequest
		billItems        []*entities.BillItem
		data             mockdata.DataForRecurringProduct
		err              error
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	switch testcase {
	case "valid withdrawal request disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, data, err = s.createScheduleBasePackageForWithdrawalDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestScheduleBasePackageDisabledProrating(&insertOrderReq, billItems, data)
	case "valid withdrawal request with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, data, err = s.createScheduleBasePackageForWithdrawal(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestScheduleBasePackage(&insertOrderReq, billItems, data)
	case "valid graduate request with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, data, err = s.createScheduleBasePackageForWithdrawal(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validGraduationRequestScheduleBasePackage(&insertOrderReq, billItems, data)
	case "empty billed at order with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, data, err = s.createScheduleBasePackageForWithdrawal(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestScheduleBasePackageEmptyBilledAtOrder(&insertOrderReq, billItems, data)
	case "empty upcoming billing with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		insertOrderReq, billItems, data, err = s.createScheduleBasePackageForWithdrawalEmptyUpcomingBilling(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestScheduleBasePackageEmptyUpcomingBilling(&insertOrderReq, billItems, data)
	case "empty billing items":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		insertOrderReq, billItems, data, err = s.createScheduleBasePackageForWithdrawalEmptyBillingItems(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestScheduleBasePackageEmptyBillingItems(&insertOrderReq, billItems, data)
	case "duplicate products":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		insertOrderReq, billItems, data, err = s.createScheduleBasePackageForWithdrawalDuplicateProducts(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestScheduleBasePackageDuplicateProducts(&insertOrderReq, billItems, data)
	}

	leavingReasonIDs, err := s.insertLeavingReasonsAndReturnID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	withdrawOrderReq.LeavingReasonIds = []string{leavingReasonIDs[0]}

	stepState.Request = &withdrawOrderReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderWithdrawScheduleBasePackageSuccess(ctx context.Context) (context.Context, error) {
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

	billingItems, err := s.getBillItems(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	foundBillItem := countBillItemForRecurringProduct(billingItems, req.BillingItems, pb.BillingStatus_BILLING_STATUS_BILLED, pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING, req.LocationId)
	if foundBillItem < len(req.BillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing billing item")
	}

	foundUpcomingBillItem := countBillItemForRecurringProduct(billingItems, req.UpcomingBillingItems, pb.BillingStatus_BILLING_STATUS_PENDING, pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING, req.LocationId)
	if foundUpcomingBillItem < len(req.UpcomingBillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing upcoming billing item")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validWithdrawalRequestScheduleBasePackageDuplicateProducts(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem, data mockdata.DataForRecurringProduct) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 17).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
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
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[2].StudentProductID.String,
			},
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
		// with 1/2 prorating
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2) * (1 / 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2)*(1/2), 20),
			},
			FinalPrice:      getScheduleBasePrice(100, 2) * (1 / 2),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * (3.0 / 4.0)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 0)*(1/2), 10),
			},
		},
		// with 1/2 prorating
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[2].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2) * (1 / 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2)*(1/2), 20),
			},
			FinalPrice:      getScheduleBasePrice(100, 2) * (1 / 2),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * (3.0 / 4.0)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 0)*(1/2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[2].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalRequestScheduleBasePackageEmptyBillingItems(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem, data mockdata.DataForRecurringProduct) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 3, 0).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
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

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalRequestScheduleBasePackageEmptyUpcomingBilling(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem, data mockdata.DataForRecurringProduct) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 15).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
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
		// with 1/2 prorating
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2) * (1 / 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10)*(1/2), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * 1 / 2,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * (3.0 / 4.0)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2)*(1/2), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalRequestScheduleBasePackageEmptyBilledAtOrder(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem, data mockdata.DataForRecurringProduct) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 15).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
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
		// with 1/2 prorating
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2) * (1 / 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10)*(1/2), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * 1 / 2,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * (3.0 / 4.0)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2)*(1/2), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalRequestScheduleBasePackageDisabledProrating(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem, data mockdata.DataForRecurringProduct) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 15).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
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
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 0), 10)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalRequestScheduleBasePackage(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem, data mockdata.DataForRecurringProduct) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 15).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
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
		// with 1/2 prorating
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2) * (1 / 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2)*(1/2), 20),
			},
			FinalPrice:      getScheduleBasePrice(100, 2) * (1 / 2),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * (3.0 / 4.0)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 0)*(1/2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validGraduationRequestScheduleBasePackage(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem, data mockdata.DataForRecurringProduct) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 15).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     data.ProductIDs[0],
			EffectiveDate: effectiveDate,
			DiscountId:    &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
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
		// with 1/2 prorating
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2) * (1 / 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 2)*(1/2), 20),
			},
			FinalPrice:      getScheduleBasePrice(100, 2) * (1 / 2),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10) * (3.0 / 4.0)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 0)*(1/2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: data.ProductIDs[0],
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice:      getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10)},
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test graduate schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_GRADUATE
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) createScheduleBasePackageForWithdrawalDisabledProrating(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	data mockdata.DataForRecurringProduct,
	err error,
) {
	options.InsertPackageCourses = true
	options.InsertPackageCourseScheduleBased = true

	data, err = mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	err = mockdata.UpdateDisabledProratingFlagByProductID(ctx, s.FatimaDBTrace, data.ProductIDs[0], true)
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return
	}

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	insertOrderResp, err := pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), &req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), &req)
	}
	if err != nil {
		return
	}
	billItems, err = s.getBillItemsByOrderIDAndProductID(ctx, insertOrderResp.OrderId, req.OrderItems[0].ProductId)
	if err != nil {
		return
	}

	return
}

func (s *suite) createScheduleBasePackageForWithdrawal(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	data mockdata.DataForRecurringProduct,
	err error,
) {
	options.InsertPackageCourses = true
	options.InsertPackageCourseScheduleBased = true

	data, err = mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return
	}

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	insertOrderResp, err := pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), &req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), &req)
	}
	if err != nil {
		return
	}
	billItems, err = s.getBillItemsByOrderIDAndProductID(ctx, insertOrderResp.OrderId, req.OrderItems[0].ProductId)
	if err != nil {
		return
	}

	return
}

func (s *suite) createScheduleBasePackageForWithdrawalEmptyUpcomingBilling(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	data mockdata.DataForRecurringProduct,
	err error,
) {
	options.InsertPackageCourses = true
	options.InsertPackageCourseScheduleBased = true

	data, err = mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return
	}

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	insertOrderResp, err := pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), &req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), &req)
	}
	if err != nil {
		return
	}
	billItems, err = s.getBillItemsByOrderIDAndProductID(ctx, insertOrderResp.OrderId, req.OrderItems[0].ProductId)
	if err != nil {
		return
	}

	return
}

func (s *suite) createScheduleBasePackageForWithdrawalEmptyBillingItems(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	data mockdata.DataForRecurringProduct,
	err error,
) {
	options.InsertPackageCourses = true
	options.InsertPackageCourseScheduleBased = true

	data, err = mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test withdraw schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return
	}

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	insertOrderResp, err := pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), &req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), &req)
	}
	if err != nil {
		return
	}
	billItems, err = s.getBillItemsByOrderIDAndProductID(ctx, insertOrderResp.OrderId, req.OrderItems[0].ProductId)
	if err != nil {
		return
	}

	return
}

func (s *suite) createScheduleBasePackageForWithdrawalDuplicateProducts(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	data mockdata.DataForRecurringProduct,
	err error,
) {
	var (
		insertOrderResp       *pb.CreateOrderResponse
		tmpBillItems          []*entities.BillItem
		createOrderReq        []*pb.CreateOrderRequest
		reqDuplicateProductID pb.CreateOrderRequest
	)

	options.InsertPackageCourses = true
	options.InsertPackageCourseScheduleBased = true

	data, err = mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order recurring product"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	createOrderReq = append(createOrderReq, &req)

	orderItems = []*pb.OrderItem{}
	billedAtOrderItems = []*pb.BillingItem{}
	upcomingBillingItems = []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 2},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(getScheduleBasePrice(100, 2), 10),
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
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getScheduleBasePrice(100, 2), 10),
			},
		},
	)

	reqDuplicateProductID.StudentId = data.UserID
	reqDuplicateProductID.LocationId = data.LocationID
	reqDuplicateProductID.OrderComment = "test withdraw schedule-base package"
	reqDuplicateProductID.OrderItems = orderItems
	reqDuplicateProductID.BillingItems = billedAtOrderItems
	reqDuplicateProductID.UpcomingBillingItems = upcomingBillingItems
	reqDuplicateProductID.OrderType = pb.OrderType_ORDER_TYPE_NEW

	createOrderReq = append(createOrderReq, &reqDuplicateProductID)

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return
	}

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()

	for _, orderReq := range createOrderReq {
		insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), orderReq)
		for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			time.Sleep(1000)
			insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
				CreateOrder(contextWithToken(ctx), orderReq)
		}
		if err != nil {
			return
		}
		tmpBillItems, err = s.getBillItemsByOrderIDAndProductID(ctx, insertOrderResp.OrderId, orderReq.OrderItems[0].ProductId)
		if err != nil {
			return
		}

		billItems = append(billItems, tmpBillItems...)
	}

	return
}
