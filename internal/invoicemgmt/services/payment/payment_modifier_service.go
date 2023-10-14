package paymentsvc

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

type PaymentModifierServiceRepositories struct {
	PaymentRepo                       *repositories.PaymentRepo
	BulkPaymentRequestRepo            *repositories.BulkPaymentRequestRepo
	InvoiceRepo                       *repositories.InvoiceRepo
	BulkPaymentRequestFileRepo        *repositories.BulkPaymentRequestFileRepo
	BulkPaymentRequestFilePaymentRepo *repositories.BulkPaymentRequestFilePaymentRepo
	PartnerConvenienceStoreRepo       *repositories.PartnerConvenienceStoreRepo
	StudentPaymentDetailRepo          *repositories.StudentPaymentDetailRepo
	BankBranchRepo                    *repositories.BankBranchRepo
	NewCustomerCodeHistoryRepo        *repositories.NewCustomerCodeHistoryRepo
	PrefectureRepo                    *repositories.PrefectureRepo
	PartnerBankRepo                   *repositories.PartnerBankRepo
	BulkPaymentValidationsRepo        *repositories.BulkPaymentValidationsRepo
	BulkPaymentValidationsDetailRepo  *repositories.BulkPaymentValidationsDetailRepo
	InvoiceActionLogRepo              *repositories.InvoiceActionLogRepo
	BankAccountRepo                   *repositories.BankAccountRepo
	StudentRepo                       *repositories.StudentRepo
	BulkPaymentRepo                   *repositories.BulkPaymentRepo
	UserBasicInfoRepo                 *repositories.UserBasicInfoRepo
}

type PaymentModifierService struct {
	logger      zap.SugaredLogger
	DB          database.Ext
	PaymentRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Payment) error
		GetLatestPaymentDueDateByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Payment, error)
		Update(ctx context.Context, db database.QueryExecer, e *entities.Payment) error
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.Payment, fieldsToUpdate []string) error
		FindByPaymentID(ctx context.Context, db database.QueryExecer, paymentID string) (*entities.Payment, error)
		FindByPaymentSequenceNumber(ctx context.Context, db database.QueryExecer, paymentSequenceNumber int) (*entities.Payment, error)
		FindPaymentInvoiceByIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string) ([]*entities.PaymentInvoiceMap, error)
		UpdateIsExportedByPaymentRequestFileID(ctx context.Context, db database.QueryExecer, filePaymentID string, isExported bool) error
		UpdateIsExportedByPaymentIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string, isExported bool) error
		FindAllByBulkPaymentID(ctx context.Context, db database.QueryExecer, bulkPaymentID string) ([]*entities.Payment, error)
		UpdateStatusAndAmountByPaymentIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string, status string, amount float64) error
		CountOtherPaymentsByBulkPaymentIDNotInStatus(ctx context.Context, db database.QueryExecer, bulkPaymentID, paymentID, paymentStatus string) (int, error)
		UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, payments []*entities.Payment, fields []string) error
		FindPaymentInvoiceUserFromTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.PaymentInvoiceUserMap, error)
		InsertPaymentNumbersTempTable(ctx context.Context, db database.QueryExecer, paymentSeqNumbers []int) error
	}
	BulkPaymentRequestRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequest) (string, error)
		Update(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequest) error
		FindByPaymentRequestID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentRequest, error)
	}
	InvoiceRepo interface {
		UpdateIsExportedByPaymentRequestFileID(ctx context.Context, db database.QueryExecer, filePaymentID string, isExported bool) error
		UpdateIsExportedByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string, isExported bool) error
		RetrieveInvoiceByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Invoice, error)
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.Invoice, fieldsToUpdate []string) error
		UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, invoices []*entities.Invoice, fields []string) error
	}
	BulkPaymentRequestFileRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFile) (string, error)
		FindByPaymentFileID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentRequestFile, error)
	}
	BulkPaymentRequestFilePaymentRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFilePayment) (string, error)
		FindByPaymentID(ctx context.Context, db database.QueryExecer, paymentID string) (*entities.BulkPaymentRequestFilePayment, error)
		FindPaymentInvoiceByRequestFileID(ctx context.Context, db database.QueryExecer, id string) ([]*entities.FilePaymentInvoiceMap, error)
	}
	PartnerConvenienceStoreRepo interface {
		FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerConvenienceStore, error)
	}
	StudentPaymentDetailRepo interface {
		FindStudentBillingByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBillingDetailsMap, error)
		FindStudentBankDetailsByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBankDetailsMap, error)
		FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentPaymentDetail, error)
	}
	BankBranchRepo interface {
		FindRelatedBankOfBankBranches(ctx context.Context, db database.QueryExecer, branchIDs []string) ([]*entities.BankRelationMap, error)
	}
	NewCustomerCodeHistoryRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.NewCustomerCodeHistory) error
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.NewCustomerCodeHistory, error)
		FindByAccountNumbers(ctx context.Context, db database.QueryExecer, bankAccountNumbers []string) ([]*entities.NewCustomerCodeHistory, error)
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.NewCustomerCodeHistory, fieldsToUpdate []string) error
	}
	PrefectureRepo interface {
		FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.Prefecture, error)
	}
	PartnerBankRepo interface {
		FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerBank, error)
	}
	BulkPaymentValidationsRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, bulkPaymentValidationsID string) (*entities.BulkPaymentValidations, error)
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidations, fieldsToUpdate []string) error
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidations) (string, error)
	}
	BulkPaymentValidationsDetailRepo interface {
		RetrieveRecordsByBulkPaymentValidationsID(ctx context.Context, db database.QueryExecer, bulkPaymentValidationsID pgtype.Text) ([]*entities.BulkPaymentValidationsDetail, error)
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidationsDetail) (string, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, validationDetails []*entities.BulkPaymentValidationsDetail) error
	}
	InvoiceActionLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceActionLog) error
		CreateMultiple(ctx context.Context, db database.QueryExecer, actionLogs []*entities.InvoiceActionLog) error
	}
	BankAccountRepo interface {
		FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BankAccount, error)
	}
	StudentRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.Student, error)
	}
	BulkPaymentRepo interface {
		UpdateBulkPaymentStatusByIDs(ctx context.Context, db database.QueryExecer, status string, bulkPaymentIDs []string) error
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPayment) error
		FindByBulkPaymentID(ctx context.Context, db database.QueryExecer, bulkPaymentID string) (*entities.BulkPayment, error)
	}
	UserBasicInfoRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, userID string) (*entities.UserBasicInfo, error)
	}

	FileStorage           filestorage.FileStorage
	UnleashClient         unleashclient.ClientInstance
	TempFileCreator       utils.ITempFileCreator
	Env                   string
	SequenceNumberService seqnumberservice.ISequenceNumberService
}

