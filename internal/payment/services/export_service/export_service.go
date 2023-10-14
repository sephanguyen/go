package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	export_entities "github.com/manabie-com/backend/internal/payment/export_entities"
	bill_item_service "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	fee_service "github.com/manabie-com/backend/internal/payment/services/domain_service/fee"
	material_service "github.com/manabie-com/backend/internal/payment/services/domain_service/material"
	package_service "github.com/manabie-com/backend/internal/payment/services/domain_service/package"
)

type IBillItemServiceForExportService interface {
	GetExportStudentBilling(
		ctx context.Context, db database.QueryExecer, locationIDs []string) (
		exportData []*export_entities.StudentBillingExport, err error)
}

type IMaterialServiceForExportService interface {
	GetAllMaterialsForExport(ctx context.Context, db database.QueryExecer) (materials []export_entities.ProductMaterialExport, err error)
}

type IFeeServiceForExportService interface {
	GetAllFeesForExport(ctx context.Context, db database.QueryExecer) (fees []export_entities.ProductFeeExport, err error)
}

type IPackageServiceForExportService interface {
	GetAllPackagesForExport(ctx context.Context, db database.Ext) (packages []*export_entities.ProductPackageExport, err error)
}
type ExportService struct {
	DB              database.Ext
	BillItemService IBillItemServiceForExportService
	MaterialService IMaterialServiceForExportService
	FeeService      IFeeServiceForExportService
	PackageService  IPackageServiceForExportService
}

func NewExportService(db database.Ext) (exportService *ExportService) {
	return &ExportService{
		DB:              db,
		BillItemService: bill_item_service.NewBillItemService(),
		MaterialService: material_service.NewMaterialService(),
		FeeService:      fee_service.NewFeeService(),
		PackageService:  package_service.NewPackageService(),
	}
}
