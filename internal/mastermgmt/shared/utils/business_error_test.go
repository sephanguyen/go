package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBusinessError_Is(t *testing.T) {
	t.Parallel()

	name := "sample name"
	t.Run("should return true with the same name of error", func(t *testing.T) {
		// arrange
		err := &BusinessError{
			Name:  "sample name",
			Error: fmt.Errorf("%s", "err detail"),
		}
		// act
		res := err.Is(name)

		// assert
		assert.Equal(t, true, res)
	})

	t.Run("should return false with the different name of error", func(t *testing.T) {
		// arrange
		err := &BusinessError{
			Name:  "sample name 2",
			Error: fmt.Errorf("%s", "err detail 2"),
		}
		// act
		res := err.Is(name)

		// assert
		assert.Equal(t, false, res)
	})
}

func TestBusinessError_NewError(t *testing.T) {
	t.Parallel()

	t.Run("should return the same error", func(t *testing.T) {
		// arrange
		err := fmt.Errorf("%s", "test err")

		// act
		bErr := NewError("ErrCode1", err)

		// assert
		assert.Equal(t, err, bErr.Error)
		assert.Equal(t, "ErrCode1", bErr.Name)
	})
}

func TestBusinessError_NewSystemError(t *testing.T) {
	t.Parallel()

	t.Run("should return the same error", func(t *testing.T) {
		// arrange
		err := fmt.Errorf("%s", "test err")

		// act
		bErr := NewSystemError(err)

		// assert
		assert.Equal(t, err, bErr.Error)
		assert.Equal(t, SystemError, bErr.Name)
	})
}
