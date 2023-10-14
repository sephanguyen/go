package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForCreateLOARequest(ctx context.Context, testcase string) (context.Context, error) {
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
		InsertNotificationDate:        true,
	}
	var (
		insertOrderReq pb.CreateOrderRequest
		orderReq       pb.CreateOrderRequest
		billItems      []*entities.BillItem
		err            error
	)

	switch testcase {
	case "valid LOA request with disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		orderReq = s.validLOARequestDisabledProrating(&insertOrderReq, billItems)
	case "valid LOA request with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialWithProratingAndDiscount(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		orderReq = s.validLOARequestWithProratingAndDiscount(&insertOrderReq, billItems)
	case "empty billed at order with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialWithProratingAndDiscount(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		orderReq = s.validLOARequestEmptyBilledAtOrderWithProratingAndDiscount(&insertOrderReq, billItems)
	case "empty upcoming billing with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialWithProratingAndDiscountEmptyUpcomingBilling(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		orderReq = s.validLOARequestEmptyUpcomingBillingWithProratingAndDiscount(&insertOrderReq, billItems)
	case "no active recurring products":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()

		orderReq = s.validLOARequestNoActiveRecurringProduct(ctx, defaultOptionPrepareData)
	}

	leavingReasonIDs, err := s.insertLeavingReasonsAndReturnID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	orderReq.LeavingReasonIds = []string{leavingReasonIDs[0]}

	stepState.Request = &orderReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreateLOARequestWithPausableTag(ctx context.Context, pausable string) (context.Context, error) {
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
		InsertLeavingReasons:          true,
		InsertNotificationDate:        true,
		InsertProductSetting:          false,
	}
	var (
		insertOrderReq pb.CreateOrderRequest
		orderReq       pb.CreateOrderRequest
		billItems      []*entities.BillItem
		productSetting entities.ProductSetting
		err            error
	)

	switch pausable {
	case "true":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		orderReq = s.validLOARequestDisabledProrating(&insertOrderReq, billItems)
		productSetting = entities.ProductSetting{
			ProductID:                    pgtype.Text{String: insertOrderReq.OrderItems[0].ProductId, Status: pgtype.Present},
			IsPausable:                   pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsEnrollmentRequired:         pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsOperationFee:               pgtype.Bool{Bool: false, Status: pgtype.Present},
		}
		err := mockdata.InsertProductSetting(ctx, s.FatimaDBTrace, productSetting)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

	case "false":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		orderReq = s.validLOARequestDisabledProrating(&insertOrderReq, billItems)
		productSetting = entities.ProductSetting{
			ProductID:                    pgtype.Text{String: insertOrderReq.OrderItems[0].ProductId, Status: pgtype.Present},
			IsPausable:                   pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsEnrollmentRequired:         pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsOperationFee:               pgtype.Bool{Bool: false, Status: pgtype.Present},
		}
		err := mockdata.InsertProductSetting(ctx, s.FatimaDBTrace, productSetting)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	stepState.Request = &orderReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) pausableTagValidatedSuccessfully(ctx context.Context, pausable string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if pausable == "false" {
		if stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "Internal") {
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("product with not pausable tag allowed for LOA")
	}

	return s.createLOARequestSuccess(ctx)
}

