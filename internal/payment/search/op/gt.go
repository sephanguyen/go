package op

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

func GreaterThan(columnName string, value interface{}) Condition {
	return &operatorGt{
		columnName: columnName,
		value:      value,
	}
}

type operatorGt struct {
	columnName string
	value      interface{}
}

func (p *operatorGt) BuildQuery() elastic.Query {
	return elastic.NewRangeQuery(p.columnName).Gt(p.value)
}

func (p *operatorGt) String() string {
	return fmt.Sprintf("operatorGt-%s-%v", p.columnName, p.value)
}
