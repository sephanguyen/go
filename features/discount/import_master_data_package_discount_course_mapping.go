package discount

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/discount/mockdata"
	"github.com/manabie-com/backend/internal/discount/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aPackageDiscountCourseMappingPayloadWithData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	headerTitles := getPackageDiscountCourseMappingHeader()
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

	courseIDOne, err := mockdata.InsertCourse(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	courseIDTwo, err := mockdata.InsertCourse(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRowOne := fmt.Sprintf("%s,%s,%s,0,%s", packageIDOne, courseIDOne, discountTagID, productGroup.ProductGroupID.String)
	validRoTwo := fmt.Sprintf("%s,%s,%s,0,%s", packageIDTwo, fmt.Sprintf("%s;%s", courseIDOne, courseIDTwo), discountTagID, productGroup.ProductGroupID.String)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRowOne, validRoTwo)),
		}
		stepState.ValidCsvRows = []string{validRowOne, validRoTwo}
	case overWriteExisting:
		var overwrittenRowOne, overwrittenRowTwo string

		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s
				%s`, headerText, validRowOne, validRoTwo)),
		}

		ctx, err := s.importsPackageDiscountCourseMapping(ctx, "school admin")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		allExistingPackageDiscountCourseMapping, err := s.selectAllPackageDiscountCourseMappings(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.New("err cannot select records from product group")
		}

		overwrittenRowOne = fmt.Sprintf("%s,%s,%s,0,%s", allExistingPackageDiscountCourseMapping[0].PackageID.String, fmt.Sprintf("%s;%s", courseIDOne, courseIDTwo), allExistingPackageDiscountCourseMapping[0].DiscountTagID.String, productGroup.ProductGroupID.String)

		_, discountTagID, err := mockdata.InsertOrgDiscount(ctx, s.FatimaDBTrace, "combo-test")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		courseIDThree, err := mockdata.InsertCourse(ctx, s.FatimaDBTrace)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		overwrittenRowTwo = fmt.Sprintf("%s,%s,%s,0,%s", allExistingPackageDiscountCourseMapping[0].PackageID.String, fmt.Sprintf("%s;%s;%s", courseIDOne, courseIDTwo, courseIDThree), discountTagID, productGroup.ProductGroupID.String)

		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s
				%s`, headerText, overwrittenRowOne, overwrittenRowTwo)),
		}

		stepState.OverwrittenCsvRows = []string{overwrittenRowOne, overwrittenRowTwo}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importsPackageDiscountCourseMapping(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.DiscountConn).
		ImportPackageDiscountCourseMapping(contextWithToken(ctx), stepState.Request.(*pb.ImportPackageDiscountCourseMappingRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidPackageDiscountCourseMappingLinesWithDataAreImportedSuccessfully(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	packageDiscountCourseMappings, err := s.selectAllPackageDiscountCourseMappings(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	csvRows := stepState.ValidCsvRows
	if rowCondition == overWriteExisting {
		csvRows = stepState.OverwrittenCsvRows
	}

	err = s.comparePackageDiscountCourseMappingCsvValuesOnDB(csvRows, packageDiscountCourseMappings, false)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPackageDiscountCourseMappingRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getPackageDiscountCourseMappingHeader()
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

	courseIDOne, err := mockdata.InsertCourse(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	courseIDTwo, err := mockdata.InsertCourse(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRowOne := fmt.Sprintf("%s,%s,%s,0,%s", packageIDOne, courseIDOne, discountTagID, productGroup.ProductGroupID.String)
	validRowTwo := fmt.Sprintf("%s,%s,%s,0,%s", packageIDTwo, fmt.Sprintf("%s;%s", courseIDOne, courseIDTwo), discountTagID, productGroup.ProductGroupID.String)

	// package id should be required
	invalidEmptyRowOne := fmt.Sprintf(",%s,%s,0,%s", courseIDOne, discountTagID, productGroup.ProductGroupID.String)

	// course combination ids should be required
	invalidEmptyRowTwo := fmt.Sprintf("%s,,%s,0,%s", packageIDOne, discountTagID, productGroup.ProductGroupID.String)

	// discount tag id should be required
	invalidEmptyRowThree := fmt.Sprintf("%s,%s,,0,%s", packageIDOne, courseIDOne, productGroup.ProductGroupID.String)

	// non existing package id
	invalidValueRowOne := fmt.Sprintf("not-existing-package-id,%s,%s,0,%s", courseIDTwo, discountTagID, productGroup.ProductGroupID.String)
	// non existing course combination ids
	invalidValueRowTwo := fmt.Sprintf("%s,not-existing-course-ids,%s,0,%s", packageIDOne, discountTagID, productGroup.ProductGroupID.String)
	// non existing discount tag id
	invalidValueRowThree := fmt.Sprintf("%s,%s,not-existing-discount-tag-id,0,%s", packageIDOne, courseIDTwo, productGroup.ProductGroupID.String)

	// stepState.ValidCsvRows = []string{}
	// stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s`, headerText, invalidEmptyRowOne, invalidEmptyRowTwo, invalidEmptyRowThree)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRowOne, invalidEmptyRowTwo, invalidEmptyRowThree}

	case "invalid value row":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(
				`%s
					%s
					%s
					%s`,
				headerText, invalidValueRowOne, invalidValueRowTwo, invalidValueRowThree,
			)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRowOne, invalidValueRowTwo, invalidValueRowThree}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s
				%s
				%s
				%s
				%s
				%s
				%s
				%s`, headerText, validRowOne, validRowTwo, invalidEmptyRowOne, invalidEmptyRowTwo, invalidEmptyRowThree, invalidValueRowOne, invalidValueRowTwo, invalidValueRowThree)),
		}

		stepState.ValidCsvRows = []string{validRowOne, validRowTwo}
		stepState.InvalidCsvRows = []string{invalidEmptyRowOne, invalidEmptyRowTwo, invalidEmptyRowThree, invalidValueRowOne, invalidValueRowTwo, invalidValueRowThree}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportPackageDiscountCourseMappingTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	packageDiscountCourseMapping, err := s.selectAllPackageDiscountCourseMappings(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// combine valid and invalid rows
	csvRows := stepState.InvalidCsvRows
	csvRows = append(csvRows, stepState.ValidCsvRows...)

	err = s.comparePackageDiscountCourseMappingCsvValuesOnDB(csvRows, packageDiscountCourseMapping, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aPackageDiscountCourseMappingInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	headerTitles := getPackageDiscountCourseMappingHeader()
	headerText := strings.Join(headerTitles, ",")

	packageIDOne, err := s.createPackageWithID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	courseIDOne, err := mockdata.InsertCourse(ctx, s.FatimaDBTrace)
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

	csvRow := fmt.Sprintf("%s,%s,%s,0,%s", packageIDOne, courseIDOne, discountTagID, productGroup.ProductGroupID.String)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{}
	case "header only":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(headerText),
		}
	case "number of column is not equal 5":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(`package_id
			1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`%s
			1`, headerText)),
		}
	case "wrong package_id column name in header":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`wrong_header,course_combination_ids,discount_tag_id,is_archived,product_group_id
			%s`, csvRow)),
		}
	case "wrong course_combination_ids column name in header":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_id,wrong_course_combination_ids,discount_tag_id,is_archived,product_group_id
			%s`, csvRow)),
		}
	case "wrong discount_tag_id column name in header":
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_id,course_combination_ids,wrong_discount_tag_id,is_archived,product_group_id
			%s`, csvRow)),
		}
	case wrongIsArchivedColumnNameInHeader:
		stepState.Request = &pb.ImportPackageDiscountCourseMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_id,course_combination_ids,discount_tag_id,wrong_is_archived,product_group_id
			%s`, csvRow)),
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) comparePackageDiscountCourseMappingCsvValuesOnDB(csvRows []string, packageDiscountCourseMappings []*entities.PackageDiscountCourseMapping, isRollBack bool) error {
	for _, row := range csvRows {
		found := false
		rowSplit := strings.Split(row, ",")
		packageID := strings.TrimSpace(rowSplit[0])
		courseCombinationIDs := strings.TrimSpace(rowSplit[1])
		discountTagID := strings.TrimSpace(rowSplit[2])
		isArchived := strings.TrimSpace(rowSplit[3])
		productGroupID := strings.TrimSpace(rowSplit[4])

		for _, e := range packageDiscountCourseMappings {
			isArchivedDBStr := "0"
			if e.IsArchived.Bool {
				isArchivedDBStr = "1"
			}

			if e.PackageID.String == packageID && e.CourseCombinationIDs.String == courseCombinationIDs && e.DiscountTagID.String == discountTagID && isArchivedDBStr == isArchived && e.ProductGroupID.String == productGroupID && e.CreatedAt.Time.Before(time.Now()) {
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

func (s *suite) selectAllPackageDiscountCourseMappings(ctx context.Context) ([]*entities.PackageDiscountCourseMapping, error) {
	var allEntities []*entities.PackageDiscountCourseMapping
	const stmt = `
		SELECT
			package_id,
			course_combination_ids,
			discount_tag_id,
			is_archived,
			product_group_id,
			created_at 
		FROM package_discount_course_mapping
		ORDER BY created_at DESC
	`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)

	if err != nil {
		return nil, errors.Wrap(err, "query package_discount_course_mapping")
	}
	defer rows.Close()
	for rows.Next() {
		var entity entities.PackageDiscountCourseMapping
		if err := rows.Scan(
			&entity.PackageID,
			&entity.CourseCombinationIDs,
			&entity.DiscountTagID,
			&entity.IsArchived,
			&entity.ProductGroupID,
			&entity.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan package_discount_course_mapping")
		}
		allEntities = append(allEntities, &entity)
	}
	return allEntities, nil
}

func getPackageDiscountCourseMappingHeader() []string {
	return []string{
		"package_id",
		"course_combination_ids",
		"discount_tag_id",
		"is_archived",
		"product_group_id",
	}
}
