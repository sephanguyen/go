package paymentfileutils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type PaymentFileValidator interface {
	Validate(ctx context.Context, paymentFile *PaymentFile) (*PaymentValidationResult, error)
}

type BasePaymentFileValidator struct {
	DB          database.Ext
	PaymentRepo interface {
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.Payment, fieldsToUpdate []string) error
		FindByPaymentSequenceNumber(ctx context.Context, db database.QueryExecer, paymentSequenceNumber int) (*entities.Payment, error)
		UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, payments []*entities.Payment, fields []string) error
		FindPaymentInvoiceUserFromTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.PaymentInvoiceUserMap, error)
		InsertPaymentNumbersTempTable(ctx context.Context, db database.QueryExecer, paymentSeqNumbers []int) error
	}
	InvoiceRepo interface {
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.Invoice, fieldsToUpdate []string) error
		RetrieveInvoiceByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Invoice, error)
		UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, invoices []*entities.Invoice, fields []string) error
	}
	BulkPaymentValidationsRepo interface {
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidations, fieldsToUpdate []string) error
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidations) (string, error)
	}
	BulkPaymentValidationsDetailRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidationsDetail) (string, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, validationDetails []*entities.BulkPaymentValidationsDetail) error
	}
	InvoiceActionLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceActionLog) error
		CreateMultiple(ctx context.Context, db database.QueryExecer, actionLogs []*entities.InvoiceActionLog) error
	}
	UserBasicInfoRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, userID string) (*entities.UserBasicInfo, error)
	}
	UnleashClient unleashclient.ClientInstance
	Env           string
}

type ResultCodeValidation struct {
	ResultCode       string
	InvoiceStatus    invoice_pb.InvoiceStatus
	PaymentStatus    invoice_pb.PaymentStatus
	SystemResultCode int
}

type GenericPaymentFile struct {
	PaymentMethod          int
	GenericPaymentData     []*GenericPaymentFileRecord
	TransferredTotalAmount float64
	TransferredNumber      int
	FailedTotalAmount      float64
	FailedNumber           int
}

type GenericPaymentFileRecord struct {
	PaymentDate   *time.Time
	ValidatedDate *time.Time
	Amount        int
	PaymentNumber string
	ResultCode    string
	CreatedDate   int
}

type ValidatedPayment struct {
	PaymentSequenceNumber int32
	ResultCode            string
	Amount                float64
	StudentID             string
	StudentName           string
	PaymentMethod         invoice_pb.PaymentMethod
	InvoiceSequenceNumber int32
	PaymentCreatedDate    time.Time
	InvoiceID             string
	PaymentStatus         string
}

type PaymentValidationResult struct {
	ValidatedPayments  []*ValidatedPayment
	ValidationDate     *time.Time
	SuccessfulPayments int32
	PendingPayments    int32
	FailedPayments     int32
}

