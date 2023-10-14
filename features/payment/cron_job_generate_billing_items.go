package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	service "github.com/manabie-com/backend/internal/payment/services/internal_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForScheduledGenerationOfBillItemsRecurringMaterialWith(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(10000)

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            true,
		InsertProductLocation:         true,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertFee:                     false,
		InsertBillingSchedule:         true,
		InsertBillingScheduleArchived: false,
		IsShorterPeriod:               true,
		IsTaxExclusive:                false,
		InsertDiscountNotAvailable:    false,
		InsertProductOutOfTime:        false,
		BillingScheduleStartDate:      time.Now(),
	}

	req := pb.CreateOrderRequest{}
	var (
		err error
	)

	switch testcase {
	case "order with single billed at order item":
		defaultOptionPrepareData.InsertMaterial = true
		req, err = s.validCaseBilledAtOrderItemsSingleItemExpected(ctx, defaultOptionPrepareData)
	case "order with single billed at order item unique material":
		defaultOptionPrepareData.InsertMaterialUnique = true
		req, err = s.validCaseBilledAtOrderItemsSingleItemExpected(ctx, defaultOptionPrepareData)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) nextBillingDateIsWithin30Days(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	ctx, err = s.signedAsAccount(ctx, UserGroupSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.createOrderForGeneratingNextBillItems(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.theScheduledGenerateBillingItemsIsTriggered(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theScheduledGenerateBillingItemsIsTriggered(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.GenerateBillingItemsRequest{}

	stepState.RequestSentAt = time.Now()
	internalService := service.NewInternalService(s.FatimaDBTrace, s.JSM, s.Kafka, s.Cfg.Common)
	stepState.Response, stepState.ResponseErr = internalService.GenerateBillingItems(contextWithToken(ctx), stepState.Request.(*pb.GenerateBillingItemsRequest))
	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(5000)
		stepState.Response, stepState.ResponseErr = internalService.GenerateBillingItems(contextWithToken(ctx), stepState.Request.(*pb.GenerateBillingItemsRequest))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) nextBillingItemsAreGenerated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	billingItems, err := s.getBillItems(ctx, stepState.OrderID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(billingItems) == stepState.NumberOfBillItems {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failure to generate next billing items")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderForGeneratingNextBillItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateOrderRequest))
	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(5000)
		stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateOrderRequest))
	}
	billingItems, err := s.getBillItems(ctx, stepState.Response.(*pb.CreateOrderResponse).OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.NumberOfBillItems = len(billingItems)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validCaseBilledAtOrderItemsSingleItemExpectedWithSimpleDiscount(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	var (
		orderItems           []*pb.OrderItem
		billedAtOrderItems   []*pb.BillingItem
		upcomingBillingItems []*pb.BillingItem
	)

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[1]},
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
				TaxAmount:     getInclusivePercentTax(PriceOrder-10, 20),
			},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
			FinalPrice: PriceOrder - 10,
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
				TaxAmount:     getInclusivePercentTax(PriceOrder-10, 20),
			},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
			FinalPrice: PriceOrder - 10,
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order recurring product"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}
