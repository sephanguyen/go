package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MissingMandatoryData = "missing mandatory data: %v"
	MigrationDateFormat  = "2006-01-02"
	MigrationEmail       = "@student.kec.gr.jp"
)

var (
	InvoiceStatusStructMap = map[string]struct{}{
		invoice_pb.InvoiceStatus_DRAFT.String():    {},
		invoice_pb.InvoiceStatus_ISSUED.String():   {},
		invoice_pb.InvoiceStatus_VOID.String():     {},
		invoice_pb.InvoiceStatus_FAILED.String():   {},
		invoice_pb.InvoiceStatus_REFUNDED.String(): {},
		invoice_pb.InvoiceStatus_PAID.String():     {},
	}
	InvoiceTypeStructMap = map[string]struct{}{
		invoice_pb.InvoiceType_MANUAL.String():    {},
		invoice_pb.InvoiceType_SCHEDULED.String(): {},
	}
)

// constant for Payment migration data columns
const (
	PaymentCSVID = iota
	PaymentID
	PaymentInvoiceID
	PaymentMethod
	PaymentStatus
	PaymentDueDate
	PaymentExpiryDate
	PaymentDate
	PaymentStudentID
	PaymentSequenceNumber
	PaymentIsExported
	PaymentCreatedAt
	ResultCode
	Amount
	PaymentInvoiceReference
)

// constant for Invoice migration data columns
const (
	InvoiceCsvID = iota
	InvoiceInvoiceID
	InvoiceStudentIDReference
	InvoiceType
	InvoiceStatus
	InvoiceSubTotal
	InvoiceTotal
	InvoiceCreatedAt
	InvoiceInvoiceSequenceNumber
	InvoiceIsExported
	InvoiceReference1
	InvoiceReference2
)

type SetFunc func(interface{}) error

func checkMandatoryColumnAndGetIndex(column []string, positions []int, entityName string) error {
	headerTitles, err := getHeaderTitles(entityName)

	if err != nil {
		return err
	}
	for _, position := range positions {
		if strings.TrimSpace(column[position]) == "" {
			return fmt.Errorf("missing mandatory data: %v", headerTitles[position])
		}
	}
	return nil
}

func getHeaderTitles(entityNameStr string) ([]string, error) {
	var headerTitles []string
	switch entityNameStr {
	case invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String():
		headerTitles = []string{
			"payment_csv_id",
			"payment_id",
			"invoice_id",
			"payment_method",
			"payment_status",
			"due_date",
			"expiry_date",
			"payment_date",
			"student_id",
			"payment_sequence_number",
			"is_exported",
			"created_at",
			"result_code",
			"amount",
			"reference",
		}
	case invoice_pb.DataMigrationEntityName_INVOICE_ENTITY.String():
		headerTitles = []string{
			"invoice_csv_id",
			"invoice_id",
			"student_id",
			"type",
			"status",
			"sub_total",
			"total",
			"created_at",
			"invoice_sequence_number",
			"is_exported",
			"reference1",
			"reference2",
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "entity name not supported")
	}

	return headerTitles, nil
}

func StringToFormatString(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	return setter(trimmedValue)
}

func StringToDate(title, value, country string, nullable bool, setter SetFunc) error {
	var (
		timeElement time.Time
		err         error
	)
	trimmedValue := strings.TrimSpace(value)

	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}

	location, err := utils.GetTimeLocationByCountry(country)
	if err != nil {
		return fmt.Errorf("error: %v getTimeLocationByCountry from date column: %v", err, title)
	}

	timeElement, err = time.ParseInLocation(MigrationDateFormat, trimmedValue, location)
	if err != nil {
		return fmt.Errorf("error parsing string to date %v: %w", title, err)
	}

	return setter(timeElement)
}

func StringToBool(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	boolValue, err := strconv.ParseBool(trimmedValue)
	if err != nil {
		return fmt.Errorf("error parsing string to bool %v: %w", title, err)
	}
	return setter(boolValue)
}

func StringToFloat64(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}

	float64Value, err := strconv.ParseFloat(trimmedValue, 64)
	if err != nil {
		return fmt.Errorf("error parsing string to float64 %v: %w", title, err)
	}
	return setter(float64Value)
}

func validateHeaderColumnRequest(req *invoice_pb.ImportDataMigrationRequest) ([][]string, error) {
	var lines [][]string

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return lines, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return lines, status.Error(codes.InvalidArgument, "no data in CSV file")
	}
	header := lines[0]
	headerTitles, err := getHeaderTitles(req.EntityName.String())
	if err != nil {
		return lines, err
	}

	if err = ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	); err != nil {
		return lines, status.Error(codes.InvalidArgument, fmt.Sprintf("%s - %s", req.EntityName.String(), err.Error()))
	}

	return lines, nil
}

func ValidateCsvHeader(expectedNumberColumns int, columnNames, expectedColumnNames []string) error {
	if len(columnNames) != expectedNumberColumns {
		return fmt.Errorf("csv file invalid format - number of column should be %d", expectedNumberColumns)
	}

	for idx, expectedColumnName := range expectedColumnNames {
		if !strings.EqualFold(columnNames[idx], expectedColumnName) {
			return fmt.Errorf("csv file invalid format - %s column (toLowerCase) should be '%s'", columnNames[idx], expectedColumnName)
		}
	}
	return nil
}
