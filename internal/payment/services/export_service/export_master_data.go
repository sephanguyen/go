package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	dbEntities "github.com/manabie-com/backend/internal/payment/entities"
	export_entities "github.com/manabie-com/backend/internal/payment/export_entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ExportService) ExportMasterData(ctx context.Context, req *pb.ExportMasterDataRequest) (exportDataResp *pb.ExportMasterDataResponse, err error) {
	colMap, entityType := GetExportColMapAndEntityType(req.ExportDataType)
	var entities []database.Entity
	switch req.ExportDataType {
	case pb.ExportMasterDataType_EXPORT_MATERIAL:
		materials, err := s.MaterialService.GetAllMaterialsForExport(ctx, s.DB)
		if err != nil {
			return nil, err
		}
		entities = make([]database.Entity, 0, len(materials))
		for _, e := range materials {
			productMaterialExport := &export_entities.ProductMaterialExport{
				MaterialID:           e.MaterialID,
				Name:                 e.Name,
				MaterialType:         e.MaterialType,
				TaxID:                e.TaxID,
				ProductTag:           e.ProductTag,
				ProductPartnerID:     e.ProductPartnerID,
				AvailableFrom:        e.AvailableFrom,
				AvailableUntil:       e.AvailableUntil,
				CustomBillingPeriod:  e.CustomBillingPeriod,
				CustomBillingDate:    e.CustomBillingDate,
				DisableProRatingFlag: e.DisableProRatingFlag,
				BillingScheduleID:    e.BillingScheduleID,
				Remarks:              e.Remarks,
				IsUnique:             e.IsUnique,
				IsArchived:           e.IsArchived,
			}
			entities = append(entities, productMaterialExport)
		}
	case pb.ExportMasterDataType_EXPORT_FEE:
		fees, err := s.FeeService.GetAllFeesForExport(ctx, s.DB)
		if err != nil {
			return nil, err
		}
		entities = make([]database.Entity, 0, len(fees))
		for _, e := range fees {
			productFeeExport := &export_entities.ProductFeeExport{
				FeeID:                e.FeeID,
				Name:                 e.Name,
				FeeType:              e.FeeType,
				TaxID:                e.TaxID,
				ProductTag:           e.ProductTag,
				ProductPartnerID:     e.ProductPartnerID,
				AvailableFrom:        e.AvailableFrom,
				AvailableUntil:       e.AvailableUntil,
				CustomBillingPeriod:  e.CustomBillingPeriod,
				BillingScheduleID:    e.BillingScheduleID,
				DisableProRatingFlag: e.DisableProRatingFlag,
				Remarks:              e.Remarks,
				IsUnique:             e.IsUnique,
				IsArchived:           e.IsArchived,
			}
			entities = append(entities, productFeeExport)
		}
	case pb.ExportMasterDataType_EXPORT_PACKAGE:
		packages, err := s.PackageService.GetAllPackagesForExport(ctx, s.DB)
		if err != nil {
			return nil, err
		}
		entities = make([]database.Entity, 0, len(packages))
		for _, e := range packages {
			productPackageExport := &export_entities.ProductPackageExport{
				PackageID:            e.PackageID,
				Name:                 e.Name,
				PackageType:          e.PackageType,
				TaxID:                e.TaxID,
				ProductTag:           e.ProductTag,
				ProductPartnerID:     e.ProductPartnerID,
				AvailableFrom:        e.AvailableFrom,
				AvailableUntil:       e.AvailableUntil,
				MaxSlot:              e.MaxSlot,
				CustomBillingPeriod:  e.CustomBillingPeriod,
				BillingScheduleID:    e.BillingScheduleID,
				DisableProRatingFlag: e.DisableProRatingFlag,
				PackageStartDate:     e.PackageStartDate,
				PackageEndDate:       e.PackageEndDate,
				Remarks:              e.Remarks,
				IsArchived:           e.IsArchived,
				IsUnique:             e.IsUnique,
			}
			entities = append(entities, productPackageExport)
		}
	default:
		entities, err = exporter.RetrieveAllData(ctx, s.DB, entityType)
		if err != nil {
			return nil, err
		}
	}

	res, err := exporter.ExportBatch(entities, colMap)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("exporter.ExportBatch err: %v", err))
	}
	csvBytes := exporter.ToCSV(res)

	return &pb.ExportMasterDataResponse{
		Data: csvBytes,
	}, nil
}

