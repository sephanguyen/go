package main

import (
	"fmt"
	"os"

	"github.com/manabie-com/backend/internal/golibs/ci/statuscheck"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

type command struct {
	*cobra.Command

	ch statuscheck.Checker
}

// defaultRequiredJobs returns a list of jobs whose statuses must pass before
// a PR can be merged.
// Note that CLI --require argument can override this value.
func defaultRequiredJobs() []string {
	return []string{
		"check-commit-messages",
		"hasura-metadata-test",
		"integration-blocker-test",
		"lint",
		"skaffold-test",
		"unit-test",
	}
}

func newRootCmd() *command {
	longHelper := `Ensures all the status checks have passed before the PR can be merged.

Since we activate test jobs based on various criteria (file diff, squad ownership, etc...),
this tool also takes that into account:
  - if a test job is skipped, its status must be "skipped"
  - if a test job is activated, its status must be "success"

Any other statuses are considered a failure for current test run.

The --data argument must be a JSON containing ALL the necessary outputs from required test jobs.
This can easily be retrieved from "${{ toJSON(needs) }}" on Github Action.
`
	example := `  # Require unit-test and lint job to succeed
  export STATUS_DATA=${{ toJSON(needs) }}
  statuscheck --data="$STATUS_DATA" --require="unit-test,lint"`
	c := &command{
		Command: &cobra.Command{
			Use:          "statuscheck [flags]",
			Short:        "Ensure all the status checks have passed before the PR can be merged",
			Long:         longHelper,
			Example:      example,
			SilenceUsage: true,
			Args:         cobra.NoArgs,
		},
	}

	var logLevelString string
	c.RunE = func(cmd *cobra.Command, args []string) error {
		logLevel, err := zapcore.ParseLevel(logLevelString)
		if err != nil {
			return err
		}
		logger.UseDevelopmentLogger(logLevel)
		return c.ch.Run()
	}

	fl := c.Flags()
	fl.StringVarP(&logLevelString, "verbosity", "v", "info",
		`log level (can be: debug, info, warn, error, dpanic, panic, fatal)`)
	fl.StringVar(&c.ch.Data, "data", "",
		`raw JSON object containing the job's output`)
	fl.StringSliceVar(&c.ch.RequiredJobs, "require", defaultRequiredJobs(),
		`required jobs to pass`)
	if err := c.MarkFlagRequired("data"); err != nil {
		panic(err)
	}

	c.AddCommand(newVersionCmd())
	return c
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("v0.0.2")
			return nil
		},
		SilenceUsage: true,
	}
}
