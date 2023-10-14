package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDescriptionWithPrc(t *testing.T) {
	t.Parallel()
	prc := -8
	result := GetDescriptionWithPrc(int64(prc))
	expectedResult := "Rejected due to PayDollar Internal/Fraud Prevention Checking"
	assert.Equal(t, result, expectedResult)
}
