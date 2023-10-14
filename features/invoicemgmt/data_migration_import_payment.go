package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) thereAreMigratedInvoicesWithStatus(ctx context.Context, invoiceRecordCount int, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	type invoiceResult struct {
		InvoiceID          string
		Err                error
		StudentID          string
		InvoiceReferenceID string
		InvoiceTotal       float64
	}

	invoiceChan := make(chan invoiceResult, invoiceRecordCount)
	var wg sync.WaitGroup

	for i := 1; i <= invoiceRecordCount; i++ {
		wg.Add(1)
		go func() {
			// Create new instance of step state
			newStepState := &common.StepState{}
			newStepState.ResourcePath = stepState.ResourcePath
			newStepState.LocationID = stepState.LocationID

			newCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			defer wg.Done()

			// Inject the new step state to new context to prevent data race when used concurrently
			newCtx = StepStateToContext(newCtx, newStepState)
			invoiceReference := idutil.ULIDNow()
			var err error
			switch status {
			case "DRAFT":
				err = s.createMigratedInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), invoice_pb.InvoiceStatus_DRAFT.String(), invoiceReference)
			case "FAILED":
				err = s.createMigratedInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), invoice_pb.InvoiceStatus_FAILED.String(), invoiceReference)
			case "ISSUED":
				err = s.createMigratedInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), invoice_pb.InvoiceStatus_ISSUED.String(), invoiceReference)
			case "PAID":
				err = s.createMigratedInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), invoice_pb.InvoiceStatus_PAID.String(), invoiceReference)
			case "REFUNDED":
				err = s.createMigratedInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), invoice_pb.InvoiceStatus_REFUNDED.String(), invoiceReference)
			}

			invoiceChan <- invoiceResult{
				InvoiceID:          newStepState.InvoiceID,
				Err:                err,
				StudentID:          newStepState.StudentID,
				InvoiceReferenceID: newStepState.InvoiceReferenceID,
				InvoiceTotal:       newStepState.InvoiceTotalFloat,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(invoiceChan)
	}()

	for item := range invoiceChan {
		if item.Err != nil {
			return StepStateToContext(ctx, stepState), item.Err
		}
		stepState.InvoiceIDs = append(stepState.InvoiceIDs, item.InvoiceID)
		stepState.StudentIds = append(stepState.StudentIds, item.StudentID)
		stepState.InvoiceStudentMap[item.InvoiceID] = item.StudentID
		stepState.InvoiceIDInvoiceReferenceMap[item.InvoiceID] = item.InvoiceReferenceID
		stepState.InvoiceIDInvoiceTotalMap[item.InvoiceID] = item.InvoiceTotal
	}

	if len(stepState.InvoiceIDs) == 1 {
		stepState.InvoiceID = stepState.InvoiceIDs[0]
	}

	if len(stepState.StudentIds) == 1 {
		stepState.StudentID = stepState.StudentIds[0]
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisPaymentCsvFileHasPaymentDataWithPaymentStatus(ctx context.Context, paymentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var payload []byte
	var csv string
	// generate header csv
	headerTitles, err := getHeaderTitles(invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	headerStr := strings.Join(headerTitles, ",")

	var paymentDateStr string

	timeNowStr := time.Now().Format("2006-01-02")
	if paymentStatus == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() {
		paymentDateStr = timeNowStr
	}

	// added header on csv
	csv = fmt.Sprintln(headerStr)

	for i := 0; i < len(stepState.InvoiceIDs); i++ {
		if i == len(stepState.InvoiceID)-1 {
			csv += fmt.Sprintf(`0,,,%v,%v,%v,%v,%v,%v,,true,%v,,,%v`, stepState.PaymentMethod, paymentStatus, timeNowStr, timeNowStr, paymentDateStr, stepState.InvoiceStudentMap[stepState.InvoiceIDs[i]], timeNowStr, stepState.InvoiceIDInvoiceReferenceMap[stepState.InvoiceIDs[i]])
		} else {
			csv += fmt.Sprintf(`0,,,%v,%v,%v,%v,%v,%v,,true,%v,,,%v`, stepState.PaymentMethod, paymentStatus, timeNowStr, timeNowStr, paymentDateStr, stepState.InvoiceStudentMap[stepState.InvoiceIDs[i]], timeNowStr, stepState.InvoiceIDInvoiceReferenceMap[stepState.InvoiceIDs[i]])
			csv += "\n"
		}
	}

	payload = []byte(csv)

	stepState.Request = &invoice_pb.ImportDataMigrationRequest{
		Payload:    payload,
		EntityName: invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importsThePaymentCsvFile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if err := try.Do(func(attempt int) (bool, error) {
		stepState.Response, stepState.ResponseErr = invoice_pb.NewDataMigrationServiceClient(s.InvoiceMgmtConn).ImportDataMigration(contextWithToken(ctx), stepState.Request.(*invoice_pb.ImportDataMigrationRequest))
		response := stepState.Response.(*invoice_pb.ImportDataMigrationResponse)

		if response.Errors == nil {
			return true, nil
		}

		if response.Errors != nil && !strings.Contains(response.Errors[0].Error, "(SQLSTATE 23505)") {
			return false, stepState.ResponseErr
		}

		return attempt < 10, fmt.Errorf("cannot create payment data migration, err %v", stepState.ResponseErr)
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentCsvFileIsImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	responseErr := stepState.Response.(*invoice_pb.ImportDataMigrationResponse).Errors

	if len(responseErr) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("payment csv file is not imported successfully")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereArePaymentRecordsWithCorrectInvoiceCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < len(stepState.InvoiceIDs); i++ {
		payment, err := s.getLatestPaymentHistoryRecordByInvoiceID(ctx, stepState.InvoiceIDs[i])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		exactPaymentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(payment.Amount, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err payment amount set: %w", err)
		}

		// check payment amount should be same on invoice total
		if exactPaymentAmount != stepState.InvoiceIDInvoiceTotalMap[stepState.InvoiceIDs[i]] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment amount: %v but got: %v", stepState.InvoiceIDInvoiceTotalMap[stepState.InvoiceIDs[i]], exactPaymentAmount)
		}

		// check payment reference invoice map
		if payment.PaymentReferenceID.String != stepState.InvoiceIDInvoiceReferenceMap[stepState.InvoiceIDs[i]] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment reference id: %v but got: %v", stepState.InvoiceIDInvoiceReferenceMap[stepState.InvoiceIDs[i]], payment.PaymentReferenceID.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAPaymentCsvFileWithPaymentMethodForTheseInvoices(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.PaymentMethod = paymentMethod

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentCsvFileIsImportedUnsuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ImportDataMigrationResponse)

	if response.Errors == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected response has error lines but got empty")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) responseHasErrorOnPaymentStatusThatShouldBe(ctx context.Context, invalidPaymentStatus, validPaymentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.ImportDataMigrationResponse)

	if response.Errors != nil && !strings.Contains(response.Errors[0].Error, fmt.Sprintf("should have payment status %v but got: %v", validPaymentStatus, invalidPaymentStatus)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected response has error lines from payment status but got %v", response.Errors[0].Error)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentCsvFileContainsInvalidStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := 0; i < len(stepState.InvoiceIDs); i++ {
		stepState.InvoiceStudentMap[stepState.InvoiceIDs[i]] = fmt.Sprintf("test-%v", i)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) responseHasErrorForInvalidPaymentStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*invoice_pb.ImportDataMigrationResponse)

	if response.Errors != nil && !strings.Contains(response.Errors[0].Error, "mismatch on invoice student id") {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected response has error lines from payment student but got %v", response.Errors[0].Error)
	}

	return StepStateToContext(ctx, stepState), nil
}
