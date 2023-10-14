package bootstrap

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"

	"go.uber.org/zap"
)

type initUnleashFunc func(url, appName, apiToken string, zapLogger *zap.Logger) (unleashclient.ClientInstance, error)

func initUnleash(ctx context.Context, config interface{}, rsc *Resources) error {
	return initUnleashf(ctx, config, rsc, unleashclient.NewUnleashClientInstance)
}

func initUnleashf(_ context.Context, config interface{}, rsc *Resources, f initUnleashFunc) error {
	c, err := extract[configs.UnleashClientConfig](config, unleashFieldName)
	if err != nil {
		return ignoreErrFieldNotFound(err)
	}

	unleashClientInstance, err := f(
		c.URL,
		c.AppName,
		c.APIToken,
		rsc.Logger(),
	)
	if err != nil {
		return err
	}

	err = unleashClientInstance.ConnectToUnleashClient()
	if err != nil {
		return err
	}
	rsc.unleash = unleashClientInstance

	return nil
}

var unleashInitF initUnleashFunc = unleashclient.NewUnleashClientInstance
