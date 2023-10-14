package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertSelectFields(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5`)
	stmt.AssertSelectedFields(t, "a1", "b2", "c3", "d4")

	stmt = ParseSQL(t, `SELECT a11, b22, c33, d44 FROM e5 f6`)
	stmt.AssertSelectedFields(t, "a11", "b22", "c33", "d44")
}

func TestGetSelectedFields(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5`)
	fields, err := stmt.getSelectedFields()
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "b2", "c3", "d4"}, fields)

	stmt = ParseSQL(t, `UPDATE e5 SET a1 = 1`)
	fields, err = stmt.getSelectedFields()
	require.EqualError(t, err, "expected *pg_query.Node_SelectStmt, got *pg_query.Node_UpdateStmt")
	assert.Nil(t, fields)

	stmt = ParseSQL(t, `SELECT t1.a1, t1.a2, t1.a3 FROM tablename t1`)
	fields, err = stmt.getSelectedFields()
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "a2", "a3"}, fields)
}

func TestAssertSelectTable(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5 f6`)
	stmt.AssertSelectedTable(t, "e5", "f6")

	stmt = ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5`)
	stmt.AssertSelectedTable(t, "e5", "")
}

func TestGetSelectedTable(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5 f6`)
	tbl, alias, err := stmt.getSelectedTable()
	require.NoError(t, err)
	assert.Equal(t, "e5", tbl)
	assert.Equal(t, "f6", alias)

	stmt = ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5`)
	tbl, alias, err = stmt.getSelectedTable()
	require.NoError(t, err)
	assert.Equal(t, "e5", tbl)
	assert.Equal(t, "", alias)

	stmt = ParseSQL(t, `UPDATE e5 SET a1 = 1`)
	tbl, alias, err = stmt.getSelectedTable()
	require.EqualError(t, err, "expected *pg_query.Node_SelectStmt, got *pg_query.Node_UpdateStmt")
	assert.Empty(t, tbl)
	assert.Empty(t, alias)
}
