package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

const (
	// File content types from the feature file
	FileContentTypeAllRCTransferred        = "all-return-codes-transferred"
	FileContentTypeMixedRC                 = "mixed-return-codes"
	FileContentTypeExistingRC              = "existing-return-codes"
	FileContentTypeDuplicatePaymentRecords = "duplicate-payment-records"
	FileContentTypeVoidNoStatusChange      = "void-invoice-no-status-change"
	FileContentTypeFailedNoStatusChange    = "failed-invoice-no-status-change"

	// Manabie system validation types used for test data setup
	SystemValidationTypeNone                          = -1
	SystemValidationTypeInvoiceNotIssued              = 0
	SystemValidationTypeAmtNotMatched                 = 1
	SystemValidationTypeInvoiceNotissuedAmtNotMatched = 2
	SystemValidationTypeInvoiceNotIssuedFailedPayment = 3
	SystemValidationTypeInvoiceIssuedFailedPayment    = 4

	// direct debit validation
	SystemValidationTypeShortOfCashInBankAccount    = 6
	SystemValidationTypeBankAccountNotExist         = 7
	SystemValidationTypeStopTransferCustomerReason  = 8
	SystemValidationTypeNoContractDD                = 9
	SystemValidationTypeStopTransferConsignorReason = 10
	SystemValidationTypeDDOthers                    = 11

	// Successful return codes for Direct Debit and Convenience Store
	SuccessfulResultCodeDD = "D-R0"
	SuccessfulResultCodeCC = "C-R0"

	// Convenience Store result codes
	CCPaidNotTransferred = "C-R1"
	CCRevokedCancelled   = "C-R2"

	// Direct Debit result codes
	DDShortInCash                 = "D-R1"
	DDBankAccountNotExist         = "D-R2"
	DDStopTransferCustomerReason  = "D-R3"
	DDNoContract                  = "D-R4"
	DDStopTransferConsignorReason = "D-R8"
	DDOthers                      = "D-R9"

	PaymentDateFutureDate = 1

	// Indicates statuses will not be updated
	ExistingInvoiceStatusNotSet = -1
	ExistingPaymentStatusNotSet = -1
)

type PaymentValidationResultChecker struct {
	ExpectedResult *ExpectedResult
	ReceivedResult *invoice_pb.ImportPaymentValidationDetail
}

func (c *PaymentValidationResultChecker) checkResultCode() error {
	if c.ExpectedResult.ResultCode != c.ReceivedResult.Result {
		return fmt.Errorf("expected result code %v but got %v", c.ExpectedResult.ResultCode, c.ReceivedResult.Result)
	}
	return nil
}

func (c *PaymentValidationResultChecker) checkPayment(payment *entities.Payment) error {
	if c.ExpectedResult.PaymentStatus.String() != payment.PaymentStatus.String {
		return fmt.Errorf("expected payment status %v but got %v from DB (payment)", c.ExpectedResult.PaymentStatus.String(), payment.PaymentStatus.String)
	}

	if c.ExpectedResult.PaymentStatus.String() != c.ReceivedResult.PaymentStatus {
		return fmt.Errorf("expected payment status %v but got %v from response", c.ExpectedResult.PaymentStatus.String(), c.ReceivedResult.PaymentStatus)
	}

	return nil
}

func (c *PaymentValidationResultChecker) checkPaymentV2(payment *entities.Payment) error {
	if c.ExpectedResult.PaymentStatus.String() != payment.PaymentStatus.String {
		return fmt.Errorf("expected payment status %v but got %v from DB (payment)", c.ExpectedResult.PaymentStatus.String(), payment.PaymentStatus.String)
	}

	if c.ExpectedResult.PaymentStatus.String() != c.ReceivedResult.PaymentStatus {
		return fmt.Errorf("expected payment status %v but got %v from response", c.ExpectedResult.PaymentStatus.String(), c.ReceivedResult.PaymentStatus)
	}

	exactPaymentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(payment.Amount, "2")
	if err != nil {
		return err
	}

	expectedAmount := c.ExpectedResult.DataRecordAmount
	if payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
		expectedAmount = 0
	}

	if exactPaymentAmount != expectedAmount {
		return fmt.Errorf("expected payment amount %v but got %v", expectedAmount, exactPaymentAmount)
	}

	return nil
}

