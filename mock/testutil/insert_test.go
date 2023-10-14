package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertInsertedTable(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `INSERT INTO ee (a1, b2, c3, d4) VALUES (1, 2, 3, 4)`)
	stmt.AssertInsertedTable(t, "ee")
}

func TestGetInsertedTable(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `INSERT INTO e5 (a1, b2, c3, d4) VALUES (1, 2, 3, 4)`)
	tbl, err := stmt.getInsertedTable()
	require.NoError(t, err)
	assert.Equal(t, "e5", tbl)

	stmt = ParseSQL(t, `UPDATE e5 SET a1 = 1`)
	tbl, err = stmt.getInsertedTable()
	require.EqualError(t, err, "expected *pg_query.Node_InsertStmt, got *pg_query.Node_UpdateStmt")
	assert.Empty(t, tbl)
}

func TestAssertInsertedFields(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `INSERT INTO ee (a1, b2, c3, d4) VALUES (1, 2, 3, 4)`)
	stmt.AssertInsertedFields(t, "a1", "b2", "c3", "d4")
}

func TestGetInsertedFields(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `INSERT INTO e5 (a1, b2, c3, d4) VALUES (1, 2, 3, 4)`)
	fields, err := stmt.getInsertedFields()
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "b2", "c3", "d4"}, fields)

	stmt = ParseSQL(t, `UPDATE e5 SET a1 = 1`)
	fields, err = stmt.getInsertedFields()
	require.EqualError(t, err, "expected *pg_query.Node_InsertStmt, got *pg_query.Node_UpdateStmt")
	assert.Nil(t, fields)
}
