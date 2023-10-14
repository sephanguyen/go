package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"

	"github.com/spf13/cobra"
)

func genInvoicemgmtRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"invoice":                           &repositories.InvoiceRepo{},
		"payment":                           &repositories.PaymentRepo{},
		"invoice_action_log":                &repositories.InvoiceActionLogRepo{},
		"invoice_bill_item":                 &repositories.InvoiceBillItemRepo{},
		"bill_item":                         &repositories.BillItemRepo{},
		"invoice_schedule":                  &repositories.InvoiceScheduleRepo{},
		"invoice_schedule_history":          &repositories.InvoiceScheduleHistoryRepo{},
		"invoice_schedule_student":          &repositories.InvoiceScheduleStudentRepo{},
		"organization":                      &repositories.OrganizationRepo{},
		"bulk_payment_request":              &repositories.BulkPaymentRequestRepo{},
		"bulk_payment_request_file":         &repositories.BulkPaymentRequestFileRepo{},
		"bulk_payment_request_file_payment": &repositories.BulkPaymentRequestFilePaymentRepo{},
		"partner_convenience_store":         &repositories.PartnerConvenienceStoreRepo{},
		"partner_bank":                      &repositories.PartnerBankRepo{},
		"bulk_payment_validations":          &repositories.BulkPaymentValidationsRepo{},
		"bulk_payment_validations_detail":   &repositories.BulkPaymentValidationsDetailRepo{},
		"user":                              &repositories.UserRepo{},
		"student_payment_detail":            &repositories.StudentPaymentDetailRepo{},
		"prefecture":                        &repositories.PrefectureRepo{},
		"bank":                              &repositories.BankRepo{},
		"bank_branch":                       &repositories.BankBranchRepo{},
		"new_customer_code_history":         &repositories.NewCustomerCodeHistoryRepo{},
		"billing_address":                   &repositories.BillingAddressRepo{},
		"bank_mapping":                      &repositories.BankMappingRepo{},
		"bank_account":                      &repositories.BankAccountRepo{},
		"order":                             &repositories.OrderRepo{},
		"student":                           &repositories.StudentRepo{},
		"invoice_adjustment":                &repositories.InvoiceAdjustmentRepo{},
		"bulk_payment":                      &repositories.BulkPaymentRepo{},
		"student_payment_detail_action_log": &repositories.StudentPaymentDetailActionLogRepo{},
		"user_basic_info":                   &repositories.UserBasicInfoRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "invoicemgmt", repos)

	interfaces := map[string][]string{
		"internal/invoicemgmt/services/filestorage": {
			"FileStorage",
		},
		"internal/invoicemgmt/services/utils": {
			"ITempFileCreator",
		},
		"internal/invoicemgmt/services/sequence_number": {
			"ISequenceNumberService",
			"IPaymentSequenceNumberService",
		},
		"internal/invoicemgmt/services": {
			"OrderService",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenInvoicemgmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "invoicemgmt [../../mock/invoicemgmt]",
		Short: "generate invoicemgmt repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genInvoicemgmtRepo,
	}
}
