package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/olivere/elastic/v7"
	"go.uber.org/multierr"
)

func DoSearchFromSourceUsingJwtToken(ctx context.Context,
	factory SearchFactory,
	indexName string, source *elastic.SearchSource, o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	somemap, err := source.Source()
	if err != nil {
		return nil, err
	}
	bs, err := json.Marshal(somemap)
	if err != nil {
		return nil, err
	}
	return factory.SearchUsingJwtToken(ctx, indexName, strings.NewReader(string(bs)), o...)
}

func DoSearchFromSource(ctx context.Context,
	factory SearchFactory,
	indexName string, source *elastic.SearchSource, o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	somemap, err := source.Source()
	if err != nil {
		return nil, err
	}
	bs, err := json.Marshal(somemap)
	if err != nil {
		return nil, err
	}
	return factory.Search(ctx, indexName, strings.NewReader(string(bs)), o...)
}

func SetSortOnSource(source *elastic.SearchSource, sortParams ...SortParam) *elastic.SearchSource {
	if len(sortParams) == 0 {
		return source
	}
	sorters := make([]elastic.Sorter, 0, len(sortParams))
	for _, sortParam := range sortParams {
		columnName := sortParam.ColumnName
		sorter := elastic.NewFieldSort(columnName).Desc()
		if sortParam.Ascending {
			sorter = sorter.Asc()
		}
		sorters = append(sorters, sorter)
	}
	return source.SortBy(sorters...)
}

func NewBoolQuery(...elastic.Query) *elastic.BoolQuery {
	return elastic.NewBoolQuery()
}
func NewTermQuery(field string, val interface{}) *elastic.TermQuery {
	return elastic.NewTermQuery(field, val)
}
func NewTermsQuery(name string, values ...interface{}) *elastic.TermsQuery {
	return elastic.NewTermsQuery(name, values...)
}

func NewMultiMatchQuery(text interface{}, fields ...string) *elastic.MultiMatchQuery {
	return elastic.NewMultiMatchQuery(text, fields...)
}
func NewSearchSource() *elastic.SearchSource {
	return elastic.NewSearchSource()
}
func NewValueCountAggregation() *elastic.ValueCountAggregation {
	return elastic.NewValueCountAggregation()
}
func NewExistQuery(field string) *elastic.ExistsQuery {
	return elastic.NewExistsQuery(field)
}
func NewRangeQuery(field string) *elastic.RangeQuery {
	return elastic.NewRangeQuery(field)
}
func NewMatchPhraseQuery(field string, text interface{}) *elastic.MatchPhraseQuery {
	return elastic.NewMatchPhraseQuery(field, text)
}
func NewWildcardQuery(field string, text string) *elastic.WildcardQuery {
	return elastic.NewWildcardQuery(field, text)
}
func NewScriptQuery(source string, params map[string]interface{}) (*elastic.ScriptQuery, error) {
	script := &elastic.Script{}
	script.Params(params)
	script.Script(source)
	if _, err := script.Source(); err != nil {
		return nil, err
	}
	return elastic.NewScriptQuery(script), nil
}

func NewError(e *elastic.ErrorDetails) error {
	rootCauses := ""
	for _, smallCause := range e.RootCause {
		rootCauses += NewError(smallCause).Error()
	}
	return ResponseErr{
		Errtype:    e.Type,
		Reason:     e.Reason,
		rootCauses: rootCauses,
	}
}

type ResponseErr struct {
	Errtype    string
	Reason     string
	rootCauses string
}

func (es ResponseErr) Error() string {
	return fmt.Sprintf("errtype: %s\nreason: %s\nrootCauses: %s\n", es.Errtype, es.Reason, es.rootCauses)
}

type SearchHit = elastic.SearchHit

// use unofficial client to parse response
func ParseSearchResponse(r io.ReadCloser, parseFunc func(*SearchHit) error) error {
	defer r.Close()
	res := elastic.SearchResult{}
	dec := json.NewDecoder(r)
	err := dec.Decode(&res)
	if err != nil {
		return err
	}
	if res.Error != nil {
		err = fmt.Errorf("res.Error: %w", NewError(res.Error))
		return err
	}

	if res.Hits != nil && res.Hits.TotalHits != nil && res.Hits.TotalHits.Value > 0 {
		for idx := range res.Hits.Hits {
			hit := res.Hits.Hits[idx]
			thisErr := parseFunc(hit)
			if thisErr != nil {
				err = multierr.Combine(err, thisErr)
			}
		}
	}
	return err
}
func ParseSearchWithResponse(r io.ReadCloser) (res elastic.SearchResult, err error) {
	defer r.Close()
	dec := json.NewDecoder(r)
	err = dec.Decode(&res)
	if err != nil {
		return
	}
	if res.Error != nil {
		err = fmt.Errorf("res.Error: %w", NewError(res.Error))
	}
	return
}

// read total of res
func ParseSearchWithTotalResponse(r io.ReadCloser, parseFunc func(*SearchHit) error) (total uint32, err error) {
	res, err := ParseSearchWithResponse(r)

	totalHits := res.TotalHits()

	if totalHits > 0 {
		for idx := range res.Hits.Hits {
			hit := res.Hits.Hits[idx]
			thisErr := parseFunc(hit)
			if thisErr != nil {
				err = multierr.Combine(err, thisErr)
			}
		}
	}

	return uint32(totalHits), err
}

func mockHandler(mockResp string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(mockResp))
		if err != nil {
			panic(err)
		}
	}
}
func CheckResponse(res *esapi.Response) error {
	// 200-299 are valid status codes
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}

	return createResponseError(res)
}

func createResponseError(res *esapi.Response) error {
	if res.Body == nil {
		return &elastic.Error{Status: res.StatusCode}
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &elastic.Error{Status: res.StatusCode}
	}
	errReply := new(elastic.Error)
	err = json.Unmarshal(data, errReply)
	if err != nil {
		return &elastic.Error{Status: res.StatusCode}
	}
	if errReply != nil {
		if errReply.Status == 0 {
			errReply.Status = res.StatusCode
		}
		return errReply
	}
	return &elastic.Error{Status: res.StatusCode}
}

func ParseAggregationValueCountResponse(r io.ReadCloser) (uint32, error) {
	defer r.Close()
	res := aggValueCountResponse{}
	dec := json.NewDecoder(r)
	err := dec.Decode(&res)
	if err != nil {
		return 0, err
	}
	return res.Aggregation.CountValue.Value, nil
}