// Performs the following for both Direct Debit (Text) and Convenience Store (CSV) files
// - Creation of records (e.g., bulk payment validation, action log, etc.)
// - Modification of invoice and payment records
// - Validation of the payment file with associated result code
func (t *BasePaymentFileValidator) GenericValidate(ctx context.Context, file *GenericPaymentFile) (*PaymentValidationResult, error) {
	enableImproveBulkPaymentValidation, err := t.UnleashClient.IsFeatureEnabled(constant.EnableImproveBulkPaymentValidation, t.Env)
	if err != nil {
		return nil, fmt.Errorf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableImproveBulkPaymentValidation, err)
	}

	// Run the improved validate function
	if enableImproveBulkPaymentValidation {
		return t.ImprovedGenericValidate(ctx, file)
	}

	validatedPayments := make([]*ValidatedPayment, 0)

	var paymentMethod invoice_pb.PaymentMethod
	switch file.PaymentMethod {
	case ConvenienceStore:
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
	case DirectDebit:
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT
	}

	paymentNumberLineMap := identifyDuplicatePaymentNumbers(file.GenericPaymentData)

	bulkPaymentValidation := new(entities.BulkPaymentValidations)
	database.AllNullEntity(bulkPaymentValidation)

	useBulkAddValidatePh2, err := t.UnleashClient.IsFeatureEnabled(constant.EnableBulkAddValidatePh2, t.Env)
	if err != nil {
		return nil, fmt.Errorf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableBulkAddValidatePh2, err)
	}

	err = database.ExecInTx(ctx, t.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		var successfulPayments, failedPayments, pendingPayments int32

		bulkPaymentValidation.PaymentMethod = database.Text(paymentMethod.String())
		bulkPaymentValidation.SuccessfulPayments = database.Int4(0)
		bulkPaymentValidation.FailedPayments = database.Int4(0)
		bulkPaymentValidation.PendingPayments = database.Int4(0)
		bulkPaymentValidation.ValidationDate = database.Timestamptz(time.Now())

		// Create a record to be updated later; ID referenced by details record
		bulkPaymentValidationID, err := t.BulkPaymentValidationsRepo.Create(ctx, tx, bulkPaymentValidation)
		if err != nil {
			return fmt.Errorf("unable to create bulk payment validations: %v", err.Error())
		}

		receiptDate := time.Now() // all successful payments will have the same receipt date
		// For each line of file, create bulk payment validation details record
		for lineNo, dataRecord := range file.GenericPaymentData {
			if paymentMethod == invoice_pb.PaymentMethod_CONVENIENCE_STORE {
				// For duplicate payments, this ensures the line with the greatest created data value is selected
				// Otherwise, select the last line with same created date value
				var paymentLineNumber int
				switch {
				case paymentNumberLineMap[dataRecord.PaymentNumber].GreaterCreatedDateFound:
					paymentLineNumber = paymentNumberLineMap[dataRecord.PaymentNumber].LineNoWithGreaterCreatedDate
				case paymentNumberLineMap[dataRecord.PaymentNumber].DuplicateCreatedDateFound:
					paymentLineNumber = paymentNumberLineMap[dataRecord.PaymentNumber].LastLineNoWithSameCreatedDate
				default:
					paymentLineNumber = paymentNumberLineMap[dataRecord.PaymentNumber].LineNo
				}

				if lineNo != paymentLineNumber {
					continue
				}
			}

			payment, invoice, err := t.getPaymentAndInvoiceFromDatRecord(ctx, dataRecord, lineNo+1)
			if err != nil {
				return err
			}

			// Validate if payment method is equal to the payment method of file
			if payment.PaymentMethod.String != paymentMethod.String() {
				return fmt.Errorf("processing %v payment file but contains a record for %v in line %v", paymentMethod.String(), payment.PaymentMethod.String, lineNo+1)
			}

			previousResultCode := payment.ResultCode.String
			previousPaymentStatus := payment.PaymentStatus.String
			previousInvoiceStatus := invoice.Status.String

			// Validate invoice and payment records and retrieve the result codes
			validationResult, invoice, payment, err := t.validateDataRecord(dataRecord, lineNo+1, file.PaymentMethod, payment, invoice, useBulkAddValidatePh2, receiptDate)
			if err != nil {
				return fmt.Errorf("file validation failed: %v", err.Error())
			}

			paymentUpdateFields := []string{"result_code", "payment_date", "validated_date", "payment_status", "receipt_date", "updated_at"}
			// Set the payment amount to 0 if the payment status is FAILED
			if useBulkAddValidatePh2 && payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
				payment.Amount = database.Numeric(0)
				paymentUpdateFields = append(paymentUpdateFields, "amount")
			}

			if err := t.PaymentRepo.UpdateWithFields(ctx, tx, payment, paymentUpdateFields); err != nil {
				return fmt.Errorf("error updating payment record at line %v: %v", lineNo+1, err.Error())
			}

			invoiceUpdateFields := []string{"status", "updated_at"}
			if invoice.Status.String == invoice_pb.InvoiceStatus_PAID.String() {
				invoiceUpdateFields = append(invoiceUpdateFields, "outstanding_balance", "amount_paid")
			}

			if err := t.InvoiceRepo.UpdateWithFields(ctx, tx, invoice, invoiceUpdateFields); err != nil {
				return fmt.Errorf("error updating payment record at line %v: %v", lineNo+1, err.Error())
			}

			// Create action log data
			actionDetails := &utils.InvoiceActionLogDetails{
				InvoiceID:                invoice.InvoiceID.String,
				PaymentSequenceNumber:    payment.PaymentSequenceNumber.Int,
				Action:                   invoice_pb.InvoiceAction_NO_ACTION, // Ensures no action log will be created if not modified
				BulkPaymentValidationsID: bulkPaymentValidationID,
			}

			switch invoice.Status.String {
			case invoice_pb.InvoiceStatus_PAID.String():
				successfulPayments++
				actionDetails.Action = invoice_pb.InvoiceAction_INVOICE_PAID
			case invoice_pb.InvoiceStatus_ISSUED.String():
				if useBulkAddValidatePh2 && payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
					failedPayments++
				} else {
					pendingPayments++
				}
			case invoice_pb.InvoiceStatus_FAILED.String():
				failedPayments++
				actionDetails.Action = invoice_pb.InvoiceAction_INVOICE_FAILED
			case invoice_pb.InvoiceStatus_VOID.String():
				failedPayments++
			}

			// If there is no changes in invoice and payment status, set the action log to payment updated
			if previousInvoiceStatus == invoice.Status.String && previousPaymentStatus == payment.PaymentStatus.String {
				actionDetails.Action = invoice_pb.InvoiceAction_PAYMENT_UPDATED
			}

			// If payment status is failed after validating
			switch useBulkAddValidatePh2 {
			case true:
				if previousPaymentStatus == payment.PaymentStatus.String {
					actionDetails.Action = invoice_pb.InvoiceAction_PAYMENT_UPDATED
				}
				if previousPaymentStatus != payment.PaymentStatus.String && payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
					actionDetails.Action = invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED
				}

				// payment status is success after validating
				if payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() {
					actionDetails.Action = invoice_pb.InvoiceAction_PAYMENT_VALIDATE_SUCCESS
				}
				// Create action log record v2
				if actionDetails.Action != invoice_pb.InvoiceAction_NO_ACTION {
					if err := utils.CreateActionLogV2(ctx, tx, actionDetails, t.InvoiceActionLogRepo); err != nil {
						return err
					}
				}
			default:
				// Create action log record v1
				if actionDetails.Action != invoice_pb.InvoiceAction_NO_ACTION {
					if err := utils.CreateActionLog(ctx, tx, actionDetails, t.InvoiceActionLogRepo); err != nil {
						return err
					}
				}
			}
			// Create bulk validation details record
			if err := t.createBulkValidationDetails(ctx, tx, bulkPaymentValidationID, invoice.InvoiceID.String, payment, validationResult.ResultCode, previousResultCode); err != nil {
				return err
			}

			// Retrieve data for validated payment
			user, err := t.UserBasicInfoRepo.FindByID(ctx, t.DB, invoice.StudentID.String)
			if err != nil {
				return fmt.Errorf("error retrieving user record at line %v: %v", lineNo+1, err.Error())
			}

			totalAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
			if err != nil {
				return fmt.Errorf("error converting invoice total amount at line %v: %v", lineNo+1, err.Error())
			}

			validatedPayment := &ValidatedPayment{
				PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
				ResultCode:            validationResult.ResultCode,
				Amount:                totalAmount,
				PaymentMethod:         paymentMethod,
				InvoiceSequenceNumber: invoice.InvoiceSequenceNumber.Int,
				StudentID:             invoice.StudentID.String,
				StudentName:           user.Name.String,
				PaymentCreatedDate:    payment.CreatedAt.Time,
				InvoiceID:             invoice.InvoiceID.String,
				PaymentStatus:         payment.PaymentStatus.String,
			}

			validatedPayments = append(validatedPayments, validatedPayment)
		}

		// After successful iteration of data records, save the summary
		bulkPaymentValidation.SuccessfulPayments = database.Int4(successfulPayments)
		bulkPaymentValidation.FailedPayments = database.Int4(failedPayments)
		bulkPaymentValidation.PendingPayments = database.Int4(pendingPayments)
		bulkPaymentValidation.ValidationDate = database.Timestamptz(time.Now())

		if err := t.BulkPaymentValidationsRepo.UpdateWithFields(ctx, tx, bulkPaymentValidation, []string{"successful_payments", "failed_payments", "pending_payments", "updated_at"}); err != nil {
			return fmt.Errorf("error updating bulk payment validations: %v", err.Error())
		}

		return nil
	})

	paymentValidationResult := &PaymentValidationResult{
		ValidatedPayments:  validatedPayments,
		ValidationDate:     &bulkPaymentValidation.ValidationDate.Time,
		SuccessfulPayments: bulkPaymentValidation.SuccessfulPayments.Int,
		PendingPayments:    bulkPaymentValidation.PendingPayments.Int,
		FailedPayments:     bulkPaymentValidation.FailedPayments.Int,
	}

	return paymentValidationResult, err
}

