package generator

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"testing"
	"time"

	helper "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration/tools"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/stretchr/testify/assert"
)

// nolint:unused,structcheck
type TestCase struct {
	name            string
	ctx             context.Context
	rawFilePath     string
	generatedCSVDir string
	expectedErr     error
	setup           func(ctx context.Context)
	maxRowPerFile   int
}

func createInvoiceFile(count int, status, receiveType string, amount int) (*utils.TempFile, error) {
	tempFileCreator := &utils.TempFileCreator{TempDirPattern: "invoicemgmt-data-migration-test-"}
	tempFile, err := tempFileCreator.CreateTempFile("test_t_invoices.csv")
	if err != nil {
		return nil, err
	}

	data := genTestRawData(count, status, receiveType, amount)

	// add the expected data to the file
	w := csv.NewWriter(tempFile.File)
	err = w.WriteAll(data)
	if err != nil {
		return nil, err
	}

	w.Flush()
	if err != nil {
		return nil, err
	}

	return tempFile, nil
}

func TestInvoiceMgmtCSVGenerator_GenerateInvoiceAndPaymentCSV(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	paidFile, err := createInvoiceFile(3, "1", "1", 1000)
	if err != nil {
		t.Error(err)
	}

	refundedFile, err := createInvoiceFile(3, "1", "3", -1000)
	if err != nil {
		t.Error(err)
	}

	issuedFile, err := createInvoiceFile(3, "3", "1", 1000)
	if err != nil {
		t.Error(err)
	}

	failedFile, err := createInvoiceFile(3, "9", "2", 1000)
	if err != nil {
		t.Error(err)
	}

	defer paidFile.CleanUp()
	defer paidFile.Close()

	defer refundedFile.CleanUp()
	defer refundedFile.Close()

	defer issuedFile.CleanUp()
	defer issuedFile.Close()

	defer failedFile.CleanUp()
	defer failedFile.Close()

	testCases := []TestCase{
		{
			name:            "happy case - paid",
			ctx:             ctx,
			rawFilePath:     paidFile.ObjectPath,
			generatedCSVDir: paidFile.TempDirName,
			maxRowPerFile:   1000,
			setup: func(ctx context.Context) {

			},
		},
		{
			name:            "happy case - refunded",
			ctx:             ctx,
			rawFilePath:     refundedFile.ObjectPath,
			generatedCSVDir: refundedFile.TempDirName,
			maxRowPerFile:   1000,
			setup: func(ctx context.Context) {

			},
		},
		{
			name:            "happy case - issued",
			ctx:             ctx,
			rawFilePath:     issuedFile.ObjectPath,
			generatedCSVDir: issuedFile.TempDirName,
			maxRowPerFile:   1000,
			setup: func(ctx context.Context) {

			},
		},
		{
			name:            "happy case - failed",
			ctx:             ctx,
			rawFilePath:     failedFile.ObjectPath,
			generatedCSVDir: failedFile.TempDirName,
			maxRowPerFile:   1000,
			setup: func(ctx context.Context) {

			},
		},
		{
			name:            "happy case with 0 max row per file",
			ctx:             ctx,
			rawFilePath:     paidFile.ObjectPath,
			generatedCSVDir: paidFile.TempDirName,
			maxRowPerFile:   0,
			setup: func(ctx context.Context) {

			},
		},
		{
			name:            "happy case with 1 max row per file",
			ctx:             ctx,
			rawFilePath:     paidFile.ObjectPath,
			generatedCSVDir: paidFile.TempDirName,
			maxRowPerFile:   1,
			setup: func(ctx context.Context) {

			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			g := NewInvoiceMgmtCSVGenerator(testCase.rawFilePath, testCase.generatedCSVDir, 10000)
			err := g.GenerateInvoiceAndPaymentCSV(testCase.ctx)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func genTestRawData(count int, status, receiveType string, amount int) [][]string {
	header := helper.GetHeaderTitles(helper.InvoiceRawData)
	testRawData := [][]string{
		header,
	}

	for i := 0; i < count; i++ {
		line := make([]string, len(header))
		line[RawID] = strconv.Itoa(i + 1)
		line[RawPaymentID] = fmt.Sprintf("payment-id-%v", i+1)
		line[RawInvoiceDate] = "20060301"
		line[RawStudentID] = fmt.Sprintf("student-id-%v", i+1)
		line[RawStatus] = status
		line[RawInvoiceAmount] = strconv.Itoa(amount)
		line[RawConsumptionTaxAmount] = "10"
		line[RawReceiveType] = receiveType
		line[RawReceiveDate] = time.Now().Format(rawReceiveDateFormat)
		line[RawPrintedLimitDate1] = time.Now().Format(rawPrintedLimitDateFormat)
		line[RawUsableLimitDate2] = time.Now().Format(rawUsableLimitDateFormat)
		line[RawEntryDateTime] = time.Now().Format(rawDataCreatedDateFormat)
		line[RawInvoiceDate] = time.Now().Format(rawInvoiceDateFormat)

		testRawData = append(testRawData, line)
	}

	return testRawData
}

func genTestRawDataWithDuplicate() [][]string {
	testRawData := [][]string{
		helper.GetHeaderTitles(helper.InvoiceRawData),
	}

	for i := 0; i < 5; i++ {

	}

	return testRawData
}
