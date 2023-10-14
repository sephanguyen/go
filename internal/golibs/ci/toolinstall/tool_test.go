package toolinstall

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func mockTooler() *tool {
	return &tool{
		name:          "mocktooler.sh",
		versionCmd:    "version",
		versionRegexp: regexp.MustCompile(`^([\d\.]+)(?:\r?\n)?$`),
	}
}

func pathToTestdata() string {
	return execwrapper.Abs("internal/golibs/ci/toolinstall/testdata")
}

func TestTool_ShouldInstall(t *testing.T) {
	mt := mockTooler()
	ctx := context.Background()

	t.Run("v1 not yet exists, thus needs to be installed", func(t *testing.T) {
		b, err := mt.ShouldInstall(ctx, "1.0.0")
		require.NoError(t, err)
		require.True(t, b)
	})

	// Add testdata/ to PATH, so that mocktooler.sh is at v1.0.0
	currentPath := os.Getenv("PATH")
	t.Setenv("PATH", fmt.Sprintf("%s:%s", pathToTestdata(), currentPath))

	// start testing
	t.Run("v1 already installed", func(t *testing.T) {
		b, err := mt.ShouldInstall(ctx, "1.0.0")
		require.NoError(t, err)
		require.False(t, b)
	})

	t.Run("not allowed to upgrade", func(t *testing.T) {
		b, err := mt.ShouldInstall(ctx, "2.0.0")
		require.NotNil(t, err)
		require.False(t, b)
	})

	t.Run("can upgrade to v2", func(t *testing.T) {
		// Set testdata/ as installDir to allow upgrading mocktooler
		mt.installDir = pathToTestdata()
		b, err := mt.ShouldInstall(ctx, "2.0.0")
		require.NoError(t, err)
		require.True(t, b)
	})

	t.Run("corrupted file error, should install to replace", func(t *testing.T) {
		oldVersionCmd := mt.versionCmd
		mt.versionCmd = "die"
		t.Cleanup(func() { mt.versionCmd = oldVersionCmd })
		b, err := mt.ShouldInstall(ctx, "1.0.0")
		require.NoError(t, err)
		require.True(t, b)
	})

	t.Run("invalid tool error", func(t *testing.T) {
		oldVersionCmd := mt.versionCmd
		mt.versionCmd = "wrongversion"
		t.Cleanup(func() { mt.versionCmd = oldVersionCmd })
		b, err := mt.ShouldInstall(ctx, "1.0.0")
		require.NotNil(t, err)
		require.False(t, b)
	})
}

func getTool(toolName string) (*tool, error) {
	tooler, ok := toolMap[toolName]
	if !ok {
		return nil, fmt.Errorf("unknown Tooler %s", toolName)
	}
	tool, ok := tooler.(*tool)
	if !ok {
		return nil, fmt.Errorf("Tooler %s cannot be converted to tool", toolName)
	}
	return tool, nil
}

func TestTool_parseVersionOutput(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		desc             string
		toolName         string
		rawVersionOutput string
		expectedVersion  string
		expectedErr      error
	}{
		{desc: "jq 1.6", toolName: "jq", rawVersionOutput: "jq-1.6", expectedVersion: "1.6"},
		{desc: "jq 1.5.2", toolName: "jq", rawVersionOutput: "jq-1.5.2\n", expectedVersion: "1.5.2"},
		{desc: "jq invalid version", toolName: "jq", rawVersionOutput: "invalid version", expectedErr: errParseVersion},
		{desc: "skaffold", toolName: "skaffold", rawVersionOutput: "v1.39.2", expectedVersion: "1.39.2"},
		{desc: "skaffold version 2", toolName: "skaffold", rawVersionOutput: "v2.0.5", expectedVersion: "2.0.5"},
		{desc: "skaffoldv2", toolName: "skaffoldv2", rawVersionOutput: "v2.4.0\n", expectedVersion: "2.4.0"},
		{desc: "kind", toolName: "kind", rawVersionOutput: "kind v0.14.0 go1.18.2 linux/amd64", expectedVersion: "0.14.0"},
		{desc: "kind missing go info", toolName: "kind", rawVersionOutput: "kind v0.14.0", expectedErr: errParseVersion},
		{desc: "yq", toolName: "yq", rawVersionOutput: "yq (https://github.com/mikefarah/yq/) version v4.28.2\n", expectedVersion: "4.28.2"},
		{desc: "yq", toolName: "yq", rawVersionOutput: "yq (https://github.com/mikefarah/yq/) version 4.34.1\n", expectedVersion: "4.34.1"},
		{desc: "gh", toolName: "gh", rawVersionOutput: `gh version 2.21.2 (2023-01-03)
https://github.com/cli/cli/releases/tag/v2.21.2


A new release of gh is available: 2.21.2 → 2.22.1
https://github.com/cli/cli/releases/tag/v2.22.1`, expectedVersion: "2.21.2"},
		{desc: "helm", toolName: "helm", rawVersionOutput: "version.BuildInfo{Version:\"v3.11.0\", GitCommit:\"472c5736ab01133de504a826bd9ee12cbe4e7904\", GitTreeState:\"clean\", GoVersion:\"go1.18.10\"}\n", expectedVersion: "3.11.0"},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			tool, err := getTool(tc.toolName)
			require.NoError(t, err)
			actualVersion, actualErr := tool.parseVersionOutput(tc.rawVersionOutput)
			require.Truef(t, errors.Is(actualErr, tc.expectedErr), "unexpected error: %s", actualErr)
			require.Equal(t, tc.expectedVersion, actualVersion)
		})
	}
}

func TestTool_tgzTargetRegexp(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		desc                string
		toolName            string
		inputFileList       []string
		expectedMatchedFile string
	}{
		{desc: "gh", toolName: "gh", inputFileList: []string{"gh_2.21.2_linux_amd64/LICENSE", "gh_2.21.2_linux_amd64/bin/gh"}, expectedMatchedFile: "gh_2.21.2_linux_amd64/bin/gh"},
		{desc: "helm", toolName: "helm", inputFileList: []string{"linux-amd64/LICENSE", "linux-amd64/helm"}, expectedMatchedFile: "linux-amd64/helm"},
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			tool, err := getTool(tc.toolName)
			require.NoError(t, err)
			re := tool.tgzTargetRegexp
			require.NotNil(t, re)
			for _, f := range tc.inputFileList {
				matched := re.MatchString(f)
				require.Equal(t, f == tc.expectedMatchedFile, matched)
			}
		})
	}
}
