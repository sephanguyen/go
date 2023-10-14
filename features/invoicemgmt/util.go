package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	ec "github.com/manabie-com/backend/features/invoicemgmt/entities_creator"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoicemgmt_entities "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoicemgmt_repo "github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// InsertEntities run all the insert function in sequence.
// Make sure that the stepState that was passed is the stepState from context
func InsertEntities(stepState *common.StepState, insertFunc ...ec.InsertEntityFunction) error {
	for _, f := range insertFunc {
		err := f(stepState)
		if err != nil {
			return err
		}
	}

	return nil
}

// createStudentWithBillItem creates a student with Bill Item.
// If a student already exist in step state, it will skip the creation of student, the same with the location.
func (s *suite) createStudentWithBillItem(ctx context.Context, status, billItemType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error

	// Create a student if there is no existing student ID in step state
	if stepState.StudentID == "" {
		ctx, err = s.createStudent(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// wait for kafka sync of bob entities that are needed for inserting fatima entities
		time.Sleep(invoiceConst.KafkaSyncSleepDuration)
	}

	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateTax(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateBillingSchedule(ctx, s.FatimaDBTrace, true),
		s.EntitiesCreator.CreateBillingSchedulePeriod(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateProduct(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateOrder(ctx, s.FatimaDBTrace, payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(), false),
		s.EntitiesCreator.CreateStudentProduct(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateBillItem(ctx, s.FatimaDBTrace, status, billItemType),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createInvoiceOfBillItem creates invoice of bill item. It also creates the invoice_bill_item.
func (s *suite) createInvoiceOfBillItem(ctx context.Context, billStatus, invoiceStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	ctx, err = s.createStudentWithBillItem(ctx, billStatus, payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	time.Sleep(invoiceConst.KafkaSyncSleepDuration) // wait for kafka sync

	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateInvoice(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceStatus),
		s.EntitiesCreator.CreateInvoiceBillItem(ctx, s.InvoiceMgmtPostgresDBTrace, "BILLING_STATUS_BILLED"),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createInvoiceBillItem creates the invoice_bill_item.
func (s *suite) createInvoiceBillItem(ctx context.Context, pastBillingStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateInvoiceBillItem(ctx, s.InvoiceMgmtPostgresDBTrace, pastBillingStatus),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// updateInvoiceStatus updates the invoice status with a given status
func (s *suite) updateInvoiceStatus(ctx context.Context, newInvoiceStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, span := interceptors.StartSpan(ctx, "suite.updateInvoiceStatus")
	defer span.End()

	stmt := fmt.Sprintf("UPDATE invoice SET status = '%s' WHERE invoice_id = '%v'", newInvoiceStatus, stepState.InvoiceID)

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt); err != nil {
		return ctx, fmt.Errorf("error updating invoice status with invoice_id %v and status %v: %v", stepState.InvoiceID, newInvoiceStatus, err)
	}

	return StepStateToContext(ctx, stepState), nil
}

// getPaymentHistoryRecordCount returns the count of payment history record
func (s *suite) getPaymentHistoryRecordCount(ctx context.Context, invoiceID string, status string) (int32, error) {
	var rowCount int32

	stmt := fmt.Sprintf("SELECT COUNT(*) FROM payment WHERE invoice_id = '%v' AND payment_status = '%v'", invoiceID, status)

	row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt)

	if err := row.Scan(&rowCount); err != nil {
		return rowCount, fmt.Errorf("error finding payment with invoice_id '%v' and status %v: %w", invoiceID, status, err)
	}

	return rowCount, nil
}

// createStudentParentRelationship creates the a binding for student and parent. Basically it inserts the data in student_parents table.
func (s *suite) createStudentParentRelationship(ctx context.Context, relationship string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateStudentParent(ctx, s.BobDBTrace, relationship),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createPayment creates or insert payment
func (s *suite) createPayment(ctx context.Context, paymentMethod invoice_pb.PaymentMethod, paymentStatus, existingResultCode string, isExported bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreatePayment(ctx, s.InvoiceMgmtPostgresDBTrace, paymentMethod.String(), paymentStatus, existingResultCode, isExported),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

// createInvoiceBasedOnStatus creates the invoice of bill item based on a given status
func (s *suite) createInvoiceBasedOnStatus(ctx context.Context, invoice string) error {
	var (
		err           error
		paymentStatus invoice_pb.PaymentStatus
		invoiceStatus invoice_pb.InvoiceStatus
	)

	switch invoice {
	case "ISSUED":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_PENDING
		invoiceStatus = invoice_pb.InvoiceStatus_ISSUED

	case "VOID":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
		invoiceStatus = invoice_pb.InvoiceStatus_VOID

	case "PAID":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL
		invoiceStatus = invoice_pb.InvoiceStatus_PAID

	case "FAILED":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
		invoiceStatus = invoice_pb.InvoiceStatus_FAILED

	case "REFUNDED":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_PENDING
		invoiceStatus = invoice_pb.InvoiceStatus_REFUNDED
	default:
		invoiceStatus = invoice_pb.InvoiceStatus_DRAFT
	}

	ctx, err = s.createInvoiceOfBillItem(ctx, payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoiceStatus.String())
	if err != nil {
		return err
	}

	if invoice != "DRAFT" {
		_, err = s.createPayment(ctx, invoice_pb.PaymentMethod_DIRECT_DEBIT, paymentStatus.String(), "", false)
	}

	if err != nil {
		return err
	}

	return nil
}

// createBillItemBasedOnStatus creates a bill item of a student based on status
func (s *suite) createBillItemBasedOnStatusAndType(ctx context.Context, status, billItemType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	splitBillItemStatuses := strings.Split(status, "-")
	for _, splitBillItemStatus := range splitBillItemStatuses {
		var billItemStatus string
		switch splitBillItemStatus {
		case "BILLED":
			billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()
		case "PENDING":
			billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_PENDING.String()
		case "INVOICED":
			billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()
		case "CANCELLED":
			billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_CANCELLED.String()
		case "WAITING_APPROVAL":
			billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_WAITING_APPROVAL.String()
		}
		ctx, err := s.createStudentWithBillItem(ctx, billItemStatus, billItemType)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error cannot create bill items %v", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

// deleteBillItemsByInvoiceID delete the binding of bill item and an invoice by invoice ID
func (s *suite) deleteBillItemsByInvoiceID(ctx context.Context, invoiceID string) error {
	stmt := `DELETE FROM invoice_bill_item where invoice_id = $1`

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, invoiceID); err != nil {
		return fmt.Errorf("error deleting invoice bill items with invoice_id %v: %v", invoiceID, err)
	}

	return nil
}

// generateRandomNumber generates a random number
func (s *suite) generateRandomNumber() int32 {
	return int32(time.Now().UTC().UnixNano())
}

// getLatestPaymentHistoryRecordByInvoiceID get the latest payment history record by invoice ID
func (s *suite) getLatestPaymentHistoryRecordByInvoiceID(ctx context.Context, invoiceID string) (*invoicemgmt_entities.Payment, error) {
	paymentRepo := invoicemgmt_repo.PaymentRepo{}
	return paymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)
}

// getLatestInvoiceActionLogByInvoiceID returns the latest invoice action log
func (s *suite) getLatestInvoiceActionLogByInvoiceID(ctx context.Context, invoiceID string) (*invoicemgmt_entities.InvoiceActionLog, error) {
	invoiceActionLogRepo := invoicemgmt_repo.InvoiceActionLogRepo{}
	return invoiceActionLogRepo.GetLatestRecordByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)
}

// getInvoiceByInvoiceID returns invoice by invoice ID
func (s *suite) getInvoiceByInvoiceID(ctx context.Context, invoiceID string) (*invoicemgmt_entities.Invoice, error) {
	invoiceRepo := invoicemgmt_repo.InvoiceRepo{}
	return invoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)
}

// getBillItemByInvoiceID returns a bill item of an invoice
func (s *suite) getBillItemByInvoiceID(ctx context.Context, invoiceID string) (*invoicemgmt_entities.BillItem, error) {

	invoiceBillItemRepo := invoicemgmt_repo.InvoiceBillItemRepo{}
	billItemRepo := invoicemgmt_repo.BillItemRepo{}

	invoiceBillItems, err := invoiceBillItemRepo.FindAllByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)

	if err != nil {
		return nil, err
	}

	if invoiceBillItems == nil || len(invoiceBillItems.ToArray()) == 0 {
		return nil, fmt.Errorf("NO RECORDS FOUND")
	}

	// Get the first item only
	invoiceBillItem := invoiceBillItems.ToArray()[0]

	return billItemRepo.FindByID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceBillItem.BillItemSequenceNumber.Int)
}

// insertInvoiceIntoInvoicemgmt creates only the invoice record.
// Unlike with other methods of creating invoice, this method only inserts an invoice.
func (s *suite) insertInvoiceIntoInvoicemgmt(ctx context.Context, invoiceStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateInvoice(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceStatus),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateBillItemFinalPriceValueByStudentID(ctx context.Context, studentID string, finalPrice float64) error {
	stmt := `UPDATE bill_item SET final_price = $1 WHERE student_id = $2 AND resource_path = $3`
	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, finalPrice, studentID, s.StepState.ResourcePath); err != nil {
		return fmt.Errorf("error updateBillItemFinalPriceValueByStudentID: %v", err)
	}

	err := s.updateInvoiceTotal(ctx, finalPrice)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) updateBillItemAdjustmentPriceValueByStudentID(ctx context.Context, studentID string, adjustmentPrice float64) error {
	stmt := `UPDATE bill_item SET adjustment_price = $1, bill_type = $2 WHERE student_id = $3 AND resource_path = $4`

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, adjustmentPrice, payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String(), studentID, s.StepState.ResourcePath); err != nil {
		return fmt.Errorf("error updateBillItemAdjustmentPriceValueByStudentID: %v", err)
	}

	err := s.updateInvoiceTotal(ctx, adjustmentPrice)
	if err != nil {
		return err
	}

	return nil
}

// updateInvoiceStatus updates the invoice status with a given status
func (s *suite) updateInvoiceType(ctx context.Context, newInvoiceType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := fmt.Sprintf("UPDATE invoice SET type = '%s' WHERE invoice_id = '%v'", newInvoiceType, stepState.InvoiceID)

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt); err != nil {
		return ctx, fmt.Errorf("error updating invoice type with invoice_id %v and type %v: %v", stepState.InvoiceID, newInvoiceType, err)
	}

	return StepStateToContext(ctx, stepState), nil
}

// updateInvoiceStatus updates the invoice status with a given status
func (s *suite) updateInvoiceTotal(ctx context.Context, invoiceTotal float64) error {
	stepState := StepStateFromContext(ctx)

	stmt := "UPDATE invoice SET total = $1, sub_total = $1 WHERE invoice_id = $2"
	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, invoiceTotal, stepState.InvoiceID); err != nil {
		return fmt.Errorf("error updating invoice total with invoice_id %v: err %v", stepState.InvoiceID, err)
	}

	return nil
}

// createGrantedRoleAccessPathLocationLevel  creates a granted role access path record
func (s *suite) createGrantedRoleAccessPath(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateGrantedRoleAccessPath(ctx, s.BobDBTrace, role),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createUser  creates a granted role access path record
func (s *suite) createUserAccessPathForStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateUserAccessPathForStudent(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

const (
	COUNTRY_JP = "COUNTRY_JP"
	COUNTRY_VN = "COUNTRY_VN"
)

var countryTZMap = map[string]string{
	COUNTRY_JP: "Asia/Tokyo",
	COUNTRY_VN: "Asia/Ho_Chi_Minh",
}

func convertDatetoCountryTZ(date time.Time, country string) (time.Time, error) {
	timezone, ok := countryTZMap[country]
	if !ok {
		timezone = countryTZMap[COUNTRY_VN]
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return date, err
	}

	return date.In(location), nil
}

func (s *suite) fromOrgInCountryLoginsToBackofficeApp(ctx context.Context, user string, org string, country string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Update country values to blank if not specified
	if country == "NO_COUNTRY" {
		country = ""
	}

	ctx, err := s.updateOrgCountry(ctx, org, country)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx = s.setResourcePathAndClaims(ctx, org)

	ctx, err = s.signedAsAccount(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.OrganizationCountry = country

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrgCountry(ctx context.Context, org string, country string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, span := interceptors.StartSpan(ctx, "suite.updateOrgCountry")
	defer span.End()

	stmt := fmt.Sprintf("UPDATE organizations SET country = '%s' WHERE organization_id = '%v'", country, org)

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt); err != nil {
		return ctx, fmt.Errorf("error updating organization with organization_id %v and country %v: %v", org, country, err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func resetTimeComponent(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func (s *suite) createStudentPaymentDetail(ctx context.Context, paymentMethod, studentID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateStudentPaymentDetail(ctx, s.InvoiceMgmtPostgresDBTrace, paymentMethod, studentID),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func getFormattedTimestampDate(dateString string) *timestamppb.Timestamp {
	var dateTimestamp *timestamppb.Timestamp
	switch dateString {
	case "TODAY":
		dateTimestamp = timestamppb.New(time.Now().Add(1 * time.Hour))
	case "TODAY+1":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, 1))
	case "TODAY+2":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, 2))
	case "TODAY+3":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, 3))
	}

	return dateTimestamp
}

func (s *suite) createStudentBankAccount(ctx context.Context, studentIDs []string, bankBranchID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	resChan := make(chan error, len(stepState.StudentIds))
	var wg sync.WaitGroup

	for _, id := range studentIDs {
		wg.Add(1)
		go func(studentID string) {

			// Create new instance of step state
			newStepState := &common.StepState{}
			newStepState.ResourcePath = stepState.ResourcePath
			newStepState.LocationID = stepState.LocationID
			newStepState.StudentID = studentID
			newStepState.BankBranchID = bankBranchID

			newCtx, cancel := context.WithCancel(ctx)

			// Inject the new step state to new context to prevent data race when used concurrently
			newCtx = StepStateToContext(newCtx, newStepState)

			defer cancel()
			defer wg.Done()

			err := InsertEntities(
				newStepState,
				s.EntitiesCreator.CreateStudentPaymentDetail(newCtx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.PaymentMethod_DIRECT_DEBIT.String(), studentID),
				s.EntitiesCreator.CreateBankAccount(newCtx, s.InvoiceMgmtPostgresDBTrace, true),
			)

			resChan <- err
		}(id)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for err := range resChan {
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

// updatePaymentStatus updates the update status with a given status
func (s *suite) updatePaymentStatusByPaymentSeqNumber(ctx context.Context, paymentSeqNumber int, newPaymentStatus string) error {
	stepState := StepStateFromContext(ctx)

	ctx, span := interceptors.StartSpan(ctx, "suite.updatePaymentStatusByPaymentSeqNumber")
	defer span.End()

	sqlQuery := "UPDATE payment SET payment_status = '%s' WHERE payment_sequence_number = '%v' AND resource_path = '%v'"
	stmt := fmt.Sprintf(sqlQuery, newPaymentStatus, paymentSeqNumber, stepState.ResourcePath)

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("error updating payment status with payment_sequence_number %v and status %v: %v", paymentSeqNumber, newPaymentStatus, err)
	}

	return nil
}

func (s *suite) updateBillItemIsReviewed(ctx context.Context, studentID string, billItemSequenceNumbers []int32, isReviewed bool) error {
	stmt := `UPDATE bill_item SET is_reviewed = $1 WHERE student_id = $2 AND bill_item_sequence_number = ANY($3)`

	var billItemSequenceNumbersArray pgtype.Int4Array
	_ = billItemSequenceNumbersArray.Set(billItemSequenceNumbers)

	if _, err := s.FatimaDBTrace.Exec(ctx, stmt, isReviewed, studentID, billItemSequenceNumbersArray); err != nil {
		return err
	}

	return nil
}

func (s *suite) updateBillItemCreatedAt(ctx context.Context, studentID string, billItemSequenceNumbers []int32, createdAt time.Time) error {
	stmt := `UPDATE bill_item SET created_at = $1 WHERE student_id = $2 AND bill_item_sequence_number = ANY($3)`

	var billItemSequenceNumbersArray pgtype.Int4Array
	_ = billItemSequenceNumbersArray.Set(billItemSequenceNumbers)

	if _, err := s.FatimaDBTrace.Exec(ctx, stmt, createdAt, studentID, billItemSequenceNumbersArray); err != nil {
		return err
	}

	return nil
}

func (s *suite) findBillItemsByBillItemSequenceNumbers(ctx context.Context, billItemSequenceNumbers []int32, org string) ([]*entities.BillItem, error) {
	e := &entities.BillItem{}
	fields, _ := e.FieldMap()

	var arr pgtype.Int4Array
	_ = arr.Set(billItemSequenceNumbers)

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE bill_item_sequence_number = ANY($1) AND resource_path = $2", strings.Join(fields, ","), e.TableName())

	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, stmt, arr, org)
	if err != nil {
		return nil, err
	}

	billItems := []*entities.BillItem{}
	defer rows.Close()
	for rows.Next() {
		billItem := new(entities.BillItem)
		database.AllNullEntity(billItem)

		_, fieldValues := billItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, billItem)
	}

	return billItems, nil
}

func (s *suite) getLatestActionLogByInvoiceIDAndAction(ctx context.Context, invoiceID, action string) (*invoicemgmt_entities.InvoiceActionLog, error) {
	e := &entities.InvoiceActionLog{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_id = $1 AND action = $2 ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, invoiceID, action).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func getPaymentMethodFromStr(paymentMethodStr string) (invoice_pb.PaymentMethod, error) {
	var paymentMethod invoice_pb.PaymentMethod
	switch paymentMethodStr {
	case "DIRECT_DEBIT":
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT
	case "CONVENIENCE_STORE":
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
	case "CASH":
		paymentMethod = invoice_pb.PaymentMethod_CASH
	case "BANK_TRANSFER":
		paymentMethod = invoice_pb.PaymentMethod_BANK_TRANSFER
	default:
		return paymentMethod, fmt.Errorf("payment method %v not supported", paymentMethodStr)
	}

	return paymentMethod, nil
}

func (s *suite) getLatestInvoicePaymentFromStepState(ctx context.Context) (*entities.Payment, error) {
	stepState := StepStateFromContext(ctx)

	currentPayment := stepState.CurrentPayment
	if currentPayment != nil {
		return currentPayment, nil
	}

	payment, err := s.getLatestPaymentHistoryRecordByInvoiceID(ctx, stepState.InvoiceID)
	if err != nil {
		return nil, err
	}
	stepState.CurrentPayment = payment

	return stepState.CurrentPayment, nil
}

func (s *suite) latestPaymentRecordHasPaymentStatus(ctx context.Context, paymentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	payment, err := s.getLatestInvoicePaymentFromStepState(ctx)

	// It's possible to have no payment records
	if err != nil && err != pgx.ErrNoRows {
		return StepStateToContext(ctx, stepState), fmt.Errorf("latestPaymentRecordHasPaymentStatus error: %v", err.Error())
	}

	// "none" expects no payment record
	if paymentStatus == "none" && payment != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected no payment record but got one for invoice_id %v", s.StepState.InvoiceID)
	}

	// if status is not "none", payment record is expected
	if paymentStatus != "none" && payment == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment record but got none for invoice_id %v", s.StepState.InvoiceID)
	}

	var expectedPaymentStatus invoice_pb.PaymentStatus

	switch paymentStatus {
	case "SUCCESSFUL":
		expectedPaymentStatus = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL
	case "FAILED":
		expectedPaymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
	default:
		expectedPaymentStatus = invoice_pb.PaymentStatus_PAYMENT_PENDING
	}

	if paymentStatus != "none" && payment.PaymentStatus.String != expectedPaymentStatus.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v payment status but got %v for invoice_id %v", expectedPaymentStatus.String(), payment.PaymentStatus.String, s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}

// this util for creating action log is for payment service
func (s *suite) actionLogRecordIsRecordedWithActionAndRemarks(ctx context.Context, action, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	actionLog, err := s.getLatestActionLogByInvoiceIDAndAction(ctx, stepState.InvoiceID, action)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// remarks is not included on query as it is optional and can be empty
	if actionLog.ActionComment.String != strings.TrimSpace(remarks) {
		return nil, fmt.Errorf("error action log remarks expected %v got actual %v", remarks, actionLog.ActionComment.String)
	}

	switch action {
	case invoice_pb.InvoiceAction_PAYMENT_CANCELLED.String():
		if actionLog.ActionDetail.String != invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
			return nil, fmt.Errorf("error action detail expected %v got actual %v", invoice_pb.PaymentStatus_PAYMENT_FAILED.String(), actionLog.ActionDetail.String)
		}
	case invoice_pb.InvoiceAction_PAYMENT_APPROVED.String():
		if actionLog.ActionDetail.String != invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() {
			return nil, fmt.Errorf("error action detail expected %v got actual %v", invoice_pb.PaymentStatus_PAYMENT_FAILED.String(), actionLog.ActionDetail.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreExistingInvoicesWithStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	type invoiceResult struct {
		InvoiceID string
		Err       error
		StudentID string
	}

	studentLength := 5
	if s.NoOfStudentsInvoiceToCreate != 0 {
		studentLength = s.NoOfStudentsInvoiceToCreate
	}

	invoiceChan := make(chan invoiceResult, studentLength)
	var wg sync.WaitGroup

	for i := 1; i <= studentLength; i++ {
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

			var err error
			switch status {
			case "DRAFT":
				_, err = s.createInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_DRAFT.String())
			case "FAILED":
				_, err = s.createInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_FAILED.String())
			case "ISSUED":
				_, err = s.createInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_ISSUED.String())
			case "PAID":
				_, err = s.createInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_PAID.String())
			case "VOID":
				_, err = s.createInvoiceOfBillItem(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_VOID.String())
			}

			invoiceChan <- invoiceResult{
				InvoiceID: newStepState.InvoiceID,
				Err:       err,
				StudentID: newStepState.StudentID,
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
	}

	if len(stepState.InvoiceIDs) == 1 {
		stepState.InvoiceID = stepState.InvoiceIDs[0]
	}

	if len(stepState.StudentIds) == 1 {
		stepState.StudentID = stepState.StudentIds[0]
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisInvoiceHasTotalAmount(ctx context.Context, invoiceTotal float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if strings.TrimSpace(stepState.InvoiceID) == "" {
		stepState.InvoiceID = stepState.InvoiceIDs[0]
	}

	e := new(entities.Invoice)
	database.AllNullEntity(e)
	stepState.InvoiceTotalAmount = []float64{invoiceTotal}

	err := multierr.Combine(
		e.InvoiceID.Set(stepState.InvoiceID),
		e.Total.Set(invoiceTotal),
		e.SubTotal.Set(invoiceTotal),
		e.UpdatedAt.Set(time.Now()),
		e.OutstandingBalance.Set(invoiceTotal),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}

	invoiceRepo := &invoicemgmt_repo.InvoiceRepo{}

	if err := invoiceRepo.UpdateWithFields(ctx, s.InvoiceMgmtPostgresDBTrace, e, []string{"total", "sub_total", "outstanding_balance", "updated_at"}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error InvoiceRepo UpdateWithFields: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreExistingPayments(ctx context.Context, count int, status string, method string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		paymentStatus invoice_pb.PaymentStatus
		invoiceStatus invoice_pb.InvoiceStatus
		billingStatus payment_pb.BillingStatus
		paymentMethod invoice_pb.PaymentMethod
	)

	switch status {
	case "PENDING":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_PENDING
		invoiceStatus = invoice_pb.InvoiceStatus_ISSUED
		billingStatus = payment_pb.BillingStatus_BILLING_STATUS_INVOICED
	case "SUCCESSFUL":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL
		invoiceStatus = invoice_pb.InvoiceStatus_PAID
		billingStatus = payment_pb.BillingStatus_BILLING_STATUS_INVOICED
	case "REFUNDED":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_REFUNDED
		invoiceStatus = invoice_pb.InvoiceStatus_REFUNDED
		billingStatus = payment_pb.BillingStatus_BILLING_STATUS_INVOICED
	case "FAILED":
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
		invoiceStatus = invoice_pb.InvoiceStatus_VOID
		billingStatus = payment_pb.BillingStatus_BILLING_STATUS_BILLED
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("payment status %s is not supported", status)
	}

	switch method {
	case "CONVENIENCE STORE", "CONVENIENCE_STORE":
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
	case "DIRECT DEBIT", "DIRECT_DEBIT":
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT
	case "CASH":
		paymentMethod = invoice_pb.PaymentMethod_CASH
	case "BANK_TRANSFER":
		paymentMethod = invoice_pb.PaymentMethod_BANK_TRANSFER
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("payment method %s is not supported", method)
	}

	type res struct {
		PaymentID string
		InvoiceID string
		StudentID string
		OrderID   string
		Err       error
	}

	resChan := make(chan res, count)

	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			// Create new instance of step state
			newStepState := &common.StepState{}
			newStepState.ResourcePath = stepState.ResourcePath
			newStepState.LocationID = stepState.LocationID

			newCtx, cancel := context.WithCancel(ctx)

			// Inject the new step state to new context to prevent data race when used concurrently
			newCtx = StepStateToContext(newCtx, newStepState)

			defer cancel()
			defer wg.Done()

			var err error
			_, err = s.createInvoiceOfBillItem(newCtx, billingStatus.String(), invoiceStatus.String())
			if err != nil {
				resChan <- res{Err: err}
				return
			}

			_, err = s.createPayment(newCtx, paymentMethod, paymentStatus.String(), "", false)
			if err != nil {
				resChan <- res{Err: err}
				return
			}

			resChan <- res{
				InvoiceID: newStepState.InvoiceID,
				PaymentID: newStepState.PaymentID,
				StudentID: newStepState.StudentID,
				OrderID:   newStepState.OrderID,
				Err:       nil,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for item := range resChan {
		if item.Err != nil {
			return StepStateToContext(ctx, stepState), item.Err
		}
		stepState.PaymentIDs = append(stepState.PaymentIDs, item.PaymentID)
		stepState.InvoiceIDs = append(stepState.InvoiceIDs, item.InvoiceID)
		stepState.StudentIds = append(stepState.StudentIds, item.StudentID)
		stepState.OrderIDs = append(stepState.OrderIDs, item.OrderID)
		stepState.PaymentStatusIDsMap[paymentStatus.String()] = append(stepState.PaymentStatusIDsMap[paymentStatus.String()], item.PaymentID)
	}

	stepState.PaymentMethod = paymentMethod.String()

	return StepStateToContext(ctx, stepState), nil
}

func getRefundMethodFromStr(refundMethodStr string) (invoice_pb.RefundMethod, error) {
	var refundMethod invoice_pb.RefundMethod
	switch refundMethodStr {
	case "CASH":
		refundMethod = invoice_pb.RefundMethod_REFUND_CASH
	case "BANK_TRANSFER":
		refundMethod = invoice_pb.RefundMethod_REFUND_BANK_TRANSFER
	default:
		return refundMethod, fmt.Errorf("refund method %v not supported", refundMethodStr)
	}

	return refundMethod, nil
}

func (s *suite) thereAreExistingInvoicesWithStatusAndAmount(ctx context.Context, status string, amount float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	type invoiceResult struct {
		InvoiceID string
		Err       error
		StudentID string
	}

	studentLength := 5
	if s.NoOfStudentsInvoiceToCreate != 0 {
		studentLength = s.NoOfStudentsInvoiceToCreate
	}

	invoiceChan := make(chan invoiceResult, studentLength)
	var wg sync.WaitGroup

	for i := 1; i <= studentLength; i++ {
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

			var err error
			switch status {
			case "DRAFT":
				err = s.createInvoiceOfBillItemV2(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_DRAFT.String(), amount)
			case "FAILED":
				err = s.createInvoiceOfBillItemV2(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_FAILED.String(), amount)
			case "ISSUED":
				err = s.createInvoiceOfBillItemV2(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_ISSUED.String(), amount)
			case "PAID":
				err = s.createInvoiceOfBillItemV2(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_PAID.String(), amount)
			case "VOID":
				err = s.createInvoiceOfBillItemV2(StepStateToContext(newCtx, newStepState), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), invoice_pb.InvoiceStatus_VOID.String(), amount)
			}

			invoiceChan <- invoiceResult{
				InvoiceID: newStepState.InvoiceID,
				Err:       err,
				StudentID: newStepState.StudentID,
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
	}

	if len(stepState.InvoiceIDs) == 1 {
		stepState.InvoiceID = stepState.InvoiceIDs[0]
	}

	if len(stepState.StudentIds) == 1 {
		stepState.StudentID = stepState.StudentIds[0]
	}

	stepState.InvoiceTotalFloat = amount

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createInvoiceOfBillItemV2(ctx context.Context, billStatus, invoiceStatus string, amount float64) error {
	var err error
	ctx, err = s.createStudentWithBillItemV2(ctx, billStatus, payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String(), amount)
	if err != nil {
		return err
	}

	time.Sleep(invoiceConst.KafkaSyncSleepDuration) // wait for kafka sync

	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateInvoiceV2(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceStatus, amount),
		s.EntitiesCreator.CreateInvoiceBillItem(ctx, s.InvoiceMgmtPostgresDBTrace, "BILLING_STATUS_BILLED"),
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *suite) createStudentWithBillItemV2(ctx context.Context, status, billItemType string, amount float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error

	// Create a student if there is no existing student ID in step state
	if stepState.StudentID == "" {
		ctx, err = s.createStudent(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// wait for kafka sync of bob entities that are needed for inserting fatima entities
		time.Sleep(invoiceConst.KafkaSyncSleepDuration)
	}

	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateTax(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateBillingSchedule(ctx, s.FatimaDBTrace, true),
		s.EntitiesCreator.CreateBillingSchedulePeriod(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateProduct(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateOrder(ctx, s.FatimaDBTrace, payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(), false),
		s.EntitiesCreator.CreateStudentProduct(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateBillItemV2(ctx, s.FatimaDBTrace, status, billItemType, amount),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
