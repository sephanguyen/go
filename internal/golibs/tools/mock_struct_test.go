package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testPaths = []string{
	"internal/foo",
	"internal/foo/",
	"internal/foo/bar",
	"internal/foo/bar/foo1/bar1/foo2/bar2/foo3/bar3",
}

func TestServiceNameFrom(t *testing.T) {
	for _, path := range testPaths {
		res := serviceNameFrom(path)
		assert.Equal(t, "foo", res)
	}
}

func TestMockPathFrom(t *testing.T) {
	expected := []string{
		"mock/foo",
		"mock/foo/",
		"mock/foo/bar",
		"mock/foo/bar/foo1/bar1/foo2/bar2/foo3/bar3",
	}
	for i, path := range testPaths {
		res := mockPathFrom(path)
		assert.Equal(t, expected[i], res)
	}
}
