package invoicemgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	pfutils "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_utils"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	gocsv "github.com/gocarina/gocsv"
	fixedwidth "github.com/ianlopshire/go-fixedwidth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) thereArePreexistingNumberOfExistingInvoicesWithStatus(ctx context.Context, numberOfInvoices string, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	numberOfInvoicesInt, _ := strconv.Atoi(numberOfInvoices)
	s.NoOfStudentsInvoiceToCreate = numberOfInvoicesInt

	ctx, err := s.thereAreExistingInvoicesWithStatus(ctx, status)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreExistingPaymentsForThoseInvoicesForPaymentMethodWithStatus(ctx context.Context, paymentMethod, paymentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.generateExistingPaymentsForInvoices(ctx, paymentMethod, paymentStatus, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.PaymentMethod = paymentMethod
	stepState.LatestPaymentStatuses = append(stepState.LatestPaymentStatuses, paymentStatus)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) signedinUserUploadsThePaymentFileForPaymentMethod(ctx context.Context, signedInuser string, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.PaymentListToValidate) > 0 {
		// generate payment file
		switch paymentMethod {
		case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
			dataRecords, err := s.generateConvenienceStoreFile(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), nil
			}

			fileBytes, _ := gocsv.MarshalBytes(&dataRecords)
			stepState.PaymentFile = fileBytes
		case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
			fileContentsStr, err := s.generateDirectDebitFile(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), nil
			}
			stepState.PaymentFile = []byte(fileContentsStr)
		}
	}

	ctx, err := s.signedAsAccount(ctx, signedInuser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	paymentMethodInt := invoice_pb.PaymentMethod_value[paymentMethod]

	content := stepState.PaymentFile
	// Encode the payload
	if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableEncodePaymentRequestFiles) {
		content, err = utils.EncodeByteToShiftJIS(stepState.PaymentFile)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	req := &invoice_pb.CreateBulkPaymentValidationRequest{
		PaymentMethod: invoice_pb.PaymentMethod(paymentMethodInt),
		Payload:       content,
	}
	// only direct debit payment method can input direct debit payment date
	if paymentMethod == "DIRECT_DEBIT" {
		req.DirectDebitPaymentDate = timestamppb.New(stepState.PaymentDate)
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).CreateBulkPaymentValidation(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

// Validates the response file and table records (invoice, payment, action log, bulk payment validations, bulk payment validation details)
func (s *suite) receivesExpectedResultWithCorrectDBRecordsBasedOnFileContentType(ctx context.Context, fileContentType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := s.StepState.Response.(*invoice_pb.CreateBulkPaymentValidationResponse)
	paymentValidationDetails := response.PaymentValidationDetail

	if len(paymentValidationDetails) != len(stepState.InvoiceIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment validation detail count %v but got %v", len(stepState.InvoiceIDs), len(paymentValidationDetails))
	}

	expectedResult := s.getExpectedResults(ctx, fileContentType)

	var successfulPayments, failedPayments, pendingPayments int

	// For each line from the endpoint's result file, check the expected result code
	for _, paymentValidationDetail := range paymentValidationDetails {
		if fileContentType == FileContentTypeAllRCTransferred {
			expectedResultCode := SuccessfulResultCodeDD
			if stepState.PaymentMethod == invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
				expectedResultCode = SuccessfulResultCodeCC
			}

			expectedResult := &ExpectedResult{
				ResultCode:           expectedResultCode,
				InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
				InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_PAID,
			}
			expectedResult.DataRecordAmount = paymentValidationDetail.Amount

			err := s.validateResult(ctx, expectedResult, paymentValidationDetail, fileContentType)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for %v type: %v", FileContentTypeAllRCTransferred, err.Error())
			}

			successfulPayments++
		} else {
			// Retrieve the response payment number's index from the payment number seq list
			// Index determines the order in expected result array
			for i, paymentSeqNumber := range stepState.PaymentSeqNumbers {
				paymentSeqNumberDetail := paymentValidationDetail.PaymentSequenceNumber
				if paymentSeqNumber == paymentSeqNumberDetail {
					expectedResult[i].DataRecordAmount = paymentValidationDetail.Amount
					err := s.validateResult(ctx, expectedResult[i], paymentValidationDetail, fileContentType)
					if err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed at %v: %v", i, err.Error())
					}
					break
				}
			}
			switch paymentValidationDetail.Result {
			case SuccessfulResultCodeDD, SuccessfulResultCodeCC:
				successfulPayments++
			case CCPaidNotTransferred, CCRevokedCancelled:
				pendingPayments++
			default:
				failedPayments++
			}
		}
	}

	err := validateResponsePaymentStatusAndValidationDate(response, successfulPayments, failedPayments, pendingPayments)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// Performs records and result code validation
