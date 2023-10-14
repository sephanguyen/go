package utils

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	MissingMandatoryData = "missing mandatory data: %v"
	ErrorParsing         = "error parsing %v: %w"
)

type SetFunc func(interface{}) error

func StringToBool(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	boolValue, err := strconv.ParseBool(trimmedValue)
	if err != nil {
		return fmt.Errorf(ErrorParsing, title, err)
	}
	return setter(boolValue)
}

func StringToFormatString(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	return setter(trimmedValue)
}

func StringToInt(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	intElement, err := strconv.Atoi(trimmedValue)
	if err != nil {
		return fmt.Errorf(ErrorParsing, title, err)
	}
	return setter(intElement)
}
