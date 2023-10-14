package main

import (
	"context"
	"fmt"
	"os"

	generator "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration/tools/generate_invoicemgmt_csv"
)

const maxRowPerFile = 20000

func main() {
	ctx := context.Background()

	g := generator.NewInvoiceMgmtCSVGenerator(getRawDataPath(), generatedCSVDir(), maxRowPerFile)
	err := g.GenerateInvoiceAndPaymentCSV(ctx)
	if err != nil {
		panic(err)
	}
}

// Change the values here depending on the file name in your local
func getRawDataPath() string {
	home, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/raw-data/t_invoices_v3.csv", home)
}

// Change the values here depending on the directory name in your local
func generatedCSVDir() string {
	return "./generated_csv"
}
