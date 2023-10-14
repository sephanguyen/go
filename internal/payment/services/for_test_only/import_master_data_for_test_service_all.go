package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	importsvc "github.com/manabie-com/backend/internal/payment/services/import_service"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type insertEntityFunc func(*ImportMasterDataForTestService, context.Context, pgx.Tx, []string) error

var insertAccountingCategory insertEntityFunc = (*ImportMasterDataForTestService).insertAccountingCategory
var insertBillingRatio insertEntityFunc = (*ImportMasterDataForTestService).insertBillingRatio
var insertBillingSchedule insertEntityFunc = (*ImportMasterDataForTestService).insertBillingSchedule
var insertBillingSchedulePeriod insertEntityFunc = (*ImportMasterDataForTestService).insertBillingSchedulePeriod
var insertDiscount insertEntityFunc = (*ImportMasterDataForTestService).insertDiscount
var insertTax insertEntityFunc = (*ImportMasterDataForTestService).insertTax
var insertFee insertEntityFunc = (*ImportMasterDataForTestService).insertFee
var insertMaterial insertEntityFunc = (*ImportMasterDataForTestService).insertMaterial
var insertPackage insertEntityFunc = (*ImportMasterDataForTestService).insertPackage
var insertProductAccountingCategory insertEntityFunc = (*ImportMasterDataForTestService).insertProductAccountingCategory
var insertProductGrade insertEntityFunc = (*ImportMasterDataForTestService).insertProductGrade
var insertProductLocation insertEntityFunc = (*ImportMasterDataForTestService).insertProductLocation
var insertProductDiscount insertEntityFunc = (*ImportMasterDataForTestService).insertProductDiscount
var insertProductPrice insertEntityFunc = (*ImportMasterDataForTestService).insertProductPrice
var insertProductSetting insertEntityFunc = (*ImportMasterDataForTestService).insertProductSetting
var insertPackageCourse insertEntityFunc = (*ImportMasterDataForTestService).insertPackageCourse
var insertPackageCourseFee insertEntityFunc = (*ImportMasterDataForTestService).insertPackageCourseFee
var insertPackageCourseMaterial insertEntityFunc = (*ImportMasterDataForTestService).insertPackageCourseMaterial
var insertLeavingReason insertEntityFunc = (*ImportMasterDataForTestService).insertLeavingReason
var insertNotificationDate insertEntityFunc = (*ImportMasterDataForTestService).insertNotificationDate

