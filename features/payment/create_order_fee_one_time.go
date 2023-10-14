package payment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const OneTimeFee = "one time fee"

func (s *suite) prepareDataForCreateOrderOneTimeFee(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		feeIDs      []string
		req         pb.CreateOrderRequest
		err         error
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
	taxID, discountIDs, locationID, feeIDs, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeFee)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order fee one time"

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

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreateOrderOneTimeFeeWithCase(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		locationID string
		userID     string
		err        error
	)
	req := pb.CreateOrderRequest{}
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
	switch testcase {
	case "not-exist student":
		req.StudentId = "invalid user"
		stepState.Request = &req
	case "not-exist location":
		defaultPrepareDataSettings.insertTax = false
		defaultPrepareDataSettings.insertDiscount = false
		defaultPrepareDataSettings.insertFee = false
		defaultPrepareDataSettings.insertLocation = false
		defaultPrepareDataSettings.insertProductLocation = false
		defaultPrepareDataSettings.insertProductPrice = false
		_, _, locationID, _, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeFee)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.StudentId = userID
		req.LocationId = locationID
		stepState.Request = &req
	case "not-exist product location":
		defaultPrepareDataSettings.insertProductLocation = false
		req, _ = s.ProductNotSupportCase(ctx, defaultPrepareDataSettings)
		stepState.Request = &req
	case "not-exist product grade":
		defaultPrepareDataSettings.insertProductGrade = false
		req, _ = s.ProductNotSupportCase(ctx, defaultPrepareDataSettings)
		stepState.Request = &req
	case "not-exist product price":
		req, _ = s.NotExistProductPriceCase(ctx, defaultPrepareDataSettings)
		stepState.Request = &req
	default:
		fmt.Println("None testcase")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesErrMessageForCreateOneTimeFee(ctx context.Context, expectedErrMessage, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	req := stepState.Request.(*pb.CreateOrderRequest)
	switch testcase {
	case "not-exist product location":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, []string{req.BillingItems[0].ProductId, req.BillingItems[1].ProductId, req.BillingItems[2].ProductId}, req.LocationId)
	case "not-exist product grade":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, []string{req.BillingItems[0].ProductId, req.BillingItems[1].ProductId, req.BillingItems[2].ProductId}, "1")
	default:
	}
	if !strings.Contains(stt.Message(), expectedErrMessage) {
		return ctx, fmt.Errorf("expecting %s, got %s error message ", expectedErrMessage, stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertFeeOutOfTime(ctx context.Context, taxID string) ([]string, error) {
	type AddProductParams struct {
		Name                 string         `json:"name"`
		ProductID            string         `json:"product_id"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
		ProductTag           sql.NullString `json:"product_tag"`
		ProductPartnerID     sql.NullString `json:"product_partner_id"`
		AvailableFrom        time.Time      `json:"available_from"`
		AvailableUtil        time.Time      `json:"available_until"`
		Remarks              sql.NullString `json:"remarks"`
		CustomBillingPeriod  sql.NullTime   `json:"custom_billing_period"`
		BillingScheduleID    sql.NullString `json:"billing_schedule_id"`
		DisableProRatingFlag bool           `json:"disable_pro_rating_flag"`
		IsArchived           bool           `json:"is_archived"`
		UpdatedAt            time.Time      `json:"updated_at"`
		CreatedAt            time.Time      `json:"created_at"`
	}

	type AddFeeParams struct {
		FeeID   string `json:"fee_id"`
		FeeType string `json:"fee_type"`
	}
	var feeIDs []string
	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var feeArg AddFeeParams
		randomStr := idutil.ULIDNow()
		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("fee-for-create-order-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_FEE.String()
		productArg.AvailableFrom = time.Now().AddDate(1, 0, 0)
		productArg.AvailableUtil = time.Now().AddDate(2, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.TaxID = sql.NullString{String: taxID, Valid: true}
		stmt := `INSERT INTO product (
				   product_id,
                   name,
                   product_type,
                   tax_id,
				   product_tag,
				   product_partner_id,
                   available_from,
                   available_until,
                   remarks,
                   custom_billing_period,
                   billing_schedule_id,
                   disable_pro_rating_flag,
                   is_archived,
                   updated_at,
                   created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now(), now())
				RETURNING product_id`
		row := s.FatimaDBTrace.QueryRow(ctx, stmt,
			productArg.ProductID,
			productArg.Name,
			productArg.ProductType,
			productArg.TaxID,
			productArg.ProductTag,
			productArg.ProductPartnerID,
			productArg.AvailableFrom,
			productArg.AvailableUtil,
			productArg.Remarks,
			productArg.CustomBillingPeriod,
			productArg.BillingScheduleID,
			productArg.DisableProRatingFlag,
			productArg.IsArchived,
		)
		err := row.Scan(&feeArg.FeeID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		feeIDs = append(feeIDs, feeArg.FeeID)
		queryInsertPackage := `INSERT INTO fee (fee_id, fee_type) VALUES ($1, $2)
		`
		feeArg.FeeType = pb.FeeType_FEE_TYPE_ONE_TIME.String()
		_, err = s.FatimaDBTrace.Exec(ctx, queryInsertPackage,
			feeArg.FeeID,
			feeArg.FeeType,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert fee, err: %s", err)
		}
	}
	return feeIDs, nil
}

func (s *suite) createOrderOneTimeFeeSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	err := s.checkCreatedOrderDetailsAndActionLogs(ctx, pb.OrderType_ORDER_TYPE_NEW)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.validateCreatedOrderItemsAndBillItemsForOneTimeProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