// Create bulk validation detail record
func (t *BasePaymentFileValidator) createBulkValidationDetails(
	ctx context.Context,
	tx pgx.Tx,
	bulkPaymentValidationID string,
	invoiceID string,
	payment *entities.Payment,
	resultCode string,
	previousResultCode string,
) error {
	bulkPaymentValidationDtl := new(entities.BulkPaymentValidationsDetail)
	database.AllNullEntity(bulkPaymentValidationDtl)

	bulkPaymentValidationDtl.BulkPaymentValidationsID = database.Text(bulkPaymentValidationID)
	bulkPaymentValidationDtl.InvoiceID = database.Text(invoiceID)
	bulkPaymentValidationDtl.PaymentID = database.Text(payment.PaymentID.String)
	bulkPaymentValidationDtl.ValidatedResultCode = database.Text(resultCode)
	bulkPaymentValidationDtl.PreviousResultCode = database.Text(previousResultCode)
	bulkPaymentValidationDtl.PaymentStatus = database.Text(payment.PaymentStatus.String)

	_, err := t.BulkPaymentValidationsDetailRepo.Create(ctx, tx, bulkPaymentValidationDtl)
	if err != nil {
		return fmt.Errorf("unable to create bulk payment validations detail: %v", err.Error())
	}

	return nil
}

