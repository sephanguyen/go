package invoicesvc

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type OrderService interface {
	UpdateBillItemStatus(ctx context.Context, in *payment_pb.UpdateBillItemStatusRequest, opts ...grpc.CallOption) (*payment_pb.UpdateBillItemStatusResponse, error)
}

type InvoiceModifierServiceRepositories struct {
	InvoiceRepo                       *repositories.InvoiceRepo
	ActionLogRepo                     *repositories.InvoiceActionLogRepo
	InvoiceBillItemRepo               *repositories.InvoiceBillItemRepo
	BillItemRepo                      *repositories.BillItemRepo
	PaymentRepo                       *repositories.PaymentRepo
	OrganizationRepo                  *repositories.OrganizationRepo
	InvoiceScheduleRepo               *repositories.InvoiceScheduleRepo
	InvoiceScheduleHistoryRepo        *repositories.InvoiceScheduleHistoryRepo
	InvoiceScheduleStudentRepo        *repositories.InvoiceScheduleStudentRepo
	BulkPaymentRequestRepo            *repositories.BulkPaymentRequestRepo
	BulkPaymentRequestFileRepo        *repositories.BulkPaymentRequestFileRepo
	BulkPaymentRequestFilePaymentRepo *repositories.BulkPaymentRequestFilePaymentRepo
	PartnerConvenienceStoreRepo       *repositories.PartnerConvenienceStoreRepo
	PartnerBankRepo                   *repositories.PartnerBankRepo
	BulkPaymentValidationsRepo        *repositories.BulkPaymentValidationsRepo
	BulkPaymentValidationsDetailRepo  *repositories.BulkPaymentValidationsDetailRepo
	StudentPaymentDetailRepo          *repositories.StudentPaymentDetailRepo
	BankBranchRepo                    *repositories.BankBranchRepo
	NewCustomerCodeHistoryRepo        *repositories.NewCustomerCodeHistoryRepo
	OrderRepo                         *repositories.OrderRepo
	StudentRepo                       *repositories.StudentRepo
	PrefectureRepo                    *repositories.PrefectureRepo
	InvoiceAdjustmentRepo             *repositories.InvoiceAdjustmentRepo
	BulkPaymentRepo                   *repositories.BulkPaymentRepo
	BankAccountRepo                   *repositories.BankAccountRepo
	UserBasicInfoRepo                 *repositories.UserBasicInfoRepo
}

