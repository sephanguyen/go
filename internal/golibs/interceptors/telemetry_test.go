package interceptors

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestIgnoreTracingInterceptors(t *testing.T) {
	t.Parallel()
	ignoreEndpoint := "ignore tracing"
	interceptors := NewIgnoredTraceInterceptor(map[string]struct{}{ignoreEndpoint: {}})

	assertNotTracedHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.True(t, isCtxIgnoringTracing(ctx))
		return nil, nil
	}
	assertHasTracingHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.False(t, isCtxIgnoringTracing(ctx))
		return nil, nil
	}
	_, err := interceptors.UnaryServerInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: ignoreEndpoint}, assertNotTracedHandler)
	assert.NoError(t, err)
	_, err = interceptors.UnaryServerInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "normal endpoint"}, assertHasTracingHandler)
	assert.NoError(t, err)
}

// This test is pretty out-dated. TODO: remove it with InitTelemetry function.
func TestTelemetryInterceptorsConfigs(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		initEnv configs.CommonConfig
	}{
		"enable all": {
			initEnv: configs.CommonConfig{
				StatsEnabled: true,
				RemoteTrace: configs.RemoteTraceConfig{
					Enabled: true,
				},
			},
		},
		"stats only": {
			initEnv: configs.CommonConfig{
				StatsEnabled: true,
				RemoteTrace: configs.RemoteTraceConfig{
					Enabled: false,
				},
			},
		},
		"missing agent endpoint": {
			initEnv: configs.CommonConfig{
				StatsEnabled: false,
				RemoteTrace: configs.RemoteTraceConfig{
					Enabled: false,
				},
			},
		},
		"missing collector endpoint": {
			initEnv: configs.CommonConfig{
				StatsEnabled: false,
				RemoteTrace: configs.RemoteTraceConfig{
					Enabled: true,
				},
			},
		},
	}

	for name, c := range testCases {
		name := name
		c := c
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			e, _, _ := InitTelemetry(&c.initEnv, "testSvc", 0.1)
			if c.initEnv.StatsEnabled && e == nil {
				t.Errorf("when configs.StatsEnabled is enabled, stats exporter must not be nil")
			}
		})
	}
}
