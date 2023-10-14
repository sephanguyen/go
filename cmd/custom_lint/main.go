package main

import (
	"github.com/manabie-com/backend/cmd/custom_lint/sqlclosecheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(sqlclosecheck.NewAnalyzer())
}
