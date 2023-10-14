package utils

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/manabie-com/backend/internal/payment/constant"
)

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

func CheckOutVersion(expectedVersionNumber int32, versionNumber int32) (err error) {
	if expectedVersionNumber != versionNumber {
		err = StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.OptimisticLockingEntityVersionMismatched,
			&errdetails.DebugInfo{Detail: fmt.Sprintf("Expected version = %d vs actual version = %d", expectedVersionNumber, versionNumber)},
		)

		return
	}
	return
}