func (s *ImportMasterDataForTestService) ImportAllForTest(ctx context.Context, req *pb.ImportAllForTestRequest) (*pb.ImportAllForTestResponse, error) {
	errLines := []*pb.ImportAllForTestResponse_ImportAllForTestError{}
	var insertFunc insertEntityFunc
	var entityName string
	var rowNumber uint16

	bytesReader := bytes.NewReader(req.Payload)
	bufReader := bufio.NewReader(bytesReader)
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		var validateHeaderErr error
		for {
			if validateHeaderErr == io.EOF {
				break
			} else if err != nil {
				return status.Error(codes.InvalidArgument, "validateHeader err: "+err.Error())
			}
			rowNumber++
			lineByte, _, err := bufReader.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				return status.Error(codes.InvalidArgument, "bufReader.ReadLine() err: "+err.Error())
			}
			// lbs is line before splited into csv format
			lbs := strings.TrimSpace(string(lineByte))

			switch lbs {
			case "-- accounting_category":
				insertFunc = insertAccountingCategory
				entityName = "accounting_category"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- billing_ratio":
				insertFunc = insertBillingRatio
				entityName = "billing_ratio"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- billing_schedule":
				insertFunc = insertBillingSchedule
				entityName = "billing_schedule"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- billing_schedule_period":
				insertFunc = insertBillingSchedulePeriod
				entityName = "billing_schedule_period"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- discount":
				insertFunc = insertDiscount
				entityName = "discount"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- fee":
				insertFunc = insertFee
				entityName = "fee"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- material":
				insertFunc = insertMaterial
				entityName = "material"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- package":
				insertFunc = insertPackage
				entityName = "package"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- tax":
				insertFunc = insertTax
				entityName = "tax"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- leaving_reason":
				insertFunc = insertLeavingReason
				entityName = "leaving_reason"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- product_accounting_category":
				insertFunc = insertProductAccountingCategory
				entityName = "product_accounting_category"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- product_discount":
				insertFunc = insertProductDiscount
				entityName = "product_discount"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- product_grade":
				insertFunc = insertProductGrade
				entityName = "product_grade"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- product_location":
				insertFunc = insertProductLocation
				entityName = "product_location"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- product_price":
				insertFunc = insertProductPrice
				entityName = "product_price"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- product_setting":
				insertFunc = insertProductSetting
				entityName = "product_setting"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- package_course":
				insertFunc = insertPackageCourse
				entityName = "package_course"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- package_course_fee":
				insertFunc = insertPackageCourseFee
				entityName = "package_course_fee"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- package_course_material":
				insertFunc = insertPackageCourseMaterial
				entityName = "package_course_material"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "-- notification_date":
				insertFunc = insertNotificationDate
				entityName = "notification_date"
				validateHeaderErr = validateHeader(bufReader, entityName)
				rowNumber++
				continue
			case "":
				continue
			}

			// line is array of string in csv format
			line := strings.Split(lbs, ",")

			errLine := insertFunc(s, ctx, tx, line)
			if errLine != nil {
				errLines = append(errLines, &pb.ImportAllForTestResponse_ImportAllForTestError{
					EntityName: entityName,
					RowNumber:  int32(rowNumber),
					Error:      errLine.Error(),
				})
			}
		}
		return nil
	})

	return &pb.ImportAllForTestResponse{
		Errors: errLines,
	}, err
}

