package payment

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aPackageQuantityTypeMappingValidRequestPayloadWithCorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	validRow1 := "1,2"
	validRow2 := "2,1"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_type,quantity_type
		      %s
		      %s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	case "overwrite existing":
		overWrittenRow := "1,2"
		updatedRow := validRow1

		err := s.upsertPackageQuantityTypeMapping(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_type,quantity_type
		      %s`, validRow1)),
		}
		stepState.ValidCsvRows = []string{updatedRow}
		stepState.OverwrittenCsvRows = []string{overWrittenRow}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPackageQuantityTypeMappingValidRequestPayloadWithIncorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	validRow1 := "1,1"
	validRow2 := "2,2"
	validRow3 := "3,3"
	validRow4 := "4,1"

	invalidEmptyRow1 := ",3"
	invalidEmptyRow2 := "4,"

	invalidValueRow1 := "a,1"
	invalidValueRow2 := "1,b"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	stepState.OverwrittenCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_type,quantity_type
		      %s
		      %s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_type,quantity_type
		      %s
		      %s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_type,quantity_type
		      %s
		      %s
		      %s
		      %s
		      %s
		      %s
		      %s
			  %s`, validRow1, validRow2, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, validRow3, validRow4)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow3}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	case "overwrite existing":
		overWrittenRow := "1,2"
		updatedRow := validRow1

		err := s.upsertPackageQuantityTypeMapping(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(fmt.Sprintf(`package_type,quantity_type
		      %s`, validRow1)),
		}
		stepState.ValidCsvRows = []string{updatedRow}
		stepState.OverwrittenCsvRows = []string{overWrittenRow}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingPackageQuantityTypeMapping(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportPackageQuantityTypeMapping(contextWithToken(ctx), stepState.Request.(*pb.ImportPackageQuantityTypeMappingRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidPackageQuantityTypeMappingLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allPackageQuantityTypeMapping, err := s.selectAllPackageQuantityTypeMapping(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")
		packageType, err := strconv.Atoi(strings.TrimSpace(rowSplit[0]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		packageTypeEnum := s.convertIntToPackageType(packageType)

		quantityType, err := strconv.Atoi(strings.TrimSpace(rowSplit[1]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quantityTypeEnum := s.convertIntToQuantityType(quantityType)

		for _, e := range allPackageQuantityTypeMapping {
			if e.PackageType.String == packageTypeEnum.String() && e.QuantityType.String == quantityTypeEnum.String() {
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
		packageType, err := strconv.Atoi(strings.TrimSpace(rowSplit[0]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		packageTypeEnum := s.convertIntToPackageType(packageType)

		quantityType, err := strconv.Atoi(strings.TrimSpace(rowSplit[1]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quantityTypeEnum := s.convertIntToQuantityType(quantityType)

		for _, e := range allPackageQuantityTypeMapping {
			if e.PackageType.String == packageTypeEnum.String() && e.QuantityType.String == quantityTypeEnum.String() {
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

func (s *suite) theImportPackageQuantityTypeMappingTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allPackageQuantityTypeMapping, err := s.selectAllPackageQuantityTypeMapping(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")
		packageType, err := strconv.Atoi(strings.TrimSpace(rowSplit[0]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		packageTypeEnum := s.convertIntToPackageType(packageType)

		quantityType, err := strconv.Atoi(strings.TrimSpace(rowSplit[1]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quantityTypeEnum := s.convertIntToQuantityType(quantityType)

		for _, e := range allPackageQuantityTypeMapping {
			if e.PackageType.String == packageTypeEnum.String() && e.QuantityType.String == quantityTypeEnum.String() {
				found = true
				break
			}
		}

		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to rollback valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidPackageQuantityTypeMappingLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportPackageQuantityTypeMappingRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportPackageQuantityTypeMappingResponse)
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

func (s *suite) aPackageQuantityTypeMappingInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{}
	case "header only":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(`package_type,quantity_type`),
		}
	case "number of column is not equal 2":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(`package_type
      1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(`package_type,quantity_type
      1`),
		}
	case "incorrect package_type column name in header":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(`incorrect_package_type,quantity_type
      1,1`),
		}
	case "incorrect quantity_type column name in header":
		stepState.Request = &pb.ImportPackageQuantityTypeMappingRequest{
			Payload: []byte(`package_type,incorrect_quantity_type
      1,1`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllPackageQuantityTypeMapping(ctx context.Context) ([]*entities.PackageQuantityTypeMapping, error) {
	var allEntities []*entities.PackageQuantityTypeMapping
	stmt := `SELECT
                package_type,
                quantity_type
            FROM
                package_quantity_type_mapping`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query package quantity type")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.PackageQuantityTypeMapping{}
		err := rows.Scan(
			&e.PackageType,
			&e.QuantityType,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan package quantity type")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) upsertPackageQuantityTypeMapping(ctx context.Context) error {
	deleteStmt := `DELETE FROM package_quantity_type_mapping
			WHERE package_type = $1`
	_, err := s.FatimaDBTrace.Exec(ctx, deleteStmt, s.convertIntToPackageType(1))
	if err != nil {
		return fmt.Errorf("cannot upsert package quantity type, err: %s", err)
	}

	insertStmt := `INSERT INTO package_quantity_type_mapping(
                package_type,
                quantity_type)
            VALUES ($1,$2) ON CONFLICT DO NOTHING`
	_, err = s.FatimaDBTrace.Exec(ctx, insertStmt, s.convertIntToPackageType(1), s.convertIntToQuantityType(2))
	if err != nil {
		return fmt.Errorf("cannot upsert package quantity type, err: %s", err)
	}
	return nil
}

func (s *suite) convertIntToQuantityType(i int) pb.QuantityType {
	quantityType := map[int]pb.QuantityType{
		1: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
		2: pb.QuantityType_QUANTITY_TYPE_SLOT,
		3: pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK,
	}
	return quantityType[i]
}

func (s *suite) convertIntToPackageType(i int) pb.PackageType {
	packageType := map[int]pb.PackageType{
		1: pb.PackageType_PACKAGE_TYPE_ONE_TIME,
		2: pb.PackageType_PACKAGE_TYPE_SLOT_BASED,
		3: pb.PackageType_PACKAGE_TYPE_FREQUENCY,
		4: pb.PackageType_PACKAGE_TYPE_SCHEDULED,
	}
	return packageType[i]
}
