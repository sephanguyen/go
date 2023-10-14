package database

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs"
)

const mockTestingDir = "../../../mock/testing/testdata"

func loadAndCheckSyncTable(tableSchemaDir string, connectorConfigDir string) error {
	syncTableList, _ := loadSyncTableList(connectorConfigDir)

	for _, table := range syncTableList {
		if table.SinkTable == "user_basic_info" {
			fmt.Println(table)
		}
		sinkSchema, err := LoadTableSchema(tableSchemaDir, table.SinkDB, table.SinkTable, table.Schema)
		if err != nil {
			return err
		}

		sourceSchema, err := LoadTableSchema(tableSchemaDir, table.SourceDB, table.SourceTable, publicSchema)
		if err != nil {
			return err
		}

		_, err = CompareSchema(sinkSchema, sourceSchema, table.SinkDB, table.SourceDB)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestCompareFromConfig(t *testing.T) {
	connectorConfigDir := "../../../deployments/helm/manabie-all-in-one/charts/hephaestus/connectors/sink"
	err := loadAndCheckSyncTable(mockTestingDir, connectorConfigDir)
	if err != nil {
		t.Error(err)
	}
}

func TestCompareFromGeneratedConfig(t *testing.T) {
	baseConnectorConfigDir := "../../../deployments/helm/manabie-all-in-one/charts/hephaestus/generated_connectors/sink"

	orgDirs, err := os.ReadDir(baseConnectorConfigDir)
	if err != nil {
		t.Error(err)
	}

	for _, orgDir := range orgDirs {
		envDirs, err := os.ReadDir(path.Join(baseConnectorConfigDir, orgDir.Name()))
		if err != nil {
			t.Error(err)
		}
		for _, envDir := range envDirs {
			connectorConfigDir := path.Join(baseConnectorConfigDir, orgDir.Name(), envDir.Name())
			err := loadAndCheckSyncTable(mockTestingDir, connectorConfigDir)
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestCheckFkData(t *testing.T) {
	connectorConfigDir := "../../../deployments/helm/manabie-all-in-one/charts/hephaestus/connectors/sink"
	syncTableList, _ := loadSyncTableList(connectorConfigDir)
	for _, table := range syncTableList {
		sinkSchema, err := LoadTableSchema(mockTestingDir, table.SinkDB, table.SinkTable, table.Schema)
		if err != nil {
			fmt.Errorf("failed to load sink schema %s", err.Error())
			continue
		}

		fk := ""
		for _, valueSink := range sinkSchema.Constraint {
			if valueSink.ConstraintType == "FOREIGN KEY" {
				fk = fk + valueSink.ConstraintName + " "
			}
		}
		if len(fk) > 1 {
			fk = ", FK : " + fk
		}
		fmt.Println(table.SourceDB + "," + table.SinkDB + "," + table.SinkTable + fk)
	}
}

func loadSyncTableList(connectorConfigDir string) ([]SyncTable, error) {
	return LoadSyncTalbeFromDir(connectorConfigDir)
}

func skipSync(name string) bool {
	skipList := []string{
		"fatima_to_elastic_order_item.json",
		"fatima_to_elastic_order.json",
		"fatima_to_elastic_product.json",
		"sync_to_datalake.json",
	}
	return golibs.InArrayString(name, skipList)
}

func LoadSyncTalbeFromDir(dir string) ([]SyncTable, error) {
	var arr []SyncTable
	files, err := os.ReadDir(dir)
	// If upsert failed, then log the connector failed status code and continue
	if err != nil {
		fmt.Println("error loading file ", err)
		return nil, err
	}
	for _, file := range files {
		if skipSync(file.Name()) {
			continue
		}
		fileExt := filepath.Ext(file.Name())
		path := filepath.Join(dir, file.Name())
		// ignore directory, not json file and empty file
		if file.IsDir() || fileExt != JSONExt || isEmpty(path) {
			continue
		}
		syncInfo, err := buildSyncInfo(strings.ReplaceAll(file.Name(), JSONExt, ""))
		if err != nil || syncInfo.SinkDB == "elastic" {
			fmt.Println(" skip table ", syncInfo, err)
			continue
		}
		// fmt.Println("sync info : ", syncInfo)
		arr = append(arr, syncInfo)
	}
	return arr, nil
}
func buildSyncInfo(fileName string) (SyncTable, error) {
	reSource := regexp.MustCompile(`^([^_]*)_to_`)
	reSink := regexp.MustCompile(`_to_([^_]*)_`)
	reTableName := regexp.MustCompile(`_to_[^_]+_((.*)_sink_connector[_v\d]*|(.*))`)
	reVersion := regexp.MustCompile(`_v\d+$`)
	source := reSource.FindStringSubmatch(fileName)[1]
	sink := reSink.FindStringSubmatch(fileName)[1]
	tableNameResult := reTableName.FindStringSubmatch(fileName)
	tableName := ""
	schemaName := publicSchema
	for _, name := range tableNameResult[1:] {
		if name != "" {
			tableName = reVersion.ReplaceAllString(name, "")
		}
	}
	schemaTable := strings.Split(tableName, ".")
	if len(schemaTable) == 2 {
		schemaName, tableName = schemaTable[0], schemaTable[1]
	}
	return SyncTable{source, tableName, sink, tableName, schemaName}, nil
}

func buildSyncInfoFromConnectorName(connectorName string) SyncTable {
	reSource := regexp.MustCompile(`([^_]*)_to_`)
	reSink := regexp.MustCompile(`_to_([^_]*)_`)

	// table name may have 3 three format with suffix _sink_connector or _connector or nothing
	reTableName := regexp.MustCompile(`_to_[^_]+_((.*)_sink_connector|(.*)_connector|(.*))`)
	source := reSource.FindStringSubmatch(connectorName)[1]
	sink := reSink.FindStringSubmatch(connectorName)[1]
	tableNameResult := reTableName.FindStringSubmatch(connectorName)
	tableName := ""
	schemaName := publicSchema
	for _, name := range tableNameResult[1:] {
		if name != "" {
			tableName = name
		}
	}
	schemaTable := strings.Split(tableName, ".")
	if len(schemaTable) == 2 {
		schemaName, tableName = schemaTable[0], schemaTable[1]
	}

	return SyncTable{source, tableName, sink, tableName, schemaName}
}
