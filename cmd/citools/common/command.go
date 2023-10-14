package common

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Command struct {
	cmd *cobra.Command

	logLevel *string
}

func NewCommand() *Command {
	return &Command{
		cmd: &cobra.Command{SilenceUsage: true},
	}
}

func (cc *Command) Do() error {
	cc.cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cc.logLevel != nil {
			logger.UseDevelopmentLoggerString(*cc.logLevel) // TODO @anhpngt: return error instead
		}
		return nil
	}
	return cc.cmd.Execute()
}

func (cc *Command) WithUsage(s string) *Command {
	cc.cmd.Use = s
	return cc
}

func (cc *Command) WithShortDescription(s string) *Command {
	cc.cmd.Short = s
	return cc
}

func (cc *Command) WithLongDescription(s string) *Command {
	cc.cmd.Long = s
	return cc
}

func (cc *Command) WithExample(s string) *Command {
	cc.cmd.Example = s
	return cc
}

func (cc *Command) WithExpectedArgs(a cobra.PositionalArgs) *Command {
	cc.cmd.Args = a
	return cc
}

func (cc *Command) WithTask(f func(cmd *cobra.Command, args []string) error) *Command {
	cc.cmd.RunE = f
	return cc
}

func (cc *Command) WithVersion(version string) *Command {
	versionCmd := &cobra.Command{
		Use:          "version",
		Short:        "Print the version information",
		Args:         cobra.NoArgs,
		Run:          func(cmd *cobra.Command, args []string) { fmt.Println(version) },
		SilenceUsage: true,
	}
	cc.cmd.AddCommand(versionCmd)
	return cc
}

func (cc *Command) WithFlag(f func(*pflag.FlagSet)) *Command {
	f(cc.cmd.Flags())
	return cc
}

func (cc *Command) WithLogLevelFlag() *Command {
	cc.logLevel = cc.cmd.Flags().StringP("verbosity", "v", "info",
		`log level (can be: debug, info, warn, error, dpanic, panic, fatal)`)
	return cc
}

func (cc *Command) WithSubCommand(sc *Command) *Command {
	cc.cmd.AddCommand(sc.cmd)
	return cc
}
