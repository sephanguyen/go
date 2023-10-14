package execwrapper

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopLevelDir(t *testing.T) {
	dir, err := GitTopLevelDir()
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "go.mod"))
	require.FileExists(t, filepath.Join(dir, "go.sum"))
	require.DirExists(t, filepath.Join(dir, "cmd"))
	require.DirExists(t, filepath.Join(dir, "deployments"))
	require.DirExists(t, filepath.Join(dir, "internal"))
}

func TestGitDiff(t *testing.T) {
	SetGitDir(t, MockedGitDir())

	parent := "45b09733a9959d704cba0d01641ba5d400243230"
	child := "5c126559ff5a252dc51fbaac64a3dcbbf96c0844"
	diff, err := GitDiff(parent, child)
	require.NoError(t, err)
	assert.Equal(t, []string{"README.md", "helloworld.go"}, diff)

	diff, err = GitDiff(child, child)
	require.NoError(t, err)
	assert.Nil(t, diff)
}
