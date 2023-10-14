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

func (s *suite) theOrganizationHasExistingBankData(ctx context.Context, org string, isArchived string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setResourcePathAndClaims(ctx, org)

	for i := 0; i < 3; i++ {
		ctx, err := s.createBankAndBankForOrganization(ctx, isArchived)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createBankAndBankForOrganization(ctx context.Context, isArchivedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// default archived false
	var isArchived bool

	if isArchivedStr == "not archived" {
		isArchived = true
	}

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, isArchived),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theOrganizationHasNoExistingBank(ctx context.Context, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setResourcePathAndClaims(ctx, org)

	bankRepo := &repositories.BankRepo{}
	bank, err := bankRepo.FindAll(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err bankRepo: %w", err)
	}

	if len(bank) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting no bank got %d", len(bank))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminExportTheBankData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.ExportBankRequest{}
	stepState.Response, stepState.ResponseErr = invoice_pb.NewExportMasterDataServiceClient(s.InvoiceMgmtConn).ExportBank(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theBankCSVHasACorrectContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportBankResponse)
	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be greater than 1
	if len(lines) < 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the context line to be greater than 2 got %d", len(lines))
	}

	// check the header record
	err = checkCSVHeader(
		[]string{"bank_id", "bank_code", "bank_name", "bank_phonetic_name", "is_archived"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	bankRepo := &repositories.BankRepo{}
	bank, err := bankRepo.FindAll(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	// check the length of existing invoice schedule should be equal or greater than the number of record.
	// greater than because this might cause a flaky test if other tests create invoice schedule after the exporting of data
	if len(bank) < len(lines)-1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there is an inequality with the exported data. Length of scheduled invoice: %d. Length of data row: %d", len(bank), len(lines)-1)
	}

	// Check the content if equal
	for _, line := range lines[1:] {
		bankID := line[0]
		bankCode := line[1]
		bankName := line[2]
		BankNamePhonetic := line[3]
		isArchived := line[4]

		e := &entities.Bank{}
		fields, _ := e.FieldMap()
		query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_id = $1", strings.Join(fields, ","), e.TableName())
		err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, bankID).ScanOne(e)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on querying bank err: %v", err)
		}

		actualIsArchivedStr := "0"
		if e.IsArchived.Bool {
			actualIsArchivedStr = "1"
		}

		if err := multierr.Combine(
			isEqual(bankID, e.BankID.String, "bank_id"),
			isEqual(bankCode, e.BankCode.String, "bank_code"),
			isEqual(bankName, e.BankName.String, "bank_name"),
			isEqual(BankNamePhonetic, e.BankNamePhonetic.String, "bank_name_phonetic"),
			isEqual(isArchived, actualIsArchivedStr, "is_archived"),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theBankCSVOnlyContainsHeaderRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportBankResponse)
	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be 1
	if len(lines) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the context line to be 1 got %d", len(lines))
	}

	// check the header record
	err = checkCSVHeader(
		[]string{"bank_id", "bank_code", "bank_name", "bank_phonetic_name", "is_archived"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func checkCSVHeader(expected []string, actual []string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("expected header length to be %d got %d", len(expected), len(actual))
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			return fmt.Errorf("expected header name to be %s got %s", expected[i], actual[i])
		}
	}

	return nil
}
