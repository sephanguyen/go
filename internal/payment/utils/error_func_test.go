package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	NotEqualValues = "a doesn't equal b: %v != %v"
)

func equalValue(a, b int32) (err error) {
	if a != b {
		err = fmt.Errorf(NotEqualValues, a, b)
	}
	return
}

func TestGroupErrorFunc(t *testing.T) {
	t.Run("All func no err", func(t *testing.T) {
		err := GroupErrorFunc(
			equalValue(1, 1),
			equalValue(2, 2),
			equalValue(3, 3),
		)
		require.Nil(t, err)
	})

	t.Run("One func have err", func(t *testing.T) {
		err := GroupErrorFunc(
			equalValue(1, 1),
			equalValue(3, 2),
			equalValue(3, 3),
		)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf(NotEqualValues, 3, 2).Error(), err.Error())
	})
	t.Run("All func have err", func(t *testing.T) {
		err := GroupErrorFunc(
			equalValue(2, 1),
			equalValue(3, 2),
			equalValue(4, 3),
		)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf(NotEqualValues, 2, 1).Error(), err.Error())
	})
}
