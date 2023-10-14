package op

import "github.com/olivere/elastic/v7"

type Condition interface {
	BuildQuery() elastic.Query
	String() string
}
