package testutil

import (
	"fmt"
	"testing"

	pg_query "github.com/pganalyze/pg_query_go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *RawStmt) AssertInsertedTable(t *testing.T, tableName string) {
	actualTableName, err := s.getInsertedTable()
	require.NoErrorf(t, err, "failed to get table name in insert into: %s", err)
	assert.Equal(t, tableName, actualTableName)
}

// Deprecated: use AssertInsertedTable instead
func (s *RawStmt) AssertInsertTo(t *testing.T, tableName string) {
	s.AssertInsertedTable(t, tableName)
}

func (s *RawStmt) getInsertedTable() (string, error) {
	_, ok := s.Stmt.GetNode().(*pg_query.Node_InsertStmt)
	if !ok {
		return "", fmt.Errorf("expected *pg_query.Node_InsertStmt, got %T", s.Stmt.GetNode())
	}
	return s.Stmt.GetInsertStmt().GetRelation().GetRelname(), nil
}

func (s *RawStmt) AssertInsertedFields(t *testing.T, expectedFields ...string) {
	actualFields, err := s.getInsertedFields()
	require.NoErrorf(t, err, "failed to get insert fields: %s", err)
	assert.Equal(t, expectedFields, actualFields)
}

// Deprecated: use AssertInsertedFields instead.
func (s *RawStmt) AssertInsertFields(t *testing.T, expectedFields ...string) {
	s.AssertInsertedFields(t, expectedFields...)
}

func (s *RawStmt) getInsertedFields() ([]string, error) {
	_, ok := s.Stmt.GetNode().(*pg_query.Node_InsertStmt)
	if !ok {
		return nil, fmt.Errorf("expected *pg_query.Node_InsertStmt, got %T", s.Stmt.GetNode())
	}
	cols := s.Stmt.GetInsertStmt().GetCols()
	res := make([]string, 0, len(cols))
	for _, c := range cols {
		name := c.GetResTarget().GetName()
		res = append(res, name)
	}
	return res, nil
}
