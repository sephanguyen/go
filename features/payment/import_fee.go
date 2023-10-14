package payment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) theInvalidFeeLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportProductRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportProductResponse)
	for _, row := range stepState.InvalidCsvRows {
		found := false
		for _, e := range resp.Errors {
			if strings.TrimSpace(reqSplit[e.RowNumber-1]) == row {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid line is not returned in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidFeeLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allFees, err := s.selectAllFees(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	allProducts, err := s.selectAllProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, row := range stepState.ValidCsvRows {
		var product entities.Product
		var fee entities.Fee
		values := strings.Split(row, ",")
		product, fee, err = getProductAndFeeFromCsv(values)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		findingProduct := foundProducts(product, allProducts)
		if findingProduct.ProductID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found product in list")
		}
		fee.FeeID = findingProduct.ProductID
		findingFee := foundFees(fee, allFees)
		if findingFee.FeeID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found fee in list")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportFeeTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allFees, err := s.selectAllFees(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	allProducts, err := s.selectAllProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		var product entities.Product
		var fee entities.Fee
		values := strings.Split(row, ",")
		product, fee, err = getProductAndFeeFromCsv(values)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		findingProduct := foundProducts(product, allProducts)
		if findingProduct.ProductID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found product in list")
		}
		fee.FeeID = findingProduct.ProductID
		findingFee := foundFees(fee, allFees)
		if findingFee.FeeID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found fee in list")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingFee(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	client := pb.NewImportMasterDataServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = client.ImportProduct(
		contextWithToken(ctx),
		stepState.Request.(*pb.ImportProductRequest),
	)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anFeeValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeFees(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(
		",Cat %s,1,,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,,1,Remarks %s,1,0",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	validRow2 := fmt.Sprintf(
		"01GXN3MJEG7PEMYKNZ3GBEBRDM,Cat %s,2,,,,2021-12-07,2021-12-08,2021-12-09,,1,Remarks %s,1,1",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "all valid rows":

		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload: []byte(fmt.Sprintf(`fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
			%s
			%s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anFeeValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeFees(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingFees, err := s.selectAllFees(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(
		",Cat %s,1,,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,,1,Remarks %s,1,0",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	validRow2 := fmt.Sprintf(
		",Cat %s,2,,,,2021-12-07,2021-12-08,2021-12-09,,1,Remarks %s,1,0",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	validRow3 := fmt.Sprintf(
		",Cat %s,2,,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,,1,,1,0",
		idutil.ULIDNow(),
	)
	validRow4 := fmt.Sprintf(
		"%s,Cat %s,2,,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,,1,Remarks %s,1,0",
		existingFees[0].FeeID.String,
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidEmptyRow1 := fmt.Sprintf(
		",Cat %s,2,,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks %s,,0",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidEmptyRow2 := fmt.Sprintf(
		"%s,Cat %s,2,,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks %s,,0",
		existingFees[1].FeeID.String,
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidValueRow1 := fmt.Sprintf(
		",Cat %s,2,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks %s,Archived,1",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidValueRow2 := fmt.Sprintf(
		"%s,Cat %s,2,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks %s,Archived,1",
		existingFees[2].FeeID.String,
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "empty value row":

		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload: []byte(fmt.Sprintf(`fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":

		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload: []byte(fmt.Sprintf(`fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":

		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload: []byte(fmt.Sprintf(`fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, validRow1, validRow2, validRow3, validRow4, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anFeeInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		}
	case "header only":
		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload:     []byte(`fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique`),
		}
	case "number of column is not equal 12":
		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload: []byte(`fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_unique
			1,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1,0
			2,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,0
			3,Cat 3,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3,0`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload: []byte(`fee_id,name,billing_schedule_id,product_tag,product_partner_id,start_date,end_date,billing_date,remarks,is_archived,is_unique
			1,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1
			2,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2
			3,Cat 3,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3`),
		}
	case "wrong fee_id column name in header":
		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
			Payload: []byte(`Number,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
			1,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1,0,0
			2,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,0,0
			3,Cat 3,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3,0,0`),
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllFees(ctx context.Context) ([]*entities.Fee, error) {
	allEntities := []*entities.Fee{}
	stmt :=
		`
		SELECT
			pm.fee_id,
			pm.fee_type,
			pm.resource_path
		FROM fee pm
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query fee")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.Fee{}
		err := rows.Scan(
			&e.FeeID,
			&e.FeeType,
			&e.ResourcePath,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan fee")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertSomeFees(ctx context.Context) error {
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

	type AddFeeParams struct {
		FeeID   string `json:"fee_id"`
		FeeType string `json:"fee_type"`
	}

	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var feeArg AddFeeParams
		randomStr := idutil.ULIDNow()
		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("fee-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_FEE.String()
		productArg.AvailableFrom = time.Now()
		productArg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
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
			return fmt.Errorf("cannot insert product, err: %s", err)
		}
		queryInsertFee := `INSERT INTO fee (fee_id, fee_type) VALUES ($1, $2)
		`
		feeArg.FeeType = pb.FeeType_FEE_TYPE_ONE_TIME.String()
		_, err = s.FatimaDBTrace.Exec(ctx, queryInsertFee,
			feeArg.FeeID,
			feeArg.FeeType,
		)
		if err != nil {
			return fmt.Errorf("cannot insert fee, err: %s", err)
		}

		productSetting := entities.ProductSetting{
			ProductID:                    pgtype.Text{String: productArg.ProductID, Status: pgtype.Present},
			IsPausable:                   pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsEnrollmentRequired:         pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsOperationFee:               pgtype.Bool{Bool: false, Status: pgtype.Present},
		}
		err = mockdata.InsertProductSetting(ctx, s.FatimaDBTrace, productSetting)
		if err != nil {
			return fmt.Errorf("cannot insert default product_setting for fee product, err: %s", err)
		}
	}

	return nil
}

func getProductAndFeeFromCsv(line []string) (product entities.Product, fee entities.Fee, err error) {
	const (
		FeeID = iota
		Name
		FeeType
		TaxID
		ProductTag
		ProductPartnerID
		AvailableFrom
		AvailableUntil
		CustomBillingPeriod
		BillingScheduleID
		DisableProRatingFlag
		Remarks
		IsArchived
	)

	if err = multierr.Combine(
		utils.StringToFormatString("fee_id", line[FeeID], true, product.ProductID.Set),
		utils.StringToFormatString("name", line[Name], false, product.Name.Set),
		product.ProductType.Set(pb.ProductType_PRODUCT_TYPE_FEE),
		utils.StringToFormatString("tax_id", line[TaxID], true, product.TaxID.Set),
		utils.StringToFormatString("product_tag", line[ProductTag], true, product.ProductTag.Set),
		utils.StringToFormatString("product_partner_id", line[ProductPartnerID], true, product.ProductPartnerID.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, product.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, product.AvailableUntil.Set),
		utils.StringToDate("custom_billing_period", line[CustomBillingPeriod], true, product.CustomBillingPeriod.Set),
		utils.StringToFormatString("billing_schedule_id", line[BillingScheduleID], true, product.BillingScheduleID.Set),
		utils.StringToBool("disable_pro_rating_flag", line[DisableProRatingFlag], true, product.DisableProRatingFlag.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, product.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], true, product.IsArchived.Set),
	); err != nil {
		return product, fee, err
	}

	if err = multierr.Combine(
		utils.StringToFormatString("ID", line[FeeID], true, fee.FeeID.Set),
		utils.StringToFeeType("FeeType", line[FeeType], fee.FeeType.Set),
	); err != nil {
		return product, fee, err
	}

	return product, fee, nil
}

func foundFees(feeNeedFinding entities.Fee, feeList []*entities.Fee) (finding entities.Fee) {
	for i, fee := range feeList {
		if feeNeedFinding.FeeID == fee.FeeID {
			finding = *feeList[i]
			break
		}
	}
	return finding
}