func (s *suite) validateResult(ctx context.Context, expectedResult *ExpectedResult, receivedResult *invoice_pb.ImportPaymentValidationDetail, fileContentType string) error {
	// stepState := StepStateFromContext(ctx)

	paymentRepo := &repositories.PaymentRepo{}
	invoiceRepo := &repositories.InvoiceRepo{}
	actionLogRepo := &repositories.InvoiceActionLogRepo{}
	bulkPaymentValidationDtlRepo := &repositories.BulkPaymentValidationsDetailRepo{}

	resultChecker := &PaymentValidationResultChecker{
		ExpectedResult: expectedResult,
		ReceivedResult: receivedResult,
	}

	// Check response result code
	err := resultChecker.checkResultCode()
	if err != nil {
		return err
	}

	payment, err := paymentRepo.FindByPaymentSequenceNumber(ctx, s.InvoiceMgmtPostgresDBTrace, int(receivedResult.PaymentSequenceNumber))
	if err != nil {
		return fmt.Errorf("paymentRepo.FindByPaymentSequenceNumber err: %v", err)
	}

	// Check payment status
	err = resultChecker.checkPayment(payment)
	if err != nil {
		return err
	}

	// check transferred date that reflected on payment date for convenience store
	err = resultChecker.checkTransferredAndPaymentDate(ctx, payment, fileContentType)
	if err != nil {
		return err
	}

	// check receipt date
	err = resultChecker.checkReceiptDate(payment)
	if err != nil {
		return err
	}

	invoice, err := invoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, receivedResult.InvoiceId)
	if err != nil {
		return fmt.Errorf("invoiceRepo.RetrieveInvoiceByInvoiceID err: %v", err)
	}

	// Check invoice status and amount paid
	err = resultChecker.checkInvoice(invoice)
	if err != nil {
		return err
	}

	actionLog, err := actionLogRepo.GetLatestRecordByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoice.InvoiceID.String)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		// ignore no rows error
		return fmt.Errorf("actionLogRepo.GetLatestRecordByInvoiceID err: %v", err)
	}

	// Check action log type
	err = resultChecker.checkActionLog(actionLog)
	if err != nil {
		return err
	}

	bulkPaymentValidationDtl, err := bulkPaymentValidationDtlRepo.FindByPaymentID(ctx, s.InvoiceMgmtPostgresDBTrace, payment.PaymentID.String)
	if err != nil {
		return fmt.Errorf("bulkPaymentValidationDtlRepo.FindByPaymentID err: %v", err)
	}

	err = resultChecker.checkBulkPaymentValidationDetail(bulkPaymentValidationDtl)
	if err != nil {
		return err
	}

	return nil
}

// ExpectedResult holds the expected statuses and result code based on the file result code and system validation
type ExpectedResult struct {
	// Values validated to confirm result
	PaymentStatus        invoice_pb.PaymentStatus
	InvoiceStatus        invoice_pb.InvoiceStatus
	InvoiceActionLogType invoice_pb.InvoiceAction
	ResultCode           string // e.g., DR-0
	PaymentDate          int
	// Values used to set up existing records
	FileResultCode        string // e.g., 01, 9
	SystemValidationType  int    // e.g., SystemValidationTypeInvoiceNotIssued
	ExistingResultCode    string // e.g., CR-1
	ExistingPaymentStatus invoice_pb.PaymentStatus
	ExistingInvoiceStatus invoice_pb.InvoiceStatus
	DataRecordAmount      float64
}

