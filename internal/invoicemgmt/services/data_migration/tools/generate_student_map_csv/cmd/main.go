package main

import (
	"context"
	"os"
	"strings"

	generator "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration/tools/generate_student_map_csv"
)

func main() {
	ctx := context.Background()
	if len(os.Args[1:]) != 1 {
		panic("error should input parameter for entity name: INVOICE_ENTITY or PAYMENT_ENTITY")
	}
	entityName := strings.TrimSpace(os.Args[1])

	g := generator.NewStudentMapCSVGenerator(entityName, useHomeDir())

	err := g.GenerateStudentMapCsv(ctx)
	if err != nil {
		panic(err)
	}
}

func useHomeDir() string {
	home, _ := os.UserHomeDir()

	return home + "/"
}
