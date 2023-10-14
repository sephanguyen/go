package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PartnerBankDetails struct {
	PartnerBankID    string
	ConsignorCode    string
	ConsignorName    string
	BankNumber       string
	BankName         string
	BankBranchNumber string
	BankBranchName   string
	DepositItems     string
	AccountNumber    string
	IsArchived       bool
	Remarks          string
	IsDefault        bool
	RecordLimit      int
}

func (s *ImportMasterDataService) ImportPartnerBank(ctx context.Context, req *invoice_pb.ImportPartnerBankRequest) (*invoice_pb.ImportPartnerBankResponse, error) {
	lines, err := s.validateHeaderColumnRequest(req.Payload, invoice_pb.ImportMasterAction_PARTNER_BANK.String())
	if err != nil {
		return nil, err
	}

	errors := []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// Skip the first item, which is the CSV header
		for i, line := range lines[1:] {
			// Validate each line after the header
			validPartnerBankDetails, err := validateImportPartnerBankCSV(line)
			if err != nil {
				errors = append(errors, s.generateImportPartnerBankError(int32(i)+2, fmt.Sprintf("unable to parse partner bank detail: %s", err)))
				continue
			}

			if validPartnerBankDetails.PartnerBankID != "" {
				// record should be existing
				_, err := s.PartnerBankRepo.RetrievePartnerBankByID(ctx, tx, validPartnerBankDetails.PartnerBankID)
				if err != nil {
					errors = append(errors, s.generateImportPartnerBankError(int32(i)+2, fmt.Sprintf("cannot find partner bank with error '%v'", err)))
					continue
				}
			}

			// generate partner bank entity
			var partnerBankEntity *entities.PartnerBank
			partnerBankEntity, err = generatePartnerBankEntity(validPartnerBankDetails)
			if err != nil {
				errors = append(errors, s.generateImportPartnerBankError(int32(i)+2, fmt.Sprintf("cannot generate partner bank entity '%v'", err)))
				continue
			}
			// Create partner bank record if there's no partner bank ID provided; otherwise, update the existing record
			if err := s.PartnerBankRepo.Upsert(ctx, tx, partnerBankEntity); err != nil {
				errors = append(errors, s.generateImportPartnerBankError(int32(i)+2, fmt.Sprintf("unable to upsert partner bank %s", err)))
			}
		}

		// This causes the changes to be rolled back if there's at least one error
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error when importing partner bank: %s", err.Error())
	}

	return &invoice_pb.ImportPartnerBankResponse{
		Errors: errors,
	}, nil
}

