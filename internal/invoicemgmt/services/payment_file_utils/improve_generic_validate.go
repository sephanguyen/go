package paymentfileutils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
)

type paymentNumberMapsAndList struct {
	PaymentLineNoMap       map[int]int
	PaymentDataRecordMap   map[int]*GenericPaymentFileRecord
	PaymentNumberEntityMap map[int]*entities.PaymentInvoiceUserMap
	PaymentNumbers         []int
}

type paymentStatusCount struct {
	SuccessfulPayments int
	FailedPayments     int
	PendingPayments    int
}

type entityLists struct {
	PaymentToUpdate                 []*entities.Payment
	InvoiceToUpdate                 []*entities.Invoice
	InvoiceActionLogToCreate        []*entities.InvoiceActionLog
	PaymentValidationDetailToCreate []*entities.BulkPaymentValidationsDetail
}

type featureFlags struct {
	UseBulkAddValidatePh2 bool
}

func (t *BasePaymentFileValidator) ImprovedGenericValidate(ctx context.Context, file *GenericPaymentFile) (*PaymentValidationResult, error) {
	var (
		entityList        *entityLists
		statusCount       *paymentStatusCount
		validatedPayments []*ValidatedPayment
		receiptDate       = time.Now()
	)

	paymentNumberLineMap := identifyDuplicatePaymentNumbers(file.GenericPaymentData)
	paymentMethod := getPaymentMethodFromFile(file)

	// Initialize BulkPaymentValidation entity
	bulkPaymentValidation := initBulkPaymentValidationEntity(paymentMethod)

	useBulkAddValidatePh2, err := t.UnleashClient.IsFeatureEnabled(constant.EnableBulkAddValidatePh2, t.Env)
	if err != nil {
		return nil, fmt.Errorf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableBulkAddValidatePh2, err)
	}

	err = database.ExecInTx(ctx, t.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// Create a record to be updated later; ID referenced by details record
		bulkPaymentValidationID, err := t.BulkPaymentValidationsRepo.Create(ctx, tx, bulkPaymentValidation)
		if err != nil {
			return fmt.Errorf("unable to create bulk payment validations: %v", err.Error())
		}

		// Get the valid data records and collect the valid payment numbers
		paymentMaps, err := getPaymentNumberMaps(file, paymentMethod, paymentNumberLineMap)
		if err != nil {
			return err
		}

		// Insert the sequence numbers to temporary table
		err = t.PaymentRepo.InsertPaymentNumbersTempTable(ctx, tx, paymentMaps.PaymentNumbers)
		if err != nil {
			return fmt.Errorf("t.PaymentRepo.InsertPaymentNumbersTempTable err: %v", err)
		}

		// Get the invoice, payment and user entities from the temp table
		invoicePaymentUser, err := t.PaymentRepo.FindPaymentInvoiceUserFromTempTable(ctx, tx)
		if err != nil {
			return fmt.Errorf("t.PaymentRepo.FindPaymentInvoiceUserFromTempTable err: %v", err)
		}

		// Check if the result length is same with number of payment sequence numbers
		if len(invoicePaymentUser) != len(paymentMaps.PaymentNumbers) {
			return fmt.Errorf("there are missing %d invoice/payment/user data", len(paymentMaps.PaymentNumbers)-len(invoicePaymentUser))
		}

		// Map the entities to payment sequence number
		paymentNumberEntityMap := make(map[int]*entities.PaymentInvoiceUserMap)
		for _, e := range invoicePaymentUser {
			paymentNumberEntityMap[int(e.Payment.PaymentSequenceNumber.Int)] = e
		}
		paymentMaps.PaymentNumberEntityMap = paymentNumberEntityMap

		// Get the entities to be updated and created
		// It also validates the payment
		entityList, statusCount, validatedPayments, err = t.validateAndGetEntitiesToUpdateAndCreate(ctx, paymentMaps,
			&featureFlags{UseBulkAddValidatePh2: useBulkAddValidatePh2},
			&getEntitiesToUpdateAndCreateParam{
				BulkPaymentValidationID: bulkPaymentValidationID,
				ReceiptDate:             receiptDate,
				File:                    file,
				PaymentMethod:           paymentMethod,
			})
		if err != nil {
			return err
		}

		// Update payments
		err = t.PaymentRepo.UpdateMultipleWithFields(ctx, tx, entityList.PaymentToUpdate,
			[]string{"result_code", "payment_date", "validated_date", "payment_status", "receipt_date", "updated_at", "amount"})
		if err != nil {
			return fmt.Errorf("t.PaymentRepo.UpdateMultipleWithFields err: %v", err)
		}

		// Update invoices
		err = t.InvoiceRepo.UpdateMultipleWithFields(ctx, tx, entityList.InvoiceToUpdate, []string{"status", "updated_at", "outstanding_balance", "amount_paid"})
		if err != nil {
			return fmt.Errorf("t.InvoiceRepo.UpdateMultipleWithFields err: %v", err)
		}

		// Create Action Logs
		err = t.InvoiceActionLogRepo.CreateMultiple(ctx, tx, entityList.InvoiceActionLogToCreate)
		if err != nil {
			return fmt.Errorf("t.InvoiceActionLogRepo.CreateMultiple err: %v", err)
		}

		// Create Bulk Payment Validation Details
		err = t.BulkPaymentValidationsDetailRepo.CreateMultiple(ctx, tx, entityList.PaymentValidationDetailToCreate)
		if err != nil {
			return fmt.Errorf("t.BulkPaymentValidationsDetailRepo.CreateMultiple err: %v", err)
		}

		// After successful iteration of data records, save the summary
		bulkPaymentValidation.SuccessfulPayments = database.Int4(int32(statusCount.SuccessfulPayments))
		bulkPaymentValidation.FailedPayments = database.Int4(int32(statusCount.FailedPayments))
		bulkPaymentValidation.PendingPayments = database.Int4(int32(statusCount.PendingPayments))
		bulkPaymentValidation.ValidationDate = database.Timestamptz(time.Now())

		if err := t.BulkPaymentValidationsRepo.UpdateWithFields(ctx, tx, bulkPaymentValidation, []string{"successful_payments", "failed_payments", "pending_payments", "updated_at"}); err != nil {
			return fmt.Errorf("error updating bulk payment validations: %v", err.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	paymentValidationResult := &PaymentValidationResult{
		ValidatedPayments:  validatedPayments,
		ValidationDate:     &bulkPaymentValidation.ValidationDate.Time,
		SuccessfulPayments: bulkPaymentValidation.SuccessfulPayments.Int,
		PendingPayments:    bulkPaymentValidation.PendingPayments.Int,
		FailedPayments:     bulkPaymentValidation.FailedPayments.Int,
	}

	return paymentValidationResult, nil
}

func getPaymentNumberMaps(file *GenericPaymentFile, paymentMethod invoice_pb.PaymentMethod, paymentNumberLineMap map[string]*DataRecordCreatedDateLine) (*paymentNumberMapsAndList, error) {
	paymentLineNoMap := make(map[int]int)
	paymentDataRecordMap := make(map[int]*GenericPaymentFileRecord)
	paymentNumbers := []int{}
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

		paymentNo, err := strconv.Atoi(dataRecord.PaymentNumber)
		if err != nil {
			return nil, fmt.Errorf("payment number %v not numeric at line %v", dataRecord.PaymentNumber, lineNo)
		}

		paymentLineNoMap[paymentNo] = lineNo
		paymentDataRecordMap[paymentNo] = dataRecord
		paymentNumbers = append(paymentNumbers, paymentNo)
	}

	return &paymentNumberMapsAndList{
		PaymentLineNoMap:     paymentLineNoMap,
		PaymentDataRecordMap: paymentDataRecordMap,
		PaymentNumbers:       paymentNumbers,
	}, nil
}

func initBulkPaymentValidationEntity(paymentMethod invoice_pb.PaymentMethod) *entities.BulkPaymentValidations {
	bulkPaymentValidation := new(entities.BulkPaymentValidations)
	database.AllNullEntity(bulkPaymentValidation)

	bulkPaymentValidation.PaymentMethod = database.Text(paymentMethod.String())
	bulkPaymentValidation.SuccessfulPayments = database.Int4(0)
	bulkPaymentValidation.FailedPayments = database.Int4(0)
	bulkPaymentValidation.PendingPayments = database.Int4(0)
	bulkPaymentValidation.ValidationDate = database.Timestamptz(time.Now())

	return bulkPaymentValidation
}

type getEntitiesToUpdateAndCreateParam struct {
	BulkPaymentValidationID string
	ReceiptDate             time.Time
	File                    *GenericPaymentFile
	PaymentMethod           invoice_pb.PaymentMethod
}

func (t *BasePaymentFileValidator) validateAndGetEntitiesToUpdateAndCreate(
	ctx context.Context,
	paymentMaps *paymentNumberMapsAndList,
	featureFlags *featureFlags,
	param *getEntitiesToUpdateAndCreateParam,
) (*entityLists, *paymentStatusCount, []*ValidatedPayment, error) {
	validatedPayments := make([]*ValidatedPayment, 0)

	invoiceToUpdate := make([]*entities.Invoice, len(paymentMaps.PaymentNumbers))
	paymentToUpdate := make([]*entities.Payment, len(paymentMaps.PaymentNumbers))
	actionLogToCreate := []*entities.InvoiceActionLog{}
	validationDetailToCreate := make([]*entities.BulkPaymentValidationsDetail, len(paymentMaps.PaymentNumbers))

	statusCount := &paymentStatusCount{}

	for i, paymentNo := range paymentMaps.PaymentNumbers {
		paymentInvoice := paymentMaps.PaymentNumberEntityMap[paymentNo]
		dataRecord := paymentMaps.PaymentDataRecordMap[paymentNo]
		lineNo := paymentMaps.PaymentLineNoMap[paymentNo]

		payment := paymentInvoice.Payment
		invoice := paymentInvoice.Invoice

		if payment.PaymentMethod.String != param.PaymentMethod.String() {
			return nil, nil, nil, fmt.Errorf("processing %v payment file but contains a record for %v in line %v", param.PaymentMethod.String(), payment.PaymentMethod.String, lineNo+1)
		}

		previousResultCode := payment.ResultCode.String
		previousPaymentStatus := payment.PaymentStatus.String
		previousInvoiceStatus := invoice.Status.String

		validationResult, invoice, payment, err := t.validateDataRecord(dataRecord, lineNo+1, param.File.PaymentMethod, payment, invoice, featureFlags.UseBulkAddValidatePh2, param.ReceiptDate)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("file validation failed: %v", err.Error())
		}

		if featureFlags.UseBulkAddValidatePh2 && payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
			payment.Amount = database.Numeric(0)
		}

		totalAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error converting invoice total amount at line %v: %v", lineNo+1, err.Error())
		}

		// Create action log data
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:                invoice.InvoiceID.String,
			PaymentSequenceNumber:    payment.PaymentSequenceNumber.Int,
			Action:                   invoice_pb.InvoiceAction_NO_ACTION, // Ensures no action log will be created if not modified
			BulkPaymentValidationsID: param.BulkPaymentValidationID,
		}

		switch invoice.Status.String {
		case invoice_pb.InvoiceStatus_PAID.String():
			statusCount.SuccessfulPayments++
			actionDetails.Action = invoice_pb.InvoiceAction_INVOICE_PAID
		case invoice_pb.InvoiceStatus_ISSUED.String():
			if featureFlags.UseBulkAddValidatePh2 && payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
				statusCount.FailedPayments++
			} else {
				statusCount.PendingPayments++
			}
		case invoice_pb.InvoiceStatus_FAILED.String():
			statusCount.FailedPayments++
			actionDetails.Action = invoice_pb.InvoiceAction_INVOICE_FAILED
		case invoice_pb.InvoiceStatus_VOID.String():
			statusCount.FailedPayments++
		}

		// If there is no changes in invoice and payment status, set the action log to payment updated
		if previousInvoiceStatus == invoice.Status.String && previousPaymentStatus == payment.PaymentStatus.String {
			actionDetails.Action = invoice_pb.InvoiceAction_PAYMENT_UPDATED
		}

		// If payment status is failed after validating
		switch featureFlags.UseBulkAddValidatePh2 {
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

			// Generate action log V2
			if actionDetails.Action != invoice_pb.InvoiceAction_NO_ACTION {
				actionLogEntity, err := utils.GenActionLogEntity(ctx, actionDetails)
				if err != nil {
					return nil, nil, nil, err
				}
				actionLogToCreate = append(actionLogToCreate, actionLogEntity)
			}

		default:
			// Generate action log V1
			if actionDetails.Action != invoice_pb.InvoiceAction_NO_ACTION {
				actionLogEntity, err := utils.GenActionLogEntityV1(ctx, actionDetails)
				if err != nil {
					return nil, nil, nil, err
				}
				actionLogToCreate = append(actionLogToCreate, actionLogEntity)
			}
		}

		// Create bulk validation details record
		bulkPaymentValidationDtl := new(entities.BulkPaymentValidationsDetail)
		database.AllNullEntity(bulkPaymentValidationDtl)

		bulkPaymentValidationDtl.BulkPaymentValidationsID = database.Text(param.BulkPaymentValidationID)
		bulkPaymentValidationDtl.InvoiceID = database.Text(invoice.InvoiceID.String)
		bulkPaymentValidationDtl.PaymentID = database.Text(payment.PaymentID.String)
		bulkPaymentValidationDtl.ValidatedResultCode = database.Text(validationResult.ResultCode)
		bulkPaymentValidationDtl.PreviousResultCode = database.Text(previousResultCode)
		bulkPaymentValidationDtl.PaymentStatus = database.Text(payment.PaymentStatus.String)

		invoiceToUpdate[i] = invoice
		paymentToUpdate[i] = payment
		validationDetailToCreate[i] = bulkPaymentValidationDtl

		validatedPayment := &ValidatedPayment{
			PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
			ResultCode:            validationResult.ResultCode,
			Amount:                totalAmount,
			PaymentMethod:         param.PaymentMethod,
			InvoiceSequenceNumber: invoice.InvoiceSequenceNumber.Int,
			StudentID:             invoice.StudentID.String,
			StudentName:           paymentInvoice.UserBasicInfo.Name.String,
			PaymentCreatedDate:    payment.CreatedAt.Time,
			InvoiceID:             invoice.InvoiceID.String,
			PaymentStatus:         payment.PaymentStatus.String,
		}

		validatedPayments = append(validatedPayments, validatedPayment)
	}

	entityList := &entityLists{
		InvoiceToUpdate:                 invoiceToUpdate,
		PaymentToUpdate:                 paymentToUpdate,
		InvoiceActionLogToCreate:        actionLogToCreate,
		PaymentValidationDetailToCreate: validationDetailToCreate,
	}

	return entityList, statusCount, validatedPayments, nil
}
