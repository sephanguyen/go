package testutil

import (
	"fmt"
	"testing"

	pg_query "github.com/pganalyze/pg_query_go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *RawStmt) AssertSelectedTable(t *testing.T, tableName, alias string) {
	actualTableName, actualAlias, err := s.getSelectedTable()
	require.NoErrorf(t, err, "failed to get table name in from clause: %s", err)
	assert.Equal(t, tableName, actualTableName)
	assert.Equal(t, alias, actualAlias)
}

// Deprecated: use AssertSelectedTable instead.
func (s *RawStmt) AssertFromClause(t *testing.T, tableName, alias string) {
	s.AssertSelectedTable(t, tableName, alias)
}

func (s *RawStmt) getSelectedTable() (string, string, error) {
	_, ok := s.Stmt.GetNode().(*pg_query.Node_SelectStmt)
	if !ok {
		return "", "", fmt.Errorf("expected *pg_query.Node_SelectStmt, got %T", s.Stmt.GetNode())
	}
	list := s.Stmt.GetSelectStmt().GetFromClause()
	if len(list) != 1 {
		return "", "", fmt.Errorf("expected one from clause, got %d", len(list))
	}
	v := list[0].GetRangeVar()
	if v == nil {
		return "", "", fmt.Errorf("expected range var, got %T", list[0])
	}
	return v.GetRelname(), v.GetAlias().GetAliasname(), nil
}

func (s *RawStmt) AssertSelectedFields(t *testing.T, expectedFields ...string) {
	actualFields, err := s.getSelectedFields()
	require.NoErrorf(t, err, "failed to get selected fields: %s", err)
	assert.Equal(t, expectedFields, actualFields)
}

// Deprecated: use AssertSelectedFields instead.
func (s *RawStmt) AssertSelectField(t *testing.T, expectedFields ...string) {
	s.AssertSelectedFields(t, expectedFields...)
}

func (s *RawStmt) getSelectedFields() ([]string, error) {
	_, ok := s.Stmt.GetNode().(*pg_query.Node_SelectStmt)
	if !ok {
		return nil, fmt.Errorf("expected *pg_query.Node_SelectStmt, got %T", s.Stmt.GetNode())
	}
	list := s.Stmt.GetSelectStmt().GetTargetList()
	res := make([]string, 0, len(list))
	for _, v := range list {
		v2 := v.GetResTarget().GetVal().GetColumnRef().GetFields()
		if len(v2) > 2 {
			return nil, fmt.Errorf("expected one (column name) or two (table name + column name) fields, got %d (%+v)", len(v2), v2)
		}
		v3 := v2[len(v2)-1].GetString_().GetStr()
		res = append(res, v3)
	}
	return res, nil
}
