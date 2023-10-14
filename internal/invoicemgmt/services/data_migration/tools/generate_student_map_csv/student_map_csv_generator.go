package generator

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	helper "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration/tools"
)

type StudentMapCSVGenerator struct {
	entityName  string
	baseDirPath string
}

func NewStudentMapCSVGenerator(entityName, baseDirPath string) *StudentMapCSVGenerator {
	return &StudentMapCSVGenerator{
		entityName:  entityName,
		baseDirPath: baseDirPath,
	}
}

func (g *StudentMapCSVGenerator) GenerateStudentMapCsv(ctx context.Context) error {
	err := g.validateRequestParams(g.entityName)
	if err != nil {
		return err
	}

	filesToMapDir, destinationMapDir := g.getCsvFolderPath()

	userIDsMapped, err := g.retrieveUserIDsFromCsv(g.baseDirPath)
	if err != nil {
		return err
	}

	// loop the files inside the directory
	fileInfos, err := os.ReadDir(filesToMapDir)
	if err != nil {
		return err
	}
	// get invoice/payment csvs to map
	var existingFile bool
	for _, fileInfo := range fileInfos {
		var lines [][]string

		if fileInfo.IsDir() {
			continue
		}
		existingFile = true
		lines, err := validateCsvFile(fileInfo, g.entityName, filesToMapDir)
		if err != nil {
			return err
		}
		newDataLines := mappedCsvToStudentID(lines, userIDsMapped, g.entityName)

		err = generateCsvWithStudentMapped(newDataLines, destinationMapDir+"new_"+fileInfo.Name())
		if err != nil {
			return err
		}
	}

	if !existingFile {
		return fmt.Errorf("no existing csv for entity name: %v", g.entityName)
	}

	return nil
}

func (g *StudentMapCSVGenerator) validateRequestParams(entityName string) error {
	if !EntityNameMap[entityName] {
		return fmt.Errorf("invalid entity name: %v should be: INVOICE_ENTITY or PAYMENT_ENTITY", entityName)
	}

	return nil
}

func (g *StudentMapCSVGenerator) retrieveUserIDsFromCsv(dir string) (map[string]string, error) {
	// mapped of list student ids from csv
	mapUserIDs := make(map[string]string)
	file, err := os.Open(dir + "user_mapping_id.csv")
	if err != nil {
		return mapUserIDs, err
	}
	defer file.Close()

	// read the csv
	r := csv.NewReader(file)
	lines, err := r.ReadAll()
	if err != nil {
		return mapUserIDs, err
	}

	if len(lines) < 2 {
		return mapUserIDs, fmt.Errorf("no data in CSV file of entity: %v", g.entityName)
	}

	header := lines[0]
	headerTitles, err := getHeaderTitles(UserEntity)
	if err != nil {
		return mapUserIDs, err
	}

	if err = helper.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	); err != nil {
		return mapUserIDs, err
	}
	var (
		userExternalIDCountEmpty int
		userIDCountEmpty         int
	)
	for _, line := range lines[1:] {
		if strings.TrimSpace(line[UserExternalID]) == "" {
			userExternalIDCountEmpty++
			continue
		}

		if strings.TrimSpace(line[UserID]) == "" {
			userIDCountEmpty++
			continue
		}

		mapUserIDs[strings.TrimSpace(line[UserExternalID])] = strings.TrimSpace(line[UserID])
	}

	if userExternalIDCountEmpty > 0 {
		fmt.Printf("there are %v user_external_id that is empty\n", userExternalIDCountEmpty)
	}

	if userIDCountEmpty > 0 {
		fmt.Printf("there are %v user_id that is empty\n", userIDCountEmpty)
	}

	return mapUserIDs, nil
}

func (g *StudentMapCSVGenerator) getCsvFolderPath() (string, string) {
	var (
		filesToMapDir     string
		destinationMapDir string
	)

	switch g.entityName {
	case InvoiceEntity:
		filesToMapDir = g.baseDirPath + "invoice_csv/"
		destinationMapDir = filesToMapDir + MappedInvoiceCsvDir
	case PaymentEntity:
		filesToMapDir = g.baseDirPath + "payment_csv/"
		destinationMapDir = filesToMapDir + MappedPaymentCsvDir
	}

	return filesToMapDir, destinationMapDir
}
