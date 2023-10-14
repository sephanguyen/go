package payment

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aProductSettingValidRequestPayloadWithCorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	productIDs, err := s.insertSomeProducts(ctx)
	if err != nil || len(productIDs) < 2 {
		return ctx, fmt.Errorf("error inserting mock products for product setting test, err: %s", err)
	}

	validRow1 := fmt.Sprintf("%s,false,false,false,false", productIDs[0])
	validRow2 := fmt.Sprintf("%s,false,false,false,false", productIDs[1])

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(fmt.Sprintf(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
		      %s
		      %s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	case "overwrite existing":
		overWrittenRow := fmt.Sprintf("%s,false,false,false,false", productIDs[0])
		updatedRow := fmt.Sprintf("%s,true,true,false,false", productIDs[0])

		err := s.insertProductSetting(ctx, productIDs[0])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(fmt.Sprintf(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
		      %s`, updatedRow)),
		}
		stepState.ValidCsvRows = []string{updatedRow}
		stepState.OverwrittenCsvRows = []string{overWrittenRow}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aProductSettingValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	productIDs, err := s.insertSomeProducts(ctx)
	if err != nil || len(productIDs) < 2 {
		return ctx, fmt.Errorf("error inserting mock products for product setting test, err: %s", err)
	}

	validRow1 := fmt.Sprintf("%s,false,false,false,false", productIDs[0])
	validRow2 := fmt.Sprintf("%s,false,false,false,false", productIDs[1])

	invalidEmptyRow1 := ",false,false,false,false"
	invalidEmptyRow2 := fmt.Sprintf("%s,,false,false,false", productIDs[0])

	invalidValueRow1 := "a,false,false,false,false"
	invalidValueRow2 := fmt.Sprintf("%s,a,false,false,false", productIDs[0])

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(fmt.Sprintf(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
		      %s
		      %s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(fmt.Sprintf(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
		      %s
		      %s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(fmt.Sprintf(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
		      %s
		      %s
		      %s
		      %s
		      %s
			  %s`, validRow1, validRow2, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductSetting(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportProductSetting(contextWithToken(ctx), stepState.Request.(*pb.ImportProductSettingRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidProductSettingLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductSetting, err := s.selectAllProductSetting(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")
		productID := strings.TrimSpace(rowSplit[0])

		isEnrollmentRequired, err := strconv.ParseBool(strings.TrimSpace(rowSplit[1]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allProductSetting {
			if e.ProductID.String == productID && e.IsEnrollmentRequired.Bool == isEnrollmentRequired {
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

func (s *suite) theImportProductSettingTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	allProductSetting, err := s.selectAllProductSetting(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")
		productID := rowSplit[0]

		isEnrollmentRequired, err := strconv.ParseBool(strings.TrimSpace(rowSplit[1]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allProductSetting {
			if e.ProductID.String == productID && e.IsEnrollmentRequired.Bool == isEnrollmentRequired {
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

func (s *suite) theInvalidProductSettingLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportProductSettingRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportProductSettingResponse)
	for _, row := range stepState.InvalidCsvRows {
		found := false
		if resp != nil {
			for _, e := range resp.Errors {
				if strings.TrimSpace(reqSplit[e.RowNumber-1]) == row {
					found = true
					break
				}
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid line is not returned in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aProductSettingInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductSettingRequest{}
	case "header only":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee`),
		}
	case "number of column is not equal 5":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default
      1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
      1,true`),
		}
	case "incorrect product_id column name in header":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(`incorrect_product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
      1,true,true,false,false`),
		}
	case "incorrect is_enrollment_required column name in header":
		stepState.Request = &pb.ImportProductSettingRequest{
			Payload: []byte(`package_type,incorrect_is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
      1,true,true,false,false`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllProductSetting(ctx context.Context) ([]*entities.ProductSetting, error) {
	var allEntities []*entities.ProductSetting
	stmt := `SELECT
                product_id,
                is_enrollment_required,
				is_pausable,
				is_added_to_enrollment_by_default,
				is_operation_fee
            FROM
                product_setting`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query product_setting")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.ProductSetting{}
		err := rows.Scan(
			&e.ProductID,
			&e.IsEnrollmentRequired,
			&e.IsPausable,
			&e.IsAddedToEnrollmentByDefault,
			&e.IsOperationFee,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan product setting")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertProductSetting(ctx context.Context, productID string) error {
	insertStmt := `INSERT INTO product_setting(
                product_id,
                is_enrollment_required,
				is_pausable,
				is_added_to_enrollment_by_default,
				is_operation_fee)
            VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`
	_, err := s.FatimaDBTrace.Exec(ctx, insertStmt, productID, false, true, false, false)
	if err != nil {
		return fmt.Errorf("cannot insert product setting, err: %s", err)
	}
	return nil
}

func (s *suite) insertSomeProducts(ctx context.Context) ([]string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
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

	var productIDs []string
	for i := 0; i < 2; i++ {
		var arg AddProductParams
		var productID string
		randomStr := idutil.ULIDNow()
		arg.ProductID = randomStr
		arg.Name = fmt.Sprintf("product-%v", randomStr)
		arg.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
		arg.AvailableFrom = time.Now()
		arg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		arg.DisableProRatingFlag = false
		arg.IsArchived = false

		stmt := `INSERT INTO product(
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
                    created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), now())
                RETURNING product_id`
		row := s.FatimaDBTrace.QueryRow(ctx, stmt,
			arg.ProductID,
			arg.Name,
			arg.ProductType,
			arg.TaxID,
			arg.AvailableFrom,
			arg.AvailableUtil,
			arg.Remarks,
			arg.CustomBillingPeriod,
			arg.BillingScheduleID,
			arg.DisableProRatingFlag,
			arg.IsArchived)

		err := row.Scan(&productID)
		if err != nil {
			return productIDs, fmt.Errorf("cannot insert product, err: %s", err)
		}

		productIDs = append(productIDs, productID)
	}
	return productIDs, nil
}
