package sqlparser

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	sqlFilePath string
)

func runSqlParser(cmd *cobra.Command, args []string) error {
	if len(sqlFilePath) == 0 {
		return errors.New("filepath is required")
	}
	return parse(sqlFilePath)
}

func parse(filePath string) error {
	parser := NewSqlParser()
	queries, err := parser.ParseFromFile(filePath)
	if err != nil {
		return err
	}

	tables := make(map[string]bool)
	functions := make(map[string]bool)
	procedures := make(map[string]bool)
	for _, query := range queries {
		if query.Entity == TableEntity || query.Entity == IndexEntity {
			tables[query.Name] = true
		}
		if query.Entity == FunctionEntity {
			functions[query.Name] = true
		}
		if query.Entity == ProcedureEntity {
			procedures[query.Name] = true
		}
	}

	var (
		tableNames     = []string{}
		functionNames  = []string{}
		procedureNames = []string{}
	)
	for name := range tables {
		tableNames = append(tableNames, name)
	}
	for name := range functions {
		functionNames = append(functionNames, name)
	}
	for name := range procedures {
		procedureNames = append(procedureNames, name)
	}

	printToTerminal(tableNames)
	printToTerminal(functionNames)
	printToTerminal(procedureNames)
	return nil
}

func printToTerminal(names []string) {
	bNames, err := json.Marshal(names)
	if err == nil {
		fmt.Println(string(bNames))
	}
}
