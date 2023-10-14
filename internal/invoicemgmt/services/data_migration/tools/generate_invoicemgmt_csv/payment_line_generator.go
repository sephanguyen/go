package generator

import (
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

var (
	rawPrintedLimitDateFormat = "2006-01-02"
	rawUsableLimitDateFormat  = "2006-01-02"
	rawReceiveDateFormat      = "2006-01-02"
	rawInvoiceDateFormat      = "2006-01-02"

	paymentOutDateFormat = "2006-01-02"

	paymentMethodMap = map[string]string{
		"1": invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
		"2": invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
		"3": invoice_pb.PaymentMethod_CASH.String(),
		"4": invoice_pb.PaymentMethod_BANK_TRANSFER.String(),
	}
)

const (
	PaymentOutID = iota
	PaymentOutPaymentID
	PaymentOutInvoiceID
	PaymentOutPaymentMethod
	PaymentOutPaymentStatus
	PaymentOutDueDate
	PaymentOutExpiryDate
	PaymentOutPaymentDate
	PaymentOutStudentID
	PaymentOutPaymentSequenceNumber
	PaymentOutIsExported
	PaymentOutCreatedAt
	PaymentOutResultCode
	PaymentOutAmount
	PaymentOutReference
)

func generatePaymentLineFromRawLine(line []string, invoiceStatus string, invoiceTotal string) ([]string, error) {
	paymentLine := make([]string, 15)

	var paymentStatus string
	switch invoiceStatus {
	case invoice_pb.InvoiceStatus_ISSUED.String():
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
	case invoice_pb.InvoiceStatus_FAILED.String():
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_FAILED.String()
	case invoice_pb.InvoiceStatus_PAID.String(), invoice_pb.InvoiceStatus_REFUNDED.String():
		paymentStatus = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()
	}

	paymentDueDateStr, err := reformatTimeString(line[RawPrintedLimitDate1], rawPrintedLimitDateFormat, paymentOutDateFormat, "printed limit date")
	if err != nil {
		return nil, err
	}

	paymentExpiryDateStr, err := reformatTimeString(line[RawUsableLimitDate2], rawUsableLimitDateFormat, paymentOutDateFormat, "usable limit date")
	if err != nil {
		return nil, err
	}

	paymentDateStr, err := reformatTimeString(line[RawReceiveDate], rawReceiveDateFormat, paymentOutDateFormat, "receive date")
	if err != nil {
		return nil, err
	}

	createdDateStr, err := reformatTimeString(line[RawInvoiceDate], rawInvoiceDateFormat, paymentOutDateFormat, "invoice date")
	if err != nil {
		return nil, err
	}

	paymentLine[PaymentOutPaymentMethod] = paymentMethodMap[line[RawReceiveType]]
	paymentLine[PaymentOutPaymentStatus] = paymentStatus
	paymentLine[PaymentOutDueDate] = paymentDueDateStr
	paymentLine[PaymentOutExpiryDate] = paymentExpiryDateStr
	paymentLine[PaymentOutPaymentDate] = paymentDateStr
	paymentLine[PaymentOutStudentID] = line[RawStudentID]
	paymentLine[PaymentOutIsExported] = "TRUE"
	paymentLine[PaymentOutCreatedAt] = createdDateStr
	paymentLine[PaymentOutAmount] = invoiceTotal
	paymentLine[PaymentOutReference] = line[RawID]

	return paymentLine, nil
}
