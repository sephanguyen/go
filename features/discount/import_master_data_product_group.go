package discount

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/pkg/errors"
)

func (s *suite) productGroupValidRequestPayloadWithCorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getProductGroupHeader()

	headerText := strings.Join(headerTitles, ",")
	// sample rows for csv
	randomID := idutil.ULIDNow()

	validRow1 := fmt.Sprintf(",group-name-%s,group-tag-%s,discount-type-%s,0", randomID, randomID, randomID)
	validRow2 := fmt.Sprintf(",group-name-%s,group-tag-%s,discount-type-%s,0", randomID, randomID, randomID)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}

	case "overwrite existing":
		var overwrittenRow1, overwrittenRow2 string

		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
		}

		ctx, err := s.importingProductGroup(ctx, "school admin")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		allExistingProductGroup, err := s.selectAllProductGroup(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.New("err cannot select records from product group")
		}
		newRandomID := idutil.ULIDNow()
		overwrittenRow1 = fmt.Sprintf("%s,group-name-override-%s,group-tag-override-%s,discount-type-%s,0", allExistingProductGroup[0].ProductGroupID.String, newRandomID, newRandomID, newRandomID)
		overwrittenRow2 = fmt.Sprintf("%s,group-name-override-%s,group-tag-override-%s,discount-type-override-%s,1", allExistingProductGroup[1].ProductGroupID.String, newRandomID, newRandomID, newRandomID)

		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, overwrittenRow1, overwrittenRow2)),
		}

		stepState.OverwrittenCsvRows = []string{overwrittenRow1, overwrittenRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductGroup(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.DiscountConn).
		ImportProductGroup(contextWithToken(ctx), stepState.Request.(*pb.ImportProductGroupRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidProductGroupLinesWithDataAreImportedSuccessfully(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	productGroups, err := s.selectAllProductGroup(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	csvRows := stepState.ValidCsvRows
	if rowCondition == "overwrite existing" {
		csvRows = stepState.OverwrittenCsvRows
	}

	err = s.checkAndCompareCsvValuesOnDB(csvRows, productGroups, false)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllProductGroup(ctx context.Context) ([]*entities.ProductGroup, error) {
	var allEntities []*entities.ProductGroup
	const stmt = `
		SELECT
			product_group_id,
			group_name,
			group_tag,
			discount_type,
			is_archived,
			created_at 
		FROM product_group
		ORDER BY created_at DESC
	`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)

	if err != nil {
		return nil, errors.Wrap(err, "query product_group")
	}
	defer rows.Close()
	for rows.Next() {
		var entity entities.ProductGroup
		if err := rows.Scan(
			&entity.ProductGroupID,
			&entity.GroupName,
			&entity.GroupTag,
			&entity.DiscountType,
			&entity.IsArchived,
			&entity.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan product_group")
		}
		allEntities = append(allEntities, &entity)
	}
	return allEntities, nil
}

func (s *suite) aProductGroupValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := idutil.ULIDNow()

	headerTitles := getProductGroupHeader()
	headerText := strings.Join(headerTitles, ",")

	validRow1 := fmt.Sprintf(",group-name-%s,group-tag-%s,discount-type-%s,0", randomID, randomID, randomID)
	validRow2 := fmt.Sprintf(",group-name-%s,group-tag-%s,discount-type-%s,0", randomID, randomID, randomID)

	// group name should be required
	invalidEmptyRow2 := fmt.Sprintf(",,group-tag-%s,discount-type-%s,0", randomID, randomID)
	// group tag should be required
	invalidEmptyRow3 := fmt.Sprintf(",group-name,,discount-type-%s,0", randomID)
	// is archived should be required
	invalidEmptyRow5 := fmt.Sprintf(",group-name,group-tag-%s,discount-type-%s,", randomID, randomID)

	invalidValueRow1 := fmt.Sprintf("non-exist-product-group-id,group-name-%s,group-tag-%s,discount-type-%s,test-invalid", randomID, randomID, randomID)

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s`, headerText, invalidEmptyRow2, invalidEmptyRow3, invalidEmptyRow5)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow2, invalidEmptyRow3, invalidEmptyRow5}

	case "invalid value row":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(
				`%s
				%s`,
				headerText, invalidValueRow1,
			)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s
			%s
			%s
			%s`, headerText, validRow1, validRow2, invalidEmptyRow2, invalidEmptyRow3, invalidEmptyRow5, invalidValueRow1)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
		stepState.InvalidCsvRows = []string{invalidEmptyRow2, invalidEmptyRow3, invalidEmptyRow5, invalidValueRow1}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkAndCompareCsvValuesOnDB(csvRows []string, productGroups []*entities.ProductGroup, isRollBack bool) error {
	for _, row := range csvRows {
		found := false
		rowSplit := strings.Split(row, ",")
		groupName := strings.TrimSpace(rowSplit[1])
		groupTag := strings.TrimSpace(rowSplit[2])
		discountType := strings.TrimSpace(rowSplit[3])
		isArchived := strings.TrimSpace(rowSplit[4])

		for _, e := range productGroups {
			isArchivedDBStr := "0"
			if e.IsArchived.Bool {
				isArchivedDBStr = "1"
			}

			if e.GroupName.String == groupName && e.GroupTag.String == groupTag && e.DiscountType.String == discountType && isArchivedDBStr == isArchived && e.CreatedAt.Time.Before(time.Now()) {
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

func (s *suite) aProductGroupInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := idutil.ULIDNow()

	headerTitles := getProductGroupHeader()
	headerText := strings.Join(headerTitles, ",")
	csvRow := fmt.Sprintf(",group-name-%s,group-tag-%s,discount-type-%s,0", randomID, randomID, randomID)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductGroupRequest{}
	case "header only":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(headerText),
		}
	case "number of column is not equal 5":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(`product_group_id
			1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`%s
			1`, headerText)),
		}
	case "wrong product_group_id column name in header":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`wrong_header,group_name,group_tag,discount_type,is_archived
			%s`, csvRow)),
		}
	case "wrong group_name column name in header":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`product_group_id,wrong_header,group_tag,discount_type,is_archived
			%s`, csvRow)),
		}
	case "wrong group_tag column name in header":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`product_group_id,group_name,wrong_header,discount_type,is_archived
			%s`, csvRow)),
		}
	case "wrong discount_type column name in header":
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`product_group_id,group_name,group_tag,wrong_header,is_archived
			%s`, csvRow)),
		}
	case wrongIsArchivedColumnNameInHeader:
		stepState.Request = &pb.ImportProductGroupRequest{
			Payload: []byte(fmt.Sprintf(`product_group_id,group_name,group_tag,discount_type,wrong_header
			%s`, csvRow)),
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportProductGroupTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	productGroups, err := s.selectAllProductGroup(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// combine valid and invalid rows
	csvRows := stepState.InvalidCsvRows
	csvRows = append(csvRows, stepState.ValidCsvRows...)

	err = s.checkAndCompareCsvValuesOnDB(csvRows, productGroups, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func getProductGroupHeader() []string {
	return []string{
		"product_group_id",
		"group_name",
		"group_tag",
		"discount_type",
		"is_archived",
	}
}
