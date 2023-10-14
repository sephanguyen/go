package generator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

var (
	rawDataCreatedDateFormat = "2006-01-02 15:04:05"
	invoiceOutDateFormat     = "2006-01-02"

	invoiceStatusMap = map[string]string{
		"1": invoice_pb.InvoiceStatus_PAID.String(),
		"2": invoice_pb.InvoiceStatus_ISSUED.String(),
		"3": invoice_pb.InvoiceStatus_ISSUED.String(),
		"4": invoice_pb.InvoiceStatus_ISSUED.String(),
		"5": invoice_pb.InvoiceStatus_ISSUED.String(),
		"6": invoice_pb.InvoiceStatus_PAID.String(),
		"7": invoice_pb.InvoiceStatus_FAILED.String(),
		"9": invoice_pb.InvoiceStatus_FAILED.String(),
	}
)

const (
	InvoiceOutID = iota
	InvoiceOutInvoiceID
	InvoiceOutStudentID
	InvoiceOutType
	InvoiceOutStatus
	InvoiceOutSubTotal
	InvoiceOutTotal
	InvoiceOutCreatedAt
	InvoiceOutInvoiceSeq
	InvoiceOutIsExported
	InvoiceOutReference1
	InvoiceOutReference2
)

func generateInvoiceLineFromRawLine(line []string) ([]string, error) {
	invoiceLine := make([]string, 12)

	invoiceAmount, err := strconv.ParseInt(line[RawInvoiceAmount], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot convert raw invoice amount to int err: %v", err)
	}

	consumptionTaxAmount, err := strconv.ParseInt(line[RawConsumptionTaxAmount], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot convert raw consumption tax amount to int err: %v", err)
	}

	invoiceTotal := invoiceAmount + consumptionTaxAmount

	createdAtStr, err := reformatTimeString(line[RawEntryDateTime], rawDataCreatedDateFormat, invoiceOutDateFormat, "entry date time")
	if err != nil {
		return nil, err
	}

	invoiceStatus := invoiceStatusMap[line[RawStatus]]
	if invoiceStatus == invoice_pb.InvoiceStatus_PAID.String() && invoiceTotal < 0 {
		invoiceStatus = invoice_pb.InvoiceStatus_REFUNDED.String()
	}

	invoiceLine[InvoiceOutStudentID] = line[RawStudentID]
	invoiceLine[InvoiceOutType] = invoice_pb.InvoiceType_MANUAL.String()
	invoiceLine[InvoiceOutStatus] = invoiceStatus
	invoiceLine[InvoiceOutSubTotal] = strconv.FormatInt(invoiceAmount, 10)
	invoiceLine[InvoiceOutTotal] = strconv.FormatInt(invoiceTotal, 10)
	invoiceLine[InvoiceOutCreatedAt] = createdAtStr
	invoiceLine[InvoiceOutIsExported] = "TRUE"
	invoiceLine[InvoiceOutReference1] = line[RawID]
	invoiceLine[InvoiceOutReference2] = line[RawPaymentID]

	return invoiceLine, nil
}

func reformatTimeString(timeStr, oldFormat, newFormat string, name string) (string, error) {
	if strings.TrimSpace(timeStr) == "" || timeStr == "NULL" {
		return "2099-12-31", nil
	}

	newTime, err := time.Parse(oldFormat, timeStr)
	if err != nil {
		return "", fmt.Errorf("cannot parse the raw %v err: %v", name, err)
	}

	return newTime.Format(newFormat), nil
}
