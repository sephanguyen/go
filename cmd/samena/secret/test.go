package secret

import (
	"github.com/manabie-com/backend/internal/golibs/execwrapper"

	"github.com/spf13/cobra"
)

func NewCmdSecretTest() *cobra.Command {
	longHelper := `
Run unit test for secrets, defined in 'github.com/manabie-com/backend/cmd/samena/secret/tests' package.

Running this command is equivalent to running:
  go test github.com/manabie-com/backend/cmd/samena/secret/tests

This command requires user to have super privilege (project owner or
belonging to dev.infras@manabie.com group) to decrypt production secrets.
`

	example := `
  # Most basic usage
  go run cmd/samena/main.go secret test

  # Pass extra args to the 'go test' command using two dashes (--)
  go run cmd/samena/main.go secret test -- -run TestExampleFunction`

	cmd := &cobra.Command{
		Use:     "test",
		Short:   "test encrypted secrets in this repository",
		Long:    longHelper,
		Example: example,
		RunE: func(cmd *cobra.Command, args []string) error {
			return execwrapper.GoTest("github.com/manabie-com/backend/cmd/samena/secret/tests", args...)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	return cmd
}
