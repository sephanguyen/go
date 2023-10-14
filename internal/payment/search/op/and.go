package op

import (
	"github.com/olivere/elastic/v7"
)

func And(ops ...Condition) Condition {
	return &operatorAnd{
		ops: ops,
	}
}

type operatorAnd struct {
	ops []Condition
}

func (p *operatorAnd) BuildQuery() elastic.Query {
	queries := make([]elastic.Query, 0, len(p.ops))
	for _, op := range p.ops {
		queries = append(queries, op.BuildQuery())
	}
	return elastic.NewBoolQuery().Must(queries...)
}

func (p *operatorAnd) String() string {
	str := ""
	for _, op := range p.ops {
		str += "-" + op.String()
	}
	return "operatorAnd" + str
}