// Creates combination of statuses and expected result code
func (s *suite) getExpectedResults(ctx context.Context, fileContentType string) []*ExpectedResult {
	stepState := StepStateFromContext(ctx)

	expectedResults := make([]*ExpectedResult, 0)

	switch fileContentType {
	case FileContentTypeMixedRC:
		if stepState.PaymentMethod == "DIRECT_DEBIT" {
			expectedResults = []*ExpectedResult{
				{
					FileResultCode:       pfutils.FileRCodeDDAlreadyTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_PAID,
					ResultCode:           SuccessfulResultCodeDD,
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDAlreadyTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "D-R0-1",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDShortCash,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "D-R1",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:        pfutils.FileRCodeDDShortCash,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "D-R1-2",
					SystemValidationType:  SystemValidationTypeInvoiceNotIssued,
					ExistingInvoiceStatus: invoice_pb.InvoiceStatus_PAID,
					ExistingPaymentStatus: ExistingPaymentStatusNotSet,
				},
				{
					FileResultCode:        pfutils.FileRCodeDDShortCash,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "D-R1-2",
					SystemValidationType:  SystemValidationTypeInvoiceNotIssued,
					ExistingInvoiceStatus: ExistingInvoiceStatusNotSet,
					ExistingPaymentStatus: invoice_pb.PaymentStatus_PAYMENT_FAILED,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDAcctNonExisting,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "D-R2",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDStopCustReason,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "D-R3",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:        pfutils.FileRCodeDDStopCustReason,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "D-R3-3",
					SystemValidationType:  SystemValidationTypeInvoiceNotissuedAmtNotMatched,
					ExistingInvoiceStatus: invoice_pb.InvoiceStatus_PAID,
					ExistingPaymentStatus: ExistingPaymentStatusNotSet,
				},
				{
					FileResultCode:        pfutils.FileRCodeDDStopCustReason,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "D-R3-3",
					SystemValidationType:  SystemValidationTypeInvoiceNotissuedAmtNotMatched,
					ExistingInvoiceStatus: ExistingInvoiceStatusNotSet,
					ExistingPaymentStatus: invoice_pb.PaymentStatus_PAYMENT_FAILED,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDNoContract,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "D-R4",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDStopConsignReason,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "D-R8",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDOthers,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "D-R9",
					SystemValidationType: SystemValidationTypeNone,
				},
			}
		} else {
			expectedResults = []*ExpectedResult{
				{
					FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_PAID,
					ResultCode:           SuccessfulResultCodeCC,
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "C-R0-1",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
				{
					FileResultCode:        pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "C-R0-2",
					SystemValidationType:  SystemValidationTypeInvoiceNotIssued,
					ExistingInvoiceStatus: invoice_pb.InvoiceStatus_PAID,
					ExistingPaymentStatus: ExistingPaymentStatusNotSet,
				},
				{
					FileResultCode:        pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "C-R0-2",
					SystemValidationType:  SystemValidationTypeInvoiceNotIssued,
					ExistingInvoiceStatus: ExistingInvoiceStatusNotSet,
					ExistingPaymentStatus: invoice_pb.PaymentStatus_PAYMENT_FAILED,
				},
				{
					FileResultCode:        pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "C-R0-3",
					SystemValidationType:  SystemValidationTypeInvoiceNotissuedAmtNotMatched,
					ExistingInvoiceStatus: invoice_pb.InvoiceStatus_PAID,
					ExistingPaymentStatus: ExistingPaymentStatusNotSet,
				},
				{
					FileResultCode:        pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "C-R0-3",
					SystemValidationType:  SystemValidationTypeInvoiceNotissuedAmtNotMatched,
					ExistingInvoiceStatus: ExistingInvoiceStatusNotSet,
					ExistingPaymentStatus: invoice_pb.PaymentStatus_PAYMENT_FAILED,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "C-R1",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "C-R1-1",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
				{
					FileResultCode:        pfutils.FileRCodeCCPaidNotTransferred,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "C-R1-2",
					SystemValidationType:  SystemValidationTypeInvoiceNotIssued,
					ExistingInvoiceStatus: invoice_pb.InvoiceStatus_PAID,
					ExistingPaymentStatus: ExistingPaymentStatusNotSet,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCRevokedCancelled,
					InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "C-R2",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCRevokedCancelled,
					InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:           "C-R2-1",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
				{
					FileResultCode:        pfutils.FileRCodeCCRevokedCancelled,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "C-R2-3",
					SystemValidationType:  SystemValidationTypeInvoiceNotissuedAmtNotMatched,
					ExistingInvoiceStatus: invoice_pb.InvoiceStatus_PAID,
					ExistingPaymentStatus: ExistingPaymentStatusNotSet,
				},
				{
					FileResultCode:        pfutils.FileRCodeCCRevokedCancelled,
					InvoiceStatus:         invoice_pb.InvoiceStatus_FAILED,
					PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType:  invoice_pb.InvoiceAction_INVOICE_FAILED,
					ResultCode:            "C-R2-3",
					SystemValidationType:  SystemValidationTypeInvoiceNotissuedAmtNotMatched,
					ExistingInvoiceStatus: ExistingInvoiceStatusNotSet,
					ExistingPaymentStatus: invoice_pb.PaymentStatus_PAYMENT_FAILED,
				},
			}
		}
	case FileContentTypeExistingRC:
		expectedResults = []*ExpectedResult{
			{
				ResultCode:           "C-R2",
				InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				FileResultCode:       pfutils.FileRCodeCCRevokedCancelled,
				ExistingResultCode:   "C-R1",
			},
			{
				ResultCode:           "C-R1",
				InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
				ExistingResultCode:   "C-R2",
			},
			{
				ResultCode:           "C-R0",
				InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
				InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_PAID,
				FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
				ExistingResultCode:   "C-R1",
			},
			{
				ResultCode:           "C-R0",
				InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
				InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_PAID,
				FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
				ExistingResultCode:   "C-R2",
			},
		}
	case FileContentTypeVoidNoStatusChange:
		if stepState.PaymentMethod == "DIRECT_DEBIT" {
			expectedResults = []*ExpectedResult{
				{
					FileResultCode:       pfutils.FileRCodeDDAlreadyTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "D-R0-2",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDShortCash,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "D-R1-2",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDAlreadyTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "D-R0-3",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
				{
					FileResultCode:       pfutils.FileRCodeDDShortCash,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "D-R1-3",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
			}
		} else {
			expectedResults = []*ExpectedResult{
				{
					FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "C-R0-2",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "C-R1-2",
					SystemValidationType: SystemValidationTypeNone,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "C-R0-3",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
				{
					FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
					InvoiceStatus:        invoice_pb.InvoiceStatus_VOID,
					PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
					InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
					ResultCode:           "C-R1-3",
					SystemValidationType: SystemValidationTypeAmtNotMatched,
				},
			}
		}
	case FileContentTypeFailedNoStatusChange:
		expectedResults = []*ExpectedResult{
			{
				FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
				InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				ResultCode:           "C-R0-2",
				SystemValidationType: SystemValidationTypeNone,
			},
			{
				FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
				InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				ResultCode:           "C-R1-2",
				SystemValidationType: SystemValidationTypeNone,
			},
			{
				FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
				InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				ResultCode:           "C-R0-3",
				SystemValidationType: SystemValidationTypeAmtNotMatched,
			},
			{
				FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
				InvoiceStatus:        invoice_pb.InvoiceStatus_FAILED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_FAILED,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				ResultCode:           "C-R1-3",
				SystemValidationType: SystemValidationTypeAmtNotMatched,
			},
		}
	}

	return expectedResults
}

// Creates text file bytes for the endpoint
func (s *suite) createFileForDirectDebit(ctx context.Context, fileContentType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	expectedResults := s.getExpectedResults(ctx, fileContentType)

	// Create the header record part
	headerRecord := pfutils.DirectDebitFileHeaderRecord{
		DataCategory: pfutils.DataTypeHeaderRecord,
	}

	headerRecordBytes, err := fixedwidth.Marshal(headerRecord)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error marshalling header record: %v", err)
	}

	// Create the data record part
	dataRecordsBytes := make([][]byte, 0)

	for i, paymentSeqNumber := range stepState.PaymentSeqNumbers {
		dataRecord := &pfutils.DirectDebitFileDataRecord{
			DataCategory:   pfutils.DataTypeDataRecord,
			CustomerNumber: fmt.Sprintf("%v", paymentSeqNumber),
			DepositAmount:  int(stepState.InvoiceTotalAmount[i]),
		}

		// Use successful result code; otherwise, create a file with multiple result codes
		if fileContentType == FileContentTypeAllRCTransferred {
			dataRecord.ResultCode = pfutils.FileRCodeDDAlreadyTransferred
		} else {
			expectedResult := expectedResults[i]
			dataRecord.ResultCode = expectedResult.FileResultCode

			amount, err := s.setUpAmountAndInvoiceStatus(ctx, expectedResult, paymentSeqNumber, int(stepState.InvoiceTotalAmount[i]))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error setUpAmountAndInvoiceStatus: %v", err.Error())
			}

			dataRecord.DepositAmount = amount
		}

		dataRecordBytesData, err := fixedwidth.Marshal(dataRecord)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error marshalling data record: %v", err)
		}

		dataRecordsBytes = append(dataRecordsBytes, dataRecordBytesData)
	}

	// Create the trailer record part
	trailerRecord := pfutils.DirectDebitFileTrailerRecord{
		DataCategory: pfutils.DataTypeTrailerRecord,
	}

	trailerRecordBytes, err := fixedwidth.Marshal(trailerRecord)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error marshalling trailer record: %v", err)
	}

	// Create the end record part
	endRecord := pfutils.DirectDebitFileEndRecord{
		DataCategory: pfutils.DataTypeEndRecord,
	}

	endRecordBytes, err := fixedwidth.Marshal(endRecord)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error marshalling end record: %v", err)
	}

	// Combine all file parts into a single string before converting into bytes
	fileContentsStr := fmt.Sprintf("%v\n", string(headerRecordBytes))
	for _, dataBytes := range dataRecordsBytes {
		fileContentsStr += fmt.Sprintf("%v\n", string(dataBytes))
	}
	fileContentsStr += fmt.Sprintf("%v\n%v", string(trailerRecordBytes), string(endRecordBytes))

	stepState.PaymentFile = []byte(fileContentsStr)

	return StepStateToContext(ctx, stepState), nil
}

// Creates CSV file bytes for the endpoint
func (s *suite) createFileForConvenienceStore(ctx context.Context, fileContentType string, formattedDate time.Time) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	expectedResults := s.getExpectedResults(ctx, fileContentType)
	dataRecords := make([]*pfutils.ConvenienceStoreFileDataRecord, 0)

	for i, paymentSeqNumber := range stepState.PaymentSeqNumbers {
		var err error
		dataRecord := &pfutils.ConvenienceStoreFileDataRecord{
			Amount:          int(stepState.InvoiceTotalAmount[i]),
			CodeForUser2:    fmt.Sprintf("%v", paymentSeqNumber),
			TransferredDate: 0,
		}

		switch fileContentType {
		case FileContentTypeAllRCTransferred, FileContentTypeDuplicatePaymentRecords:
			dataRecord.Category = pfutils.FileRCodeCCPaidTransferred
			dataRecord.TransferredDate, err = strconv.Atoi(generatePaymentDateFormat(formattedDate))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int transferred date: %v", err.Error())
			}

			dataRecord.CreatedDate, err = strconv.Atoi(generatePaymentDateFormat(formattedDate))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int created date: %v", err.Error())
			}

			dataRecord.DateOfReceipt, err = strconv.Atoi(generatePaymentDateFormat(stepState.ValidatedDate))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int receive date: %v", err.Error())
			}

		case FileContentTypeExistingRC:
			expectedResult := expectedResults[i]
			dataRecord.Category = expectedResult.FileResultCode

			// Add transfer date to successful payment
			if expectedResult.InvoiceActionLogType == invoice_pb.InvoiceAction_INVOICE_PAID {
				dataRecord.TransferredDate, err = strconv.Atoi(generatePaymentDateFormat(formattedDate))
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int transferred date: %v", err.Error())
				}
			}

			dataRecord.DateOfReceipt, err = strconv.Atoi(generatePaymentDateFormat(stepState.ValidatedDate))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int receive date: %v", err.Error())
			}

			if err = s.updatePaymentResultCodeByPaymentSeqNumber(ctx, int(paymentSeqNumber), expectedResult.ExistingResultCode); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error updateInvoiceByPaymentSeqNumber: %v", err.Error())
			}
		case FileContentTypeMixedRC, FileContentTypeVoidNoStatusChange, FileContentTypeFailedNoStatusChange:
			expectedResult := expectedResults[i]
			dataRecord.Category = expectedResult.FileResultCode

			amount, err := s.setUpAmountAndInvoiceStatus(ctx, expectedResult, paymentSeqNumber, int(stepState.InvoiceTotalAmount[i]))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error setUpAmountAndInvoiceStatus: %v", err.Error())
			}

			// Add transfer date to the successful payment
			if expectedResult.SystemValidationType == SystemValidationTypeNone {
				dataRecord.TransferredDate, err = strconv.Atoi(generatePaymentDateFormat(formattedDate))
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int transferred date: %v", err.Error())
				}

				dataRecord.DateOfReceipt, err = strconv.Atoi(generatePaymentDateFormat(stepState.ValidatedDate))
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int receive date: %v", err.Error())
				}
			}

			dataRecord.Amount = amount
		}

		dataRecords = append(dataRecords, dataRecord)
	}

	// Create duplicate records
	if fileContentType == FileContentTypeDuplicatePaymentRecords {
		futureDate := time.Now().AddDate(0, 1, 0)
		futureDateInt, _ := strconv.Atoi(fmt.Sprintf("%v%02d%02d", futureDate.Year(), int(futureDate.Month()), futureDate.Day()))

		duplicatePaymentGreaterCreatedDate := dataRecords[1]
		duplicatePaymentGreaterCreatedDate.CreatedDate = futureDateInt
		duplicatePaymentGreaterCreatedDate.TransferredDate = futureDateInt
		dataRecords = append(dataRecords, duplicatePaymentGreaterCreatedDate)

		duplicatePaymentSameCreatedDate := dataRecords[2]
		duplicatePaymentSameCreatedDate.TransferredDate = futureDateInt
		dataRecords = append(dataRecords, duplicatePaymentGreaterCreatedDate)
	}

	fileBytes, _ := gocsv.MarshalBytes(&dataRecords)
	stepState.PaymentFile = fileBytes

	return StepStateToContext(ctx, stepState), nil
}

