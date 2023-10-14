package op

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

func LessThan(columnName string, value interface{}) Condition {
	return &operatorLt{
		columnName: columnName,
		value:      value,
	}
}

type operatorLt struct {
	columnName string
	value      interface{}
}

func (p *operatorLt) BuildQuery() elastic.Query {
	return elastic.NewRangeQuery(p.columnName).Lt(p.value)
}

func (p *operatorLt) String() string {
	return fmt.Sprintf("operatorLt-%s-%v", p.columnName, p.value)
}
