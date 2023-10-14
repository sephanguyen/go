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

func (s *suite) aTaxValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.setDefaultTaxIfNotExist(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingTaxID, err := s.getNewTaxID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	uniqueID := idutil.ULIDNow()
	validRow1 := fmt.Sprintf(",import-tax-test-1-%s,10,1,0,0", uniqueID)
	validRow2 := fmt.Sprintf(",import-tax-test-2-%s,15,1,0,1", uniqueID)
	validRow3 := fmt.Sprintf("%s,import-tax-test-3-%s,20,1,0,0", existingTaxID, uniqueID)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(fmt.Sprintf(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
        %s
		%s
        %s`, validRow1, validRow2, validRow3)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aTaxValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.setDefaultTaxIfNotExist(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingTaxID, err := s.getNewTaxID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	uniqueID := idutil.ULIDNow()
	validRow1 := fmt.Sprintf(",import-tax-test-1-%s,10,1,0,0", uniqueID)
	validRow2 := fmt.Sprintf(",import-tax-test-2-%s,15,1,0,1", uniqueID)
	validRow3 := fmt.Sprintf("%s,import-tax-test-3-%s,20,1,0,0", existingTaxID, uniqueID)

	invalidEmptyRow1 := ",Tax 70,,,0,"
	invalidEmptyRow2 := "69,,13,1,,0"

	invalidValueRow1 := ",Tax 1,a,1,0,-1"
	invalidValueRow2 := "a,Tax 2,12,3,default,1"
	duplicateTrueDefaultFlag := ",Tax 1,15,1,1,0"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(fmt.Sprintf(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
        %s
        %s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(fmt.Sprintf(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
        %s
        %s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(fmt.Sprintf(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
        %s
        %s
        %s
        %s
        %s
        %s
		%s
        %s`, validRow1, validRow2, validRow3, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, duplicateTrueDefaultFlag)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, duplicateTrueDefaultFlag}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingTax(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportTax(contextWithToken(ctx), stepState.Request.(*pb.ImportTaxRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidTaxLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportTaxRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportTaxResponse)
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

func (s *suite) theValidTaxLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allTaxes, err := s.selectAllTaxes(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")

		taxID := rowSplit[0]
		name := rowSplit[1]

		taxPercentage, err := strconv.Atoi(strings.TrimSpace(rowSplit[2]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		taxCategory, err := strconv.Atoi(strings.TrimSpace(rowSplit[3]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		defaultFlag, err := strconv.ParseBool(rowSplit[4])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		isArchived, err := strconv.ParseBool(rowSplit[5])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allTaxes {
			if taxID == "" {
				if e.Name.String == name &&
					int(e.TaxPercentage.Int) == taxPercentage &&
					e.TaxCategory.Get() == convertToTaxCategoryString(taxCategory) &&
					e.IsArchived.Bool == isArchived &&
					e.DefaultFlag.Bool == defaultFlag &&
					e.CreatedAt.Time.Equal(e.UpdatedAt.Time) {
					found = true
					break
				}
			} else {
				taxId := strings.TrimSpace(taxID)
				if e.TaxID.String == taxId &&
					e.Name.String == name &&
					int(e.TaxPercentage.Int) == taxPercentage &&
					e.TaxCategory.Get() == convertToTaxCategoryString(taxCategory) &&
					e.IsArchived.Bool == isArchived && e.DefaultFlag.Bool == defaultFlag &&
					e.CreatedAt.Time.Before(e.UpdatedAt.Time) {
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

func (s *suite) theImportTaxTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	allTaxes, err := s.selectAllTaxes(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")

		taxID := rowSplit[0]
		name := rowSplit[1]

		taxPercentage, err := strconv.Atoi(strings.TrimSpace(rowSplit[2]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		taxCategory, err := strconv.Atoi(strings.TrimSpace(rowSplit[3]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		defaultFlag, err := strconv.ParseBool(rowSplit[4])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		isArchived, err := strconv.ParseBool(rowSplit[5])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allTaxes {
			if taxID == "" {
				if e.Name.String == name &&
					int(e.TaxPercentage.Int) == taxPercentage &&
					e.TaxCategory.Get() == convertToTaxCategoryString(taxCategory) &&
					e.IsArchived.Bool == isArchived &&
					e.DefaultFlag.Bool == defaultFlag &&
					e.CreatedAt.Time.Equal(e.UpdatedAt.Time) {
					found = true
					break
				}
			} else {
				taxId := strings.TrimSpace(taxID)
				if e.TaxID.String == taxId &&
					e.Name.String == name &&
					int(e.TaxPercentage.Int) == taxPercentage &&
					e.TaxCategory.Get() == convertToTaxCategoryString(taxCategory) &&
					e.IsArchived.Bool == isArchived && e.DefaultFlag.Bool == defaultFlag &&
					e.CreatedAt.Time.Before(e.UpdatedAt.Time) {
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

func (s *suite) aTaxInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportTaxRequest{}
	case "header only":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived`),
		}
	case "number of column is not equal 6":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,name,tax_percentage,tax_category,is_archived
      ,Tax 1,10,1,0`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
      ,10,1,0,0`),
		}
	case "wrong tax_id column name in header":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`wrong_header,name,tax_percentage,tax_category,default_flag,is_archived
      ,Tax 1,10,1,0,0`),
		}
	case "wrong name column name in header":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,wrong_header,tax_percentage,tax_category,default_flag,is_archived
      ,Tax 1,10,1,0,0`),
		}
	case "wrong tax_percentage column name in header":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,name,wrong_header,tax_category,default_flag,is_archived
      ,Tax 1,10,1,0,0`),
		}
	case "wrong tax_category column name in header":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,name,tax_percentage,wrong_header,default_flag,is_archived
      ,Tax 1,10,1,0,0`),
		}
	case "wrong default_flag column name in header":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,name,tax_percentage,tax_category,wrong_header,is_archived
      ,Tax 1,10,1,0,0`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportTaxRequest{
			Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,wrong_header
      ,Tax 1,10,1,0,0`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func convertToTaxCategoryString(i int) string {
	taxCategories := []string{pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(), pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String()}
	return taxCategories[i-1]
}

func (s *suite) selectAllTaxes(ctx context.Context) ([]*entities.Tax, error) {
	allEntities := []*entities.Tax{}
	stmt :=
		`
        SELECT
            tax_id,
            name,
            tax_percentage,
            tax_category,
            default_flag,
            is_archived,
            created_at,
            updated_at
        FROM
            tax
        `
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query tax")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.Tax{}
		err := rows.Scan(
			&e.TaxID,
			&e.Name,
			&e.TaxPercentage,
			&e.TaxCategory,
			&e.DefaultFlag,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan tax")
		}
		allEntities = append(allEntities, e)
	}

	return allEntities, nil
}

func (s *suite) insertTax(ctx context.Context, name string, defaultFlag bool) error {
	stmt :=
		`
		INSERT INTO tax(
			tax_id,
			name,
			tax_percentage,
			tax_category,
			default_flag,
			is_archived,
			created_at,
			updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())
        ON CONFLICT DO NOTHING
		`
	_, err := s.FatimaDBTrace.Exec(ctx, stmt, idutil.ULIDNow(), name, 10, "TAX_CATEGORY_INCLUSIVE", defaultFlag, false)
	if err != nil {
		return fmt.Errorf("cannot insert tax, err: %s", err)
	}

	return nil
}

func (s *suite) setDefaultTaxIfNotExist(ctx context.Context) error {
	uniqueID := idutil.ULIDNow()
	name := fmt.Sprintf("default-tax-%s", uniqueID)
	return s.insertTax(ctx, name, true)
}

func (s *suite) getNewTaxID(ctx context.Context) (string, error) {
	var taxID string

	uniqueID := idutil.ULIDNow()
	name := fmt.Sprintf("import-tax-test-%s", uniqueID)

	err := s.insertTax(ctx, name, false)
	if err != nil {
		return "", fmt.Errorf("cannot insert tax, err: %s", err)
	}

	stmt := `SELECT tax_id FROM tax WHERE name = $1`

	row := s.FatimaDBTrace.QueryRow(ctx, stmt, name)
	err = row.Scan(&taxID)
	if err != nil {
		return "", fmt.Errorf("cannot insert tax, err: %s", err)
	}

	return taxID, nil
}
