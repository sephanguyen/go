package generator

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"strings"

	helper "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration/tools"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	InvoiceStudentIDReference = 2
	PaymentStudentIDReference = 8
)

const (
	UserID = iota
	UserExternalID
	UserEmail
	UserResourcePath
)

const (
	InvoiceEntity       = "INVOICE_ENTITY"
	PaymentEntity       = "PAYMENT_ENTITY"
	MappedInvoiceCsvDir = "mapped_invoice_csv/"
	MappedPaymentCsvDir = "mapped_payment_csv/"
	UserEntity          = "USER_ENTITY"
)

var (
	EntityNameMap = map[string]bool{
		"INVOICE_ENTITY": true,
		"PAYMENT_ENTITY": true,
	}
)

func validateCsvFile(fileInfo fs.DirEntry, entityName, dir string) ([][]string, error) {
	file, err := os.Open(dir + fileInfo.Name())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	lines, err := r.ReadAll()
	if err != nil {
		return lines, err
	}

	if len(lines) < 2 {
		return lines, fmt.Errorf("no data in CSV file: %v", fileInfo.Name())
	}

	header := lines[0]
	headerTitles := helper.GetHeaderTitles(entityName)
	if err != nil {
		return lines, err
	}

	if err = helper.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	); err != nil {
		return lines, err
	}

	return lines, nil
}

func mappedCsvToStudentID(lines [][]string, studentIDs map[string]string, entityName string) [][]string {
	data := make([][]string, 0, len(lines))

	for i, line := range lines {
		// for header of csv just append to the new data
		if i == 0 {
			data = append(data, line)
			continue
		}

		var (
			studentIDTrim          string
			studentColumnReference int
		)

		switch entityName {
		case InvoiceEntity:
			studentColumnReference = InvoiceStudentIDReference
		case PaymentEntity:
			studentColumnReference = PaymentStudentIDReference
		}
		studentIDTrim = strings.TrimSpace(line[studentColumnReference])

		if studentIDTrim == "" {
			fmt.Printf("missing student id reference on csv row :%v\n", line[0])
		}

		studentID, ok := studentIDs[studentIDTrim]

		if !ok {
			fmt.Printf("there is no student id mapped in this external student id: %v on csv row: %v\n", line[studentColumnReference], line[0])
			// will not skip loop, should remain the external value for debugging
			studentID = studentIDTrim
		}

		line[studentColumnReference] = studentID
		data = append(data, line)
	}

	return data
}

func generateCsvWithStudentMapped(newDataLines [][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %v: ", err)
	}

	defer file.Close()

	if err != nil {
		return fmt.Errorf("failed to create file %v: ", err)
	}

	w := csv.NewWriter(file)
	defer w.Flush()

	err = w.WriteAll(newDataLines)
	if err != nil {
		return fmt.Errorf("failed to write all data in csv: %v: ", err)
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
	case "USER_ENTITY":
		headerTitles = []string{
			"user_id",
			"user_external_id",
			"email",
			"resource_path",
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "entity name not supported")
	}

	return headerTitles, nil
}