// Validate the data record and return invoice and payment entities
func (t *BasePaymentFileValidator) validateDataRecord(
	dataRecord *GenericPaymentFileRecord,
	lineNo int,
	paymentMethod int,
	payment *entities.Payment,
	invoice *entities.Invoice,
	useBulkAddValidatePh2 bool,
	receiptDate time.Time,
) (*ResultCodeValidation, *entities.Invoice, *entities.Payment, error) {
	var (
		validationResult *ResultCodeValidation
		err              error
	)
	if useBulkAddValidatePh2 {
		validationResult, err = t.getResultCodeAndStatusesPhase2(paymentMethod, dataRecord, lineNo, payment, invoice)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		validationResult, err = t.getResultCodeAndStatuses(paymentMethod, dataRecord, lineNo, payment, invoice)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// Remove payment date if the present payment status is successful but validation result failed
	// Happens at succeeding uploads/validations with the same payment number
	if payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() && validationResult.SystemResultCode != 0 {
		payment.PaymentDate = pgtype.Timestamptz{
			Status: pgtype.Null,
		}
	}
	// set initially receipt date null for update field
	payment.ReceiptDate = pgtype.Timestamptz{
		Status: pgtype.Null,
	}

	if validationResult.PaymentStatus == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL {
		if dataRecord.PaymentDate == nil {
			return nil, nil, nil, fmt.Errorf("payment date required at line %v", lineNo)
		}
		payment.PaymentDate = database.Timestamptz(*dataRecord.PaymentDate)

		// Set the validated date for CONVENIENCE STORE payment
		if payment.PaymentMethod.String == invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
			if dataRecord.ValidatedDate == nil {
				return nil, nil, nil, fmt.Errorf("validated date required if payment method is convenience store at line %v", lineNo)
			}
			payment.ValidatedDate = database.Timestamptz(*dataRecord.ValidatedDate)
		}
		// success payment set the receipt date
		payment.ReceiptDate = database.Timestamptz(receiptDate)
	}

	// Set the amount_paid and outstanding_balance of an invoice if it will be updated to PAID status
	// amount_paid = current amount paid value + amount value of the payment record
	// outstanding_balance = invoice total - the updated value of the amount paid
	if validationResult.InvoiceStatus.String() == invoice_pb.InvoiceStatus_PAID.String() {
		exactAmountPaid, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.AmountPaid, "2")
		if err != nil {
			return nil, nil, nil, err
		}
		newAmountPaid := exactAmountPaid + float64(dataRecord.Amount)

		exactInvoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return nil, nil, nil, err
		}
		newOutstandingBalance := exactInvoiceTotal - newAmountPaid

		invoice.AmountPaid = database.Numeric(float32(newAmountPaid))
		invoice.OutstandingBalance = database.Numeric(float32(newOutstandingBalance))
	}

	invoice.Status = database.Text(validationResult.InvoiceStatus.String())
	payment.PaymentStatus = database.Text(validationResult.PaymentStatus.String())
	payment.ResultCode = database.Text(validationResult.ResultCode)

	return validationResult, invoice, payment, nil
}

