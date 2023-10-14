package mastermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/configurations"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func init() {
	bootstrap.RegisterJob("scan_es_resource_path", scanESResourcePath).
		Desc("Scan elastic search resource_path").
		DescLong("Scan elastic search resource_path if type key word")
}

type indexMapping struct {
	Mapping indexProperties `json:"mappings"`
}

type indexProperties struct {
	Properties esResourcePath `json:"properties"`
}

type esType struct {
	EsType string `json:"type"`
}
type esResourcePath struct {
	ResourcePath esType `json:"resource_path"`
}

func getCreatedESIndices(ctx context.Context, searchClient elastic.SearchFactory) ([]string, error) {
	client := searchClient.GetClient()
	opts := make([]func(*esapi.CatIndicesRequest), 0)
	opts = append(opts, client.Cat.Indices.WithContext(ctx))
	opts = append(opts, client.Cat.Indices.WithPretty())
	opts = append(opts, client.Cat.Indices.WithH("index"))

	// other options
	rsp, err := client.Cat.Indices(opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get created indices: %w", err)
	}
	defer rsp.Body.Close()
	rs, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}
	indices := strings.Split(string(rs), "\n")

	excludeIndicesPrefix := []string{"security-auditlog", ".kibana", ".opendistro", " "}
	createdIndeces := sliceutils.Filter(indices, func(index string) bool {
		for _, prefix := range excludeIndicesPrefix {
			if strings.HasPrefix(index, prefix) {
				return false
			}
		}
		return true
	})
	return createdIndeces, nil
}

func getIndicesDetails(ctx context.Context, searchClient elastic.SearchFactory, indices []string) (map[string]indexMapping, error) {
	client := searchClient.GetClient()
	opts := make([]func(*esapi.IndicesGetRequest), 0)
	opts = append(opts, client.Indices.Get.WithContext(ctx))
	opts = append(opts, client.Indices.Get.WithPretty())

	// other options
	rsp, err := client.Indices.Get(indices, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get created indices: %w", err)
	}
	defer rsp.Body.Close()

	dec := json.NewDecoder(rsp.Body)
	result := make(map[string]indexMapping)
	err = dec.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return result, nil
}

func scanESResourcePath(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()

	searchClient, err := elastic.NewSearchFactory(zapLogger, c.ElasticSearch.Addresses, c.ElasticSearch.Username, c.ElasticSearch.Password, "", "")
	if err != nil {
		return fmt.Errorf("unable to connect elasticsearch: %s", err)
	}
	log.Println("====", searchClient)

	createdIndices, err := getCreatedESIndices(ctx, searchClient)
	if err != nil {
		return fmt.Errorf("unable to getCreatedESIndices: %s", err)
	}
	indicesDetails, err := getIndicesDetails(ctx, searchClient, createdIndices)
	if err != nil {
		return fmt.Errorf("unable to getIndicesDetails: %s", err)
	}
	for index, indexDetail := range indicesDetails {
		if indexDetail.Mapping.Properties.ResourcePath.EsType != "keyword" {
			return fmt.Errorf("resource_path does not have type keyword: index: %s has resource_path with type: %s", index, indexDetail.Mapping.Properties.ResourcePath.EsType)
		}
	}
	return nil
}