func (c *PaymentValidationResultChecker) checkTransferredAndPaymentDate(ctx context.Context, payment *entities.Payment, fileContentType string) error {
	stepState := StepStateFromContext(ctx)
	var err error
	if stepState.PaymentMethod == invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
		const paymentDateField = "payment date"
		const validatedDate = "validated date"

		switch fileContentType {
		case FileContentTypeAllRCTransferred:
			err = multierr.Combine(
				isEqual(generatePaymentDateFormat(payment.PaymentDate.Time), generatePaymentDateFormat(stepState.PaymentDate), paymentDateField),
				isEqual(generatePaymentDateFormat(payment.ValidatedDate.Time), generatePaymentDateFormat(stepState.ValidatedDate), validatedDate),
			)
		case FileContentTypeMixedRC:
			if c.ExpectedResult.ResultCode == SuccessfulResultCodeCC {
				err = multierr.Combine(
					isEqual(generatePaymentDateFormat(payment.PaymentDate.Time), generatePaymentDateFormat(stepState.PaymentDate), paymentDateField),
					isEqual(generatePaymentDateFormat(payment.ValidatedDate.Time), generatePaymentDateFormat(stepState.ValidatedDate), validatedDate),
				)
			}
		case FileContentTypeExistingRC, FileContentTypeDuplicatePaymentRecords:
			if c.ExpectedResult.InvoiceActionLogType == invoice_pb.InvoiceAction_INVOICE_PAID {
				err = multierr.Combine(
					isEqual(generatePaymentDateFormat(payment.PaymentDate.Time), generatePaymentDateFormat(stepState.PaymentDate), paymentDateField),
					isEqual(generatePaymentDateFormat(payment.ValidatedDate.Time), generatePaymentDateFormat(stepState.ValidatedDate), validatedDate),
				)
			}
		}

		if err != nil {
			return err
		}
	}

	paymentDateNull := pgtype.Timestamptz{
		Status: pgtype.Null,
	}

	// Payment date should be null if result code is unsuccessful
	if c.ExpectedResult.InvoiceStatus != invoice_pb.InvoiceStatus_PAID && payment.PaymentDate != paymentDateNull {
		return fmt.Errorf("expected nil payment date but got %v", payment.PaymentDate)
	}

	return nil
}

func (c *PaymentValidationResultChecker) checkInvoice(invoice *entities.Invoice) error {
	// Check invoice status
	if c.ExpectedResult.InvoiceStatus.String() != invoice.Status.String {
		return fmt.Errorf("expected invoice status %v but got %v", c.ExpectedResult.InvoiceStatus.String(), invoice.Status.String)
	}

	// Check invoice amount paid
	if invoice.Status.String == invoice_pb.InvoiceStatus_PAID.String() {
		exactInvoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return err
		}
		exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
		if err != nil {
			return err
		}
		exactAmountPaid, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.AmountPaid, "2")
		if err != nil {
			return err
		}

		if exactAmountPaid != c.ExpectedResult.DataRecordAmount {
			return fmt.Errorf("expected amount paid %v but got %v", c.ExpectedResult.DataRecordAmount, exactAmountPaid)
		}

		newOutstandingBalance := exactInvoiceTotal - c.ExpectedResult.DataRecordAmount
		if exactOutstandingBalance != newOutstandingBalance {
			return fmt.Errorf("expected outstanding balance %v but got %v", newOutstandingBalance, exactOutstandingBalance)
		}
	}

	return nil
}

func (c *PaymentValidationResultChecker) checkActionLog(actionLog *entities.InvoiceActionLog) error {
	// No expected action log created
	if c.ExpectedResult.InvoiceActionLogType == invoice_pb.InvoiceAction_NO_ACTION {
		if actionLog != nil {
			return fmt.Errorf("expected no invoice action log created but found one")
		}
	} else {
		if c.ExpectedResult.InvoiceActionLogType.String() != actionLog.Action.String {
			return fmt.Errorf("expected invoice action log type %v but got %v", c.ExpectedResult.InvoiceActionLogType.String(), actionLog.Action.String)
		}
	}

	return nil
}

func (c *PaymentValidationResultChecker) checkBulkPaymentValidationDetail(bulkPaymentValidationDtl *entities.BulkPaymentValidationsDetail) error {
	// Check response result code
	if c.ExpectedResult.ResultCode != bulkPaymentValidationDtl.ValidatedResultCode.String {
		return fmt.Errorf("expected validated result code %v but got %v", c.ExpectedResult.ResultCode, bulkPaymentValidationDtl.ValidatedResultCode.String)
	}

	// Check payment status
	if c.ExpectedResult.PaymentStatus.String() != bulkPaymentValidationDtl.PaymentStatus.String {
		return fmt.Errorf("expected payment status %v but got %v from DB (bulk payment validation detail)", c.ExpectedResult.PaymentStatus.String(), bulkPaymentValidationDtl.PaymentStatus.String)
	}

	return nil
}

