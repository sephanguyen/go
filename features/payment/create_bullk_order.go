package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForCreateBulkOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		feeIDs      []string
		err         error
		req         *pb.CreateBulkOrderRequest
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertStudent:         true,
		insertFee:             true,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertMaterial:        false,
		insertProductDiscount: true,
	}
	req = &pb.CreateBulkOrderRequest{}
	for i := 0; i < 3; i++ {
		reqItem := &pb.CreateBulkOrderRequest_CreateNewOrderRequest{}
		taxID, discountIDs, locationID, feeIDs, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeFee)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		reqItem.StudentId = userID
		reqItem.LocationId = locationID
		reqItem.OrderComment = "test create order fee one time"

		orderItems := make([]*pb.OrderItem, 0, len(feeIDs))
		billingItems := make([]*pb.BillingItem, 0, len(feeIDs))

		orderItems = append(orderItems, &pb.OrderItem{ProductId: feeIDs[0]},
			&pb.OrderItem{
				ProductId:  feeIDs[1],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			},
			&pb.OrderItem{
				ProductId:  feeIDs[2],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			})
		billingItems = append(billingItems, &pb.BillingItem{
			ProductId: feeIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice: PriceOrder,
		}, &pb.BillingItem{
			ProductId: feeIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
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
		}, &pb.BillingItem{
			ProductId: feeIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
				DiscountAmountValue: 20,
				DiscountAmount:      PriceOrder * 20 / 100,
			},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     66.666664,
			},
			FinalPrice: PriceOrder - PriceOrder*20/100,
		})

		reqItem.OrderItems = orderItems
		reqItem.BillingItems = billingItems
		reqItem.OrderType = pb.OrderType_ORDER_TYPE_NEW
		req.NewOrderRequests = append(req.NewOrderRequests, reqItem)
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createBulkOrder(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState = StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
		CreateBulkOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateBulkOrderRequest))
	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1500)
		stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
			CreateBulkOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateBulkOrderRequest))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createBulkOrderSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	reqList := stepState.Request.(*pb.CreateBulkOrderRequest)
	resList := stepState.Response.(*pb.CreateBulkOrderResponse)
	mapOrderWithStudentIDinReq := make(map[string]entities.Order, len(reqList.NewOrderRequests))

	for _, res := range resList.NewOrderResponses {
		if !res.Successful {
			return StepStateToContext(ctx, stepState), fmt.Errorf("create order didn't success")
		}

		order, err := s.getOrder(ctx, res.OrderId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, request := range reqList.NewOrderRequests {
			if order.StudentID.String == request.StudentId {
				mapOrderWithStudentIDinReq[request.StudentId] = *order
				break
			}
		}
	}

	for _, req := range reqList.NewOrderRequests {
		order := mapOrderWithStudentIDinReq[req.StudentId]
		if order.OrderComment.String != req.OrderComment ||
			order.LocationID.String != req.LocationId ||
			order.OrderType.String != req.OrderType.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error create order wrong data")
		}

		orderItems, err := s.getOrderItems(ctx, order.OrderID.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		foundOrderItem := countOrderItem(orderItems, req.OrderItems)
		if foundOrderItem < len(req.OrderItems) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("create miss orderItem")
		}

		billItems, err := s.getBillItems(ctx, order.OrderID.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		foundBillItem := countBillItem(billItems, req.BillingItems, req.LocationId)
		if foundBillItem < len(req.BillingItems) {
			fmt.Println(foundBillItem, len(req.BillingItems))
			return StepStateToContext(ctx, stepState), fmt.Errorf("create miss billItem")
		}

		orderActionLogs, err := s.getOrderActionLogs(ctx, order.OrderID.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if len(orderActionLogs) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("create order action log fail")
		}

		if !(orderActionLogs[0].Action.String == pb.OrderActionStatus_ORDER_ACTION_SUBMITTED.String() &&
			orderActionLogs[0].Comment.String == order.OrderComment.String &&
			orderActionLogs[0].UserID.String == stepState.CurrentUserID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("create order action log invalid content")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