func (s *suite) createLOARequestSuccess(ctx context.Context) (context.Context, error) {
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

	foundOrderItem := countOrderItemForRecurringProduct(orderItems, req.OrderItems)
	if foundOrderItem < len(req.OrderItems) && req.OrderItems[0].StudentProductId != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing order item")
	}

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

	for _, item := range orderItems {
		if time.Now().AddDate(0, 0, -1).After(item.StartDate.Time) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid LOA start date saved to order")
		}
		if !req.OrderItems[0].EndDate.AsTime().Equal(item.EndDate.Time) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid LOA end date saved to order")
		}
	}

	if len(req.OrderItems) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	studentProducts, err := s.getListStudentProductBaseOnProductID(ctx, req.OrderItems[0].ProductId)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve student products for LOA")
	}

	if req.OrderItems[0].StartDate.AsTime().Day() == time.Now().Day() &&
		studentProducts[0].StudentProductLabel.String != pb.StudentProductLabel_PAUSED.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student product label not updated to PAUSED")
	} else if req.OrderItems[0].StartDate.AsTime().Day() != time.Now().Day() &&
		studentProducts[0].StudentProductLabel.String != pb.StudentProductLabel_PAUSE_SCHEDULED.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student product label not updated to PAUSE_SCHEDULED")
	}

	err = s.checkLOANotifications(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot publish notification by kafka: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkLOANotifications(ctx context.Context, orderID string) (err error) {
	var systemNotificationEntity *model.SystemNotification
	err = try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)
		systemNotificationEntity = &model.SystemNotification{}
		fieldNames, fieldValues := systemNotificationEntity.FieldMap()
		query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE reference_id = $1 AND deleted_at IS NULL`,
			strings.Join(fieldNames, ","), systemNotificationEntity.TableName())
		err = s.NotificationMgmtDBTrace.QueryRow(ctx, query, orderID).Scan(fieldValues...)
		if err == nil {
			return false, nil
		}
		retry := attempt <= 3
		if retry {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return err
	}
	return
}

func (s *suite) validLOARequestDisabledProrating(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	loaStartDate := &timestamppb.Timestamp{Seconds: time.Now().Unix()}
	loaEndDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 0).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: oldRequest.OrderItems[0].ProductId,
			StartDate: loaStartDate,
			EndDate:   loaEndDate,
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
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -PriceOrder},
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
	req.OrderComment = "test LOA request"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_LOA
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.StartDate = loaStartDate
	req.EndDate = loaEndDate
	req.LeavingReasonIds = oldRequest.LeavingReasonIds
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}

	return
}

func (s *suite) validLOARequestWithProratingAndDiscount(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	loaStartDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 9).Unix()}
	loaEndDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 2, 9).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  oldRequest.OrderItems[0].ProductId,
			DiscountId: oldRequest.OrderItems[0].DiscountId,
			StartDate:  loaStartDate,
			EndDate:    loaEndDate,
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
	req.OrderComment = "test LOA request"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_LOA
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.StartDate = loaStartDate
	req.EndDate = loaEndDate
	req.LeavingReasonIds = oldRequest.LeavingReasonIds
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}
	return
}

func (s *suite) validLOARequestEmptyBilledAtOrderWithProratingAndDiscount(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	loaStartDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 9).Unix()}
	loaEndDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 2, 9).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  oldRequest.OrderItems[0].ProductId,
			DiscountId: oldRequest.OrderItems[0].DiscountId,
			StartDate:  loaStartDate,
			EndDate:    loaEndDate,
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
	req.OrderComment = "test LOA request"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_LOA
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.StartDate = loaStartDate
	req.EndDate = loaEndDate
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}

	return
}

func (s *suite) validLOARequestEmptyUpcomingBillingWithProratingAndDiscount(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	loaStartDate := &timestamppb.Timestamp{Seconds: time.Now().Unix()}
	loaEndDate := &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 0).Unix()}
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  oldRequest.OrderItems[0].ProductId,
			DiscountId: oldRequest.OrderItems[0].DiscountId,
			StartDate:  loaStartDate,
			EndDate:    loaEndDate,
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
	req.OrderComment = "test LOA request"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_LOA
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.StartDate = loaStartDate
	req.EndDate = loaEndDate
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}

	return
}

func (s *suite) validLOARequestNoActiveRecurringProduct(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test LOA request"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_LOA
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.StartDate = &timestamppb.Timestamp{Seconds: time.Now().Unix()}
	req.EndDate = &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 1, 0).Unix()}
	req.StudentDetailPath = &wrapperspb.StringValue{Value: "/students/student-id-1/details"}

	return
}
