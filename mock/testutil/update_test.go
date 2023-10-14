package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertUpdatedTable(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `UPDATE mytbl1 SET a1 = 1;`)
	stmt.AssertUpdatedTable(t, "mytbl1")
}

func TestGetUpdatedTable(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `UPDATE mytbl1 SET a1 = 1;`)
	tbl, err := stmt.GetUpdatedTable()
	require.NoError(t, err)
	assert.Equal(t, "mytbl1", tbl)

	stmt = ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5`)
	tbl, err = stmt.GetUpdatedTable()
	require.EqualError(t, err, "expected *pg_query.Node_UpdateStmt, got *pg_query.Node_SelectStmt")
	assert.Empty(t, tbl)
}

func TestGetUpdatedFields(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `UPDATE mytbl1 SET a1 = 1, b2 = 2, c3 = 3, d4 = 4;`)
	fields, err := stmt.GetUpdatedFields()
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "b2", "c3", "d4"}, fields)

	stmt = ParseSQL(t, `UPDATE mytbl1 SET xyz=1234;`)
	fields, err = stmt.GetUpdatedFields()
	require.NoError(t, err)
	assert.Equal(t, []string{"xyz"}, fields)

	stmt = ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5`)
	fields, err = stmt.GetUpdatedFields()
	require.EqualError(t, err, "expected *pg_query.Node_UpdateStmt, got *pg_query.Node_SelectStmt")
	assert.Empty(t, fields)
}

func TestAssertUpdatedFields(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `UPDATE mytbl1 SET a1 = 1, b2 = 2, c3 = 3, d4 = 4;`)
	stmt.AssertUpdatedFields(t, "a1", "b2", "c3", "d4")
}
