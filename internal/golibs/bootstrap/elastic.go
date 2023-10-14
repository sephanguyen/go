package bootstrap

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/elastic"

	"go.uber.org/zap"
)

func initElastic(_ context.Context, config interface{}, rsc *Resources) error {
	c, err := extract[configs.ElasticSearchConfig](config, elasticFieldName)
	if err != nil {
		return ignoreErrFieldNotFound(err)
	}

	_ = rsc.WithElasticC(c)
	return nil
}

// Elasticer handles the connection to an Elasticsearch instance.
type Elasticer interface {
	// Init initializes and returns a new elastic client.
	Init(l *zap.Logger, addrs []string, user, password, cloudID, apiKey string) (*elastic.SearchFactoryImpl, error)
}

// elasticImpl implements Elasticer using elastic.NewSearchFactory function.
type elasticImpl struct{}

func newElasticImpl() *elasticImpl {
	return &elasticImpl{}
}

func (e *elasticImpl) Init(l *zap.Logger, addrs []string, user, password, cloudID, apiKey string) (*elastic.SearchFactoryImpl, error) {
	return elastic.NewSearchFactory(l, addrs, user, password, cloudID, apiKey)
}
