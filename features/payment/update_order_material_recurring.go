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

func (s *suite) prepareDataForCreateOrderRecurringMaterial(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
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
	}

	req := pb.CreateOrderRequest{}
	var err error

	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, defaultOptionPrepareData, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var orderItems []*pb.OrderItem
	var billedAtOrderItems []*pb.BillingItem
	var upcomingBillingItems []*pb.BillingItem

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:  data.ProductIDs[0],
			DiscountId: &wrapperspb.StringValue{Value: data.DiscountIDs[1]},
			StartDate:  &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 18).Unix()},
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

func (s *suite) prepareDataForUpdateOrderRecurringMaterial(ctx context.Context) (context.Context, error) {
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
			EffectiveDate:    &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               oldReq.OrderItems[0].ProductId,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: oldReq.BillingItems[2].BillingSchedulePeriodId.Value},
			Price:                   PriceOrder - 50,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldReq.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder-50, 20),
			},
			FinalPrice:       PriceOrder - 50,
			AdjustmentPrice:  wrapperspb.Float(7.5),
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               oldReq.OrderItems[0].ProductId,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: oldReq.UpcomingBillingItems[0].BillingSchedulePeriodId.Value},
			Price:                   PriceOrder + 150,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldReq.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder+150, 20),
			},
			FinalPrice:       PriceOrder + 150,
			DiscountItem:     nil,
			AdjustmentPrice:  wrapperspb.Float(10),
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
		},
	)
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

func (s *suite) prepareDataForCancelOrderRecurringMaterial(ctx context.Context) (context.Context, error) {
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
			DiscountId:       oldReq.OrderItems[0].DiscountId,
			EffectiveDate:    &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 2).Unix()},
			CancellationDate: &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 15).Unix()},
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
		},
	)
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               oldReq.OrderItems[0].ProductId,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: oldReq.BillingItems[2].BillingSchedulePeriodId.Value},
			Price:                   PriceOrder + 100,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldReq.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder+100, 20),
			},
			FinalPrice:       PriceOrder + 90,
			AdjustmentPrice:  wrapperspb.Float(-(PriceOrder - 57.5)),
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
			IsCancelBillItem: wrapperspb.Bool(true),
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          oldReq.OrderItems[0].DiscountId.Value,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               oldReq.OrderItems[0].ProductId,
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: oldReq.UpcomingBillingItems[0].BillingSchedulePeriodId.Value},
			Price:                   PriceOrder + 150,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         oldReq.BillingItems[2].TaxItem.TaxId,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder+150, 20),
			},
			FinalPrice: PriceOrder + 140,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          oldReq.OrderItems[0].DiscountId.Value,
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
			AdjustmentPrice:  wrapperspb.Float(-640),
			StudentProductId: wrapperspb.String(oldStudentProduct.StudentProductID.String),
			IsCancelBillItem: wrapperspb.Bool(true),
		},
	)
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

func (s *suite) getStudentProductBaseOnProductID(ctx context.Context, productID string) (*entities.StudentProduct, error) {
	studentProduct := &entities.StudentProduct{}
	studentProductFieldNames, studentProductFieldValues := studentProduct.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProduct.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, productID)
	err := row.Scan(studentProductFieldValues...)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	return studentProduct, nil
}

func (s *suite) getListStudentProductBaseOnProductID(ctx context.Context, productID string) ([]*entities.StudentProduct, error) {
	studentProducts := make([]*entities.StudentProduct, 0, 2)
	studentProduct := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProduct.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1
		ORDER BY start_date
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProduct.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, productID)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	for rows.Next() {
		tmpStudentProduct := &entities.StudentProduct{}
		_, tmpStudentProductFieldValues := tmpStudentProduct.FieldMap()
		err = rows.Scan(tmpStudentProductFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		studentProducts = append(studentProducts, tmpStudentProduct)
	}
	return studentProducts, nil
}

func (s *suite) getListBillItemBaseOnStudentProductID(ctx context.Context, studentProductID string) ([]*entities.BillItem, error) {
	billItems := make([]*entities.BillItem, 0, 2)
	billItem := &entities.BillItem{}
	billItemFieldNames, _ := billItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = $1
		ORDER BY billing_from
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		billItem.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, studentProductID)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	for rows.Next() {
		tmpBillItem := &entities.BillItem{}
		_, tmpBillItemFieldValues := tmpBillItem.FieldMap()
		err = rows.Scan(tmpBillItemFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, tmpBillItem)
	}
	return billItems, nil
}

func (s *suite) updateOrderForRecurringMaterialSuccess(ctx context.Context) (context.Context, error) {
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
