package testutil

import (
	"testing"

	pg_query "github.com/pganalyze/pg_query_go/v2"
	"github.com/stretchr/testify/require"
)

type RawStmt pg_query.RawStmt

func ParseSQL(t *testing.T, sql string) *RawStmt {
	tree, err := pg_query.Parse(sql)
	require.Nil(t, err)
	require.Equal(t, 1, len(tree.Stmts))
	return (*RawStmt)(tree.Stmts[0])
}
