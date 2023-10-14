package payment

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	export "github.com/manabie-com/backend/internal/payment/services/export_service"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"go.uber.org/multierr"
)

func (s *suite) addDataForExportMasterData(ctx context.Context, dataType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	taxID, err := mockdata.InsertOneTax(ctx, s.FatimaDBTrace, "tax_name")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	switch dataType {
	case "accounting category":
		err = s.insertSomeAccountingCategories(ctx)
	case "billing ratio":
		err = s.insertBillingRatioForExport(ctx)
	case "billing schedule":
		err = s.insertSomeBillingSchedules(ctx)
	case "billing schedule period":
		err = s.insertSomeBillingSchedulePeriods(ctx)
	case "discount":
		_, err = mockdata.InsertOneDiscountAmount(ctx, s.FatimaDBTrace, "discount_name")
	case "fee":
		err = s.insertSomeFees(ctx)
	case "leaving reason":
		err = s.insertLeavingReasons(ctx)
	case "material":
		_, err = s.insertMaterial(ctx, taxID)
	case "product setting":
		err = s.insertProductSettingForExport(ctx)
	case "product discount":
		err = s.insertProductDiscountForExport(ctx)
	case "product location":
		err = s.insertProductLocationForExport(ctx, taxID)
	case "product grade":
		err = s.insertProductGradeForExport(ctx)
	case "package course", "product accounting category", "package quantity type mapping", "product price":
		err = s.insertPackageAssociationForExport(ctx, taxID)
	case "package":
		err = s.insertSomePackages(ctx)
	case "notification date":
		err = s.insertNotificationDates(ctx)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theUserExportMasterData(ctx context.Context, user string, dataType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	clientExport := pb.NewExportServiceClient(s.PaymentConn)
	var req *pb.ExportMasterDataRequest
	switch dataType {
	case "accounting category":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_ACCOUNTING_CATEGORY,
		}
	case "billing ratio":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_BILLING_RATIO,
		}
	case "billing schedule":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_BILLING_SCHEDULE,
		}
	case "billing schedule period":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_BILLING_SCHEDULE_PERIOD,
		}
	case "discount":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_DISCOUNT,
		}
	case "fee":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_FEE,
		}
	case "leaving reason":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_LEAVING_REASON,
		}
	case "material":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_MATERIAL,
		}
	case "tax":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_TAX,
		}
	case "product setting":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PRODUCT_SETTING,
		}
	case "product price":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PRODUCT_PRICE,
		}
	case "product discount":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PRODUCT_DISCOUNT,
		}
	case "product location":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PRODUCT_ASSOCIATED_LOCATION,
		}
	case "product grade":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PRODUCT_ASSOCIATED_GRADE,
		}
	case "product accounting category":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PRODUCT_ASSOCIATED_ACCOUNTING_CATEGORY,
		}
	case "package quantity type mapping":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PACKAGE_QUANTITY_TYPE_MAPPING,
		}
	case "package course":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PACKAGE_ASSOCIATED_COURSE,
		}
	case "package":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_PACKAGE,
		}
	case "notification date":
		req = &pb.ExportMasterDataRequest{
			ExportDataType: pb.ExportMasterDataType_EXPORT_NOTIFICATION_DATE,
		}
	}

	resp, err := clientExport.ExportMasterData(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = req
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theMasterDataCSVHasCorrectContent(ctx context.Context, dataType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*pb.ExportMasterDataRequest)
	response := stepState.Response.(*pb.ExportMasterDataResponse)

	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be greater than 1
	if len(lines) < 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the context line to be greater than or equal to 1 got %d", len(lines))
	}
	colMap, _ := export.GetExportColMapAndEntityType(request.ExportDataType)
	csvColumn := make([]string, 0, len(colMap))
	for _, col := range colMap {
		csvColumn = append(csvColumn, col.CSVColumn)
	}
	// check the header record
	err = checkCSVHeaderForExport(
		csvColumn,
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if dataType == "product price" {
		err = s.validateDataConversionInProductPrice(ctx, lines[1])
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertLeavingReasons(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		randomStr := idutil.ULIDNow()
		name := fmt.Sprintf("Cat " + randomStr)
		leavingReasonType := database.Text("1")
		remarks := fmt.Sprintf("Remark " + randomStr)
		isArchived := true
		stmt := `INSERT INTO leaving_reason
		(leaving_reason_id, name, leaving_reason_type,  remark, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())`

		_, err := s.FatimaDBTrace.Exec(ctx, stmt, randomStr, name, leavingReasonType, remarks, isArchived)
		if err != nil {
			return fmt.Errorf("cannot insert leaving_reason, err: %s", err)
		}
	}
	return nil
}

func (s *suite) insertNotificationDates(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		notificationDateID := idutil.ULIDNow()
		orderType := pb.OrderType_ORDER_TYPE_NEW.String()
		notificationDate := 10
		isArchived := true
		stmt := `INSERT INTO notification_date
		(notification_date_id, order_type, notification_date, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())`

		_, err := s.FatimaDBTrace.Exec(ctx, stmt, notificationDateID, orderType, notificationDate, isArchived)
		if err != nil {
			return fmt.Errorf("cannot insert notification_date, err: %s", err)
		}
	}
	return nil
}

