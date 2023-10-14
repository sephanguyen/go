package utils

import (
	"math"

	"github.com/jackc/pgtype"
)

func IsEqualNumericAndFloat32(numeric pgtype.Numeric, float32Value float32) bool {
	tmpFloatValue := ConvertNumericToFloat32(numeric)
	return tmpFloatValue == float32Value
}

func ConvertNumericToFloat32(numeric pgtype.Numeric) (float32Value float32) {
	_ = numeric.AssignTo(&float32Value)
	return
}

func CompareAmountValue(valueCompare float32, valueNeedCompare float32) bool {
	result := math.Abs(float64(valueCompare) - float64(valueNeedCompare))
	return result < 0.01
}
