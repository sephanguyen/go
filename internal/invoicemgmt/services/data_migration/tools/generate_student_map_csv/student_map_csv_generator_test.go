package generator

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"testing"
	"time"

	services "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration"
	helper "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration/tools"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
)

// nolint:unused,structcheck
type TestCase struct {
	name            string
	ctx             context.Context
	entityName      string
	generatedCSVDir string
	expectedErr     error
	setup           func(ctx context.Context)
}

func TestStudentMapCSVGenerator_GenerateStudentMapCSV(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	baseDir, err := createTempBaseDir()
	if err != nil {
		t.Error(err)
	}

	// create user csv
	userFileCsv, userFilePath, err := createData(3, UserEntity, baseDir)

	// create invoice csv
	invoiceFileCsv, invoiceFilePath, err := createData(3, InvoiceEntity, baseDir)

	paymentFileCsv, paymentFilePath, err := createData(3, PaymentEntity, baseDir)

	defer userFileCsv.Close()
	defer invoiceFileCsv.Close()
	defer paymentFileCsv.Close()

	// remove all folders subfolders and files contains in base directory
	defer os.RemoveAll(baseDir)

	baseDirForTest := baseDir + "/"

	testCases := []TestCase{
		{
			name:            "happy case invoice entity",
			ctx:             ctx,
			entityName:      InvoiceEntity,
			generatedCSVDir: baseDirForTest,
			setup: func(ctx context.Context) {

			},
		},
		{
			name:            "happy case payment entity",
			ctx:             ctx,
			entityName:      PaymentEntity,
			generatedCSVDir: baseDirForTest,
			setup: func(ctx context.Context) {

			},
		},
		{
			name:            "no user mapping csv existing",
			ctx:             ctx,
			entityName:      InvoiceEntity,
			generatedCSVDir: baseDir,
			expectedErr:     fmt.Errorf("user_mapping_id.csv: no such file or directory"),
			setup: func(ctx context.Context) {
				os.Remove(userFilePath)
			},
		},
		{
			name:            "no invoice csv existing",
			ctx:             ctx,
			entityName:      InvoiceEntity,
			generatedCSVDir: baseDirForTest,
			expectedErr:     fmt.Errorf("no existing csv for entity name: INVOICE_ENTITY"),
			setup: func(ctx context.Context) {
				userFileCsvTest, _, _ := createData(3, UserEntity, baseDir)
				defer userFileCsvTest.Close()

				os.Remove(invoiceFilePath)
			},
		},
		{
			name:            "no payment csv existing",
			ctx:             ctx,
			entityName:      PaymentEntity,
			generatedCSVDir: baseDirForTest,
			expectedErr:     fmt.Errorf("no existing csv for entity name: PAYMENT_ENTITY"),
			setup: func(ctx context.Context) {
				os.Remove(paymentFilePath)
			},
		},
		{
			name:            "no data invoice csv",
			ctx:             ctx,
			entityName:      InvoiceEntity,
			generatedCSVDir: baseDirForTest,
			expectedErr:     fmt.Errorf("no data in CSV file: invoice_test.csv"),
			setup: func(ctx context.Context) {
				invoiceFileCsvTest, _, _ := createData(0, InvoiceEntity, baseDir)
				defer invoiceFileCsvTest.Close()
			},
		},
		{
			name:            "no data payment csv",
			ctx:             ctx,
			entityName:      PaymentEntity,
			generatedCSVDir: baseDirForTest,
			expectedErr:     fmt.Errorf("no data in CSV file: payment_test.csv"),
			setup: func(ctx context.Context) {
				paymentFileCsvTest, _, _ := createData(0, PaymentEntity, baseDir)
				defer paymentFileCsvTest.Close()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			g := NewStudentMapCSVGenerator(testCase.entityName, testCase.generatedCSVDir)
			err := g.GenerateStudentMapCsv(testCase.ctx)

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}
		})
	}
}
func createTempBaseDir() (string, error) {
	dname, err := os.MkdirTemp("", "sampledir")

	return dname, err
}

