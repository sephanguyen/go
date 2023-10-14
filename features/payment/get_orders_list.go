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

func (s *suite) calculateTaxAmount(price, discountAmount, taxPercentage float32) float32 {
	priceAfterDiscount := price - discountAmount
	return float32(float64(priceAfterDiscount*taxPercentage) / float64(100+taxPercentage))
}

func (s *suite) prepareDataForGetOrderList(ctx context.Context) (context.Context, error) {
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
	startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	orderItems = append(
		orderItems, &pb.OrderItem{ProductId: materialIDs[0], StartDate: startDate},
		&pb.OrderItem{
			ProductId:  materialIDs[1],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			StartDate:  startDate,
		},
		&pb.OrderItem{
			ProductId:  materialIDs[2],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			StartDate:  startDate,
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

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrders(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resps := make([]*pb.CreateOrderResponse, 0)
	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	for i := 0; i < 6; i++ {
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
	}
	stepState.Response = resps

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getOrderListWithFilter(ctx context.Context, userGroup, getOrdersFilter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now().UTC()
	switch getOrdersFilter {
	case "without filter":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter: &pb.RetrieveListOfOrdersFilter{
				OrderTypes: []pb.OrderType{
					pb.OrderType_ORDER_TYPE_NEW,
				},
				ProductIds:  []string{},
				CreatedFrom: timestamppb.New(now.AddDate(-1, 0, 0)),
				CreatedTo:   timestamppb.New(now.AddDate(1, 0, 0)),
			},
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 5,
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

		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items: []*pb.RetrieveListOfOrdersResponse_Order{},
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 15,
				},
			},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
		}

		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "filter with empty response":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter: &pb.RetrieveListOfOrdersFilter{
				OrderTypes: []pb.OrderType{
					pb.OrderType_ORDER_TYPE_NEW,
				},
				ProductIds:  []string{},
				CreatedFrom: timestamppb.New(now.AddDate(5, 0, 0)),
				CreatedTo:   timestamppb.New(now.AddDate(6, 0, 0)),
			},
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 10,
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

		if len(resp.Items) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong count item response")
		}

		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items:            []*pb.RetrieveListOfOrdersResponse_Order{},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
		}

		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "empty filter":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter:      nil,
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 5,
				},
			},
		}
		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items: []*pb.RetrieveListOfOrdersResponse_Order{},
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 15,
				},
			},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
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

		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "filter with paginated result":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter:      nil,
			Paging: &cpb.Paging{
				Limit: 1,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 2,
				},
			},
		}
		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items: []*pb.RetrieveListOfOrdersResponse_Order{},
			PreviousPage: &cpb.Paging{
				Limit: 1,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 1,
				},
			},
			NextPage: &cpb.Paging{
				Limit: 1,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 3,
				},
			},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
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

		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "keyword filter case insensitive match":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_ALL,
			Filter:      nil,
			Keyword:     "TudEn",
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 5,
				},
			},
		}
		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items: []*pb.RetrieveListOfOrdersResponse_Order{},
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 15,
				},
			},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
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
		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "keyword filter no student match":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_ALL,
			Filter:      nil,
			Keyword:     "impossiblestudentmatch",
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items: []*pb.RetrieveListOfOrdersResponse_Order{},
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 10,
				},
			},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
		}
		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfOrders(contextWithToken(ctx), req)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}

		if len(resp.Items) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong count item response")
		}

		stepState.Response = resp
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}
		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "product id filter empty response":
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter: &pb.RetrieveListOfOrdersFilter{
				OrderTypes: []pb.OrderType{
					pb.OrderType_ORDER_TYPE_NEW,
				},
				ProductIds:  []string{"ANGELOGUIAM"},
				CreatedFrom: timestamppb.New(now.AddDate(-1, 0, 0)),
				CreatedTo:   timestamppb.New(now.AddDate(1, 0, 0)),
			},
			Paging: &cpb.Paging{
				Limit: 10,
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

		if len(resp.Items) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong count item response")
		}

		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items:            []*pb.RetrieveListOfOrdersResponse_Order{},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
		}

		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "valid product id filter":
		productIDs, err := s.getProductIDsForSearchFilter(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error getting product IDs for filter")
		}
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter: &pb.RetrieveListOfOrdersFilter{
				OrderTypes: []pb.OrderType{
					pb.OrderType_ORDER_TYPE_NEW,
				},
				ProductIds:  productIDs,
				CreatedFrom: timestamppb.New(now.AddDate(-1, 0, 0)),
				CreatedTo:   timestamppb.New(now.AddDate(1, 0, 0)),
			},
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 5,
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
		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items: []*pb.RetrieveListOfOrdersResponse_Order{},
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 15,
				},
			},
			TotalItems:       6,
			TotalOfSubmitted: 6,
			TotalOfPending:   0,
			TotalOfRejected:  0,
			TotalOfVoided:    0,
			TotalOfInvoiced:  0,
		}
		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	case "only is not reviewed filter":
		productIDs, err := s.getProductIDsForSearchFilter(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error getting product IDs for filter")
		}
		req := &pb.RetrieveListOfOrdersRequest{
			CurrentTime: timestamppb.New(time.Now()),
			OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			Filter: &pb.RetrieveListOfOrdersFilter{
				OrderTypes: []pb.OrderType{
					pb.OrderType_ORDER_TYPE_NEW,
				},
				ProductIds:      productIDs,
				CreatedFrom:     timestamppb.New(now.AddDate(-1, 0, 0)),
				CreatedTo:       timestamppb.New(now.AddDate(1, 0, 0)),
				OnlyNotReviewed: true,
			},
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 5,
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
		expectedResponse := &pb.RetrieveListOfOrdersResponse{
			Items: []*pb.RetrieveListOfOrdersResponse_Order{},
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 15,
				},
			},
			TotalItems:               6,
			TotalOfSubmitted:         6,
			TotalOfPending:           0,
			TotalOfRejected:          0,
			TotalOfVoided:            0,
			TotalOfInvoiced:          0,
			TotalOfOrderNeedToReview: 6,
		}
		err = s.checkResponse(expectedResponse, resp)
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkResponse(expectedResponse *pb.RetrieveListOfOrdersResponse, resp *pb.RetrieveListOfOrdersResponse) error {
	for _, item := range resp.Items {
		if item.OrderSequenceNumber == 0 {
			return fmt.Errorf("wrong OrderSequenceNumber")
		}
		if len(item.OrderId) == 0 {
			return fmt.Errorf("wrong OrderID")
		}
		if len(item.StudentId) == 0 {
			return fmt.Errorf("wrong StudentId")
		}
		if item.CreatorInfo == nil {
			return fmt.Errorf("wrong CreatorInfo")
		}
	}

	if resp.NextPage != nil && expectedResponse.NextPage != nil {
		if resp.NextPage.GetOffsetInteger() != expectedResponse.NextPage.GetOffsetInteger() {
			return fmt.Errorf("wrong NextPage.Offset, %v", resp.NextPage.GetOffsetInteger())
		}

		if resp.NextPage.Limit != expectedResponse.NextPage.Limit {
			return fmt.Errorf("wrong NextPage.Limit")
		}
	}

	if resp.PreviousPage != nil && expectedResponse.PreviousPage != nil {
		if resp.PreviousPage.GetOffsetInteger() != expectedResponse.PreviousPage.GetOffsetInteger() {
			return fmt.Errorf("wrong PreviousPage.Offset, %v", resp.PreviousPage.GetOffsetInteger())
		}

		if resp.PreviousPage.Limit != expectedResponse.PreviousPage.Limit {
			return fmt.Errorf("wrong PreviousPage.Limit")
		}
	}

	return nil
}

func (s *suite) getProductIDsForSearchFilter(ctx context.Context) (productIDs []string, err error) {
	product := &entities.Product{}
	productFieldNames, productFieldValues := product.FieldMap()
	stmt :=
		`
		SELECT
			%s
		FROM 
			%s
		ORDER BY created_at DESC
		LIMIT 5 
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productFieldNames, ","),
		product.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(productFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		productIDs = append(productIDs, product.ProductID.String)
	}

	return
}
