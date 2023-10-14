package op

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

func NewTermQuery(columnName string, text string) Condition {
	return &operatorTermQuery{
		columnName: columnName,
		text:       text,
	}
}

type operatorTermQuery struct {
	columnName string
	text       string
}

func (p *operatorTermQuery) BuildQuery() elastic.Query {
	return elastic.NewTermQuery(p.columnName, p.text).CaseInsensitive(true)
}

func (p *operatorTermQuery) String() string {
	return fmt.Sprintf("operatorMatch-%s-%s", p.columnName, p.text)
}