func createFile(tempDir, objectName string) (*os.File, string, error) {
	objectPath := fmt.Sprintf("%v/%v", tempDir, objectName)
	file, err := os.Create(objectPath)
	if err != nil {
		return nil, "", fmt.Errorf("os.Create err: %v", err)
	}

	return file, objectPath, nil
}

func createData(count int, entityName, dir string) (*os.File, string, error) {
	var testFilename string

	switch entityName {
	case invoice_pb.DataMigrationEntityName_INVOICE_ENTITY.String():
		err := os.MkdirAll(dir+"/invoice_csv/mapped_invoice_csv", os.ModePerm)
		if err != nil {
			return nil, "", err
		}
		dir += "/invoice_csv"

		testFilename = "invoice_test.csv"
	case invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String():
		err := os.MkdirAll(dir+"/payment_csv/mapped_payment_csv", os.ModePerm)
		if err != nil {
			return nil, "", err
		}
		dir += "/payment_csv"
		testFilename = "payment_test.csv"
	case "USER_ENTITY":
		testFilename = "user_mapping_id.csv"
	}

	file, objectPath, err := createFile(dir, testFilename)
	if err != nil {
		return nil, "", err
	}

	data := genTestData(count, entityName)

	// add the expected data to the file
	w := csv.NewWriter(file)
	err = w.WriteAll(data)
	if err != nil {
		return nil, "", err
	}

	w.Flush()
	if err != nil {
		return nil, "", err
	}

	return file, objectPath, nil
}

func genTestData(count int, entityName string) [][]string {
	header := helper.GetHeaderTitles(entityName)
	testData := [][]string{
		header,
	}

	for i := 0; i < count; i++ {
		var line []string

		switch entityName {
		case invoice_pb.DataMigrationEntityName_INVOICE_ENTITY.String():
			line = genInvoiceTestData(len(header), i)
		case invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String():
			line = genPaymentTestData(len(header), i)
		case "USER_ENTITY":
			line = genUserTestData(len(header), i)
		}

		testData = append(testData, line)
	}

	return testData
}

func genUserTestData(dataLength, index int) []string {
	line := make([]string, dataLength)
	line[UserID] = fmt.Sprintf("student-id-%v", index+1)
	line[UserExternalID] = fmt.Sprintf("external-student-id-%v", index+1)
	line[UserResourcePath] = "-2147483642"
	return line
}

func genInvoiceTestData(dataLength, index int) []string {
	line := make([]string, dataLength)
	line[services.InvoiceCsvID] = fmt.Sprintf("%v", index+1)
	line[services.InvoiceStudentIDReference] = fmt.Sprintf("external-student-id-%v", index+1)
	line[services.InvoiceType] = fmt.Sprintf("test-type-%v", index+1)
	line[services.InvoiceStatus] = fmt.Sprintf("test-status-%v", index+1)
	line[services.InvoiceSubTotal] = "500"
	line[services.InvoiceTotal] = "500"
	line[services.InvoiceCreatedAt] = "2002-03-01"
	line[services.InvoiceReference1] = fmt.Sprintf("invoice-ref-%v", index+1)
	return line
}

func genPaymentTestData(dataLength, index int) []string {
	line := make([]string, dataLength)
	line[services.PaymentCSVID] = fmt.Sprintf("%v", index+1)
	line[services.PaymentMethod] = fmt.Sprintf("test-payment-method-%v", index+1)
	line[services.PaymentStatus] = fmt.Sprintf("test-status-%v", index+1)
	line[services.PaymentDueDate] = "2002-03-01"
	line[services.PaymentExpiryDate] = "2002-03-01"
	line[services.PaymentDate] = "2002-03-01"
	line[services.PaymentStudentID] = fmt.Sprintf("external-student-id-%v", index+1)
	line[services.PaymentInvoiceReference] = fmt.Sprintf("invoice-ref-%v", index+1)
	return line
}
