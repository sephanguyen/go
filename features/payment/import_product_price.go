package payment

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

func (s *suite) theInvalidProductPriceLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportProductPriceRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportProductPriceResponse)
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

func (s *suite) theValidProductPriceLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductPrices, err := s.selectAllProductPrices(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we should use map for allProductPrices but it leads to some more code and not many items in
	// stepState.ValidCsvRows and allProductPrices, so we can do like below to make it simple
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		productID := rowSplit[0]
		billingSchedulePeriodID := "" // If null, then e.BillingSchedulePeriodID.String = ""
		if rowSplit[1] != "" {
			billingSchedulePeriodID = rowSplit[1]
		}
		quantity := int64(0)
		if rowSplit[2] != "" {
			quantity, err = strconv.ParseInt(rowSplit[2], 10, 64)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
		price, err := strconv.ParseFloat(rowSplit[3], 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		priceNumeric := pgtype.Numeric{}
		if err = priceNumeric.Set(price); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		priceType := rowSplit[4]

		found := false
		for _, e := range allProductPrices {
			if e.ProductID.String == productID &&
				e.BillingSchedulePeriodID.String == billingSchedulePeriodID &&
				e.Quantity.Int == int32(quantity) &&
				isNumericEqual(e.Price, priceNumeric) &&
				e.PriceType.String == priceType {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductPrice(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportProductPrice(contextWithToken(ctx), stepState.Request.(*pb.ImportProductPriceRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anProductPriceValidRequestPayloadWithCorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if err := s.insertSomePackages(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := s.insertSomeBillingSchedulePeriods(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	products, err := s.selectProducts(ctx, 4)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	billingSchedulePeriod, err := s.selectBillingSchedulePeriods(ctx, 2)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf("%s,%s,2,8,DEFAULT_PRICE", products[0].ProductID.String, billingSchedulePeriod[0].BillingSchedulePeriodID.String)
	validRow2 := fmt.Sprintf("%s,%s,2,12.25,DEFAULT_PRICE", products[0].ProductID.String, billingSchedulePeriod[1].BillingSchedulePeriodID.String)
	validRow3 := fmt.Sprintf("%s,%s,2,8,ENROLLED_PRICE", products[0].ProductID.String, billingSchedulePeriod[0].BillingSchedulePeriodID.String)
	validRow4 := fmt.Sprintf("%s,%s,2,12.25,ENROLLED_PRICE", products[0].ProductID.String, billingSchedulePeriod[1].BillingSchedulePeriodID.String)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(fmt.Sprintf(`product_id,billing_schedule_period_id,quantity,price,price_type
			%s
			%s
			%s
			%s`, validRow1, validRow2, validRow3, validRow4)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anProductPriceValidRequestPayloadWithIncorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if err := s.insertSomePackages(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := s.insertSomeBillingSchedulePeriods(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	products, err := s.selectProducts(ctx, 4)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	billingSchedulePeriod, err := s.selectBillingSchedulePeriods(ctx, 2)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf("%s,%s,2,7,DEFAULT_PRICE", products[0].ProductID.String, billingSchedulePeriod[0].BillingSchedulePeriodID.String)
	validRow2 := fmt.Sprintf("%s,%s,3,12.25,DEFAULT_PRICE", products[0].ProductID.String, billingSchedulePeriod[1].BillingSchedulePeriodID.String)
	validRow3 := fmt.Sprintf("%s,,2,7,DEFAULT_PRICE", products[1].ProductID.String)
	validRow4 := fmt.Sprintf("%s,,,7,DEFAULT_PRICE", products[2].ProductID.String)
	validRow5 := fmt.Sprintf("%s,,,7,DEFAULT_PRICE", products[3].ProductID.String)
	invalidEmptyRow1 := fmt.Sprintf("%s,%s,2,,DEFAULT_PRICE", products[0].ProductID.String, billingSchedulePeriod[0].BillingSchedulePeriodID.String)
	invalidEmptyRow2 := fmt.Sprintf(",%s,2,7,DEFAULT_PRICE", billingSchedulePeriod[0].BillingSchedulePeriodID.String)
	invalidValueRow1 := fmt.Sprintf("%s,%s,3,12..25,DEFAULT_PRICE", products[0].ProductID.String, billingSchedulePeriod[1].BillingSchedulePeriodID.String)
	invalidValueRow2 := fmt.Sprintf("%s,%s,a,12.25,DEFAULT_PRICE", products[0].ProductID.String, billingSchedulePeriod[1].BillingSchedulePeriodID.String)
	invalidValueRow3 := fmt.Sprintf("%s,0,3,12.25,DEFAULT_PRICE", products[0].ProductID.String)                            // Not existed billing schedule period id
	invalidValueRow4 := fmt.Sprintf("0,%s,3,12.25,DEFAULT_PRICE", billingSchedulePeriod[1].BillingSchedulePeriodID.String) // Not existed product id

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(fmt.Sprintf(`product_id,billing_schedule_period_id,quantity,price,price_type
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(fmt.Sprintf(`product_id,billing_schedule_period_id,quantity,price,price_type
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(fmt.Sprintf(`product_id,billing_schedule_period_id,quantity,price,price_type
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`,
				validRow1,
				validRow2,
				validRow3,
				validRow4,
				validRow5,
				invalidEmptyRow1,
				invalidEmptyRow2,
				invalidValueRow1,
				invalidValueRow2,
				invalidValueRow3,
				invalidValueRow4,
			)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anProductPriceInvalidRequestPayloadWith(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductPriceRequest{}
	case "header only":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type`),
		}
	case "number of column is not equal 5":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,quantity,price,price_type
			1,3,12.25,DEFAULT_PRICE`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
			1,1,12.25,DEFAULT_PRICE`),
		}
	case "wrong product_id column name in header":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`some_id,billing_schedule_period_id,quantity,price,price_type
			1,1,3,12.25,DEFAULT_PRICE`),
		}
	case "wrong billing_schedule_period_id column name in header":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,billing_schedule_id,quantity,price,price_type
			1,1,3,12.25,DEFAULT_PRICE`),
		}
	case "wrong quantity column name in header":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,billing_schedule_period_id,number_items,price,price_type
			1,1,3,12.25,DEFAULT_PRICE`),
		}
	case "wrong price column name in header":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,billing_schedule_period_id,quantity,item_price,price_type
			1,1,3,12.25,DEFAULT_PRICE`),
		}
	case "wrong price type column name in header":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,wrong_price_type
			1,1,3,12.25,DEFAULT_PRICE`),
		}
	case "missing default_price value by product id":
		stepState.Request = &pb.ImportProductPriceRequest{
			Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,wrong_price_type
			1,1,3,12.25,ENROLLED_PRICE`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportProductPriceTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	allProductPrices, err := s.selectAllProductPrices(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		productID := rowSplit[0]
		billingSchedulePeriodID := "" // If null, then e.BillingSchedulePeriodID.String = ""
		if rowSplit[1] != "" {
			billingSchedulePeriodID = rowSplit[1]
		}
		quantity := int64(0)
		if rowSplit[2] != "" {
			quantity, err = strconv.ParseInt(rowSplit[2], 10, 64)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
		price, err := strconv.ParseFloat(rowSplit[3], 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		priceNumeric := pgtype.Numeric{}
		if err = priceNumeric.Set(price); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		found := false
		for _, e := range allProductPrices {
			if e.ProductID.String == productID &&
				e.BillingSchedulePeriodID.String == billingSchedulePeriodID &&
				e.Quantity.Int == int32(quantity) &&
				isNumericEqual(e.Price, priceNumeric) {
				found = true
				break
			}
		}
		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to rollback valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllProductPrices(ctx context.Context) ([]*entities.ProductPrice, error) {
	allEntities := []*entities.ProductPrice{}
	stmt :=
		`
		SELECT 
			product_id,
			billing_schedule_period_id,
			quantity,
			price,
			price_type
		FROM
			product_price
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query product_price")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.ProductPrice{}
		err := rows.Scan(
			&e.ProductID,
			&e.BillingSchedulePeriodID,
			&e.Quantity,
			&e.Price,
			&e.PriceType,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan product_price")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) selectProducts(ctx context.Context, nItems int) ([]*entities.Product, error) {
	var allEntities []*entities.Product
	getProductsStm :=
		`SELECT
			product_id, 
			name, 
			product_type, 
			tax_id, 
			available_from, 
			available_until, 
			remarks, 
			custom_billing_period, 
			billing_schedule_id, 
			disable_pro_rating_flag, 
			is_archived, 
			updated_at, 
			created_at 
		FROM
			product
		LIMIT
			$1
`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		getProductsStm,
		nItems,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query products")
	}
	defer rows.Close()
	for rows.Next() {
		var i entities.Product
		if err := rows.Scan(
			&i.ProductID,
			&i.Name,
			&i.ProductType,
			&i.TaxID,
			&i.AvailableFrom,
			&i.AvailableUntil,
			&i.Remarks,
			&i.CustomBillingPeriod,
			&i.BillingScheduleID,
			&i.DisableProRatingFlag,
			&i.IsArchived,
			&i.UpdatedAt,
			&i.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan Product")
		}
		allEntities = append(allEntities, &i)
	}
	return allEntities, nil
}

func (s *suite) selectBillingSchedulePeriods(ctx context.Context, nItems int) ([]*entities.BillingSchedulePeriod, error) {
	allEntities := []*entities.BillingSchedulePeriod{}
	stmt :=
		`
		SELECT
			billing_schedule_period_id,
			name,
			billing_schedule_id,
			start_date,
			end_date,
			billing_date,
			remarks,
			is_archived
		FROM
			billing_schedule_period
		LIMIT $1
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
		nItems,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query billing_schedule_period")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.BillingSchedulePeriod{}
		err := rows.Scan(
			&e.BillingSchedulePeriodID,
			&e.Name,
			&e.BillingScheduleID,
			&e.StartDate,
			&e.EndDate,
			&e.BillingDate,
			&e.Remarks,
			&e.IsArchived,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan billing schedule period")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}
