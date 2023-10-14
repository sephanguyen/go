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

func (s *suite) prepareDataForCreateOrderWithdrawRecurringMaterial(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            true,
		InsertProductLocation:         true,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertFee:                     false,
		InsertMaterial:                true,
		InsertBillingSchedule:         true,
		InsertBillingScheduleArchived: false,
		IsTaxExclusive:                false,
		InsertDiscountNotAvailable:    false,
		InsertProductOutOfTime:        false,
		InsertProductDiscount:         true,
		BillingScheduleStartDate:      time.Now(),
		InsertProductSetting:          true,
	}
	var (
		insertOrderReq   pb.CreateOrderRequest
		withdrawOrderReq pb.CreateOrderRequest
		billItems        []*entities.BillItem
		err              error
	)

	switch testcase {
	case "valid withdrawal request with disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestDisabledProrating(&insertOrderReq, billItems)
	case "valid graduate request with disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validGraduateRequestDisabledProrating(&insertOrderReq, billItems)
	case "empty billed at order disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyBilledAtOrderDisabledProrating(&insertOrderReq, billItems)
	case "empty upcoming billing disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProratingEmptyUpcomingBilling(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyUpcomingBillingDisabledProrating(&insertOrderReq, billItems)
	case "empty billed at order with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialWithProratingAndDiscount(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyBilledAtOrderWithProratingAndDiscount(&insertOrderReq, billItems)
	case "empty upcoming billing with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialWithProratingAndDiscountEmptyUpcomingBilling(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyUpcomingBillingWithProratingAndDiscount(&insertOrderReq, billItems)
	case "empty billing items":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		insertOrderReq, billItems, err = s.createRecurringMaterialForWithdrawEmptyBillItems(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyBillItems(&insertOrderReq, billItems)
	case "created and withdrawn on same ratio":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		insertOrderReq, billItems, err = s.createRecurringMaterialWithProratingAndDiscountProratedFirstBilling(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestCreatedAndWithdrawnOnSameRatio(&insertOrderReq, billItems)
	case "duplicate products":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialForWithdrawDuplicateProducts(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalDuplicateProducts(&insertOrderReq, billItems)
	case "withdrawal with no products":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		withdrawOrderReq = s.validWithdrawalEmptyProducts(ctx, defaultOptionPrepareData)
	}

	leavingReasonIDs, err := s.insertLeavingReasonsAndReturnID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	withdrawOrderReq.LeavingReasonIds = []string{leavingReasonIDs[0]}

	stepState.Request = &withdrawOrderReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderWithdrawRecurringMaterialSuccess(ctx context.Context) (context.Context, error) {
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

	order, err := s.getOrder(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if time.Now().AddDate(0, 0, -1).After(order.WithdrawalEffectiveDate.Time) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid withdrawal effective date saved to order")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validWithdrawalRequestCreatedAndWithdrawnOnSameRatio(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 27).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			DiscountId:    oldRequest.OrderItems[0].DiscountId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)
	// prorating disabled
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.BillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder / 4,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[0].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     oldRequest.BillingItems[0].TaxItem.TaxAmount,
			},
			FinalPrice:      getPercentDiscountedPrice(PriceOrder/4, 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(PriceOrder/4, 10)},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          oldRequest.BillingItems[0].DiscountItem.DiscountId,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder/4, 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.UpcomingBillingItems[0].TaxItem.TaxId,
				TaxPercentage: oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.UpcomingBillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     oldRequest.UpcomingBillingItems[0].TaxItem.TaxAmount,
			},
			FinalPrice:      getPercentDiscountedPrice(PriceOrder, 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -getPercentDiscountedPrice(PriceOrder, 10)},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          oldRequest.UpcomingBillingItems[0].DiscountItem.DiscountId,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalRequestDisabledProrating(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 1).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)
	// prorating disabled
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.BillingItems[2].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[2].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[2].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: 0},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.UpcomingBillingItems[0].TaxItem.TaxId,
				TaxPercentage: oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.UpcomingBillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -PriceOrder},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validGraduateRequestDisabledProrating(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 1).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)
	// prorating disabled
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.BillingItems[2].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[2].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[2].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: 0},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.UpcomingBillingItems[0].TaxItem.TaxId,
				TaxPercentage: oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.UpcomingBillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -PriceOrder},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test graduate recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_GRADUATE
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalEmptyBilledAtOrderDisabledProrating(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 10).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)
	// prorating disabled
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[2].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[2].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: 0},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalEmptyBilledAtOrderWithProratingAndDiscount(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 9).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			DiscountId:    oldRequest.OrderItems[0].DiscountId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)
	// prorating enabled
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   getProratedPrice(PriceOrder, 1, 4),
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.UpcomingBillingItems[0].TaxItem.TaxId,
				TaxPercentage: oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.UpcomingBillingItems[0].TaxItem.TaxCategory,
				TaxAmount: getInclusivePercentTax(
					getPercentDiscountedPrice(getProratedPrice(PriceOrder, 1, 4), 10),
					oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage),
			},
			FinalPrice:      getPercentDiscountedPrice(getProratedPrice(PriceOrder, 1, 4), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -(oldRequest.UpcomingBillingItems[0].FinalPrice * (1.0 / 2.0))},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          oldRequest.UpcomingBillingItems[0].DiscountItem.DiscountId,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getProratedPrice(PriceOrder, 1, 4), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalEmptyUpcomingBillingDisabledProrating(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 1).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)
	// prorating disabled
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.BillingItems[3].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[3].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[3].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[3].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[3].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: 0},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalEmptyUpcomingBillingWithProratingAndDiscount(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 1).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			DiscountId:    oldRequest.OrderItems[0].DiscountId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)
	// prorating disabled
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
			BillingSchedulePeriodId: oldRequest.BillingItems[3].BillingSchedulePeriodId,
			Price:                   getProratedPrice(PriceOrder, 0, 4),
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[3].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[3].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[3].TaxItem.TaxCategory,
				TaxAmount: getInclusivePercentTax(
					getPercentDiscountedPrice(getProratedPrice(PriceOrder, 0, 4), 10),
					oldRequest.BillingItems[3].TaxItem.TaxPercentage),
			},
			FinalPrice:      getPercentDiscountedPrice(getProratedPrice(PriceOrder, 0, 4), 10),
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -(oldRequest.BillingItems[3].FinalPrice * (3.0 / 4.0))},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          oldRequest.BillingItems[3].DiscountItem.DiscountId,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(getProratedPrice(PriceOrder, 0, 4), 10),
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalEmptyBillItems(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 3, 2).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalDuplicateProducts(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	effectiveDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[0].StudentProductID.String,
			},
		},
		&pb.OrderItem{
			ProductId:     oldRequest.OrderItems[0].ProductId,
			EffectiveDate: effectiveDate,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItems[5].StudentProductID.String,
			},
		},
	)

	billedAtOrderItems = append(billedAtOrderItems,
		// 3/4 prorating
		&pb.BillingItem{
			ProductId:               oldRequest.OrderItems[0].ProductId,
			StudentProductId:        wrapperspb.String(billItems[0].StudentProductID.String),
			BillingSchedulePeriodId: oldRequest.BillingItems[2].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[2].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[2].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -PriceOrder * (1.0 / 2.0)},
		},
		// 3/4 prorating
		&pb.BillingItem{
			ProductId:               oldRequest.OrderItems[0].ProductId,
			StudentProductId:        wrapperspb.String(billItems[5].StudentProductID.String),
			BillingSchedulePeriodId: oldRequest.BillingItems[2].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: oldRequest.BillingItems[2].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.BillingItems[2].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -(PriceOrder - 50) * (3.0 / 4.0)},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               oldRequest.OrderItems[0].ProductId,
			StudentProductId:        wrapperspb.String(billItems[0].StudentProductID.String),
			BillingSchedulePeriodId: oldRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.UpcomingBillingItems[0].TaxItem.TaxId,
				TaxPercentage: oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.UpcomingBillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -PriceOrder},
		},
		&pb.BillingItem{
			ProductId:               oldRequest.OrderItems[0].ProductId,
			StudentProductId:        wrapperspb.String(billItems[5].StudentProductID.String),
			BillingSchedulePeriodId: oldRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldRequest.UpcomingBillingItems[0].TaxItem.TaxId,
				TaxPercentage: oldRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   oldRequest.UpcomingBillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, oldRequest.BillingItems[2].TaxItem.TaxPercentage),
			},
			FinalPrice:      PriceOrder,
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -(PriceOrder - 50)},
		},
	)

	req.StudentId = oldRequest.StudentId
	req.LocationId = oldRequest.LocationId
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = effectiveDate

	return
}

