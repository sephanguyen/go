package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/payment/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForGetOrderProductList(ctx context.Context, orderType string, productType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch productType {
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
			insertProductSetting:             true,
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
			insertProductSetting:  true,
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
			insertProductSetting:  true,
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
	case "recurring material":
		switch orderType {
		case "new":
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
				InsertLeavingReasons:          true,
				InsertProductSetting:          true,
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
		case "loa":
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
				InsertLeavingReasons:          true,
				InsertProductSetting:          true,
			}
			var (
				insertOrderReq pb.CreateOrderRequest
				req            pb.CreateOrderRequest
				billItems      []*entities.BillItem
				err            error
			)

			defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
			insertOrderReq, billItems, err = s.createRecurringMaterialWithProratingAndDiscount(ctx, defaultOptionPrepareData)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			req = s.validLOARequestWithProratingAndDiscount(&insertOrderReq, billItems)

			stepState.Request = &req
			return StepStateToContext(ctx, stepState), nil
		case "resume":
			var (
				loaReq pb.CreateOrderRequest
				req    pb.CreateOrderRequest
				err    error
			)

			loaReq, _, err = s.createLOAForResumeProductDisabledProrating(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			req = s.validResumeRequestDisabledProrating(&loaReq)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			stepState.Request = &req
			return StepStateToContext(ctx, stepState), nil
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrdersForOrderProduct(ctx context.Context, account string, orderType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resps := make([]*pb.CreateOrderResponse, 0)
	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	itirator := 6
	if orderType == "loa" || orderType == "resume" {
		itirator = 1
	}

	for i := 0; i < itirator; i++ {
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

func (s *suite) getOrderProductList(ctx context.Context, account string, locationType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reqCreateOrders := stepState.Request.(*pb.CreateOrderRequest)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var req pb.RetrieveListOfOrderProductsRequest
	switch locationType {
	case "valid location":
		req = pb.RetrieveListOfOrderProductsRequest{
			StudentId:   reqCreateOrders.StudentId,
			LocationIds: []string{reqCreateOrders.LocationId},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
	case "invalid location":
		locationIDInvalid, err := mockdata.InsertOneLocation(ctx, s.FatimaDBTrace)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
		req = pb.RetrieveListOfOrderProductsRequest{
			StudentId:   reqCreateOrders.StudentId,
			LocationIds: []string{locationIDInvalid},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
	case "empty location":
		req = pb.RetrieveListOfOrderProductsRequest{
			StudentId:   reqCreateOrders.StudentId,
			LocationIds: []string{},
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.RetrieveListOfOrderProducts(contextWithToken(ctx), &req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkResponseOrderProduct(ctx context.Context, orderType string, productType string, locationType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.CreateOrderRequest)
	resp := stepState.Response.(*pb.RetrieveListOfOrderProductsResponse)
	mapProduct := make(map[string]float32, len(resp.Items))

	var expectedResponse *pb.RetrieveListOfOrderProductsResponse

	for _, itemExpect := range req.BillingItems {
		mapProduct[itemExpect.ProductId] = itemExpect.FinalPrice
	}

	if locationType == "invalid location" {
		if len(resp.Items) != 0 {
			return ctx, fmt.Errorf("error response not empty data with invalid data")
		}
		return StepStateToContext(ctx, stepState), nil
	}

	switch orderType {
	case "loa":
		for _, orderProduct := range resp.Items {
			err := s.validatePausedProduct(orderProduct)
			if err != nil {
				return ctx, fmt.Errorf("incorrect data for paused products: %v", err)
			}
		}
	case "resume":
		for _, orderProduct := range resp.Items {
			err := s.validateResumedProduct(orderProduct)
			if err != nil {
				return ctx, fmt.Errorf("incorrect data for resume products: %v", err)
			}
		}
	default:
		expectedResponse = &pb.RetrieveListOfOrderProductsResponse{
			Items: []*pb.RetrieveListOfOrderProductsResponse_OrderProduct{},
			NextPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 2,
				},
			},
			PreviousPage: &cpb.Paging{},
			TotalItems:   6,
		}

		if len(resp.Items) != 2 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect 2 items returned")
		}

		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}

		for _, billItems := range req.BillingItems {
			if _, ok := mapProduct[billItems.ProductId]; !ok {
				return ctx, fmt.Errorf("error response product ID")
			}
			itemOfRequest := &pb.RetrieveListOfOrderProductsResponse_OrderProduct{
				LocationInfo: &pb.RetrieveListOfOrderProductsResponse_OrderProduct_LocationInfo{
					LocationId: req.LocationId,
				},
				ProductId: billItems.ProductId,
			}
			expectedResponse.Items = append(expectedResponse.Items, itemOfRequest)
		}

		for i, item := range resp.Items {
			if item.LocationInfo == nil {
				return ctx, fmt.Errorf("wrong location info")
			}
			if productType == "recurring material" {
				if item.LocationInfo.LocationId != expectedResponse.Items[0].LocationInfo.LocationId {
					return ctx, fmt.Errorf("wrong loaciton  id")
				}
			} else {
				if item.LocationInfo.LocationId != expectedResponse.Items[i].LocationInfo.LocationId {
					return ctx, fmt.Errorf("wrong loaciton  id")
				}
			}

			if item.Duration == nil {
				return ctx, fmt.Errorf("wrong duration info")
			}

			if item.Status != pb.StudentProductStatus_ORDERED {
				return ctx, fmt.Errorf("wrong BillingStatus: %v", item.Status)
			}
		}

		if resp.NextPage != nil && expectedResponse.NextPage != nil {
			if resp.NextPage.GetOffsetInteger() != expectedResponse.NextPage.GetOffsetInteger() {
				return ctx, fmt.Errorf("wrong NextPage.Offset, %v", resp.NextPage.GetOffsetInteger())
			}

			if resp.NextPage.Limit != expectedResponse.NextPage.Limit {
				return ctx, fmt.Errorf("wrong NextPage.Limit")
			}
		}

		if resp.PreviousPage != nil && expectedResponse.PreviousPage != nil {
			if resp.PreviousPage.GetOffsetInteger() != expectedResponse.PreviousPage.GetOffsetInteger() {
				return ctx, fmt.Errorf("wrong PreviousPage.Offset, %v", resp.PreviousPage.GetOffsetInteger())
			}

			if resp.PreviousPage.Limit != expectedResponse.PreviousPage.Limit {
				return ctx, fmt.Errorf("wrong PreviousPage.Limit")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validatePausedProduct(product *pb.RetrieveListOfOrderProductsResponse_OrderProduct) error {
	if product.StudentProductLabel.String() != pb.StudentProductLabel_PAUSE_SCHEDULED.String() {
		return fmt.Errorf("incorrect student product label for paused schedule product")
	}

	return nil
}

func (s *suite) validateResumedProduct(product *pb.RetrieveListOfOrderProductsResponse_OrderProduct) error {
	if product.StudentProductLabel.String() != pb.StudentProductLabel_CREATED.String() {
		return fmt.Errorf("incorrect student product label for resumed schedule product")
	}

	return nil
}
