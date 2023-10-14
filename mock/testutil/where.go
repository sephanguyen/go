package testutil

import (
	"fmt"
	"regexp"
	"testing"

	pg_query "github.com/pganalyze/pg_query_go/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EqualExpr struct {
	ColumnRef string
	IndexArg  int
	Type      string // bool
	Value     interface{}
}

type BetweenExpr struct {
	Field string
	Args  []string
}

type CheckWhereClauseOpt struct {
	HasNullTest bool
	BetweenExpr *BetweenExpr
	EqualExpr   *EqualExpr
}

func MergeWhereConditions(c1, c2 map[string]*CheckWhereClauseOpt) map[string]*CheckWhereClauseOpt {
	res := make(map[string]*CheckWhereClauseOpt)
	for k, v := range c1 {
		res[k] = v
	}
	for k, v := range c2 {
		if _, exist := res[k]; !exist {
			res[k] = v
		} else {
			res[k].HasNullTest = res[k].HasNullTest || v.HasNullTest
			if v.BetweenExpr != nil {
				if res[k].BetweenExpr != nil {
					panic("multiple between expressions")
				} else {
					res[k].BetweenExpr = v.BetweenExpr
				}
			}
			if v.EqualExpr != nil {
				if res[k].EqualExpr != nil {
					panic("multiple equal expressions")
				} else {
					res[k].EqualExpr = v.EqualExpr
				}
			}
		}
	}
	return res
}

// AssertWhereConditions checks all the conditions in opts are satisfied. opts can contain fewer
// conditions than the ones in the where clause. As a result, AssertWhereConditions(t, nil) never fails.
func (s *RawStmt) AssertWhereConditions(t *testing.T, opts map[string]*CheckWhereClauseOpt) {
	conditions, err := s.getWhereConditions()
	require.NoError(t, err, "failed to get where conditions: %s", err)
	for fieldName := range opts {
		require.Contains(t, conditions, fieldName, "field %q not found in where clause", fieldName)
		assert.Equalf(t, opts[fieldName], conditions[fieldName], "conditions for field %q does not match (expected %+v, got %+v)", fieldName, opts[fieldName], conditions[fieldName])
	}
}

// Deprecated: use AssertWhereConditions instead.
func (s *RawStmt) AssertWhereClause(t *testing.T, opts map[string]*CheckWhereClauseOpt) {
	s.AssertWhereClause(t, opts)
}

func (s *RawStmt) getWhereConditions() (map[string]*CheckWhereClauseOpt, error) {
	// Extract the where clause from the statement
	var whereClause *pg_query.Node
	switch s.Stmt.GetNode().(type) {
	case *pg_query.Node_SelectStmt:
		whereClause = s.Stmt.GetSelectStmt().GetWhereClause()
	case *pg_query.Node_UpdateStmt:
		whereClause = s.Stmt.GetUpdateStmt().GetWhereClause()
	default:
		return nil, fmt.Errorf("unexpected statement type %T", s.Stmt.GetNode())
	}
	if whereClause == nil {
		return nil, nil
	}

	// Parse the where clause
	switch whereClause := whereClause.GetNode().(type) {
	case *pg_query.Node_AExpr: // single condition, for example: a = $1, b = true
		return s.extractConditionsFromAExpr(whereClause.AExpr)
	case *pg_query.Node_BoolExpr: // multiple conditions with AND/OR, for example: a = $1 AND b = true
		return s.extractConditionFromBoolExpr(whereClause.BoolExpr)
	case *pg_query.Node_NullTest: // null test, for example: a IS NOT NULL
		return s.extractConditionsFromNullTest(whereClause.NullTest)
	default:
		return nil, fmt.Errorf("unexpected where clause type %T", whereClause)
	}
}

