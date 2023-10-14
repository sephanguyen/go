package database

import "strings"

var specialChars = []string{`\`, "%", "_"}

func ReplaceSpecialChars(search string) string {
	for _, char := range specialChars {
		new := `\` + char
		search = strings.ReplaceAll(search, char, new)
	}
	return search
}
