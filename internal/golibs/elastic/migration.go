package elastic

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/tidwall/gjson"
)

func parseResponse(res *esapi.Response) string {
	defer res.Body.Close()
	bs, _ := ioutil.ReadAll(res.Body)
	return string(bs)
}

// TODO UpsertAnalyzerDefinition

func (s *SearchFactoryImpl) UpsertFieldDefinition(ctx context.Context, index string, field string, fieldDefinition string) error {
	cl := s.basicAuthClient
	findFieldMapping := esapi.IndicesGetFieldMappingRequest{
		Index:  []string{index},
		Fields: []string{field},
	}
	res3, err := findFieldMapping.Do(ctx, cl)
	if err != nil {
		return fmt.Errorf("findFieldMapping %s %w", field, err)
	}
	somejson := parseResponse(res3)
	node := gjson.Get(somejson, fmt.Sprintf("%s.mappings.%s", index, field))
	if node.Exists() {
		return nil
	}

	s.logger.Info(fmt.Sprintf("%s not exist in index %s, creating with definition %s", field, index, fieldDefinition))
	mpreq := esapi.IndicesPutMappingRequest{
		Index: []string{index},
		Body:  strings.NewReader(fieldDefinition),
	}
	res, err := mpreq.Do(context.Background(), cl)
	if err != nil {
		return fmt.Errorf("creating index for field %s with definition %s has error %w", field, fieldDefinition, err)
	}
	somejson = parseResponse(res)
	node = gjson.Get(somejson, "error")
	if node.Exists() {
		return fmt.Errorf("creating index for field %s with definition %s has error %v", field, fieldDefinition, node)
	}
	return nil
}
