package bootstrap

import (
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	mock_bootstrap "github.com/manabie-com/backend/mock/golibs/bootstrap"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitElastic(t *testing.T) {
	t.Parallel()
	l := zap.NewNop()

	t.Run("with elastic config", func(t *testing.T) {
		type testConfig struct {
			ElasticSearch configs.ElasticSearchConfig
		}

		c := testConfig{ElasticSearch: configs.ElasticSearchConfig{Addresses: []string{"address"}, Username: "username", Password: "password"}}
		e := &elastic.SearchFactoryImpl{}
		elasticer := &mock_bootstrap.Elasticer{}
		elasticer.On("Init", l, []string{"address"}, "username", "password", "", "").
			Return(&elastic.SearchFactoryImpl{}, nil).Once()
		r := NewResources().WithLogger(l)
		r.elasticer = elasticer

		err := initElastic(nil, c, r)
		assert.NoError(t, err)
		assert.Equal(t, e, r.Elastic())
	})

	t.Run("without elastic config", func(t *testing.T) {
		type testConfig struct{}

		c := testConfig{}
		r := NewResources().WithLogger(l)
		r.elasticer = &mock_bootstrap.Elasticer{}

		err := initElastic(nil, c, r)
		assert.NoError(t, err)
		assert.Panics(t, func() { _ = r.Elastic() })
	})

	t.Run("with invalid elastic config", func(t *testing.T) {
		type testConfig struct {
			ElasticSearch configs.ElasticSearchConfig
		}

		c := testConfig{ElasticSearch: configs.ElasticSearchConfig{}}
		elasticer := &mock_bootstrap.Elasticer{}
		elasticer.On("Init", l, []string(nil), "", "", "", "").
			Return(nil, errors.New("missing user and password")).Once()
		r := NewResources().WithLogger(l)
		r.elasticer = elasticer
		err := initElastic(nil, c, r)
		assert.NoError(t, err)
		assert.PanicsWithError(t, "failed to initialize elastic: missing user and password", func() { _ = r.Elastic() })
	})
}
