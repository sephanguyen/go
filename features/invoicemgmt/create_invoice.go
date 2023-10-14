package invoicemgmt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/try"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) thereIsAStudentThatHasBillItemWithStatus(ctx context.Context, status string) (context.Context, error) {
	return s.createBillItemBasedOnStatusAndType(ctx, status, payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String())
}

func (s *suite) billItemExistsInInvoicemgmtDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stmt := `SELECT bill_item_sequence_number FROM bill_item WHERE student_id = $1 AND resource_path = $2`

	if err := try.Do(func(attempt int) (bool, error) {
		var billItemSequenceNumber int32
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.StudentID, stepState.ResourcePath)
		err := row.Scan(&billItemSequenceNumber)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if billItemSequenceNumber != 0 {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("bill item sequence number is not found in invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreStudentsThatHasBillItemWithStatusAndType(ctx context.Context, number int, billItemCount int, status, billItemType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	type studentBillItem struct {
		StudentID          string
		BillItemSeqNumbers []int32
		Err                error
		InvoiceTotal       int64
	}

	studentBillItemChan := make(chan studentBillItem, number)

	var wg sync.WaitGroup

	for i := 0; i < number; i++ {
		wg.Add(1)
		stepState.InvoiceTotal = 0

		var totalAmount int64
		go func() {
			// Create new instance of step state
			newStepState := &common.StepState{}
			newStepState.ResourcePath = stepState.ResourcePath
			newStepState.LocationID = stepState.LocationID
			newStepState.InvoiceTotal = 0
			newCtx, cancel := context.WithCancel(ctx)

			// Inject the new step state to new context to prevent data race when used concurrently
			newCtx = StepStateToContext(newCtx, newStepState)

			defer cancel()
			defer wg.Done()

			var err error
			billItems := []int32{}
			for j := 0; j < billItemCount; j++ {
				_, err = s.createBillItemBasedOnStatusAndType(newCtx, status, billItemType)
				if err != nil {
					break
				}

				billItems = append(billItems, newStepState.BillItemSequenceNumber)

				if status == "BILLED" || status == "PENDING" {
					totalAmount += newStepState.InvoiceTotal
				}
			}

			studentBillItemChan <- studentBillItem{
				StudentID:          newStepState.StudentID,
				BillItemSeqNumbers: billItems,
				Err:                err,
				InvoiceTotal:       totalAmount,
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
		stepState.StudentBillItemMap[item.StudentID] = item.BillItemSeqNumbers
		stepState.StudentIds = append(stepState.StudentIds, item.StudentID)
		stepState.StudentInvoiceTotalMap[item.StudentID] = item.InvoiceTotal
	}

	// add delay to wait for the syncing of bill item
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateInvoiceEndpointIsCalledToCreateMultipleInvoice(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoiceRequest := &invoice_pb.GenerateInvoicesRequest{}
	for studentID, billItemIDs := range stepState.StudentBillItemMap {
		invoiceRequest.Invoices = append(invoiceRequest.Invoices, &invoice_pb.GenerateInvoiceDetail{
			InvoiceType: invoice_pb.InvoiceType_MANUAL,
			BillItemIds: billItemIDs,
			StudentId:   studentID,
			SubTotal:    float32(stepState.StudentInvoiceTotalMap[studentID]),
			Total:       int32(stepState.StudentInvoiceTotalMap[studentID]),
		})
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).GenerateInvoices(contextWithToken(ctx), invoiceRequest)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreNoErrorsInResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp, ok := stepState.Response.(*invoice_pb.GenerateInvoicesResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response is nil")
	}

	if len(resp.Errors) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are errors in response")
	}

	if !resp.Successful {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting successful response to be true got false")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreResponseError(ctx context.Context, count int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp, ok := stepState.Response.(*invoice_pb.GenerateInvoicesResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response is nil")
	}

	if len(resp.Errors) != count {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting response error count to be %d got %d", count, len(resp.Errors))
	}

	if resp.Successful {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting successful response to be false got true")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereArestudentDraftInvoicesCreatedSuccessfully(ctx context.Context, count int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := "SELECT invoice_id, student_id FROM invoice WHERE student_id = ANY($1) AND status = $2"

	var students pgtype.TextArray
	_ = students.Set(stepState.StudentIds)

	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, stmt, students, invoice_pb.InvoiceStatus_DRAFT.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			invoiceID string
			studentID string
		)

		err := rows.Scan(&invoiceID, &studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("row.Scan: %w", err)
		}

		stepState.InvoiceStudentMap[invoiceID] = studentID
		stepState.InvoiceIDs = append(stepState.InvoiceIDs, invoiceID)
	}

	if len(stepState.InvoiceStudentMap) != count {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number of student with invoice to be %d got %d", count, len(stepState.InvoiceStudentMap))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceBillItemIsCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoiceIDs := []string{}
	for invoiceID := range stepState.InvoiceStudentMap {
		invoiceIDs = append(invoiceIDs, invoiceID)
	}

	stmt := "SELECT invoice_id, bill_item_sequence_number FROM invoice_bill_item WHERE invoice_id = ANY($1)"

	var i pgtype.TextArray
	_ = i.Set(invoiceIDs)

	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, stmt, i)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			invoiceID              string
			billItemSequenceNumber int32
		)

		err := rows.Scan(&invoiceID, &billItemSequenceNumber)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("row.Scan: %w", err)
		}

		studentID, ok := stepState.InvoiceStudentMap[invoiceID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error in InvoiceStudentMap")
		}

		expectedStudentBillItems := stepState.StudentBillItemMap[studentID]
		if !sliceutils.ContainFunc(expectedStudentBillItems, billItemSequenceNumber, func(x, y int32) bool { return x == y }) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("the created invoice bill item %d does not exist in student %s", billItemSequenceNumber, studentID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceDataIsPresentInTheResponseWithCount(ctx context.Context, count int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*invoice_pb.GenerateInvoicesResponse).InvoicesData

	if len(resp) != count {
		return StepStateToContext(ctx, stepState), fmt.Errorf("count of invoices data must be equal to count of students")
	}

	var ids []string

	if len(resp) > 0 {
		for _, invoice := range resp {
			ids = append(ids, invoice.InvoiceId)
			if invoice.InvoiceId == "" {
				return StepStateToContext(ctx, stepState), fmt.Errorf("invoice ID must not be empty")
			}
		}
	}

	if checkDuplicateCount(ids) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invoice ID must be unique")
	}

	return StepStateToContext(ctx, stepState), nil
}