// Update invoice status and amount depending on the validation type
func (s *suite) setUpAmountAndInvoiceStatus(ctx context.Context, expectedResult *ExpectedResult, paymentSeqNumber int32, amount int) (int, error) {
	switch expectedResult.SystemValidationType {
	case SystemValidationTypeInvoiceNotIssued:
		if expectedResult.ExistingInvoiceStatus != ExistingInvoiceStatusNotSet {
			err := s.updateInvoiceByPaymentSeqNumber(ctx, int(paymentSeqNumber), expectedResult.ExistingInvoiceStatus.String())
			if err != nil {
				return amount, err
			}
		}
		if expectedResult.ExistingPaymentStatus != ExistingPaymentStatusNotSet {
			err := s.updatePaymentStatusByPaymentSeqNumber(ctx, int(paymentSeqNumber), expectedResult.ExistingPaymentStatus.String())
			if err != nil {
				return amount, err
			}
		}
	case SystemValidationTypeAmtNotMatched:
		amount = int(amount) * 2
	case SystemValidationTypeInvoiceNotissuedAmtNotMatched:
		if expectedResult.ExistingInvoiceStatus != ExistingInvoiceStatusNotSet {
			err := s.updateInvoiceByPaymentSeqNumber(ctx, int(paymentSeqNumber), expectedResult.ExistingInvoiceStatus.String())
			if err != nil {
				return amount, err
			}
		}

		if expectedResult.ExistingPaymentStatus != ExistingPaymentStatusNotSet {
			err := s.updatePaymentStatusByPaymentSeqNumber(ctx, int(paymentSeqNumber), expectedResult.ExistingPaymentStatus.String())
			if err != nil {
				return amount, err
			}
		}

		amount = int(amount) * 2
	}

	return amount, nil
}

