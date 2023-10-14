package op

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

func Equal(columnName string, value interface{}) Condition {
	return &operatorEqual{
		columnName: columnName,
		value:      value,
	}
}

type operatorEqual struct {
	columnName string
	value      interface{}
}

func (p *operatorEqual) BuildQuery() elastic.Query {
	return elastic.NewMatchQuery(p.columnName, p.value).Operator("AND")
}

func (p *operatorEqual) String() string {
	return fmt.Sprintf("operatorEqual-%s-%v", p.columnName, p.value)
}
