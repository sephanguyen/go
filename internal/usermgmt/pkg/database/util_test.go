package database

import (
	"testing"

	"gotest.tools/assert"
)

func TestGenerateUpdatePlaceholders(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		expected := "user_id = EXCLUDED.user_id"

		actual := GenerateUpdatePlaceholders([]string{"user_id"})

		assert.Equal(t, actual, expected)
	})

	t.Run("happy case with created_at", func(t *testing.T) {
		expected := ""

		actual := GenerateUpdatePlaceholders([]string{"created_at"})

		assert.Equal(t, actual, expected)
	})
}
