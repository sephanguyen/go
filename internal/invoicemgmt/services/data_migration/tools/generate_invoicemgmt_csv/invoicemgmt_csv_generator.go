package generator

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	services "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration"
	helper "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration/tools"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

type InvoiceMgmtCSVGenerator struct {
	rawFilePath            string
	destinationDir         string
	rawInvoicePaymentIDMap map[string][][]string
	chunkSize              int
}

func NewInvoiceMgmtCSVGenerator(rawFilePath, destinationDir string, chunkSize int) *InvoiceMgmtCSVGenerator {
	return &InvoiceMgmtCSVGenerator{
		rawFilePath:            rawFilePath,
		destinationDir:         destinationDir,
		rawInvoicePaymentIDMap: make(map[string][][]string),
		chunkSize:              chunkSize,
	}
}

func (g *InvoiceMgmtCSVGenerator) GenerateInvoiceAndPaymentCSV(ctx context.Context) error {
	// Open the raw t_invoices file
	f, err := os.Open(g.rawFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read from the start of the file
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}

	// Read the raw invoice CSV
	csvReader := csv.NewReader(f)

	line, err := csvReader.Read()
	if err != nil {
		return err
	}

	// Validate the CSV header of raw data
	headerTitles := helper.GetHeaderTitles(helper.InvoiceRawData)
	err = services.ValidateCsvHeader(
		len(headerTitles),
		line,
		headerTitles,
	)
	if err != nil {
		return err
	}

	for {
		line, err := csvReader.Read()
		if err != nil && err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		// Map the invoice by payment ID
		g.mapRawDataWithPaymentID(line)
	}

	// Generate the invoice and payment CSV based on the info
	err = g.processCreatingCSV()
	if err != nil {
		return err
	}

	return nil
}

func (g *InvoiceMgmtCSVGenerator) mapRawDataWithPaymentID(line []string) {
	existingData := g.rawInvoicePaymentIDMap[line[RawPaymentID]]
	g.rawInvoicePaymentIDMap[line[RawPaymentID]] = append(existingData, line)
}

func (g *InvoiceMgmtCSVGenerator) processCreatingCSV() error {
	validLines := [][]string{}
	for _, lines := range g.rawInvoicePaymentIDMap {
		var (
			latestData      []string
			latestCreatedAt time.Time
		)

		// For each invoice data of payment ID, only get the invoice with latest created date
		for _, line := range lines {
			// Filter out rows that have null invoice date
			if strings.TrimSpace(line[RawInvoiceDate]) == "" || line[RawInvoiceDate] == "NULL" {
				continue
			}

			createdAt, err := time.Parse(rawDataCreatedDateFormat, line[RawEntryDateTime])
			if err != nil {
				return err
			}

			if createdAt.After(latestCreatedAt) {
				latestData = line
			}
		}

		// If there are data assigned, append to valid lines
		if len(latestData) != 0 {
			validLines = append(validLines, latestData)
		}
	}

	// Generate the CSV files using the valid lines
	err := g.genCSVFromRawLine(validLines)
	if err != nil {
		return err
	}

	return nil
}

func (g *InvoiceMgmtCSVGenerator) genCSVFromRawLine(rawLines [][]string) error {
	// Generate the invoice and payment CSV data lines from raw line
	invoiceLines, paymentLines, err := genInvoiceAndPaymentLines(rawLines)
	if err != nil {
		return err
	}

	// Sort the records by the t_invoices.id in ascending order
	sortSliceByIndex(invoiceLines, InvoiceOutReference1)
	sortSliceByIndex(paymentLines, PaymentOutReference)

	// Assign the CSV row ID
	assignRowIDToLines(invoiceLines, InvoiceOutID)
	assignRowIDToLines(paymentLines, PaymentOutID)

	now := time.Now().Local()

	invoiceDirName := fmt.Sprintf("%v/version-%v/invoice", g.destinationDir, now.Format("20060102-150405"))
	if err := os.MkdirAll(invoiceDirName, os.ModePerm); err != nil {
		return err
	}

	// Chunk invoice lines and create files
	chunkedInvoiceLines := chunkCSVLine(invoiceLines, g.chunkSize)
	for i, lines := range chunkedInvoiceLines {
		newLine := [][]string{helper.GetHeaderTitles(invoice_pb.DataMigrationEntityName_INVOICE_ENTITY.String())}
		newLine = append(newLine, lines...)

		err = createFileAndWrite(newLine, "invoice", i+1, invoiceDirName)
		if err != nil {
			return err
		}
	}

	paymentDirName := fmt.Sprintf("%v/version-%v/payment", g.destinationDir, now.Format("20060102-150405"))
	if err := os.MkdirAll(paymentDirName, os.ModePerm); err != nil {
		return err
	}

	// Chunk payment lines and create files
	chunkedPaymentLines := chunkCSVLine(paymentLines, g.chunkSize)
	for i, lines := range chunkedPaymentLines {
		newLine := [][]string{helper.GetHeaderTitles(invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String())}
		newLine = append(newLine, lines...)

		err = createFileAndWrite(newLine, "payment", i+1, paymentDirName)
		if err != nil {
			return err
		}
	}

	return nil
}

func createFileAndWrite(lines [][]string, entityName string, index int, dirName string) error {
	f, err := os.Create(fmt.Sprintf("%v/%v_%v.csv", dirName, entityName, index))
	if err != nil {
		return fmt.Errorf("unable to create %s CSV file err: %v", entityName, err)
	}
	defer f.Close()

	csvWriter := csv.NewWriter(f)
	err = writeLines(csvWriter, lines, entityName)
	if err != nil {
		return err
	}

	return nil
}

func chunkCSVLine(lines [][]string, chunkSize int) [][][]string {
	var chunks [][][]string

	if chunkSize <= 0 {
		return [][][]string{lines}
	}

	for i := 0; i < len(lines); i += chunkSize {
		end := i + chunkSize

		if end > len(lines) {
			end = len(lines)
		}

		chunks = append(chunks, lines[i:end])
	}

	return chunks
}

func genInvoiceAndPaymentLines(rawLines [][]string) (invoiceLines [][]string, paymentLines [][]string, err error) {
	invoiceLines = [][]string{}
	paymentLines = [][]string{}

	for _, rawLine := range rawLines {
		invoiceLine, err := generateInvoiceLineFromRawLine(rawLine)
		if err != nil {
			return nil, nil, err
		}

		paymentLine, err := generatePaymentLineFromRawLine(rawLine, invoiceLine[InvoiceOutStatus], invoiceLine[InvoiceOutTotal])
		if err != nil {
			return nil, nil, err
		}

		invoiceLines = append(invoiceLines, invoiceLine)
		paymentLines = append(paymentLines, paymentLine)
	}

	return invoiceLines, paymentLines, nil
}
