package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertWhereConditions_Select(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `SELECT * FROM tbl WHERE (a1 IS NULL OR a1 = $1) AND (b2 IS NOT NULL) AND c3 = $2`)
	stmt.AssertWhereConditions(t, map[string]*CheckWhereClauseOpt{
		"a1": {
			HasNullTest: true,
			EqualExpr: &EqualExpr{
				IndexArg: 1,
			},
		},
		"b2": {
			HasNullTest: true,
		},
		"c3": {
			EqualExpr: &EqualExpr{
				IndexArg: 2,
			},
		},
	})
}

func TestAssertWhereConditions_Update(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `UPDATE e5 SET a1 = 1 WHERE b2 = $2 AND c3 IS NOT NULL`)
	stmt.AssertWhereConditions(t, map[string]*CheckWhereClauseOpt{
		"b2": {
			EqualExpr: &EqualExpr{
				IndexArg: 2,
			},
		},
		"c3": {
			HasNullTest: true,
		},
	})
}

func TestGetWhereConditions_Select_AExpr(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5 WHERE a1 = $1`)
	cond, err := stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"a1": {
			EqualExpr: &EqualExpr{
				IndexArg: 1,
			},
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5 WHERE b2 = true`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"b2": {
			EqualExpr: &EqualExpr{
				Type:  "bool",
				Value: true,
			},
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT * FROM e5 WHERE NOW() BETWEEN a1 AND b2`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"now": {
			BetweenExpr: &BetweenExpr{
				Field: "now",
				Args:  []string{"a1", "b2"},
			},
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT * FROM e5 WHERE a1 IS NOT NULL`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"a1": {
			HasNullTest: true,
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT * FROM e5 WHERE $1::int4 = 0`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"$1": {
			EqualExpr: &EqualExpr{
				Type:  "int32",
				Value: int32(0),
			},
		},
	}, cond)
}

func TestGetWhereConditions_Select_BoolExpr(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `SELECT a1, b2, c3, d4 FROM e5 WHERE a1 = $1 AND b2 = true AND c3 = 'a_string'`)
	cond, err := stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"a1": {
			EqualExpr: &EqualExpr{
				IndexArg: 1,
			},
		},
		"b2": {
			EqualExpr: &EqualExpr{
				Type:  "bool",
				Value: true,
			},
		},
		"c3": {
			EqualExpr: &EqualExpr{
				Type:  "string",
				Value: "a_string",
			},
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT * FROM e5 WHERE $1::int[] IS NULL OR a1 = ANY($1)`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"$1": {HasNullTest: true},
		"a1": {HasNullTest: true, EqualExpr: &EqualExpr{IndexArg: 1}},
	}, cond)
}

func TestGetWhereConditions_Update_AExpr(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `UPDATE e5 SET a1 = 1 WHERE b2 = $2`)
	cond, err := stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"b2": {
			EqualExpr: &EqualExpr{
				IndexArg: 2,
			},
		},
	}, cond)

	stmt = ParseSQL(t, `UPDATE e5 SET a1 = 1 WHERE b2 = true`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"b2": {
			EqualExpr: &EqualExpr{
				Type:  "bool",
				Value: true,
			},
		},
	}, cond)
}

func TestGetWhereConditions_Update_BoolExpr(t *testing.T) {
	t.Parallel()
	stmt := ParseSQL(t, `UPDATE e5 SET a1 = 1 WHERE b3 = $1 AND c4 = true AND d5 = 'another_string'`)
	cond, err := stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"b3": {
			EqualExpr: &EqualExpr{
				IndexArg: 1,
			},
		},
		"c4": {
			EqualExpr: &EqualExpr{
				Type:  "bool",
				Value: true,
			},
		},
		"d5": {
			EqualExpr: &EqualExpr{
				Type:  "string",
				Value: "another_string",
			},
		},
	}, cond)
}

func TestGetWhereConditions_BackwardCompatibility(t *testing.T) {
	t.Parallel()

	stmt := ParseSQL(t, `SELECT user_id,user_group,country,name,given_name,avatar,phone_number,email,device_token,allow_notification,updated_at,created_at,is_tester,facebook_id,phone_verified,email_verified,deleted_at,resource_path,last_login_date,birthday,gender FROM users  WHERE (email = ANY($1))`)
	cond, err := stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"email": {
			EqualExpr: &EqualExpr{
				IndexArg: 1,
			},
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT * FROM tablename WHERE deleted_at IS NULL AND ($1::text[] IS NULL OR user_id = ANY($1))`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"deleted_at": {
			HasNullTest: true,
		},
		"user_id": {
			HasNullTest: true,
			EqualExpr: &EqualExpr{
				IndexArg: 1,
			},
		},
		"$1": {
			HasNullTest: true,
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT * FROM e5, e6 WHERE e5.a1 IS NOT NULL AND e6.a1 IS NOT NULL`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"a1": {
			HasNullTest: true,
		},
	}, cond)

	stmt = ParseSQL(t, `SELECT school_id, name, country, city_id, district_id, point, is_system_school, is_merge, phone_number, created_at, updated_at FROM schools
						WHERE country = $1
						AND ($2::int = 0 OR city_id = $2)
						AND ($3::int = 0 OR district_id = $3)
						AND ($4::boolean = false OR is_system_school = $4)
						AND ($5::int[] IS NULL OR school_id = ANY($5))`)
	cond, err = stmt.getWhereConditions()
	require.NoError(t, err)
	assert.Equal(t, map[string]*CheckWhereClauseOpt{
		"country":          {EqualExpr: &EqualExpr{IndexArg: 1}},
		"city_id":          {EqualExpr: &EqualExpr{IndexArg: 2}},
		"district_id":      {EqualExpr: &EqualExpr{IndexArg: 3}},
		"is_system_school": {EqualExpr: &EqualExpr{IndexArg: 4}},
		"school_id":        {HasNullTest: true, EqualExpr: &EqualExpr{IndexArg: 5}},
		"$2":               {EqualExpr: &EqualExpr{Type: "int32", Value: int32(0)}},
		"$3":               {EqualExpr: &EqualExpr{Type: "int32", Value: int32(0)}},
		"$4":               {EqualExpr: &EqualExpr{Type: "bool", Value: false}},
		"$5":               {HasNullTest: true},
	}, cond)
}