// Util function to update the invoice status based on associated payment's sequence number
func (s *suite) updateInvoiceByPaymentSeqNumber(ctx context.Context, paymentSeqNumber int, invoiceStatus string) error {
	paymentRepo := &repositories.PaymentRepo{}
	invoiceRepo := &repositories.InvoiceRepo{}

	payment, err := paymentRepo.FindByPaymentSequenceNumber(ctx, s.InvoiceMgmtPostgresDBTrace, paymentSeqNumber)
	if err != nil {
		return fmt.Errorf("paymentRepo.FindByPaymentSequenceNumber err: %v", err)
	}

	invoice, err := invoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, payment.InvoiceID.String)
	if err != nil {
		return fmt.Errorf("invoiceRepo.RetrieveInvoiceByInvoiceID err: %v", err)
	}

	invoice.Status = database.Text(invoiceStatus)
	err = invoiceRepo.Update(ctx, s.InvoiceMgmtPostgresDBTrace, invoice)
	if err != nil {
		return fmt.Errorf("invoiceRepo.Update err: %v", err)
	}

	return nil
}

func (s *suite) updatePaymentResultCodeByPaymentSeqNumber(ctx context.Context, paymentSeqNumber int, expectedResultCode string) error {
	paymentRepo := &repositories.PaymentRepo{}

	payment, err := paymentRepo.FindByPaymentSequenceNumber(ctx, s.InvoiceMgmtPostgresDBTrace, paymentSeqNumber)
	if err != nil {
		return fmt.Errorf("paymentRepo.FindByPaymentSequenceNumber err: %v", err)
	}

	payment.ResultCode = database.Text(expectedResultCode)

	if err := paymentRepo.UpdateWithFields(ctx, s.InvoiceMgmtPostgresDBTrace, payment, []string{"result_code"}); err != nil {
		return fmt.Errorf("paymentRepo.Update err: %v", err)
	}

	return nil
}

