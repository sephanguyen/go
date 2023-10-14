package payment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) theInvalidPackageLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportProductRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportProductResponse)
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

func (s *suite) theValidPackageLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	allProducts, err := s.selectAllProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, row := range stepState.ValidCsvRows {
		var product entities.Product
		var pkg entities.Package
		values := strings.Split(row, ",")
		product, pkg, err = productAndPackageFromCsv(values)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		findingProduct := foundProducts(product, allProducts)
		if findingProduct.ProductID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found product in list")
		}
		pkg.PackageID = findingProduct.ProductID
		findingPackage := foundPackages(pkg, allPackages)
		if findingPackage.PackageID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found package in list")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportPackageTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	allProducts, err := s.selectAllProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		var product entities.Product
		var pkg entities.Package
		values := strings.Split(row, ",")
		product, pkg, err = productAndPackageFromCsv(values)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		findingProduct := foundProducts(product, allProducts)
		if findingProduct.ProductID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found product in list")
		}
		pkg.PackageID = findingProduct.ProductID
		findingPackage := foundPackages(pkg, allPackages)
		if findingPackage.PackageID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found package in list")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingPackage(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportProduct(contextWithToken(ctx), stepState.Request.(*pb.ImportProductRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anPackageValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomePackages(ctx)
	if err != nil {
		fmt.Printf("error when insert package %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	headerTitles := []string{
		"package_id",
		"name",
		"package_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"max_slot",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"package_start_date",
		"package_end_date",
		"remarks",
		"is_archived",
		"is_unique",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Package %s,2,,,,2021-12-07,2022-10-07,2,2022-10-07,,1,2021-12-07,2022-10-07,Remarks,0,0", idutil.ULIDNow())
	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportProductRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anPackageValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
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
	headerTitles := []string{
		"package_id",
		"name",
		"package_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"max_slot",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"package_start_date",
		"package_end_date",
		"remarks",
		"is_archived",
		"is_unique",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Package %s,2,,,,2021-12-07,2022-10-07,2,2022-10-07,,1,2021-12-07,2022-10-07,Remarks,0,0", idutil.ULIDNow())
	validRow3 := fmt.Sprintf("%s,Package %s,2,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", existingPackages[0].PackageID.String, idutil.ULIDNow())
	validRow4 := fmt.Sprintf("%s,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", existingPackages[1].PackageID.String, idutil.ULIDNow())
	invalidEmptyRow1 := fmt.Sprintf(",Package %s,,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	invalidEmptyRow2 := fmt.Sprintf("%s,Package %s,0,,,,,2022-10-07T00:00:00-07:00,,2,2022-10-07T00:00:00-07:00,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", existingPackages[2].PackageID.String, idutil.ULIDNow())
	invalidValueRow1 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,,,2,2022-10-07T00:00:00-07:00,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	invalidValueRow3 := fmt.Sprintf(",Package %s,1,,,,1221,2022-10-07T00:00:00-07:00,,2,2022-10-07T00:00:00-07:00,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	invalidValueRow4 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,212,,2,2022-10-07T00:00:00-07:00,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	invalidValueRow5 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,fgfgfg,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	invalidValueRow6 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,fgfgfg,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,,Remarks,0,0", idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportProductRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidEmptyRow1, invalidEmptyRow2)),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportProductRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidValueRow1, invalidValueRow2)),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, headerText, validRow1, validRow2, validRow3, validRow4, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5, invalidValueRow6)),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5, invalidValueRow6}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anPackageInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	headerTitles := []string{
		"package_id",
		"name",
		"package_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"max_slot",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"package_start_date",
		"package_end_date",
		"remarks",
		"is_archived",
		"is_unique",
	}
	headerTitleMisfield := []string{
		"package_id",
		"name",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"max_slot",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"package_start_date",
		"package_end_date",
		"remarks",
		"is_unique",
	}
	headerTitlesWrongNameOfID := []string{
		"PackageID123",
		"name",
		"package_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"max_slot",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"package_start_date",
		"package_end_date",
		"remarks",
		"is_archived",
		"is_unique",
	}
	headerText := strings.Join(headerTitles, ",")
	headerTitleMisfieldText := strings.Join(headerTitleMisfield, ",")
	headerTitlesWrongNameOfIDText := strings.Join(headerTitlesWrongNameOfID, ",")
	validRow1 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	invalidRow1 := fmt.Sprintf(",Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,0", idutil.ULIDNow())
	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductRequest{
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
	case "header only":
		stepState.Request = &pb.ImportProductRequest{
			Payload:     []byte(headerText),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
	case "number of column is not equal 15":
		stepState.Request = &pb.ImportProductRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s`, headerTitleMisfieldText, validRow1)),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportProductRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s`, headerText, invalidRow1)),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
	case "wrong package_id column name in header":
		stepState.Request = &pb.ImportProductRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s`, headerTitlesWrongNameOfIDText, validRow1)),
			ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllPackages(ctx context.Context) ([]*entities.Package, error) {
	var allEntities []*entities.Package
	stmt :=
		`
		SELECT 
			package_id,package_type,max_slot,package_start_date,package_end_date
		FROM
			package
		ORDER BY
			package_id ASC
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query discount")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entities.Package{}
		err := rows.Scan(
			&e.PackageID,
			&e.PackageType,
			&e.MaxSlot,
			&e.PackageStartDate,
			&e.PackageEndDate,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan discount")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) selectAllProducts(ctx context.Context) ([]*entities.Product, error) {
	var allEntities []*entities.Product
	const getProducts = `-- name: GetProducts :many
SELECT product_id, name, product_type, tax_id, product_tag, product_partner_id, available_from, available_until, remarks, custom_billing_period, billing_schedule_id, disable_pro_rating_flag, is_archived, updated_at, created_at FROM product
`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		getProducts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query discount")
	}
	defer rows.Close()
	for rows.Next() {
		var i entities.Product
		if err := rows.Scan(
			&i.ProductID,
			&i.Name,
			&i.ProductType,
			&i.TaxID,
			&i.ProductTag,
			&i.ProductPartnerID,
			&i.AvailableFrom,
			&i.AvailableUntil,
			&i.Remarks,
			&i.CustomBillingPeriod,
			&i.BillingScheduleID,
			&i.DisableProRatingFlag,
			&i.IsArchived,
			&i.UpdatedAt,
			&i.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan Product")
		}
		allEntities = append(allEntities, &i)
	}
	return allEntities, nil
}

