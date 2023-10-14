package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConvertNumberToUppercaseChar(t *testing.T) {
	t.Parallel()

	t.Run("Happy all case", func(t *testing.T) {
		for i := 0; i < 26; i++ {
			char, err := ConvertNumberToUppercaseChar(i)
			assert.Nil(t, err)
			assert.Equal(t, string(rune('A'-0+i)), char)
		}
	})

	t.Run("Happy one case", func(t *testing.T) {
		char, err := ConvertNumberToUppercaseChar(0)
		assert.Nil(t, err)
		assert.Equal(t, "A", char)
	})

	t.Run("Out of range (-1)", func(t *testing.T) {
		char, err := ConvertNumberToUppercaseChar(-1)
		assert.EqualError(t, err, "cannot convert a number is less than 0 or great than 25 to one uppercase letter")
		assert.Equal(t, "", char)
	})

	t.Run("Out of range (26)", func(t *testing.T) {
		char, err := ConvertNumberToUppercaseChar(-1)
		assert.EqualError(t, err, "cannot convert a number is less than 0 or great than 25 to one uppercase letter")
		assert.Equal(t, "", char)
	})
}
