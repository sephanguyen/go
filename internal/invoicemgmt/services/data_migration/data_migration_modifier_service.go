package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

type DataMigrationModifierServiceRepositories struct {
	InvoiceRepo         *repositories.InvoiceRepo
	PaymentRepo         *repositories.PaymentRepo
	BillItemRepo        *repositories.BillItemRepo
	InvoiceBillItemRepo *repositories.InvoiceBillItemRepo
	StudentRepo         *repositories.StudentRepo
}

type DataMigrationModifierService struct {
	logger      zap.SugaredLogger
	DB          database.Ext
	InvoiceRepo interface {
		RetrieveInvoiceByInvoiceReferenceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Invoice, error)
		Create(ctx context.Context, db database.QueryExecer, e *entities.Invoice) (pgtype.Text, error)
		RetrievedMigratedInvoices(ctx context.Context, db database.QueryExecer) ([]*entities.Invoice, error)
	}
	PaymentRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Payment) error
	}
	BillItemRepo interface {
		RetrieveBillItemsByInvoiceReferenceNum(ctx context.Context, db database.QueryExecer, referenceID string) ([]*entities.BillItem, error)
		GetBillItemTotalByStudentAndReference(ctx context.Context, db database.QueryExecer, studentID, invoiceReferenceID string) (pgtype.Numeric, error)
	}
	InvoiceBillItemRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceBillItem) error
		FindAllByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.InvoiceBillItems, error)
	}
	StudentRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.Student, error)
	}
}

func NewDataMigrationModifierService(
	logger zap.SugaredLogger,
	db database.Ext,
	serviceRepo *DataMigrationModifierServiceRepositories,
) *DataMigrationModifierService {
	return &DataMigrationModifierService{
		logger:              logger,
		DB:                  db,
		InvoiceRepo:         serviceRepo.InvoiceRepo,
		PaymentRepo:         serviceRepo.PaymentRepo,
		BillItemRepo:        serviceRepo.BillItemRepo,
		InvoiceBillItemRepo: serviceRepo.InvoiceBillItemRepo,
		StudentRepo:         serviceRepo.StudentRepo,
	}
}
