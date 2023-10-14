package secret

import "github.com/spf13/cobra"

func NewCmdSecret() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "manage secrets",
	}
	cmd.AddCommand(NewCmdSecretTest())
	return cmd
}