func (s *suite) validWithdrawalEmptyProducts(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test withdraw recurring material"
	req.OrderItems = []*pb.OrderItem{}
	req.BillingItems = []*pb.BillingItem{}
	req.UpcomingBillingItems = []*pb.BillingItem{}
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 1).Unix()}

	return
}

func (s *suite) createRecurringMaterialDisabledProrating(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: data.ProductIDs[1],
			StartDate: &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[1],
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
			ProductId:               data.ProductIDs[1],
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
			ProductId:               data.ProductIDs[1],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
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
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[1],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
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
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order recurring product"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	req.LeavingReasonIds = data.LeavingReasonIDs
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}

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

func (s *suite) createRecurringMaterialWithProratingAndDiscount(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
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
	req.LeavingReasonIds = data.LeavingReasonIDs
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}

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

func (s *suite) createRecurringMaterialWithProratingAndDiscountProratedFirstBilling(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 26).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		// with 1/4 prorating
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getProratedPrice(PriceOrder, 1, 4),
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder/4, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder/4, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder/4, 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
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

func (s *suite) createRecurringMaterialWithProratingAndDiscountEmptyUpcomingBilling(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
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
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}

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

func (s *suite) createRecurringMaterialDisabledProratingEmptyUpcomingBilling(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: data.ProductIDs[1],
			StartDate: &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[1],
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
			ProductId:               data.ProductIDs[1],
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
			ProductId:               data.ProductIDs[1],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
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
			ProductId:               data.ProductIDs[1],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
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
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order recurring product"
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

func (s *suite) createRecurringMaterialForWithdrawEmptyBillItems(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: data.ProductIDs[1],
			StartDate: &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 5).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[1],
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
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[1],
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
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order recurring product"
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

func (s *suite) createRecurringMaterialForWithdrawDuplicateProducts(
	ctx context.Context,
	options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct,
) (
	req pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
	var (
		insertOrderResp       *pb.CreateOrderResponse
		tmpBillItems          []*entities.BillItem
		createOrderReq        []*pb.CreateOrderRequest
		reqDuplicateProductID pb.CreateOrderRequest
	)

	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: data.ProductIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
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
			ProductId:               data.ProductIDs[0],
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
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
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
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
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 15).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[3]},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		// with 3/4 prorating
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   PriceOrder * (3.0 / 4.0),
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder*(3.0/4.0), 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder*(3.0/4.0), 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder*(3.0/4.0), 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getPercentDiscountedPrice(PriceOrder, 10), 20),
			},
			FinalPrice: getPercentDiscountedPrice(PriceOrder, 10),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[3],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 10,
				DiscountAmount:      getPercentDiscountValue(PriceOrder, 10),
			},
		},
	)

	reqDuplicateProductID.StudentId = data.UserID
	reqDuplicateProductID.LocationId = data.LocationID
	reqDuplicateProductID.OrderComment = "test create order recurring product"
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

	for i, orderReq := range createOrderReq {
		insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), orderReq)
		for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			time.Sleep(1000)
			insertOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
				CreateOrder(contextWithToken(ctx), orderReq)
		}
		if err != nil {
			fmt.Println(i, err)
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

func (s *suite) getBillItemsByOrderIDAndProductID(ctx context.Context, orderID string, productID string) ([]*entities.BillItem, error) {
	billItem := &entities.BillItem{}
	billItems := []*entities.BillItem{}
	billItemFieldNames, billItemFieldValues := billItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1 AND product_id = $2
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		billItem.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, orderID, productID)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(billItemFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, billItem)
	}
	return billItems, nil
}
