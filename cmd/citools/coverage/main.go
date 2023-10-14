package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/manabie-com/backend/internal/golibs/ci/coverage"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	c := &coverage.C{}
	rootCmd := &cobra.Command{}
	fl := rootCmd.PersistentFlags()
	fl.IntVar(&c.TimeoutInSeconds, "timeout", 60,
		`timeout in seconds, setting to zero or below will cause immediate timeout`)
	fl.StringVarP(&c.LogLevelString, "verbosity", "v", "info",
		`log level (can be: debug, info, warn, error, dpanic, panic, fatal)`)

	rootCmd.AddCommand(newUpdateCoverageCmd(c))
	rootCmd.AddCommand(newCompareCoverageCmd(c))
	rootCmd.AddCommand(newVersionCmd())
	return rootCmd
}

func newUpdateCoverageCmd(c *coverage.C) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update code coverage of a branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.UseDevelopmentLoggerString(c.LogLevelString)
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*time.Duration(c.TimeoutInSeconds))
			defer cancel()
			if err := c.Dial(ctx); err != nil {
				return err
			}
			defer c.Close()
			return c.UpdateCoverage(ctx)
		},
		SilenceUsage: true,
	}
	fl := cmd.Flags()
	fl.StringVar(&c.Ref, "ref", "develop",
		`target branch or ref, or "${{ github.ref_name }}" in Github Action`)
	fl.StringVar(&c.CoverageFilepath, "coverfile", "./cover.func",
		`path to the file containing the final coverage`)
	fl.StringVar(&c.RepositoryName, "repo", "manabie-com/backend",
		`repository name`)
	fl.StringVar(&c.SecretKey, "key", "",
		`secret key for authentication`)
	fl.StringVar(&c.ServerAddr, "address", "api.staging.manabie.io:443",
		"address to the Draft server")
	fl.BoolVar(&c.IsIntegrationTest, "integration", false,
		`whether this is integration test`)
	return cmd
}

func newCompareCoverageCmd(c *coverage.C) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Verifiy code coverage is satisfactory for a target branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.UseDevelopmentLoggerString(c.LogLevelString)
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*time.Duration(c.TimeoutInSeconds))
			defer cancel()
			if err := c.Dial(ctx); err != nil {
				return err
			}
			defer c.Close()
			return c.CompareCoverage(ctx)
		},
		SilenceUsage: true,
	}
	fl := cmd.Flags()
	fl.StringVar(&c.HeadRef, "head-ref", "",
		`head ref, e.g. "my-feature-branch" or "${{ github.head_ref }}`)
	fl.StringVar(&c.BaseRef, "base-ref", "develop",
		`base ref, e.g. "develop" or "${{ github.base_ref }}`)
	fl.StringVar(&c.CoverageFilepath, "coverfile", "./cover.func",
		`path to the file containing the final coverage`)
	fl.StringVar(&c.RepositoryName, "repo", "manabie-com/backend",
		`repository name`)
	fl.StringVar(&c.SecretKey, "key", "",
		`secret key for authentication`)
	fl.StringVar(&c.ServerAddr, "address", "api.staging.manabie.io:443",
		"address to the Draft server")
	fl.BoolVar(&c.IsIntegrationTest, "integration", false,
		`whether this is integration test`)
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("v0.0.3")
			return nil
		},
		SilenceUsage: true,
	}
}
