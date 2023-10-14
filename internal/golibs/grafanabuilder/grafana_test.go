package grafanabuilder

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	cfg, err := os.Open("testdata/test_dashboard_cfg.jsonnet")
	require.NoError(t, err)
	defer cfg.Close()
	extend, err := os.Open("testdata/test_panel_target.jsonnet")
	require.NoError(t, err)
	defer extend.Close()

	tmpFolder := "tmp_" + idutil.ULIDNow()
	err = os.Mkdir(tmpFolder, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(tmpFolder)

	err = (&Builder{}).
		AddDashboardConfigFiles(cfg, map[string]io.Reader{"test_panel_target.jsonnet": extend}).
		AddDestinationFilePath(tmpFolder+"/res.json").
		AddGrafanaConfig("testdata/test_grafana_config.yaml", tmpFolder+"/values.yaml", "example", "dashboards/example.json").
		Build()
	require.NoError(t, err)

	// validate new dashboard config json file
	f1, err := os.Open(tmpFolder + "/res.json")
	require.NoError(t, err)
	defer f1.Close()

	h1 := sha256.New()
	_, err = io.Copy(h1, f1)
	require.NoError(t, err)

	f2, err := os.Open("testdata/expected_dashboard.json")
	require.NoError(t, err)
	defer f2.Close()

	h2 := sha256.New()
	_, err = io.Copy(h2, f2)
	require.NoError(t, err)

	assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

	// validate grafana config yaml file
	f3, err := os.Open(tmpFolder + "/values.yaml")
	require.NoError(t, err)
	defer f3.Close()

	h3 := sha256.New()
	_, err = io.Copy(h3, f3)
	require.NoError(t, err)

	f4, err := os.Open("testdata/expected_grafana_config.yaml")
	require.NoError(t, err)
	defer f4.Close()

	h4 := sha256.New()
	_, err = io.Copy(h4, f4)
	require.NoError(t, err)

	assert.Equal(t, 0, bytes.Compare(h3.Sum(nil), h4.Sum(nil)))
}
