package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) createOrderWithOneTimeProductsSuccessfully(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var (
		createOrderReq *pb.CreateOrderRequest
		orderItems     []*pb.OrderItem
		billingItems   []*pb.BillingItem
	)
	settings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertProductGrade:    true,
		insertStudent:         true,
		insertMaterial:        true,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertFee:             true,
		insertPackage:         true,
		insertProductDiscount: true,
	}
	taxID, locationID, feeIDs, materialIDs, packageIDs, courseIDs, discountIDs, userID,
		err := s.insertAllDataForInsertOrderWithOneTimeProducts(ctx, settings)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	{
		// Make data for package order items, bill items
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
		orderItems = append(orderItems,
			&pb.OrderItem{
				ProductId:   packageIDs[0],
				CourseItems: courseItems},
			&pb.OrderItem{
				ProductId:   packageIDs[1],
				CourseItems: courseItems,
			},
			&pb.OrderItem{
				ProductId:   packageIDs[2],
				CourseItems: courseItems,
			})
		billingItems = append(billingItems,
			&pb.BillingItem{
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
			},
			&pb.BillingItem{
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
			},
			&pb.BillingItem{
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
			},
		)

		// Make data for material order items, bill items
		orderItems = append(orderItems,
			&pb.OrderItem{ProductId: materialIDs[0]},
			&pb.OrderItem{
				ProductId:  materialIDs[1],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			},
			&pb.OrderItem{
				ProductId:  materialIDs[2],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			})
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

		// make data for fee order items, bill items
		orderItems = append(orderItems,
			&pb.OrderItem{ProductId: feeIDs[0]},
			&pb.OrderItem{
				ProductId:  feeIDs[1],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			},
			&pb.OrderItem{
				ProductId:  feeIDs[2],
				DiscountId: &wrapperspb.StringValue{Value: discountIDs[1]},
			})
		billingItems = append(billingItems,
			&pb.BillingItem{
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
			},
			&pb.BillingItem{
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
			},
			&pb.BillingItem{
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
	}

	createOrderReq = &pb.CreateOrderRequest{
		OrderItems:   orderItems,
		BillingItems: billingItems,
		OrderType:    pb.OrderType_ORDER_TYPE_NEW,
		StudentId:    userID,
		LocationId:   locationID,
		OrderComment: "test create order with one-time products (package, fee, material)",
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	createOrderResp, err := client.CreateOrder(contextWithToken(ctx), createOrderReq)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		createOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), createOrderReq)
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = createOrderReq
	stepState.Response = createOrderResp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) voidOrderWithOneTimeProducts(ctx context.Context, account string, orderTypeTestcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var orderID string
	switch orderTypeTestcase {
	case "custom billing":
		createOrdersRsp := stepState.Response.(*pb.CreateCustomBillingResponse)
		orderID = createOrdersRsp.OrderId
	default:
		createOrdersRsp := stepState.Response.(*pb.CreateOrderResponse)
		orderID = createOrdersRsp.OrderId
	}
	voidOrderReq := &pb.VoidOrderRequest{
		OrderId: orderID,
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	voidOrderResp, err := client.VoidOrder(contextWithToken(ctx), voidOrderReq)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		voidOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			VoidOrder(contextWithToken(ctx), voidOrderReq)
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = voidOrderReq
	stepState.Response = voidOrderResp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) voidOrderWithOneTimeProductsSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	voidedOrderResp := stepState.Response.(*pb.VoidOrderResponse)
	if !voidedOrderResp.Successful {
		return StepStateToContext(ctx, stepState), fmt.Errorf("void order failed")
	}

	order, err := s.getOrder(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if order.OrderStatus.String != pb.OrderStatus_ORDER_STATUS_VOIDED.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when order status wrong data")
	}

	billItems, err := s.getBillItems(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for _, billItem := range billItems {
		if billItem.BillStatus.String != pb.BillingStatus_BILLING_STATUS_CANCELLED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when bill item status wrong data")
		}
	}

	orderItems, err := s.getOrderItems(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	studentProductIDs := make([]string, 0, len(orderItems))
	for _, orderItem := range orderItems {
		studentProductIDs = append(studentProductIDs, orderItem.StudentProductID.String)
	}
	studentProducts, err := s.getStudentProductsByIDs(ctx, studentProductIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for _, studentProduct := range studentProducts {
		if studentProduct.ProductStatus.String != pb.StudentProductStatus_CANCELLED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when student product status wrong data")
		}
	}

	orderActionLogs, err := s.getOrderActionLogs(ctx, voidedOrderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	isVoidedAction := false
	for _, actionLog := range orderActionLogs {
		if actionLog.Action.String == pb.OrderActionStatus_ORDER_ACTION_VOIDED.String() && actionLog.UserID.String == stepState.CurrentUserID {
			isVoidedAction = true
		}
	}
	if !isVoidedAction {
		return StepStateToContext(ctx, stepState), fmt.Errorf("void order action log invalid content")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertAllDataForInsertOrderWithOneTimeProducts(ctx context.Context, settings PrepareDataForCreatingOrderSettings) (
	taxID string,
	locationID string,
	feeIDs []string,
	materialIDs []string,
	packageIDs []string,
	courseIDs []string,
	discountIDs []string,
	userID string,
	err error,
) {
	gradeID, err := mockdata.InsertOneGrade(ctx, s.FatimaDBTrace)
	if err != nil {
		return
	}

	var productIDs []string
	if settings.insertTax {
		taxID, err = mockdata.InsertOneTax(ctx, s.FatimaDBTrace, "test-insert-one-time-products-tax")
		if err != nil {
			return
		}
	}
	if settings.insertDiscount {
		discountIDs, err = mockdata.InsertOneDiscountAmount(ctx, s.FatimaDBTrace, "test-insert-one-time-products-discount")
		if err != nil {
			return
		}
	}
	if settings.insertLocation {
		locationID, err = mockdata.InsertOneLocation(ctx, s.FatimaDBTrace)
		if err != nil {
			return
		}
	} else {
		locationID = constants.ManabieOrgLocation
	}
	if settings.insertStudent {
		userID, err = mockdata.InsertOneUser(ctx, s.FatimaDBTrace, gradeID)
		if err != nil {
			return
		}
	}
	if settings.insertFee {
		feeIDs, err = mockdata.InsertFee(ctx, s.FatimaDBTrace, taxID)
		if err != nil {
			return
		}
		productIDs = append(productIDs, feeIDs...)
	}
	if settings.insertMaterial {
		materialIDs, err = s.insertMaterial(ctx, taxID)
		if err != nil {
			return
		}
		productIDs = append(productIDs, materialIDs...)
	}
	if settings.insertPackage {
		packageIDs, err = s.insertPackage(ctx, taxID)
		if err != nil {
			return
		}
		courseIDs, err = s.insertCourses(ctx)
		if err != nil {
			return
		}
		productIDs = append(productIDs, packageIDs...)
		if err = utils.GroupErrorFunc(
			s.insertProductPriceForPackage(ctx, packageIDs),
			s.insertPackageCourses(ctx, packageIDs, courseIDs),
		); err != nil {
			return
		}
	}
	if settings.insertProductLocation {
		err = mockdata.InsertProductLocation(ctx, s.FatimaDBTrace, locationID, productIDs)
		if err != nil {
			return
		}
	}
	if settings.insertProductPrice {
		err = s.insertProductPrice(ctx, productIDs, pb.ProductPriceType_DEFAULT_PRICE.String())
		if err != nil {
			return
		}
	}
	if settings.insertProductGrade {
		err = mockdata.InsertProductGrade(ctx, s.FatimaDBTrace, gradeID, productIDs)
		if err != nil {
			return
		}
	}

	if settings.insertProductDiscount {
		err = mockdata.InsertProductDiscount(ctx, s.FatimaDBTrace, productIDs, discountIDs)
		if err != nil {
			return
		}
	}
	return
}

func (s *suite) getStudentProductsByIDs(ctx context.Context, studentProductIDs []string) (studentProducts []entities.StudentProduct, err error) {
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = ANY($1)
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, studentProductIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		studentProducts = append(studentProducts, *studentProduct)
	}
	return studentProducts, nil
}

func (s *suite) voidOrderWithOneTimeProductsOutOfVersion(ctx context.Context, account string, orderTypeTestcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var orderID string
	createOrdersRsp := stepState.Response.(*pb.CreateOrderResponse)
	orderID = createOrdersRsp.OrderId
	voidOrderReq := &pb.VoidOrderRequest{
		OrderId:            orderID,
		OrderVersionNumber: int32(5),
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	voidOrderResp, err := client.VoidOrder(contextWithToken(ctx), voidOrderReq)

	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		voidOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			VoidOrder(contextWithToken(ctx), voidOrderReq)
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}
	stepState.Request = voidOrderReq
	stepState.Response = voidOrderResp

	return StepStateToContext(ctx, stepState), nil
}
