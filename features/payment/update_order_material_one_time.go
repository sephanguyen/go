package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) getBillItemByOrderIDAndProductID(ctx context.Context, orderID string, productID string) (*entities.BillItem, error) {
	billItem := &entities.BillItem{}
	billItemFieldNames, billItemFieldValues := billItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1 AND product_id = $2
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		billItem.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, orderID, productID)
	err := row.Scan(billItemFieldValues...)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	return billItem, nil
}

func (s *suite) getBillItemByStudentIDs(ctx context.Context, studentProductIDs []string) ([]entities.BillItem, error) {
	var billItems []entities.BillItem
	billItem := &entities.BillItem{}
	billItemFieldNames, billItemFieldValues := billItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = ANY($1)
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		billItem.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, studentProductIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(billItemFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, *billItem)
	}
	return billItems, nil
}

func (s *suite) prepareForUpdateOneTimeMaterialWithStatusInvoiced(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		materialIDs []string
		req         *pb.CreateOrderRequest
		resp        *pb.CreateOrderResponse
		billItem    *entities.BillItem
		err         error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertStudent:         true,
		insertMaterial:        true,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertFee:             false,
		insertProductDiscount: true,
	}
	req = &pb.CreateOrderRequest{}
	taxID,
		discountIDs,
		locationID,
		materialIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))

	orderItems = append(orderItems, &pb.OrderItem{ProductId: materialIDs[0]})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: materialIDs[0],
		Price:     PriceOrder,
		Quantity:  &wrapperspb.Int32Value{Value: 1},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     s.calculateTaxAmount(PriceOrder, 0, 20),
		},
		FinalPrice: PriceOrder,
	})

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState = StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	resp, err = pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		resp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), req)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	billItem, err = s.getBillItemByOrderIDAndProductID(ctx, resp.OrderId, materialIDs[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.OrderComment = "test update order material one time"
	req.OrderItems = []*pb.OrderItem{{
		ProductId: materialIDs[0],
		StudentProductId: &wrapperspb.StringValue{
			Value: billItem.StudentProductID.String,
		},
		DiscountId: &wrapperspb.StringValue{
			Value: discountIDs[0],
		},
	}}
	req.BillingItems = []*pb.BillingItem{
		{
			ProductId: materialIDs[0],
			Price:     PriceOrder,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[0],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 20,
				DiscountAmount:      20,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     80,
			},
			FinalPrice: PriceOrder - 20,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItem.StudentProductID.String,
			},
			AdjustmentPrice: &wrapperspb.FloatValue{Value: -20},
		},
	}
	req.OrderType = pb.OrderType_ORDER_TYPE_UPDATE
	stepState.Request = req
	err = s.updateBillItemStatus(ctx, resp.OrderId, materialIDs[0], pb.BillingStatus_BILLING_STATUS_INVOICED)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateBillItemStatus(ctx context.Context, orderID string, productID string, billingStatus pb.BillingStatus) error {
	stmt :=
		`
		UPDATE public.bill_item
		SET billing_status= $1
		WHERE order_id = $2 AND product_id = $3;
		`
	_, err := s.FatimaDBTrace.Exec(ctx, stmt, billingStatus.String(), orderID, productID)
	return err
}

