package unleashclient

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/unleashclient/configurations"

	"github.com/Unleash/unleash-client-go/v3"
	unleash_ctx "github.com/Unleash/unleash-client-go/v3/context"
	"go.uber.org/zap"
)

type DebugListener struct {
	logger *zap.Logger
}

// OnError error log through zapLogger
func (l DebugListener) OnError(err error) {
	l.logger.Error("unleash error", zap.Error(err))
}

// OnWarning warn log through zapLogger
func (l DebugListener) OnWarning(warning error) {
	l.logger.Warn("unleash warn", zap.Error(warning))
}

// OnReady debug log through zapLogger when the repository is ready.
func (l DebugListener) OnReady() {
	l.logger.Debug("READY")
}

// OnCount debug log through zapLogger when the feature is queried.
func (l DebugListener) OnCount(name string, enabled bool) {
	l.logger.Debug("unleash feature request", zap.String("feature", name), zap.Bool("enabled", enabled))
}

// OnSent debug log through zapLogger when the server has uploaded metrics.
func (l DebugListener) OnSent(payload unleash.MetricsData) {
	l.logger.Debug(fmt.Sprintf("Sent: %+v\n", payload))
}

// OnRegistered debug log through zapLogger when the client has registered.
func (l DebugListener) OnRegistered(payload unleash.ClientData) {
	l.logger.Debug(fmt.Sprintf("Registered: %+v\n", payload))
}

type ClientInstance interface {
	ConnectToUnleashClient() error
	IsFeatureEnabled(featureName, backendEnv string) (bool, error)
	IsFeatureEnabledOnOrganization(featureName, backendEnv, resourcePath string) (bool, error)
	WaitForUnleashReady()
}

type unleashClientInstanceImpl struct {
	sync.RWMutex
	url      string
	appName  string
	apiToken string
	logger   *zap.Logger
}

func NewUnleashClientInstance(url, appName, apiToken string, zapLogger *zap.Logger) (ClientInstance, error) {
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

	unleashClientInstance := &unleashClientInstanceImpl{
		url:      url,
		appName:  appName,
		apiToken: apiToken,
		logger:   zapLogger,
	}

	return unleashClientInstance, nil
}
func (u *unleashClientInstanceImpl) ConnectToUnleashClient() error {
	err := unleash.Initialize(
		unleash.WithListener(&DebugListener{logger: u.logger}),
		unleash.WithAppName(u.appName),
		unleash.WithUrl(u.url),
		unleash.WithCustomHeaders(http.Header{"Authorization": {u.apiToken}}),
		unleash.WithStrategies(&configurations.EnvStrategy{}, &configurations.OrgStrategy{}),
	)
	return err
}

func (u *unleashClientInstanceImpl) IsFeatureEnabled(featureName, backendEnv string) (bool, error) {
	// env strategy
	envStrategy := getEnvStrategy(backendEnv)

	envCtx := unleash_ctx.Context{
		Properties: map[string]string{
			"env": envStrategy,
		},
	}
	isFeatureEnabled := unleash.IsEnabled(featureName, unleash.WithContext(envCtx))

	return isFeatureEnabled, nil
}

func (u *unleashClientInstanceImpl) IsFeatureEnabledOnOrganization(featureName, backendEnv, resourcePath string) (bool, error) {
	// env strategy
	envStrategy := getEnvStrategy(backendEnv)

	envCtx := unleash_ctx.Context{
		Properties: map[string]string{
			"env": envStrategy,
			"org": resourcePath,
		},
	}
	isFeatureEnabled := unleash.IsEnabled(featureName, unleash.WithContext(envCtx))

	return isFeatureEnabled, nil
}

func (u *unleashClientInstanceImpl) WaitForUnleashReady() {
	unleash.WaitForReady()
}

func getEnvStrategy(backendEnv string) string {
	// env strategy
	var envStrategy string
	if backendEnv == "local" || backendEnv == "staging" {
		envStrategy = "stag"
	} else {
		envStrategy = backendEnv
	}
	return envStrategy
}