// extractConditionsFromAExpr returns the conditions in the input.
// pg_query.A_Expr contains a single condition.
func (s *RawStmt) extractConditionsFromAExpr(aExpr *pg_query.A_Expr) (map[string]*CheckWhereClauseOpt, error) {
	res := make(map[string]*CheckWhereClauseOpt)
	comparsion := aExpr.GetName()[0].GetString_().GetStr()
	rExpr := aExpr.GetRexpr()
	lExpr := aExpr.GetLexpr()
	switch comparsion {
	case "=":
		var opt EqualExpr
		switch rExpr.GetNode().(type) {
		case *pg_query.Node_ColumnRef:
			v := rExpr.GetColumnRef().GetFields()
			if len(v) > 2 {
				return nil, fmt.Errorf("expected one (column name) or two (table name + column name) fields, got %d (%+v)", len(v), v)
			}
			opt.ColumnRef = v[len(v)-1].GetString_().GetStr()
		case *pg_query.Node_ParamRef:
			opt.IndexArg = int(rExpr.GetParamRef().GetNumber())
		case *pg_query.Node_AConst:
			t, v, err := s.typeAndValOfConst(rExpr.GetAConst())
			if err != nil {
				return nil, fmt.Errorf("failed to extract value from const: %s", err)
			}
			opt.Type = t
			opt.Value = v
		case *pg_query.Node_TypeCast:
			t, v, err := s.typeAndValOfTypeCast(rExpr.GetTypeCast())
			if err != nil {
				return nil, fmt.Errorf("failed to extract value from type cast: %s", err)
			}
			opt.Type = t
			opt.Value = v
		default:
			return nil, fmt.Errorf("expression %q is not valid: unexpected type for rexpr %s: %T", aExpr, rExpr.GetNode(), rExpr.GetNode())
		}

		var fieldName string
		switch lExpr.GetNode().(type) {
		case *pg_query.Node_ColumnRef:
			fieldName = lExpr.GetColumnRef().GetFields()[0].GetString_().GetStr()
		case *pg_query.Node_TypeCast:
			fieldName = fmt.Sprintf("$%d", lExpr.GetTypeCast().GetArg().GetParamRef().GetNumber())
		default:
			return nil, fmt.Errorf("expression %q is not valid: lexpr is not a ColumnRef but a %T", aExpr, lExpr.GetNode())
		}
		res[fieldName] = &CheckWhereClauseOpt{EqualExpr: &opt}

	case "BETWEEN":
		items := rExpr.GetList().Items
		args := make([]string, 0, len(items))
		for _, item := range items {
			switch item.GetNode().(type) {
			case *pg_query.Node_ColumnRef:
				fieldName := item.GetColumnRef().GetFields()[0].GetString_().GetStr()
				args = append(args, fieldName)
			default:
				return nil, fmt.Errorf("expression %q is not valid: unexpected type for argument in BETWEEN expression %s: %T", aExpr, item.GetNode(), item.GetNode())
			}
		}

		switch lExpr.GetNode().(type) {
		case *pg_query.Node_FuncCall:
			funcName := lExpr.GetFuncCall().GetFuncname()[0].GetString_().GetStr()
			res[funcName] = &CheckWhereClauseOpt{
				BetweenExpr: &BetweenExpr{
					Field: funcName,
					Args:  args,
				},
			}
		default:
			return nil, fmt.Errorf("expression %q is not valid: lexpr is not a FuncCall but a %T", aExpr, lExpr.GetNode())
		}

	default:
		return nil, fmt.Errorf("unexpected comparsion %q", comparsion)
	}

	return res, nil
}

// typeAndValOfConst returns the type in string format and value of the input.
func (s *RawStmt) typeAndValOfConst(c *pg_query.A_Const) (string, interface{}, error) {
	v := c.GetVal()
	switch v.GetNode().(type) {
	case *pg_query.Node_Integer:
		return "int32", v.GetInteger().GetIval(), nil
	case *pg_query.Node_String_:
		return "string", v.GetString_().GetStr(), nil
	default:
		return "", nil, fmt.Errorf("unexpected type for constant %s: %T", v.GetNode(), v.GetNode())
	}
}

