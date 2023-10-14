package main

import (
	"os"

	"github.com/manabie-com/backend/cmd/citools/common"
	"github.com/manabie-com/backend/internal/golibs/ci/toolinstall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	if err := newInstallerCmd().Do(); err != nil {
		os.Exit(1)
	}
}

func newInstallerCmd() *common.Command {
	example := `
  # Install jq@1.6
  toolinstall jq@1.6`
	return common.NewCommand().
		WithUsage("install NAME[@VERSION] ...").
		WithShortDescription("Install toolings and binaries").
		WithExample(example).
		WithExpectedArgs(cobra.MinimumNArgs(1)).
		WithLogLevelFlag().
		WithFlag(func(fs *pflag.FlagSet) {
			fs.StringP("install-dir", "d", "", "installation directory (default: $MANABIE_HOME or ~/.manabie/bin)")
		}).
		WithVersion("v0.2.0").
		WithTask(func(cmd *cobra.Command, args []string) error {
			installDir, err := cmd.Flags().GetString("install-dir")
			if err != nil {
				return err
			}
			return toolinstall.Run(cmd.Context(), installDir, args)
		})
}
