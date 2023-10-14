package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/status"
)

func (s *suite) prepareDataForCreateOrderRecurringMaterialWithValidRequest(ctx context.Context, testcase string) (context.Context, error) {
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
	}

	req := pb.CreateOrderRequest{}
	var err error

	switch testcase {
	case "order with discount applied fix amount with null recurring valid duration":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseCorrectDiscountAppliedFixAmountWithNullRecurringValidDuration(ctx, defaultOptionPrepareData)
	case "order with discount applied fix amount with finite recurring valid duration":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseCorrectDiscountAppliedFixAmountWithFiniteRecurringValidDuration(ctx, defaultOptionPrepareData)
	case "order with discount applied percent type with null recurring valid duration":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseCorrectDiscountAppliedPercentTypeWithNullRecurringValidDuration(ctx, defaultOptionPrepareData)
	case "order with discount applied percent type with finite recurring valid duration":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseCorrectDiscountAppliedPercentTypeWithFiniteRecurringValidDuration(ctx, defaultOptionPrepareData)
	case "order with discount and prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseCorrectFinalPriceWithDiscountAndProrating(ctx, defaultOptionPrepareData)
	case "order with empty billed at order items":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, 1, 0)
		req, err = s.validCaseBilledAtOrderItemsEmpty(ctx, defaultOptionPrepareData)
	case "order with single billed at order item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		req, err = s.validCaseBilledAtOrderItemsSingleItemExpected(ctx, defaultOptionPrepareData)
	case "order with multiple billed at order item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseBilledAtOrderItemsMultipleItemsExpected(ctx, defaultOptionPrepareData)
	case "order with prorating applied":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseBilledAtOrderItemsCorrectRatioApplied(ctx, defaultOptionPrepareData)
	case "order without prorating applied":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.validCaseBilledAtOrderItemsNoRatioAppliedDisabledProrating(ctx, defaultOptionPrepareData)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreateOrderRecurringMaterialWithInvalidRequest(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                  true,
		InsertDiscount:             true,
		InsertStudent:              true,
		InsertProductPrice:         true,
		InsertProductLocation:      true,
		InsertLocation:             false,
		InsertProductGrade:         true,
		InsertMaterial:             true,
		InsertBillingSchedule:      true,
		IsTaxExclusive:             false,
		InsertDiscountNotAvailable: false,
		InsertProductOutOfTime:     false,
		InsertProductDiscount:      true,
		BillingScheduleStartDate:   time.Now(),
	}

	req := pb.CreateOrderRequest{}
	var err error

	switch testcase {
	case "billed at order items no items added":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		req, err = s.invalidCaseBilledAtOrderItemsNoItemAdded(ctx, defaultOptionPrepareData)
	case "billed at order items should be empty":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, 1, 0)
		req, err = s.invalidCaseBilledAtOrderItemsShouldBeEmpty(ctx, defaultOptionPrepareData)
	case "incorrect first billing period added":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		req, err = s.invalidCaseBilledAtOrderItemsIncorrectBillingPeriodAdded(ctx, defaultOptionPrepareData)
	case "incorrect upcoming billing period added":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseBilledAtOrderItemsMissingBillingPeriod(ctx, defaultOptionPrepareData)
	case "incorrect upcoming billing multiple items added":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -1, 0)
		req, err = s.invalidCaseUpcomingBillingMultipleItemsAdded(ctx, defaultOptionPrepareData)
	case "incorrect ratio applied with prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -1, 0)
		req, err = s.invalidCaseBilledAtOrderItemsIncorrectRatioApplied(ctx, defaultOptionPrepareData)
	case "incorrect ratio applied with disabled prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -1, 0)
		req, err = s.invalidCaseBilledAtOrderItemsRatioAppliedOnDisabledProrating(ctx, defaultOptionPrepareData)
	case "incorrect final price missing discount billed at order":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseMissingDiscountBilledAtOrderFixAmountWithNullRecurringValidDuration(ctx, defaultOptionPrepareData)
	case "incorrect final price missing discount upcoming billing":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseMissingDiscountUpcomingBillingFixAmountWithNullRecurringValidDuration(ctx, defaultOptionPrepareData)
	case "incorrect tax percent applied":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		req, err = s.invalidCaseIncorrectTaxPercentApplied(ctx, defaultOptionPrepareData)
	case "incorrect tax amount applied without discount and prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -3, 0)
		req, err = s.invalidCaseIncorrectTaxAmountApplied(ctx, defaultOptionPrepareData)
	case "incorrect tax amount applied with discount and prorating":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseIncorrectTaxAmountAppliedWithDiscountAndProrating(ctx, defaultOptionPrepareData)
	case "maximum discount reached fix amount":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseMaximumDiscountReachedFixAmount(ctx, defaultOptionPrepareData)
	case "maximum discount reached percent type":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseMaximumDiscountReachedPercentType(ctx, defaultOptionPrepareData)
	case "new order with archived billing schedule":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		defaultOptionPrepareData.InsertBillingScheduleArchived = true
		req, err = s.invalidCaseNewOrderWithArchivedBillingSchedule(ctx, defaultOptionPrepareData)
	case "start date outside billing schedule":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		req, err = s.invalidCaseStartDateOutsideBillingSchedule(ctx, defaultOptionPrepareData)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesErrMessageForCreateRecurringMaterial(ctx context.Context, expectedErrMessage, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	req := stepState.Request.(*pb.CreateOrderRequest)
	switch testcase {
	case "incorrect ratio applied with prorating":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].ProductId, 250, 375)
	case "incorrect ratio applied with disabled prorating":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].ProductId, 250, 500)
	case "incorrect tax percent applied":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].ProductId, 10, 20)
	case "incorrect tax amount applied without discount and prorating":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, 50, 83.333336)
	case "incorrect tax amount applied with discount and prorating":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, 81.666664, 40)
	case "incorrect final price missing discount billed at order":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].ProductId, 500, 490)
	case "incorrect final price missing discount upcoming billing":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].ProductId, 500, 490)

	default:
	}
	if !strings.Contains(stt.Message(), expectedErrMessage) {
		return ctx, fmt.Errorf("expecting %s, got %s error message ", expectedErrMessage, stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderRecurringMaterialSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	err := s.checkCreatedOrderDetailsAndActionLogs(ctx, pb.OrderType_ORDER_TYPE_NEW)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.validateCreatedOrderItemsAndBillItemsForRecurringProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
