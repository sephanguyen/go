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

func (s *suite) createOrderWithRecurringProductsSuccessfully(ctx context.Context, account string, orderTypeUsecase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		req             pb.CreateOrderRequest
		createOrderResp *pb.CreateOrderResponse
	)

	switch orderTypeUsecase {
	case "new":
		ctx, err := s.signedAsAccount(ctx, account)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
			InsertTax:                     true,
			InsertDiscount:                true,
			InsertStudent:                 true,
			InsertProductPrice:            true,
			InsertProductLocation:         true,
			InsertLocation:                false,
			InsertProductGrade:            true,
			InsertFee:                     true,
			InsertMaterial:                false,
			InsertBillingSchedule:         true,
			InsertBillingScheduleArchived: false,
			IsShorterPeriod:               true,
			IsTaxExclusive:                false,
			InsertDiscountNotAvailable:    false,
			InsertProductOutOfTime:        false,
			InsertProductDiscount:         true,
			BillingScheduleStartDate:      time.Now(),
			InsertProductSetting:          true,
		}
		req, err = s.validCaseBilledAtOrderItemsSingleItemExpected(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "update":
		ctx, err := s.signedAsAccount(ctx, account)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
			InsertTax:                            true,
			InsertDiscount:                       true,
			InsertStudent:                        true,
			InsertProductPrice:                   false,
			InsertProductPriceWithDifferentPrice: true,
			InsertProductLocation:                true,
			InsertLocation:                       false,
			InsertProductGrade:                   true,
			InsertFee:                            false,
			InsertMaterial:                       true,
			InsertBillingSchedule:                true,
			InsertBillingScheduleArchived:        false,
			IsTaxExclusive:                       false,
			InsertDiscountNotAvailable:           false,
			InsertProductOutOfTime:               false,
			InsertProductDiscount:                true,
			BillingScheduleStartDate:             time.Now(),
			InsertProductSetting:                 true,
		}
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		createOrderReq, err := s.prepareRequestForCreateOrderWithRecurringProducts(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), &createOrderReq)
		for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			time.Sleep(1500)
			_, err = pb.NewOrderServiceClient(s.PaymentConn).
				CreateOrder(contextWithToken(ctx), &createOrderReq)
		}
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		oldStudentProduct, err := s.getStudentProductBaseOnProductID(ctx, createOrderReq.OrderItems[0].ProductId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req = s.prepareRequestForUpdateOrderWithRecurringProducts(&createOrderReq, oldStudentProduct)
	case "withdraw":
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
			InsertProductSetting:             true,
		}

		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		createOrderReq, billItems, data, err := s.createFrequencyBasePackageForWithdrawalDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req = s.validWithdrawalRequestFrequencyBasePackageDisabledProrating(&createOrderReq, billItems, data)
	default:
		err := fmt.Errorf("error when invalid order type usecase: %s ", orderTypeUsecase)
		return StepStateToContext(ctx, stepState), err
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	createOrderResp, err := client.CreateOrder(contextWithToken(ctx), &req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		createOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), &req)
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = &req
	stepState.Response = createOrderResp
	stepState.OrderID = createOrderResp.OrderId

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareRequestForUpdateOrderWithRecurringProducts(createOrderReq *pb.CreateOrderRequest, studentProduct *entities.StudentProduct) (updateOrderReq pb.CreateOrderRequest) {
	var (
		orderItems           []*pb.OrderItem
		billedAtOrderItems   []*pb.BillingItem
		upcomingBillingItems []*pb.BillingItem
	)

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:        createOrderReq.OrderItems[0].ProductId,
			DiscountId:       nil,
			EffectiveDate:    &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
			StudentProductId: wrapperspb.String(studentProduct.StudentProductID.String),
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               createOrderReq.OrderItems[0].ProductId,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: createOrderReq.BillingItems[2].BillingSchedulePeriodId.Value},
			Price:                   PriceOrder - 50,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         createOrderReq.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder-50, 20),
			},
			FinalPrice:       PriceOrder - 50,
			AdjustmentPrice:  wrapperspb.Float(7.5),
			StudentProductId: wrapperspb.String(studentProduct.StudentProductID.String),
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               createOrderReq.OrderItems[0].ProductId,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: createOrderReq.UpcomingBillingItems[0].BillingSchedulePeriodId.Value},
			Price:                   PriceOrder + 150,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         createOrderReq.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder+150, 20),
			},
			FinalPrice:       PriceOrder + 150,
			DiscountItem:     nil,
			AdjustmentPrice:  wrapperspb.Float(10),
			StudentProductId: wrapperspb.String(studentProduct.StudentProductID.String),
		},
	)
	updateOrderReq.StudentId = createOrderReq.StudentId
	updateOrderReq.LocationId = createOrderReq.LocationId
	updateOrderReq.OrderComment = "test create update order recurring products"
	updateOrderReq.OrderItems = orderItems
	updateOrderReq.BillingItems = billedAtOrderItems
	updateOrderReq.UpcomingBillingItems = upcomingBillingItems
	updateOrderReq.OrderType = pb.OrderType_ORDER_TYPE_UPDATE
	return
}

