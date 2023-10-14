package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForCreateOrderWithdrawRecurringFee(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            true,
		InsertProductLocation:         true,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertMaterial:                false,
		InsertFee:                     true,
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
		insertOrderReq, billItems, err = s.createRecurringFeeDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalRequestDisabledProrating(&insertOrderReq, billItems)
	case "valid graduate request with disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringFeeDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validGraduateRequestDisabledProrating(&insertOrderReq, billItems)
	case "empty billed at order disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringFeeDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyBilledAtOrderDisabledProrating(&insertOrderReq, billItems)
	case "empty upcoming billing disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		insertOrderReq, billItems, err = s.createRecurringFeeDisabledProratingEmptyUpcomingBilling(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyUpcomingBillingDisabledProrating(&insertOrderReq, billItems)
	case "empty billed at order with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringFeeWithProratingAndDiscount(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyBilledAtOrderWithProratingAndDiscount(&insertOrderReq, billItems)
	case "empty upcoming billing with prorating and discount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		insertOrderReq, billItems, err = s.createRecurringFeeWithProratingAndDiscountEmptyUpcomingBilling(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyUpcomingBillingWithProratingAndDiscount(&insertOrderReq, billItems)
	case "empty billing items":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		insertOrderReq, billItems, err = s.createRecurringFeeForWithdrawEmptyBillItems(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.validWithdrawalEmptyBillItems(&insertOrderReq, billItems)
	case "out of version":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringFeeDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		withdrawOrderReq = s.invalidWithdrawalRequestDisabledProratingWithOutOfVersion(&insertOrderReq, billItems)
	case "non-enrolled status":
		defaultOptionPrepareData.InsertPotentialStudent = true
		withdrawOrderReq = s.invalidWithdrawalRequestNonEnrolledStudent(ctx, defaultOptionPrepareData)
	}

	leavingReasonIDs, err := s.insertLeavingReasonsAndReturnID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	withdrawOrderReq.LeavingReasonIds = []string{leavingReasonIDs[0]}

	stepState.Request = &withdrawOrderReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderWithdrawRecurringFeeSuccessFor(ctx context.Context, statusType string, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch statusType {
	case "successfully":
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

	case "unsuccessfully":
		if testcase == "out of version" && stepState.ResponseErr.Error() != status.Error(codes.FailedPrecondition, constant.OptimisticLockingEntityVersionMismatched).Error() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("valid version number")
		}
		if testcase == "non-enrolled status" && stepState.ResponseErr.Error() != status.Error(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatusUnavailable).Error() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("valid student status")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createRecurringFeeDisabledProrating(
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

func (s *suite) createRecurringFeeWithProratingAndDiscount(
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

func (s *suite) createRecurringFeeWithProratingAndDiscountEmptyUpcomingBilling(
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

func (s *suite) createRecurringFeeDisabledProratingEmptyUpcomingBilling(
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

func (s *suite) createRecurringFeeForWithdrawEmptyBillItems(
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

func (s *suite) invalidWithdrawalRequestDisabledProratingWithOutOfVersion(oldRequest *pb.CreateOrderRequest, billItems []*entities.BillItem) (req pb.CreateOrderRequest) {
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
			StudentProductVersionNumber: int32(5),
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

func (s *suite) invalidWithdrawalRequestNonEnrolledStudent(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest) {
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, options, "")
	if err != nil {
		return
	}

	err = mockdata.UpdateStudentStatus(ctx, s.FatimaDBTrace, data.PotentialUserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_POTENTIAL", time.Now().AddDate(0, -1, 0), time.Now().AddDate(0, 1, 0))
	if err != nil {
		return
	}

	req.StudentId = data.PotentialUserID
	req.LocationId = data.LocationID
	req.OrderComment = "test withdrawal with non-enrolled student"
	req.OrderItems = []*pb.OrderItem{}
	req.BillingItems = []*pb.BillingItem{}
	req.UpcomingBillingItems = []*pb.BillingItem{}
	req.OrderType = pb.OrderType_ORDER_TYPE_WITHDRAWAL
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}
	req.EffectiveDate = &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()}

	return
}

func (s *suite) insertLeavingReasonsAndReturnID(ctx context.Context) (leavingReasonIDs []string, err error) {
	for i := 0; i < 3; i++ {
		randomStr := idutil.ULIDNow()
		name := fmt.Sprintf("Cat " + randomStr)
		leavingReasonType := database.Text("1")
		remarks := fmt.Sprintf("Remark " + randomStr)
		isArchived := true
		stmt := `INSERT INTO leaving_reason
		(leaving_reason_id, name, leaving_reason_type,  remark, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())`

		_, err := s.FatimaDBTrace.Exec(ctx, stmt, randomStr, name, leavingReasonType, remarks, isArchived)
		if err != nil {
			return nil, fmt.Errorf("cannot insert leaving_reason, err: %s", err)
		}
		leavingReasonIDs = append(leavingReasonIDs, randomStr)
	}
	return
}