func NewPaymentModifierService(
	logger zap.SugaredLogger,
	db database.Ext,
	serviceRepo *PaymentModifierServiceRepositories,
	fileStorage filestorage.FileStorage,
	unleashClient unleashclient.ClientInstance,
	tempFileCreator utils.ITempFileCreator,
	env string,
) *PaymentModifierService {
	return &PaymentModifierService{
		logger:                            logger,
		DB:                                db,
		PaymentRepo:                       serviceRepo.PaymentRepo,
		BulkPaymentRequestRepo:            serviceRepo.BulkPaymentRequestRepo,
		InvoiceRepo:                       serviceRepo.InvoiceRepo,
		BulkPaymentRequestFileRepo:        serviceRepo.BulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: serviceRepo.BulkPaymentRequestFilePaymentRepo,
		PartnerConvenienceStoreRepo:       serviceRepo.PartnerConvenienceStoreRepo,
		StudentPaymentDetailRepo:          serviceRepo.StudentPaymentDetailRepo,
		BankBranchRepo:                    serviceRepo.BankBranchRepo,
		NewCustomerCodeHistoryRepo:        serviceRepo.NewCustomerCodeHistoryRepo,
		PrefectureRepo:                    serviceRepo.PrefectureRepo,
		PartnerBankRepo:                   serviceRepo.PartnerBankRepo,
		BulkPaymentValidationsRepo:        serviceRepo.BulkPaymentValidationsRepo,
		BulkPaymentValidationsDetailRepo:  serviceRepo.BulkPaymentValidationsDetailRepo,
		InvoiceActionLogRepo:              serviceRepo.InvoiceActionLogRepo,
		BankAccountRepo:                   serviceRepo.BankAccountRepo,
		StudentRepo:                       serviceRepo.StudentRepo,
		BulkPaymentRepo:                   serviceRepo.BulkPaymentRepo,
		UserBasicInfoRepo:                 serviceRepo.UserBasicInfoRepo,
		FileStorage:                       fileStorage,
		UnleashClient:                     unleashClient,
		TempFileCreator:                   tempFileCreator,
		Env:                               env,
		SequenceNumberService: &seqnumberservice.SequenceNumberService{
			PaymentRepo: serviceRepo.PaymentRepo,
			Logger:      logger,
		},
	}
}
