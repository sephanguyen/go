package tests

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
)

// TestUtilChartUpdated ensures that `make update-deps` always run for every commit.
//
// Its purpose is to ensure each backend chart have an up-to-dated libary chart from deployments/helm/libs/util.
func TestUtilChartUpdated(t *testing.T) {
	t.Parallel()

	getDirHash := func(dirpath string) (string, error) {
		return dirhash.HashDir(dirpath, "", dirhash.DefaultHash)
	}

	checkDirHash := func(t *testing.T, chartDir, sourceHash string) {
		t.Parallel()

		destDir := filepath.Join(chartDir, "templates/util")
		destHash, err := getDirHash(destDir)
		require.NoError(t, err)
		require.Equal(t, sourceHash, destHash, "dir hash mismatched, have you run `make update-deps` ?")
	}

	sourceHash, err := getDirHash(execwrapper.Abs("deployments/helm/libs/util/templates"))
	require.NoError(t, err)

	// Check for deployments/helm/backend/* charts
	rootChartDir := execwrapper.Abs("deployments/helm/backend")
	err = filepath.WalkDir(rootChartDir, func(chartDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || chartDir == rootChartDir {
			return nil
		}

		t.Run(chartDir, func(t *testing.T) { checkDirHash(t, chartDir, sourceHash) })

		return filepath.SkipDir // skip sub-directories
	})
	require.NoError(t, err)

	// Check for deployments/helm/integrations chart
	chartDir := execwrapper.Abs("deployments/helm/integrations")
	t.Run(chartDir, func(t *testing.T) { checkDirHash(t, chartDir, sourceHash) })
}