type InvoiceModifierService struct {
	logger               zap.SugaredLogger
	DB                   database.Ext
	OrderService         OrderService
	InternalOrderService OrderService
	InvoiceRepo          interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Invoice) (pgtype.Text, error)
		Update(ctx context.Context, db database.QueryExecer, e *entities.Invoice) error
		RetrieveInvoiceByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Invoice, error)
		RetrieveRecordsByStudentID(ctx context.Context, db database.QueryExecer, studentID string, limit, offset pgtype.Int8) ([]*entities.Invoice, error)
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.Invoice, fieldsToUpdate []string) error
		UpdateIsExportedByPaymentRequestFileID(ctx context.Context, db database.QueryExecer, filePaymentID string, isExported bool) error
		UpdateIsExportedByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string, isExported bool) error
		InsertInvoiceIDsTempTable(ctx context.Context, db database.QueryExecer, invoiceIDs []string) error
		FindInvoicesFromInvoiceIDTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.Invoice, error)
		UpdateStatusFromInvoiceIDTempTable(ctx context.Context, db database.QueryExecer, status string) error
		UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, invoices []*entities.Invoice, fields []string) error
		RetrieveInvoiceData(ctx context.Context, db database.QueryExecer, limit, offset pgtype.Int8, sqlFilter string) ([]*entities.InvoicePaymentMap, error)
		RetrieveInvoiceStatusCount(ctx context.Context, db database.QueryExecer, sqlFilter string) (map[string]int32, error)
	}
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
		CountOtherPaymentsByBulkPaymentIDNotInStatus(ctx context.Context, db database.QueryExecer, bulkPaymentID, paymentID, paymentStatus string) (int, error)
		FindByPaymentIDs(ctx context.Context, db database.QueryExecer, paymentIds []string) ([]*entities.Payment, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, payments []*entities.Payment) error
		GetLatestPaymentSequenceNumber(ctx context.Context, db database.QueryExecer) (int32, error)
		PaymentSeqNumberLockAdvisory(ctx context.Context, db database.QueryExecer) (bool, error)
		PaymentSeqNumberUnLockAdvisory(ctx context.Context, db database.QueryExecer) error
		UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, payments []*entities.Payment, fields []string) error
		FindPaymentInvoiceUserFromTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.PaymentInvoiceUserMap, error)
		InsertPaymentNumbersTempTable(ctx context.Context, db database.QueryExecer, paymentSeqNumbers []int) error
	}
	InvoiceActionLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceActionLog) error
		CreateMultiple(ctx context.Context, db database.QueryExecer, actionLogs []*entities.InvoiceActionLog) error
	}
	BillItemRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, billItemID int32) (*entities.BillItem, error)
		FindByStatuses(ctx context.Context, db database.QueryExecer, billItemStatuses []string) ([]*entities.BillItem, error)
		FindByOrderID(ctx context.Context, db database.QueryExecer, orderID string) ([]*entities.BillItem, error)
		FindInvoiceBillItemMapByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string) ([]*entities.InvoiceBillItemMap, error)
	}
	InvoiceBillItemRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceBillItem) error
		FindAllByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.InvoiceBillItems, error)
	}
	OrganizationRepo interface {
		GetOrganizations(ctx context.Context, db database.QueryExecer) ([]*entities.Organization, error)
		FindByID(ctx context.Context, db database.QueryExecer, organizationID string) (*entities.Organization, error)
	}
	InvoiceScheduleRepo interface {
		GetByStatusAndInvoiceDate(ctx context.Context, db database.QueryExecer, status string, invoiceDate time.Time) (*entities.InvoiceSchedule, error)
		GetByStatusAndScheduledDate(ctx context.Context, db database.QueryExecer, status string, scheduledDate time.Time) (*entities.InvoiceSchedule, error)
		Update(ctx context.Context, db database.QueryExecer, e *entities.InvoiceSchedule) error
		GetCurrentEarliestInvoiceSchedule(ctx context.Context, db database.QueryExecer, status string) (*entities.InvoiceSchedule, error)
	}
	InvoiceScheduleHistoryRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceScheduleHistory) (string, error)
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.InvoiceScheduleHistory, fields []string) error
	}
	InvoiceScheduleStudentRepo interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, e []*entities.InvoiceScheduleStudent) error
	}
	BulkPaymentRequestRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequest) (string, error)
		Update(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequest) error
		FindByPaymentRequestID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentRequest, error)
	}
	BulkPaymentRequestFileRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFile) (string, error)
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFile, fieldsToUpdate []string) error
		FindByPaymentFileID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentRequestFile, error)
	}
	BulkPaymentRequestFilePaymentRepo interface {
		FindByPaymentID(ctx context.Context, db database.QueryExecer, paymentID string) (*entities.BulkPaymentRequestFilePayment, error)
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFilePayment) (string, error)
		FindPaymentInvoiceByRequestFileID(ctx context.Context, db database.QueryExecer, id string) ([]*entities.FilePaymentInvoiceMap, error)
	}
	PartnerConvenienceStoreRepo interface {
		FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerConvenienceStore, error)
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
	StudentPaymentDetailRepo interface {
		FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentPaymentDetail, error)
		FindStudentBillingByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBillingDetailsMap, error)
		FindStudentBankDetailsByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBankDetailsMap, error)
		FindFromInvoiceIDTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.StudentPaymentDetail, error)
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
	OrderRepo interface {
		FindByOrderID(ctx context.Context, db database.QueryExecer, orderID string) (*entities.Order, error)
	}
	StudentRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.Student, error)
	}
	PrefectureRepo interface {
		FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.Prefecture, error)
	}
	InvoiceAdjustmentRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, invoiceAdjustments []*entities.InvoiceAdjustment) error
		SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
		FindByID(ctx context.Context, db database.QueryExecer, id string) (*entities.InvoiceAdjustment, error)
		FindByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string) ([]*entities.InvoiceAdjustment, error)
	}
	BulkPaymentRepo interface {
		UpdateBulkPaymentStatusByIDs(ctx context.Context, db database.QueryExecer, status string, bulkPaymentIDs []string) error
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPayment) error
	}
	BankAccountRepo interface {
		FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BankAccount, error)
	}
	FileStorage           filestorage.FileStorage
	UnleashClient         unleashclient.ClientInstance
	TempFileCreator       utils.ITempFileCreator
	Env                   string
	SequenceNumberService seqnumberservice.ISequenceNumberService
	UserBasicInfoRepo     interface {
		FindByID(ctx context.Context, db database.QueryExecer, userID string) (*entities.UserBasicInfo, error)
	}
}

