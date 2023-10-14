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

func (s *suite) anProductDiscountValidRequestPayloadWithCorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomePackages(ctx)
	if err != nil {
		fmt.Printf("error when insert package %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}
	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertSomeDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingDiscounts, err := s.selectAllDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	headerTitles := []string{
		"product_id",
		"discount_id",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf("%s,%s", existingPackages[0].PackageID.String, existingDiscounts[0].DiscountID.String)
	validRow2 := fmt.Sprintf("%s,%s", existingPackages[1].PackageID.String, existingDiscounts[1].DiscountID.String)
	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	case "overwrite existing":
		var overwrittenRow string

		allExistingProductDiscount, err := s.selectAllProductDiscount(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(allExistingProductDiscount) == 0 {
			err = s.insertSomeProductAssociationDataDiscount(
				ctx,
				existingPackages[len(existingPackages)-1].PackageID.String,
				existingDiscounts[len(existingDiscounts)-1].DiscountID.String)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			allExistingProductDiscount, err = s.selectAllProductDiscount(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			overwrittenRow = fmt.Sprintf("%s,%s", allExistingProductDiscount[0].ProductID.String, allExistingProductDiscount[0].DiscountID.String)
		} else {
			overwrittenRow = fmt.Sprintf("%s,%s", allExistingProductDiscount[0].ProductID.String, allExistingProductDiscount[0].DiscountID.String)
		}

		err = s.insertSomeProductAssociationDataDiscount(
			ctx,
			allExistingProductDiscount[0].ProductID.String,
			allExistingProductDiscount[0].DiscountID.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		validRow := fmt.Sprintf("%s,%s", allExistingProductDiscount[0].ProductID.String, allExistingProductDiscount[0].DiscountID.String)

		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(fmt.Sprintf(`product_id,discount_id
          %s`, validRow)),
		}
		stepState.ValidCsvRows = []string{validRow}
		stepState.OverwrittenCsvRows = []string{overwrittenRow}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anProductDiscountInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{}
	case "header only":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(`product_id,discount_id
			1`),
		}
	case "number of column is not equal 2":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(`product_id
			1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(`product_id,discount_id
      1`),
		}
	case "wrong product_id column name in header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(`wrong_header,discount_id
			1,1`),
		}
	case "wrong discount_id column name in header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(`product_id,wrong_header
     		1,1`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anProductDiscountValidRequestPayloadWithIncorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomePackages(ctx)
	if err != nil {
		fmt.Printf("error when insert package %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}
	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertSomeDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingDiscounts, err := s.selectAllDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	headerTitles := []string{
		"product_id",
		"discount_id",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf("%s,%s", existingPackages[len(existingPackages)-3].PackageID.String, existingDiscounts[len(existingDiscounts)-3].DiscountID.String)
	validRow2 := fmt.Sprintf("%s,%s", existingPackages[len(existingPackages)-4].PackageID.String, existingDiscounts[len(existingDiscounts)-4].DiscountID.String)
	invalidEmptyRow1 := ",3"
	invalidEmptyRow2 := "4,"
	invalidValueRow1 := "a,5"
	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(fmt.Sprintf(
				`%s
				%s`,
				headerText, invalidValueRow1,
			)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s
			%s
			%s`, headerText, validRow1, validRow2, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductDiscount(ctx context.Context, userGroup string) (context.Context, error) {
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

func (s *suite) theValidProductDiscountLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductDiscount, err := s.selectAllProductDiscount(ctx)
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

		discountID := strings.TrimSpace(rowSplit[1])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allProductDiscount {
			if e.ProductID.String == productID && e.DiscountID.String == discountID && e.CreatedAt.Time.Before(time.Now()) {
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

func (s *suite) theImportProductDiscountTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductDiscount, err := s.selectAllProductDiscount(ctx)
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

		discountID := strings.TrimSpace(rowSplit[1])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allProductDiscount {
			if e.ProductID.String == productID && e.DiscountID.String == discountID && e.CreatedAt.Time.Before(time.Now()) {
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

		discountID := strings.TrimSpace(rowSplit[1])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, e := range allProductDiscount {
			if e.ProductID.String == productID && e.DiscountID.String == discountID {
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

func (s *suite) theInvalidProductDiscountLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
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

func (s *suite) selectAllProductDiscount(ctx context.Context) ([]*entities.ProductDiscount, error) {
	var allEntities []*entities.ProductDiscount
	const stmt = `
		SELECT
			product_id,
			discount_id,
			created_at 
		FROM product_discount
	`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query product_discount")
	}
	defer rows.Close()
	for rows.Next() {
		var entity entities.ProductDiscount
		if err := rows.Scan(
			&entity.ProductID,
			&entity.DiscountID,
			&entity.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan product_discount")
		}
		allEntities = append(allEntities, &entity)
	}
	return allEntities, nil
}

func (s *suite) insertSomeProductAssociationDataDiscount(ctx context.Context, productID string, discountID string) error {
	stmt := `INSERT INTO product_discount(
                product_id,
                discount_id,
                created_at)
            VALUES ($1, $2, now()) ON CONFLICT DO NOTHING`
	_, err := s.FatimaDBTrace.Exec(ctx, stmt, productID, discountID)
	if err != nil {
		return fmt.Errorf("cannot insert product associated data discount, err: %s", err)
	}

	return nil
}
