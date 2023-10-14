package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	internalelastic "github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/search/op"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func NewElasticSearch(searchFactory internalelastic.SearchFactory) Engine {
	return &elasticSearch{
		searchFactory: searchFactory,
	}
}

type elasticSearch struct {
	searchFactory internalelastic.SearchFactory
}

func (p *elasticSearch) Insert(ctx context.Context, tableName string, contents []InsertionContent) (totalSuccess int, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ElasticSearch.Insert")
	defer span.End()

	var (
		data = make(map[string]internalelastic.Doc)
	)

	for _, content := range contents {
		data[content.ID] = internalelastic.NewDoc(content.Data)
	}
	totalSuccess, err = p.searchFactory.BulkIndexWithResourcePath(ctx, data, tableName)
	if err != nil {
		return totalSuccess, fmt.Errorf("failed to insert document to elasticsearch: %w", err)
	}
	return
}

func (p *elasticSearch) GetAll(ctx context.Context, tableName string, funcRecv func(data []byte) (interface{}, error), pagingParam PagingParam, sortParams ...SortParam) ([]interface{}, error) {
	internalSortParams := make([]internalelastic.SortParam, 0, len(sortParams))
	for _, sortParam := range sortParams {
		internalSortParams = append(internalSortParams, internalelastic.SortParam{
			ColumnName: sortParam.ColumnName,
			Ascending:  sortParam.Ascending,
		})
	}
	esClient := p.searchFactory.GetClient()
	source := internalelastic.SetSortOnSource(internalelastic.NewSearchSource(), internalSortParams...)
	res, err := internalelastic.DoSearchFromSourceUsingJwtToken(
		ctx,
		p.searchFactory,
		tableName,
		source,
		esClient.Search.WithFrom(int(pagingParam.FromIdx)),
		esClient.Search.WithSize(int(pagingParam.NumberRows)),
	)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0)
	err = internalelastic.ParseSearchResponse(res.Body, func(hit *internalelastic.SearchHit) error {
		item, err := funcRecv(hit.Source)
		if err == nil {
			result = append(result, item)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *elasticSearch) Search(ctx context.Context, tableName string, condition op.Condition, funcRecv func(data []byte) (interface{}, error), pagingParam PagingParam, sortParams ...SortParam) ([]interface{}, error) {
	internalSortParams := make([]internalelastic.SortParam, 0, len(sortParams))
	for _, sortParam := range sortParams {
		internalSortParams = append(internalSortParams, internalelastic.SortParam{
			ColumnName: sortParam.ColumnName,
			Ascending:  sortParam.Ascending,
		})
	}
	source := internalelastic.SetSortOnSource(internalelastic.NewSearchSource(), internalSortParams...)
	if condition != nil {
		q := condition.BuildQuery()
		source = source.Query(q)
	}
	esClient := p.searchFactory.GetClient()
	res, err := internalelastic.DoSearchFromSourceUsingJwtToken(
		ctx,
		p.searchFactory,
		tableName,
		source,
		esClient.Search.WithFrom(int(pagingParam.FromIdx)),
		esClient.Search.WithSize(int(pagingParam.NumberRows)),
	)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0)
	err = internalelastic.ParseSearchResponse(res.Body, func(hit *internalelastic.SearchHit) error {
		item, err := funcRecv(hit.Source)
		if err == nil {
			result = append(result, item)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *elasticSearch) SearchWithoutPaging(ctx context.Context, tableName string, condition op.Condition, funcRecv func(data []byte) (interface{}, error), sortParams ...SortParam) ([]interface{}, error) {
	internalSortParams := make([]internalelastic.SortParam, 0, len(sortParams))
	for _, sortParam := range sortParams {
		internalSortParams = append(internalSortParams, internalelastic.SortParam{
			ColumnName: sortParam.ColumnName,
			Ascending:  sortParam.Ascending,
		})
	}
	source := internalelastic.SetSortOnSource(internalelastic.NewSearchSource(), internalSortParams...)
	if condition != nil {
		q := condition.BuildQuery()
		source = source.Query(q)
	}
	esClient := p.searchFactory.GetClient()
	res, err := internalelastic.DoSearchFromSourceUsingJwtToken(
		ctx,
		p.searchFactory,
		tableName,
		source,
		esClient.Search.WithFrom(0),
		esClient.Search.WithSize(10000),
	)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0)
	err = internalelastic.ParseSearchResponse(res.Body, func(hit *internalelastic.SearchHit) error {
		item, err := funcRecv(hit.Source)
		if err == nil {
			result = append(result, item)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *elasticSearch) CountValue(ctx context.Context, tableName, columnName string, condition op.Condition) (uint32, error) {
	source := internalelastic.NewSearchSource().Size(0)
	if condition != nil {
		q := condition.BuildQuery()
		source = source.Query(q)
	}
	source = source.Aggregation("count_value", internalelastic.NewValueCountAggregation().Field(columnName))
	res, err := internalelastic.DoSearchFromSourceUsingJwtToken(
		ctx,
		p.searchFactory,
		tableName,
		source,
	)
	if err != nil {
		return 0, err
	}

	var order entities.ElasticOrder
	total, err := internalelastic.ParseSearchWithTotalResponse(res.Body, func(hit *internalelastic.SearchHit) error {
		err = json.Unmarshal(hit.Source, &order)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("error ParseSearchResponse: %w", err)
	}

	return total, nil
}

func (p *elasticSearch) CheckIndexExists(index string) (bool, error) {
	return p.searchFactory.CheckIndexExists(index)
}

func (p *elasticSearch) CreateIndex(index string, body io.Reader) (*esapi.Response, error) {
	return p.searchFactory.CreateIndex(index, body)
}

func (p *elasticSearch) DeleteIndex(index string) (*esapi.Response, error) {
	return p.searchFactory.DeleteIndex(index)
}
