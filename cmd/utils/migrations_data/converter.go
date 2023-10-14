package migrationsdata

import "github.com/manabie-com/backend/internal/golibs/scanner"

type Converter interface {
	GetHeader() []string
	GetLineConverted(sc scanner.CSVScanner, orgID string) []string
	ValidationData(sc scanner.CSVScanner) []string
}
