package database

import (
	"fmt"
	"reflect"

	"github.com/jackc/pgtype"
)

// fieldSchema implements database.Entity and contains information about a column's name
// and datatype in a table.	 It also implements database.Entity interface.
type fieldSchema struct {
	FieldName     pgtype.Text `json:"column_name"`
	TypeName      pgtype.Text `json:"data_type"`
	ColumnDefault pgtype.Text `json:"column_default"`
	IsNullable    pgtype.Text `json:"is_nullable"`
}

// fieldConstraint implements database.Entity and contains information about a table's constrains
type fieldConstraint struct {
	ConstraintName pgtype.Text `json:"constraint_name"`
	ColumnName     pgtype.Text `json:"column_name"`
	ConstraintType pgtype.Text `json:"constraint_type"`
}

// FieldMap returns fields' names and values of ColumnSchema entity.
func (fs *fieldSchema) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"column_name", "data_type", "column_default", "is_nullable"}
	values = []interface{}{&fs.FieldName, &fs.TypeName, &fs.ColumnDefault, &fs.IsNullable}
	return
}

func (fs *fieldConstraint) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"constraint_name", "column_name", "constraint_type"}
	values = []interface{}{&fs.ConstraintName, &fs.ColumnName, &fs.ConstraintType}
	return
}

// TableName returns the table name of ColumnSchema entity.
func (fs *fieldSchema) TableName() string {
	return "information_schema.columns"
}

// TableName returns the table name of table_constraints entity.
func (fs *fieldConstraint) TableName() string {
	return "information_schema.table_constraints"
}

// fieldSchemas implements database.Entities for fieldSchema struct.
type fieldSchemas []*fieldSchema

// tableConstraints implements database.Entities for tableConstraint struct.
type fieldConstraints []*fieldConstraint

// Add adds to FieldSchemas and returns a new FieldSchema entity.
func (u *fieldSchemas) Add() Entity {
	e := &fieldSchema{}
	*u = append(*u, e)
	return e
}

// Add adds to FieldSchemas and returns a new FieldSchema entity.
func (u *fieldConstraints) Add() Entity {
	e := &fieldConstraint{}
	*u = append(*u, e)
	return e
}

// table implements database.Entity and is used to query all user-created table names.
type table struct {
	Name   pgtype.Text
	Type   pgtype.Text
	Owner  pgtype.Text
	Schema pgtype.Text
}

// FieldMap returns fields' names and values of table entity.
func (t *table) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"table_name", "table_type", "owner", "table_schema"}
	values = []interface{}{&t.Name, &t.Type, &t.Owner, &t.Schema}
	return
}

// TableName returns the table name of table entity.
func (t *table) TableName() string {
	return "information_schema.tables"
}

// tables implements database.Entities for table struct.
type tables []*table

// tableSchema represents a table schema in a database. It contains information
// about every column's name and datatype in a table.
type tableSchema struct {
	Schema     []*fieldSchema     `json:"schema"`
	Policies   []*tablePolicy     `json:"policies"`
	Constraint []*fieldConstraint `json:"constraint"`
	TableName  string             `json:"table_name"`
	Type       string             `json:"type"`
	Owner      string             `json:"owner"`
}

// Add adds to tables and returns a new table entity.
func (u *tables) Add() Entity {
	e := &table{}
	*u = append(*u, e)
	return e
}

// matchEntity checks if e's fields are a subset of current TableSchema's fields, i.e. all fields from e
// must exist in TableSchema and the type must matchEntity (conversely, fields from TableSchema can be missing from e).
func (ts *tableSchema) matchEntity(e Entity) (bool, error) {
	if ts.TableName != e.TableName() {
		return false, fmt.Errorf("table name mismatched (%s vs %s)", ts.TableName, e.TableName())
	}

	// O(N^2) here because map is memory-expensive
	names, values := e.FieldMap()
outerLoop:
	for i := range names {
		for j := range ts.Schema {
			if names[i] != ts.Schema[j].FieldName.String {
				continue
			}

			// names match, check for datatype
			if !ts.checkType(values[i], ts.Schema[j].TypeName.String) {
				return false, fmt.Errorf(`data type mismatched (%s vs %s for field %s)`, reflect.TypeOf(values[i]), ts.Schema[j].TypeName.String, names[i])
			}
			continue outerLoop
		}

		// reaching here means no matches
		return false, fmt.Errorf("entity field %s is missing from table %s", names[i], ts.TableName)
	}

	return true, nil
}

// checkType returns whether v is of type require in PostgreSQL.
//
// For example, if v is of type pgtype.Bool, then require should be boolean.
func (*tableSchema) checkType(v interface{}, require string) bool {
	t := reflect.TypeOf(v)
	lookup, ok := typemap[t]
	if !ok {
		panic(fmt.Errorf("missing type %s in psql typemap", t))
	}
	return lookup(require)
}
