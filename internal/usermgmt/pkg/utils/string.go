package utils

import "strings"

func LowerCaseFirstLetter(text string) string {
	if len(text) < 1 {
		return text
	}
	if len(text) == 1 {
		return strings.ToLower(text)
	}
	return strings.ToLower(string(text[0])) + text[1:]
}

func UpperCaseFirstLetter(text string) string {
	if len(text) < 1 {
		return text
	}
	if len(text) == 1 {
		return strings.ToUpper(text)
	}
	return strings.ToUpper(string(text[0])) + text[1:]
}
