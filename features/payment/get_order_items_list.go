package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForGetOderItemsList(ctx context.Context, typeOfOrder string, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch testcase {
	case "package type one time":
		defaultOptionPrepareData := optionToPrepareDataForCreateOrderPackageOneTime{
			insertStudent:                    true,
			insertPackage:                    true,
			insertPackageCourse:              true,
			insertCourse:                     true,
			insertProductPrice:               true,
			insertProductLocation:            true,
			insertLocation:                   false,
			insertProductGrade:               true,
			insertPackageQuantityTypeMapping: true,
		}
		taxID,
			locationID,
			packageIDs,
			courseIDs,
			userID,
			err := s.insertAllDataForInsertOrderPackageOneTime(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		orderItems := make([]*pb.OrderItem, 0, len(packageIDs))
		billingItems := make([]*pb.BillingItem, 0, len(packageIDs))
		courseItems := []*pb.CourseItem{
			{
				CourseId:   courseIDs[0],
				CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[0]),
				Weight:     &wrapperspb.Int32Value{Value: 1},
			},
			{
				CourseId:   courseIDs[1],
				CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[1]),
				Weight:     &wrapperspb.Int32Value{Value: 2},
			},
			{
				CourseId:   courseIDs[2],
				CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[2]),
				Weight:     &wrapperspb.Int32Value{Value: 3},
			},
		}
		startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
		orderItems = append(
			orderItems,
			&pb.OrderItem{ProductId: packageIDs[0], CourseItems: courseItems, StartDate: startDate},
			&pb.OrderItem{
				ProductId:   packageIDs[1],
				CourseItems: courseItems,
				StartDate:   startDate,
			},
			&pb.OrderItem{
				ProductId:   packageIDs[2],
				CourseItems: courseItems,
				StartDate:   startDate,
			})

		billingItems = append(billingItems, &pb.BillingItem{
			ProductId: packageIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     s.calculateTaxAmount(PriceOrder, 0, 20),
			},
			CourseItems: courseItems,
			FinalPrice:  PriceOrder,
		}, &pb.BillingItem{
			ProductId: packageIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     s.calculateTaxAmount(PriceOrder, 0, 20),
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		}, &pb.BillingItem{
			ProductId: packageIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 6},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     s.calculateTaxAmount(PriceOrder, 0, 20),
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		})

		stepState.Request = &pb.CreateOrderRequest{
			OrderItems:   orderItems,
			BillingItems: billingItems,
			OrderType:    pb.OrderType_ORDER_TYPE_NEW,
			StudentId:    userID,
			LocationId:   locationID,
			OrderComment: "test create order",
		}
	case "material type one time":
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

		startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
		orderItems = append(orderItems, &pb.OrderItem{ProductId: materialIDs[0], StartDate: startDate},
			&pb.OrderItem{
				ProductId:  materialIDs[1],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
				StartDate:  startDate,
			},
			&pb.OrderItem{
				ProductId:  materialIDs[2],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
				StartDate:  startDate,
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
	case "fee type one time":
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
		const NameOneTimeFee = "one time fee"
		taxID, discountIDs, locationID, feeIDs, userID, err := s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, NameOneTimeFee)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		orderItems := make([]*pb.OrderItem, 0, len(feeIDs))
		billingItems := make([]*pb.BillingItem, 0, len(feeIDs))

		startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
		orderItems = append(orderItems, &pb.OrderItem{ProductId: feeIDs[0], StartDate: startDate},
			&pb.OrderItem{
				ProductId:  feeIDs[1],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
				StartDate:  startDate,
			},
			&pb.OrderItem{
				ProductId:  feeIDs[2],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
				StartDate:  startDate,
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
				TaxAmount:     s.calculateTaxAmount(PriceOrder, 20, 20),
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
			OrderComment: "test create order fee one time",
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrdersForOrderItems(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
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

func (s *suite) checkResponseData(expectedResponse *pb.RetrieveListOfOrderItemsResponse, resp *pb.RetrieveListOfOrderItemsResponse) error {
	for _, item := range resp.Items {
		if len(item.LocationInfo.LocationId) == 0 {
			return fmt.Errorf("wrong locationID")
		}

		if len(item.OrderId) == 0 {
			return fmt.Errorf("wrong OrderID")
		}

		if len(item.ProductDetails) == 0 {
			return fmt.Errorf("wrong ProductDetails")
		}
		if item.OrderStatus != pb.OrderStatus_ORDER_STATUS_SUBMITTED {
			return fmt.Errorf("wrong OrderStatus: %v", item.OrderStatus)
		}
		if item.OrderType != pb.OrderType_ORDER_TYPE_NEW {
			return fmt.Errorf("wrong OrderType: %v", item.OrderType)
		}
	}

	if resp.NextPage.GetOffsetInteger() != expectedResponse.NextPage.GetOffsetInteger() {
		return fmt.Errorf("wrong NextPage.Offset")
	}

	if resp.NextPage.Limit != expectedResponse.NextPage.Limit {
		return fmt.Errorf("wrong NextPage.Limit")
	}

	if resp.PreviousPage != nil {
		return fmt.Errorf("wrong PreviousPage.Offset")
	}

	return nil
}

func (s *suite) getOrderItemsListWithFilter(ctx context.Context, account, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch testcase {
	case "without filter":
		reqCreateOrders := stepState.Request.(*pb.CreateOrderRequest)
		req := &pb.RetrieveListOfOrderItemsRequest{
			StudentId:   reqCreateOrders.StudentId,
			LocationIds: []string{reqCreateOrders.LocationId},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfOrderItems(contextWithToken(ctx), req)

		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response = resp

		if len(resp.Items) != 2 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect 2 items returned")
		}

		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}

		expectedResponse := &pb.RetrieveListOfOrderItemsResponse{
			Items: []*pb.RetrieveListOfOrderItemsResponse_OrderItems{},
			NextPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 2,
				},
			},
			TotalItems: 6,
		}
		err = s.checkResponseData(expectedResponse, resp)
		if err != nil {
			return nil, err
		}
	case "empty filter":
		req := &pb.RetrieveListOfOrderItemsRequest{
			StudentId:   "",
			LocationIds: []string{},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfOrderItems(contextWithToken(ctx), req)
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
	case "filter with empty response":
		req := &pb.RetrieveListOfOrderItemsRequest{
			StudentId:   "invalid_student_id",
			LocationIds: []string{"invalid_location_id"},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfOrderItems(contextWithToken(ctx), req)
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
	}

	return StepStateToContext(ctx, stepState), nil
}
