package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var NumberNames = [...]string{
	"first",
	"second",
	"third",
	"fourth",
	"fifth",
	"sixth",
	"seventh",
	"eighth",
	"ninth",
	"tenth",
	"eleventh",
	"twelveth",
	"thirdteenth",
	"fourteenth",
}

func validateCsvHeader(expectedNumberColumns int, columnNames, expectedColumnNames []string) error {
	if len(columnNames) != expectedNumberColumns {
		return fmt.Errorf("number of column should be %d", expectedNumberColumns)
	}

	for idx, expectedColumnName := range expectedColumnNames {
		if !strings.EqualFold(columnNames[idx], expectedColumnName) {
			return fmt.Errorf("%s column should be '%s'", NumberNames[idx], expectedColumnName)
		}
	}
	return nil
}

func GenerateHeaderTitles(importAction string) []string {
	var headerTitles []string
	switch importAction {
	case "INVOICE_SCHEDULE":
		headerTitles = []string{
			"invoice_schedule_id",
			"invoice_date",
			"is_archived",
			"remarks",
		}
	case "PARTNER_BANK":
		headerTitles = []string{
			"partner_bank_id",
			"consignor_code",
			"consignor_name",
			"bank_number",
			"bank_name",
			"bank_branch_number",
			"bank_branch_name",
			"deposit_items",
			"account_number",
			"is_archived",
			"remarks",
			"is_default",
			"record_limit",
		}
	}
	return headerTitles
}

func validateHalfWidthNumber(numValue, fieldName string) error {
	halfWidthNumValidation := fmt.Sprintf(
		"^[%s]+$",
		utils.RegexHalfWidthKanaValidationNumbers,
	)

	regex := regexp.MustCompile(halfWidthNumValidation)
	if strings.TrimSpace(numValue) != "" && !regex.MatchString(numValue) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%s field has invalid half width number", fieldName))
	}
	return nil
}

func validateHalfWidthCharacters(strValue, fieldName string) error {
	halfWidthNumValidation := fmt.Sprintf(
		"^[%s%s%s%s]+$",
		utils.RegexHalfWidthKanaValidationNumbers,
		utils.RegexHalfWidthKanaValidationCapitalAlphabets,
		utils.HalfWidthKanaValidationHalfWidthKatakana,
		utils.RegexHalfWidthKanaValidationSymbols,
	)

	regex := regexp.MustCompile(halfWidthNumValidation)
	if strings.TrimSpace(strValue) != "" && !regex.MatchString(strValue) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%s field has invalid half width character", fieldName))
	}
	return nil
}
