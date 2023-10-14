package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type SetFunc func(interface{}) error

func StringToInt(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf("missing mandatory data: %v", title)
	}
	intElement, err := strconv.Atoi(trimmedValue)
	if err != nil {
		return fmt.Errorf("error parsing %v: %w", title, err)
	}
	return setter(intElement)
}

func StringToBool(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf("missing mandatory data: %v", title)
	}
	boolValue, err := strconv.ParseBool(trimmedValue)
	if err != nil {
		return fmt.Errorf("error parsing %v: %w", title, err)
	}
	return setter(boolValue)
}

func StringToFormatString(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf("missing mandatory data: %v", title)
	}
	return setter(trimmedValue)
}
