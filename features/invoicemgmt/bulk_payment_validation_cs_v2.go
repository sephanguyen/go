package invoicemgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	pfutils "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_utils"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) hasResultCodeCategoryOnItsFileContent(ctx context.Context, category string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.generatePaymentDataBasedOnCategory(ctx, category)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generatePaymentDataBasedOnCategory(ctx context.Context, category string) context.Context {
	stepState := StepStateFromContext(ctx)
	paymentRecords := make([]*entities.Payment, 0)

	for i, paymentSeqNumber := range stepState.PaymentSeqNumbers {
		paymentEntity := &entities.Payment{
			Amount:                database.Numeric(float32(stepState.InvoiceTotalAmount[i])),
			PaymentSequenceNumber: database.Int4(paymentSeqNumber),
			PaymentDate: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
			ValidatedDate: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
			ResultCode: database.Text(category),
		}

		paymentRecords = append(paymentRecords, paymentEntity)
	}

	stepState.PaymentListToValidate = paymentRecords
	return StepStateToContext(ctx, stepState)
}

func (s *suite) hasPaymentDateOnItsFileContent(ctx context.Context, paymentDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	formattedDate := getFormattedTimestampDate(paymentDate)

	stepState.PaymentDate = formattedDate.AsTime()
	// date where the bulk validate process happen
	stepState.ValidatedDate = time.Now()

	for i := 0; i < len(stepState.PaymentListToValidate); i++ {
		stepState.PaymentListToValidate[i].PaymentDate = database.Timestamptz(stepState.PaymentDate)
		stepState.PaymentListToValidate[i].ValidatedDate = database.Timestamptz(stepState.ValidatedDate)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreNumberOfExistingInvoicesWithTotalAmount(ctx context.Context, numberOfInvoices int, status string, amount float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	s.NoOfStudentsInvoiceToCreate = numberOfInvoices

	ctx, err := s.thereAreExistingInvoicesWithStatusAndAmount(ctx, status, amount)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentInvoiceStatus = status

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateConvenienceStoreFile(ctx context.Context) ([]*pfutils.ConvenienceStoreFileDataRecord, error) {
	stepState := StepStateFromContext(ctx)
	dataRecords := make([]*pfutils.ConvenienceStoreFileDataRecord, 0)
	for _, paymentToValidate := range stepState.PaymentListToValidate {
		exactAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(paymentToValidate.Amount, "2")
		if err != nil {
			return nil, err
		}
		dataRecord := &pfutils.ConvenienceStoreFileDataRecord{
			Amount:          int(exactAmount),
			CodeForUser2:    fmt.Sprintf("%v", paymentToValidate.PaymentSequenceNumber.Int),
			Category:        paymentToValidate.ResultCode.String,
			TransferredDate: 0,
		}

		if paymentToValidate.ResultCode.String == FileRCodeCCPaidTransferred {
			dataRecord.TransferredDate, err = strconv.Atoi(generatePaymentDateFormat(paymentToValidate.PaymentDate.Time))
			if err != nil {
				return nil, fmt.Errorf("error cannot parse int transferred date: %v", err.Error())
			}

			dataRecord.CreatedDate, err = strconv.Atoi(generatePaymentDateFormat(paymentToValidate.CreatedAt.Time))
			if err != nil {
				return nil, fmt.Errorf("error cannot parse int created date: %v", err.Error())
			}

			dataRecord.DateOfReceipt, err = strconv.Atoi(generatePaymentDateFormat(stepState.ValidatedDate))
			if err != nil {
				return nil, fmt.Errorf("error cannot parse int receive date: %v", err.Error())
			}
		}

		dataRecords = append(dataRecords, dataRecord)
	}
	return dataRecords, nil
}

func (s *suite) paymentsHaveResultCodeWithCorrectExpectedResult(ctx context.Context, resultCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := s.StepState.Response.(*invoice_pb.CreateBulkPaymentValidationResponse)
	paymentValidationDetails := response.PaymentValidationDetail

	if len(paymentValidationDetails) != len(stepState.InvoiceIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment validation detail count %v but got %v", len(stepState.InvoiceIDs), len(paymentValidationDetails))
	}
	var expectedResult ExpectedResult

	switch stepState.CurrentInvoiceStatus {
	case invoice_pb.InvoiceStatus_VOID.String():
		if stepState.PaymentMethod == invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
			expectedResult = CSCategoryResultFromFileMapVoidInvoice[resultCode]
		} else {
			expectedResult = DDCategoryResultFromFileMapVoidInvoice[resultCode]
		}
	default:
		if stepState.PaymentMethod == invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
			expectedResult = CSCategoryResultFromFileMap[resultCode]
		} else {
			expectedResult = DDCategoryResultFromFileMap[resultCode]
		}
	}

	var successfulPayments, failedPayments, pendingPayments int

	// For each line from the endpoint's result file, check the expected result code
	for _, paymentValidationDetail := range paymentValidationDetails {
		expectedResult.DataRecordAmount = paymentValidationDetail.Amount

		err := s.validateResultV2(ctx, &expectedResult, paymentValidationDetail)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for file paid transferred: %v", err.Error())
		}
		switch expectedResult.FileResultCode {
		case FileRCodeCCPaidTransferred, FileRCodeDDAlreadyTransferred:
			if expectedResult.SystemValidationType == SystemValidationTypeNone {
				successfulPayments++
			} else {
				failedPayments++
			}
		default:
			if expectedResult.SystemValidationType == SystemValidationTypeNone {
				pendingPayments++
			} else {
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

func (s *suite) validateResultV2(ctx context.Context, expectedResult *ExpectedResult, receivedResult *invoice_pb.ImportPaymentValidationDetail) error {
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
	err = resultChecker.checkPaymentV2(payment)
	if err != nil {
		return err
	}

	// check transferred date that reflected on payment date for convenience store
	err = resultChecker.checkTransferredAndPaymentDateV2(ctx, payment)
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

func (c *PaymentValidationResultChecker) checkTransferredAndPaymentDateV2(ctx context.Context, payment *entities.Payment) error {
	stepState := StepStateFromContext(ctx)
	var err error
	const (
		paymentDateField = "payment date"
		validatedDate    = "validated date"
	)

	if payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() {
		err = isEqual(generatePaymentDateFormat(payment.PaymentDate.Time), generatePaymentDateFormat(stepState.PaymentDate), paymentDateField)
		if err != nil {
			return err
		}

		if stepState.PaymentMethod == invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
			err = isEqual(generatePaymentDateFormat(payment.ValidatedDate.Time), generatePaymentDateFormat(stepState.ValidatedDate), validatedDate)
			if err != nil {
				return err
			}
		}
	}

	nullDate := pgtype.Timestamptz{
		Status: pgtype.Null,
	}

	// Payment date should be null if result code is unsuccessful
	if c.ExpectedResult.InvoiceStatus != invoice_pb.InvoiceStatus_PAID && payment.PaymentDate != nullDate {
		return fmt.Errorf("expected nil payment date but got %v", payment.PaymentDate)
	}

	// Payment validated date should be null if result code is unsuccessful
	if c.ExpectedResult.InvoiceStatus != invoice_pb.InvoiceStatus_PAID && payment.ValidatedDate != nullDate {
		return fmt.Errorf("expected nil payment validated date but got %v", payment.ValidatedDate)
	}

	// Payment receipt date should be null if result code is unsuccessful
	if c.ExpectedResult.InvoiceStatus != invoice_pb.InvoiceStatus_PAID && payment.ReceiptDate != nullDate {
		return fmt.Errorf("expected nil payment receipt date but got %v", payment.ReceiptDate)
	}

	return nil
}

func (s *suite) hasAmountMismatchedOnItsFileContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < len(stepState.PaymentListToValidate); i++ {
		stepState.PaymentListToValidate[i].Amount = database.Numeric(0.00)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAPaymentThatIsNotMatchInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.PaymentListToValidate) > 0 {
		stepState.PaymentListToValidate[0].PaymentSequenceNumber = database.Int4(9999)
	}

	return StepStateToContext(ctx, stepState), nil
}