// typeAndValOfTypeCast returns the type in string format and value of the input.
func (s *RawStmt) typeAndValOfTypeCast(tc *pg_query.TypeCast) (string, interface{}, error) {
	// items[0] is always pg_catalog, items[1] is the type
	t := tc.GetTypeName().GetNames()[1].GetString_().GetStr()
	switch t {
	case "bool":
		vstr := tc.GetArg().GetAConst().GetVal().GetString_().GetStr()
		return "bool", vstr == "t", nil
	default:
		return "", nil, fmt.Errorf("unexpected type for type cast %s", t)
	}
}

var paramRe = regexp.MustCompile(`^\$\d+$`)

// extractConditionFromBoolExpr returns the conditions in the input.
// pg_query.BoolExpr contains multiple conditions, which can be pg_query.AExpr, pg_query.NullTest, or even
// another sub-expression of pg_query.BoolExpr.
func (s *RawStmt) extractConditionFromBoolExpr(boolExpr *pg_query.BoolExpr) (map[string]*CheckWhereClauseOpt, error) {
	res := make(map[string]*CheckWhereClauseOpt)
	for _, arg := range boolExpr.GetArgs() {
		switch arg.GetNode().(type) {
		case *pg_query.Node_AExpr:
			aExpr := arg.GetAExpr()
			m, err := s.extractConditionsFromAExpr(aExpr)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to extract conditions from aexpr %s", aExpr)
			}
			res = MergeWhereConditions(res, m)
		case *pg_query.Node_NullTest:
			nullTest := arg.GetNullTest()
			m, err := s.extractConditionsFromNullTest(nullTest)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to extract conditions from null test %s: %s", nullTest, err)
			}
			res = MergeWhereConditions(res, m)
		case *pg_query.Node_BoolExpr:
			boolExpr := arg.GetBoolExpr()
			m, err := s.extractConditionFromBoolExpr(boolExpr)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to extract conditions from bool expr %s", boolExpr)
			}
			res = MergeWhereConditions(res, m)
		default:
			return nil, fmt.Errorf("unexpected type for bool expr %s: %T", arg.GetNode(), arg.GetNode())
		}
	}

	s.postProcessForBoolExpr(res)
	return res, nil
}

// For clauses like WHERE $1::text IS NOT NULL AND colA = $1
// we mark colA's HasNullTest as true.
func (s *RawStmt) postProcessForBoolExpr(res map[string]*CheckWhereClauseOpt) {
	// Reverse look-up from $1 to colA
	paramRef := make(map[string]string)
	for fieldName, fieldOpt := range res {
		if fieldOpt.EqualExpr != nil && fieldOpt.EqualExpr.IndexArg > 0 { // match found
			paramRef[fmt.Sprintf("$%d", fieldOpt.EqualExpr.IndexArg)] = fieldName
		}
	}

	// Look for $1 in res
	for fieldName, fieldOpt := range res {
		if !fieldOpt.HasNullTest {
			continue
		}
		loc := paramRe.FindStringIndex(fieldName)
		if loc == nil { // no match found
			continue
		}

		reversedColumnRef, ok := paramRef[fieldName]
		if !ok { // no column is using = $1
			continue
		}

		// Mark the column as having null test
		res[reversedColumnRef].HasNullTest = true
	}
}

func (s *RawStmt) extractConditionsFromNullTest(nullTest *pg_query.NullTest) (map[string]*CheckWhereClauseOpt, error) {
	arg := nullTest.GetArg()
	var fieldName string
	switch arg.GetNode().(type) {
	case *pg_query.Node_ColumnRef: // example: deleted_at IS NOT NULL
		v := arg.GetColumnRef().GetFields()
		if len(v) > 2 {
			return nil, fmt.Errorf("expected one (column name) or two (table name + column name) fields, got %d (%+v)", len(v), v)
		}
		fieldName = v[len(v)-1].GetString_().GetStr()
	case *pg_query.Node_TypeCast: // example: $1::text IS NOT NULL
		fieldName = fmt.Sprintf("$%d", arg.GetTypeCast().GetArg().GetParamRef().GetNumber())
	default:
		return nil, fmt.Errorf("unexpected type for null test %s: %T", arg, arg)
	}
	return map[string]*CheckWhereClauseOpt{
		fieldName: {
			HasNullTest: true,
		},
	}, nil
}
