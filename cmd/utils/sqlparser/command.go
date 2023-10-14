package sqlparser

import (
	"github.com/spf13/cobra"
)

// SqlParserCmd parses a SQL file
var SqlParserCmd = &cobra.Command{
	Use:   "parse",
	Short: "parse a sql file",
	RunE:  runSqlParser,
}

// RootCmd for sql
var RootCmd = &cobra.Command{
	Use:   "sql",
	Short: "interact with a sql file",
}

func init() {
	RootCmd.AddCommand(
		SqlParserCmd,
	)

	SqlParserCmd.PersistentFlags().StringVarP(
		&sqlFilePath,
		"filepath",
		"f",
		"",
		"SQL file path",
	)
}
