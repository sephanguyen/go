// package handles running git commands.
// References: github.com/cli/cli
package execwrapper

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	once          sync.Once
	rootDirectory string
)

// RootDirectory returns the absolute path of the current project's
// root directory (based on git's output).
//
// It panics if current working directory does not belong to a git project.
func RootDirectory() string {
	once.Do(func() {
		var err error
		rootDirectory, err = GitTopLevelDir()
		if err != nil {
			panic("`git rev-parse --show-toplevel` failed (please `cd` to the backend's repository)")
		}
	})
	return rootDirectory
}

// Abs is a shorthand for filepath.Join(RootDirectory(), filepath.Join(subpaths...)).
func Abs(subpaths ...string) string {
	return filepath.Join(RootDirectory(), filepath.Join(subpaths...))
}

// Absf is a shorthand for Abs(fmt.Sprintf(path, args...)).
func Absf(path string, args ...any) string {
	return Abs(fmt.Sprintf(path, args...))
}

func GitCommand(args ...string) (*exec.Cmd, error) {
	gitExe, err := LookPath("git")
	if err != nil {
		return nil, err
	}
	return exec.Command(gitExe, args...), nil
}

func GitTopLevelDir() (string, error) {
	showCmd, err := GitCommand("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	showCmd.Stderr = os.Stderr
	output, err := showCmd.Output()
	return firstLine(output), err
}

func firstLine(output []byte) string {
	if i := bytes.IndexAny(output, "\n"); i >= 0 {
		return string(output)[0:i]
	}
	return string(output)
}

// GitDiff returns the list of changed files between 2 commits.
// Under the hood, `git diff commit1...commit2 --name-only --diff-filter=ACMDRT` is run.
func GitDiff(commit1, commit2 string) ([]string, error) {
	diffCmd, err := GitCommand("diff", "--name-only", "--diff-filter=ACMDRT", commit1+"..."+commit2)
	if err != nil {
		return nil, err
	}
	diffCmd.Stderr = os.Stderr
	output, err := diffCmd.Output()
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	outputStr := strings.TrimRight(string(output), "\n")
	return strings.Split(outputStr, "\n"), nil
}

// MockedGitDir returns the path to the mocked git directory used in testings.
func MockedGitDir() string {
	return filepath.Join(RootDirectory(), "internal/golibs/execwrapper/testdata/repo/test.git")
}

// SetGitDir sets GIT_DIR to a custom .git directory to manipulate the
// output of git. Used in testings only.
// See https://git-scm.com/book/en/v2/Git-Internals-Environment-Variables
// See internal/golibs/execwrapper/testdata/repo/README.md to know how to set
// up your mocked commits.
func SetGitDir(t *testing.T, dir string) {
	t.Setenv("GIT_DIR", dir)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}
