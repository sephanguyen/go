package downloader

import (
	"fmt"
	"strconv"
	"time"

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

func GetTimeInJST(t time.Time) (time.Time, error) {
	timezone := "Asia/Tokyo"
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return t.In(location), nil
}
