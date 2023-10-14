package database

import (
	"fmt"
	"strings"

	fileio "github.com/manabie-com/backend/internal/golibs/io"

	"go.uber.org/multierr"
)

type SinkTable struct {
	SrcDB     string
	DesDB     string
	TableName string
}

func splitFileName(filename string) []string {
	return strings.Split(filename, "_")
}

func removeExtJSON(filename string) string {
	return strings.ReplaceAll(filename, JSONExt, "")
}

func newSinkTable(filename string) SinkTable {
	strArr := splitFileName(filename)
	tableName := removeExtJSON(strings.Join(strArr[3:], "_"))
	return SinkTable{SrcDB: strArr[0], DesDB: strArr[2], TableName: tableName}
}

func validSinkTable(filename string) bool {
	return len(splitFileName(filename)) >= 4
}

func getSinkTables(connectorDir string) ([]SinkTable, error) {
	files, err := fileio.GetFileNamesOnDir(connectorDir)
	if err != nil {
		return nil, err
	}

	sinkTables := []SinkTable{}
	for _, file := range files {
		if validSinkTable(file) {
			sinkTables = append(sinkTables, newSinkTable(file))
		}
	}

	return sinkTables, nil
}

func getSinkTablesBy(connectorDir string, svc string) ([]SinkTable, error) {
	sinkTables, err := getSinkTables(connectorDir)
	if err != nil {
		return nil, err
	}

	desSinkTables := []SinkTable{}
	for _, sinkTable := range sinkTables {
		if sinkTable.DesDB == svc {
			desSinkTables = append(desSinkTables, sinkTable)
		}
	}
	return desSinkTables, nil
}

func verifySinkTable(sinkTable SinkTable, groupedAC map[string]map[string]bool) error {
	if acTables, ok := groupedAC[sinkTable.SrcDB]; ok && acTables[sinkTable.TableName] {
		if desACTables, ok := groupedAC[sinkTable.DesDB]; !ok || !desACTables[sinkTable.TableName] {
			return fmt.Errorf("table %s in db %s missing AC ", sinkTable.TableName, sinkTable.DesDB)
		}
	}
	return nil
}

func VerifyACForAllSinkTable() error {
	sinkDir := "../../../deployments/helm/manabie-all-in-one/charts/hephaestus/generated_connectors/sink/manabie/local"
	sinkTables, err := getSinkTables(sinkDir)
	if err != nil {
		return err
	}

	groupedAC, err := groupACByServiceAndTable(fileACStageDir)
	if err != nil {
		return err
	}

	var errs error
	for _, sinkTable := range sinkTables {
		err := verifySinkTable(sinkTable, groupedAC)
		errs = multierr.Append(errs, err)
	}
	return errs
}
