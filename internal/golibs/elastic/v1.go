package elastic

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// SearchFactory implemented by SearchFactoryImpl
type SearchFactory interface {
	NewBulkIndexer() (esutil.BulkIndexer, error)

	// Deprecated: DLS does not support update API
	Update(index string, id string, body io.Reader, o ...func(*esapi.UpdateRequest)) (*esapi.Response, error)

	// Deprecated: DLS does not support update API
	UpdateCtx(ctx context.Context, index string, id string, body io.Reader, o ...func(*esapi.UpdateRequest)) (*esapi.Response, error)
	Search(ctx context.Context, indexName string, read *strings.Reader, o ...func(*esapi.SearchRequest)) (*esapi.Response, error)
	GetClient() *elasticsearch.Client
	// Deprecated: use BulkIndexWithResourcePath in SearchFactoryV2 instead
	CreateDocuments(ctx context.Context, data map[string][]byte, actionKind, indexName string) (int, error)
	// Deprecated: use BulkIndexWithResourcePath in SearchFactoryV2 instead
	BulkIndex(ctx context.Context, data map[string][]byte, actionKind, indexName string) (int, error)
	CheckIndexExists(index string) (bool, error)
	CreateIndex(index string, body io.Reader) (*esapi.Response, error)
	DeleteIndex(index string) (*esapi.Response, error)

	SearchFactoryV2
}

type SearchFactoryImpl struct {
	basicAuthClient *elasticsearch.Client
	jwtAuthClient   *elasticsearch.Client
	logger          *zap.Logger
}

func (e *SearchFactoryImpl) NewBulkIndexer() (esutil.BulkIndexer, error) {
	// In current usecase, we only want 1 worker
	// multiple worker only useful in case of pipeline/batching
	return esutil.NewBulkIndexer(
		esutil.BulkIndexerConfig{
			Client:     e.basicAuthClient,
			NumWorkers: 1,
			FlushBytes: 1e+4,
		},
	)
}
func (e *SearchFactoryImpl) GetClient() *elasticsearch.Client {
	return e.basicAuthClient
}

func (e *SearchFactoryImpl) UpdateCtx(ctx context.Context, index string, id string, body io.Reader, o ...func(*esapi.UpdateRequest)) (*esapi.Response, error) {
	req := esapi.UpdateRequest{Index: index, DocumentID: id, Body: body}
	for _, f := range o {
		f(&req)
	}

	res, err := req.Do(ctx, e.basicAuthClient)
	return res, err
}

func (e *SearchFactoryImpl) Update(index string, id string, body io.Reader, o ...func(*esapi.UpdateRequest)) (*esapi.Response, error) {
	return e.basicAuthClient.Update(index, id, body, o...)
}

func (e *SearchFactoryImpl) Search(ctx context.Context, indexName string, read *strings.Reader, o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	if e.GetClient() == nil {
		return nil, fmt.Errorf("the connection closed")
	}
	opts := make([]func(*esapi.SearchRequest), 0)
	opts = append(opts, e.basicAuthClient.Search.WithContext(ctx))
	opts = append(opts, e.basicAuthClient.Search.WithIndex(indexName))
	opts = append(opts, e.basicAuthClient.Search.WithBody(read))
	opts = append(opts, e.basicAuthClient.Search.WithTrackTotalHits(true))
	opts = append(opts, e.basicAuthClient.Search.WithPretty())
	// other options
	opts = append(opts, o...)
	return e.basicAuthClient.Search(opts...)
}

