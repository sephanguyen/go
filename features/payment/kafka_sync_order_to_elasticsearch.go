package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareOrderRecordUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
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
	taxID, discountIDs, locationID, materialIDs, studentID, err := s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, "get order list test")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	orderItems = append(
		orderItems, &pb.OrderItem{ProductId: materialIDs[0]},
		&pb.OrderItem{
			ProductId:  materialIDs[1],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
		},
		&pb.OrderItem{
			ProductId:  materialIDs[2],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
		},
	)

	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))
	billingItems = append(billingItems,
		&pb.BillingItem{
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
		},
		&pb.BillingItem{
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
		},
		&pb.BillingItem{
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
		},
	)

	stepState.Request = &pb.CreateOrderRequest{
		OrderItems:   orderItems,
		BillingItems: billingItems,
		OrderType:    pb.OrderType_ORDER_TYPE_NEW,
		StudentId:    studentID,
		LocationId:   locationID,
		OrderComment: "test create order",
	}

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

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
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) orderRecordUpdatedInDB(ctx context.Context, operation string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res := stepState.Response.(*pb.CreateOrderResponse)
	if !res.Successful {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create recurring material order")
	}

	order, err := s.getOrder(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch operation {
	case "update":
		err := s.updateOrderRecordInDB(ctx, order)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
	case "delete":
		err := s.deleteOrderRecordInDB(ctx, order)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) orderRecordReflectedInES(ctx context.Context, operation string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(3 * time.Second)

	switch operation {
	case "update":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_ALL,
			Filter:      nil,
			Keyword:     "update-order-for-sync-test",
			Paging: &cpb.Paging{
				Limit: 1,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfOrders(contextWithToken(ctx), req)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}

		stepState.Response = resp
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}

		if len(resp.Items) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to sync update in order data")
		}
		return StepStateToContext(ctx, stepState), err
	case "delete":
		createOrderRequest := stepState.Request.(*pb.CreateOrderRequest)
		productID := createOrderRequest.OrderItems[0].ProductId

		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter: &pb.RetrieveListOfOrdersFilter{
				OrderTypes: []pb.OrderType{
					pb.OrderType_ORDER_TYPE_NEW,
				},
				ProductIds:  []string{productID},
				CreatedFrom: timestamppb.New(time.Now().AddDate(-1, 0, 0)),
				CreatedTo:   timestamppb.New(time.Now().AddDate(1, 0, 0)),
			},
			Paging: &cpb.Paging{
				Limit: 1,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfOrders(contextWithToken(ctx), req)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}

		stepState.Response = resp
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}

		if len(resp.Items) > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to sync delete in order data")
		}
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrderRecordInDB(ctx context.Context, order *entities.Order) (err error) {
	orderEntity := entities.Order{}
	stmt := fmt.Sprintf(
		`UPDATE public.%s SET student_full_name = $1 WHERE order_id = $2;`,
		orderEntity.TableName(),
	)
	cmdTag, err := s.FatimaDBTrace.Exec(ctx, stmt, "update-order-for-sync-test", order.OrderID.String)
	if err != nil {
		err = fmt.Errorf("err update OrderItem: %w %s %s", err, stmt, order.OrderID.String)
		return
	}
	if cmdTag.RowsAffected() == 0 {
		err = fmt.Errorf("err update OrderItem: %d RowsAffected", cmdTag.RowsAffected())
		return
	}
	return
}

func (s *suite) deleteOrderRecordInDB(ctx context.Context, order *entities.Order) (err error) {
	billItemEntity := entities.BillItem{}
	orderEntity := entities.Order{}
	orderItemEntity := entities.OrderItem{}
	orderActionLog := entities.OrderActionLog{}
	tables := [4]string{
		billItemEntity.TableName(),
		orderActionLog.TableName(),
		orderItemEntity.TableName(),
		orderEntity.TableName(),
	}

	for _, table := range tables {
		stmt := fmt.Sprintf(
			`DELETE FROM public.%s WHERE order_id = $1;`,
			table,
		)
		_, deleteErr := s.FatimaDBTrace.Exec(ctx, stmt, order.OrderID)
		if deleteErr != nil {
			return deleteErr
		}
	}
	return
}
