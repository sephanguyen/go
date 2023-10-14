package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateClassCode(t *testing.T) {
	t.Parallel()
	testNum := 10
	codeLen := 20
	results := make([]string, testNum)
	assert := assert.New(t)
	for i := range results {
		results[i] = GenerateClassCode(codeLen)
		assert.Len(results[i], codeLen)

		// Check duplication
		for j := 0; j < i; j++ {
			if results[j] == results[i] {
				assert.FailNowf("GenerateClassCode generated duplicated code", "code generated: %s", results[j])
			}
		}
	}
}
