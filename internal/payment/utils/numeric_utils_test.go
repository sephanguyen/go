package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareAmountValue(t *testing.T) {
	t.Run("2 number is equal", func(t *testing.T) {
		isEqual := CompareAmountValue(12.3456789, 12.3456789010)
		assert.True(t, isEqual)
	})
	t.Run("2 number is equal with float less than 4", func(t *testing.T) {
		isEqual := CompareAmountValue(12.345, 12.345)
		assert.True(t, isEqual)
	})
	t.Run("2 number is not equal with float equal 4", func(t *testing.T) {
		isEqual := CompareAmountValue(12.349999999, 12.35)
		assert.True(t, isEqual)
	})

	t.Run("2 number with zero value is equal", func(t *testing.T) {
		isEqual := CompareAmountValue(-0, 0)
		assert.True(t, isEqual)
	})

	t.Run("test real data", func(t *testing.T) {
		tmpTaxAmount := 1033.3333333333333 * float32(10) / float32(100+10)
		isEqual := CompareAmountValue(tmpTaxAmount, 93.93939393939392)
		assert.True(t, isEqual)
	})
}
