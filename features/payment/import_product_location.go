package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) anProductLocationsValidRequestPayloadWithCorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.insertSomePackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = s.insertSomeLocations(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingLocations, err := s.selectAllLocations(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	headerTitles := []string{
		"product_id",
		"location_id",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf("%s,%s", existingPackages[0].PackageID.String, existingLocations[0].LocationID.String)
	validRow2 := fmt.Sprintf("%s,%s", existingPackages[0].PackageID.String, existingLocations[1].LocationID.String)
	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anProductLocationsValidRequestPayloadWithIncorrectDataWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.insertSomePackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = s.insertSomeLocations(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingLocations, err := s.selectAllLocations(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	headerTitles := []string{
		"product_id",
		"location_id",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf("%s,%s", existingPackages[0].PackageID.String, existingLocations[0].LocationID.String)
	validRow2 := fmt.Sprintf("%s,%s", existingPackages[0].PackageID.String, existingLocations[1].LocationID.String)
	validRow3 := fmt.Sprintf("%s,%s", existingPackages[1].PackageID.String, existingLocations[0].LocationID.String)
	validRow4 := fmt.Sprintf("%s,%s", existingPackages[1].PackageID.String, existingLocations[1].LocationID.String)
	invalidEmptyRow1 := fmt.Sprintf(",%s", existingLocations[0].LocationID.String)
	invalidEmptyRow2 := fmt.Sprintf("%s,", existingPackages[0].PackageID.String)
	invalidValueRow1 := fmt.Sprintf("sd,%s", existingLocations[0].LocationID.String)
	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidEmptyRow1, invalidEmptyRow2)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(
				`%s
				%s`,
				headerText, invalidValueRow1,
			)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, headerText, validRow1, validRow2, validRow3, validRow4, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anProductLocationsInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	headerTitles := []string{
		"product_id",
		"location_id",
	}
	headerTitleMisfield := []string{
		"product_id",
	}
	headerTitlesWrongNameOfID := []string{
		"product_idabc",
		"location_id",
	}
	headerText := strings.Join(headerTitles, ",")
	headerTitleMisfieldText := strings.Join(headerTitleMisfield, ",")
	headerTitlesWrongNameOfIDText := strings.Join(headerTitlesWrongNameOfID, ",")
	validRow1 := fmt.Sprintf("1,Location-%s", idutil.ULIDNow())
	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
	case "header only":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload:                   []byte(headerText),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
	case "number of column is not equal 2":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s`, headerTitleMisfieldText, validRow1)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
	case "wrong product_id column name in header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s`, headerTitlesWrongNameOfIDText, validRow1)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingProductLocations(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).ImportProductAssociatedData(
		contextWithToken(ctx), stepState.Request.(*pb.ImportProductAssociatedDataRequest),
	)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidProductLocationsLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductLocations, err := s.selectAllProductLocations(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	const (
		ProductID = iota
		LocationID
	)
	for _, row := range stepState.ValidCsvRows {
		var productLocation entities.ProductLocation
		values := strings.Split(row, ",")

		err = multierr.Combine(
			utils.StringToFormatString("product_id", values[ProductID], false, productLocation.ProductID.Set),
			utils.StringToFormatString("location_id", values[LocationID], false, productLocation.LocationID.Set),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		findingProduct := foundProductLocations(productLocation, allProductLocations)
		if findingProduct.ProductID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found product in list")
		}
	}

	resp := stepState.Response.(*pb.ImportProductAssociatedDataResponse)
	if len(stepState.ValidCsvRows) != 0 && len(stepState.InvalidCsvRows) == 0 {
		if len(resp.Errors) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error returned")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportProductLocationTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductLocations, err := s.selectAllProductLocations(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	const (
		ProductID = iota
		LocationID
	)
	for _, row := range stepState.ValidCsvRows {
		var productLocation entities.ProductLocation
		values := strings.Split(row, ",")

		err = multierr.Combine(
			utils.StringToFormatString("product_id", values[ProductID], false, productLocation.ProductID.Set),
			utils.StringToFormatString("location_id", values[LocationID], false, productLocation.LocationID.Set),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		findingProduct := foundProductLocations(productLocation, allProductLocations)
		if findingProduct.ProductID.Get() != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to rollback valid csv row")
		}
	}

	resp := stepState.Response.(*pb.ImportProductAssociatedDataResponse)
	if len(stepState.ValidCsvRows) != 0 && len(stepState.InvalidCsvRows) == 0 {
		if len(resp.Errors) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error returned")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidProductLocationsLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
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

func (s *suite) selectAllProductLocations(ctx context.Context) ([]*entities.ProductLocation, error) {
	var allEntities []*entities.ProductLocation
	const stmt = `
		SELECT
			pl.product_id,
			pl.location_id,
			pl.created_at 
		FROM product_location pl
	`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query product_location")
	}
	defer rows.Close()
	for rows.Next() {
		var entity entities.ProductLocation
		if err := rows.Scan(
			&entity.ProductID,
			&entity.LocationID,
			&entity.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan Product")
		}
		allEntities = append(allEntities, &entity)
	}
	return allEntities, nil
}

func foundProductLocations(productLocationNeedFinding entities.ProductLocation, productLocationsList []*entities.ProductLocation) (finding entities.ProductLocation) {
	for i, productLocation := range productLocationsList {
		if productLocationNeedFinding.ProductID == productLocation.ProductID &&
			productLocationNeedFinding.LocationID == productLocation.LocationID {
			finding = *productLocationsList[i]
			break
		}
	}
	return finding
}

func (s *suite) insertSomeLocations(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		id := idutil.ULIDNow()
		locationName := fmt.Sprintf("location_name_%d", i)
		stmt := `INSERT INTO locations
		(location_id, name, created_at, updated_at)
		VALUES ($1, $2, now(), now())`
		_, err := s.FatimaDBTrace.Exec(ctx, stmt, id, locationName)
		if err != nil {
			return fmt.Errorf("cannot insert location: %v", err.Error())
		}
	}
	return nil
}

func (s *suite) selectAllLocations(ctx context.Context) ([]*entities.Location, error) {
	var allEntities []*entities.Location
	stmt :=
		`
		SELECT 
			location_id
		FROM
			locations
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query location")
	}
	defer rows.Close()
	for rows.Next() {
		e := &entities.Location{}
		err := rows.Scan(
			&e.LocationID,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan location")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}
