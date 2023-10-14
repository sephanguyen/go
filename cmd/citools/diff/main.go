package main

import (
	"fmt"
	"os"

	"github.com/manabie-com/backend/internal/golibs/ci/diff"
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
	differ diff.Differ
}

func newRootCmd() *command {
	longHelper := `Determine the required tests to run on CI.

See https://manabie.atlassian.net/l/cp/HpHNPpFb for more info on running CI on backend.

.github/scripts/diff.yaml can be updated to reflect the criteria of when a rule is triggered.
The syntax of a rule is as follow:

  - name       		: rule's name, also affects its Github Action output name
  - paths      		: a list of regexp paths; when any files in these paths are updated, this rule is triggered
  - paths-ignore	: a list of regexp paths; files in these paths are always ignored for this rule
  - force_value		: (optional) value to set when force test is enabled
  - run_only   		: (default is false) when true, the rule is triggered only when ALL files match at least one path
  - values     		: (optional) these values replace the {{.VALUE}} in paths, for convenience
                 when the rule is triggered, instead of 0 or 1, it outputs the triggered value instead
  - enabled_squads 	: squads for whom this rule is enabled for; mutually exclusive with "disabled_squads".
                      When specified, this rule only triggered for actors that belongs to these squads.
					  When both "enabled_squads" and "disabled_squads" are empty, the rule is triggered by default.
  - disabled_squads	: squads for whom this rule is disabled for; mutually exclusive with "enabled_squads."
  					  When specified, this rule only triggered for actors that do not belongs to these squads.`

	example := `
  # Force all test
  diff -f

  # Get the requirement based on 2 branches and PR description
  diff --base-ref=develop --head-ref=my-feature-branch --pr-desc="My PR summary"

  # Get the requirement based on PR description only
  diff --pr-desc="My PR summary" --pr-desc-only`

	c := &command{
		Command: &cobra.Command{
			Use:          "diff [flags]",
			Short:        "Determine the required tests to run on CI.",
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
		return c.differ.Run()
	}

	fl := c.Flags()
	fl.StringVarP(&logLevelString, "verbosity", "v", "info",
		`log level (can be: debug, info, warn, error, dpanic, panic, fatal)`)
	fl.BoolVar(&c.differ.Force, "force", false,
		`force running all tests`)
	fl.StringVar(&c.differ.PRDesc, "pr-desc", "",
		`pull request description that allows adding extra tests to run`)
	fl.BoolVar(&c.differ.PRDescOnly, "pr-desc-only", false,
		`only extract test requirements from pull request description`)
	fl.StringVar(&c.differ.BaseRef, "base-ref", "",
		`base ref of the pull request, e.g. "develop" or "${{ github.base_ref }}"`)
	fl.StringVar(&c.differ.HeadRef, "head-ref", "",
		`head ref of the pull request, e.g. "my-feature-branch" or "${{ github.head_ref }}"`)
	fl.StringVar(&c.differ.ConfigPath, "config-path", ".github/scripts/diff.yaml",
		`path to the config file containing diff rules`)
	fl.StringVarP(&c.differ.OutputPath, "output-path", "o", "",
		`location to write output to, defaults to os.Stdout if not specified`)
	fl.StringSliceVar(&c.differ.Squads, "squads", nil,
		`list of squads that the workflow actor belongs to`)

	c.AddCommand(newVersionCmd())
	return c
}

// newVersionCmd returns a command showing the current version of diff.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("v0.0.8")
			return nil
		},
		SilenceUsage: true,
	}
}
