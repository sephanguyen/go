package payment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	PriceOrder           = 500
	EnrolledProductPrice = 300
	OneTimeMaterial      = "one time material"
)

type PrepareDataForCreatingOrderSettings struct {
	insertTax                        bool
	insertDiscount                   bool
	insertProductGrade               bool
	insertStudent                    bool
	insertMaterial                   bool
	insertProductPrice               bool
	insertEnrolledProductPrice       bool
	insertProductLocation            bool
	insertLocation                   bool
	insertFee                        bool
	insertPackage                    bool
	insertPackageCourse              bool
	insertCourse                     bool
	insertPackageQuantityTypeMapping bool
	insertProductDiscount            bool
	insertOrgLevelDiscount           bool
	insertMaterialUnique             bool
	insertFeeUnique                  bool
	insertProductSetting             bool
}

func (s *suite) prepareDataForCreateOrderOneTimeMaterial(ctx context.Context, discountType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch discountType {
	case "product discount":
		defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
			insertTax:              true,
			insertDiscount:         true,
			insertStudent:          true,
			insertMaterial:         true,
			insertProductPrice:     true,
			insertProductLocation:  true,
			insertLocation:         false,
			insertProductGrade:     true,
			insertFee:              false,
			insertProductDiscount:  true,
			insertOrgLevelDiscount: false,
			insertMaterialUnique:   false,
			insertProductSetting:   true,
		}
		req, err := s.validOneTimeMaterialOrderWithProductDiscount(ctx, defaultPrepareDataSettings)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Request = req
	case "org-level discount":
		defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
			insertTax:              true,
			insertDiscount:         false,
			insertStudent:          true,
			insertMaterial:         true,
			insertProductPrice:     true,
			insertProductLocation:  true,
			insertLocation:         false,
			insertProductGrade:     true,
			insertFee:              false,
			insertProductDiscount:  false,
			insertOrgLevelDiscount: true,
			insertMaterialUnique:   false,
			insertProductSetting:   true,
		}
		req, err := s.validOneTimeMaterialOrderWithOrgLevelDiscount(ctx, defaultPrepareDataSettings)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Request = req
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validOneTimeMaterialOrderWithProductDiscount(ctx context.Context, defaultPrepareDataSettings PrepareDataForCreatingOrderSettings) (*pb.CreateOrderRequest, error) {
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		materialIDs []string
		req         pb.CreateOrderRequest
		err         error
	)

	req = pb.CreateOrderRequest{}

	taxID, discountIDs, locationID, materialIDs, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return &req, err
	}

	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))

	startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	orderItems = append(orderItems,
		&pb.OrderItem{ProductId: materialIDs[0]},
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
	billingItems = append(billingItems,
		&pb.BillingItem{
			ProductId: materialIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
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
				TaxAmount:     80,
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
				TaxAmount:     66.666664,
			},
			FinalPrice: PriceOrder - PriceOrder*20/100,
		},
	)

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return &req, nil
}

func (s *suite) validOneTimeMaterialOrderWithOrgLevelDiscount(ctx context.Context, defaultPrepareDataSettings PrepareDataForCreatingOrderSettings) (*pb.CreateOrderRequest, error) {
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		materialIDs []string
		req         pb.CreateOrderRequest
		err         error
	)

	req = pb.CreateOrderRequest{}

	taxID, discountIDs, locationID, materialIDs, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return &req, err
	}

	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))

	startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	orderItems = append(orderItems,
		&pb.OrderItem{ProductId: materialIDs[0]},
		&pb.OrderItem{
			ProductId:  materialIDs[1],
			DiscountId: &wrapperspb.StringValue{Value: discountIDs[0]},
			StartDate:  startDate,
		},
	)
	billingItems = append(billingItems,
		&pb.BillingItem{
			ProductId: materialIDs[0],
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
			ProductId: materialIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 1},
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          discountIDs[0],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_FAMILY,
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
		},
	)

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW

	return &req, nil
}

