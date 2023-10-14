package testutil

import (
	"fmt"
	"testing"

	pg_query "github.com/pganalyze/pg_query_go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *RawStmt) AssertUpdatedTable(t *testing.T, tableName string) {
	actualTableName, err := s.GetUpdatedTable()
	require.NoErrorf(t, err, "failed to get table name in update: %s", err)
	assert.Equal(t, tableName, actualTableName)
}

// Deprecated: use AssertUpdatedTable instead.
func (s *RawStmt) AssertUpdateTo(t *testing.T, tableName string) {
	s.AssertUpdatedTable(t, tableName)
}

func (s *RawStmt) GetUpdatedTable() (string, error) {
	_, ok := s.Stmt.GetNode().(*pg_query.Node_UpdateStmt)
	if !ok {
		return "", fmt.Errorf("expected *pg_query.Node_UpdateStmt, got %T", s.Stmt.GetNode())
	}
	return s.Stmt.GetUpdateStmt().GetRelation().GetRelname(), nil
}

func (s *RawStmt) MustGetUpdatedTable() string {
	tableName, err := s.GetUpdatedTable()
	if err != nil {
		panic(err)
	}
	return tableName
}

func (s *RawStmt) AssertUpdatedFields(t *testing.T, fields ...string) {
	actualFields, err := s.GetUpdatedFields()
	require.NoErrorf(t, err, "failed to get update fields: %s", err)
	assert.Equal(t, fields, actualFields)
}

// Deprecated: use AssertUpdatedFields instead.
func (s *RawStmt) AssertUpdateFields(t *testing.T, fields ...string) {
	s.AssertUpdatedFields(t, fields...)
}

func (s *RawStmt) GetUpdatedFields() ([]string, error) {
	_, ok := s.Stmt.GetNode().(*pg_query.Node_UpdateStmt)
	if !ok {
		return nil, fmt.Errorf("expected *pg_query.Node_UpdateStmt, got %T", s.Stmt.GetNode())
	}
	cols := s.Stmt.GetUpdateStmt().GetTargetList()
	res := make([]string, 0, len(cols))
	for _, v := range cols {
		colname := v.GetResTarget().GetName()
		res = append(res, colname)
	}
	return res, nil
}

func (s *RawStmt) MustGetUpdatedFields() []string {
	fields, err := s.GetUpdatedFields()
	if err != nil {
		panic(err)
	}
	return fields
}
