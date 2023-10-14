package payment

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) AddDataForStudentBillingExport(ctx context.Context) (context.Context, error) {
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
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theOrganizationHasExistingBankBranchData(ctx context.Context, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(1000)
	ctx, err := s.AddDataForStudentBillingExport(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theUserExportStudentBilling(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	clientCreate := pb.NewOrderServiceClient(s.PaymentConn)
	clientExport := pb.NewExportServiceClient(s.PaymentConn)

	stepState.RequestSentAt = time.Now()
	for i := 0; i < 6; i++ {
		req := stepState.Request.(*pb.CreateOrderRequest)
		_, err := clientCreate.CreateOrder(contextWithToken(ctx), req)
		for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			time.Sleep(1000)
			_, err = pb.NewOrderServiceClient(s.PaymentConn).
				CreateOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateOrderRequest))
		}
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}

	}
	resp, err := clientExport.ExportStudentBilling(contextWithToken(ctx), &pb.ExportStudentBillingRequest{})
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theStudentBillingCSVHasCorrectContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*pb.ExportStudentBillingResponse)

	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be greater than 1
	if len(lines) < 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting the context line to be greater than or equal to 1 got %d", len(lines))
	}

	// check the header record
	err = checkCSVHeaderForExport(
		[]string{"student_name", "student_id", "grade", "location", "created_date", "status", "billing_item_name", "courses", "discount_name", "discount_amount", "tax_amount", "billing_amount"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