func GetExportColMapAndEntityType(exportDataType pb.ExportMasterDataType) (colMap []exporter.ExportColumnMap, entityType database.Entity) {
	switch exportDataType {
	case pb.ExportMasterDataType_EXPORT_ACCOUNTING_CATEGORY:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "accounting_category_id",
				DBColumn:  "accounting_category_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "remarks",
				DBColumn:  "remarks",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
			{
				CSVColumn: "updated_at",
				DBColumn:  "updated_at",
			},
		}
		entityType = &dbEntities.AccountingCategory{}
	case pb.ExportMasterDataType_EXPORT_ASSOCIATED_PRODUCTS_MATERIAL:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "package_id",
				DBColumn:  "package_id",
			},
			{
				CSVColumn: "course_id",
				DBColumn:  "course_id",
			},
			{
				CSVColumn: "material_id",
				DBColumn:  "material_id",
			},
			{
				CSVColumn: "available_from",
				DBColumn:  "available_from",
			},
			{
				CSVColumn: "available_until",
				DBColumn:  "available_until",
			},
		}
		entityType = &dbEntities.PackageCourseMaterial{}
	case pb.ExportMasterDataType_EXPORT_ASSOCIATED_PRODUCTS_FEE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "package_id",
				DBColumn:  "package_id",
			},
			{
				CSVColumn: "course_id",
				DBColumn:  "course_id",
			},
			{
				CSVColumn: "fee_id",
				DBColumn:  "fee_id",
			},
			{
				CSVColumn: "available_from",
				DBColumn:  "available_from",
			},
			{
				CSVColumn: "available_until",
				DBColumn:  "available_until",
			},
		}
		entityType = &dbEntities.PackageCourseFee{}
	case pb.ExportMasterDataType_EXPORT_BILLING_RATIO:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "billing_ratio_id",
				DBColumn:  "billing_ratio_id",
			},
			{
				CSVColumn: "start_date",
				DBColumn:  "start_date",
			},
			{
				CSVColumn: "end_date",
				DBColumn:  "end_date",
			},
			{
				CSVColumn: "billing_schedule_period_id",
				DBColumn:  "billing_schedule_period_id",
			},
			{
				CSVColumn: "billing_ratio_numerator",
				DBColumn:  "billing_ratio_numerator",
			},
			{
				CSVColumn: "billing_ratio_denominator",
				DBColumn:  "billing_ratio_denominator",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
		}
		entityType = &dbEntities.BillingRatio{}
	case pb.ExportMasterDataType_EXPORT_BILLING_SCHEDULE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "billing_schedule_id",
				DBColumn:  "billing_schedule_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "remarks",
				DBColumn:  "remarks",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
		}
		entityType = &dbEntities.BillingSchedule{}
	case pb.ExportMasterDataType_EXPORT_BILLING_SCHEDULE_PERIOD:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "billing_schedule_period_id",
				DBColumn:  "billing_schedule_period_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "billing_schedule_id",
				DBColumn:  "billing_schedule_id",
			},
			{
				CSVColumn: "start_date",
				DBColumn:  "start_date",
			},
			{
				CSVColumn: "end_date",
				DBColumn:  "end_date",
			},
			{
				CSVColumn: "billing_date",
				DBColumn:  "billing_date",
			},
			{
				CSVColumn: "remarks",
				DBColumn:  "remarks",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
		}
		entityType = &dbEntities.BillingSchedulePeriod{}
	case pb.ExportMasterDataType_EXPORT_DISCOUNT:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "discount_id",
				DBColumn:  "discount_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "discount_type",
				DBColumn:  "discount_type",
			},
			{
				CSVColumn: "discount_amount_type",
				DBColumn:  "discount_amount_type",
			},
			{
				CSVColumn: "discount_amount_value",
				DBColumn:  "discount_amount_value",
			},
			{
				CSVColumn: "recurring_valid_duration",
				DBColumn:  "recurring_valid_duration",
			},
			{
				CSVColumn: "available_from",
				DBColumn:  "available_from",
			},
			{
				CSVColumn: "available_until",
				DBColumn:  "available_until",
			},
			{
				CSVColumn: "remarks",
				DBColumn:  "remarks",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
			{
				CSVColumn: "student_tag_id_validation",
				DBColumn:  "student_tag_id_validation",
			},
			{
				CSVColumn: "parent_tag_id_validation",
				DBColumn:  "parent_tag_id_validation",
			},
			{
				CSVColumn: "discount_tag_id",
				DBColumn:  "discount_tag_id",
			},
		}
		entityType = &dbEntities.Discount{}
	case pb.ExportMasterDataType_EXPORT_FEE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "fee_id",
				DBColumn:  "fee_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "fee_type",
				DBColumn:  "fee_type",
			},
			{
				CSVColumn: "tax_id",
				DBColumn:  "tax_id",
			},
			{
				CSVColumn: "product_tag",
				DBColumn:  "product_tag",
			},
			{
				CSVColumn: "product_partner_id",
				DBColumn:  "product_partner_id",
			},
			{
				CSVColumn: "available_from",
				DBColumn:  "available_from",
			},
			{
				CSVColumn: "available_until",
				DBColumn:  "available_until",
			},
			{
				CSVColumn: "custom_billing_period",
				DBColumn:  "custom_billing_period",
			},
			{
				CSVColumn: "billing_schedule_id",
				DBColumn:  "billing_schedule_id",
			},
			{
				CSVColumn: "disable_pro_rating_flag",
				DBColumn:  "disable_pro_rating_flag",
			},
			{
				CSVColumn: "remarks",
				DBColumn:  "remarks",
			},
			{
				CSVColumn: "is_unique",
				DBColumn:  "is_unique",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
		}
		entityType = &dbEntities.Discount{}
	case pb.ExportMasterDataType_EXPORT_LEAVING_REASON:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "leaving_reason_id",
				DBColumn:  "leaving_reason_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "leaving_reason_type",
				DBColumn:  "leaving_reason_type",
			},
			{
				CSVColumn: "remark",
				DBColumn:  "remark",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
			{
				CSVColumn: "updated_at",
				DBColumn:  "updated_at",
			},
		}
		entityType = &dbEntities.LeavingReason{}
	case pb.ExportMasterDataType_EXPORT_MATERIAL:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "material_id",
				DBColumn:  "material_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "material_type",
				DBColumn:  "material_type",
			},
			{
				CSVColumn: "tax_id",
				DBColumn:  "tax_id",
			},
			{
				CSVColumn: "product_tag",
				DBColumn:  "product_tag",
			},
			{
				CSVColumn: "product_partner_id",
				DBColumn:  "product_partner_id",
			},
			{
				CSVColumn: "available_from",
				DBColumn:  "available_from",
			},
			{
				CSVColumn: "available_until",
				DBColumn:  "available_until",
			},
			{
				CSVColumn: "custom_billing_period",
				DBColumn:  "custom_billing_period",
			},
			{
				CSVColumn: "custom_billing_date",
				DBColumn:  "custom_billing_date",
			},
			{
				CSVColumn: "disable_pro_rating_flag",
				DBColumn:  "disable_pro_rating_flag",
			},
			{
				CSVColumn: "billing_schedule_id",
				DBColumn:  "billing_schedule_id",
			},
			{
				CSVColumn: "remarks",
				DBColumn:  "remarks",
			},
			{
				CSVColumn: "is_unique",
				DBColumn:  "is_unique",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
		}
		entityType = &dbEntities.Material{}
	case pb.ExportMasterDataType_EXPORT_PACKAGE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "package_id",
				DBColumn:  "package_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "package_type",
				DBColumn:  "package_type",
			},
			{
				CSVColumn: "tax_id",
				DBColumn:  "tax_id",
			},
			{
				CSVColumn: "product_tag",
				DBColumn:  "product_tag",
			},
			{
				CSVColumn: "product_partner_id",
				DBColumn:  "product_partner_id",
			},
			{
				CSVColumn: "available_from",
				DBColumn:  "available_from",
			},
			{
				CSVColumn: "available_until",
				DBColumn:  "available_until",
			},
			{
				CSVColumn: "max_slot",
				DBColumn:  "max_slot",
			},
			{
				CSVColumn: "custom_billing_period",
				DBColumn:  "custom_billing_period",
			},
			{
				CSVColumn: "billing_schedule_id",
				DBColumn:  "billing_schedule_id",
			},
			{
				CSVColumn: "disable_pro_rating_flag",
				DBColumn:  "disable_pro_rating_flag",
			},
			{
				CSVColumn: "package_start_date",
				DBColumn:  "package_start_date",
			},
			{
				CSVColumn: "package_end_date",
				DBColumn:  "package_end_date",
			},
			{
				CSVColumn: "remarks",
				DBColumn:  "remarks",
			},
			{
				CSVColumn: "is_unique",
				DBColumn:  "is_unique",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
		}
		entityType = &dbEntities.Package{}
	case pb.ExportMasterDataType_EXPORT_PACKAGE_ASSOCIATED_COURSE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "package_id",
				DBColumn:  "package_id",
			},
			{
				CSVColumn: "course_id",
				DBColumn:  "course_id",
			},
			{
				CSVColumn: "mandatory_flag",
				DBColumn:  "mandatory_flag",
			},
			{
				CSVColumn: "course_weight",
				DBColumn:  "course_weight",
			},
			{
				CSVColumn: "max_slots_per_course",
				DBColumn:  "max_slots_per_course",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
		}
		entityType = &dbEntities.PackageCourse{}
	case pb.ExportMasterDataType_EXPORT_PACKAGE_QUANTITY_TYPE_MAPPING:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "package_type",
				DBColumn:  "package_type",
			},
			{
				CSVColumn: "quantity_type",
				DBColumn:  "quantity_type",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
		}
		entityType = &dbEntities.PackageQuantityTypeMapping{}
	case pb.ExportMasterDataType_EXPORT_PRODUCT_ASSOCIATED_ACCOUNTING_CATEGORY:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "product_id",
				DBColumn:  "product_id",
			},
			{
				CSVColumn: "accounting_category_id",
				DBColumn:  "accounting_category_id",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
		}
		entityType = &dbEntities.ProductAccountingCategory{}
	case pb.ExportMasterDataType_EXPORT_PRODUCT_ASSOCIATED_GRADE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "product_id",
				DBColumn:  "product_id",
			},
			{
				CSVColumn: "grade_id",
				DBColumn:  "grade_id",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
		}
		entityType = &dbEntities.ProductGrade{}
	case pb.ExportMasterDataType_EXPORT_PRODUCT_ASSOCIATED_LOCATION:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "product_id",
				DBColumn:  "product_id",
			},
			{
				CSVColumn: "location_id",
				DBColumn:  "location_id",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
		}
		entityType = &dbEntities.ProductLocation{}
	case pb.ExportMasterDataType_EXPORT_PRODUCT_DISCOUNT:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "discount_id",
				DBColumn:  "discount_id",
			},
			{
				CSVColumn: "product_id",
				DBColumn:  "product_id",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
		}
		entityType = &dbEntities.ProductDiscount{}
	case pb.ExportMasterDataType_EXPORT_PRODUCT_PRICE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "product_price_id",
				DBColumn:  "product_price_id",
			},
			{
				CSVColumn: "product_id",
				DBColumn:  "product_id",
			},
			{
				CSVColumn: "billing_schedule_period_id",
				DBColumn:  "billing_schedule_period_id",
			},
			{
				CSVColumn: "quantity",
				DBColumn:  "quantity",
			},
			{
				CSVColumn: "price",
				DBColumn:  "price",
			},
			{
				CSVColumn: "price_type",
				DBColumn:  "price_type",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
		}
		entityType = &dbEntities.ProductPrice{}
	case pb.ExportMasterDataType_EXPORT_PRODUCT_SETTING:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "product_id",
				DBColumn:  "product_id",
			},
			{
				CSVColumn: "is_enrollment_required",
				DBColumn:  "is_enrollment_required",
			},
			{
				CSVColumn: "is_pausable",
				DBColumn:  "is_pausable",
			},
			{
				CSVColumn: "is_added_to_enrollment_by_default",
				DBColumn:  "is_added_to_enrollment_by_default",
			},
			{
				CSVColumn: "is_operation_fee",
				DBColumn:  "is_operation_fee",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
			{
				CSVColumn: "updated_at",
				DBColumn:  "updated_at",
			},
		}
		entityType = &dbEntities.ProductSetting{}
	case pb.ExportMasterDataType_EXPORT_TAX:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "tax_id",
				DBColumn:  "tax_id",
			},
			{
				CSVColumn: "name",
				DBColumn:  "name",
			},
			{
				CSVColumn: "tax_percentage",
				DBColumn:  "tax_percentage",
			},
			{
				CSVColumn: "tax_category",
				DBColumn:  "tax_category",
			},
			{
				CSVColumn: "default_flag",
				DBColumn:  "default_flag",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
			{
				CSVColumn: "updated_at",
				DBColumn:  "updated_at",
			},
		}
		entityType = &dbEntities.Tax{}
	case pb.ExportMasterDataType_EXPORT_NOTIFICATION_DATE:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "notification_date_id",
				DBColumn:  "notification_date_id",
			},
			{
				CSVColumn: "order_type",
				DBColumn:  "order_type",
			},
			{
				CSVColumn: "notification_date",
				DBColumn:  "notification_date",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
			{
				CSVColumn: "created_at",
				DBColumn:  "created_at",
			},
			{
				CSVColumn: "updated_at",
				DBColumn:  "updated_at",
			},
		}
		entityType = &dbEntities.NotificationDate{}
	default:
	}

	return
}
