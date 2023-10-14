package op

import (
	"github.com/olivere/elastic/v7"
)

func Or(ops ...Condition) Condition {
	return &operatorOr{
		ops: ops,
	}
}

type operatorOr struct {
	ops []Condition
}

func (p *operatorOr) BuildQuery() elastic.Query {
	queries := make([]elastic.Query, 0, len(p.ops))
	for _, op := range p.ops {
		queries = append(queries, op.BuildQuery())
	}
	return elastic.NewBoolQuery().Should(queries...)
}

func (p *operatorOr) String() string {
	str := ""
	for _, op := range p.ops {
		str += "-" + op.String()
	}
	return "operatorOr" + str
}
