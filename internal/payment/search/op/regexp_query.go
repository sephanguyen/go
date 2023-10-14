package op

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

func NewRegexpQuery(columnName string, text string) Condition {
	return &operatorRegexpQuery{
		columnName: columnName,
		text:       text,
	}
}

type operatorRegexpQuery struct {
	columnName string
	text       string
}

func (p *operatorRegexpQuery) BuildQuery() elastic.Query {
	regexp := ".*" + p.text + ".*"
	return elastic.NewRegexpQuery(p.columnName, regexp).CaseInsensitive(true)
}

func (p *operatorRegexpQuery) String() string {
	return fmt.Sprintf("operatorMatch-%s-%s", p.columnName, p.text)
}
