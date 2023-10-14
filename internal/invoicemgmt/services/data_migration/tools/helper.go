package helper

import (
	"fmt"
	"strings"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

const (
	InvoiceRawData = "invoice_raw_data"
	UserEntity     = "USER_ENTITY"
)

// nolint:misspell
func GetHeaderTitles(entityName string) []string {
	var headerTitles []string

	switch entityName {
	case invoice_pb.DataMigrationEntityName_INVOICE_ENTITY.String():
		headerTitles = []string{
			"invoice_csv_id",
			"invoice_id",
			"student_id",
			"type",
			"status",
			"sub_total",
			"total",
			"created_at",
			"invoice_sequence_number",
			"is_exported",
			"reference1",
			"reference2",
		}
	case invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String():
		headerTitles = []string{
			"payment_csv_id",
			"payment_id",
			"invoice_id",
			"payment_method",
			"payment_status",
			"due_date",
			"expiry_date",
			"payment_date",
			"student_id",
			"payment_sequence_number",
			"is_exported",
			"created_at",
			"result_code",
			"amount",
			"reference",
		}
	case InvoiceRawData:
		headerTitles = []string{
			"id",
			"payment_id",
			"invoice_date",
			"invoice_month",
			"m_student_id",
			"invoice_type",
			"m_department_id",
			"make_type",
			"status",
			"invoice_timing",
			"invoice_amout",
			"comsumption_tax_amount",
			"deposit_flag",
			"recieve_type",
			"collection_date",
			"prompt_report_date",
			"recieve_date",
			"printed_limit_date_1",
			"usable_limit_date_2",
			"m_invoice_institute_id",
			"convient_store_name",
			"note",
			"m_invoice_pattern_id",
			"m_bank_id",
			"m_bank_branch_id",
			"account_type",
			"account_number",
			"holder_name",
			"invoice_flag",
			"customer_note",
			"invoice_family_name",
			"invoice_first_name",
			"postal_code",
			"prefecture",
			"city",
			"apartment_name",
			"t_invoice_management_id",
			"output_cv_date_time",
			"output_cv_user_id",
			"entry_date_time",
			"entry_user_id",
			"update_date_time",
			"update_user_id",
			"delete_flg",
		}
	case UserEntity:
		headerTitles = []string{
			"user_id",
			"user_external_id",
			"email",
			"resource_path",
		}
	}

	return headerTitles
}

func ValidateCsvHeader(expectedNumberColumns int, columnNames, expectedColumnNames []string) error {
	if len(columnNames) != expectedNumberColumns {
		return fmt.Errorf("csv file invalid format - number of column should be %d", expectedNumberColumns)
	}

	for idx, expectedColumnName := range expectedColumnNames {
		if !strings.EqualFold(columnNames[idx], expectedColumnName) {
			return fmt.Errorf("csv file invalid format - %s column (toLowerCase) should be '%s'", columnNames[idx], expectedColumnName)
		}
	}
	return nil
}
