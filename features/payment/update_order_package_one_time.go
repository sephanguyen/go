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

func (s *suite) prepareDataForUpdateOrderOneTimePackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		discountIDs []string
		err         error
		productIDs  []string
		req         *pb.CreateOrderRequest
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	discountIDs, err = mockdata.InsertOneDiscountAmount(ctx, s.FatimaDBTrace, "discount update order")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req = stepState.Request.(*pb.CreateOrderRequest)
	for index := range req.OrderItems {
		productIDs = append(productIDs, req.BillingItems[index].ProductId)
		req.OrderItems[index].DiscountId = wrapperspb.String(discountIDs[0])
		req.OrderType = pb.OrderType_ORDER_TYPE_UPDATE
		req.BillingItems[index].DiscountItem = &pb.DiscountBillItem{
			DiscountId:          discountIDs[0],
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
			DiscountAmountValue: 20,
			DiscountAmount:      20,
		}
		req.BillingItems[index].FinalPrice = PriceOrder - 20
		req.BillingItems[index].AdjustmentPrice = wrapperspb.Float(-20)
		req.BillingItems[index].TaxItem.TaxAmount = s.calculateTaxAmount(PriceOrder, 20, 20)
		studentProduct, err := s.getStudentProductByProductID(ctx, req.BillingItems[index].ProductId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.BillingItems[index].StudentProductId = wrapperspb.String(studentProduct.StudentProductID.String)
		req.OrderItems[index].StudentProductId = wrapperspb.String(studentProduct.StudentProductID.String)
	}
	err = mockdata.InsertProductDiscount(ctx, s.FatimaDBTrace, productIDs, discountIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCancelOrderOneTimePackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		err error
		req *pb.CreateOrderRequest
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req = stepState.Request.(*pb.CreateOrderRequest)
	startDate := timestamppb.New(time.Now().AddDate(0, 1, 0))
	for index := range req.OrderItems {
		req.OrderType = pb.OrderType_ORDER_TYPE_UPDATE
		req.BillingItems[index].FinalPrice = PriceOrder
		req.BillingItems[index].AdjustmentPrice = wrapperspb.Float(-PriceOrder)
		req.BillingItems[index].IsCancelBillItem = wrapperspb.Bool(true)
		req.BillingItems[index].TaxItem.TaxAmount = s.calculateTaxAmount(PriceOrder, 0, 20)
		studentProduct, err := s.getStudentProductByProductID(ctx, req.BillingItems[index].ProductId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.BillingItems[index].StudentProductId = wrapperspb.String(studentProduct.StudentProductID.String)
		req.OrderItems[index].StudentProductId = wrapperspb.String(studentProduct.StudentProductID.String)
		req.OrderItems[index].CancellationDate = startDate
		req.OrderItems[index].StartDate = startDate
	}
	req.EffectiveDate = startDate
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudentProductByProductID(ctx context.Context, productID string) (*entities.StudentProduct, error) {
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

func (s *suite) updateOrderOneTimePackageSuccess(ctx context.Context) (context.Context, error) {
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

	billItems, err := s.getBillItems(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	foundBillItem := countAffectedBillItem(billItems, req.BillingItems)
	if foundBillItem < len(req.BillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("create miss billItem")
	}

	return StepStateToContext(ctx, stepState), nil
}