func (s *suite) prepareDataForCreateOrderOneTimeMaterialWithCase(ctx context.Context, testcase string) (context.Context, error) {
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
		insertMaterial:        true,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertFee:             false,
		insertProductDiscount: true,
		insertMaterialUnique:  false,
	}
	switch testcase {
	case "not-exist student":
		req.StudentId = "invalid user"
		stepState.Request = &req
	case "not-exist location":
		defaultPrepareDataSettings.insertTax = false
		defaultPrepareDataSettings.insertDiscount = false
		defaultPrepareDataSettings.insertMaterial = false
		defaultPrepareDataSettings.insertLocation = false
		defaultPrepareDataSettings.insertProductLocation = false
		defaultPrepareDataSettings.insertProductPrice = false
		_,
			_,
			locationID,
			_,
			userID,
			err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
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
	case "product doesn't support":
		defaultPrepareDataSettings.insertMaterial = false
		defaultPrepareDataSettings.insertFee = true
		req, _ = s.ProductNotSupportCase(ctx, defaultPrepareDataSettings)
		stepState.Request = &req
	case "not-exist product price":
		req, _ = s.NotExistProductPriceCase(ctx, defaultPrepareDataSettings)
		stepState.Request = &req
	case "not-exist product discount":
		defaultPrepareDataSettings.insertProductDiscount = false
		req, _ = s.ProductNotSupportCase(ctx, defaultPrepareDataSettings)
		stepState.Request = &req
	default:
		fmt.Println("None testcase")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesErrMessageForCreateOneTimeMaterial(ctx context.Context, expectedErrMessage, testcase string) (context.Context, error) {
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
	case "not-exist product discount":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.BillingItems[1].ProductId, req.BillingItems[1].DiscountItem.DiscountId)
	default:
	}
	if !strings.Contains(stt.Message(), expectedErrMessage) {
		return ctx, fmt.Errorf("expecting %s, got %s error message ", expectedErrMessage, stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertAllDataForInsertOrder(ctx context.Context, defaultPrepareDataSettings PrepareDataForCreatingOrderSettings, name string) (
	taxID string,
	discountIDs []string,
	locationID string,
	productIDs []string,
	userID string,
	err error,
) {
	gradeID, err := mockdata.InsertOneGrade(ctx, s.FatimaDBTrace)
	if err != nil {
		return
	}

	if defaultPrepareDataSettings.insertTax {
		taxID, err = mockdata.InsertOneTax(ctx, s.FatimaDBTrace, name)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertDiscount {
		discountIDs, err = mockdata.InsertOneDiscountAmount(ctx, s.FatimaDBTrace, name)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertLocation {
		locationID, err = mockdata.InsertOneLocation(ctx, s.FatimaDBTrace)
		if err != nil {
			return
		}
	} else {
		locationID = constants.ManabieOrgLocation
	}

	if defaultPrepareDataSettings.insertStudent {
		userID, err = mockdata.InsertOneUser(ctx, s.FatimaDBTrace, gradeID)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertMaterial {
		productIDs, err = s.insertMaterial(ctx, taxID)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertMaterialUnique {
		productIDs, err = s.insertMaterialUnique(ctx, taxID)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertFeeUnique {
		productIDs, err = s.InsertFeeUnique(ctx, taxID)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertFee {
		productIDs, err = mockdata.InsertFee(ctx, s.FatimaDBTrace, taxID)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertProductLocation {
		err = s.insertProductLocation(ctx, locationID, productIDs)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertProductPrice {
		err = s.insertProductPrice(ctx, productIDs, pb.ProductPriceType_DEFAULT_PRICE.String())
		if err != nil {
			return
		}
	}
	if defaultPrepareDataSettings.insertEnrolledProductPrice {
		err = s.insertProductPrice(ctx, productIDs, pb.ProductPriceType_ENROLLED_PRICE.String())
		if err != nil {
			return
		}
	}
	if defaultPrepareDataSettings.insertProductGrade {
		err = s.insertProductGrade(ctx, gradeID, productIDs)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertProductDiscount {
		err = s.insertProductDiscount(ctx, productIDs, discountIDs)
		if err != nil {
			return
		}
	}

	if defaultPrepareDataSettings.insertOrgLevelDiscount {
		var orgLevelDiscountID string
		orgLevelDiscountID, err = mockdata.InsertOrgLevelDiscount(ctx, s.FatimaDBTrace, userID)
		if err != nil {
			return
		}

		discountIDs = append(discountIDs, orgLevelDiscountID)
	}

	if defaultPrepareDataSettings.insertProductSetting {
		for _, productID := range productIDs {
			err = s.insertProductSetting(ctx, productID)
			if err != nil {
				return
			}
		}
	}

	return
}

func (s *suite) insertMaterial(ctx context.Context, taxID string) ([]string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
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

	type AddMaterialParams struct {
		MaterialID        string       `json:"material_id"`
		MaterialType      string       `json:"material_type"`
		CustomBillingDate sql.NullTime `json:"custom_billing_date"`
	}
	var materialIDs []string
	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var materialArg AddMaterialParams
		randomStr := idutil.ULIDNow()
		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("material-for-create-order-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
		productArg.AvailableFrom = time.Now()
		productArg.AvailableUtil = time.Now().AddDate(1, 0, 0)
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
		err := row.Scan(&materialArg.MaterialID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		materialIDs = append(materialIDs, materialArg.MaterialID)
		queryInsertPackage := `INSERT INTO material (material_id, material_type, custom_billing_date) VALUES ($1, $2, $3)
		`
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String()
		materialArg.CustomBillingDate = sql.NullTime{Time: time.Now(), Valid: true}
		_, err = s.FatimaDBTrace.Exec(ctx, queryInsertPackage,
			materialArg.MaterialID,
			materialArg.MaterialType,
			materialArg.CustomBillingDate,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert material, err: %s", err)
		}
	}
	return materialIDs, nil
}

func (s *suite) insertMaterialUnique(ctx context.Context, taxID string) ([]string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
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
		IsUnique             bool           `json:"is_unique"`
	}

	type AddMaterialParams struct {
		MaterialID        string       `json:"material_id"`
		MaterialType      string       `json:"material_type"`
		CustomBillingDate sql.NullTime `json:"custom_billing_date"`
	}
	var materialIDs []string
	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var materialArg AddMaterialParams
		randomStr := idutil.ULIDNow()
		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("material-for-create-order-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
		productArg.AvailableFrom = time.Now()
		productArg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.IsUnique = true
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
				   is_unique,
                   updated_at,
                   created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now(), now())
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
			productArg.IsUnique,
		)
		err := row.Scan(&materialArg.MaterialID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		materialIDs = append(materialIDs, materialArg.MaterialID)
		queryInsertPackage := `INSERT INTO material (material_id, material_type, custom_billing_date) VALUES ($1, $2, $3)
		`
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String()
		materialArg.CustomBillingDate = sql.NullTime{Time: time.Now(), Valid: true}
		_, err = s.FatimaDBTrace.Exec(ctx, queryInsertPackage,
			materialArg.MaterialID,
			materialArg.MaterialType,
			materialArg.CustomBillingDate,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert material, err: %s", err)
		}
	}
	return materialIDs, nil
}

func (s *suite) InsertFeeUnique(ctx context.Context, taxID string) ([]string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
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
		IsUnique             bool           `json:"is_unique"`
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
		productArg.AvailableFrom = time.Now()
		productArg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.IsUnique = true
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
				   is_unique,
                   updated_at,
                   created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now(), now())
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
			productArg.IsUnique,
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

func (s *suite) insertMaterialOutOfTime(ctx context.Context, taxID string) ([]string, error) {
	type AddProductParams struct {
		Name                 string         `json:"name"`
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

	type AddMaterialParams struct {
		MaterialID        string       `json:"material_id"`
		MaterialType      string       `json:"material_type"`
		CustomBillingDate sql.NullTime `json:"custom_billing_date"`
	}
	var materialIDs []string
	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var materialArg AddMaterialParams
		randomStr := idutil.ULIDNow()
		productArg.Name = fmt.Sprintf("material-for-create-order-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
		productArg.AvailableFrom = time.Now().AddDate(1, 0, 0)
		productArg.AvailableUtil = time.Now().AddDate(2, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.TaxID = sql.NullString{String: taxID, Valid: true}
		stmt := `INSERT INTO product (
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
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now(), now())
				RETURNING product_id`
		row := s.FatimaDBTrace.QueryRow(ctx, stmt,
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
		err := row.Scan(&materialArg.MaterialID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		materialIDs = append(materialIDs, materialArg.MaterialID)
		queryInsertPackage := `INSERT INTO material (material_id, material_type, custom_billing_date) VALUES ($1, $2, $3)
		`
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String()
		materialArg.CustomBillingDate = sql.NullTime{Time: time.Now(), Valid: true}
		_, err = s.FatimaDBTrace.Exec(ctx, queryInsertPackage,
			materialArg.MaterialID,
			materialArg.MaterialType,
			materialArg.CustomBillingDate,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert material, err: %s", err)
		}
	}
	return materialIDs, nil
}

func (s *suite) insertProductLocation(ctx context.Context, locationID string, materialIDs []string) error {
	insertProductLocationStmt := `INSERT INTO product_location (location_id, product_id, created_at) VALUES ($1, $2, now())
		`
	for _, materialID := range materialIDs {
		_, err := s.FatimaDBTrace.Exec(ctx, insertProductLocationStmt,
			locationID,
			materialID,
		)
		if err != nil {
			return fmt.Errorf("cannot insert product_location, err: %s", err)
		}
	}
	return nil
}

func (s *suite) insertProductGrade(ctx context.Context, gradeID string, materialIDs []string) error {
	insertProductGradeStmt := `INSERT INTO product_grade (grade_id, product_id, created_at) VALUES ($1, $2, now())
		`
	for _, materialID := range materialIDs {
		_, err := s.FatimaDBTrace.Exec(ctx, insertProductGradeStmt,
			gradeID,
			materialID,
		)
		if err != nil {
			return fmt.Errorf("cannot insert product_grade, err: %s", err)
		}
	}
	return nil
}

func (s *suite) insertProductDiscount(ctx context.Context, productIDs []string, discountIDs []string) error {
	insertProductDiscountStmt := `INSERT INTO product_discount (product_id, discount_id, created_at) VALUES ($1, $2, now())
		`
	for _, productID := range productIDs {
		for _, discountID := range discountIDs {
			_, err := s.FatimaDBTrace.Exec(ctx, insertProductDiscountStmt, productID, discountID)
			if err != nil {
				return fmt.Errorf("canot insert product_discount, err: %s", err)
			}
		}
	}
	return nil
}

func (s *suite) insertProductPrice(ctx context.Context, productIDs []string, priceType string) error {
	insertDefaultProductPriceStmt := `INSERT INTO product_price (product_id, price, created_at, price_type) VALUES ($1, $2, now(), $3)
		`
	for _, materialID := range productIDs {
		_, err := s.FatimaDBTrace.Exec(ctx, insertDefaultProductPriceStmt,
			materialID,
			PriceOrder,
			priceType,
		)
		if err != nil {
			return fmt.Errorf("cannot insert product_price, err: %s", err)
		}
	}
	return nil
}

func (s *suite) userSubmitOrder(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState = StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateOrderRequest))
	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1500)
		stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateOrderRequest))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrderOneTimeMaterialSuccess(ctx context.Context) (context.Context, error) {
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

func countOrderItem(dbOrderItems []entities.OrderItem, orderItems []*pb.OrderItem) int {
	foundOrderItem := 0
	for _, item := range orderItems {
		for _, dbItem := range dbOrderItems {
			if item.ProductId == dbItem.ProductID.String &&
				((dbItem.DiscountID.Status == pgtype.Null) ||
					(dbItem.DiscountID.Status == pgtype.Present && dbItem.DiscountID.String == item.DiscountId.Value)) &&
				((dbItem.StartDate.Status == pgtype.Null) ||
					(dbItem.StartDate.Status == pgtype.Present && dbItem.StartDate.Time == item.StartDate.AsTime())) {
				foundOrderItem++
			}
		}
	}
	return foundOrderItem
}

func countBillItem(billItems []entities.BillItem, billingItems []*pb.BillingItem, locationID string) int {
	foundBillItem := 0
	for _, item := range billingItems {
		for _, dbItem := range billItems {
			if item.ProductId == dbItem.ProductID.String &&
				dbItem.TaxCategory.String == item.TaxItem.TaxCategory.String() &&
				float32(dbItem.TaxPercentage.Int) == item.TaxItem.TaxPercentage &&
				dbItem.TaxID.String == item.TaxItem.TaxId &&
				IsEqualNumericAndFloat32(dbItem.TaxAmount, item.TaxItem.TaxAmount) &&
				float32(dbItem.ProductPricing.Int) == item.Price &&
				dbItem.BillSchedulePeriodID.Status == pgtype.Null &&
				dbItem.BillStatus.String == pb.BillingStatus_BILLING_STATUS_BILLED.String() &&
				dbItem.BillType.String == pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String() &&
				dbItem.BillFrom.Status == pgtype.Null &&
				dbItem.BillTo.Status == pgtype.Null &&
				IsEqualNumericAndFloat32(dbItem.FinalPrice, item.FinalPrice) &&
				dbItem.LocationID.String == locationID {
				foundBillItem++
			}
		}
	}
	return foundBillItem
}

func (s *suite) getOrder(ctx context.Context, orderID string) (*entities.Order, error) {
	order := &entities.Order{}
	orderFieldNames, orderFieldValues := order.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(orderFieldNames, ","),
		"public.order",
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, orderID)
	err := row.Scan(orderFieldValues...)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	return order, nil
}

func (s *suite) getOrderItems(ctx context.Context, orderID string) ([]entities.OrderItem, error) {
	var orderItems []entities.OrderItem
	orderItem := &entities.OrderItem{}
	orderItemFieldNames, orderItemFieldValues := orderItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(orderItemFieldNames, ","),
		orderItem.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(orderItemFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		orderItems = append(orderItems, *orderItem)
	}
	return orderItems, nil
}

func (s *suite) getOrderActionLogs(ctx context.Context, orderID string) ([]entities.OrderActionLog, error) {
	var orderActionLogs []entities.OrderActionLog
	orderActionLog := &entities.OrderActionLog{}
	orderActionLogFieldNames, orderActionLogFieldValues := orderActionLog.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(orderActionLogFieldNames, ","),
		orderActionLog.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(orderActionLogFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		orderActionLogs = append(orderActionLogs, *orderActionLog)
	}
	return orderActionLogs, nil
}

func (s *suite) getBillItems(ctx context.Context, orderID string) ([]entities.BillItem, error) {
	var billItems []entities.BillItem
	billItem := &entities.BillItem{}
	billItemFieldNames, billItemFieldValues := billItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		billItem.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(billItemFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, *billItem)
	}
	return billItems, nil
}

func (s *suite) ProductNotSupportCase(ctx context.Context, defaultPrepareDataSettings PrepareDataForCreatingOrderSettings) (req pb.CreateOrderRequest, err error) {
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		materialIDs []string
	)
	taxID,
		discountIDs,
		locationID,
		materialIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order"

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))
	startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	// for _, materialID := range materialIDs {
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

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	return
}

func (s *suite) NotExistProductPriceCase(ctx context.Context, defaultPrepareDataSettings PrepareDataForCreatingOrderSettings) (req pb.CreateOrderRequest, err error) {
	var (
		taxID       string
		discountIDs []string
		locationID  string
		userID      string
		materialIDs []string
	)
	taxID,
		discountIDs,
		locationID,
		materialIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order"

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))

	// for _, materialID := range materialIDs {
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
		Price:     0,
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
		Price:     15,
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
		Price:     20,
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

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	return
}
