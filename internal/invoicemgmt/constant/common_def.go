package constant

import (
	"time"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

const (
	PageLimit = 100

	RoleSchoolAdmin = "School Admin"
	RoleTeacher     = "Teacher"
	RoleParent      = "Parent"
	RoleStudent     = "Student"
	RoleOpenAPI     = "OpenAPI"

	UserGroupSchoolAdmin = "USER_GROUP_SCHOOL_ADMIN"

	EnableGCloudUploadFeatureFlag = "BACKEND_Invoice_InvoiceManagement_CreatePaymentRequest_GCloud_File_Upload"
	InvoicemgmtTemporaryDir       = "invoicemgmt-temp-dir-"

	EnableAutoSetCCFeatureFlag               = "BACKEND_Invoice_InvoiceManagement_AutoSetConvenienceStore"
	EnableKECFeedbackPh1                     = "Invoice_InvoiceManagement_BackOffice_KecFeedback"
	EnableBulkAddValidatePh2                 = "Invoice_InvoiceManagement_BackOffice_BulkAddAndValidatePayments"
	EnableSetDirectDebitFeatureFlag          = "BACKEND_Invoice_InvoiceManagement_SetDirectDebit"
	EnableImproveCronImportInvoiceChecker    = "BACKEND_Invoice_InvoiceManagement_ImproveImportInvoiceChecker"
	EnableImproveBulkIssueInvoice            = "BACKEND_Invoice_InvoiceManagement_ImproveBulkIssueInvoice"
	EnableInvoiceScheduleCronJobAlert        = "BACKEND_Invoice_InvoiceManagement_InvoiceScheduleCronJobAlert"
	EnableReviewOrderChecking                = "Payment_OrderManagement_BackOffice_Reviewed_Flag"
	EnablePaymentSequenceNumberManualSetting = "BACKEND_Invoice_InvoiceManagement_PaymentSequenceNumberManualSetting"
	EnableOptionalValidationInPaymentRequest = "BACKEND_Invoice_InvoiceManagement_OptionalValidationInPaymentRequest"
	EnableImproveBulkPaymentValidation       = "BACKEND_Invoice_InvoiceManagement_ImproveBulkPaymentValidation"
	EnableRetryFailedInvoiceSchedule         = "BACKEND_Invoice_InvoiceManagement_RetryFailedInvoiceSchedule"
	EnableBillingMessageInCSVMessages        = "Invoice_InvoiceManagement_ConvenienceStoreCsvMessageFields"
	EnableSingleIssueInvoiceWithPayment      = "Invoice_InvoiceManagement_BackOffice_SingleIssueInvoiceWithPayment"
	EnableFormatPaymentRequestFileFields     = "BACKEND_Invoice_InvoiceManagement_FormatPaymentRequestFileFields"
	EnableEncodePaymentRequestFiles          = "BACKEND_Invoice_InvoiceManagement_EncodePaymentRequestFiles"

	NatsEventRetryQuerySleep = 500 * time.Millisecond

	// postgres errors
	PgConnForeignKeyError = "pgconn: foreign key error"
	TableRLSError         = "new row violates row-level security policy for table"

	BulkAddPaymentAndIssueMaxRetry = 100

	NumericRegex             = `^[0-9]+$`
	AdjustmentBillingKeyword = "[金額調整分]"
)

var (
	SingleInvoicePaymentMethods = map[string]bool{
		invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(): true,
		invoice_pb.PaymentMethod_CASH.String():              true,
		invoice_pb.PaymentMethod_BANK_TRANSFER.String():     true,
	}

	AddInvoicePaymentAllowedMethods = map[string]struct{}{
		invoice_pb.PaymentMethod_DIRECT_DEBIT.String():      {},
		invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(): {},
		invoice_pb.PaymentMethod_CASH.String():              {},
		invoice_pb.PaymentMethod_BANK_TRANSFER.String():     {},
	}

	RefundInvoiceAllowedMethods = map[string]struct{}{
		invoice_pb.RefundMethod_REFUND_CASH.String():          {},
		invoice_pb.RefundMethod_REFUND_BANK_TRANSFER.String(): {},
	}

	PartnerBankDepositItems = map[int]string{
		1: "ORDINARY_BANK_ACCOUNT",
		2: "CHECKING_ACCOUNT",
		3: "SAVINGS_ACCOUNT",
		4: "OTHER_BANK_ACCOUNT_TYPE",
	}

	PaymentMethodsConvertToEnums = map[string]invoice_pb.PaymentMethod{
		"CONVENIENCE_STORE": invoice_pb.PaymentMethod_CONVENIENCE_STORE,
		"DIRECT_DEBIT":      invoice_pb.PaymentMethod_DIRECT_DEBIT,
	}

	StudentPaymentMethods = map[string]bool{
		invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(): true,
		invoice_pb.PaymentMethod_DIRECT_DEBIT.String():      true,
	}

	ApprovePaymentAllowedMethods = map[string]bool{
		invoice_pb.PaymentMethod_CASH.String():          true,
		invoice_pb.PaymentMethod_BANK_TRANSFER.String(): true,
	}
)
