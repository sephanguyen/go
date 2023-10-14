package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/go-kafka/connect"
	dplparser "github.com/manabie-com/backend/cmd/utils/data_pipeline_parser"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
)

var (
	generatedConnectorDir  = "../../../deployments/helm/manabie-all-in-one/charts/hephaestus/generated_connectors/sink/%s/%s"
	definedDatapipelineDir = "../../../deployments/helm/platforms/kafka-connect/postgresql2postgresql"
	tableSchemaDir         = "../../../mock/testing/testdata"
)

func getCurrentColumnFromSinkTableInGeneratedFile(org, env string, sinkFileName string) ([]string, error) {
	// work around for dir name
	// due older code only verify local manabie connector
	// now we have to verify first env manabie instead
	dir := fmt.Sprintf(generatedConnectorDir, org, env)
	b, err := os.ReadFile(path.Join(dir, sinkFileName))
	if err != nil {
		return nil, err
	}
	connect := connect.Connector{}
	err = json.Unmarshal(b, &connect)
	if err != nil {
		return nil, err
	}

	columns := connect.Config["fields.whitelist"]
	return strings.Split(columns, ","), nil
}

func getAllColumnsFromSourceTable(dbName, tableName, schema string) ([]string, error) {
	tableSchema, err := database.LoadTableSchema(tableSchemaDir, dbName, tableName, schema)
	if err != nil {
		return nil, err
	}

	cols := make([]string, 0)
	for _, col := range tableSchema.Schema {
		cols = append(cols, col.ColumnName)
	}

	return cols, nil
}

func compareExcludeColumns(current, expect []string) error {
	sort.Strings(current)
	sort.Strings(expect)
	if len(current) != len(expect) {
		return fmt.Errorf("not define exclude columns expect %s but got %s", expect, current)
	}

	for i := 0; i < len(current); i++ {
		if current[i] != expect[i] {
			return fmt.Errorf("exclude column not match expect %s but got %s", expect, current)
		}
	}

	return nil
}

func getOrgByEnv(env string, orgs []string) string {
	if len(orgs) == 1 {
		return orgs[0]
	} else if env == "local" || env == "stag" || env == "uat" {
		return "manabie"
	}
	return orgs[0]
}

func TestCompareSchemaWithExcludeColumn(t *testing.T) {
	es, err := os.ReadDir(definedDatapipelineDir)
	if err != nil {
		t.Error(err)
		return
	}
	for _, e := range es {
		filePath := path.Join(definedDatapipelineDir, e.Name())
		dpl, err := dplparser.NewDataPipelineParser(filePath)
		if err != nil {
			t.Error(err)
			return
		}
		for _, pl := range dpl.DataPipelineDef.Datapipelines {
			dbName := dpl.DataPipelineDef.Database

			if err != nil {
				t.Error(err)
				return
			}
			for _, sink := range pl.Sinks {
				for _, schema := range sink.DeploySchemas {
					if schema == "public" {
						sink.FileName = strings.ReplaceAll(sink.FileName, "SCHEMA.", "")
					} else {
						sink.FileName = strings.ReplaceAll(sink.FileName, "SCHEMA", schema)
					}
					sourceColumns, err := getAllColumnsFromSourceTable(sink.Database, pl.Table, schema)
					firstEnv := sink.DeployEnvs[0]
					org := getOrgByEnv(firstEnv, sink.DeployOrgs)
					sinkColumns, err := getCurrentColumnFromSinkTableInGeneratedFile(org, firstEnv, sink.FileName)
					if err != nil {
						t.Error(err)
						return
					}

					excludeColumns := make([]string, 0)
					for i := 0; i < len(sinkColumns); i++ {
						if !golibs.InArrayString(sinkColumns[i], sourceColumns) {
							excludeColumns = append(excludeColumns, sinkColumns[i])
						}
					}

					excludeColumnsField := sink.ExcludeColumns

					err = compareExcludeColumns(excludeColumnsField, excludeColumns)
					if err != nil {
						t.Errorf("sync from %s.%s to %s.%s mismatch exclude columns, err: %s", dbName, pl.Table, sink.Database, sink.Table, err.Error())
						return
					}
				}
			}
		}
	}
}
