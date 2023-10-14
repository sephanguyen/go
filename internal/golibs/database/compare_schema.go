package database

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const JSONExt = ".json"
const publicSchema = "public"

type Column struct {
	ColumnName    string      `json:"column_name"`
	DataType      string      `json:"data_type"`
	ColumnDefault interface{} `json:"column_default"`
	IsNullable    string      `json:"is_nullable"`
}

type FieldConstraint struct {
	ConstraintName string `json:"constraint_name"`
	ColumName      string `json:"column_name"`
	ConstraintType string `json:"constraint_type"`
}

type TableSchema struct {
	Schema     []Column          `json:"schema"`
	Constraint []FieldConstraint `json:"constraint"`
	TableName  string            `json:"table_name"`
}

type SyncTable struct {
	SourceDB    string
	SourceTable string
	SinkDB      string
	SinkTable   string
	Schema      string
}

// compareSchema compare columns of sink table to columns of source table
func CompareSchema(sinkSchema, sourceSchema *TableSchema, sinkName string, sourceName string) (*SyncTable, error) {
	tableName := sinkSchema.TableName
	for _, sinkCol := range sinkSchema.Schema {
		found := false
		for _, sourceCol := range sourceSchema.Schema {
			if sinkCol.ColumnName == sourceCol.ColumnName {
				found = true
				if sinkCol.DataType != sourceCol.DataType {
					return nil, fmt.Errorf("column %s.%s type is not match [%s] [%s] sink : %s, source : %s", tableName, sinkCol.ColumnName, sinkCol.DataType, sourceCol.DataType, sinkName, sourceName)
				}

				if sinkCol.IsNullable == "NO" && sourceCol.IsNullable == "YES" {
					return nil, fmt.Errorf("column %s.%s nullable is not match [%s] [%s] sink : %s, source : %s", tableName, sinkCol.ColumnName, sinkCol.IsNullable, sourceCol.IsNullable, sinkName, sourceName)
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("not found column %s.%s in source table. sink : %s, source : %s", tableName, sinkCol.ColumnName, sinkName, sourceName)
		}
	}

	exists := make(map[string]string)

	for _, value := range sourceSchema.Constraint {
		exists[value.ColumName+"_"+value.ConstraintType] = value.ConstraintType
	}

	existsSink := make(map[string]string)
	for _, valueSink := range sinkSchema.Constraint {
		existsSink[valueSink.ColumName+"_"+valueSink.ConstraintType] = valueSink.ConstraintType
		val, ok := exists[valueSink.ColumName+"_"+valueSink.ConstraintType]
		if valueSink.ConstraintType == "FOREIGN KEY" {
			continue
		}
		if !ok || val != valueSink.ConstraintType {
			return nil, fmt.Errorf("mismatched constraint %s in table: %s sink : %s, source: %s", valueSink.ColumName, tableName, sinkName, sourceName)
		}
	}

	for _, value := range sourceSchema.Constraint {
		val, ok := existsSink[value.ColumName+"_"+value.ConstraintType]
		if value.ConstraintType != "PRIMARY KEY" {
			continue
		}
		if !ok || val != value.ConstraintType {
			return nil, fmt.Errorf("missing primary key  %s in table: %s sink : %s, source: %s", value.ColumName, tableName, sinkName, sourceName)
		}
	}
	var ecap = SyncTable{SourceDB: dbname, SinkTable: sinkSchema.TableName, SourceTable: sourceSchema.TableName}
	return &ecap, nil
}

func isEmpty(path string) bool {
	b, _ := os.ReadFile(path)
	return len(strings.TrimSpace(string(b))) == 0
}

func LoadTableSchema(dir, dbname, tablename, schemaName string) (*TableSchema, error) {
	if schemaName == publicSchema {
		schemaName = ""
	} else {
		schemaName = fmt.Sprintf("%s.", schemaName)
	}
	path := fmt.Sprintf("%s/%s/%s%s.json", dir, dbname, schemaName, tablename)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	schema := &TableSchema{}
	err = json.Unmarshal(b, schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
