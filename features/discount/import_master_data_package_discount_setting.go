package discount

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/discount/mockdata"
	"github.com/manabie-com/backend/internal/discount/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aPackageDiscountSettingPayloadWithData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	headerTitles := getPackageDiscountSettingHeader()
	headerText := strings.Join(headerTitles, ",")

	packageIDOne, err := s.createPackageWithID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	packageIDTwo, err := s.createPackageWithID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	discountTagID, ok := stepState.DiscountTagTypeAndIDMap[paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()]
	if !ok {
		return StepStateToContext(ctx, stepState), errors.New("there is no existing discount tag record with combo discount type")
	}
	discountType := paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()
	productGroup, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, "groupTag", discountType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf("%s,1,2,%s,0,%s", packageIDOne, discountTagID, productGroup.ProductGroupID.String)
	validRow2 := fmt.Sprintf("%s,3,5,%s,0,%s", packageIDTwo, discountTagID, productGroup.ProductGroupID.String)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	case overWriteExisting:
		var overwrittenRow1, overwrittenRow2 string

		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
		}

		ctx, err := s.importingPackageDiscountSetting(ctx, "school admin")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		allExistingPackageDiscountSetting, err := s.selectAllPackageDiscountSetting(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.New("err cannot select records from product group")
		}

		overwrittenRow1 = fmt.Sprintf("%s,2,2,%s,0,%s", allExistingPackageDiscountSetting[0].PackageID.String, allExistingPackageDiscountSetting[0].DiscountTagID.String, productGroup.ProductGroupID.String)

		_, discountTagID, err := mockdata.InsertOrgDiscount(ctx, s.FatimaDBTrace, "combo-test")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		overwrittenRow2 = fmt.Sprintf("%s,2,2,%s,0,%s", allExistingPackageDiscountSetting[1].PackageID.String, discountTagID, productGroup.ProductGroupID.String)

		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, overwrittenRow1, overwrittenRow2)),
		}

		stepState.OverwrittenCsvRows = []string{overwrittenRow1, overwrittenRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingPackageDiscountSetting(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.DiscountConn).
		ImportPackageDiscountSetting(contextWithToken(ctx), stepState.Request.(*pb.ImportPackageDiscountSettingRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidPackageDiscountSettingLinesWithDataAreImportedSuccessfully(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	packageDiscountSettings, err := s.selectAllPackageDiscountSetting(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	csvRows := stepState.ValidCsvRows
	if rowCondition == overWriteExisting {
		csvRows = stepState.OverwrittenCsvRows
	}

	err = s.comparePackageDiscountSettingMappingCsvValuesOnDB(csvRows, packageDiscountSettings, false)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPackageDiscountSettingRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getPackageDiscountSettingHeader()
	headerText := strings.Join(headerTitles, ",")

	packageIDOne, err := s.createPackageWithID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	packageIDTwo, err := s.createPackageWithID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	discountTagID, ok := stepState.DiscountTagTypeAndIDMap[paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()]
	if !ok {
		return StepStateToContext(ctx, stepState), errors.New("there is no existing discount tag record with combo discount type")
	}
	discountType := paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()
	productGroup, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, "groupTag", discountType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf("%s,0,0,%s,0,%s", packageIDOne, discountTagID, productGroup.ProductGroupID.String)
	validRow2 := fmt.Sprintf("%s,1,2,%s,0,%s", packageIDTwo, discountTagID, productGroup.ProductGroupID.String)

	// package id should be required
	invalidEmptyRow1 := fmt.Sprintf(",0,0,%s,0,%s", discountTagID, productGroup.ProductGroupID.String)
	// discount tag id should be required
	invalidEmptyRow2 := fmt.Sprintf("%s,0,0,,0,%s", packageIDOne, productGroup.ProductGroupID.String)

	// non existing package id
	invalidValueRow1 := fmt.Sprintf("not-existing-package-id,0,0,%s,0,%s", discountTagID, productGroup.ProductGroupID.String)
	// non existing discount tag id
	invalidValueRow2 := fmt.Sprintf("%s,0,0,not-existing-discount-tag-id,0,%s", packageIDOne, productGroup.ProductGroupID.String)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}

	case "invalid value row":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(
				`%s
				%s
				%s`,
				headerText, invalidValueRow1, invalidValueRow2,
			)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s
			%s
			%s
			%s`, headerText, validRow1, validRow2, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}

		stepState.ValidCsvRows = []string{validRow1, validRow2}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createPackageWithID(ctx context.Context) (string, error) {
	var packageID string
	taxID, err := mockdata.InsertOneTax(ctx, s.FatimaDBTrace)
	if err != nil {
		return packageID, err
	}

	packageID, err = mockdata.InsertPackage(ctx, taxID, s.FatimaDBTrace)
	if err != nil {
		return packageID, err
	}

	return packageID, nil
}

func (s *suite) theImportPackageDiscountSettingTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	packageDiscountSettingMapping, err := s.selectAllPackageDiscountSetting(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// combine valid and invalid rows
	csvRows := stepState.InvalidCsvRows
	csvRows = append(csvRows, stepState.ValidCsvRows...)

	err = s.comparePackageDiscountSettingMappingCsvValuesOnDB(csvRows, packageDiscountSettingMapping, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPackageDiscountSettingInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getPackageDiscountSettingHeader()
	headerText := strings.Join(headerTitles, ",")

	taxID, err := mockdata.InsertOneTax(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	packageID, err := mockdata.InsertPackage(ctx, taxID, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	discountTagID, ok := stepState.DiscountTagTypeAndIDMap[paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()]
	if !ok {
		return StepStateToContext(ctx, stepState), errors.New("there is no existing discount tag record with combo discount type")
	}
	discountType := paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()
	productGroup, err := mockdata.InsertProductGroup(ctx, s.FatimaDBTrace, "groupTag", discountType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	csvRow := fmt.Sprintf("%s,0,0,%s,0,%s", packageID, discountTagID, productGroup.ProductGroupID.String)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{}
	case "header only":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(headerText),
		}
	case "number of column is not equal 5":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(`package_id
			1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			1`, headerText)),
		}
	case "wrong package_id column name in header":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`wrong_header,min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
			%s`, csvRow)),
		}
	case "wrong min_slot_trigger column name in header":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`package_id,wrong_min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
			%s`, csvRow)),
		}
	case "wrong max_slot_trigger column name in header":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`package_id,min_slot_trigger,wrong_max_slot_trigger,discount_tag_id,is_archived,product_group_id
			%s`, csvRow)),
		}
	case "wrong discount_tag_id column name in header":
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`package_id,min_slot_trigger,max_slot_trigger,wrong_discount_tag_id,is_archived,product_group_id
			%s`, csvRow)),
		}
	case wrongIsArchivedColumnNameInHeader:
		stepState.Request = &pb.ImportPackageDiscountSettingRequest{
			Payload: []byte(fmt.Sprintf(`package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,wrong_is_archived,product_group_id
			%s`, csvRow)),
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllPackageDiscountSetting(ctx context.Context) ([]*entities.PackageDiscountSetting, error) {
	var allEntities []*entities.PackageDiscountSetting
	const stmt = `
		SELECT
			package_id,
			min_slot_trigger,
			max_slot_trigger,
			discount_tag_id,
			is_archived,
			product_group_id,
			created_at 
		FROM package_discount_setting
		ORDER BY created_at DESC
	`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)

	if err != nil {
		return nil, errors.Wrap(err, "query package_discount_setting")
	}
	defer rows.Close()
	for rows.Next() {
		var entity entities.PackageDiscountSetting
		if err := rows.Scan(
			&entity.PackageID,
			&entity.MinSlotTrigger,
			&entity.MaxSlotTrigger,
			&entity.DiscountTagID,
			&entity.IsArchived,
			&entity.ProductGroupID,
			&entity.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan package_discount_setting")
		}
		allEntities = append(allEntities, &entity)
	}
	return allEntities, nil
}

func (s *suite) comparePackageDiscountSettingMappingCsvValuesOnDB(csvRows []string, packageDiscountSettingMappings []*entities.PackageDiscountSetting, isRollBack bool) error {
	for _, row := range csvRows {
		found := false
		rowSplit := strings.Split(row, ",")
		packageID := strings.TrimSpace(rowSplit[0])
		minSlotTriggerStr := strings.TrimSpace(rowSplit[1])
		maxSlotTriggerStr := strings.TrimSpace(rowSplit[2])
		discountTagID := strings.TrimSpace(rowSplit[3])
		isArchived := strings.TrimSpace(rowSplit[4])
		productGroupID := strings.TrimSpace(rowSplit[5])

		minSlotTrigger, err := strconv.Atoi(minSlotTriggerStr)
		if err != nil {
			return err
		}

		maxSlotTrigger, err := strconv.Atoi(maxSlotTriggerStr)
		if err != nil {
			return err
		}

		for _, e := range packageDiscountSettingMappings {
			isArchivedDBStr := "0"
			if e.IsArchived.Bool {
				isArchivedDBStr = "1"
			}

			if e.PackageID.String == packageID && int(e.MinSlotTrigger.Int) == minSlotTrigger && int(e.MaxSlotTrigger.Int) == maxSlotTrigger && e.DiscountTagID.String == discountTagID && isArchivedDBStr == isArchived && e.ProductGroupID.String == productGroupID && e.CreatedAt.Time.Before(time.Now()) {
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

func getPackageDiscountSettingHeader() []string {
	return []string{
		"package_id",
		"min_slot_trigger",
		"max_slot_trigger",
		"discount_tag_id",
		"is_archived",
		"product_group_id",
	}
}
