package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestInitUnleashf(t *testing.T) {
	t.Parallel()
	unleash := new(mock_unleash_client.UnleashClientInstance)

	f := func(url, appName, apiToken string, zapLogger *zap.Logger) (unleashclient.ClientInstance, error) {
		if url == "" {
			return nil, errors.New("missing Unleash URL")
		}

		if appName == "" {
			return nil, errors.New("missing App Name")
		}

		if apiToken == "" {
			return nil, errors.New("missing API Token")
		}

		if zapLogger == nil {
			return nil, errors.New("missing logger")
		}

		return unleash, nil
	}

	t.Run("with unleash config", func(t *testing.T) {
		type testConfig struct {
			UnleashClientConfig configs.UnleashClientConfig
		}

		c := testConfig{
			UnleashClientConfig: configs.UnleashClientConfig{
				URL:      "url",
				AppName:  "appname",
				APIToken: "apitoken",
			},
		}
		rsc := NewResources().WithLogger(zap.NewNop())
		unleash.On("ConnectToUnleashClient").Once().Return(nil)
		err := initUnleashf(context.Background(), c, rsc, f)
		assert.NoError(t, err)
		assert.Equal(t, unleash, rsc.Unleash())
	})

	t.Run("without unleash config", func(t *testing.T) {
		type testConfig struct{}
		rsc := NewResources().WithLogger(zap.NewNop())
		err := initUnleashf(context.Background(), testConfig{}, rsc, f)
		assert.NoError(t, err)
		assert.Nil(t, rsc.unleash)
	})

	t.Run("with invalid unleash config", func(t *testing.T) {
		type testConfig struct {
			UnleashClientConfig configs.UnleashClientConfig
		}

		c := testConfig{
			UnleashClientConfig: configs.UnleashClientConfig{
				URL:     "url",
				AppName: "appname",
			},
		}
		rsc := NewResources().WithLogger(zap.NewNop())
		err := initUnleashf(context.Background(), c, rsc, f)
		assert.Error(t, err)
		assert.EqualError(t, err, "missing API Token")
	})
}