func (c *PaymentValidationResultChecker) checkReceiptDate(payment *entities.Payment) error {
	if payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() {
		err := compareReceiptDateWhenBulkValidateProcess(payment, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}

func compareReceiptDateWhenBulkValidateProcess(payment *entities.Payment, dateBulkProcess time.Time) error {
	dateFormat := "2006-01-02"
	if payment.ReceiptDate.Time.Format(dateFormat) != dateBulkProcess.Format(dateFormat) {
		return fmt.Errorf("payment receipt date expected: %v but got: %v", dateBulkProcess.Format(dateFormat), payment.ReceiptDate.Time.Format(dateFormat))
	}
	return nil
}

const (
	FileRCodeCCPaidNotTransferred = "01"
	FileRCodeCCPaidTransferred    = "02"
	FileRCodeCCRevokedCancelled   = "03"

	FileRCodeDDAlreadyTransferred = "0"
	FileRCodeDDShortCash          = "1"
	FileRCodeDDAcctNonExisting    = "2"
	FileRCodeDDStopCustReason     = "3"
	FileRCodeDDNoContract         = "4"
	FileRCodeDDStopConsignReason  = "8"
	FileRCodeDDOthers             = "9"
)

var (
	// Convenience Store Result Code From File
	CSCategoryResultFromFileMap = map[string]ExpectedResult{
		// result codes: CR0, CR1,CR2
		// paid by customer but not transferred yet, leave the status as it is for both payment and invoice
		"C-R1": {
			FileResultCode:       FileRCodeCCPaidNotTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           CCPaidNotTransferred,
			SystemValidationType: SystemValidationTypeNone,
		},
		// paid by customer & already transferred
		"C-R0": {
			FileResultCode:       FileRCodeCCPaidTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_SUCCESS,
			ResultCode:           SuccessfulResultCodeCC,
			SystemValidationType: SystemValidationTypeNone,
		},
		// report was cancelled
		"C-R2": {
			FileResultCode:       FileRCodeCCRevokedCancelled,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           CCRevokedCancelled,
			SystemValidationType: SystemValidationTypeNone,
		},
		// amount mismatched paid by customer & already transferred
		"C-R0-1": {
			FileResultCode:       FileRCodeCCPaidTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "C-R0-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		// amount mismatched paid by customer but not transferred yet
		"C-R1-1": {
			FileResultCode:       FileRCodeCCPaidNotTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "C-R1-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		// amount mismatched report was cancelled
		"C-R2-1": {
			FileResultCode:       FileRCodeCCRevokedCancelled,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "C-R2-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		// issued failed payment remains both invoice and payment status C-R0-2, C-R1-2 and C-R2-2
		"C-R0-2": {
			FileResultCode:       FileRCodeCCPaidTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R0-2",
			SystemValidationType: SystemValidationTypeInvoiceIssuedFailedPayment,
		},
		"C-R1-2": {
			FileResultCode:       FileRCodeCCPaidNotTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R1-2",
			SystemValidationType: SystemValidationTypeInvoiceIssuedFailedPayment,
		},
		"C-R2-2": {
			FileResultCode:       FileRCodeCCRevokedCancelled,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R2-2",
			SystemValidationType: SystemValidationTypeInvoiceIssuedFailedPayment,
		},
		"C-R0-3": {
			FileResultCode:       FileRCodeCCPaidTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R0-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"C-R1-3": {
			FileResultCode:       FileRCodeCCPaidNotTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R1-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"C-R2-3": {
			FileResultCode:       FileRCodeCCRevokedCancelled,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R2-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
	}

	CSCategoryResultFromFileMapVoidInvoice = map[string]ExpectedResult{
		"C-R0-2": {
			FileResultCode:       FileRCodeCCPaidTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R0-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"C-R1-2": {
			FileResultCode:       FileRCodeCCPaidNotTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R1-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"C-R2-2": {
			FileResultCode:       FileRCodeCCRevokedCancelled,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R2-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"C-R0-3": {
			FileResultCode:       FileRCodeCCPaidTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R0-3",
			SystemValidationType: SystemValidationTypeInvoiceNotissuedAmtNotMatched,
		},
		"C-R1-3": {
			FileResultCode:       FileRCodeCCPaidNotTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R1-3",
			SystemValidationType: SystemValidationTypeInvoiceNotissuedAmtNotMatched,
		},
		"C-R2-3": {
			FileResultCode:       FileRCodeCCRevokedCancelled,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "C-R2-3",
			SystemValidationType: SystemValidationTypeInvoiceNotissuedAmtNotMatched,
		},
	}
	// Direct Debit Category Result codes
	DDCategoryResultFromFileMap = map[string]ExpectedResult{
		"D-R0": {
			FileResultCode:       FileRCodeDDAlreadyTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_SUCCESS,
			ResultCode:           SuccessfulResultCodeDD,
			SystemValidationType: SystemValidationTypeNone,
		},
		"D-R1": {
			FileResultCode:       FileRCodeDDShortCash,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           DDShortInCash,
			SystemValidationType: SystemValidationTypeShortOfCashInBankAccount,
		},
		"D-R2": {
			FileResultCode:       FileRCodeDDAcctNonExisting,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           DDBankAccountNotExist,
			SystemValidationType: SystemValidationTypeBankAccountNotExist,
		},
		"D-R3": {
			FileResultCode:       FileRCodeDDStopCustReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           DDStopTransferCustomerReason,
			SystemValidationType: SystemValidationTypeStopTransferCustomerReason,
		},
		"D-R4": {
			FileResultCode:       FileRCodeDDNoContract,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           DDNoContract,
			SystemValidationType: SystemValidationTypeNoContractDD,
		},
		"D-R8": {
			FileResultCode:       FileRCodeDDStopConsignReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           DDStopTransferConsignorReason,
			SystemValidationType: SystemValidationTypeStopTransferConsignorReason,
		},
		"D-R9": {
			FileResultCode:       FileRCodeDDOthers,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           DDOthers,
			SystemValidationType: SystemValidationTypeDDOthers,
		},
		"D-R0-1": {
			FileResultCode:       FileRCodeDDAlreadyTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "D-R0-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R1-1": {
			FileResultCode:       FileRCodeDDShortCash,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "D-R1-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R2-1": {
			FileResultCode:       FileRCodeDDAcctNonExisting,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "D-R2-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R3-1": {
			FileResultCode:       FileRCodeDDStopCustReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "D-R3-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R4-1": {
			FileResultCode:       FileRCodeDDNoContract,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "D-R4-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R8-1": {
			FileResultCode:       FileRCodeDDStopConsignReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "D-R8-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R9-1": {
			FileResultCode:       FileRCodeDDOthers,
			InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED,
			ResultCode:           "D-R9-1",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
	}

	DDCategoryResultFromFileMapVoidInvoice = map[string]ExpectedResult{
		"D-R0-2": {
			FileResultCode:       FileRCodeDDAlreadyTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R0-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"D-R1-2": {
			FileResultCode:       FileRCodeDDShortCash,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R1-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"D-R2-2": {
			FileResultCode:       FileRCodeDDAcctNonExisting,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R2-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"D-R3-2": {
			FileResultCode:       FileRCodeDDStopCustReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R3-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"D-R4-2": {
			FileResultCode:       FileRCodeDDNoContract,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R4-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"D-R8-2": {
			FileResultCode:       FileRCodeDDStopConsignReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R8-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"D-R9-2": {
			FileResultCode:       FileRCodeDDOthers,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R9-2",
			SystemValidationType: SystemValidationTypeInvoiceNotIssuedFailedPayment,
		},
		"D-R0-3": {
			FileResultCode:       FileRCodeDDAlreadyTransferred,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R0-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R1-3": {
			FileResultCode:       FileRCodeDDShortCash,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R1-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R2-3": {
			FileResultCode:       FileRCodeDDAcctNonExisting,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R2-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R3-3": {
			FileResultCode:       FileRCodeDDStopCustReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R3-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R4-3": {
			FileResultCode:       FileRCodeDDNoContract,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R4-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R8-3": {
			FileResultCode:       FileRCodeDDStopConsignReason,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R8-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
		"D-R9-3": {
			FileResultCode:       FileRCodeDDOthers,
			InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
			PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
			InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
			ResultCode:           "D-R9-3",
			SystemValidationType: SystemValidationTypeAmtNotMatched,
		},
	}
)

func generateDuplicatePaymentExpectedResultV2(resultCode string) []*ExpectedResult {
	expectedResults := []*ExpectedResult{}

	resultCodeSlice := strings.Split(resultCode, "/")

	for _, resultCode := range resultCodeSlice {
		expectedResult := CSCategoryResultFromFileMap[resultCode]
		expectedResults = append(expectedResults, &expectedResult)
	}

	return expectedResults
}
