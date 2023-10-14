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

func (s *suite) theOrganizationHasExistingBankMappings(ctx context.Context, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setResourcePathAndClaims(ctx, org)

	for i := 0; i < 3; i++ {
		err := InsertEntities(
			stepState,
			s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
			s.EntitiesCreator.CreatePartnerBank(ctx, s.InvoiceMgmtPostgresDBTrace, true),
			s.EntitiesCreator.CreateBankMapping(ctx, s.InvoiceMgmtPostgresDBTrace),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theOrganizationHasNoExistingBankMapping(ctx context.Context, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setResourcePathAndClaims(ctx, org)

	bankMappingRepo := &repositories.BankMappingRepo{}
	bankMapping, err := bankMappingRepo.FindAll(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err bankMappingRepo.FindAll: %w", err)
	}

	if len(bankMapping) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting no bank mapping got %d", len(bankMapping))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminExportsTheBankMappingData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.ExportBankMappingRequest{}
	stepState.Response, stepState.ResponseErr = invoice_pb.NewExportMasterDataServiceClient(s.InvoiceMgmtConn).ExportBankMapping(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theBankMappingCSVHasACorrectContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportBankMappingResponse)
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
	err = checkCSVHeaderForExport(
		[]string{"bank_mapping_id", "bank_id", "partner_bank_id", "is_archived", "remarks"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	bankMappingRepo := &repositories.BankMappingRepo{}
	bankMapping, err := bankMappingRepo.FindAll(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	// check the length of existing bank mapping should be equal or greater than the number of record.
	// greater than because this might cause a flaky test if other tests create bank mapping after the exporting of data
	if len(bankMapping) < len(lines)-1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there is an inequality with the exported data. Length of scheduled invoice: %d. Length of data row: %d", len(bankMapping), len(lines)-1)
	}

	// Check the content if equal
	for _, line := range lines[1:] {
		bankMappingID := line[0]
		bankID := line[1]
		partnerBankID := line[2]
		isArchived := line[3]
		remarks := line[4]

		e := &entities.BankMapping{}
		fields, _ := e.FieldMap()
		query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_mapping_id = $1", strings.Join(fields, ","), e.TableName())
		err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, bankMappingID).ScanOne(e)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on querying bank_mapping err: %v", err)
		}

		actualIsArchivedStr := "0"
		if e.IsArchived.Bool {
			actualIsArchivedStr = "1"
		}

		if err := multierr.Combine(
			isEqual(bankMappingID, e.BankMappingID.String, "bank_mapping_id"),
			isEqual(bankID, e.BankID.String, "bank_id"),
			isEqual(partnerBankID, e.PartnerBankID.String, "partner_bank_id"),
			isEqual(isArchived, actualIsArchivedStr, "is_archived"),
			isEqual(remarks, e.Remarks.String, "remarks"),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theBankMappingCSVOnlyContainsTheHeaderRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportBankMappingResponse)
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
	err = checkCSVHeaderForExport(
		[]string{"bank_mapping_id", "bank_id", "partner_bank_id", "is_archived", "remarks"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
