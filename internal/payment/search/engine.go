package search

import (
	"context"
	"io"

	"github.com/manabie-com/backend/internal/payment/search/op"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type Engine interface {
	Insert(ctx context.Context, tableName string, contents []InsertionContent) (totalSuccess int, err error)
	GetAll(ctx context.Context, tableName string, funcRecv func(data []byte) (interface{}, error), pagingParam PagingParam, sortParams ...SortParam) ([]interface{}, error)
	Search(ctx context.Context, tableName string, condition op.Condition, funcRecv func(data []byte) (interface{}, error), pagingParam PagingParam, sortParams ...SortParam) ([]interface{}, error)
	SearchWithoutPaging(ctx context.Context, tableName string, condition op.Condition, funcRecv func(data []byte) (interface{}, error), sortParams ...SortParam) ([]interface{}, error)
	CountValue(ctx context.Context, tableName, columnName string, condition op.Condition) (uint32, error)
	CheckIndexExists(index string) (bool, error)
	CreateIndex(index string, body io.Reader) (*esapi.Response, error)
	DeleteIndex(index string) (*esapi.Response, error)
}
