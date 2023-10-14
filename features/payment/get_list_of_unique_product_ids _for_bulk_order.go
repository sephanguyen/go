package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForUniqueProductListForBulkOrder(ctx context.Context) (context.Context, error) {
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
		insertFee:             false,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertMaterial:        false,
		insertProductDiscount: true,
		insertMaterialUnique:  false,
		insertFeeUnique:       true,
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

func (s *suite) createBulkOrdersForUniqueProduct(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
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

func (s *suite) getUniqueProductListForBulkOrder(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	reqList := stepState.Request.(*pb.CreateBulkOrderRequest)
	resList := stepState.Response.(*pb.CreateBulkOrderResponse)
	var studentIds []string
	var UniqueProductIds []string
	mapUniqueProduct := make(map[string]*pb.BillingItem, len(resList.NewOrderResponses))
	for _, res := range resList.NewOrderResponses {
		if !res.Successful {
			return StepStateToContext(ctx, stepState), fmt.Errorf("create order didn't success")
		}

		order, err := s.getOrder(ctx, res.OrderId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, request := range reqList.NewOrderRequests {
			for _, bill := range request.BillingItems {
				if _, ok := mapUniqueProduct[bill.ProductId]; ok {
					continue
				}
				mapUniqueProduct[bill.ProductId] = bill
				UniqueProductIds = append(UniqueProductIds, bill.ProductId)

			}
			if order.StudentID.String == request.StudentId {
				studentIds = append(studentIds, request.StudentId)
				break
			}

		}

	}

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &pb.RetrieveListOfUniqueProductIDForBulkOrderRequest{
		StudentIds: studentIds,
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.RetrieveListOfUniqueProductIDForBulkOrder(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response = resp

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkResponseUniqueProductForBulkOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveListOfUniqueProductIDForBulkOrderResponse)
	if len(resp.UniqueProductOfStudent) != 3 {
		return ctx, fmt.Errorf("wrong unique product list")
	}

	return StepStateToContext(ctx, stepState), nil
}