func (s *suite) insertProductDiscountForExport(ctx context.Context) (err error) {
	err = s.insertSomePackages(ctx)
	if err != nil {
		return
	}
	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return
	}

	err = s.insertSomeDiscounts(ctx)
	if err != nil {
		return
	}

	existingDiscounts, err := s.selectAllDiscounts(ctx)
	if err != nil {
		return
	}
	err = s.insertSomeProductAssociationDataDiscount(
		ctx,
		existingPackages[len(existingPackages)-1].PackageID.String,
		existingDiscounts[len(existingDiscounts)-1].DiscountID.String)
	if err != nil {
		return
	}
	return
}

func (s *suite) insertPackageAssociationForExport(ctx context.Context, taxID string) (err error) {
	productIDs, err := s.insertPackage(ctx, taxID)
	if err != nil {
		return
	}
	err = mockdata.InsertPackageTypeQuantityTypeMapping(ctx, s.FatimaDBTrace)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"package_quantity_type_mapping_pk\"") {
			return
		}
	}

	err = s.insertProductPriceForPackage(ctx, productIDs)
	if err != nil {
		return
	}

	courseIDs, err := s.insertCourses(ctx)
	if err != nil {
		err = fmt.Errorf("error when insert course %v", err.Error())
		return
	}

	err = s.insertPackageCourses(ctx, productIDs, courseIDs)
	if err != nil {
		return
	}
	return
}

func (s *suite) insertProductGradeForExport(ctx context.Context) (err error) {
	gradeID, err := mockdata.InsertOneGrade(ctx, s.FatimaDBTrace)
	if err != nil {
		err = fmt.Errorf("error when insert grade %v", err.Error())
		return
	}
	err = s.insertSomePackages(ctx)
	if err != nil {
		err = fmt.Errorf("error when insert package %v", err.Error())
		return
	}

	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return
	}
	packageIDs := []string{existingPackages[len(existingPackages)-1].PackageID.String, existingPackages[len(existingPackages)-2].PackageID.String}
	err = s.insertProductGrade(ctx, gradeID, packageIDs)
	if err != nil {
		return
	}
	return
}

func (s *suite) insertProductLocationForExport(ctx context.Context, taxID string) (err error) {
	productIDs, err := s.insertPackage(ctx, taxID)
	if err != nil {
		return
	}
	locationID := constants.ManabieOrgLocation

	err = s.insertProductLocation(ctx, locationID, productIDs)
	if err != nil {
		return
	}
	return
}

func (s *suite) insertProductSettingForExport(ctx context.Context) (err error) {
	productIDs, err := s.insertSomeProducts(ctx)
	if err != nil || len(productIDs) < 2 {
		err = fmt.Errorf("error inserting mock products for product setting test, err: %s", err)
		return
	}
	for _, productID := range productIDs {
		err = s.insertProductSetting(ctx, productID)
		if err != nil {
			return
		}
	}
	return
}

func (s *suite) insertBillingRatioForExport(ctx context.Context) (err error) {
	err = s.insertSomeBillingSchedulePeriods(ctx)
	if err != nil {
		return
	}
	err = s.insertSomeBillingRatios(ctx)
	if err != nil {
		return
	}
	return
}

func (s *suite) validateDataConversionInProductPrice(ctx context.Context, line []string) error {
	exportedValues, err := s.productPriceFromCsv(line)
	if err != nil {
		return err
	}

	databaseValues, err := s.getProductPriceByProductID(ctx, exportedValues.ProductID.String)
	if err != nil {
		return err
	}

	var floatValExported float64
	_ = exportedValues.Price.AssignTo(&floatValExported)

	var floatValDBValue float64
	_ = databaseValues.Price.AssignTo(&floatValDBValue)

	if floatValExported != floatValDBValue {
		return fmt.Errorf("incorrect numeric conversion for product price")
	}

	return nil
}

func (s *suite) productPriceFromCsv(line []string) (productPrice entities.ProductPrice, err error) {
	const (
		ProductPriceID = iota
		ProductID
		BillingSchedulePeriodID
		Quantity
		Price
		PriceType
	)

	if err = multierr.Combine(
		utils.StringToInt("product_price_id", line[ProductPriceID], false, productPrice.ProductPriceID.Set),
		utils.StringToFormatString("product_id", line[ProductID], false, productPrice.ProductID.Set),
		utils.StringToFormatString("billing_schedule_period_id", line[BillingSchedulePeriodID], true, productPrice.BillingSchedulePeriodID.Set),
		utils.StringToInt("quantity", line[Quantity], true, productPrice.Quantity.Set),
		utils.StringToFloat("price", line[Price], false, productPrice.Price.Set),
		utils.StringToFloat("price_type", line[PriceType], false, productPrice.PriceType.Set),
	); err != nil {
		return productPrice, err
	}

	return productPrice, nil
}

func (s *suite) getProductPriceByProductID(ctx context.Context, productID string) (productPrice entities.ProductPrice, err error) {
	productPriceFieldNames, productPriceFieldValues := productPrice.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productPriceFieldNames, ","),
		"public.product_price",
	)

	row := s.FatimaDBTrace.QueryRow(ctx, stmt, productID)
	err = row.Scan(productPriceFieldValues...)
	if err != nil {
		err = fmt.Errorf("row.Scan: %w", err)
		return
	}

	return
}
