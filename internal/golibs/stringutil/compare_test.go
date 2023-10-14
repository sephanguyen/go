package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceEqual(t *testing.T) {
	t.Parallel()
	assert.True(t, SliceEqual(nil, nil))
	assert.True(t, SliceEqual(nil, []string{}))
	assert.True(t, SliceEqual([]string{"a", "b"}, []string{"a", "b"}))
	assert.False(t, SliceEqual(nil, []string{"a"}))
	assert.False(t, SliceEqual([]string{"a"}, []string{"a", "b"}))
	assert.False(t, SliceEqual([]string{"a", "b"}, []string{"a"}))
	assert.False(t, SliceEqual([]string{"a", "b"}, []string{"b", "a"}))
	assert.False(t, SliceEqual([]string{"a", "b"}, []string{"a", "c"}))
}

func TestSliceDiff(t *testing.T) {
	t.Parallel()
	var nilSl []string
	assert.Equal(t, nilSl, SliceElementsDiff(nil, nil))
	assert.Equal(t, []string{"c"}, SliceElementsDiff([]string{"a", "b", "c"}, []string{"a", "b"}))
	assert.Equal(t, []string{"c"}, SliceElementsDiff([]string{"a", "b", "c"}, []string{"a", "b", "d"}))
	assert.Equal(t, []string{"a", "b", "c"}, SliceElementsDiff([]string{"a", "b", "c"}, []string{}))
	assert.Equal(t, []string{"a", "b", "c"}, SliceElementsDiff([]string{"a", "b", "c"}, nil))
}

func TestSliceElementsMatch(t *testing.T) {
	t.Parallel()
	assert.True(t, SliceElementsMatch(nil, nil))
	assert.True(t, SliceElementsMatch(nil, []string{}))
	assert.True(t, SliceElementsMatch([]string{"a", "b"}, []string{"a", "b"}))
	assert.True(t, SliceElementsMatch([]string{"a", "b", "c"}, []string{"b", "c", "a"}))
	assert.False(t, SliceElementsMatch(nil, []string{"a"}))
	assert.False(t, SliceElementsMatch([]string{"a"}, []string{"a", "b"}))
	assert.False(t, SliceElementsMatch([]string{"a", "b"}, []string{"a"}))
	assert.False(t, SliceElementsMatch([]string{"a", "b"}, []string{"a", "c"}))
}