func (s *suite) thereIsAnExistingPaymentFileForPaymentMethod(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// set payment method
	stepState.PaymentMethod = paymentMethod
	// empty payment file before adding content
	stepState.PaymentFile = []byte("")

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasFileContentTypeWithTransferredDateForSuccessfulPayments(ctx context.Context, fileContentType, transferredDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	formattedDate := getFormattedTimestampDate(transferredDate)

	stepState.PaymentDate = formattedDate.AsTime()
	stepState.ValidatedDate = formattedDate.AsTime().Add(24 * time.Hour)

	switch stepState.PaymentMethod {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
		ctx, err = s.createFileForConvenienceStore(ctx, fileContentType, formattedDate.AsTime())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error creating payment file: %v", err)
		}
	case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
		ctx, err = s.createFileForDirectDebit(ctx, fileContentType)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error creating payment file: %v", err)
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid payment method: %v", stepState.PaymentMethod)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasDuplicatePaymentRecordsWithDateAndResultCodeSequenceOnPaymentFile(ctx context.Context, createdDateStrFormat, resultCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.PaymentDate = time.Now()
	stepState.ValidatedDate = time.Now().Add(24 * time.Hour)

	createdDateSlice := strings.Split(createdDateStrFormat, "-")

	createdDates := make([]time.Time, 0, len(createdDateSlice))
	for _, createdDateStr := range createdDateSlice {
		createdDates = append(createdDates, getFormattedTimestampDate(createdDateStr).AsTime())
	}

	var expectedResults []*ExpectedResult
	switch isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableBulkAddValidatePh2) {
	case true:
		expectedResults = generateDuplicatePaymentExpectedResultV2(resultCode)
	default:
		expectedResults = generateDuplicatePaymentExpectedResult(resultCode)
	}

	dataRecords := make([]*pfutils.ConvenienceStoreFileDataRecord, 0)

	// use the first payment to be duplicated in a payment file
	for i := 0; i < len(expectedResults); i++ {
		var err error
		dataRecord := &pfutils.ConvenienceStoreFileDataRecord{
			Amount:          int(stepState.InvoiceTotalAmount[0]),
			CodeForUser2:    fmt.Sprintf("%v", stepState.PaymentSeqNumbers[0]),
			TransferredDate: 0,
		}

		expectedResult := expectedResults[i]
		dataRecord.Category = expectedResult.FileResultCode

		amount, err := s.setUpAmountAndInvoiceStatus(ctx, expectedResult, stepState.PaymentSeqNumbers[0], int(stepState.InvoiceTotalAmount[0]))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error setUpAmountAndInvoiceStatus: %v", err.Error())
		}

		// Add transfer date to the successful payment
		if expectedResult.SystemValidationType == SystemValidationTypeNone {
			dataRecord.TransferredDate, err = strconv.Atoi(generatePaymentDateFormat(stepState.PaymentDate))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int transferred date: %v", err.Error())
			}

			dataRecord.DateOfReceipt, err = strconv.Atoi(generatePaymentDateFormat(stepState.ValidatedDate))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int receive date: %v", err.Error())
			}
		}

		dataRecord.CreatedDate, err = strconv.Atoi(generatePaymentDateFormat(createdDates[i]))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot parse int created date: %v", err.Error())
		}

		dataRecord.Amount = amount

		dataRecords = append(dataRecords, dataRecord)
	}

	fileBytes, _ := gocsv.MarshalBytes(&dataRecords)
	stepState.PaymentFile = fileBytes

	return StepStateToContext(ctx, stepState), nil
}

