package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareForUpdateOneTimeFeeWithStatusInvoiced(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		feeIDs      []string
		req         *pb.CreateOrderRequest
		resp        *pb.CreateOrderResponse
		billItem    *entities.BillItem
		err         error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertStudent:         true,
		insertMaterial:        false,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertFee:             true,
		insertProductDiscount: true,
	}
	req = &pb.CreateOrderRequest{}
	taxID,
		discountIDs,
		locationID,
		feeIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeFee)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(feeIDs))
	billingItems := make([]*pb.BillingItem, 0, len(feeIDs))

	orderItems = append(orderItems, &pb.OrderItem{ProductId: feeIDs[0]})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: feeIDs[0],
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
	billItem, err = s.getBillItemByOrderIDAndProductID(ctx, resp.OrderId, feeIDs[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.OrderComment = "test update order fee one time"
	req.OrderItems = []*pb.OrderItem{{
		ProductId: feeIDs[0],
		StudentProductId: &wrapperspb.StringValue{
			Value: billItem.StudentProductID.String,
		},
		DiscountId: &wrapperspb.StringValue{
			Value: discountIDs[0],
		},
	}}
	req.BillingItems = []*pb.BillingItem{
		{
			ProductId: feeIDs[0],
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
	err = s.updateBillItemStatus(ctx, resp.OrderId, feeIDs[0], pb.BillingStatus_BILLING_STATUS_INVOICED)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrderOneTimeFeeWithStatusInvoicedSuccess(ctx context.Context) (context.Context, error) {
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

func (s *suite) prepareForCancelOneTimeFeeWithStatusOrdered(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		feeIDs      []string
		req         *pb.CreateOrderRequest
		resp        *pb.CreateOrderResponse
		billItem    *entities.BillItem
		err         error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertStudent:         true,
		insertMaterial:        false,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertFee:             true,
		insertProductDiscount: true,
	}
	req = &pb.CreateOrderRequest{}
	taxID,
		discountIDs,
		locationID,
		feeIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeFee)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order fee one time"

	orderItems := make([]*pb.OrderItem, 0, len(feeIDs))
	billingItems := make([]*pb.BillingItem, 0, len(feeIDs))

	orderItems = append(orderItems, &pb.OrderItem{ProductId: feeIDs[0]})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: feeIDs[0],
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
	billItem, err = s.getBillItemByOrderIDAndProductID(ctx, resp.OrderId, feeIDs[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.OrderComment = "test update order material one time"
	req.OrderItems = []*pb.OrderItem{{
		ProductId: feeIDs[0],
		StudentProductId: &wrapperspb.StringValue{
			Value: billItem.StudentProductID.String,
		},
		DiscountId: &wrapperspb.StringValue{
			Value: discountIDs[0],
		},
	}}
	req.BillingItems = []*pb.BillingItem{
		{
			ProductId: feeIDs[0],
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

func (s *suite) cancelOrderOneTimeFeeWithStatusOrderedSuccess(ctx context.Context) (context.Context, error) {
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

func (s *suite) prepareForUpdateOneTimeFeeWithStatusInvoicedAndOutVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		feeIDs      []string
		req         *pb.CreateOrderRequest
		resp        *pb.CreateOrderResponse
		billItem    *entities.BillItem
		err         error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertStudent:         true,
		insertMaterial:        false,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertFee:             true,
		insertProductDiscount: true,
	}
	req = &pb.CreateOrderRequest{}
	taxID,
		discountIDs,
		locationID,
		feeIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeFee)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(feeIDs))
	billingItems := make([]*pb.BillingItem, 0, len(feeIDs))

	orderItems = append(orderItems, &pb.OrderItem{
		ProductId: feeIDs[0],
	})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: feeIDs[0],
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
	billItem, err = s.getBillItemByOrderIDAndProductID(ctx, resp.OrderId, feeIDs[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.OrderComment = "test update order fee one time"
	req.OrderItems = []*pb.OrderItem{{
		ProductId: feeIDs[0],
		StudentProductId: &wrapperspb.StringValue{
			Value: billItem.StudentProductID.String,
		},
		DiscountId: &wrapperspb.StringValue{
			Value: discountIDs[0],
		},
		StudentProductVersionNumber: int32(5),
	}}
	req.BillingItems = []*pb.BillingItem{
		{
			ProductId: feeIDs[0],
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
	err = s.updateBillItemStatus(ctx, resp.OrderId, feeIDs[0], pb.BillingStatus_BILLING_STATUS_INVOICED)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrderOutOfVersionUnsuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr.Error() != status.Error(codes.FailedPrecondition, constant.OptimisticLockingEntityVersionMismatched).Error() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("valid version number")
	}

	return StepStateToContext(ctx, stepState), nil
}