func (s *suite) updateOrderOneTimeMaterialWithStatusInvoicedSuccess(ctx context.Context) (context.Context, error) {
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
	foundOrderItem := countOrderItem(orderItems, req.OrderItems)
	if foundOrderItem < len(req.OrderItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("create miss orderItem")
	}
	studentProductIDs := make([]string, 0, len(req.OrderItems))
	for _, item := range req.OrderItems {
		studentProductIDs = append(studentProductIDs, item.StudentProductId.Value)
	}
	billItems, err := s.getBillItemByStudentIDs(ctx, studentProductIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	foundBillItem := countAffectedBillItem(billItems, req.BillingItems)
	if foundBillItem < len(req.BillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("create miss billItem")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareForCancelOneTimeMaterialWithStatusOrdered(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		materialIDs []string
		req         *pb.CreateOrderRequest
		resp        *pb.CreateOrderResponse
		billItem    *entities.BillItem
		err         error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertStudent:         true,
		insertMaterial:        true,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertFee:             false,
		insertProductDiscount: true,
	}
	req = &pb.CreateOrderRequest{}
	taxID,
		discountIDs,
		locationID,
		materialIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))

	orderItems = append(orderItems, &pb.OrderItem{ProductId: materialIDs[0]})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: materialIDs[0],
		Price:     PriceOrder,
		Quantity:  &wrapperspb.Int32Value{Value: 1},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     s.calculateTaxAmount(PriceOrder, 0, 20),
		},
		FinalPrice: PriceOrder,
	})

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState = StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	resp, err = pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		resp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), req)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	billItem, err = s.getBillItemByOrderIDAndProductID(ctx, resp.OrderId, materialIDs[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.OrderComment = "test update order material one time"
	req.OrderItems = []*pb.OrderItem{{
		ProductId: materialIDs[0],
		StudentProductId: &wrapperspb.StringValue{
			Value: billItem.StudentProductID.String,
		},
		DiscountId: &wrapperspb.StringValue{
			Value: discountIDs[0],
		},
	}}
	req.BillingItems = []*pb.BillingItem{
		{
			ProductId: materialIDs[0],
			Price:     PriceOrder,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[0],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 20,
				DiscountAmount:      20,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     80,
			},
			FinalPrice: PriceOrder - 20,
			StudentProductId: &wrapperspb.StringValue{
				Value: billItem.StudentProductID.String,
			},
			IsCancelBillItem: wrapperspb.Bool(true),
		},
	}
	req.OrderType = pb.OrderType_ORDER_TYPE_UPDATE
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) cancelOrderOneTimeMaterialWithStatusOrderedSuccess(ctx context.Context) (context.Context, error) {
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
	foundOrderItem := countOrderItem(orderItems, req.OrderItems)
	if foundOrderItem < len(req.OrderItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("create miss orderItem")
	}
	studentProductIDs := make([]string, 0, len(req.OrderItems))
	for _, item := range req.OrderItems {
		studentProductIDs = append(studentProductIDs, item.StudentProductId.Value)
	}
	billItems, err := s.getBillItemByStudentIDs(ctx, studentProductIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	foundBillItem := countCancelBillItem(billItems, req.BillingItems)
	if foundBillItem < len(req.BillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("create miss billItem")
	}

	return StepStateToContext(ctx, stepState), nil
}

func countAffectedBillItem(billItems []entities.BillItem, billingItems []*pb.BillingItem) int {
	foundBillItem := 0
	for _, item := range billingItems {
		for _, dbItem := range billItems {
			if item.ProductId == dbItem.ProductID.String &&
				dbItem.TaxCategory.String == item.TaxItem.TaxCategory.String() &&
				float32(dbItem.TaxPercentage.Int) == item.TaxItem.TaxPercentage &&
				dbItem.TaxID.String == item.TaxItem.TaxId &&
				IsEqualNumericAndFloat32(dbItem.TaxAmount, item.TaxItem.TaxAmount) &&
				float32(dbItem.ProductPricing.Int) == item.Price &&
				dbItem.BillSchedulePeriodID.Status == pgtype.Null &&
				dbItem.BillType.String == pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String() &&
				dbItem.BillFrom.Status == pgtype.Null &&
				dbItem.BillTo.Status == pgtype.Null &&
				IsEqualNumericAndFloat32(dbItem.AdjustmentPrice, item.AdjustmentPrice.Value) {
				foundBillItem++
			}
		}
	}
	return foundBillItem
}

func countCancelBillItem(billItems []entities.BillItem, billingItems []*pb.BillingItem) int {
	foundBillItem := 0
	for _, item := range billingItems {
		for _, dbItem := range billItems {
			if item.ProductId == dbItem.ProductID.String &&
				dbItem.BillStatus.String == pb.BillingStatus_BILLING_STATUS_BILLED.String() &&
				dbItem.BillType.String == pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String() {
				foundBillItem++
			}
		}
	}
	return foundBillItem
}