func (e *SearchFactoryImpl) BulkIndex(ctx context.Context, data map[string][]byte, actionKind, indexName string) (totalSuccess int, err error) {
	errChan := make(chan error, 1)
	bi, err := e.NewBulkIndexer()
	if err != nil {
		err = fmt.Errorf("unable to create bulk indexer: %w", err)
		return
	}
	maxTimesRetries := 5

	for id, value := range data {
		smallErr := bi.Add(ctx, esutil.BulkIndexerItem{
			Index:      indexName,
			Action:     actionKind,
			DocumentID: id,
			Body:       bytes.NewReader(value),
			OnFailure: func(c context.Context, bii esutil.BulkIndexerItem, biri esutil.BulkIndexerResponseItem, e error) {
				if e != nil {
					ctxzap.Extract(ctx).Error("fail to create document", zap.String("document_index", indexName), zap.Error(e))
				} else {
					ctxzap.Extract(ctx).Error("fail to import document", zap.String("document_index", indexName), zap.String("document_id", biri.DocumentID), zap.String(biri.Error.Type, biri.Error.Reason))
				}
				errbytes, merr := json.Marshal(biri.Error)
				if merr != nil {
					errChan <- fmt.Errorf("item: %s has error %v", string(value), biri.Error)
				} else {
					errChan <- fmt.Errorf("item: %s has error %s", string(value), string(errbytes))
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

// CreateDocuments using map[string][]byte with string is a ID of document and []byte is data of the document
func (e *SearchFactoryImpl) CreateDocuments(ctx context.Context, data map[string][]byte, actionKind, indexName string) (totalSuccess int, err error) {
	return e.BulkIndex(ctx, data, actionKind, indexName)
}

func (e *SearchFactoryImpl) CheckIndexExists(index string) (bool, error) {
	res, err := e.basicAuthClient.Indices.Exists([]string{index})
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusNotFound:
		return false, nil
	case http.StatusOK:
		return true, nil
	default:
		return false, fmt.Errorf("unable to check the index exist: %s", res.String())
	}
}

func (e *SearchFactoryImpl) CreateIndex(index string, body io.Reader) (*esapi.Response, error) {
	return e.basicAuthClient.Indices.Create(index, func(req *esapi.IndicesCreateRequest) {
		req.Body = body
	})
}

func (e *SearchFactoryImpl) DeleteIndex(index string) (*esapi.Response, error) {
	return e.basicAuthClient.Indices.Delete([]string{index})
}

func (e *SearchFactoryImpl) DeletebyQuery(index string, body io.Reader) (*esapi.Response, error) {
	return e.basicAuthClient.DeleteByQuery([]string{index}, body)
}

func (e *SearchFactoryImpl) Count(index string, body io.Reader) (*esapi.Response, error) {
	count := e.basicAuthClient.Count
	opts := []func(*esapi.CountRequest){
		count.WithBody(body),
		count.WithIndex(index),
	}
	return e.basicAuthClient.Count(opts...)
}

// NewSearchFactory creates new SearchFactory
func NewSearchFactory(zapLogger *zap.Logger, addrs []string, user, password, cloudID, apiKey string) (*SearchFactoryImpl, error) {
	if len(addrs) == 0 && cloudID == "" {
		return nil, fmt.Errorf("elastic: neither addrs nor cloudID is empty")
	}
	if user == "" && password == "" {
		return nil, fmt.Errorf("elastic: missing user and password")
	}
	//nolint:gosec
	cfg := elasticsearch.Config{
		Username: user,
		Password: password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	// connect using addrs or cloudID, we priority cloudID
	if cloudID != "" {
		cfg.CloudID = cloudID
		cfg.APIKey = apiKey
	} else if len(addrs) != 0 {
		cfg.Addresses = addrs
	}

	basicAuthClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	//nolint:gosec
	jwtAuthcfg := elasticsearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	// connect using addrs or cloudID, we priority cloudID
	if cloudID != "" {
		jwtAuthcfg.CloudID = cloudID
		jwtAuthcfg.APIKey = apiKey
	} else if len(addrs) != 0 {
		jwtAuthcfg.Addresses = addrs
	}

	jwtAuthClient, err := elasticsearch.NewClient(jwtAuthcfg)
	if err != nil {
		return nil, err
	}
	e := &SearchFactoryImpl{
		basicAuthClient: basicAuthClient,
		jwtAuthClient:   jwtAuthClient,
		logger:          zapLogger,
	}
	return e, nil
}

func NewMockSearchFactory(mockResp string) (*SearchFactoryImpl, func()) {
	ts := httptest.NewServer(mockHandler(mockResp))
	esclient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})
	if err != nil {
		panic(err)
	}
	return &SearchFactoryImpl{
		basicAuthClient: esclient,
		jwtAuthClient:   esclient,
		logger:          zap.NewNop(),
	}, ts.Close
}
