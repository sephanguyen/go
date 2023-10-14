package main

import (
	"os"

	"github.com/manabie-com/backend/cmd/samena/secret"
	"github.com/manabie-com/backend/cmd/samena/sql"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "samena subcommand",
		Short: "CLI for working with Manabie's backend",
	}
	rootCmd.AddCommand(secret.NewCmdSecret())
	rootCmd.AddCommand(sql.NewCmdCloudSQLProxy())
	rootCmd.AddCommand(sql.NewCmdPsql())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