func (t *BasePaymentFileValidator) getPaymentAndInvoiceFromDatRecord(ctx context.Context, dataRecord *GenericPaymentFileRecord, lineNo int) (*entities.Payment, *entities.Invoice, error) {
	if len(strings.TrimSpace(dataRecord.PaymentNumber)) == 0 {
		return nil, nil, fmt.Errorf("payment number is required at line %v", lineNo)
	}

	paymentNo, err := strconv.Atoi(dataRecord.PaymentNumber)
	if err != nil {
		return nil, nil, fmt.Errorf("payment number %v not numeric at line %v", dataRecord.PaymentNumber, lineNo)
	}

	payment, err := t.PaymentRepo.FindByPaymentSequenceNumber(ctx, t.DB, paymentNo)
	if err != nil {
		return nil, nil, fmt.Errorf("error finding payment record at line %v: %v", lineNo, err.Error())
	}

	invoice, err := t.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, t.DB, payment.InvoiceID.String)
	if err != nil {
		return nil, nil, fmt.Errorf("error finding invoice record at line %v: %v", lineNo, err.Error())
	}

	return payment, invoice, nil
}

func (t *BasePaymentFileValidator) getResultCodeAndStatuses(
	paymentMethod int,
	dataRecord *GenericPaymentFileRecord,
	lineNo int,
	payment *entities.Payment,
	invoice *entities.Invoice,
) (*ResultCodeValidation, error) {
	var (
		// This map is used to validate if the result code from file is valid
		fileResultCodeMap map[string]string

		// This map is used to get the equivalent invoice status of the result code from file
		fileResultCodeInvoiceStatusMap map[string]invoice_pb.InvoiceStatus

		// This map is used to get the equivalent payment status of the result code from file
		fileResultCodePaymentStatusMap map[string]invoice_pb.PaymentStatus

		// This map is used to get the equivalent invoice status of the system code (amount not matched, invoice not issued, or both)
		systemResultCodeInvoiceStatusMap map[int]invoice_pb.InvoiceStatus

		// This map is used to get the equivalent payment status of the system code (amount not matched, invoice not issued, or both)
		systemResultCodePaymentStatusMap map[int]invoice_pb.PaymentStatus

		// Below are the equivalent system code based on Manabie validation (amount not matched, invoice not issued, or both)
		amountNotMatchedSystemCode          int
		notIssuedSystemCode                 int
		amountNotMatchedNotIssuedSystemCode int

		isInvoiceStatusIssued = invoice.Status.String == invoice_pb.InvoiceStatus_ISSUED.String()
		isInvoiceStatusVoid   = invoice.Status.String == invoice_pb.InvoiceStatus_VOID.String()
		isInvoiceStatusFailed = invoice.Status.String == invoice_pb.InvoiceStatus_FAILED.String()
		isInvoiceStatusPaid   = invoice.Status.String == invoice_pb.InvoiceStatus_PAID.String()

		isPaymentStatusFailed     = payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String()
		isPaymentStatusPending    = payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
		isPaymentStatusSuccessful = payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()

		systemResultCode  int
		paymentMethodName string
	)

	if paymentMethod != DirectDebit && paymentMethod != ConvenienceStore {
		return nil, errors.New("invalid payment method")
	}

	switch paymentMethod {
	case DirectDebit:
		paymentMethodName = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
	case ConvenienceStore:
		paymentMethodName = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
	}

	// handling scenario 17 for paid invoice and successful payment
	if isInvoiceStatusPaid && isPaymentStatusSuccessful {
		return nil, fmt.Errorf("invalid invoice paid status and payment successful status on payment method: %v", paymentMethodName)
	}

	// handling scenario 17 for failed invoice and failed payment with existing result code
	if isInvoiceStatusFailed && isPaymentStatusFailed && payment.ResultCode.String != "" {
		return nil, fmt.Errorf("invalid invoice failed status and payment failed status with existing result code: %v on payment method: %v", payment.ResultCode.String, paymentMethodName)
	}

	totalFloat, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return nil, fmt.Errorf("error converting invoice total to float at line %v: %v", lineNo, err.Error())
	}

	// Assign the maps and values depending on payment method
	switch paymentMethod {
	case DirectDebit:
		fileResultCodeMap = fileResultCodeDDMapping
		fileResultCodeInvoiceStatusMap = fileResultCodeDDInvoiceStatusMapping
		fileResultCodePaymentStatusMap = fileResultCodeDDPaymentStatusMapping
		systemResultCodeInvoiceStatusMap = systemResultCodeDDInvoiceStatusMapping
		systemResultCodePaymentStatusMap = systemResultCodeDDPaymentStatusMapping

		amountNotMatchedSystemCode = SysRCodeDDAmtNotMatched
		notIssuedSystemCode = SysRCodeDDNotIssued
		amountNotMatchedNotIssuedSystemCode = SysRCodeDDAmtNotMatchedNotIssued
	case ConvenienceStore:
		fileResultCodeMap = fileResultCodeCCMapping
		fileResultCodeInvoiceStatusMap = fileResultCodeCCInvoiceStatusMapping
		fileResultCodePaymentStatusMap = fileResultCodeCCPaymentStatusMapping
		systemResultCodeInvoiceStatusMap = systemResultCodeCCInvoiceStatusMapping
		systemResultCodePaymentStatusMap = systemResultCodeCCPaymentStatusMapping

		amountNotMatchedSystemCode = SysRCodeCCAmtNotMatched
		notIssuedSystemCode = SysRCodeCCNotIssued
		amountNotMatchedNotIssuedSystemCode = SysRCodeCCAmtNotMatchedNotIssued
	}

	// Check if the result code from file is valid
	_, ok := fileResultCodeMap[dataRecord.ResultCode]
	if !ok {
		return nil, fmt.Errorf("invalid %s result code at line %v: %v", paymentMethodName, lineNo, dataRecord.ResultCode)
	}

	// Get the invoice and payment status based on the result code from file
	// Scenario 1-8
	invoiceStatus := fileResultCodeInvoiceStatusMap[dataRecord.ResultCode]
	paymentStatus := fileResultCodePaymentStatusMap[dataRecord.ResultCode]

	fileResultCode := dataRecord.ResultCode
	if paymentMethod == ConvenienceStore {
		// Map the correct result code of Convenience Store based on the file result code. e.g. 02: 0, 01: 1, 03: 02
		fileResultCode = fileResultCodeCCMapping[fileResultCode]
	}

	// Check statuses and amount based on system validation and assign added system result code
	// The sequence of cases are important. It is required to check first both statues and the amount before checking them separately.
	// So that if status is not issued and amount is not matched, we can assign the right system result code.
	// Scenario 9-10
	switch {
	case (!isInvoiceStatusIssued || !isPaymentStatusPending) && totalFloat != float64(dataRecord.Amount):
		systemResultCode = amountNotMatchedNotIssuedSystemCode
	case totalFloat != float64(dataRecord.Amount):
		systemResultCode = amountNotMatchedSystemCode
	case !isInvoiceStatusIssued || !isPaymentStatusPending:
		systemResultCode = notIssuedSystemCode
	}

	// Generate the final result code.
	// If there are invalid status and amount, system result code will be greater than 0. And these system result codes have equivalent payment and invoice statuses.
	prefixCode := prefixCodePaymentMethodMapping[paymentMethod]
	finalResultCode := fmt.Sprintf("%v-R%v", prefixCode, fileResultCode)
	if systemResultCode > 0 {
		finalResultCode = fmt.Sprintf("%v-%v", finalResultCode, systemResultCode)
		invoiceStatus = systemResultCodeInvoiceStatusMap[systemResultCode]
		paymentStatus = systemResultCodePaymentStatusMap[systemResultCode]
	}

	// If the invoice status is VOIDED and payment status is FAILED, disregard the invoice and payment status change
	// Scenario 11-13
	if isInvoiceStatusVoid && isPaymentStatusFailed {
		invoiceStatus = invoice_pb.InvoiceStatus_VOID
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
	}

	// If both invoice and payment status is FAILED and payment method is convenience store, disregard the invoice and payment status change
	// Scenario 14-16
	if isInvoiceStatusFailed && isPaymentStatusFailed && paymentMethod == ConvenienceStore {
		invoiceStatus = invoice_pb.InvoiceStatus_FAILED
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
	}

	validationResult := &ResultCodeValidation{
		ResultCode:       finalResultCode,
		InvoiceStatus:    invoiceStatus,
		PaymentStatus:    paymentStatus,
		SystemResultCode: systemResultCode,
	}

	return validationResult, nil
}

