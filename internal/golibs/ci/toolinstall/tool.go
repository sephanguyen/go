package toolinstall

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/pkg/errors"
)

var toolMap = map[string]Tooler{
	"jq": &tool{
		name:              "jq",
		versionCmd:        "--version",
		versionRegexp:     regexp.MustCompile(`^jq-([\d\.]+)(?:\r?\n)?$`),
		downloadURLFormat: "https://github.com/stedolan/jq/releases/download/jq-{{.Version}}/jq-linux64",
	},
	"skaffold": &tool{
		name:              "skaffold",
		versionCmd:        "version",
		versionRegexp:     regexp.MustCompile(`^v([0-9\.]+)(?:\r?\n)?$`),
		downloadURLFormat: "https://github.com/GoogleContainerTools/skaffold/releases/download/v{{.Version}}/skaffold-linux-amd64",
	},
	"skaffoldv2": &tool{
		name:              "skaffoldv2",
		versionCmd:        "version",
		versionRegexp:     regexp.MustCompile(`^v([0-9\.]+)(?:\r?\n)?$`),
		downloadURLFormat: "https://github.com/GoogleContainerTools/skaffold/releases/download/v{{.Version}}/skaffold-linux-amd64",
	},
	"kind": &tool{
		name:              "kind",
		versionCmd:        "version",
		versionRegexp:     regexp.MustCompile(`^kind v([0-9\.]+) [a-z0-9\.\/ ]+(?:\r?\n)?$`),
		downloadURLFormat: "https://github.com/kubernetes-sigs/kind/releases/download/v{{.Version}}/kind-linux-amd64",
	},
	"yq": &tool{
		name:              "yq",
		versionCmd:        "--version",
		versionRegexp:     regexp.MustCompile(`^yq \(.+\) version v?([0-9\.]+)(?:\r?\n)?$`),
		downloadURLFormat: "https://github.com/mikefarah/yq/releases/download/v{{.Version}}/yq_linux_amd64",
	},
	"gh": &tool{
		name:              "gh",
		versionCmd:        "version",
		versionRegexp:     regexp.MustCompile(`(?m)^gh version ([0-9\.]+) .+(?:\r?\n)?`),
		downloadURLFormat: "https://github.com/cli/cli/releases/download/v{{.Version}}/gh_{{.Version}}_linux_amd64.tar.gz",
		tgzTargetRegexp:   regexp.MustCompile(`[^\/]+/bin/gh$`),
	},
	"helm": &tool{
		name:              "helm",
		versionCmd:        "version",
		versionRegexp:     regexp.MustCompile(`^version.BuildInfo{Version:"v([0-9\.]+)", GitCommit:".+", GitTreeState:".+", GoVersion:".+"}(?:\r?\n)?$`),
		downloadURLFormat: "https://get.helm.sh/helm-v{{.Version}}-linux-amd64.tar.gz",
		tgzTargetRegexp:   regexp.MustCompile(`[^\/]+/helm$`),
	},
}

// Tooler is an interface that must be implemented to download a specific tool.
//
// TODO @anhpngt: implement checksum when downloading archives.
type Tooler interface {
	// Name returns the name of the binary.
	Name() string

	// SetInstallDir sets the target installation directory for this tool.
	SetInstallDir(targetInstallDir string)

	// ShouldInstall reports whether the tool should be installed/upgraded
	// e.g. it does not exist or is outdated.
	ShouldInstall(ctx context.Context, targetVersion string) (bool, error)

	// Install checks whether tool is already at targetVersion,
	// and installs it if not.
	//
	// It also verifies the tool post-installing.
	Install(ctx context.Context, targetVersion string) error
}

// tool implements Tooler.
type tool struct {
	// name is the name of the binary.
	name string

	// installDir is the installation directory of this tool, if installed by
	// this program.
	installDir string

	// versionCmd is the command which will be used to get the current version.
	// e.g. "version" or "--version"
	versionCmd string

	// versionRegexp is the regexp used to extract the version value
	// when running the version command.
	versionRegexp *regexp.Regexp

	// downloadURLFormat is the URL used to download the binary.
	// "{{.Version}}" template can be used to replace some parts of the URL
	// with the target version.
	downloadURLFormat string

	// tgzTargetRegexp is used to extract the binary from a tar archive.
	// If specified, the downloaded content is assumed to be an tar archive.
	// Otherwise, the download content is the binary itself.
	tgzTargetRegexp *regexp.Regexp
}

func (t *tool) Name() string {
	return t.name
}

func (t *tool) SetInstallDir(targetInstallDir string) {
	t.installDir = targetInstallDir
}

func (t *tool) ShouldInstall(ctx context.Context, targetVersion string) (bool, error) {
	toolName := t.Name()
	currentBinaryPath, currentVersion, err := t.getCurrentVersion(ctx)
	switch {
	case err == nil:
		if currentVersion == targetVersion {
			logger.Infof("%s %s already exists at %s, skipping installation", toolName, targetVersion, currentBinaryPath)
			return false, nil
		}

		// If binary is not in ~/.manabie (or whichever path we are managing), never touch it
		if !t.canUpdate(currentBinaryPath) {
			return false, fmt.Errorf("program %s is installed at %s, which is not managed by this scripts (at %s); "+
				"please either install %s %s manually, or uninstall the existing version and run this script again",
				toolName, currentBinaryPath, t.installDir, toolName, targetVersion,
			)
		}

		logger.Infof("%s is at version %s, target is %s, will update", toolName, currentVersion, targetVersion)
		return true, nil
	case errors.Is(err, errBinaryNotFound):
		logger.Infof("%s does not exist, will install", toolName)
		return true, nil
	case errors.Is(err, errExecVersionCmd):
		// this can happen when the binary is corrupted, so we just replace it with a fresh install
		logger.Warnf("failed to run \"%s %s\", will reinstall", toolName, t.versionCmd)
		return true, nil
	default:
		return false, fmt.Errorf("failed to get current version: %s", err)
	}
}

