package toolinstall

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func init() {
	defaultInstallDir = pathToTestdata()
}

func TestInstaller_parseNameAndVersion(t *testing.T) {
	t.Parallel()

	expectedTool := toolMap["jq"]
	tool, version, err := parseToolNameAndVersion("jq@1.6")
	require.NoError(t, err)
	require.Equal(t, expectedTool, tool)
	require.Equal(t, "1.6", version)

	tool, version, err = parseToolNameAndVersion("jq@v1.6")
	require.NoError(t, err)
	require.Equal(t, expectedTool, tool)
	require.Equal(t, "1.6", version)

	tool, version, err = parseToolNameAndVersion("unsupportedtool@1.6")
	require.EqualError(t, err, "program unsupportedtool is not supported")
	require.Equal(t, nil, tool)
	require.Equal(t, "", version)

	tool, version, err = parseToolNameAndVersion("jq-1.6")
	require.EqualError(t, err, "invalid input format, expected \"name@version\"")
	require.Equal(t, nil, tool)
	require.Equal(t, "", version)
}

func httptestServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/mocktooler/1.0.0", func(w http.ResponseWriter, r *http.Request) {
		mocktooler, err := os.ReadFile(filepath.Join(pathToTestdata(), "mocktooler.sh"))
		require.NoError(t, err)
		w.Write(mocktooler)
	})
	mux.HandleFunc("/tgz/mocktooler/1.0.0", func(w http.ResponseWriter, r *http.Request) {
		zw := gzip.NewWriter(w)
		defer zw.Close()
		tw := tar.NewWriter(zw)
		defer tw.Close()

		mocktooler, err := os.ReadFile(filepath.Join(pathToTestdata(), "mocktooler.sh"))
		require.NoError(t, err)
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: "mocktooler.sh", Size: int64(len(mocktooler)), Mode: 0o600}))
		_, err = tw.Write(mocktooler)
		require.NoError(t, err)

		// add another random file
		readme := []byte("this is a readme")
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: "README.md", Size: int64(len(readme)), Mode: 0o600}))
		_, err = tw.Write(readme)
		require.NoError(t, err)
	})
	return httptest.NewServer(mux)
}

func TestTool_Install(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ts := httptestServer(t)
	defer ts.Close()

	t.Run("install mocktooler@1.0.0", func(t *testing.T) {
		tool := &tool{
			name:              "mocktooler.sh",
			installDir:        t.TempDir(),
			versionCmd:        "version",
			versionRegexp:     regexp.MustCompile(`^([\d\.]+)(?:\r?\n)?$`),
			downloadURLFormat: ts.URL + "/mocktooler/{{.Version}}",
		}
		err := tool.Install(ctx, "1.0.0")
		require.NoError(t, err)
	})

	t.Run("install mocktooler@1.0.0 in tgz", func(t *testing.T) {
		tool := &tool{
			name:              "mocktooler.sh",
			installDir:        t.TempDir(),
			versionCmd:        "version",
			versionRegexp:     regexp.MustCompile(`^([\d\.]+)(?:\r?\n)?$`),
			downloadURLFormat: ts.URL + "/tgz/mocktooler/{{.Version}}",
			tgzTargetRegexp:   regexp.MustCompile(`mocktooler\.sh`),
		}
		err := tool.Install(ctx, "1.0.0")
		require.NoError(t, err)
	})
}

func TestInstaller_getToolCurrentVersion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testcases := []struct {
		desc               string
		tool               *tool
		expectedBinaryPath string
		expectedVersion    string
		expectedErr        error
	}{
		{desc: "successfully get version", tool: &tool{name: "mocktooler.sh", installDir: pathToTestdata(), versionCmd: "version", versionRegexp: regexp.MustCompile(`^([\d\.]+)(?:\r?\n)?$`)}, expectedBinaryPath: filepath.Join(defaultInstallDir, "mocktooler.sh"), expectedVersion: "1.0.0"},
		{desc: "binary not found", tool: &tool{name: "nonexistent.sh", installDir: pathToTestdata()}, expectedErr: errBinaryNotFound},
		{desc: "version command failed", tool: &tool{name: "mocktooler.sh", installDir: pathToTestdata(), versionCmd: "die"}, expectedErr: errExecVersionCmd},
		{desc: "failed to parse version", tool: &tool{name: "mocktooler.sh", installDir: pathToTestdata(), versionCmd: "wrongversion", versionRegexp: regexp.MustCompile(`^([\d\.]+)(?:\r?\n)?$`)}, expectedErr: errParseVersion},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			actualBinaryPath, actualVersion, actualErr := tc.tool.getCurrentVersion(ctx)
			require.True(t, errors.Is(actualErr, tc.expectedErr), "unexpected error")
			require.Equal(t, tc.expectedVersion, actualVersion)
			require.Equal(t, tc.expectedBinaryPath, actualBinaryPath)
		})
	}
}
