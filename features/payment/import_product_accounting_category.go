package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aProductAccountingCategoryValidRequestPayloadWithCorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.insertSomePackages(ctx)
	if err != nil {
		fmt.Printf("error when insert packages %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}

	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertSomeAccountingCategories(ctx)
	if err != nil {
		fmt.Printf("error when insert accounting categories %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}

	existingAccountingCategories, err := s.selectAllAccountingCategories(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "all valid rows":
		validRow1 := fmt.Sprintf("%s,%s", existingPackages[len(existingPackages)-1].PackageID.String, existingAccountingCategories[len(existingAccountingCategories)-1].AccountingCategoryID.String)
		validRow2 := fmt.Sprintf("%s,%s", existingPackages[len(existingPackages)-2].PackageID.String, existingAccountingCategories[len(existingAccountingCategories)-2].AccountingCategoryID.String)

		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(fmt.Sprintf(`product_id,accounting_category_id
          %s
          %s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}

	case "overwrite existing":
		var overwrittenRow string

		allProductAccountingCategory, err := s.selectAllProductAccountingCategory(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(allProductAccountingCategory) == 0 {
			err = s.insertSomeProductAssociatedAccountingCategory(
				ctx,
				existingPackages[len(existingPackages)-1].PackageID.String,
				existingAccountingCategories[len(existingAccountingCategories)-1].AccountingCategoryID.String)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			allProductAccountingCategory, err = s.selectAllProductAccountingCategory(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			overwrittenRow = fmt.Sprintf("%s,%s", existingPackages[len(existingPackages)-1].PackageID.String, existingAccountingCategories[len(existingAccountingCategories)-1].AccountingCategoryID.String)
		} else {
			overwrittenRow = fmt.Sprintf("%s,%s", allProductAccountingCategory[0].ProductID.String, allProductAccountingCategory[0].AccountingCategoryID.String)
		}

		err = s.insertSomeAccountingCategories(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		newAccountingCategories, err := s.selectAllAccountingCategories(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		err = s.insertSomeProductAssociatedAccountingCategory(
			ctx,
			allProductAccountingCategory[0].ProductID.String,
			newAccountingCategories[len(newAccountingCategories)-1].AccountingCategoryID.String)

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		validRow := fmt.Sprintf("%s,%s", allProductAccountingCategory[0].ProductID.String, newAccountingCategories[len(newAccountingCategories)-1].AccountingCategoryID.String)

		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(fmt.Sprintf(`product_id,accounting_category_id
          %s`, validRow)),
		}

		stepState.ValidCsvRows = []string{validRow}
		stepState.OverwrittenCsvRows = []string{overwrittenRow}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aProductAccountingCategoryValidRequestPayloadWithIncorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.insertSomePackages(ctx)
	if err != nil {
		fmt.Printf("error when insert packages %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}

	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertSomeAccountingCategories(ctx)
	if err != nil {
		fmt.Printf("error when insert accounting categories %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}

	existingAccountingCategories, err := s.selectAllAccountingCategories(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	invalidEmptyRow1 := ",3"
	invalidEmptyRow2 := "4,"

	invalidValueRow1 := "a,5"
	invalidValueRow2 := "6,b"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(fmt.Sprintf(`product_id,accounting_category_id
          %s
          %s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}

	case "invalid value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(fmt.Sprintf(`product_id,accounting_category_id
          %s
          %s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}

	case "valid and invalid rows":
		validRow1 := fmt.Sprintf("%s,%s", existingPackages[0].PackageID.String, existingAccountingCategories[0].AccountingCategoryID.String)
		validRow2 := fmt.Sprintf("%s,%s", existingPackages[1].PackageID.String, existingAccountingCategories[1].AccountingCategoryID.String)

		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(fmt.Sprintf(`product_id,accounting_category_id
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

func (s *suite) theValidProductAccountingCategoryLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductAccountingCategory, err := s.selectAllProductAccountingCategory(ctx)
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
		accountingCategoryID := strings.TrimSpace(rowSplit[1])

		for _, e := range allProductAccountingCategory {
			if e.ProductID.String == productID && e.AccountingCategoryID.String == accountingCategoryID && e.CreatedAt.Time.Before(time.Now()) {
				found = true
				break
			}
		}

		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	for _, row := range stepState.OverwrittenCsvRows {
		found := false
		rowSplit := strings.Split(row, ",")

		productID := strings.TrimSpace(rowSplit[0])
		accountingCategoryID := strings.TrimSpace(rowSplit[1])

		for _, e := range allProductAccountingCategory {
			if e.ProductID.String == productID && e.AccountingCategoryID.String == accountingCategoryID {
				found = true
				break
			}
		}

		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to overwrite existing association")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportProductAccountingCategoryTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductAccountingCategory, err := s.selectAllProductAccountingCategory(ctx)
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
		accountingCategoryID := strings.TrimSpace(rowSplit[1])

		for _, e := range allProductAccountingCategory {
			if e.ProductID.String == productID && e.AccountingCategoryID.String == accountingCategoryID && e.CreatedAt.Time.Before(time.Now()) {
				found = true
				break
			}
		}

		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to rollback valid csv row")
		}
	}

	for _, row := range stepState.OverwrittenCsvRows {
		found := false
		rowSplit := strings.Split(row, ",")

		productID := strings.TrimSpace(rowSplit[0])
		accountingCategoryID := strings.TrimSpace(rowSplit[1])

		for _, e := range allProductAccountingCategory {
			if e.ProductID.String == productID && e.AccountingCategoryID.String == accountingCategoryID {
				found = true
				break
			}
		}

		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to rollback valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidProductAccountingCategoryLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportProductAssociatedDataRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportProductAssociatedDataResponse)
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

func (s *suite) aProductAccountingCategoryInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{}
	case "header only":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload:                   []byte(`product_id,accounting_category_id`),
		}
	case "number of column is not equal 2 product_id only":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(`product_id
      1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(`product_id,accounting_category_id
      1`),
		}
	case "wrong product_id column name in csv header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(`wrong_header,accounting_category_id
      1,1`),
		}
	case "wrong accounting_category_id column name in header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			Payload: []byte(`product_id,wrong_header
      1,1`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductAccountingCategory(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportProductAssociatedData(contextWithToken(ctx), stepState.Request.(*pb.ImportProductAssociatedDataRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllProductAccountingCategory(ctx context.Context) ([]*entities.ProductAccountingCategory, error) {
	allEntities := []*entities.ProductAccountingCategory{}
	stmt := `SELECT
                product_id,
                accounting_category_id,
                created_at
            FROM
                product_accounting_category`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query product_accounting_category")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.ProductAccountingCategory{}
		err := rows.Scan(
			&e.ProductID,
			&e.AccountingCategoryID,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan product_accounting_category")
		}
		allEntities = append(allEntities, e)
	}

	return allEntities, nil
}

func (s *suite) insertSomeProductAssociatedAccountingCategory(ctx context.Context, productID, accountingCategoryID string) error {
	stmt := `INSERT INTO product_accounting_category(
                product_id,
                accounting_category_id,
                created_at)
            VALUES ($1, $2, now())`
	_, err := s.FatimaDBTrace.Exec(ctx, stmt, productID, accountingCategoryID)
	if err != nil {
		return fmt.Errorf("cannot insert product associated accounting category, err: %s", err)
	}

	return nil
}