var (
	errExecVersionCmd = errors.New("failed to exec version command")
	errBinaryNotFound = exec.ErrNotFound
	errParseVersion   = errors.New("failed to parse version string")
)

// getToolCurrentVersion returns the existing tool's path and its versions.
func (t tool) getCurrentVersion(ctx context.Context) (string, string, error) {
	// Look in the installation dir first, then fallback to PATH
	binaryPath, err := exec.LookPath(t.binaryInstallPath())
	if err != nil {
		binaryPath, err = exec.LookPath(t.Name())
		if err != nil {
			return "", "", errors.Wrap(err, "exec.LookPath")
		}
	}
	execCmd := exec.CommandContext(ctx, binaryPath, t.versionCmd) //nolint:gosec
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Wrapf(errExecVersionCmd, "execCmd.Output: %s", err)
	}

	currentVersion, err := t.parseVersionOutput(string(output))
	if err != nil {
		return "", "", err
	}

	return binaryPath, currentVersion, nil
}

// parseVersionOutput parses and returns the version value from the output
// of version command.
func (t tool) parseVersionOutput(rawVersionOutput string) (string, error) {
	re := t.versionRegexp
	m := re.FindStringSubmatch(rawVersionOutput)
	if len(m) != 2 {
		return "", errors.Wrapf(
			errParseVersion,
			"failed to extract version from output %q, please delete your current installation of %s then try again",
			rawVersionOutput, t.Name(),
		)
	}
	return m[1], nil
}

func (t *tool) Install(ctx context.Context, targetVersion string) error {
	toolName := t.Name()
	if err := t.downloadAndInstall(ctx, targetVersion); err != nil {
		return err
	}
	logger.Debugf("finished downloading %s", toolName)
	if err := t.verifyInstall(ctx, targetVersion); err != nil {
		return fmt.Errorf("post-download verification failed: %s", err)
	}
	logger.Infof("%s %s has been successfully installed", toolName, targetVersion)
	return nil
}

func (t tool) downloadAndInstall(ctx context.Context, targetVersion string) error {
	downloadURL := t.downloadURL(targetVersion)
	installPath := t.binaryInstallPath()
	logger.Debugf("downloading %s with URL %q and destination %q", t.Name(), downloadURL, installPath)

	outFile, err := os.Create(installPath)
	if err != nil {
		return fmt.Errorf("failed to created new file at %s: %s", installPath, err)
	}
	defer outFile.Close()

	downloadCtx, downloadCancel := context.WithTimeout(ctx, time.Minute*5) // prevent background context
	defer downloadCancel()
	req, err := http.NewRequestWithContext(downloadCtx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %s", err)
	}
	logger.Debugf("downloading tool with http request %+v", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.Get: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected HTTP status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if t.tgzTargetRegexp != nil {
		err := t.decompress(outFile, resp.Body)
		if err != nil {
			return fmt.Errorf("t.Extract: %s", err)
		}
	} else {
		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			return fmt.Errorf("io.Copy: %s", err)
		}
	}

	logger.Debugf("making the binary executable")
	if err := os.Chmod(installPath, 0o755); err != nil {
		return fmt.Errorf("os.Chmod: %s", err)
	}
	return nil
}

func (t tool) downloadURL(targetVersion string) string {
	return strings.ReplaceAll(t.downloadURLFormat, `{{.Version}}`, targetVersion)
}

func (t tool) binaryInstallPath() string {
	return filepath.Join(t.installDir, t.Name())
}

func (t tool) verifyInstall(ctx context.Context, targetVersion string) error {
	_, currentVersion, err := t.getCurrentVersion(ctx)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			// a common error is binary installed but PATH does not include the binary
			expectedBinaryPath := t.binaryInstallPath()
			_, err := os.Stat(expectedBinaryPath)
			if err == nil {
				return fmt.Errorf(`%s exists at %s but is not in your PATH.
Please add %s to your PATH. You can usually do this by adding:
	export PATH=$PATH:%s
to your $HOME/.profile or $HOME/.bashrc (if using bash shell)`,
					t.Name(), expectedBinaryPath, t.installDir, t.installDir)
			}

			// error is not due to PATH, then
			return fmt.Errorf("could not find the installed binary (unknown error); please contact platform squad")
		}
	}
	if currentVersion != targetVersion {
		return fmt.Errorf("expected version %s after install for tool %s, got %s", targetVersion, t.Name(), currentVersion)
	}
	return nil
}

func (t tool) decompress(out io.Writer, in io.Reader) error {
	logger.Debugf("decompressing %s's archive", t.Name())
	gzipReader, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("gzip.NewReader: %s", err)
	}
	defer gzipReader.Close()

	re := t.tgzTargetRegexp
	// to prevent gosec's G110, we set a max copy of 200 MB
	const bytesToCopy int64 = 200 * 1024 * 1024
	tarReader := tar.NewReader(gzipReader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %s", err)
		}

		filename := hdr.Name
		if re.MatchString(filename) {
			logger.Debugf("found %s in archive", filename)
			bytesCopied, err := io.CopyN(out, tarReader, bytesToCopy)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if bytesCopied == bytesToCopy {
				return fmt.Errorf("input file exceed size limit (%d bytes)", bytesToCopy)
			}
		}
	}
	return nil
}

// canUpdate reports whether the existing binary was installed by this tool
// (by checking if the binary is installed in targetInstallDir).
func (t tool) canUpdate(binaryPath string) bool {
	return filepath.Dir(binaryPath) == filepath.Clean(t.installDir)
}
