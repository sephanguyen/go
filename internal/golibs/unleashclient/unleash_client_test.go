package unleashclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/unleashclient/configurations"
	"go.uber.org/zap"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewUnleashClientInstance(t *testing.T) {
	t.Parallel()
	zapLogger := zap.NewNop()
	t.Run("Valid Instance", func(t *testing.T) {
		resp, err := NewUnleashClientInstance("test_url", "unit_test_app_name", "mock_token", zapLogger)
		assert.Equal(t, err, nil)
		assert.Equal(t, resp, &unleashClientInstanceImpl{
			url:      "test_url",
			appName:  "unit_test_app_name",
			apiToken: "mock_token",
			logger:   zapLogger,
		})
	})
	t.Run("Nil Url", func(t *testing.T) {
		resp, err := NewUnleashClientInstance("", "unit_test_app_name", "mock_token", zapLogger)
		assert.Equal(t, err, errors.New("missing Unleash URL"))
		assert.Equal(t, resp, nil)
	})
	t.Run("Nil App Name", func(t *testing.T) {
		resp, err := NewUnleashClientInstance("", "unit_test_app_name", "mock_token", zapLogger)
		assert.Equal(t, err, errors.New("missing Unleash URL"))
		assert.Equal(t, resp, nil)
	})
	t.Run("Nil Token", func(t *testing.T) {
		resp, err := NewUnleashClientInstance("test_url", "unit_test_app_name", "", zapLogger)
		assert.Equal(t, err, errors.New("missing API Token"))
		assert.Equal(t, resp, nil)
	})
}
func TestUnleashConn(t *testing.T) {
	t.Parallel()
	zapLogger := zap.NewNop()
	t.Run("Valid connection", func(t *testing.T) {
		err := unleash.Initialize(
			unleash.WithListener(&DebugListener{logger: zapLogger}),
			unleash.WithAppName("test_app_name"),
			unleash.WithUrl("http://unleash:4242/unleash/api"),
			unleash.WithCustomHeaders(http.Header{"Authorization": {"test"}}),
			unleash.WithStrategies(&configurations.EnvStrategy{}),
		)
		assert.Equal(t, err, nil)
	})
	t.Run("Invalid connection", func(t *testing.T) {
		err := unleash.Initialize(
			unleash.WithListener(&DebugListener{logger: zapLogger}),
			unleash.WithAppName("test_app_name"),
			unleash.WithUrl(""),
			unleash.WithCustomHeaders(http.Header{"Authorization": {""}}),
			unleash.WithStrategies(&configurations.EnvStrategy{}),
		)
		assert.NotEqual(t, err, nil)
	})
}
func TestUnleashToggle(t *testing.T) {
	zapLogger := zap.NewNop()
	resp, err := NewUnleashClientInstance("test_url", "unit_test_app_name", "mock_token", zapLogger)
	assert.Equal(t, err, nil)
	err = resp.ConnectToUnleashClient()
	assert.Equal(t, err, nil)
	isToggledMock, err := resp.IsFeatureEnabled("test_invalid", "local")
	assert.Equal(t, err, nil)
	assert.Equal(t, resp, &unleashClientInstanceImpl{
		url:      "test_url",
		appName:  "unit_test_app_name",
		apiToken: "mock_token",
		logger:   zapLogger,
	})
	assert.Equal(t, isToggledMock, false)
}
func TestUnleashToggleOrganization(t *testing.T) {
	zapLogger := zap.NewNop()
	resp, err := NewUnleashClientInstance("test_url", "unit_test_app_name", "mock_token", zapLogger)
	assert.Equal(t, err, nil)
	err = resp.ConnectToUnleashClient()
	assert.Equal(t, err, nil)
	isToggledMock, err := resp.IsFeatureEnabledOnOrganization("test_invalid", "local", "-2147483635")
	assert.Equal(t, err, nil)
	assert.Equal(t, resp, &unleashClientInstanceImpl{
		url:      "test_url",
		appName:  "unit_test_app_name",
		apiToken: "mock_token",
		logger:   zapLogger,
	})
	assert.Equal(t, isToggledMock, false)
}
