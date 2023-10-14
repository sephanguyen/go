package dplparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveColumnsInExludeList(t *testing.T) {
	t.Run("exclude list of columns", func(t *testing.T) {
		s := SinkConfig{}
		s.Columns = []string{"a", "b", "c", "d", "e", "f", "g"}
		s.ExcludeColumns = []string{"a", "b", "c"}
		s.RemoveColumnsInExludeList()
		assert.Equal(t, s.Columns, []string{"d", "e", "f", "g"})
	})

	t.Run("empty exclude column list", func(t *testing.T) {
		s := SinkConfig{}
		s.Columns = []string{"a", "b", "c", "d", "e", "f", "g"}
		s.ExcludeColumns = []string{}
		s.RemoveColumnsInExludeList()
		assert.Equal(t, s.Columns, []string{"a", "b", "c", "d", "e", "f", "g"})
	})

	t.Run("full exclude column list", func(t *testing.T) {
		s := SinkConfig{}
		s.Columns = []string{"a", "b", "c"}
		s.ExcludeColumns = []string{"a", "b", "c"}
		s.RemoveColumnsInExludeList()
		assert.Equal(t, s.Columns, []string{})
	})

	t.Run("exclude columns not in the list", func(t *testing.T) {
		s := SinkConfig{}
		s.Columns = []string{"a", "b", "c"}
		s.ExcludeColumns = []string{"x", "y", "z"}
		s.RemoveColumnsInExludeList()
		assert.Equal(t, s.Columns, []string{"a", "b", "c"})
	})
}