func (t *BasePaymentFileValidator) getResultCodeAndStatusesPhase2(
	paymentMethod int,
	dataRecord *GenericPaymentFileRecord,
	lineNo int,
	payment *entities.Payment,
	invoice *entities.Invoice,
) (*ResultCodeValidation, error) {
	var (
		// This map is used to validate if the result code from file is valid
		fileResultCodeMap map[string]string

		// This map is used to get the equivalent invoice status of the result code from file
		fileResultCodeInvoiceStatusMap map[string]invoice_pb.InvoiceStatus

		// This map is used to get the equivalent payment status of the result code from file
		fileResultCodePaymentStatusMap map[string]invoice_pb.PaymentStatus

		// Below are the equivalent system code based on Manabie validation (amount not matched, invoice not issued, or both)
		amountNotMatchedSystemCode          int
		notIssuedSystemCode                 int
		amountNotMatchedNotIssuedSystemCode int

		isInvoiceStatusIssued = invoice.Status.String == invoice_pb.InvoiceStatus_ISSUED.String()
		isInvoiceStatusVoid   = invoice.Status.String == invoice_pb.InvoiceStatus_VOID.String()
		isInvoiceStatusPaid   = invoice.Status.String == invoice_pb.InvoiceStatus_PAID.String()

		isPaymentStatusFailed     = payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String()
		isPaymentStatusPending    = payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
		isPaymentStatusSuccessful = payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()

		systemResultCode  int
		paymentMethodName string
	)

	if paymentMethod != DirectDebit && paymentMethod != ConvenienceStore {
		return nil, errors.New("invalid payment method")
	}

	switch paymentMethod {
	case DirectDebit:
		paymentMethodName = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
	case ConvenienceStore:
		paymentMethodName = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
	}

	// handling scenario 17 for paid invoice and successful payment
	if isInvoiceStatusPaid && isPaymentStatusSuccessful {
		return nil, fmt.Errorf("invalid invoice paid status and payment successful status on payment method: %v", paymentMethodName)
	}

	// handling scenario 17 for failed invoice and failed payment with existing result code
	// phase 2 changes no failed invoice but issued or void with failed payment should terminate the validation process
	if (isInvoiceStatusIssued || isInvoiceStatusVoid) && isPaymentStatusFailed && payment.ResultCode.String != "" {
		return nil, fmt.Errorf("invalid invoice issued status and payment failed status with existing result code: %v on payment method: %v", payment.ResultCode.String, paymentMethodName)
	}

	totalFloat, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return nil, fmt.Errorf("error converting invoice total to float at line %v: %v", lineNo, err.Error())
	}

	// Assign the maps and values depending on payment method
	switch paymentMethod {
	case DirectDebit:
		fileResultCodeMap = fileResultCodeDDMapping
		fileResultCodeInvoiceStatusMap = fileResultCodeDDInvoiceStatusMappingV2
		fileResultCodePaymentStatusMap = fileResultCodeDDPaymentStatusMapping
		amountNotMatchedSystemCode = SysRCodeDDAmtNotMatched
		notIssuedSystemCode = SysRCodeDDNotIssued
		amountNotMatchedNotIssuedSystemCode = SysRCodeDDAmtNotMatchedNotIssued
	case ConvenienceStore:
		fileResultCodeMap = fileResultCodeCCMapping
		fileResultCodeInvoiceStatusMap = fileResultCodeCCInvoiceStatusMapping
		fileResultCodePaymentStatusMap = fileResultCodeCCPaymentStatusMapping

		amountNotMatchedSystemCode = SysRCodeCCAmtNotMatched
		notIssuedSystemCode = SysRCodeCCNotIssued
		amountNotMatchedNotIssuedSystemCode = SysRCodeCCAmtNotMatchedNotIssued
	}

	// Check if the result code from file is valid
	_, ok := fileResultCodeMap[dataRecord.ResultCode]
	if !ok {
		return nil, fmt.Errorf("invalid %s result code at line %v: %v", paymentMethodName, lineNo, dataRecord.ResultCode)
	}

	// Get the invoice and payment status based on the result code from file
	// Scenario 1-8
	invoiceStatus := fileResultCodeInvoiceStatusMap[dataRecord.ResultCode]
	paymentStatus := fileResultCodePaymentStatusMap[dataRecord.ResultCode]

	fileResultCode := dataRecord.ResultCode
	if paymentMethod == ConvenienceStore {
		// Map the correct result code of Convenience Store based on the file result code. e.g. 02: 0, 01: 1, 03: 02
		fileResultCode = fileResultCodeCCMapping[fileResultCode]
	}

	// Check statuses and amount based on system validation and assign added system result code
	// The sequence of cases are important. It is required to check first both statues and the amount before checking them separately.
	// So that if status is not issued and amount is not matched, we can assign the right system result code.
	// Scenario 9-10

	switch {
	case (!isInvoiceStatusIssued || !isPaymentStatusPending) && totalFloat != float64(dataRecord.Amount):
		systemResultCode = amountNotMatchedNotIssuedSystemCode
	case totalFloat != float64(dataRecord.Amount):
		systemResultCode = amountNotMatchedSystemCode
	case !isInvoiceStatusIssued || !isPaymentStatusPending:
		systemResultCode = notIssuedSystemCode
	}

	// Generate the final result code.
	// If there are invalid status and amount, system result code will be greater than 0. And these system result codes have equivalent payment and invoice statuses.
	prefixCode := prefixCodePaymentMethodMapping[paymentMethod]
	finalResultCode := fmt.Sprintf("%v-R%v", prefixCode, fileResultCode)
	if systemResultCode > 0 {
		finalResultCode = fmt.Sprintf("%v-%v", finalResultCode, systemResultCode)
		// leave the invoice status as it is and failed the payment
		invoiceStatus = invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[invoice.Status.String])
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
	}

	// If the invoice status is VOIDED and payment status is FAILED, disregard the invoice and payment status change
	// Scenario 11-13
	if isInvoiceStatusVoid && isPaymentStatusFailed {
		invoiceStatus = invoice_pb.InvoiceStatus_VOID
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
	}

	// If both invoice and payment status is FAILED and payment method is convenience store, disregard the invoice and payment status change
	// Scenario 14-16
	// no invoice failed status on bulk validate phase 2

	if isInvoiceStatusIssued && isPaymentStatusFailed && paymentMethod == ConvenienceStore {
		invoiceStatus = invoice_pb.InvoiceStatus_ISSUED
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED
	}

	validationResult := &ResultCodeValidation{
		ResultCode:       finalResultCode,
		InvoiceStatus:    invoiceStatus,
		PaymentStatus:    paymentStatus,
		SystemResultCode: systemResultCode,
	}

	return validationResult, nil
}