func (s *suite) insertSomePackages(ctx context.Context) error {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
		ProductTag           sql.NullString `json:"product_tag"`
		ProductPartnerID     sql.NullString `json:"product_partner_id"`
		AvailableFrom        time.Time      `json:"available_from"`
		AvailableUtil        time.Time      `json:"available_until"`
		Remarks              sql.NullString `json:"remarks"`
		CustomBillingPeriod  sql.NullTime   `json:"custom_billing_period"`
		BillingScheduleID    sql.NullString `json:"billing_schedule_id"`
		DisableProRatingFlag bool           `json:"disable_pro_rating_flag"`
		IsArchived           bool           `json:"is_archived"`
		UpdatedAt            time.Time      `json:"updated_at"`
		CreatedAt            time.Time      `json:"created_at"`
	}

	type AddPackageParams struct {
		PackageID        string       `json:"package_id"`
		PackageType      string       `json:"package_type"`
		MaxSlot          int32        `json:"max_slot"`
		PackageStartDate sql.NullTime `json:"package_start_date"`
		PackageEndDate   sql.NullTime `json:"package_end_date"`
	}
	for i := 0; i < 5; i++ {
		var arg AddProductParams
		var packageArg AddPackageParams
		randomStr := idutil.ULIDNow()
		arg.ProductID = randomStr
		arg.Name = fmt.Sprintf("package-%v", randomStr)
		arg.ProductType = pb.ProductType_PRODUCT_TYPE_PACKAGE.String()
		arg.AvailableFrom = time.Now()
		arg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		arg.DisableProRatingFlag = false
		arg.IsArchived = false

		stmt := `INSERT INTO product(
					product_id,
                    name,
                    product_type,
                    tax_id,
                    product_tag,
                    product_partner_id,
                    available_from,
                    available_until,
                    remarks,
                    custom_billing_period,
                    billing_schedule_id,
                    disable_pro_rating_flag,
                    is_archived,
                    updated_at,
                    created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now(), now())
                RETURNING product_id`
		_, err := s.FatimaDBTrace.Exec(ctx, stmt,
			arg.ProductID,
			arg.Name,
			arg.ProductType,
			arg.TaxID,
			arg.ProductTag,
			arg.ProductPartnerID,
			arg.AvailableFrom,
			arg.AvailableUtil,
			arg.Remarks,
			arg.CustomBillingPeriod,
			arg.BillingScheduleID,
			arg.DisableProRatingFlag,
			arg.IsArchived)

		if err != nil {
			return fmt.Errorf("cannot insert product, err: %s", err)
		}

		querySelectProduct := `SELECT product_id
                            FROM
                                product
                            WHERE
                                name = $1`

		row := s.FatimaDBTrace.QueryRow(ctx, querySelectProduct, arg.Name)
		err = row.Scan(&packageArg.PackageID)

		if err != nil {
			return fmt.Errorf("cannot insert product, err: %s", err)
		}

		queryInsertPackage := `INSERT INTO package(
									package_id,
                                    package_type,
                                    max_slot,
                                    package_start_date,
                                    package_end_date)
                                VALUES ($1, $2, $3, $4, $5)`

		packageArg.PackageType = pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()
		packageArg.MaxSlot = 34
		packageArg.PackageStartDate = sql.NullTime{Time: time.Now(), Valid: true}
		packageArg.PackageEndDate = sql.NullTime{Time: time.Now().AddDate(1, 0, 0), Valid: true}
		_, err = s.FatimaDBTrace.Exec(ctx, queryInsertPackage,
			packageArg.PackageID,
			packageArg.PackageType,
			packageArg.MaxSlot,
			packageArg.PackageStartDate,
			packageArg.PackageEndDate,
		)
		if err != nil {
			return fmt.Errorf("cannot insert package, err: %s", err)
		}

		productSetting := entities.ProductSetting{
			ProductID:                    pgtype.Text{String: packageArg.PackageID, Status: pgtype.Present},
			IsPausable:                   pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsEnrollmentRequired:         pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsOperationFee:               pgtype.Bool{Bool: false, Status: pgtype.Present},
		}
		err = mockdata.InsertProductSetting(ctx, s.FatimaDBTrace, productSetting)
		if err != nil {
			return fmt.Errorf("cannot insert default product_setting for fee product, err: %s", err)
		}
	}
	return nil
}

