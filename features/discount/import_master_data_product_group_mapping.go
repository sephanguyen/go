package discount

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/discount/mockdata"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) productGroupMappingValidRequestPayloadWithCorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getProductGroupMappingHeader()

	headerText := strings.Join(headerTitles, ",")

	productIDs, err := mockdata.InsertRecurringProducts(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	productGroupA, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, fmt.Sprintf("COMBO-%s", idutil.ULIDNow()), paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	productGroupB, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, fmt.Sprintf("SIBLING-%s", idutil.ULIDNow()), paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf("%s,%s", productGroupA.ProductGroupID.String, productIDs[0])
	validRow2 := fmt.Sprintf("%s,%s", productGroupB.ProductGroupID.String, productIDs[1])

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":

		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}

	case overWriteExisting:
		var overwrittenRow1, overwrittenRow2 string

		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s
				%s`, headerText, validRow1, validRow2)),
		}

		ctx, err := s.importingProductGroupMapping(ctx, "school admin")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		allExistingProductGroupMapping, err := s.selectAllProductGroupMapping(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.New("err cannot select records from product group")
		}

		overwrittenRow1 = fmt.Sprintf("%s,%s", allExistingProductGroupMapping[0].ProductGroupID.String, allExistingProductGroupMapping[0].ProductID.String)
		overwrittenRow2 = fmt.Sprintf("%s,%s", allExistingProductGroupMapping[1].ProductGroupID.String, allExistingProductGroupMapping[1].ProductID.String)

		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s
				%s`, headerText, overwrittenRow1, overwrittenRow2)),
		}

		stepState.OverwrittenCsvRows = []string{overwrittenRow1, overwrittenRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductGroupMapping(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.DiscountConn).
		ImportProductGroupMapping(contextWithToken(ctx), stepState.Request.(*pb.ImportProductGroupMappingRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidProductGroupMappingLinesWithDataAreImportedSuccessfully(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	productGroupMappings, err := s.selectAllProductGroupMapping(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	csvRows := stepState.ValidCsvRows
	if rowCondition == overWriteExisting {
		csvRows = stepState.OverwrittenCsvRows
	}

	err = s.compareProductGroupMappingCsvValuesOnDB(csvRows, productGroupMappings, false)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllProductGroupMapping(ctx context.Context) ([]*entities.ProductGroupMapping, error) {
	var allEntities []*entities.ProductGroupMapping
	const stmt = `
		SELECT
			product_group_id,
			product_id,
			created_at 
		FROM product_group_mapping
		ORDER BY created_at DESC
	`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)

	if err != nil {
		return nil, errors.Wrap(err, "query product_group_mapping")
	}
	defer rows.Close()
	for rows.Next() {
		var entity entities.ProductGroupMapping
		if err := rows.Scan(
			&entity.ProductGroupID,
			&entity.ProductID,
			&entity.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan product_group_mapping")
		}
		allEntities = append(allEntities, &entity)
	}
	return allEntities, nil
}

func (s *suite) compareProductGroupMappingCsvValuesOnDB(csvRows []string, productGroupMappings []*entities.ProductGroupMapping, isRollBack bool) error {
	for _, row := range csvRows {
		found := false
		rowSplit := strings.Split(row, ",")
		productGroupID := strings.TrimSpace(rowSplit[0])
		productID := strings.TrimSpace(rowSplit[1])
		for _, e := range productGroupMappings {
			if e.ProductGroupID.String == productGroupID && e.ProductID.String == productID && e.CreatedAt.Time.Before(time.Now()) {
				found = true
				break
			}
		}

		switch isRollBack {
		case true:
			if found {
				return fmt.Errorf("failed to rollback valid csv row")
			}

		default:
			if !found {
				return fmt.Errorf("failed to import valid csv row")
			}
		}
	}

	return nil
}

func (s *suite) aProductGroupMappingInvalidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getProductGroupMappingHeader()
	headerText := strings.Join(headerTitles, ",")

	productIDs, err := mockdata.InsertRecurringProducts(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	productGroupA, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, fmt.Sprintf("COMBO-%s", idutil.ULIDNow()), paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	productGroupB, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, fmt.Sprintf("SIBLING-%s", idutil.ULIDNow()), paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf("%s,%s", productGroupA.ProductGroupID.String, productIDs[0])
	validRow2 := fmt.Sprintf("%s,%s", productGroupB.ProductGroupID.String, productIDs[1])

	// product group id should be required
	invalidEmptyRow1 := fmt.Sprintf(",%s", productIDs[0])
	// product id should be required
	invalidEmptyRow2 := fmt.Sprintf("%s,", productGroupA.ProductGroupID.String)

	invalidValueRow1 := fmt.Sprintf("non-exist-product-group-id,product-id-%s", productIDs[0])

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}

	case "invalid value row":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(
				`%s
				%s`,
				headerText, invalidValueRow1,
			)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
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

func (s *suite) theImportProductGroupMappingTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	productGroupMapping, err := s.selectAllProductGroupMapping(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// combine valid and invalid rows
	csvRows := stepState.InvalidCsvRows
	csvRows = append(csvRows, stepState.ValidCsvRows...)

	err = s.compareProductGroupMappingCsvValuesOnDB(csvRows, productGroupMapping, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aProductGroupMappingInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getProductGroupMappingHeader()
	headerText := strings.Join(headerTitles, ",")

	productIDs, err := mockdata.InsertRecurringProducts(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	productGroupA, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, fmt.Sprintf("COMBO-%s", idutil.ULIDNow()), paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	csvRow := fmt.Sprintf("%s,%s", productGroupA.ProductGroupID.String, productIDs[0])

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductGroupMappingRequest{}
	case "header only":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(headerText),
		}
	case "number of column is not equal 2":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(`product_group_id
			1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			1`, headerText)),
		}
	case "wrong product_group_id column name in header":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(`wrong_header,product_id
			%s`, csvRow)),
		}
	case "wrong product_id column name in header":
		stepState.Request = &pb.ImportProductGroupMappingRequest{
			Payload: []byte(fmt.Sprintf(`product_group_id,wrong_header
			%s`, csvRow)),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func getProductGroupMappingHeader() []string {
	return []string{
		"product_group_id",
		"product_id",
	}
}
