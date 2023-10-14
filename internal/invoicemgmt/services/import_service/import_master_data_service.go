package services

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"

	"go.uber.org/zap"
)

type ImportMasterDataServiceRepositories struct {
	InvoiceScheduleRepo *repositories.InvoiceScheduleRepo
	PartnerBankRepo     *repositories.PartnerBankRepo
	OrganizationRepo    *repositories.OrganizationRepo
}

type ImportMasterDataService struct {
	logger              zap.SugaredLogger
	DB                  database.Ext
	InvoiceScheduleRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceSchedule) error
		RetrieveInvoiceScheduleByID(ctx context.Context, db database.QueryExecer, invoiceScheduleID string) (*entities.InvoiceSchedule, error)
		CancelScheduleIfExists(ctx context.Context, db database.QueryExecer, invoiceScheduleDate time.Time) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.InvoiceSchedule) error
	}
	PartnerBankRepo interface {
		RetrievePartnerBankByID(ctx context.Context, db database.QueryExecer, partnerBankID string) (*entities.PartnerBank, error)
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.PartnerBank) error
	}
	OrganizationRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, organizationID string) (*entities.Organization, error)
	}
	UnleashClient unleashclient.ClientInstance
	Env           string
}

func NewImportMasterDataService(
	logger zap.SugaredLogger,
	db database.Ext,
	serviceRepo *ImportMasterDataServiceRepositories,
	unleashClient unleashclient.ClientInstance,
	env string,
) *ImportMasterDataService {
	return &ImportMasterDataService{
		logger:              logger,
		DB:                  db,
		InvoiceScheduleRepo: serviceRepo.InvoiceScheduleRepo,
		PartnerBankRepo:     serviceRepo.PartnerBankRepo,
		OrganizationRepo:    serviceRepo.OrganizationRepo,
		UnleashClient:       unleashClient,
		Env:                 env,
	}
}
