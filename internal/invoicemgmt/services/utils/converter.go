package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetFloat64ExactValueAndDecimalPlaces(amount pgtype.Numeric, decimal string) (float64, error) {
	var floatAmount float64
	err := amount.AssignTo(&floatAmount)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, err.Error())
	}

	getExactValueWithDecimalPlaces, err := strconv.ParseFloat(fmt.Sprintf("%."+decimal+"f", floatAmount), 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, err.Error())
	}

	return getExactValueWithDecimalPlaces, nil
}

// FormatCurrency formats float to string in a currency format
// Note that this can only support 2 decimal places. If we need to support decimal places, we can still improve the function
func FormatCurrency(value float64) string {
	// Convert the float to a string
	amountStr := strconv.FormatFloat(value, 'f', -1, 64)

	// Check if there is a fractional part
	hasFraction := strings.Contains(amountStr, ".")

	// Split the string into the integer and fractional parts
	parts := strings.Split(amountStr, ".")
	wholeNumber := parts[0]

	// Check if the whole number is negative
	var isNegative bool
	if wholeNumber[0] == '-' {
		isNegative = true
		wholeNumber = wholeNumber[1:]
	}

	// Add commas to the integer part
	integerLen := len(wholeNumber)
	for i := integerLen - 3; i > 0; i -= 3 {
		wholeNumber = wholeNumber[:i] + "," + wholeNumber[i:]
	}

	// If negative, add the negative sign on the start of the value
	if isNegative {
		wholeNumber = fmt.Sprintf("-%s", wholeNumber)
	}

	// If there is a fractional part, add it to the formatted string
	if hasFraction {
		fracPart := parts[1]
		return wholeNumber + "." + fracPart
	}

	return wholeNumber
}
