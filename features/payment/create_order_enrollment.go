package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) orderEnrollmentIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
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

func (s *suite) prepareDataForCreateOrderEnrollmentWith(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                  true,
		InsertDiscount:             true,
		InsertStudent:              true,
		InsertEnrolledStudent:      false,
		InsertPotentialStudent:     true,
		InsertProductPrice:         true,
		InsertEnrolledProductPrice: true,
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
	case "order with single billed at order item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		req, err = s.validCaseBilledAtOrderItemsSingleItemExpectedForEnrollment(ctx, defaultOptionPrepareData)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesErrorMessageForCreateOrderEnrollmentWith(ctx context.Context, expectedErrMessage, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if !strings.Contains(stt.Message(), expectedErrMessage) {
		return ctx, fmt.Errorf("expecting %s, got %s error message ", expectedErrMessage, stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestForCreateOrderEnrollmentWith(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                  true,
		InsertDiscount:             true,
		InsertStudent:              true,
		InsertEnrolledStudent:      true,
		InsertProductPrice:         true,
		InsertEnrolledProductPrice: true,
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
	case "student is already enrolled":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		req, err = s.studentIsAlreadyEnrolled(ctx, defaultOptionPrepareData)
	case "student is LOA in location":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		req, err = s.studentIsLOAInLocation(ctx, defaultOptionPrepareData)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentIsAlreadyEnrolled(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			Price:                   EnrolledProductPrice,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: EnrolledProductPrice,
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   EnrolledProductPrice,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: EnrolledProductPrice,
		},
	)

	req.StudentId = data.EnrolledUserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order enrollment"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_ENROLLMENT

	return
}

func (s *suite) studentIsLOAInLocation(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			Price:                   EnrolledProductPrice,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: EnrolledProductPrice,
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
				TaxAmount:     getInclusivePercentTax(PriceOrder, 20),
			},
			FinalPrice: EnrolledProductPrice,
		},
	)

	req.StudentId = data.EnrolledUserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order enrollment"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_ENROLLMENT

	err = mockdata.UpdateStudentStatus(ctx, s.FatimaDBTrace, data.EnrolledUserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_LOA", time.Now(), time.Now().AddDate(1, 0, 0))
	return
}

func (s *suite) validCaseBilledAtOrderItemsSingleItemExpectedForEnrollment(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			Price:                   EnrolledProductPrice,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(EnrolledProductPrice, 20),
			},
			FinalPrice: EnrolledProductPrice,
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   EnrolledProductPrice,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(EnrolledProductPrice, 20),
			},
			FinalPrice: EnrolledProductPrice,
		},
	)

	req.StudentId = data.PotentialUserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order enrollment"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_ENROLLMENT

	return
}
