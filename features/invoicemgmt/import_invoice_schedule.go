package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	invoicemgmt_entities "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) signedinUserImportsInvoiceScheduleFileWithFileContentType(ctx context.Context, signedInUser string, fileContentType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error

	req := &invoice_pb.ImportInvoiceScheduleRequest{}
	partnerDtNow, err := convertDatetoCountryTZ(time.Now(), COUNTRY_JP)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, req.Payload, err = s.generateImportScheduleInvoicePayload(ctx, fileContentType, partnerDtNow)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewImportMasterDataServiceClient(s.InvoiceMgmtConn).ImportInvoiceSchedule(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) errorListIsEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := s.StepState.Response.(*invoice_pb.ImportInvoiceScheduleResponse)
	if len(resp.Errors) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("expected empty errors arr from response but received %v: %v", len(resp.Errors), resp.Errors))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importScheduleReflectsInTheDBBasedOnFileContentType(ctx context.Context, fileContentType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Retrieve the actual invoice schedules stored in DB
	invoiceSchedules, err := s.retrieveImportSchedule(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var scheduledInvoicesCnt int
	var foundArchivedInvoice bool

	for _, invoiceSchedule := range invoiceSchedules {
		if invoiceSchedule.Status.String == invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String() {
			scheduledInvoicesCnt++

			if invoiceSchedule.UserID.String != stepState.CurrentUserID && fileContentType == "multiple-valid-dates" {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected invoice schedule's user id %v but got %v", stepState.CurrentUserID, invoiceSchedule.UserID.String)
			}
		}

		if invoiceSchedule.InvoiceScheduleID.String == stepState.InvoiceScheduleID && invoiceSchedule.Status.String == invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_CANCELLED.String() {
			foundArchivedInvoice = true
		}
	}

	if fileContentType == "multiple-valid-dates" {
		if scheduledInvoicesCnt != 5 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected 5 scheduled invoices but found %v", scheduledInvoicesCnt)
		}
	}

	if fileContentType == "one-invoice-to-archive" {
		if !foundArchivedInvoice {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot find cancelled invoice schedule id: %v", stepState.InvoiceScheduleID)
		}
	}

	if fileContentType == "duplicate-valid-dates" {
		// In this file content type, this is the date that was imported in CSV
		testDate := time.Now().AddDate(1, 0, 0)

		// Add JST timezone in the test date because this is how the invoice date is processed in the endpoint
		testDateJST, err := convertDatetoCountryTZ(testDate, COUNTRY_JP)
		if err != nil {
			return nil, fmt.Errorf("error convertDatetoCountryTZ: %v", err)
		}

		scheduledCount := 0
		for _, invoiceSchedule := range invoiceSchedules {
			// check for invoice date same with test date and with SCHEDULED status
			if invoiceSchedule.InvoiceDate.Time.Format("2006-01-02") == testDateJST.UTC().Format("2006-01-02") &&
				invoiceSchedule.Status.String == invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String() {
				scheduledCount++
			}
		}

		if scheduledCount > 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting SCHEDULED invoice of invoice_date %v to be 1 got %v", testDateJST.UTC(), scheduledCount)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesImportError(ctx context.Context, importError string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if !strings.ContainsAny(s.StepState.ResponseErr.Error(), importError) {
		return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("expected '%s' error but received '%s'", importError, s.StepState.ResponseErr.Error()))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnExistingImportSchedule(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Prevents multiple imports happening at the same time
	// Before this step, there was already an invoice schedule
	time.Sleep(time.Second * 1)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) errorListIsCorrect(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	expectedErr := []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
		{
			RowNumber: 2,
			Error:     "unable to parse invoice schedule: invoice schedule should be a future date",
		},
		{
			RowNumber: 3,
			Error:     "unable to parse invoice schedule: invoice schedule should be a future date",
		},
		{
			RowNumber: 5,
			Error:     "unable to parse invoice schedule: invalid date format",
		},
		{
			RowNumber: 6,
			Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
		},
		{
			RowNumber: 7,
			Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
		},
		{
			RowNumber: 8,
			Error:     "unable to parse invoice schedule: invalid IsArchived value",
		},
		{
			RowNumber: 9,
			Error:     "unable to parse invoice schedule: invoice date is required",
		},
		{
			RowNumber: 10,
			Error:     "unable to parse invoice schedule: cannot find invoice_schedule_id with error 'no rows in result set'",
		},
	}

	// remove the error of present date
	if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableKECFeedbackPh1) {
		expectedErr = []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
			{
				RowNumber: 2,
				Error:     "unable to parse invoice schedule: invoice schedule should be a present date or future date",
			},
			{
				RowNumber: 4,
				Error:     "unable to parse invoice schedule: invalid date format",
			},
			{
				RowNumber: 5,
				Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
			},
			{
				RowNumber: 6,
				Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
			},
			{
				RowNumber: 7,
				Error:     "unable to parse invoice schedule: invalid IsArchived value",
			},
			{
				RowNumber: 8,
				Error:     "unable to parse invoice schedule: invoice date is required",
			},
			{
				RowNumber: 9,
				Error:     "unable to parse invoice schedule: cannot find invoice_schedule_id with error 'no rows in result set'",
			},
		}
	}

	resp := s.StepState.Response.(*invoice_pb.ImportInvoiceScheduleResponse)
	if len(resp.Errors) != len(expectedErr) {
		return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("expected %v errors length from response but got %v", len(expectedErr), len(resp.Errors)))
	}

	for i, csvError := range resp.Errors {
		if csvError.RowNumber != expectedErr[i].RowNumber {
			return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("expected RowNumber %v from response error list but got %v", expectedErr[i].RowNumber, csvError.RowNumber))
		}

		if csvError.Error != expectedErr[i].Error {
			return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("expected Error %v from response error list but got %v", expectedErr[i].Error, csvError.Error))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

