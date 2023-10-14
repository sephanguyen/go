package services

import "strings"

type ItemsBankItem struct {
	LineNumber           int
	ItemID               string
	LoID                 string
	ItemName             string
	ItemDescriptionText  string
	ItemDescriptionImage string
}

func (q *ItemsBankItem) IsItemIDValid() bool {
	// Maximum of 150 characters, case insensitive and must only contain ASCII printable characters,
	// except for double quotes, single quotes and accent.
	trimmedItemID := strings.TrimSpace(q.ItemID)
	if len(trimmedItemID) > 150 {
		return false
	}
	for _, c := range q.ItemID {
		if c < 32 || c > 126 {
			return false
		}
		if c == '"' || c == '\'' || c == '`' {
			return false
		}
	}
	return true
}
