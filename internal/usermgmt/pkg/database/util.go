package database

import (
	"strings"
)

func GenerateUpdatePlaceholders(fields []string) string {
	var builder strings.Builder
	sep := ", "

	totalField := len(fields)
	for i, field := range fields {
		if field == "created_at" {
			continue
		}
		if i == totalField-1 {
			sep = ""
		}

		builder.WriteString(field + " = EXCLUDED." + field + sep)
	}

	return builder.String()
}
