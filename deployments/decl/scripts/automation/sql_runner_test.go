package automation

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSQLFilesIn(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "sql_runner_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	tmpfile1, err := os.CreateTemp(tmpdir, "0.init*.sql")
	require.NoError(t, err)
	tmpfile2, err := os.CreateTemp(tmpdir, "123456.some_db_name*.sql")
	require.NoError(t, err)
	_, err = os.CreateTemp(tmpdir, "invalid*.sql")
	require.NoError(t, err)
	_, err = os.CreateTemp(tmpdir, "0.notsql*.yaml")
	require.NoError(t, err)

	r := SQLRunner{}
	filelist, err := r.getSQLFilesIn(tmpdir)
	require.NoError(t, err)
	expectedFilelist := []string{
		path.Base(tmpfile1.Name()),
		path.Base(tmpfile2.Name()),
	}
	assert.ElementsMatch(t, filelist, expectedFilelist)
}
