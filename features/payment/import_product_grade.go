package payment

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aProductGradeValidRequestPayloadWithCorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
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

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "all valid rows":
		validRow1 := fmt.Sprintf("%s,%d", existingPackages[len(existingPackages)-1].PackageID.String, s.randomGrade())
		validRow2 := fmt.Sprintf("%s,%v", existingPackages[len(existingPackages)-2].PackageID.String, s.randomGrade())

		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(fmt.Sprintf(`product_id,grade_id
          %s
          %s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}

	case "overwrite existing":
		var overwrittenRow string

		allExistingProductGrade, err := s.selectAllProductGrade(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(allExistingProductGrade) == 0 {
			err = s.insertSomeProductAssociationDataGrade(
				ctx,
				existingPackages[len(existingPackages)-1].PackageID.String,
				s.randomGrade())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			allExistingProductGrade, err = s.selectAllProductGrade(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			overwrittenRow = fmt.Sprintf("%s,%v", allExistingProductGrade[0].ProductID.String, s.randomGrade())
		} else {
			overwrittenRow = fmt.Sprintf("%s,%v", allExistingProductGrade[0].ProductID.String, allExistingProductGrade[0].GradeID.String)
		}

		err = s.insertSomeProductAssociationDataGrade(
			ctx,
			allExistingProductGrade[0].ProductID.String,
			s.randomGrade())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		validRow := fmt.Sprintf("%s,%d", allExistingProductGrade[0].ProductID.String, s.randomGrade())

		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(fmt.Sprintf(`product_id,grade_id
          %s`, validRow)),
		}
		stepState.ValidCsvRows = []string{validRow}
		stepState.OverwrittenCsvRows = []string{overwrittenRow}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aProductGradeValidRequestPayloadWithIncorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
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
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(fmt.Sprintf(`product_id,grade_id
          %s
          %s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}

	case "invalid value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(fmt.Sprintf(`product_id,grade_id
          %s
          %s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}

	case "valid and invalid rows":
		validRow1 := fmt.Sprintf("%s,%d", existingPackages[len(existingPackages)-3].PackageID.String, s.randomGrade())
		validRow2 := fmt.Sprintf("%s,%d", existingPackages[len(existingPackages)-4].PackageID.String, s.randomGrade())

		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(fmt.Sprintf(`product_id,grade_id
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

func (s *suite) theValidProductGradeLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductGrades, err := s.selectAllProductGrade(ctx)
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

		gradeID := strings.TrimSpace(rowSplit[1])

		for _, e := range allProductGrades {
			if e.ProductID.String == productID && e.GradeID.String == gradeID && e.CreatedAt.Time.Before(time.Now()) {
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

		gradeID := strings.TrimSpace(rowSplit[1])
		for _, e := range allProductGrades {
			if e.ProductID.String == productID && e.GradeID.String == gradeID {
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

func (s *suite) theImportProductGradeTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductGrades, err := s.selectAllProductGrade(ctx)
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

		gradeID := strings.TrimSpace(rowSplit[1])

		for _, e := range allProductGrades {
			if e.ProductID.String == productID && e.GradeID.String == gradeID && e.CreatedAt.Time.Before(time.Now()) {
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

		gradeID := strings.TrimSpace(rowSplit[1])
		for _, e := range allProductGrades {
			if e.ProductID.String == productID && e.GradeID.String == gradeID {
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

func (s *suite) theInvalidProductGradeLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
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

func (s *suite) aProductGradeInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{}
	case "header only":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload:                   []byte(`product_id,grade_id`),
		}
	case "number of column is not equal 2":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(`product_id
      1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(`product_id,grade_id
      1`),
		}
	case "wrong product_id column name in header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(`wrong_header,grade_id
      1,1`),
		}
	case "wrong grade_id column name in header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE,
			Payload: []byte(`product_id,wrong_header
      1,1`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductGrade(ctx context.Context, userGroup string) (context.Context, error) {
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

func (s *suite) selectAllProductGrade(ctx context.Context) ([]*entities.ProductGrade, error) {
	allEntities := []*entities.ProductGrade{}
	stmt := `SELECT
                product_id,
                grade_id,
                created_at
            FROM
                product_grade`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query product_grade")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.ProductGrade{}
		err := rows.Scan(
			&e.ProductID,
			&e.GradeID,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan product_grade")
		}
		allEntities = append(allEntities, e)
	}

	return allEntities, nil
}

func (s *suite) insertSomeProductAssociationDataGrade(ctx context.Context, productID string, gradeID int32) error {
	stmt := `INSERT INTO product_grade(
                product_id,
                grade_id,
                created_at)
            VALUES ($1, $2, now()) ON CONFLICT DO NOTHING`
	_, err := s.FatimaDBTrace.Exec(ctx, stmt, productID, gradeID)
	if err != nil {
		return fmt.Errorf("cannot insert product associated data grade, err: %s", err)
	}

	return nil
}

// Get random grade
func (s *suite) randomGrade() int32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int31n(11) + 1
}
