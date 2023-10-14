package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForGetBillItemsList(ctx context.Context, typeOfOrder string, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch testcase {
	case "one time package":
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
	case "one time material":
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
		orderItems = append(orderItems, &pb.OrderItem{ProductId: materialIDs[0]},
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
	case "one time fee":
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
		orderItems = append(orderItems, &pb.OrderItem{ProductId: feeIDs[0]},
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
			InsertMaterial:                true,
			InsertBillingSchedule:         true,
			InsertBillingScheduleArchived: false,
			IsTaxExclusive:                false,
			InsertDiscountNotAvailable:    false,
			InsertProductOutOfTime:        false,
			InsertProductDiscount:         true,
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

func (s *suite) createOrdersForBillItems(ctx context.Context, account string) (context.Context, error) {
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

func (s *suite) checkResponseBillItems(expectedResponse *pb.RetrieveListOfBillItemsResponse, resp *pb.RetrieveListOfBillItemsResponse) error {
	for _, item := range resp.Items {
		if item.BillingNo == 0 {
			return fmt.Errorf("wrong BillSequenceNumber")
		}
		if len(item.OrderId) == 0 {
			return fmt.Errorf("wrong OrderID")
		}

		if item.BillingStatus != pb.BillingStatus_BILLING_STATUS_BILLED && item.BillingStatus != pb.BillingStatus_BILLING_STATUS_PENDING {
			return fmt.Errorf("wrong BillingStatus: %v", item.BillingStatus)
		}
		if item.BillingType != pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER && item.BillingType != pb.BillingType_BILLING_TYPE_UPCOMING_BILLING {
			return fmt.Errorf("wrong BillingType: %v", item.BillingType)
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

func (s *suite) getBillItemsListWithFilter(ctx context.Context, account, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch testcase {
	case "valid filter":
		reqCreateOrders := stepState.Request.(*pb.CreateOrderRequest)

		req := &pb.RetrieveListOfBillItemsRequest{
			StudentId:   reqCreateOrders.StudentId,
			LocationIds: []string{reqCreateOrders.LocationId},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 2,
				},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfBillItems(contextWithToken(ctx), req)
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
		expectedResponse := &pb.RetrieveListOfBillItemsResponse{
			Items: []*pb.RetrieveListOfBillItemsResponse_BillItems{},
			NextPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 4,
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
			TotalItems: 6,
		}
		err = s.checkResponseBillItems(expectedResponse, resp)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "empty filter location":
		reqCreateOrders := stepState.Request.(*pb.CreateOrderRequest)
		req := &pb.RetrieveListOfBillItemsRequest{
			StudentId:   reqCreateOrders.StudentId,
			LocationIds: []string{},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 2,
				},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfBillItems(contextWithToken(ctx), req)
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
		expectedResponse := &pb.RetrieveListOfBillItemsResponse{
			Items: []*pb.RetrieveListOfBillItemsResponse_BillItems{},
			NextPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 4,
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
			TotalItems: 6,
		}
		err = s.checkResponseBillItems(expectedResponse, resp)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "filter with empty response":
		req := &pb.RetrieveListOfBillItemsRequest{
			StudentId: "invalid_student_id",

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
		resp, err := client.RetrieveListOfBillItems(contextWithToken(ctx), req)
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