func checkDuplicateCount(list []string) int {
	frequency := 0
	duplicate_frequency := make(map[string]int)
	for _, item := range list {
		_, exist := duplicate_frequency[item]
		if exist {
			duplicate_frequency[item] += 1
			frequency++
		} else {
			duplicate_frequency[item] = 1
		}
	}

	return frequency
}

func (s *suite) billItemHasReviewRequiredTag(ctx context.Context, billItemReviewRequired string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for studentID, billItemIDs := range stepState.StudentBillItemMap {
		if len(billItemIDs) > 0 {
			switch billItemReviewRequired {
			case "one":
				if err := s.updateBillItemIsReviewed(ctx, studentID, []int32{billItemIDs[0]}, false); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("err updating bill item is_reviewed: %v", err)
				}

				// Others with is_reviewed=true
				if len(billItemIDs) > 1 {
					if err := s.updateBillItemIsReviewed(ctx, studentID, billItemIDs[1:], true); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("err updating bill item is_reviewed: %v", err)
					}
				}
			case "no":
				if err := s.updateBillItemIsReviewed(ctx, studentID, billItemIDs, true); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("err updating bill item is_reviewed: %v", err)
				}
			case "all":
				if err := s.updateBillItemIsReviewed(ctx, studentID, billItemIDs, false); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("err updating bill item is_reviewed: %v", err)
				}
			}
		}
	}

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnErrorAndNoInvoiceInTheResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, ok := stepState.Response.(*invoice_pb.GenerateInvoicesResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response is nil")
	}

	if len(resp.Errors) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting 1 error from response but got %v", len(resp.Errors))
	}

	if len(resp.InvoicesData) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting 0 invoice but got %v", len(resp.InvoicesData))
	}

	return StepStateToContext(ctx, stepState), nil
}
