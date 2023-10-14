package invoicemgmt

import (
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	dataMigrationService "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration"
	exportService "github.com/manabie-com/backend/internal/invoicemgmt/services/export_service"
	importService "github.com/manabie-com/backend/internal/invoicemgmt/services/import_service"
	invoiceService "github.com/manabie-com/backend/internal/invoicemgmt/services/invoice"
	openAPIService "github.com/manabie-com/backend/internal/invoicemgmt/services/open_api"
	paymentService "github.com/manabie-com/backend/internal/invoicemgmt/services/payment"
)

type Repositories struct {
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
	UserRepo                          *repositories.UserRepo
	StudentPaymentDetailRepo          *repositories.StudentPaymentDetailRepo
	BankBranchRepo                    *repositories.BankBranchRepo
	NewCustomerCodeHistoryRepo        *repositories.NewCustomerCodeHistoryRepo
	OrderRepo                         *repositories.OrderRepo
	StudentRepo                       *repositories.StudentRepo
	PrefectureRepo                    *repositories.PrefectureRepo
	InvoiceAdjustmentRepo             *repositories.InvoiceAdjustmentRepo
	BillingAddressRepo                *repositories.BillingAddressRepo
	BankAccountRepo                   *repositories.BankAccountRepo
	BankRepo                          *repositories.BankRepo
	BankMappingRepo                   *repositories.BankMappingRepo
	BulkPaymentRepo                   *repositories.BulkPaymentRepo
	StudentPaymentDetailActionLogRepo *repositories.StudentPaymentDetailActionLogRepo
	UserBasicInfoRepo                 *repositories.UserBasicInfoRepo
}

func initRepositories() *Repositories {
	return &Repositories{
		InvoiceRepo:                       &repositories.InvoiceRepo{},
		ActionLogRepo:                     &repositories.InvoiceActionLogRepo{},
		InvoiceBillItemRepo:               &repositories.InvoiceBillItemRepo{},
		BillItemRepo:                      &repositories.BillItemRepo{},
		PaymentRepo:                       &repositories.PaymentRepo{},
		OrganizationRepo:                  &repositories.OrganizationRepo{},
		InvoiceScheduleRepo:               &repositories.InvoiceScheduleRepo{},
		InvoiceScheduleHistoryRepo:        &repositories.InvoiceScheduleHistoryRepo{},
		InvoiceScheduleStudentRepo:        &repositories.InvoiceScheduleStudentRepo{},
		BulkPaymentRequestRepo:            &repositories.BulkPaymentRequestRepo{},
		BulkPaymentRequestFileRepo:        &repositories.BulkPaymentRequestFileRepo{},
		PartnerConvenienceStoreRepo:       &repositories.PartnerConvenienceStoreRepo{},
		PartnerBankRepo:                   &repositories.PartnerBankRepo{},
		BulkPaymentValidationsRepo:        &repositories.BulkPaymentValidationsRepo{},
		BulkPaymentValidationsDetailRepo:  &repositories.BulkPaymentValidationsDetailRepo{},
		UserRepo:                          &repositories.UserRepo{},
		StudentPaymentDetailRepo:          &repositories.StudentPaymentDetailRepo{},
		BankBranchRepo:                    &repositories.BankBranchRepo{},
		NewCustomerCodeHistoryRepo:        &repositories.NewCustomerCodeHistoryRepo{},
		OrderRepo:                         &repositories.OrderRepo{},
		StudentRepo:                       &repositories.StudentRepo{},
		PrefectureRepo:                    &repositories.PrefectureRepo{},
		InvoiceAdjustmentRepo:             &repositories.InvoiceAdjustmentRepo{},
		BillingAddressRepo:                &repositories.BillingAddressRepo{},
		BankAccountRepo:                   &repositories.BankAccountRepo{},
		BankRepo:                          &repositories.BankRepo{},
		BulkPaymentRepo:                   &repositories.BulkPaymentRepo{},
		StudentPaymentDetailActionLogRepo: &repositories.StudentPaymentDetailActionLogRepo{},
		UserBasicInfoRepo:                 &repositories.UserBasicInfoRepo{},
	}
}