// Validates the CSV contents according to master guidelines with additional input validations
//
//nolint:gocyclo
func validateImportPartnerBankCSV(line []string) (*PartnerBankDetails, error) {
	// validate header title values
	headerTitles := GenerateHeaderTitles(invoice_pb.ImportMasterAction_PARTNER_BANK.String())
	partnerBankCsv := map[string]string{}
	var isArchived bool
	var isDefault bool
	var err error
	var recordLimit int

	if len(headerTitles) != len(line) {
		return nil, status.Error(codes.InvalidArgument, "header values not matched")
	}

	// map to each header value
	for i := range line {
		// validate csv values mandatory except partner_bank_id, is archive and remarks
		csvValueTrimmed := strings.TrimSpace(line[i])
		if headerTitles[i] != "is_archived" && headerTitles[i] != "partner_bank_id" && headerTitles[i] != "remarks" && headerTitles[i] != "is_default" && headerTitles[i] != "record_limit" {
			if csvValueTrimmed == "" {
				return nil, fmt.Errorf("field %v is required", headerTitles[i])
			}
		}
		partnerBankCsv[headerTitles[i]] = csvValueTrimmed
	}

	isArchivedCsvStr := partnerBankCsv["is_archived"]
	isDefaultCsvStr := partnerBankCsv["is_default"]
	partnerBankIDCsvStr := partnerBankCsv["partner_bank_id"]
	recordLimitCsvStr := partnerBankCsv["record_limit"]

	if isArchivedCsvStr != "" {
		isArchived, err = strconv.ParseBool(isArchivedCsvStr)
		if err != nil {
			return nil, fmt.Errorf("invalid archive value")
		}
	}

	// validate archive csv
	if (partnerBankIDCsvStr != "" && !isArchived) ||
		(partnerBankIDCsvStr == "" && isArchived) {
		return nil, fmt.Errorf("partner_bank_id and is_archived can only be both present or absent")
	}

	// validate consignor code digit
	err = validateHalfWidthNumber(partnerBankCsv["consignor_code"], "consignor_code")
	if err != nil {
		return nil, err
	}

	_, err = strconv.Atoi(partnerBankCsv["consignor_code"])
	if err != nil {
		return nil, fmt.Errorf("invalid consignor code digit format")
	}

	if len(partnerBankCsv["consignor_code"]) > 10 {
		return nil, fmt.Errorf("invalid consignor code digit limit")
	}

	// validate consignor name limit
	err = validateHalfWidthCharacters(partnerBankCsv["consignor_name"], "consignor_name")
	if err != nil {
		return nil, err
	}

	if len([]rune(partnerBankCsv["consignor_name"])) > 40 {
		return nil, fmt.Errorf("invalid consignor name character limit")
	}

	// validate bank number digit
	err = validateHalfWidthNumber(partnerBankCsv["bank_number"], "bank_number")
	if err != nil {
		return nil, err
	}

	_, err = strconv.Atoi(partnerBankCsv["bank_number"])
	if err != nil {
		return nil, fmt.Errorf("invalid bank number digit format")
	}

	if len(partnerBankCsv["bank_number"]) > 4 {
		return nil, fmt.Errorf("invalid bank number digit limit")
	}

	// validate bank name limit
	err = validateHalfWidthCharacters(partnerBankCsv["bank_name"], "bank_name")
	if err != nil {
		return nil, err
	}

	if len([]rune(partnerBankCsv["bank_name"])) > 15 {
		return nil, fmt.Errorf("invalid bank name character limit")
	}

	// validate bank branch number limit
	err = validateHalfWidthNumber(partnerBankCsv["bank_branch_number"], "bank_branch_number")
	if err != nil {
		return nil, err
	}

	_, err = strconv.Atoi(partnerBankCsv["bank_branch_number"])
	if err != nil {
		return nil, fmt.Errorf("invalid bank branch number digit format")
	}
	if len(partnerBankCsv["bank_branch_number"]) > 3 {
		return nil, fmt.Errorf("invalid bank branch number digit limit")
	}

	// validate bank branch name limit
	err = validateHalfWidthCharacters(partnerBankCsv["bank_branch_name"], "bank_branch_name")
	if err != nil {
		return nil, err
	}

	if len([]rune(partnerBankCsv["bank_branch_name"])) > 15 {
		return nil, fmt.Errorf("invalid bank branch name character limit")
	}

	// validate deposit items limit
	err = validateHalfWidthNumber(partnerBankCsv["deposit_items"], "deposit_items")
	if err != nil {
		return nil, err
	}
	depositItemsDigit, err := strconv.Atoi(partnerBankCsv["deposit_items"])
	if err != nil {
		return nil, fmt.Errorf("invalid deposit items digit format")
	}
	if len(partnerBankCsv["deposit_items"]) > 1 {
		return nil, fmt.Errorf("invalid deposit items digit limit")
	}
	if constant.PartnerBankDepositItems[depositItemsDigit] == "" {
		return nil, fmt.Errorf("invalid deposit items account")
	}

	// validate account number
	err = validateHalfWidthNumber(partnerBankCsv["account_number"], "account_number")
	if err != nil {
		return nil, err
	}

	if len(partnerBankCsv["account_number"]) != 7 {
		return nil, fmt.Errorf("the account number can only accept 7 digit numbers")
	}
	// validate default flag
	if isDefaultCsvStr != "" {
		isDefault, err = strconv.ParseBool(isDefaultCsvStr)
		if err != nil {
			return nil, fmt.Errorf("invalid default value")
		}
	}

	recordLimit, err = setRecordLimit(recordLimitCsvStr)
	if err != nil {
		return nil, err
	}

	partnerBankDetails := &PartnerBankDetails{
		PartnerBankID:    partnerBankIDCsvStr,
		ConsignorCode:    partnerBankCsv["consignor_code"],
		ConsignorName:    partnerBankCsv["consignor_name"],
		BankNumber:       partnerBankCsv["bank_number"],
		BankName:         partnerBankCsv["bank_name"],
		BankBranchNumber: partnerBankCsv["bank_branch_number"],
		BankBranchName:   partnerBankCsv["bank_branch_name"],
		DepositItems:     constant.PartnerBankDepositItems[depositItemsDigit],
		AccountNumber:    partnerBankCsv["account_number"],
		IsArchived:       isArchived,
		Remarks:          partnerBankCsv["remarks"],
		IsDefault:        isDefault,
		RecordLimit:      recordLimit,
	}

	return partnerBankDetails, nil
}

func (s *ImportMasterDataService) generateImportPartnerBankError(rowNumber int32, errorMsg string) *invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError {
	return &invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
		RowNumber: rowNumber,
		Error:     errorMsg,
	}
}

func generatePartnerBankEntity(partnerBankDetails *PartnerBankDetails) (*entities.PartnerBank, error) {
	partnerBank := new(entities.PartnerBank)
	database.AllNullEntity(partnerBank)

	errs := []error{}

	if partnerBankDetails.PartnerBankID != "" {
		errs = append(errs, partnerBank.PartnerBankID.Set(partnerBankDetails.PartnerBankID))
	}
	now := time.Now()
	err := multierr.Combine(
		partnerBank.ConsignorCode.Set(partnerBankDetails.ConsignorCode),
		partnerBank.ConsignorName.Set(partnerBankDetails.ConsignorName),
		partnerBank.BankNumber.Set(partnerBankDetails.BankNumber),
		partnerBank.BankName.Set(partnerBankDetails.BankName),
		partnerBank.BankBranchNumber.Set(partnerBankDetails.BankBranchNumber),
		partnerBank.BankBranchName.Set(partnerBankDetails.BankBranchName),
		partnerBank.DepositItems.Set(partnerBankDetails.DepositItems),
		partnerBank.AccountNumber.Set(partnerBankDetails.AccountNumber),
		partnerBank.IsArchived.Set(partnerBankDetails.IsArchived),
		partnerBank.Remarks.Set(partnerBankDetails.Remarks),
		partnerBank.CreatedAt.Set(now),
		partnerBank.UpdatedAt.Set(now),
		partnerBank.IsDefault.Set(partnerBankDetails.IsDefault),
		partnerBank.RecordLimit.Set(partnerBankDetails.RecordLimit),
	)
	if err != nil {
		errs = append(errs, err)
	}

	if err := multierr.Combine(errs...); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return partnerBank, nil
}

func setRecordLimit(recordLimitString string) (int, error) {
	var recordLimit int

	if strings.TrimSpace(recordLimitString) == "" {
		return 0, nil
	}

	recordLimit, err := strconv.Atoi(recordLimitString)
	if err != nil {
		return 0, fmt.Errorf("invalid record_limit format")
	}
	if recordLimit < 0 {
		return 0, fmt.Errorf("invalid record limit: should be greater than or equal to 0")
	}

	return recordLimit, nil
}
