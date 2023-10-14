package fileio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAbsolutePathFromRepoRoot(t *testing.T) {
	s, err := GetAbsolutePathFromRepoRoot("/internal/golibs/io/utils.go")
	require.NoError(t, err)

	expected, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, expected+"/utils.go", s, fmt.Errorf("%s/utils.go but got %s", expected+"/utils.go", s))

	// get repo root path
	s, err = GetAbsolutePathFromRepoRoot("")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(expected, s))
	require.FileExists(t, filepath.Join(s, "go.mod"))
}