func getInvoiceServiceRepositories(repos *Repositories) *invoiceService.InvoiceModifierServiceRepositories {
	return &invoiceService.InvoiceModifierServiceRepositories{
		InvoiceRepo:                      repos.InvoiceRepo,
		ActionLogRepo:                    repos.ActionLogRepo,
		InvoiceBillItemRepo:              repos.InvoiceBillItemRepo,
		BillItemRepo:                     repos.BillItemRepo,
		PaymentRepo:                      repos.PaymentRepo,
		OrganizationRepo:                 repos.OrganizationRepo,
		InvoiceScheduleRepo:              repos.InvoiceScheduleRepo,
		InvoiceScheduleHistoryRepo:       repos.InvoiceScheduleHistoryRepo,
		InvoiceScheduleStudentRepo:       repos.InvoiceScheduleStudentRepo,
		BulkPaymentRequestRepo:           repos.BulkPaymentRequestRepo,
		BulkPaymentRequestFileRepo:       repos.BulkPaymentRequestFileRepo,
		PartnerConvenienceStoreRepo:      repos.PartnerConvenienceStoreRepo,
		PartnerBankRepo:                  repos.PartnerBankRepo,
		BulkPaymentValidationsRepo:       repos.BulkPaymentValidationsRepo,
		BulkPaymentValidationsDetailRepo: repos.BulkPaymentValidationsDetailRepo,
		StudentPaymentDetailRepo:         repos.StudentPaymentDetailRepo,
		BankBranchRepo:                   repos.BankBranchRepo,
		NewCustomerCodeHistoryRepo:       repos.NewCustomerCodeHistoryRepo,
		OrderRepo:                        repos.OrderRepo,
		StudentRepo:                      repos.StudentRepo,
		PrefectureRepo:                   repos.PrefectureRepo,
		InvoiceAdjustmentRepo:            repos.InvoiceAdjustmentRepo,
		BulkPaymentRepo:                  repos.BulkPaymentRepo,
		BankAccountRepo:                  repos.BankAccountRepo,
		UserBasicInfoRepo:                repos.UserBasicInfoRepo,
	}
}

func getPaymentServiceRepositories(repos *Repositories) *paymentService.PaymentModifierServiceRepositories {
	return &paymentService.PaymentModifierServiceRepositories{
		PaymentRepo:                       repos.PaymentRepo,
		BulkPaymentRequestRepo:            repos.BulkPaymentRequestRepo,
		InvoiceRepo:                       repos.InvoiceRepo,
		BulkPaymentRequestFileRepo:        repos.BulkPaymentRequestFileRepo,
		BulkPaymentRequestFilePaymentRepo: repos.BulkPaymentRequestFilePaymentRepo,
		PartnerConvenienceStoreRepo:       repos.PartnerConvenienceStoreRepo,
		StudentPaymentDetailRepo:          repos.StudentPaymentDetailRepo,
		BankBranchRepo:                    repos.BankBranchRepo,
		NewCustomerCodeHistoryRepo:        repos.NewCustomerCodeHistoryRepo,
		PrefectureRepo:                    repos.PrefectureRepo,
		PartnerBankRepo:                   repos.PartnerBankRepo,
		BulkPaymentValidationsRepo:        repos.BulkPaymentValidationsRepo,
		BulkPaymentValidationsDetailRepo:  repos.BulkPaymentValidationsDetailRepo,
		InvoiceActionLogRepo:              repos.ActionLogRepo,
		BankAccountRepo:                   repos.BankAccountRepo,
		StudentRepo:                       repos.StudentRepo,
		BulkPaymentRepo:                   repos.BulkPaymentRepo,
		UserBasicInfoRepo:                 repos.UserBasicInfoRepo,
	}
}

func getImportMasterDataServiceRepositories(repos *Repositories) *importService.ImportMasterDataServiceRepositories {
	return &importService.ImportMasterDataServiceRepositories{
		InvoiceScheduleRepo: repos.InvoiceScheduleRepo,
		PartnerBankRepo:     repos.PartnerBankRepo,
		OrganizationRepo:    repos.OrganizationRepo,
	}
}

func getExportMasterDataServiceRepositories(repos *Repositories) *exportService.ExportMasterDataServiceRepositories {
	return &exportService.ExportMasterDataServiceRepositories{
		InvoiceScheduleRepo: repos.InvoiceScheduleRepo,
		BankBranchRepo:      repos.BankBranchRepo,
		BankRepo:            repos.BankRepo,
		BankMappingRepo:     repos.BankMappingRepo,
	}
}

func getDataMigrationServiceRepositories(repos *Repositories) *dataMigrationService.DataMigrationModifierServiceRepositories {
	return &dataMigrationService.DataMigrationModifierServiceRepositories{
		InvoiceRepo:         repos.InvoiceRepo,
		PaymentRepo:         repos.PaymentRepo,
		InvoiceBillItemRepo: repos.InvoiceBillItemRepo,
		StudentRepo:         repos.StudentRepo,
		BillItemRepo:        repos.BillItemRepo,
	}
}

func getOpenAPIServiceRepositories(repos *Repositories) *openAPIService.OpenAPIModifierServiceRepositories {
	return &openAPIService.OpenAPIModifierServiceRepositories{
		StudentPaymentDetailRepo:          repos.StudentPaymentDetailRepo,
		BillingAddressRepo:                repos.BillingAddressRepo,
		PrefectureRepo:                    repos.PrefectureRepo,
		UserRepo:                          repos.UserRepo,
		BankRepo:                          repos.BankRepo,
		BankBranchRepo:                    repos.BankBranchRepo,
		BankAccountRepo:                   repos.BankAccountRepo,
		StudentPaymentDetailActionLogRepo: repos.StudentPaymentDetailActionLogRepo,
	}
}
