package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func MockSetter(_ interface{}) error {
	return nil
}

func TestStringToInt(t *testing.T) {
	t.Run("happy case with empty value", func(t *testing.T) {
		err := StringToInt("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run("fail with empty value", func(t *testing.T) {
		err := StringToInt("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("missing mandatory data: test"), err)
	})

	t.Run("happy case with value", func(t *testing.T) {
		err := StringToInt("test", "2", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run("false when convert value", func(t *testing.T) {
		err := StringToInt("test", "asa", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, `error parsing test: strconv.Atoi: parsing "asa": invalid syntax`, err.Error())
	})
}

func TestStringToBool(t *testing.T) {
	t.Run("happy case with empty value", func(t *testing.T) {
		err := StringToBool("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run("fail with empty value", func(t *testing.T) {
		err := StringToBool("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("missing mandatory data: test"), err)
	})

	t.Run("happy case with value", func(t *testing.T) {
		err := StringToBool("test", "1", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run("false when convert value", func(t *testing.T) {
		err := StringToBool("test", "asa", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, `error parsing test: strconv.ParseBool: parsing "asa": invalid syntax`, err.Error())
	})
}

func TestStringToFormatString(t *testing.T) {
	t.Run("happy case with value", func(t *testing.T) {
		err := StringToFormatString("test", "yes", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run("happy case with empty value", func(t *testing.T) {
		err := StringToFormatString("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run("fail with empty value", func(t *testing.T) {
		err := StringToFormatString("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("missing mandatory data: test"), err)
	})
}
