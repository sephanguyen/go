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

func (s *suite) associatedProductsByMaterialValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
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

	courseIDs, err := s.insertCourses(ctx)
	if err != nil {
		fmt.Printf("error when insert package %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertSomeMaterials(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingMaterials, err := s.selectAllMaterials(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "all valid rows":
		validRow1 := fmt.Sprintf("%s,%s,%s,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true", existingPackages[len(existingPackages)-1].PackageID.String, courseIDs[len(courseIDs)-1], existingMaterials[len(existingMaterials)-1].MaterialID.String)
		validRow2 := fmt.Sprintf("%s,%s,%s,2021-12-07,2022-12-07,true", existingPackages[len(existingPackages)-2].PackageID.String, courseIDs[len(courseIDs)-2], existingMaterials[len(existingMaterials)-2].MaterialID.String)

		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(fmt.Sprintf(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
          %s
          %s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	case "overwrite existing":
		var overwrittenRow string

		allAssociatedProductsByMaterial, err := s.selectAllAssociatedProductsByMaterial(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(allAssociatedProductsByMaterial) == 0 {
			err = s.insertSomeAssociatedProductsByMaterial(
				ctx,
				existingPackages[len(existingPackages)-1].PackageID.String,
				courseIDs[len(courseIDs)-1],
				existingMaterials[len(existingMaterials)-1].MaterialID.String)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			allAssociatedProductsByMaterial, err = s.selectAllAssociatedProductsByMaterial(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			overwrittenRow = fmt.Sprintf("%s,%s,%s,,,true", existingPackages[len(existingPackages)-1].PackageID.String, courseIDs[len(courseIDs)-1], existingMaterials[len(existingMaterials)-1].MaterialID.String)
		} else {
			overwrittenRow = fmt.Sprintf("%s,%s,%s,,,true", allAssociatedProductsByMaterial[0].PackageID.String, allAssociatedProductsByMaterial[0].CourseID.String, allAssociatedProductsByMaterial[0].MaterialID.String)
		}

		newCourseIDs, err := s.insertCourses(ctx)
		if err != nil {
			fmt.Printf("error when insert package %v\n", err.Error())
			return StepStateToContext(ctx, stepState), err
		}

		err = s.insertSomeAssociatedProductsByMaterial(
			ctx,
			allAssociatedProductsByMaterial[0].PackageID.String,
			newCourseIDs[len(newCourseIDs)-1],
			allAssociatedProductsByMaterial[0].MaterialID.String)

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		validRow := fmt.Sprintf("%s,%s,%s,,,true", allAssociatedProductsByMaterial[0].PackageID.String, newCourseIDs[len(newCourseIDs)-1], allAssociatedProductsByMaterial[0].MaterialID.String)

		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(fmt.Sprintf(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
          %s`, validRow)),
		}

		stepState.ValidCsvRows = []string{validRow}
		stepState.OverwrittenCsvRows = []string{overwrittenRow}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) associatedProductsByMaterialInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportAssociatedProductsRequest{}
	case "header only":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload:                []byte(`package_id,course_id,material_id,available_from,available_until,is_added_by_default`),
		}
	case "number of column is not equal 2 package_id only":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(`package_id
      1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
      1,Course-2,3`),
		}
	case "wrong package_id column name in csv header":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(`wrong_header,course_id,material_id,available_from,available_until,is_added_by_default
      1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
		}
	case "wrong course_id column name in csv header":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(`package_id,wrong_header,material_id,available_from,available_until,is_added_by_default
      1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
		}
	case "wrong material_id column name in csv header":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(`package_id,course_id,wrong_header,available_from,available_until,is_added_by_default
      1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
		}
	case "wrong available_from column name in csv header":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(`package_id,course_id,material_id,wrong_header,available_until,is_added_by_default
      1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
		}
	case "wrong available_until column name in csv header":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(`package_id,course_id,material_id,available_from,wrong_header,is_added_by_default
      1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingAssociatedProductsByMaterial(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportAssociatedProducts(contextWithToken(ctx), stepState.Request.(*pb.ImportAssociatedProductsRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidAssociatedProductsByMaterialLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportAssociatedProductsRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportAssociatedProductsResponse)
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

func (s *suite) associatedProductsByMaterialValidRequestPayloadWithIncorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
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

	courseIDs, err := s.insertCourses(ctx)
	if err != nil {
		fmt.Printf("error when insert package %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertSomeMaterials(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingMaterials, err := s.selectAllMaterials(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	invalidEmptyRow1 := ",Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false"
	invalidEmptyRow2 := "1,,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false"
	invalidEmptyRow3 := "1,Course-2,,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false"

	invalidValueRow1 := "a,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true"
	invalidValueRow2 := "1,Course-2,b,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true"
	invalidValueRow3 := "1,Course-2,3,c,2022-12-07T00:00:00-07:00,true"
	invalidValueRow4 := "1,Course-2,3,2021-12-07T00:00:00-07:00,d,true"

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(fmt.Sprintf(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
          %s
          %s
          %s`, invalidEmptyRow1, invalidEmptyRow2, invalidEmptyRow3)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidEmptyRow3}
	case "invalid value row":
		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(fmt.Sprintf(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
          %s
          %s
          %s
          %s`, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4}
	case "valid and invalid rows":
		validRow1 := fmt.Sprintf("%s,%s,%s,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true", existingPackages[0].PackageID.String, courseIDs[0], existingMaterials[0].MaterialID.String)
		validRow2 := fmt.Sprintf("%s,%s,%s,2021-12-07,2022-12-07,true", existingPackages[1].PackageID.String, courseIDs[1], existingMaterials[1].MaterialID.String)

		stepState.Request = &pb.ImportAssociatedProductsRequest{
			AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			Payload: []byte(fmt.Sprintf(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
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

func (s *suite) theImportAssociatedProductsByMaterialTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	allAssociatedProductsByMaterial, err := s.selectAllAssociatedProductsByMaterial(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false
		rowSplit := strings.Split(row, ",")

		packageID := strings.TrimSpace(rowSplit[0])

		courseID := strings.TrimSpace(rowSplit[1])

		materialID := strings.TrimSpace(rowSplit[2])

		for _, e := range allAssociatedProductsByMaterial {
			if e.PackageID.String == packageID && e.MaterialID.String == materialID && courseID == e.CourseID.String {
				found = true
				break
			}
		}

		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("ailed to rollback valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidAssociatedProductsByMaterialLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allAssociatedProductsByMaterial, err := s.selectAllAssociatedProductsByMaterial(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false
		rowSplit := strings.Split(row, ",")

		packageID := strings.TrimSpace(rowSplit[0])

		courseID := strings.TrimSpace(rowSplit[1])

		materialID := strings.TrimSpace(rowSplit[2])

		for _, e := range allAssociatedProductsByMaterial {
			if e.PackageID.String == packageID && e.MaterialID.String == materialID && courseID == e.CourseID.String {
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

func (s *suite) selectAllAssociatedProductsByMaterial(ctx context.Context) ([]*entities.PackageCourseMaterial, error) {
	var allEntities []*entities.PackageCourseMaterial
	stmt := `SELECT
                *
            FROM
                package_course_material`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query package_course_material")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.PackageCourseMaterial{}
		err := rows.Scan(
			&e.PackageID,
			&e.CourseID,
			&e.MaterialID,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.CreatedAt,
			&e.ResourcePath,
			&e.IsAddedByDefault,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan package_course_material")
		}
		allEntities = append(allEntities, e)
	}

	return allEntities, nil
}

func (s *suite) insertSomeAssociatedProductsByMaterial(ctx context.Context, packageID string, courseID string, materialID string) error {
	stmt := `INSERT INTO package_course_material(
                package_id,
                course_id,
                material_id,
                created_at,
				is_added_by_default)
            VALUES ($1, $2, $3, now(),$4)`
	_, err := s.FatimaDBTrace.Exec(ctx, stmt, packageID, courseID, materialID, true)
	if err != nil {
		return fmt.Errorf("cannot insert associated products by material, err: %s", err)
	}

	return nil
}
