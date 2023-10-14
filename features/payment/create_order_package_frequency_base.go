package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForCreateOrderFrequencyBasePackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            false,
		InsertProductLocation:         true,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertFee:                     false,
		InsertMaterial:                false,
		InsertBillingSchedule:         true,
		InsertBillingScheduleArchived: false,
		IsTaxExclusive:                false,
		InsertDiscountNotAvailable:    false,
		InsertProductOutOfTime:        false,
		InsertPackageCourses:          true,
		InsertProductDiscount:         true,
		BillingScheduleStartDate:      time.Now(),
	}

	var (
		req pb.CreateOrderRequest
		err error
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	req, err = s.validCaseWithDiscountAndProrating(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderFrequencyBasePackageSuccess(ctx context.Context) (context.Context, error) {
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

func (s *suite) prepareDataForCreateOrderFrequencyBasePackageWithInvalidRequest(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            false,
		InsertProductLocation:         true,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertFee:                     false,
		InsertMaterial:                false,
		InsertBillingSchedule:         true,
		InsertBillingScheduleArchived: false,
		IsTaxExclusive:                false,
		InsertDiscountNotAvailable:    false,
		InsertProductOutOfTime:        false,
		InsertPackageCourses:          true,
		InsertProductDiscount:         true,
		BillingScheduleStartDate:      time.Now(),
	}

	req := pb.CreateOrderRequest{}
	var err error

	switch testcase {
	case "bill item no course":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseBillItemNoCourse(ctx, defaultOptionPrepareData)
	case "course-slot is nil":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseSlotNil(ctx, defaultOptionPrepareData)
	case "courses in order item does not match in courses in bill item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseInOrderItemDoesNotMatchBillItem(ctx, defaultOptionPrepareData)
	case "duplicate course in order item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseDuplicateCourseInOrderItem(ctx, defaultOptionPrepareData)
	case "duplicate course in bill item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseDuplicateCourseInBillItem(ctx, defaultOptionPrepareData)
	case "course info in order item and bill item does not match":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseInfoDoesNotMatch(ctx, defaultOptionPrepareData)
	case "course slot in order item and bill item does not match":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseCourseSlotDoesNotMatch(ctx, defaultOptionPrepareData)
	case "missing mandatory course in bill item":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		defaultOptionPrepareData.ArePackageCoursesMandatory = true
		req, err = s.invalidCaseMissingMandatoryCourse(ctx, defaultOptionPrepareData)
	case "slot in order greater than max slot per course in DB":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseSlotsGreaterThanMaxSlot(ctx, defaultOptionPrepareData)
	case "quantity in bill item empty":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		req, err = s.invalidCaseQuantityBillItemEmpty(ctx, defaultOptionPrepareData)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesErrMessageForCreateOrderFrequencyBasePackage(ctx context.Context, expectedErrMessage, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	req := stepState.Request.(*pb.CreateOrderRequest)
	switch testcase {
	case "course-slot is nil":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[0].CourseId)
	case "duplicate course in order item":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[0].CourseId)
	case "duplicate course in bill item":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.BillingItems[0].CourseItems[0].CourseId)
	case "course info in order item and bill item does not match":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.OrderItems[0].CourseItems[0].CourseId)
	case "course slot in order item and bill item does not match":
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

func (s *suite) getStudentPackageByOrderByStudentID(ctx context.Context, studentID string) (*entities.StudentPackages, error) {
	studentPackageByOrder := &entities.StudentPackages{}
	studentPackageByOrderFieldNames, studentPackageByOrderFieldValues := studentPackageByOrder.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentPackageByOrderFieldNames, ","),
		studentPackageByOrder.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, studentID)
	err := row.Scan(studentPackageByOrderFieldValues...)
	if err != nil {
		return studentPackageByOrder, err
	}

	return studentPackageByOrder, nil
}

func (s *suite) getStudentPackagesByStudentID(ctx context.Context, studentID string) (*entities.StudentPackages, error) {
	studentPackage := &entities.StudentPackages{}
	studentPackageFieldNames, studentPackageFieldValues := studentPackage.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentPackageFieldNames, ","),
		studentPackage.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, studentID)
	err := row.Scan(studentPackageFieldValues...)
	if err != nil {
		return studentPackage, err
	}

	return studentPackage, nil
}

func (s *suite) validCaseWithDiscountAndProrating(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[1]},
			StartDate:  &timestamppb.Timestamp{Seconds: options.BillingScheduleStartDate.AddDate(0, 0, 18).Unix()},
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getProratedPrice(getFrequencyBasePrice(100, 3), 1, 2),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getProratedPrice(getFrequencyBasePrice(100, 3), 1, 2)-5, 20),
			},
			// 1/2 prorating applied
			FinalPrice: getProratedPrice(getFrequencyBasePrice(100, 3), 1, 2) - 5,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      5,
			},
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3)-10, 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3) - 10,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3)-10, 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3) - 10,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3)-10, 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3) - 10,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseBillItemNoCourse(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice:  getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseSlotNil(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseInOrderItemDoesNotMatchBillItem(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseDuplicateCourseInOrderItem(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseDuplicateCourseInBillItem(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseInfoDoesNotMatch(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseCourseSlotDoesNotMatch(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(2),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseMissingMandatoryCourse(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[2].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[2].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseSlotsGreaterThanMaxSlot(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(10),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(10),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(10),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(10),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(10),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}

func (s *suite) invalidCaseQuantityBillItemEmpty(ctx context.Context, options mockdata.OptionToPrepareDataForCreateOrderRecurringProduct) (req pb.CreateOrderRequest, err error) {
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
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   getFrequencyBasePrice(100, 3),
			Quantity:                &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(getFrequencyBasePrice(100, 3), 20),
			},
			FinalPrice: getFrequencyBasePrice(100, 3),
			CourseItems: []*pb.CourseItem{
				{
					CourseId:   data.PackageCourses[0].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[0].CourseID.String),
					Slot:       wrapperspb.Int32(1),
				},
				{
					CourseId:   data.PackageCourses[1].CourseID.String,
					CourseName: fmt.Sprintf("course-%s", data.PackageCourses[1].CourseID.String),
					Slot:       wrapperspb.Int32(2),
				},
			},
		},
	)

	req.StudentId = data.UserID
	req.LocationId = data.LocationID
	req.OrderComment = "test create order frequency-base package"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return
}
