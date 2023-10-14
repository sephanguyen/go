package invoicemgmt

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoiceMigrationService "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

func (s *suite) thereAreStudentsThatAreMigrated(ctx context.Context, studentCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create the students
	for i := 0; i < studentCount; i++ {
		ctx, err := s.createStudent(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.StudentIds = append(stepState.StudentIds, stepState.StudentID)
	}

	// Update the email of students to be aligned with migrated students' email
	for _, studentID := range stepState.StudentIds {
		randomReferenceID := fmt.Sprintf("student-reference-%s", idutil.ULIDNow())
		stmt := "UPDATE users SET email = $1 WHERE user_id = $2"

		if _, err := s.BobDBTrace.Exec(ctx, stmt, fmt.Sprintf("%s%s", randomReferenceID, invoiceMigrationService.MigrationEmail), studentID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error user email: %v", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsInvoiceCSVFileForTheseStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	bytesData, err := s.genCSVDataWithStatusAmountMap(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &invoice_pb.ImportDataMigrationRequest{
		Payload:    bytesData,
		EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminImportsInvoiceMigrationData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, ok := stepState.Request.(*invoice_pb.ImportDataMigrationRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting stepState.Request to be *invoice_pb.ImportDataMigrationRequest got %T", req)
	}

	var (
		resp *invoice_pb.ImportDataMigrationResponse
		err  error
	)

	retryCount := 20
	if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		resp, err = invoice_pb.NewDataMigrationServiceClient(s.InvoiceMgmtConn).ImportDataMigration(contextWithToken(ctx), req)
		if err != nil {
			return false, err
		}

		// collect all errors from response
		errList := []string{}
		for _, e := range resp.Errors {
			errList = append(errList, e.Error)
		}

		// check if there are duplicate error in the response error
		allErrs := strings.Join(errList, " ")
		if !strings.Contains(allErrs, "(SQLSTATE 23505)") {
			// return nil error since the error is in the response and not from gRPC
			// this will be checked in the next step
			return false, nil
		}

		time.Sleep(invoiceConst.DuplicateSleepDuration)
		return attempt < retryCount, fmt.Errorf("cannot import invoice, err %v", err)
	}, retryCount); err != nil {
		stepState.Response = nil
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	stepState.Response = resp
	stepState.ResponseErr = nil

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreInvoicesOfStudentsMigratedSuccessfully(ctx context.Context, studentCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoices, err := s.getInvoicesOfStudents(ctx, stepState.StudentIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, invoice := range invoices {
		stepState.InvoiceIDs = append(stepState.InvoiceIDs, invoice.InvoiceID.String)
	}

	if len(stepState.InvoiceIDs) != studentCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number of student with invoice to be %d got %d", studentCount, len(stepState.InvoiceStudentMap))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) migratedInvoicesHaveCorrectAmountBasedOnItsStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoices, err := s.getInvoicesOfStudents(ctx, stepState.StudentIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, invoice := range invoices {
		exactTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		exactAmountPaid, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.AmountPaid, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		exactAmountRefunded, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.AmountRefunded, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		switch invoice.Status.String {
		case invoice_pb.InvoiceStatus_ISSUED.String(), invoice_pb.InvoiceStatus_FAILED.String():
			if exactOutstandingBalance != exactTotal && exactAmountPaid != 0 && exactAmountRefunded != 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("amount is incorrect for invoice with status %v", invoice.Status.String)
			}
		case invoice_pb.InvoiceStatus_PAID.String():
			if exactOutstandingBalance != 0 && exactAmountPaid != exactTotal && exactAmountRefunded != 0 {
				return StepStateToContext(ctx, stepState), errors.New("amount is incorrect for invoice with status PAID")
			}
		case invoice_pb.InvoiceStatus_REFUNDED.String():
			if exactOutstandingBalance != 0 && exactAmountPaid != 0 && exactAmountRefunded != exactTotal {
				return StepStateToContext(ctx, stepState), errors.New("amount is incorrect for invoice with status REFUNDED")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) migratedInvoicesHaveSavedReferenceNumberAndMigratedAt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoices, err := s.getInvoicesOfStudents(ctx, stepState.StudentIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, invoice := range invoices {
		if invoice.InvoiceReferenceID.Status == pgtype.Null || strings.TrimSpace(invoice.InvoiceReferenceID.String) == "" {
			return StepStateToContext(ctx, stepState), errors.New("invoice reference1 is empty")
		}

		if invoice.MigratedAt.Status == pgtype.Null {
			return StepStateToContext(ctx, stepState), errors.New("invoice migrated_at is null")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getInvoicesOfStudents(ctx context.Context, studentIDs []string) ([]*entities.Invoice, error) {
	e := &entities.Invoice{}

	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM invoice WHERE student_id = ANY($1)", strings.Join(fields, ", "))

	var students pgtype.TextArray
	_ = students.Set(studentIDs)

	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, stmt, students)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []*entities.Invoice{}
	for rows.Next() {
		invoice := &entities.Invoice{}
		_, values := invoice.FieldMap()

		err := rows.Scan(values...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (s *suite) thereAreNoErrorLinesInImportInvoiceResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, ok := stepState.Response.(*invoice_pb.ImportDataMigrationResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the response to be *invoice_pb.ImportDataMigrationResponse got %T", resp)
	}

	if len(resp.Errors) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are %v error/s in the response", len(resp.Errors))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreErrorLinesInImportInvoiceResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, ok := stepState.Response.(*invoice_pb.ImportDataMigrationResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the response to be *invoice_pb.ImportDataMigrationResponse got %T", resp)
	}

	if len(resp.Errors) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("expecting to have an error got zero")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsInvoiceCSVFileForNonExistingStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	csvData := [][]string{
		getImportInvoiceHeader(),
	}

	for i := 0; i < 5; i++ {
		studentReference := fmt.Sprintf("Invalid Reference %v", i)
		csvData = append(csvData, []string{
			fmt.Sprintf("%d", i+1),
			"",
			studentReference,
			"MANUAL",
			"ISSUED",
			"10000",
			"10000",
			time.Now().Format("2006-01-02"),
			"",
			"TRUE",
			fmt.Sprintf("invoice-reference-1-%s", studentReference),
			fmt.Sprintf("invoice-reference-2-%s", studentReference),
		})
	}

	bytesData, err := generateCSVByte(csvData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &invoice_pb.ImportDataMigrationRequest{
		Payload:    bytesData,
		EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsInvoiceCSVFileForTheseStudentsWithInvalidAmount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// student invoice total is mismatch from bill item total
	for i := 0; i < len(stepState.StudentIds); i++ {
		randomNumber, err := genRandomNumber(int64(len(stepState.StudentIds)))
		if err != nil {
			return nil, err
		}
		stepState.StudentBillItemTotalPrice[stepState.StudentIds[i]] = float64(randomNumber)
	}
	bytesData, err := s.genCSVDataWithStatusAmountMap(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &invoice_pb.ImportDataMigrationRequest{
		Payload:    bytesData,
		EntityName: invoice_pb.DataMigrationEntityName_INVOICE_ENTITY,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) genCSVDataWithStatusAmountMap(ctx context.Context) ([]byte, error) {
	stepState := StepStateFromContext(ctx)

	csvData := [][]string{
		getImportInvoiceHeader(),
	}

	statuses := []string{invoice_pb.InvoiceStatus_ISSUED.String(), invoice_pb.InvoiceStatus_PAID.String(), invoice_pb.InvoiceStatus_FAILED.String()}

	for i, studentID := range stepState.StudentIds {
		randomNumber, err := genRandomNumber(int64(len(statuses)))
		if err != nil {
			return nil, err
		}

		var status string
		// this gets the invoice status of the student randomly
		// if total bill item price is negative set the status to refunded
		switch {
		case stepState.StudentBillItemTotalPrice[studentID] < 0:
			status = invoice_pb.InvoiceStatus_REFUNDED.String()
		default:
			status = statuses[randomNumber]
		}

		amount := fmt.Sprintf("%.2f", stepState.StudentBillItemTotalPrice[studentID])

		csvData = append(csvData, []string{
			fmt.Sprintf("%d", i+1),
			"",
			studentID,
			"MANUAL",
			status,
			amount,
			amount,
			time.Now().Format("2006-01-02"),
			"",
			"TRUE",
			stepState.StudentInvoiceReferenceMap[studentID],
			stepState.StudentInvoiceReference2Map[studentID],
		})
	}

	bytesData, err := generateCSVByte(csvData)
	if err != nil {
		return nil, err
	}

	return bytesData, nil
}

func getImportInvoiceHeader() []string {
	return []string{
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
}

func generateCSVByte(csvData [][]string) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	err := writer.WriteAll(csvData)
	if err != nil {
		return nil, err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func genRandomNumber(limit int64) (int64, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(limit))
	if err != nil {
		return 0, err
	}

	return num.Int64(), nil
}

func (s *suite) thereAreStudentsThatHaveBillItemsMigratedWithTotalPrice(ctx context.Context, studentCount, billItemCount int, totalPrices string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	totalPricesSplit := strings.Split(totalPrices, "&")

	type studentBillItem struct {
		StudentID         string
		Err               error
		InvoiceReference  string
		InvoiceReference2 string
		BillItemTotal     float64
	}

	studentBillItemChan := make(chan studentBillItem, studentCount)

	var wg sync.WaitGroup

	for i := 0; i < studentCount; i++ {
		wg.Add(1)

		stepState.BillItemTotalFloat = 0
		var billItemTotal float64

		billItemTotalPrice, err := strconv.ParseFloat(strings.TrimSpace(totalPricesSplit[i]), 64)
		if err != nil {
			return nil, fmt.Errorf("error on converting amount: %v to float", totalPricesSplit[i])
		}

		go func() {
			// Create new instance of step state
			newStepState := &common.StepState{}
			newStepState.ResourcePath = stepState.ResourcePath
			newStepState.LocationID = stepState.LocationID
			newStepState.StudentID = stepState.StudentID
			newCtx, cancel := context.WithCancel(ctx)

			// Inject the new step state to new context to prevent data race when used concurrently
			newCtx = StepStateToContext(newCtx, newStepState)

			defer cancel()
			defer wg.Done()

			var err error

			invoiceReference := idutil.ULIDNow()
			invoiceReference2 := fmt.Sprintf("reference-2-%v", invoiceReference)

			finalPrice := billItemTotalPrice / float64(billItemCount)

			for j := 0; j < billItemCount; j++ {
				_, err = s.createMigrationStudentWithBillItem(newCtx, finalPrice, invoiceReference2)
				if err != nil {
					break
				}

				billItemTotal += newStepState.BillItemTotalFloat
			}
			studentBillItemChan <- studentBillItem{
				StudentID:         newStepState.StudentID,
				InvoiceReference:  invoiceReference,
				InvoiceReference2: invoiceReference2,
				Err:               err,
				BillItemTotal:     billItemTotal,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(studentBillItemChan)
	}()

	for item := range studentBillItemChan {
		if item.Err != nil {
			return StepStateToContext(ctx, stepState), item.Err
		}
		stepState.StudentInvoiceReferenceMap[item.StudentID] = item.InvoiceReference
		stepState.StudentInvoiceReference2Map[item.StudentID] = item.InvoiceReference2
		stepState.StudentIds = append(stepState.StudentIds, item.StudentID)
		stepState.StudentBillItemTotalPrice[item.StudentID] = item.BillItemTotal
	}

	// add delay to wait for the syncing of bill item
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	return StepStateToContext(ctx, stepState), nil
}
