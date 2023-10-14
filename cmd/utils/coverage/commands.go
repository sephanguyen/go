package coverage

import (
	"github.com/spf13/cobra"
)

// port to set flag port in utils command
var (
	address     string
	branch      string
	repository  string
	key         string
	baseBranch  string
	integration bool
)

// CoverageCompareCmd to compare coverage
var CoverageCompareCmd = &cobra.Command{
	Use:   "compare current branch",
	Short: "parse and compare cover function output of current branch",

	// Args:  verifyCompareCoverageArgs,
	RunE: compareCoverage,
}

// RootCmd for coverage
var RootCmd = &cobra.Command{
	Use:   "coverage [command]",
	Short: "coverage compare/create",
}

// CreateTargetCmd to create target branch's code coverage
var CreateTargetCmd = &cobra.Command{
	Use:   "create cover.func",
	Short: "store code coverage of base branch ",
	RunE:  createTargetCoverage,
}

// UpdateTargetCmd to update target branch's code coverage
var UpdateTargetCmd = &cobra.Command{
	Use:   "update cover.func",
	Short: "update code coverage of base branch",
	RunE:  updateTargetCoverage,
}

func init() {
	CoverageCompareCmd.PersistentFlags().StringVar(&address, "address", "", "address server")
	CoverageCompareCmd.PersistentFlags().StringVar(&branch, "branch", "", "branch name")
	CoverageCompareCmd.PersistentFlags().StringVar(&repository, "repo", "", "repository name")
	CoverageCompareCmd.PersistentFlags().StringVar(&baseBranch, "base", "develop", "base branch")
	CoverageCompareCmd.PersistentFlags().StringVar(&key, "key", "", "secret key")
	CoverageCompareCmd.PersistentFlags().BoolVar(&integration, "integration", false, "integration tests")

	CreateTargetCmd.PersistentFlags().StringVar(&address, "address", "", "address server")
	CreateTargetCmd.PersistentFlags().StringVar(&branch, "branch", "", "branch name")
	CreateTargetCmd.PersistentFlags().StringVar(&repository, "repo", "", "repository name")
	CreateTargetCmd.PersistentFlags().StringVar(&key, "key", "", "secret key")
	CreateTargetCmd.PersistentFlags().BoolVar(&integration, "integration", false, "integration tests")

	UpdateTargetCmd.PersistentFlags().StringVar(&address, "address", "", "address server")
	UpdateTargetCmd.PersistentFlags().StringVar(&branch, "branch", "", "branch name")
	UpdateTargetCmd.PersistentFlags().StringVar(&repository, "repo", "", "repository name")
	UpdateTargetCmd.PersistentFlags().StringVar(&key, "key", "", "secret key")
	UpdateTargetCmd.PersistentFlags().BoolVar(&integration, "integration", false, "integration tests")

	RootCmd.AddCommand(
		CoverageCompareCmd,
		CreateTargetCmd,
		UpdateTargetCmd,
	)
}
