package toolinstall

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/logger"
)

var (
	homeDir           string
	defaultInstallDir string
)

func init() {
	var ok bool
	defaultInstallDir, ok = os.LookupEnv("MANABIE_HOME")
	if !ok {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		defaultInstallDir = filepath.Join(homeDir, ".manabie/bin")
	}
	logger.Infof("default installation directory: %s", defaultInstallDir)
}

// Run is the entrypoint for tool installer.
func Run(ctx context.Context, targetInstallDir string, args []string) error {
	var createDirOnce sync.Once
	if targetInstallDir == "" {
		targetInstallDir = defaultInstallDir
	}
	logger.Infof("target installation directory: %s", targetInstallDir)
	for _, nameVersion := range args {
		tool, targetVersion, err := parseToolNameAndVersion(nameVersion)
		if err != nil {
			return fmt.Errorf("failed to parse name version string %q: %s", nameVersion, err)
		}
		tool.SetInstallDir(targetInstallDir)
		needInstall, err := tool.ShouldInstall(ctx, targetVersion)
		if err != nil {
			return err
		}
		if !needInstall {
			continue
		}

		createDirOnce.Do(func() { err = os.MkdirAll(targetInstallDir, 0o750) })
		if err != nil {
			return fmt.Errorf("failed to create installation directory: %s", err)
		}

		if err := tool.Install(ctx, targetVersion); err != nil {
			return err
		}
	}
	return nil
}

// parseToolNameAndVersion returns the tool and desired version specified in cli inputs.
// For example: jq@1.6 returns tool "jq" and "1.6".
func parseToolNameAndVersion(nameVersion string) (Tooler, string, error) {
	ss := strings.SplitN(nameVersion, "@", 2)
	if len(ss) != 2 {
		return nil, "", fmt.Errorf("invalid input format, expected \"name@version\"")
	}

	name := ss[0]
	version := strings.TrimPrefix(ss[1], "v")
	tool, ok := toolMap[name]
	if !ok {
		return nil, "", fmt.Errorf("program %s is not supported", name)
	}
	return tool, version, nil
}
