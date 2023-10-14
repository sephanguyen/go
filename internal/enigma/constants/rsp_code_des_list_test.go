package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDescriptionWithPrcAndSrc(t *testing.T) {
	t.Parallel()
	prc := -8
	src := 1000
	expectedResult := "Skipped transaction"
	result := GetDescriptionWithPrcAndSrc(int64(prc), int64(src))
	assert.Equal(t, result, expectedResult)
}
