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

func (s *suite) prepareDataForCreateOrderScheduleBasePackage(ctx context.Context) (context.Context, error) {
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
		InsertPackageCourseScheduleBased: true,
		InsertProductDiscount:            true,
		BillingScheduleStartDate:         time.Now(),
	}

	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	req, err := s.validCaseHappyCaseScheduleBasePackage(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderScheduleBasePackageSuccess(ctx context.Context) (context.Context, error) {
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

func (s *suite) validCaseHappyCaseScheduleBasePackage(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
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
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
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
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
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
	req.OrderComment = "test create order recurring fee"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) receivesErrorMessageForCreateOrderSchedulebasePackageWith(ctx context.Context, expectedErrMessage, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	req := stepState.Request.(*pb.CreateOrderRequest)
	switch testcase {
	case "course-weight is nil":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[0].CourseId)
	case "duplicate course in order item":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[0].CourseId)
	case "duplicate course in bill item":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.BillingItems[0].CourseItems[0].CourseId)
	case "course info in order item and bill item does not match":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[0].CourseId)
	case "course weight in order item and bill item does not match":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[0].CourseId)
	case "slot in order greater than max slot per course in DB":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[1].CourseId)
	default:
	}
	if !strings.Contains(stt.Message(), expectedErrMessage) {
		return ctx, fmt.Errorf("expecting %s, got %s error message ", expectedErrMessage, stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestForCreateOrderSchedulebasePackageWith(ctx context.Context, testcase string) (context.Context, error) {
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
		InsertPackageCourseScheduleBased: true,
		InsertProductDiscount:            true,
		BillingScheduleStartDate:         time.Now(),
	}

	req := pb.CreateOrderRequest{}
	var err error

	switch testcase {
	case "bill item no course":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseBillItemNoCourseScheduleBase(ctx, defaultOptionPrepareData)
	case "course-weight is nil":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseWeightNil(ctx, defaultOptionPrepareData)
	case "courses in order item does not match in courses in bill item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseInOrderItemDoesNotMatchBillItemScheduleBase(ctx, defaultOptionPrepareData)
	case "duplicate course in order item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseDuplicateCourseInOrderItemScheduleBase(ctx, defaultOptionPrepareData)
	case "duplicate course in bill item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseDuplicateCourseInBillItemScheduleBase(ctx, defaultOptionPrepareData)
	case "course info in order item and bill item does not match":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseInfoDoesNotMatchScheduleBase(ctx, defaultOptionPrepareData)
	case "course weight in order item and bill item does not match":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseWeightDoesNotMatch(ctx, defaultOptionPrepareData)
	case "missing mandatory course in bill item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		defaultOptionPrepareData.ArePackageCoursesMandatory = true
		req, err = s.invalidCaseMissingMandatoryCourseScheduleBase(ctx, defaultOptionPrepareData)
	case "quantity in bill item empty":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseQuantityBillItemEmptyScheduleBase(ctx, defaultOptionPrepareData)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invalidCaseBillItemNoCourseScheduleBase(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice:  getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseWeightNil(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseInOrderItemDoesNotMatchBillItemScheduleBase(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseDuplicateCourseInOrderItemScheduleBase(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseDuplicateCourseInBillItemScheduleBase(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseInfoDoesNotMatchScheduleBase(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseWeightDoesNotMatch(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseMissingMandatoryCourseScheduleBase(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseQuantityBillItemEmptyScheduleBase(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getScheduleBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getScheduleBasePrice(100, 3), 20),
			},
			FinalPrice: getScheduleBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Weight:     wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Weight:     wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order schedule-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}
