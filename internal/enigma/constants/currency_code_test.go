package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrency(t *testing.T) {
	t.Parallel()
	cur := "704"
	expectedResult := "VND"
	result := GetCurrency(cur)
	assert.Equal(t, result, expectedResult)
}
