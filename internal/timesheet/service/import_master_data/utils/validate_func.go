package utils

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var NumberNames = [...]string{
	"first",
	"second",
	"third",
	"fourth",
	"fifth",
	"sixth",
	"seventh",
	"eighth",
	"ninth",
	"tenth",
	"eleventh",
	"twelveth",
	"thirdteenth",
	"fourteenth",
}

func ValidateCsvHeader(expectedNumberColumns int, columnNames, expectedColumnNames []string) error {
	if len(columnNames) != expectedNumberColumns {
		return status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("csv file invalid format - number of column should be %d", expectedNumberColumns),
		)
	}
	for idx, expectedColumnName := range expectedColumnNames {
		if !strings.EqualFold(columnNames[idx], expectedColumnName) {
			return status.Error(
				codes.InvalidArgument,
				fmt.Sprintf("csv file invalid format - %s column (toLowerCase) should be '%s'", NumberNames[idx], expectedColumnName),
			)
		}
	}
	return nil
}
