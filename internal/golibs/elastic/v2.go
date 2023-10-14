package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

// For reference only, best practice
// is to use smallest interface
type SearchFactoryV2 interface {
	ElasticsearchClientProxy
	ElasticsearchIndexer
}

type ElasticsearchClientProxy interface {
	SearchUsingJwtToken(ctx context.Context, indexName string, read io.Reader, o ...func(*esapi.SearchRequest)) (*esapi.Response, error)
}

// ElasticsearchIndexer used when dls enabled only
type ElasticsearchIndexer interface {
	// BulkIndex index documents, resource_path of documents will be set based on ctx
	BulkIndexWithResourcePath(ctx context.Context, data map[string]Doc, indexName string) (int, error)

	//TODO: support more operations: update/update_by_query with multi-tenancy safe
}

func (e *SearchFactoryImpl) SearchUsingJwtToken(ctx context.Context, indexName string, read io.Reader, o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	opts := make([]func(*esapi.SearchRequest), 0)
	opts = append(opts, e.basicAuthClient.Search.WithContext(ctx))
	opts = append(opts, e.basicAuthClient.Search.WithIndex(indexName))
	opts = append(opts, e.basicAuthClient.Search.WithBody(read))
	opts = append(opts, e.basicAuthClient.Search.WithTrackTotalHits(true))
	opts = append(opts, e.basicAuthClient.Search.WithPretty())
	opts = append(opts, o...)

	// TODO
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("SearchUsingJwtToken: ctx has no incoming grpc metadata")
	}
	toks := md.Get("token")
	if len(toks) != 1 {
		return nil, fmt.Errorf("SearchUsingJwtToken: want 1 token item in context, has %d", len(toks))
	}
	jwtOpt := e.basicAuthClient.Search.WithHeader(map[string]string{
		"Authorization": "Bearer " + toks[0],
	})
	opts = append(opts, jwtOpt)
	// other options
	return e.jwtAuthClient.Search(opts...)
}

func (e *SearchFactoryImpl) BulkIndexWithResourcePath(ctx context.Context, datas map[string]Doc, indexName string) (totalSuccess int, err error) {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim.Manabie == nil || claim.Manabie.ResourcePath == "" {
		return 0, fmt.Errorf("no multi-tenancy context")
	}
	resourcePath := claim.Manabie.ResourcePath
	commonMandatoryField := Mandatory{
		ResourcePath: resourcePath,
	}

	errChan := make(chan error, 1)
	bi, err := e.NewBulkIndexer()
	if err != nil {
		err = fmt.Errorf("unable to create bulk indexer: %w", err)
		return
	}
	maxTimesRetries := 5

	for id, value := range datas {
		value._mandatory = commonMandatoryField
		raw, err2 := json.Marshal(value)
		if err2 != nil {
			err = multierr.Combine(err, fmt.Errorf("json.Marshal %w", err2))
			break
		}
		smallErr := bi.Add(ctx, esutil.BulkIndexerItem{
			Index:      indexName,
			Action:     "index",
			DocumentID: id,
			Body:       strings.NewReader(string(raw)),
			OnFailure: func(c context.Context, bii esutil.BulkIndexerItem, biri esutil.BulkIndexerResponseItem, e error) {
				if e != nil {
					ctxzap.Extract(ctx).Error("fail to create document", zap.String("document_index", indexName), zap.Error(e))
				} else {
					ctxzap.Extract(ctx).Error("fail to import document", zap.String("document_index", indexName), zap.String("document_id", biri.DocumentID), zap.String(biri.Error.Type, biri.Error.Reason))
				}
				errbytes, merr := json.Marshal(biri.Error)
				if merr != nil {
					errChan <- fmt.Errorf("item: %s has error %v", string(raw), biri.Error)
				} else {
					errChan <- fmt.Errorf("item: %s has error %s", string(raw), string(errbytes))
				}
			},
			RetryOnConflict: &maxTimesRetries,
		})
		if smallErr != nil {
			err = multierr.Combine(err, smallErr)
		}
	}
	go func() {
		// must be called only after adding
		closeErr := bi.Close(ctx)
		if closeErr != nil {
			e.logger.Error("failed to close bulk indexer", zap.Error(closeErr))
		}

		// this is safe, OnFailure will no longer be called after bi.Close()
		close(errChan)
	}()
	for e := range errChan {
		err = multierr.Combine(err, e)
	}
	totalSuccess = int(bi.Stats().NumFlushed)
	return
}

type Mandatory struct {
	ResourcePath string `json:"resource_path"`
}
type Doc struct {
	_mandatory Mandatory
	Inner      interface{}
}

func NewDoc(inner interface{}) Doc {
	return Doc{Inner: inner}
}

func (o Doc) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(o._mandatory)
	if err != nil {
		return nil, err
	}
	m, err := json.Marshal(o.Inner)
	if err != nil {
		return nil, err
	}
	// inner struct marshal into {}
	if len(m) == 2 {
		return b, nil
	}
	// appending like "field_1":"val_1",....} with a final closing bracket
	b[len(b)-1] = ','
	return append(b, m[1:]...), nil
}
