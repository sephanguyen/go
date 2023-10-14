package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForCreateOrderRecurringPackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                        true,
		InsertDiscount:                   false,
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

	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, defaultOptionPrepareData, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var (
		orderItems           []*pb.OrderItem
		billedAtOrderItems   []*pb.BillingItem
		upcomingBillingItems []*pb.BillingItem
	)

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId: data.ProductIDs[0],
			StartDate: &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
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

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForUpdateOrderRecurringPackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	oldReq := stepState.Request.(*pb.CreateOrderRequest)
	oldStudentProduct, err := s.getStudentProductBaseOnProductID(ctx, oldReq.OrderItems[0].ProductId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertDiscount: true,
	}
	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, defaultOptionPrepareData, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := pb.CreateOrderRequest{}
	var orderItems []*pb.OrderItem
	var billedAtOrderItems []*pb.BillingItem
	var upcomingBillingItems []*pb.BillingItem
	err = mockdata.InsertProductDiscount(ctx, s.FatimaDBTrace, []string{oldReq.OrderItems[0].ProductId}, []string{data.DiscountIDs[1]})
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:        oldReq.OrderItems[0].ProductId,
			DiscountId:       wrapperspb.String(data.DiscountIDs[1]),
			EffectiveDate:    &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
			CourseItems:      oldReq.OrderItems[0].CourseItems,
		},
	)
	billedAtOrderItem := oldReq.BillingItems[2]
	billedAtOrderItem.DiscountItem = &pb.DiscountBillItem{
		DiscountId:          data.DiscountIDs[1],
		DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
		DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
		DiscountAmountValue: 10,
		DiscountAmount:      7.5,
	}
	billedAtOrderItem.StudentProductId = wrapperspb.String(oldStudentProduct.StudentProductID.String)
	billedAtOrderItem.AdjustmentPrice = wrapperspb.Float(-7.5)
	billedAtOrderItem.Price = 150
	billedAtOrderItem.FinalPrice = 142.5
	billedAtOrderItem.TaxItem.TaxAmount = 23.75
	billedAtOrderItems = append(billedAtOrderItems, billedAtOrderItem)
	upcomingBillingItem := oldReq.UpcomingBillingItems[0]
	upcomingBillingItem.DiscountItem = &pb.DiscountBillItem{
		DiscountId:          data.DiscountIDs[1],
		DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
		DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
		DiscountAmountValue: 10,
		DiscountAmount:      10,
	}
	upcomingBillingItem.StudentProductId = wrapperspb.String(oldStudentProduct.StudentProductID.String)
	upcomingBillingItem.AdjustmentPrice = wrapperspb.Float(-10)
	upcomingBillingItem.FinalPrice = 190
	upcomingBillingItem.TaxItem.TaxAmount = 31.666666
	upcomingBillingItems = append(upcomingBillingItems, upcomingBillingItem)
	req.StudentId = oldReq.StudentId
	req.LocationId = oldReq.LocationId
	req.OrderComment = "test create order recurring fee"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_UPDATE

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCancelOrderRecurringPackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	oldReq := stepState.Request.(*pb.CreateOrderRequest)
	oldStudentProduct, err := s.getStudentProductBaseOnProductID(ctx, oldReq.OrderItems[0].ProductId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := pb.CreateOrderRequest{}
	var orderItems []*pb.OrderItem
	var billedAtOrderItems []*pb.BillingItem
	var upcomingBillingItems []*pb.BillingItem

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:        oldReq.OrderItems[0].ProductId,
			DiscountId:       nil,
			EffectiveDate:    &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 2).Unix()},
			CancellationDate: &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 15).Unix()},
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
			CourseItems:      oldReq.OrderItems[0].CourseItems,
		},
	)
	billedAtOrderItem := oldReq.BillingItems[2]
	billedAtOrderItem.StudentProductId = wrapperspb.String(oldStudentProduct.StudentProductID.String)
	billedAtOrderItem.AdjustmentPrice = wrapperspb.Float(-150)
	billedAtOrderItem.IsCancelBillItem = wrapperspb.Bool(true)
	billedAtOrderItems = append(billedAtOrderItems, billedAtOrderItem)
	upcomingBillingItem := oldReq.UpcomingBillingItems[0]
	upcomingBillingItem.StudentProductId = wrapperspb.String(oldStudentProduct.StudentProductID.String)
	upcomingBillingItem.AdjustmentPrice = wrapperspb.Float(-200)
	upcomingBillingItem.IsCancelBillItem = wrapperspb.Bool(true)
	upcomingBillingItems = append(upcomingBillingItems, upcomingBillingItem)
	req.StudentId = oldReq.StudentId
	req.LocationId = oldReq.LocationId
	req.OrderComment = "test create order recurring fee"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_UPDATE

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrderForRecurringPackageSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.CreateOrderRequest)
	productID := req.OrderItems[0].ProductId
	studentProducts, err := s.getListStudentProductBaseOnProductID(ctx, productID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(studentProducts) != 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err when don't create 2 student product")
	}

	if studentProducts[0].StudentProductLabel.String != pb.StudentProductLabel_UPDATE_SCHEDULED.String() ||
		studentProducts[1].StudentProductLabel.String != pb.StudentProductLabel_CREATED.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong create student product")
	}
	billItems, err := s.getListBillItemBaseOnStudentProductID(ctx, studentProducts[1].StudentProductID.String)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(billItems) != 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err when don't create 2 bill item")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForUpdateOrderRecurringPackageOutOfVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	oldReq := stepState.Request.(*pb.CreateOrderRequest)
	oldStudentProduct, err := s.getStudentProductBaseOnProductID(ctx, oldReq.OrderItems[0].ProductId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertDiscount: true,
	}
	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, defaultOptionPrepareData, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := pb.CreateOrderRequest{}
	var orderItems []*pb.OrderItem
	var billedAtOrderItems []*pb.BillingItem
	var upcomingBillingItems []*pb.BillingItem
	err = mockdata.InsertProductDiscount(ctx, s.FatimaDBTrace, []string{oldReq.OrderItems[0].ProductId}, []string{data.DiscountIDs[1]})
	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:                   oldReq.OrderItems[0].ProductId,
			DiscountId:                  wrapperspb.String(data.DiscountIDs[1]),
			EffectiveDate:               &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
			StudentProductId:            wrapperspb.String(oldStudentProduct.StudentProductID.String),
			CourseItems:                 oldReq.OrderItems[0].CourseItems,
			StudentProductVersionNumber: int32(5),
		},
	)
	billedAtOrderItem := oldReq.BillingItems[2]
	billedAtOrderItem.DiscountItem = &pb.DiscountBillItem{
		DiscountId:          data.DiscountIDs[1],
		DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
		DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
		DiscountAmountValue: 10,
		DiscountAmount:      7.5,
	}
	billedAtOrderItem.StudentProductId = wrapperspb.String(oldStudentProduct.StudentProductID.String)
	billedAtOrderItem.AdjustmentPrice = wrapperspb.Float(-7.5)
	billedAtOrderItem.Price = 150
	billedAtOrderItem.FinalPrice = 142.5
	billedAtOrderItem.TaxItem.TaxAmount = 23.75
	billedAtOrderItems = append(billedAtOrderItems, billedAtOrderItem)
	upcomingBillingItem := oldReq.UpcomingBillingItems[0]
	upcomingBillingItem.DiscountItem = &pb.DiscountBillItem{
		DiscountId:          data.DiscountIDs[1],
		DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
		DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
		DiscountAmountValue: 10,
		DiscountAmount:      10,
	}
	upcomingBillingItem.StudentProductId = wrapperspb.String(oldStudentProduct.StudentProductID.String)
	upcomingBillingItem.AdjustmentPrice = wrapperspb.Float(-10)
	upcomingBillingItem.FinalPrice = 190
	upcomingBillingItem.TaxItem.TaxAmount = 31.666666
	upcomingBillingItems = append(upcomingBillingItems, upcomingBillingItem)
	req.StudentId = oldReq.StudentId
	req.LocationId = oldReq.LocationId
	req.OrderComment = "test create order recurring fee"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_UPDATE

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}
