package op

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

func LessThanOrEqual(columnName string, value interface{}) Condition {
	return &operatorLte{
		columnName: columnName,
		value:      value,
	}
}

type operatorLte struct {
	columnName string
	value      interface{}
}

func (p *operatorLte) BuildQuery() elastic.Query {
	return elastic.NewRangeQuery(p.columnName).Lte(p.value)
}

func (p *operatorLte) String() string {
	return fmt.Sprintf("operatorLte-%s-%v", p.columnName, p.value)
}
