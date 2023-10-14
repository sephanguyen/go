package op

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

func GreaterThanOrEqual(columnName string, value interface{}) Condition {
	return &operatorGte{
		columnName: columnName,
		value:      value,
	}
}

type operatorGte struct {
	columnName string
	value      interface{}
}

func (p *operatorGte) BuildQuery() elastic.Query {
	return elastic.NewRangeQuery(p.columnName).Gte(p.value)
}

func (p *operatorGte) String() string {
	return fmt.Sprintf("operatorGte-%s-%v", p.columnName, p.value)
}
