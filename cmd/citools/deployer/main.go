package main

import (
	"os"

	"github.com/manabie-com/backend/cmd/citools/common"
	"github.com/manabie-com/backend/internal/golibs/ci/deployer"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	if err := newDeployerCmd().Do(); err != nil {
		os.Exit(1)
	}
}

func newDeployerCmd() *common.Command {
	return common.NewCommand().
		WithUsage("deployer").
		WithShortDescription("Deploy services in Kubernetes").
		WithLogLevelFlag().
		WithVersion("v0.0.0-rc2").
		WithTask(func(cmd *cobra.Command, args []string) error {
			return deployer.DoDeploy(args)
		}).
		WithSubCommand(newRenderCmd()).
		WithSubCommand(newBuildCmd()).
		WithSubCommand(newDeployCmd()).
		WithSubCommand(newRunCmd())
}

func newRenderCmd() *common.Command {
	return common.NewCommand().WithUsage("render").
		WithShortDescription("Run `skaffold render`").
		WithLogLevelFlag().
		WithFlag(func(f *pflag.FlagSet) {
			f.StringP("filename", "f", "skaffold.manaverse.yaml", "Path or URL to the Skaffold config file")
		}).
		WithTask(func(cmd *cobra.Command, _ []string) error {
			filename, err := cmd.Flags().GetString("filename")
			if err != nil {
				return err
			}
			return deployer.SkaffoldRender("-f", filename)
		})
}

func newBuildCmd() *common.Command {
	return common.NewCommand().
		WithUsage("build").
		WithTask(func(cmd *cobra.Command, args []string) error {
			return deployer.SkaffoldBuild()
		})
}

func newDeployCmd() *common.Command {
	return common.NewCommand().
		WithUsage("deploy").
		WithTask(func(cmd *cobra.Command, args []string) error {
			return deployer.SkaffoldDeploy()
		})
}

func newRunCmd() *common.Command {
	return common.NewCommand().WithUsage("run").
		WithLogLevelFlag().
		WithTask(func(cmd *cobra.Command, args []string) error {
			return deployer.SkaffoldRun()
		})
}
