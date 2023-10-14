package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePlaceholdersForBulkUpsert(t *testing.T) {
	t.Parallel()

	t.Run("success returns empty string", func(t *testing.T) {
		t.Parallel()
		cases := []int{0, -1, -123456}
		for _, n := range cases {
			actual := GeneratePlaceHolderForBulkUpsert(2, n)
			assert.Empty(t, actual)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		actual := GeneratePlaceHolderForBulkUpsert(1, 4)
		assert.Exactly(t, "($1, $2, $3, $4)", actual)

		actual = GeneratePlaceHolderForBulkUpsert(2, 1)
		assert.Exactly(t, "($1), ($2)", actual)

		actual = GeneratePlaceHolderForBulkUpsert(1, 3)
		assert.Exactly(t, "($1, $2, $3)", actual)
	})
}