func generateDuplicatePaymentExpectedResult(resultCode string) []*ExpectedResult {
	expectedResults := []*ExpectedResult{}

	resulCodeSlice := strings.Split(resultCode, "-")

	for _, resultCode := range resulCodeSlice {
		switch resultCode {
		case "CR0":
			expectedResults = append(expectedResults, &ExpectedResult{
				FileResultCode:       pfutils.FileRCodeCCPaidTransferred,
				InvoiceStatus:        invoice_pb.InvoiceStatus_PAID,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
				InvoiceActionLogType: invoice_pb.InvoiceAction_INVOICE_PAID,
				ResultCode:           SuccessfulResultCodeCC,
				SystemValidationType: SystemValidationTypeNone,
			})
		case "CR1":
			expectedResults = append(expectedResults, &ExpectedResult{
				FileResultCode:       pfutils.FileRCodeCCPaidNotTransferred,
				InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				ResultCode:           "C-R1",
				SystemValidationType: SystemValidationTypeNone,
			})
		case "CR2":
			expectedResults = append(expectedResults, &ExpectedResult{
				FileResultCode:       pfutils.FileRCodeCCRevokedCancelled,
				InvoiceStatus:        invoice_pb.InvoiceStatus_ISSUED,
				PaymentStatus:        invoice_pb.PaymentStatus_PAYMENT_PENDING,
				InvoiceActionLogType: invoice_pb.InvoiceAction_PAYMENT_UPDATED,
				ResultCode:           "C-R2",
				SystemValidationType: SystemValidationTypeNone,
			})
		}
	}

	return expectedResults
}