func (s *ImportMasterDataForTestService) insertAccountingCategory(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toAccountingCategory(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertBillingRatio(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toBillingRatio(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertBillingSchedule(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toBillingSchedule(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertBillingSchedulePeriod(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toBillingSchedulePeriod(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertDiscount(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toDiscount(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertTax(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toTax(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertFee(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toFee(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	if err = s.ForTestRepo.Insert(ctx, tx, &ent.Product, []string{"resource_path"}); err != nil {
		return err
	}
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertMaterial(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toMaterial(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	if err = s.ForTestRepo.Insert(ctx, tx, &ent.Product, []string{"resource_path"}); err != nil {
		return err
	}
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertPackage(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toPackage(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	if err = s.ForTestRepo.Insert(ctx, tx, &ent.Product, []string{"resource_path"}); err != nil {
		return err
	}
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertLeavingReason(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toLeavingReason(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertProductSetting(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toProductSetting(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	_ = ent.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertProductPrice(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toProductPrice(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"product_price_id", "resource_path"})
}

func (s *ImportMasterDataForTestService) insertProductAccountingCategory(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toProductAccountingCategory(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertProductGrade(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toProductGrade(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertProductLocation(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toProductLocation(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertProductDiscount(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toProductDiscount(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertPackageCourse(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toPackageCourse(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, &ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertPackageCourseFee(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toPackageCourseFee(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertPackageCourseMaterial(ctx context.Context, tx pgx.Tx, line []string) error {
	ent, err := toPackageCourseMaterial(line)
	if err != nil {
		return err
	}
	_ = ent.CreatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, ent, []string{"resource_path"})
}

func (s *ImportMasterDataForTestService) insertNotificationDate(ctx context.Context, tx pgx.Tx, line []string) error {
	notiDate, err := toNotificationDate(line)
	if err != nil {
		return err
	}
	_ = notiDate.CreatedAt.Set(time.Now())
	_ = notiDate.UpdatedAt.Set(time.Now())
	return s.ForTestRepo.Insert(ctx, tx, notiDate, []string{"resource_path"})
}

func toAccountingCategory(line []string) (*entities.AccountingCategory, error) {
	headerTitles, err := getHeaderTitles("accounting_category")
	if err != nil {
		return nil, err
	}
	return importsvc.AccountingCategoryFromCsv(line, headerTitles)
}

func toBillingRatio(line []string) (*entities.BillingRatio, error) {
	headerTitles, err := getHeaderTitles("billing_ratio")
	if err != nil {
		return nil, err
	}
	return importsvc.CreateBillingRatioEntityFromCsv(line, headerTitles)
}

func toBillingSchedule(line []string) (*entities.BillingSchedule, error) {
	headerTitles, err := getHeaderTitles("billing_schedule")
	if err != nil {
		return nil, err
	}
	return importsvc.BillingScheduleFromCsv(line, headerTitles)
}

func toBillingSchedulePeriod(line []string) (*entities.BillingSchedulePeriod, error) {
	headerTitles, err := getHeaderTitles("billing_schedule_period")
	if err != nil {
		return nil, err
	}
	return importsvc.ReadBillingSchedulePeriodFromCsv(line, headerTitles)
}

func toDiscount(line []string) (*entities.Discount, error) {
	headerTitles, err := getHeaderTitles("discount")
	if err != nil {
		return nil, err
	}
	return importsvc.DiscountFromCsv(line, headerTitles)
}

func toTax(line []string) (*entities.Tax, error) {
	return importsvc.CreateTaxEntityFromCsv(line)
}

func toFee(line []string) (entities.Fee, error) {
	headerTitles, err := getHeaderTitles("fee")
	if err != nil {
		return entities.Fee{}, err
	}
	return importsvc.ReadFeeFromCsv(line, headerTitles)
}

func toMaterial(line []string) (entities.Material, error) {
	headerTitles, err := getHeaderTitles("material")
	if err != nil {
		return entities.Material{}, err
	}
	return importsvc.ReadMaterialFromCsv(line, headerTitles)
}

func toPackage(line []string) (entities.Package, error) {
	return importsvc.ProductAndPackageFromCsv(line)
}

func toProductAccountingCategory(line []string) (*entities.ProductAccountingCategory, error) {
	return importsvc.CreateProductAccountingCategoryFromCsv(line)
}

func toProductGrade(line []string) (*entities.ProductGrade, error) {
	return importsvc.CreateProductGradeFromCsv(line)
}

func toProductLocation(line []string) (entities.ProductLocation, error) {
	return importsvc.ReadProductLocationFromCsv(line)
}

func toProductDiscount(line []string) (*entities.ProductDiscount, error) {
	return importsvc.ProductDiscountFromCsv(line)
}

func toProductPrice(line []string) (entities.ProductPrice, error) {
	return importsvc.ProductPriceFromCsv(line)
}

func toProductSetting(line []string) (entities.ProductSetting, error) {
	return importsvc.ProductSettingFromCsv(line)
}

func toPackageCourse(line []string) (entities.PackageCourse, error) {
	return importsvc.PackageCourseFromCsv(line)
}

func toPackageCourseFee(line []string) (*entities.PackageCourseFee, error) {
	return importsvc.ConvertToAssociatedProductByFeeFromCsv(line)
}

func toPackageCourseMaterial(line []string) (*entities.PackageCourseMaterial, error) {
	return importsvc.ConvertToAssociatedProductByMaterialFromCsv(line)
}

func toLeavingReason(line []string) (entities.LeavingReason, error) {
	return importsvc.LeavingReasonFromCsv(line)
}

func toNotificationDate(line []string) (*entities.NotificationDate, error) {
	return importsvc.NotificationDateEntityFromCsv(line)
}

func validateHeader(bufReader *bufio.Reader, entityName string) error {
	headerByte, _, err := bufReader.ReadLine()
	if err != nil {
		return err
	}
	header := strings.Split(strings.TrimSpace(string(headerByte)), ",")

	headerTitles, err := getHeaderTitles(entityName)
	if err != nil {
		return err
	}

	err = utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%s - csv file invalid format - %s", entityName, err.Error()))
	}
	return nil
}

func getHeaderTitles(entityName string) ([]string, error) {
	var headerTitles []string
	switch entityName {
	case "accounting_category":
		headerTitles = []string{
			"accounting_category_id",
			"name",
			"remarks",
			"is_archived",
		}
	case "billing_ratio":
		headerTitles = []string{
			"billing_ratio_id",
			"start_date",
			"end_date",
			"billing_schedule_period_id",
			"billing_ratio_numerator",
			"billing_ratio_denominator",
			"is_archived",
		}
	case "billing_schedule":
		headerTitles = []string{
			"billing_schedule_id",
			"name",
			"remarks",
			"is_archived",
		}
	case "billing_schedule_period":
		headerTitles = []string{
			"billing_schedule_period_id",
			"name",
			"billing_schedule_id",
			"start_date",
			"end_date",
			"billing_date",
			"remarks",
			"is_archived",
		}
	case "discount":
		headerTitles = []string{
			"discount_id",
			"name",
			"discount_type",
			"discount_amount_type",
			"discount_amount_value",
			"recurring_valid_duration",
			"available_from",
			"available_until",
			"remarks",
			"is_archived",
			"student_tag_id_validation",
			"parent_tag_id_validation",
			"discount_tag_id",
		}
	case "fee":
		headerTitles = []string{
			"fee_id",
			"name",
			"fee_type",
			"tax_id",
			"available_from",
			"available_until",
			"custom_billing_period",
			"billing_schedule_id",
			"disable_pro_rating_flag",
			"remarks",
			"is_archived",
			"is_unique",
		}
	case "material":
		headerTitles = []string{
			"material_id",
			"name",
			"material_type",
			"tax_id",
			"available_from",
			"available_until",
			"custom_billing_period",
			"custom_billing_date",
			"billing_schedule_id",
			"disable_pro_rating_flag",
			"remarks",
			"is_archived",
			"is_unique",
		}
	case "package":
		headerTitles = []string{
			"package_id",
			"name",
			"package_type",
			"tax_id",
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
	case "tax":
		headerTitles = []string{
			"tax_id",
			"name",
			"tax_percentage",
			"tax_category",
			"default_flag",
			"is_archived",
		}
	case "leaving_reason":
		headerTitles = []string{
			"leaving_reason_id",
			"name",
			"leaving_reason_type",
			"remark",
			"is_archived",
		}
	case "product_accounting_category":
		headerTitles = []string{
			"product_id",
			"accounting_category_id",
		}
	case "product_discount":
		headerTitles = []string{
			"product_id",
			"discount_id",
		}
	case "product_grade":
		headerTitles = []string{
			"product_id",
			"grade_id",
		}
	case "product_location":
		headerTitles = []string{
			"product_id",
			"location_id",
		}
	case "product_price":
		headerTitles = []string{
			"product_id",
			"billing_schedule_period_id",
			"quantity",
			"price",
		}
	case "product_setting":
		headerTitles = []string{
			"product_id",
			"is_enrollment_required",
		}
	case "package_course":
		headerTitles = []string{
			"package_id",
			"course_id",
			"mandatory_flag",
			"max_slots_per_course",
			"course_weight",
		}
	case "package_course_fee":
		headerTitles = []string{
			"package_id",
			"course_id",
			"fee_id",
			"available_from",
			"available_until",
		}
	case "package_course_material":
		headerTitles = []string{
			"package_id",
			"course_id",
			"material_id",
			"available_from",
			"available_until",
		}
	case "notification_date":
		headerTitles = []string{
			"notification_date_id",
			"order_type",
			"notification_date",
			"is_archived",
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "entity name not supported")
	}

	return headerTitles, nil
}

type ForTestRepo struct{}

func (s *ForTestRepo) Insert(ctx context.Context, db database.QueryExecer, e database.Entity, excludedFields []string) error {
	cmdTag, err := database.InsertExcept(ctx, e, excludedFields, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert %s: %w", e.TableName(), err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert %s: %d RowsAffected", e.TableName(), cmdTag.RowsAffected())
	}

	return nil
}
