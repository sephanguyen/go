package paymentfileutils

import invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

type DataRecordCreatedDateLine struct {
	LineNo                        int
	CreatedDate                   int
	LineNoWithGreaterCreatedDate  int
	LastLineNoWithSameCreatedDate int
	GreaterCreatedDateFound       bool // to check if there's a created date greater than the previous csv line for the same payment (top to bottom)
	DuplicateCreatedDateFound     bool // to check whether same payment record has same created date
}

const (
	ConvenienceStore = iota
	DirectDebit
)

const (
	FileRCodeDDAlreadyTransferred = "0"
	FileRCodeDDShortCash          = "1"
	FileRCodeDDAcctNonExisting    = "2"
	FileRCodeDDStopCustReason     = "3"
	FileRCodeDDNoContract         = "4"
	FileRCodeDDStopConsignReason  = "8"
	FileRCodeDDOthers             = "9"

	SysRCodeDDAmtNotMatched          = 1
	SysRCodeDDNotIssued              = 2
	SysRCodeDDAmtNotMatchedNotIssued = 3

	FileRCodeCCPaidTransferred    = "02"
	FileRCodeCCPaidNotTransferred = "01"
	FileRCodeCCRevokedCancelled   = "03"

	SysRCodeCCAmtNotMatched          = 1
	SysRCodeCCNotIssued              = 2
	SysRCodeCCAmtNotMatchedNotIssued = 3
)

var (
	// DD Result Codes
	fileResultCodeDDMapping = map[string]string{
		FileRCodeDDAlreadyTransferred: FileRCodeDDAlreadyTransferred,
		FileRCodeDDShortCash:          FileRCodeDDShortCash,
		FileRCodeDDAcctNonExisting:    FileRCodeDDAcctNonExisting,
		FileRCodeDDStopCustReason:     FileRCodeDDStopCustReason,
		FileRCodeDDNoContract:         FileRCodeDDNoContract,
		FileRCodeDDStopConsignReason:  FileRCodeDDStopConsignReason,
		FileRCodeDDOthers:             FileRCodeDDOthers,
	}

	fileResultCodeDDInvoiceStatusMapping = map[string]invoice_pb.InvoiceStatus{
		FileRCodeDDAlreadyTransferred: invoice_pb.InvoiceStatus_PAID,
		FileRCodeDDShortCash:          invoice_pb.InvoiceStatus_FAILED,
		FileRCodeDDAcctNonExisting:    invoice_pb.InvoiceStatus_FAILED,
		FileRCodeDDStopCustReason:     invoice_pb.InvoiceStatus_FAILED,
		FileRCodeDDNoContract:         invoice_pb.InvoiceStatus_FAILED,
		FileRCodeDDStopConsignReason:  invoice_pb.InvoiceStatus_FAILED,
		FileRCodeDDOthers:             invoice_pb.InvoiceStatus_FAILED,
	}

	systemResultCodeDDInvoiceStatusMapping = map[int]invoice_pb.InvoiceStatus{
		SysRCodeDDAmtNotMatched:          invoice_pb.InvoiceStatus_FAILED,
		SysRCodeDDNotIssued:              invoice_pb.InvoiceStatus_FAILED,
		SysRCodeDDAmtNotMatchedNotIssued: invoice_pb.InvoiceStatus_FAILED,
	}

	fileResultCodeDDPaymentStatusMapping = map[string]invoice_pb.PaymentStatus{
		FileRCodeDDAlreadyTransferred: invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
		FileRCodeDDShortCash:          invoice_pb.PaymentStatus_PAYMENT_FAILED,
		FileRCodeDDAcctNonExisting:    invoice_pb.PaymentStatus_PAYMENT_FAILED,
		FileRCodeDDStopCustReason:     invoice_pb.PaymentStatus_PAYMENT_FAILED,
		FileRCodeDDNoContract:         invoice_pb.PaymentStatus_PAYMENT_FAILED,
		FileRCodeDDStopConsignReason:  invoice_pb.PaymentStatus_PAYMENT_FAILED,
		FileRCodeDDOthers:             invoice_pb.PaymentStatus_PAYMENT_FAILED,
	}

	systemResultCodeDDPaymentStatusMapping = map[int]invoice_pb.PaymentStatus{
		SysRCodeDDAmtNotMatched:          invoice_pb.PaymentStatus_PAYMENT_FAILED,
		SysRCodeDDNotIssued:              invoice_pb.PaymentStatus_PAYMENT_FAILED,
		SysRCodeDDAmtNotMatchedNotIssued: invoice_pb.PaymentStatus_PAYMENT_FAILED,
	}

	// CC Result Codes
	fileResultCodeCCMapping = map[string]string{
		FileRCodeCCPaidTransferred:    "0",
		FileRCodeCCPaidNotTransferred: "1",
		FileRCodeCCRevokedCancelled:   "2",
	}

	fileResultCodeCCInvoiceStatusMapping = map[string]invoice_pb.InvoiceStatus{
		FileRCodeCCPaidTransferred:    invoice_pb.InvoiceStatus_PAID,
		FileRCodeCCPaidNotTransferred: invoice_pb.InvoiceStatus_ISSUED,
		FileRCodeCCRevokedCancelled:   invoice_pb.InvoiceStatus_ISSUED,
	}

	systemResultCodeCCInvoiceStatusMapping = map[int]invoice_pb.InvoiceStatus{
		SysRCodeCCAmtNotMatched:          invoice_pb.InvoiceStatus_FAILED,
		SysRCodeCCNotIssued:              invoice_pb.InvoiceStatus_FAILED,
		SysRCodeCCAmtNotMatchedNotIssued: invoice_pb.InvoiceStatus_FAILED,
	}

	fileResultCodeCCPaymentStatusMapping = map[string]invoice_pb.PaymentStatus{
		FileRCodeCCPaidTransferred:    invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL,
		FileRCodeCCPaidNotTransferred: invoice_pb.PaymentStatus_PAYMENT_PENDING,
		FileRCodeCCRevokedCancelled:   invoice_pb.PaymentStatus_PAYMENT_PENDING,
	}

	systemResultCodeCCPaymentStatusMapping = map[int]invoice_pb.PaymentStatus{
		SysRCodeCCAmtNotMatched:          invoice_pb.PaymentStatus_PAYMENT_FAILED,
		SysRCodeCCNotIssued:              invoice_pb.PaymentStatus_PAYMENT_FAILED,
		SysRCodeCCAmtNotMatchedNotIssued: invoice_pb.PaymentStatus_PAYMENT_FAILED,
	}

	prefixCodePaymentMethodMapping = map[int]string{
		DirectDebit:      "D",
		ConvenienceStore: "C",
	}

	// Phase 2 Result Codes
	fileResultCodeDDInvoiceStatusMappingV2 = map[string]invoice_pb.InvoiceStatus{
		FileRCodeDDAlreadyTransferred: invoice_pb.InvoiceStatus_PAID,
		FileRCodeDDShortCash:          invoice_pb.InvoiceStatus_ISSUED,
		FileRCodeDDAcctNonExisting:    invoice_pb.InvoiceStatus_ISSUED,
		FileRCodeDDStopCustReason:     invoice_pb.InvoiceStatus_ISSUED,
		FileRCodeDDNoContract:         invoice_pb.InvoiceStatus_ISSUED,
		FileRCodeDDStopConsignReason:  invoice_pb.InvoiceStatus_ISSUED,
		FileRCodeDDOthers:             invoice_pb.InvoiceStatus_ISSUED,
	}
)