func (s *suite) receivesExpectedRecordForDuplicatePaymentWithResultCode(ctx context.Context, actualResultCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := s.StepState.Response.(*invoice_pb.CreateBulkPaymentValidationResponse)
	paymentValidationDetails := response.PaymentValidationDetail

	if len(paymentValidationDetails) != len(stepState.InvoiceIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment validation detail count %v but got %v", len(stepState.InvoiceIDs), len(paymentValidationDetails))
	}

	var successfulPayments, failedPayments, pendingPayments int

	// For each line from the endpoint's result file, check the expected result code
	for _, paymentValidationDetail := range paymentValidationDetails {
		paymentSeqNumberDetail := paymentValidationDetail.PaymentSequenceNumber
		if stepState.PaymentSeqNumbers[0] == paymentSeqNumberDetail {
			var expectedResult *ExpectedResult

			switch isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableBulkAddValidatePh2) {
			case true:
				expectedResult = generateDuplicatePaymentExpectedResultV2(actualResultCode)[0]
			default:
				expectedResult = generateDuplicatePaymentExpectedResult(actualResultCode)[0]
			}

			expectedResult.DataRecordAmount = paymentValidationDetail.Amount

			err := s.validateResult(ctx, expectedResult, paymentValidationDetail, FileContentTypeDuplicatePaymentRecords)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed at payment sequence number %v: %v", stepState.PaymentSeqNumbers[0], err.Error())
			}
		}

		switch paymentValidationDetail.Result {
		case SuccessfulResultCodeDD, SuccessfulResultCodeCC:
			successfulPayments++
		case CCPaidNotTransferred, CCRevokedCancelled:
			pendingPayments++
		default:
			failedPayments++
		}
	}

	err := validateResponsePaymentStatusAndValidationDate(response, successfulPayments, failedPayments, pendingPayments)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateExistingPaymentsForInvoices(ctx context.Context, paymentMethod, paymentStatus, existingResultCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	paymentMethodInt := invoice_pb.PaymentMethod_value[paymentMethod]

	paymentRepo := &repositories.PaymentRepo{}
	invoiceRepo := &repositories.InvoiceRepo{}
	for i, invoiceID := range stepState.InvoiceIDs {
		// Inject the new invoice ID for payment creation
		stepState.InvoiceID = invoiceID
		stepState.StudentID = stepState.StudentIds[i]
		ctx = StepStateToContext(ctx, stepState)

		ctx, err := s.createPayment(ctx, invoice_pb.PaymentMethod(paymentMethodInt), paymentStatus, existingResultCode, true)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("createPayment error: %v", err.Error())
		}

		payment, err := paymentRepo.FindByPaymentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.PaymentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("paymentRepo.FindByPaymentID err: %v", err)
		}

		invoice, err := invoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invoiceRepo.RetrieveInvoiceByInvoiceID err: %v", err)
		}

		totalAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("utils.GetFloat64ExactValueAndDecimalPlaces err: %v", err)
		}

		stepState.PaymentSeqNumbers = append(stepState.PaymentSeqNumbers, payment.PaymentSequenceNumber.Int)
		stepState.InvoiceTotalAmount = append(stepState.InvoiceTotalAmount, totalAmount)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseExistingPaymentsHaveExistingResultCode(ctx context.Context, existingResultCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, paymentSeqNumber := range stepState.PaymentSeqNumbers {
		if err := s.updatePaymentResultCodeByPaymentSeqNumber(ctx, int(paymentSeqNumber), existingResultCode); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error updateInvoiceByPaymentSeqNumber: %v", err.Error())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func validateResponsePaymentStatusAndValidationDate(response *invoice_pb.CreateBulkPaymentValidationResponse, successfulPayments, failedPayments, pendingPayments int) error {
	if int(response.SuccessfulPayments) != successfulPayments {
		return fmt.Errorf("expected successful validation count %v but got %v", successfulPayments, int(response.SuccessfulPayments))
	}

	if int(response.FailedPayments) != failedPayments {
		return fmt.Errorf("expected failed validation count %v but got %v", failedPayments, int(response.FailedPayments))
	}

	if int(response.PendingPayments) != pendingPayments {
		return fmt.Errorf("expected pending validation count %v but got %v", pendingPayments, int(response.PendingPayments))
	}

	if response.ValidationDate == nil {
		return fmt.Errorf("expected non-nil validation date")
	}
	return nil
}
