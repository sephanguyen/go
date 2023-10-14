package invoicemgmt

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"go.uber.org/multierr"
)

func (s *suite) theOrganizationHasExistingBankBranchData(ctx context.Context, org, isArchivedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setResourcePathAndClaims(ctx, org)

	for i := 0; i < 3; i++ {
		ctx, err := s.createBankAndBankBranchForOrganization(ctx, isArchivedStr)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theUserExportBankBranchData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.ExportBankBranchRequest{}
	stepState.Response, stepState.ResponseErr = invoice_pb.NewExportMasterDataServiceClient(s.InvoiceMgmtConn).ExportBankBranch(contextWithToken(ctx), req)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theBankBranchCSVHasCorrectContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportBankBranchResponse)

	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be greater than 1
	if len(lines) < 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting the context line to be greater than or equal to 1 got %d", len(lines))
	}

	// check the header record
	err = checkCSVHeaderForExport(
		[]string{"bank_branch_id", "bank_branch_code", "bank_branch_name", "bank_branch_phonetic_name", "bank_code", "is_archived"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	bankBranchRepo := &repositories.BankBranchRepo{}
	bankBranches, err := bankBranchRepo.FindExportableBankBranches(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	// check the length of existing bank branch should be equal or greater than the number of record.
	// greater than because this might cause a flaky test if other tests create bank branch after the exporting of data
	if len(bankBranches) < len(lines)-1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("There is an inequality with the exported data. Length of bank branch: %d. Length of data row: %d", len(bankBranches), len(lines)-1)
	}

	// Check the content if equal
	for _, line := range lines[1:] {
		// csv lines value
		bankBranchID := line[0]
		bankBranchCode := line[1]
		bankBranchName := line[2]
		BankBranchPhoneticName := line[3]
		bankCode := line[4]
		isArchived := line[5]

		bankBranchEntity := &entities.BankBranch{}
		fields, _ := bankBranchEntity.FieldMap()

		// find bank branch using bank branch id if existing and compare csv values
		bankBranchQuery := fmt.Sprintf("SELECT %s FROM %s WHERE bank_branch_id = $1", strings.Join(fields, ","), bankBranchEntity.TableName())
		err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, bankBranchQuery, bankBranchID).ScanOne(bankBranchEntity)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Error on selecting bank branch record: %w", err)
		}

		// find bank related record using bank id
		bankEntity := &entities.Bank{}
		bankFields, _ := bankEntity.FieldMap()

		bankQuery := fmt.Sprintf("SELECT %s FROM %s WHERE bank_id = $1", strings.Join(bankFields, ","), bankEntity.TableName())
		err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, bankQuery, bankBranchEntity.BankID).ScanOne(bankEntity)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Error on selecting related bank record: %w", err)
		}

		actualIsArchivedStr := "0"
		if bankBranchEntity.IsArchived.Bool {
			actualIsArchivedStr = "1"
		}

		if err := multierr.Combine(
			isEqual(bankBranchID, bankBranchEntity.BankBranchID.String, "bank_branch_id"),
			isEqual(bankBranchCode, bankBranchEntity.BankBranchCode.String, "bank_branch_code"),
			isEqual(bankBranchName, bankBranchEntity.BankBranchName.String, "bank_branch_name"),
			isEqual(BankBranchPhoneticName, bankBranchEntity.BankBranchPhoneticName.String, "bank_branch_phonetic_name"),
			isEqual(bankCode, bankEntity.BankCode.String, "bank_code"),
			isEqual(isArchived, actualIsArchivedStr, "is_archived"),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theOrganizationHasNoExistingBankBranch(ctx context.Context, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setResourcePathAndClaims(ctx, org)

	bankBranchRepo := &repositories.BankBranchRepo{}
	bankBranch, err := bankBranchRepo.FindExportableBankBranches(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err bankBranchRepo.FindExportableBankBranches: %w", err)
	}

	if len(bankBranch) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting no bank branch data got %d", len(bankBranch))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theBankBranchCSVOnlyContainsTheHeaderRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportBankBranchResponse)
	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be 1 for header only
	if len(lines) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting the context line to be 1 got %d", len(lines))
	}

	// check the header record
	err = checkCSVHeaderForExport(
		[]string{"bank_branch_id", "bank_branch_code", "bank_branch_name", "bank_branch_phonetic_name", "bank_code", "is_archived"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createBankAndBankBranchForOrganization(ctx context.Context, isArchivedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// default archived false
	var isArchived bool

	if isArchivedStr == "not archived" {
		isArchived = true
	}

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
		s.EntitiesCreator.CreateBankBranch(ctx, s.InvoiceMgmtPostgresDBTrace, isArchived),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
