package payment

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aBillingRatioValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.insertSomeBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertSomeBillingRatios(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingBillingSchedulePeriods, err := s.selectAllBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingBillingRatios, err := s.selectAllBillingRatios(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf(",2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,%s,1,2,0", existingBillingSchedulePeriods[0].BillingSchedulePeriodID.String)
	validRow2 := fmt.Sprintf(",2021-12-07,2021-12-08,%s,3,3,1", existingBillingSchedulePeriods[1].BillingSchedulePeriodID.String)
	validRow3 := fmt.Sprintf("%s,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,%s,4,4,0", existingBillingRatios[0].BillingRatioID.String, existingBillingSchedulePeriods[2].BillingSchedulePeriodID.String)
	stepState.ValidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(fmt.Sprintf(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
        %s
        %s
        %s`, validRow1, validRow2, validRow3)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aBillingRatioValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	existingBillingSchedulePeriods, err := s.selectAllBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf(",2022-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,%s,1,2,0", existingBillingSchedulePeriods[0].BillingSchedulePeriodID.String)
	validRow2 := fmt.Sprintf(",2022-12-07,2021-12-08,%s,3,3,1", existingBillingSchedulePeriods[1].BillingSchedulePeriodID.String)

	invalidEmptyRow1 := ",,,3,1,2,0"
	invalidEmptyRow2 := ",2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,,,,"

	invalidValueRow1 := ",2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,a,1,2,0"
	invalidValueRow2 := "a,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(fmt.Sprintf(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
        %s
        %s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(fmt.Sprintf(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
        %s
        %s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(fmt.Sprintf(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
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

func (s *suite) theImportBillingRatioTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	allBillingRatios, err := s.selectAllBillingRatios(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")

		id := rowSplit[0]
		startDate, err := parseToDate(rowSplit[1])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		endDate, err := parseToDate(rowSplit[2])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		billingSchedulePeriodID := strings.TrimSpace(rowSplit[3])
		billingRatioNumerator, _ := strconv.Atoi(strings.TrimSpace(rowSplit[4]))
		billingRatioDenominator, _ := strconv.Atoi(strings.TrimSpace(rowSplit[5]))
		isArchived, _ := strconv.ParseBool(rowSplit[6])

		for _, e := range allBillingRatios {
			if id == "" {
				if e.StartDate.Time.Equal(startDate) && e.EndDate.Time.Equal(endDate) && e.BillingSchedulePeriodID.String == billingSchedulePeriodID && int(e.BillingRatioNumerator.Int) == billingRatioNumerator && int(e.BillingRatioDenominator.Int) == billingRatioDenominator && e.IsArchived.Bool == isArchived && e.CreatedAt.Time.Equal(e.UpdatedAt.Time) {
					found = true
					break
				}
			} else {
				id := strings.TrimSpace(id)
				if e.BillingRatioID.String == id && e.StartDate.Time.Equal(startDate) && e.EndDate.Time.Equal(endDate) && e.BillingSchedulePeriodID.String == billingSchedulePeriodID && int(e.BillingRatioNumerator.Int) == billingRatioNumerator && int(e.BillingRatioDenominator.Int) == billingRatioDenominator && e.IsArchived.Bool == isArchived && e.CreatedAt.Time.Before(e.UpdatedAt.Time) {
					found = true
					break
				}
			}
		}

		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingBillingRatio(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportBillingRatio(contextWithToken(ctx), stepState.Request.(*pb.ImportBillingRatioRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidBillingRatioLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportBillingRatioRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportBillingRatioResponse)
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

func (s *suite) theValidBillingRatioLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allBillingRatios, err := s.selectAllBillingRatios(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")

		id := rowSplit[0]
		startDate, err := parseToDate(rowSplit[1])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		endDate, err := parseToDate(rowSplit[2])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		billingSchedulePeriodID := strings.TrimSpace(rowSplit[3])
		billingRatioNumerator, _ := strconv.Atoi(strings.TrimSpace(rowSplit[4]))
		billingRatioDenominator, _ := strconv.Atoi(strings.TrimSpace(rowSplit[5]))
		isArchived, _ := strconv.ParseBool(rowSplit[6])

		for _, e := range allBillingRatios {
			if id == "" {
				if e.StartDate.Time.Equal(startDate) && e.EndDate.Time.Equal(endDate) && e.BillingSchedulePeriodID.String == billingSchedulePeriodID && int(e.BillingRatioNumerator.Int) == billingRatioNumerator && int(e.BillingRatioDenominator.Int) == billingRatioDenominator && e.IsArchived.Bool == isArchived && e.CreatedAt.Time.Equal(e.UpdatedAt.Time) {
					found = true
					break
				}
			} else {
				id := strings.TrimSpace(id)
				if e.BillingRatioID.String == id && e.StartDate.Time.Equal(startDate) && e.EndDate.Time.Equal(endDate) && e.BillingSchedulePeriodID.String == billingSchedulePeriodID && int(e.BillingRatioNumerator.Int) == billingRatioNumerator && int(e.BillingRatioDenominator.Int) == billingRatioDenominator && e.IsArchived.Bool == isArchived && e.CreatedAt.Time.Before(e.UpdatedAt.Time) {
					found = true
					break
				}
			}
		}

		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aBillingRatioInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportBillingRatioRequest{}
	case "header only":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived`),
		}
	case "number of column is not equal 7":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator
      1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3`),
		}
	case "wrong billing_ratio_id column name in header":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`wrong_header,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
		}
	case "wrong start_date column name in header":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,wrong_header,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
		}
	case "wrong end_date column name in header":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,wrong_header,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
		}
	case "wrong billing_schedule_period_id column name in header":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,end_date,wrong_header,billing_ratio_numerator,billing_ratio_denominator,is_archived
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
		}
	case "wrong billing_ratio_numerator column name in header":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,wrong_header,billing_ratio_denominator,is_archived
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
		}
	case "wrong billing_ratio_denominator column name in header":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,wrong_header,is_archived
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportBillingRatioRequest{
			Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,wrong_header
      ,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllBillingRatios(ctx context.Context) ([]*entities.BillingRatio, error) {
	allEntities := []*entities.BillingRatio{}
	stmt :=
		`
    SELECT 
			billing_ratio_id,
			start_date,
			end_date,
			billing_schedule_period_id,
			billing_ratio_numerator,
			billing_ratio_denominator,
			is_archived,
			created_at,
			updated_at
    FROM
      billing_ratio
    `
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query billing_ratio")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entities.BillingRatio{}
		err := rows.Scan(
			&e.BillingRatioID,
			&e.StartDate,
			&e.EndDate,
			&e.BillingSchedulePeriodID,
			&e.BillingRatioNumerator,
			&e.BillingRatioDenominator,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan billing_ratio")
		}
		allEntities = append(allEntities, e)
	}

	return allEntities, nil
}

func (s *suite) insertSomeBillingRatios(ctx context.Context) error {
	existingBillingSchedulePeriods, err := s.selectAllBillingSchedulePeriods(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		billingSchedulePeriodID := existingBillingSchedulePeriods[i].BillingSchedulePeriodID.String
		billingRatioNumerator := 1
		billingRatioDenominator := 2
		isArchived := i%2 == 0
		stmt := `INSERT INTO billing_ratio
		(start_date, end_date, billing_ratio_id, billing_schedule_period_id, billing_ratio_numerator, billing_ratio_denominator, is_archived, created_at, updated_at)
		VALUES (now(), now(), $1, $2, $3, $4, $5, now(), now())`
		_, err = s.FatimaDBTrace.Exec(ctx, stmt, idutil.ULIDNow(), billingSchedulePeriodID, billingRatioNumerator, billingRatioDenominator, isArchived)
		if err != nil {
			return fmt.Errorf("cannot insert billing ratio, err: %s", err)
		}
	}
	return nil
}