func productAndPackageFromCsv(line []string) (product entities.Product, pkg entities.Package, err error) {
	const (
		PackageID = iota
		Name
		PackageType
		TaxID
		ProductTag
		ProductPartnerID
		AvailableFrom
		AvailableUntil
		MaxSlot
		CustomBillingPeriod
		BillingScheduleID
		DisableProRatingFlag
		PackageStartDate
		PackageEndDate
		Remarks
		IsArchived
	)

	if err = multierr.Combine(
		utils.StringToFormatString("ID", line[PackageID], true, product.ProductID.Set),
		utils.StringToFormatString("name", line[Name], false, product.Name.Set),
		product.ProductType.Set(pb.ProductType_PRODUCT_TYPE_PACKAGE),
		utils.StringToFormatString("tax_id", line[TaxID], true, product.TaxID.Set),
		utils.StringToFormatString("product_tag", line[ProductTag], true, product.ProductTag.Set),
		utils.StringToFormatString("product_partner_id", line[ProductPartnerID], true, product.ProductPartnerID.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, product.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, product.AvailableUntil.Set),
		utils.StringToDate("custom_billing_period", line[CustomBillingPeriod], true, product.CustomBillingPeriod.Set),
		utils.StringToFormatString("billing_schedule_id", line[BillingScheduleID], true, product.BillingScheduleID.Set),
		utils.StringToBool("disable_pro_rating_flag", line[DisableProRatingFlag], true, product.DisableProRatingFlag.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, product.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], true, product.IsArchived.Set),
	); err != nil {
		return product, pkg, err
	}

	if err = multierr.Combine(
		utils.StringToFormatString("package_id", line[PackageID], true, pkg.PackageID.Set),
		utils.StringToPackageType("package_type", line[PackageType], pkg.PackageType.Set),
		utils.StringToInt("max_slot", line[MaxSlot], true, pkg.MaxSlot.Set),
		utils.StringToDate("package_start_date", line[PackageStartDate], true, pkg.PackageStartDate.Set),
		utils.StringToDate("package_end_date", line[PackageEndDate], true, pkg.PackageEndDate.Set),
	); err != nil {
		return product, pkg, err
	}

	return product, pkg, nil
}

func foundProducts(productNeedFinding entities.Product, productList []*entities.Product) (finding entities.Product) {
	for i, product := range productList {
		if productNeedFinding.Name == product.Name &&
			productNeedFinding.ProductType == product.ProductType &&
			productNeedFinding.Remarks == product.Remarks &&
			productNeedFinding.IsArchived == product.IsArchived &&
			productNeedFinding.DisableProRatingFlag == product.DisableProRatingFlag &&
			productNeedFinding.TaxID == product.TaxID &&
			productNeedFinding.ProductTag == product.ProductTag &&
			productNeedFinding.ProductPartnerID == product.ProductPartnerID &&
			productNeedFinding.AvailableFrom == product.AvailableFrom &&
			productNeedFinding.AvailableUntil == product.AvailableUntil &&
			productNeedFinding.CustomBillingPeriod == product.CustomBillingPeriod &&
			productNeedFinding.BillingScheduleID == product.BillingScheduleID {
			finding = *productList[i]
			break
		}
	}
	return finding
}

func foundPackages(packageNeedFinding entities.Package, packageList []*entities.Package) (finding entities.Package) {
	for i, pkg := range packageList {
		if packageNeedFinding.PackageID == pkg.PackageID &&
			packageNeedFinding.MaxSlot == pkg.MaxSlot &&
			packageNeedFinding.PackageStartDate == pkg.PackageStartDate &&
			packageNeedFinding.PackageEndDate == pkg.PackageEndDate &&
			packageNeedFinding.PackageType == pkg.PackageType {
			finding = *packageList[i]
			break
		}
	}
	return finding
}
