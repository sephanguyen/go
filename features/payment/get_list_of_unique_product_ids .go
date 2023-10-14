package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForUniqueProductList(ctx context.Context, typeOfOrder string, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch testcase {
	case "one time material":
		defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
			insertTax:             true,
			insertDiscount:        true,
			insertStudent:         true,
			insertMaterialUnique:  false,
			insertProductPrice:    true,
			insertProductLocation: true,
			insertLocation:        false,
			insertProductGrade:    true,
			insertFee:             false,
			insertProductDiscount: true,
			insertMaterial:        false,
			insertFeeUnique:       true,
		}
		taxID,
			discountIDs,
			locationID,
			materialIDs,
			userID,
			err := s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
		billingItems := make([]*pb.BillingItem, 0, len(materialIDs))

		orderItems = append(orderItems, &pb.OrderItem{ProductId: materialIDs[0]},
			&pb.OrderItem{
				ProductId:  materialIDs[1],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			},
			&pb.OrderItem{
				ProductId:  materialIDs[2],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			})
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
		}, &pb.BillingItem{
			ProductId: materialIDs[1],
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
				TaxAmount:     s.calculateTaxAmount(PriceOrder, 20, 20),
			},
			FinalPrice: PriceOrder - 20,
		}, &pb.BillingItem{
			ProductId: materialIDs[2],
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
				TaxAmount:     s.calculateTaxAmount(PriceOrder, PriceOrder*20/100, 20),
			},
			FinalPrice: PriceOrder - PriceOrder*20/100,
		})

		stepState.Request = &pb.CreateOrderRequest{
			OrderItems:   orderItems,
			BillingItems: billingItems,
			OrderType:    pb.OrderType_ORDER_TYPE_NEW,
			StudentId:    userID,
			LocationId:   locationID,
			OrderComment: "test create order material one time",
		}
	case "recurring material":
		defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
			InsertTax:                     true,
			InsertDiscount:                true,
			InsertStudent:                 true,
			InsertProductPrice:            true,
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
			InsertProductDiscount:         true,
			InsertMaterialUnique:          true,
			BillingScheduleStartDate:      time.Now(),
		}

		var req pb.CreateOrderRequest
		var err error

		defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
		req, err = s.validCaseBilledAtOrderItemsSingleItemExpected(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.Request = &req
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrdersForUniqueProduct(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resps := make([]*pb.CreateOrderResponse, 0)
	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	req := stepState.Request.(*pb.CreateOrderRequest)
	resp, err := client.CreateOrder(contextWithToken(ctx), req)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		resp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateOrderRequest))
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	resps = append(resps, resp)

	stepState.Response = resps

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getUniqueProductList(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reqCreateOrders := stepState.Request.(*pb.CreateOrderRequest)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &pb.RetrieveListOfUniqueProductIDsRequest{
		StudentId: reqCreateOrders.StudentId,
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.RetrieveListOfUniqueProductIDs(contextWithToken(ctx), req)
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

func (s *suite) checkResponseUniqueProduct(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveListOfUniqueProductIDsResponse)

	for _, item := range resp.ProductDetails {
		switch testcase {
		case "one time material":
			if item.EndTime != nil {
				return ctx, fmt.Errorf("wrong end time of one time material")
			}
		case "recurring material":
			if item.EndTime != nil {
				return ctx, fmt.Errorf("wrong end time of recurring material")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