func NewInvoiceModifierService(
	logger zap.SugaredLogger,
	db database.Ext,
	internalOrderServiceClient OrderService,
	fileStorage filestorage.FileStorage,
	serviceRepo *InvoiceModifierServiceRepositories,
	unleashClient unleashclient.ClientInstance,
	env string,
	tempFileCreator utils.ITempFileCreator,
) *InvoiceModifierService {
	return &InvoiceModifierService{
		logger:                            logger,
		DB:                                db,
		InternalOrderService:              internalOrderServiceClient,
		FileStorage:                       fileStorage,
		InvoiceRepo:                       serviceRepo.InvoiceRepo,
		InvoiceActionLogRepo:              serviceRepo.ActionLogRepo,
		InvoiceBillItemRepo:               serviceRepo.InvoiceBillItemRepo,
		BillItemRepo:                      serviceRepo.BillItemRepo,
		PaymentRepo:                       serviceRepo.PaymentRepo,
		OrganizationRepo:                  serviceRepo.OrganizationRepo,
		InvoiceScheduleRepo:               serviceRepo.InvoiceScheduleRepo,
		InvoiceScheduleHistoryRepo:        serviceRepo.InvoiceScheduleHistoryRepo,
		InvoiceScheduleStudentRepo:        serviceRepo.InvoiceScheduleStudentRepo,
		BulkPaymentRequestRepo:            serviceRepo.BulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:        serviceRepo.BulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: serviceRepo.BulkPaymentRequestFilePaymentRepo,
		PartnerConvenienceStoreRepo:       serviceRepo.PartnerConvenienceStoreRepo,
		PartnerBankRepo:                   serviceRepo.PartnerBankRepo,
		BulkPaymentValidationsRepo:        serviceRepo.BulkPaymentValidationsRepo,
		BulkPaymentValidationsDetailRepo:  serviceRepo.BulkPaymentValidationsDetailRepo,
		StudentPaymentDetailRepo:          serviceRepo.StudentPaymentDetailRepo,
		BankBranchRepo:                    serviceRepo.BankBranchRepo,
		NewCustomerCodeHistoryRepo:        serviceRepo.NewCustomerCodeHistoryRepo,
		OrderRepo:                         serviceRepo.OrderRepo,
		StudentRepo:                       serviceRepo.StudentRepo,
		PrefectureRepo:                    serviceRepo.PrefectureRepo,
		InvoiceAdjustmentRepo:             serviceRepo.InvoiceAdjustmentRepo,
		BulkPaymentRepo:                   serviceRepo.BulkPaymentRepo,
		BankAccountRepo:                   serviceRepo.BankAccountRepo,
		UserBasicInfoRepo:                 serviceRepo.UserBasicInfoRepo,
		UnleashClient:                     unleashClient,
		Env:                               env,
		TempFileCreator:                   tempFileCreator,
		SequenceNumberService: &seqnumberservice.SequenceNumberService{
			PaymentRepo: serviceRepo.PaymentRepo,
			Logger:      logger,
		},
	}
}
