package bootstrap

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestSetupMonitor(t *testing.T) {
	t.Parallel()
	rsc := NewResources().WithLogger(zap.NewNop())
	t.Run("success", func(t *testing.T) {
		config := &Config{
			Common: configs.CommonConfig{
				RemoteTrace: configs.RemoteTraceConfig{
					OtelCollectorReceiver: "test",
				},
				StatsEnabled: true,
			},
		}
		b := bootstrapper[Config]{}
		mockSvc := new(MockAllService[Config])
		mockSvc.On("WithPrometheusCollectors", mock.Anything).Return(nil)
		mockSvc.On("InitMetricsValue").Return(nil)
		mockSvc.On("WithOpencensusViews").Return(nil)

		err := b.setupMonitor(mockSvc, config, rsc)
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, mockSvc)
	})
}
