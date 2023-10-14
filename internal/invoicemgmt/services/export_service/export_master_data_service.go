package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	export_entities "github.com/manabie-com/backend/internal/invoicemgmt/export_entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"

	"go.uber.org/zap"
)

type ExportMasterDataServiceRepositories struct {
	InvoiceScheduleRepo *repositories.InvoiceScheduleRepo
	BankBranchRepo      *repositories.BankBranchRepo
	BankRepo            *repositories.BankRepo
	BankMappingRepo     *repositories.BankMappingRepo
}

type ExportMasterDataService struct {
	logger         zap.SugaredLogger
	DB             database.Ext
	BankBranchRepo interface {
		FindExportableBankBranches(ctx context.Context, db database.QueryExecer) ([]*export_entities.BankBranchExport, error)
	}
	InvoiceScheduleRepo interface {
		FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.InvoiceSchedule, error)
	}
	BankRepo interface {
		FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.Bank, error)
	}
	BankMappingRepo interface {
		FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.BankMapping, error)
	}
}

func NewExportMasterDataService(
	logger zap.SugaredLogger,
	db database.Ext,
	serviceRepositories *ExportMasterDataServiceRepositories,
) *ExportMasterDataService {
	return &ExportMasterDataService{
		logger:              logger,
		DB:                  db,
		BankBranchRepo:      serviceRepositories.BankBranchRepo,
		InvoiceScheduleRepo: serviceRepositories.InvoiceScheduleRepo,
		BankRepo:            serviceRepositories.BankRepo,
		BankMappingRepo:     serviceRepositories.BankMappingRepo,
	}
}
