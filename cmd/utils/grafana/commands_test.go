package grafana

import (
	"strings"
	"testing"

	fileio "github.com/manabie-com/backend/internal/golibs/io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomDashboards(t *testing.T) {
	files, err := fileio.GetFileNamesOnDir(".")
	require.NoError(t, err)
	srvNames := make(map[string]bool)
	for _, fileName := range files {
		fileName = strings.ToLower(fileName)
		if !strings.HasSuffix(fileName, "_test.go") && fileName != "commands.go" && strings.HasSuffix(fileName, ".go") {
			srvNames[strings.Split(fileName, ".")[0]] = true
		}
	}

	assert.Len(t, customServices, len(srvNames))
	for customService := range customServices {
		assert.NotNil(t, srvNames[customService])
	}
}

func TestGenDashboards(t *testing.T) {
	path, err := fileio.GetAbsolutePathFromRepoRoot(destinationPath)
	require.NoError(t, err)
	files, err := fileio.GetFileNamesOnDir(path)
	require.NoError(t, err)
	srvNames := make(map[string]bool)
	for _, fileName := range files {
		fileName = strings.ToLower(fileName)
		if strings.HasSuffix(fileName, ".json") {
			srvNames[strings.Split(fileName, "-")[1]] = true
		}
	}

	expected, err := getFeatureServices()
	require.NoError(t, err)

	assert.Len(t, srvNames, len(expected))
	for _, expectedName := range expected {
		assert.NotNil(t, srvNames[expectedName])
	}
}

func TestGenHasuraDashboards(t *testing.T) {
	t.Skip()
	path, err := fileio.GetAbsolutePathFromRepoRoot(hasuraDestinationPath)
	require.NoError(t, err)
	files, err := fileio.GetFileNamesOnDir(path)
	require.NoError(t, err)
	srvNames := make(map[string]bool)
	for _, fileName := range files {
		fileName = strings.ToLower(fileName)
		if strings.HasSuffix(fileName, ".json") {
			srvNames[strings.Split(fileName, "-")[0]] = true
		}
	}

	expected, err := getHasuraServices()
	require.NoError(t, err)

	assert.Len(t, srvNames, len(expected))
	for _, expectedName := range expected {
		assert.NotNil(t, srvNames[expectedName])
	}
}