type ImportSchedules []*invoicemgmt_entities.InvoiceSchedule

func (u *ImportSchedules) Add() database.Entity {
	e := &invoicemgmt_entities.InvoiceSchedule{}
	*u = append(*u, e)

	return e
}

// Retrieves all the invoice schedules
func (s *suite) retrieveImportSchedule(ctx context.Context) ([]*invoicemgmt_entities.InvoiceSchedule, error) {

	fields, _ := (&invoicemgmt_entities.InvoiceSchedule{}).FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM invoice_schedule WHERE resource_path = $1`, strings.Join(fields, ","))
	importSchedules := ImportSchedules{}
	err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, s.ResourcePath).ScanAll(&importSchedules)
	if err != nil {
		return importSchedules, fmt.Errorf("error retrieveImportSchedule: %v", err)
	}

	return importSchedules, err
}

// Cancels all existing invoice schedules
func (s *suite) thereIsNoExistingImportSchedule(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `UPDATE invoice_schedule SET status = $1 WHERE resource_path = $2`

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_CANCELLED.String(), s.ResourcePath); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("err deleting all invoice schedule history: %v", err))
	}

	// Prevents data race
	time.Sleep(time.Second * 1)

	return StepStateToContext(ctx, stepState), nil
}

// Creates the payload in bytes, containing the CSV contents
func (s *suite) generateImportScheduleInvoicePayload(ctx context.Context, fileContentType string, invoiceDate time.Time) (context.Context, []byte, error) {
	stepState := StepStateFromContext(ctx)

	var payload []byte

	// Set time to 00:00
	invoiceDate = resetTimeComponent(invoiceDate)

	switch fileContentType {
	case "multiple-valid-dates":
		// initialize to empty again whenever import multiple dates as it's been used on other test also
		s.InvoiceScheduleDates = []time.Time{}

		csv := `invoice_schedule_id,invoice_date,is_archived,remarks
		,%v,,multiple-valid-dates-remarks1
		,%v,,multiple-valid-dates-remarks2
		,%v,,multiple-valid-dates-remarks3
		,%v,,multiple-valid-dates-remarks4
		,%v,,multiple-valid-dates-remarks5-next-day`

		// Add one month
		partnerDt := invoiceDate.AddDate(0, 1, 0)
		s.InvoiceScheduleDates = append(s.InvoiceScheduleDates, partnerDt)
		futureDateNxtMonth1 := fmt.Sprintf("%v/%02d/%02d", partnerDt.Year(), int(partnerDt.Month()), partnerDt.Day())

		partnerDt = partnerDt.AddDate(0, 1, 0)
		s.InvoiceScheduleDates = append(s.InvoiceScheduleDates, partnerDt)
		futureDateNxtMonth2 := fmt.Sprintf("%v/%02d/%02d", partnerDt.Year(), int(partnerDt.Month()), partnerDt.Day())

		partnerDt = partnerDt.AddDate(0, 1, 0)
		s.InvoiceScheduleDates = append(s.InvoiceScheduleDates, partnerDt)
		futureDateNxtMonth3 := fmt.Sprintf("%v/%02d/%02d", partnerDt.Year(), int(partnerDt.Month()), partnerDt.Day())

		partnerDt = partnerDt.AddDate(0, 1, 0)
		s.InvoiceScheduleDates = append(s.InvoiceScheduleDates, partnerDt)
		futureDateNxtMonth4 := fmt.Sprintf("%v/%02d/%02d", partnerDt.Year(), int(partnerDt.Month()), partnerDt.Day())

		// Add one day
		invoiceDate = invoiceDate.AddDate(0, 0, 1)
		s.InvoiceScheduleDates = append(s.InvoiceScheduleDates, invoiceDate)

		futureDateNxtDay := fmt.Sprintf("%v/%02d/%02d", invoiceDate.Year(), int(invoiceDate.Month()), invoiceDate.Day())

		payload = []byte(fmt.Sprintf(csv, futureDateNxtMonth1, futureDateNxtMonth2, futureDateNxtMonth3, futureDateNxtMonth4, futureDateNxtDay))
	case "single-valid-date":
		csv := `invoice_schedule_id,invoice_date,is_archived,remarks
		,%v,,multiple-valid-dates-remarks1`

		payload = []byte(fmt.Sprintf(csv, fmt.Sprintf("%v/%02d/%02d", invoiceDate.Year(), int(invoiceDate.Month()), invoiceDate.Day())))

	case "invalid-col-count":
		csv := `invoice_schedule_id,invoice_date,is_archived,`
		payload = []byte(csv)
	case "invalid-header":
		csv := `invoice_date`
		payload = []byte(csv)
	case "multiple-valid-and-invalid-dates":
		dt := invoiceDate.AddDate(0, -1, 0)
		pastDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
		dt = invoiceDate
		presentDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
		dt = invoiceDate.AddDate(0, 5, 0)
		futureDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())

		csv := `invoice_schedule_id,invoice_date,is_archived,remarks
		,%v,,multiple-valid-and-invalid-dates-remarks1
		,%v,,multiple-valid-and-invalid-dates-remarks2
		,%v,,multiple-valid-and-invalid-dates-remarks3
		,2022-03-01,,
		1,%v,,
		,%v,1,
		1,%v,yes,
		,,,multiple-valid-and-invalid-dates-remarks4
		123,,1,multiple-valid-and-invalid-dates-remarks5`

		payload = []byte(fmt.Sprintf(csv, pastDate, presentDate, futureDate, futureDate, futureDate, futureDate))

		// remove the present date in the csv lines
		if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableKECFeedbackPh1) {
			csv := `invoice_schedule_id,invoice_date,is_archived,remarks
			,%v,,multiple-valid-and-invalid-dates-remarks1
			,%v,,multiple-valid-and-invalid-dates-remarks2
			,2022-03-01,,
			1,%v,,
			,%v,1,
			1,%v,yes,
			,,,multiple-valid-and-invalid-dates-remarks4
			123,,1,multiple-valid-and-invalid-dates-remarks5`

			payload = []byte(fmt.Sprintf(csv, pastDate, futureDate, futureDate, futureDate, futureDate))
		}

	case "one-invoice-to-archive":

		if err := try.Do(func(attempt int) (bool, error) {

			// Just get one existing scheduled invoice to archive
			invoiceSchedules, err := s.retrieveImportSchedule(ctx)

			if len(invoiceSchedules) > 0 {
				stepState.InvoiceScheduleID = invoiceSchedules[0].InvoiceScheduleID.String
				return true, nil
			}

			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return false, err
			}

			time.Sleep(1 * time.Second)
			return attempt < 10, fmt.Errorf("unable to find an existing invoice schedule with SCHEDULED status: %v", err)
		}); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}

		if stepState.InvoiceScheduleID == "" {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to find an existing invoice schedule with SCHEDULED status")
		}

		csv := `invoice_schedule_id,invoice_date,is_archived,remarks
		%v,,1,one-invoice-to-archive-1`

		payload = []byte(fmt.Sprintf(csv, stepState.InvoiceScheduleID))
	case "duplicate-valid-dates":
		csv := `invoice_schedule_id,invoice_date,is_archived,remarks
				,%v,,duplicate-valid-dates-remarks1
				,%v,,duplicate-valid-dates-remarks2`

		// Please do not use this date with added 1 year in other invoice schedule scenario. It may cause flaky test
		partnerDt := invoiceDate.AddDate(1, 0, 0)
		futureDate1 := fmt.Sprintf("%v/%02d/%02d", partnerDt.Year(), int(partnerDt.Month()), partnerDt.Day())
		futureDate2 := fmt.Sprintf("%v/%02d/%02d", partnerDt.Year(), int(partnerDt.Month()), partnerDt.Day())

		payload = []byte(fmt.Sprintf(csv, futureDate1, futureDate2))
	}

	return StepStateToContext(ctx, stepState), payload, nil
}

func (s *suite) importedInvoiceSchedulesAreConvertedIn(ctx context.Context, givenTZ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Retrieve the actual invoice schedules stored in DB
	invoiceSchedules, err := s.retrieveImportSchedule(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var country string
	switch givenTZ {
	case "VNT":
		country = COUNTRY_VN
	case "JST":
		country = COUNTRY_JP
	}

	var matchedInvoiceScheduleDate bool

	for _, invoiceSchedule := range invoiceSchedules {
		if invoiceSchedule.Status.String == invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String() {

			// Verifies if the imported date has the same value from the DB
			for _, invScheduleImport := range stepState.InvoiceScheduleDates {
				convertedDate, err := convertDatetoCountryTZ(invoiceSchedule.InvoiceDate.Time, country)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("error converting %v to the timezone of %v: %v", invoiceSchedule.InvoiceDate.Time, country, err)
				}

				if invScheduleImport.Equal(convertedDate) {
					matchedInvoiceScheduleDate = true
					break
				}
			}
		}
	}

	if !matchedInvoiceScheduleDate {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("invoice schedule date from DB does not match any imported dates (%v) for invoice schedule ID %v",
				stepState.InvoiceScheduleDates, stepState.InvoiceScheduleID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) scheduledDateIsOneDayAheadOfInvoiceDate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Retrieve the actual invoice schedules stored in DB
	invoiceSchedules, err := s.retrieveImportSchedule(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, invoiceSchedule := range invoiceSchedules {
		if invoiceSchedule.ScheduledDate.Status == pgtype.Null || invoiceSchedule.InvoiceDate.Status == pgtype.Null {
			return StepStateToContext(ctx, stepState), errors.New("dates cannot be null")
		}

		diff := invoiceSchedule.ScheduledDate.Time.Sub(invoiceSchedule.InvoiceDate.Time)
		if diff.Hours() != 24 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the difference between scheduled and invoice date to be 24 hours got %v", diff.Hours())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