// Creates map of payment numbers and the correct line number for duplicate payments
func identifyDuplicatePaymentNumbers(dataRecords []*GenericPaymentFileRecord) map[string]*DataRecordCreatedDateLine {
	paymentNumberLineMap := make(map[string]*DataRecordCreatedDateLine)

	for i, dataRecord := range dataRecords {
		record, found := paymentNumberLineMap[dataRecord.PaymentNumber]
		// Initial map value for the payment number
		if !found {
			paymentNumberLineMap[dataRecord.PaymentNumber] = &DataRecordCreatedDateLine{
				LineNo:                    i,
				CreatedDate:               dataRecord.CreatedDate,
				GreaterCreatedDateFound:   false,
				DuplicateCreatedDateFound: false,
			}
		} else {
			// Store the last line that has greater or equal created date value than the currently stored's
			// Skip lines with created date values with less than the currently stored's
			switch {
			case dataRecord.CreatedDate > record.CreatedDate:
				// if there's a greater date found than the previous one, use it as a created date
				record.LineNoWithGreaterCreatedDate = i
				record.CreatedDate = dataRecord.CreatedDate
				record.GreaterCreatedDateFound = true
				record.DuplicateCreatedDateFound = false

				paymentNumberLineMap[dataRecord.PaymentNumber] = record
			case dataRecord.CreatedDate == record.CreatedDate:
				// if there's same created date for payment meaning it is duplicated
				record.LastLineNoWithSameCreatedDate = i
				record.CreatedDate = dataRecord.CreatedDate
				record.GreaterCreatedDateFound = false
				record.DuplicateCreatedDateFound = true

				paymentNumberLineMap[dataRecord.PaymentNumber] = record
			}
		}
	}

	return paymentNumberLineMap
}

func getPaymentMethodFromFile(file *GenericPaymentFile) invoice_pb.PaymentMethod {
	var paymentMethod invoice_pb.PaymentMethod
	switch file.PaymentMethod {
	case ConvenienceStore:
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
	case DirectDebit:
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT
	}

	return paymentMethod
}