func (s *suite) prepareRequestForCreateOrderWithRecurringProducts(ctx context.Context, settings mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, settings, "")
	if err != nil {
		return
	}

	var orderItems []*pb.OrderItem
	var billedAtOrderItems []*pb.BillingItem
	var upcomingBillingItems []*pb.BillingItem

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[1]},
			StartDate:  &timestamppb.Timestamp{Seconds: settings.BillingScheduleStartDate.AddDate(0, 0, 18).Unix()},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getProratedPrice(PriceOrder, 1, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getProratedPrice(PriceOrder, 1, 2)-5, 20),
			},
			// billing ratio 1/2 applied
			FinalPrice: getProratedPrice(PriceOrder, 1, 2) - 5,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      5,
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   PriceOrder + 50,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder+50-10, 20),
			},
			FinalPrice: PriceOrder + 50 - 10,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   PriceOrder + 100,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder+100-10, 20),
			},
			FinalPrice: PriceOrder + 100 - 10,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   PriceOrder + 150,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder+150-10, 20),
			},
			FinalPrice: PriceOrder + 150 - 10,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order recurring products"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	return
}

func (s *suite) voidOrderWithRecurringProducts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	voidOrderReq := &pb.VoidOrderRequest{
		OrderId: stepState.OrderID,
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	voidOrderResp, err := client.VoidOrder(contextWithToken(ctx), voidOrderReq)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		voidOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			VoidOrder(contextWithToken(ctx), voidOrderReq)
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = voidOrderReq
	stepState.Response = voidOrderResp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) voidOrderWithRecurringProductsSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	voidedOrderResp := stepState.Response.(*pb.VoidOrderResponse)
	if !voidedOrderResp.Successful {
		return StepStateToContext(ctx, stepState), fmt.Errorf("void order failed")
	}

	order, err := s.getOrder(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if order.OrderStatus.String != pb.OrderStatus_ORDER_STATUS_VOIDED.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when order status wrong data")
	}

	billItems, err := s.getBillItems(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	orderItems, err := s.getOrderItems(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, billItem := range billItems {
		if billItem.BillStatus.String != pb.BillingStatus_BILLING_STATUS_CANCELLED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when bill item status wrong data")
		}
	}

	studentProductIDs := make([]string, 0, len(orderItems))
	for _, orderItem := range orderItems {
		studentProductIDs = append(studentProductIDs, orderItem.StudentProductID.String)
	}
	studentProducts, err := s.getStudentProductsByIDs(ctx, studentProductIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, studentProduct := range studentProducts {
		if (order.OrderType.String == pb.OrderType_ORDER_TYPE_WITHDRAWAL.String() ||
			order.OrderType.String == pb.OrderType_ORDER_TYPE_GRADUATE.String() ||
			order.OrderType.String == pb.OrderType_ORDER_TYPE_LOA.String()) &&
			studentProduct.ProductStatus.String != pb.StudentProductStatus_ORDERED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error student product status incorrect for withdrawa/graduate/loa order")
		}

		if !(order.OrderType.String == pb.OrderType_ORDER_TYPE_WITHDRAWAL.String() ||
			order.OrderType.String == pb.OrderType_ORDER_TYPE_GRADUATE.String() ||
			order.OrderType.String == pb.OrderType_ORDER_TYPE_LOA.String()) &&
			studentProduct.ProductStatus.String != pb.StudentProductStatus_CANCELLED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error student product status incorrect for new/update order")
		}
	}

	orderActionLogs, err := s.getOrderActionLogs(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	isVoidedAction := false

	for _, actionLog := range orderActionLogs {
		if actionLog.Action.String == pb.OrderActionStatus_ORDER_ACTION_VOIDED.String() && actionLog.UserID.String == stepState.CurrentUserID {
			isVoidedAction = true
		}
	}
	if !isVoidedAction {
		return StepStateToContext(ctx, stepState), fmt.Errorf("void order action log invalid content")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) voidOrderWithRecurringProductsOutOfVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	voidOrderReq := &pb.VoidOrderRequest{
		OrderId:            stepState.OrderID,
		OrderVersionNumber: int32(5),
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	voidOrderResp, err := client.VoidOrder(contextWithToken(ctx), voidOrderReq)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		voidOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			VoidOrder(contextWithToken(ctx), voidOrderReq)
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	stepState.Request = voidOrderReq
	stepState.Response = voidOrderResp

	return StepStateToContext(ctx, stepState), nil
}
