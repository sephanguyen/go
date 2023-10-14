package invoicemgmt

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) theOrganizationHasExistingImportInvoiceSchedulesIn(ctx context.Context, org, fileContent, timezoneStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.setResourcePathAndClaims(ctx, org)

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	parseDateToday, err := time.Parse("2006/01/02 00:00", fmt.Sprintf("%v/%02d/%02d 00:00", now.Year(), int(now.Month()), now.Day()))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error unable to parse date format: %v", err)
	}

	invoiceDateInlocation, err := convertDatetoCountryTZ(parseDateToday, timezoneStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error convertDatetoCountryTZ: %v", err)
	}

	// create import schedule invoice records on valid dates
	req := &invoice_pb.ImportInvoiceScheduleRequest{}
	// add 10 more days on invoice date for testing and not align with other tests
	ctx, req.Payload, err = s.generateImportScheduleInvoicePayload(ctx, fileContent, invoiceDateInlocation.AddDate(0, 0, 10))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewImportMasterDataServiceClient(s.InvoiceMgmtConn).ImportInvoiceSchedule(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theOrganizationHasNoExistingInvoiceSchedule(ctx context.Context, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setResourcePathAndClaims(ctx, org)

	scheduledInvoiceRepo := &repositories.InvoiceScheduleRepo{}
	scheduledInvoice, err := scheduledInvoiceRepo.FindAll(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err scheduledInvoiceRepo.GetByStatusAndInvoiceDate: %w", err)
	}

	if len(scheduledInvoice) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting no scheduled invoice got %d", len(scheduledInvoice))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminExportTheInvoiceScheduleData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.ExportInvoiceScheduleRequest{}
	stepState.Response, stepState.ResponseErr = invoice_pb.NewExportMasterDataServiceClient(s.InvoiceMgmtConn).ExportInvoiceSchedule(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvoiceScheduleCSVHasACorrectContent(ctx context.Context, defaultLocation string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportInvoiceScheduleResponse)
	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be greater than 1
	if len(lines) < 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting the context line to be greater than 2 got %d", len(lines))
	}

	// check the header record
	err = checkCSVHeaderForExport(
		[]string{"invoice_schedule_id", "invoice_date", "is_archived", "remarks"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	scheduledInvoiceRepo := &repositories.InvoiceScheduleRepo{}
	scheduledInvoice, err := scheduledInvoiceRepo.FindAll(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
	)

	// check the length of existing invoice schedule should be equal or greater than the number of record.
	// greater than because this might cause a flaky test if other tests create invoice schedule after the exporting of data
	if len(scheduledInvoice) < len(lines)-1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("There is an inequality with the exported data. Length of scheduled invoice: %d. Length of data row: %d", len(scheduledInvoice), len(lines)-1)
	}

	// Check the content if equal
	for _, line := range lines[1:] {
		invoiceScheduleID := line[0]
		invoiceDate := line[1]
		isArchived := line[2]
		remarks := line[3]

		e, err := scheduledInvoiceRepo.RetrieveInvoiceScheduleByID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceScheduleID)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("scheduledInvoiceRepo.RetrieveInvoiceScheduleByID err: %v", err)
		}

		invoiceDateInlocation, err := convertDatetoCountryTZ(e.InvoiceDate.Time, defaultLocation)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error convertDatetoCountryTZ: %v", err)
		}

		actualInvoiceDateStr := invoiceDateInlocation.Format("2006/01/02")

		actualIsArchivedStr := "0"
		if e.IsArchived.Bool {
			actualIsArchivedStr = "1"
		}

		if err := multierr.Combine(
			isEqual(invoiceScheduleID, e.InvoiceScheduleID.String, "invoice_schedule_id"),
			isEqual(invoiceDate, actualInvoiceDateStr, "invoice_date"),
			isEqual(isArchived, actualIsArchivedStr, "is_archived"),
			isEqual(remarks, e.Remarks.String, "remarks"),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvoiceScheduleCSVOnlyContainsHeaderRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ExportInvoiceScheduleResponse)
	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// length of line should be 1
	if len(lines) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting the context line to be 1 got %d", len(lines))
	}

	// check the header record
	err = checkCSVHeaderForExport(
		[]string{"invoice_schedule_id", "invoice_date", "is_archived", "remarks"},
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
