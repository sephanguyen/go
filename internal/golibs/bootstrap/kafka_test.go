package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	mock_bootstrap "github.com/manabie-com/backend/mock/golibs/bootstrap"
	mock_kafka "github.com/manabie-com/backend/mock/golibs/kafka"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitKafka(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := zap.NewNop()

	type testConfig struct {
		KafkaCluster configs.KafkaClusterConfig
	}
	c := &testConfig{
		KafkaCluster: configs.KafkaClusterConfig{
			Address: "address",
			IsLocal: true,
		},
	}
	t.Run("with kafka config", func(t *testing.T) {
		kafka := new(mock_kafka.KafkaManagement)
		kafka.On("ConnectToKafka").Once()
		kafkaer := mock_bootstrap.NewKafkaer(t)
		kafkaer.On("NewKafkaManagement", l, &c.KafkaCluster).Once().Return(kafka, nil)
		rsc := NewResources().WithLogger(l)
		rsc.kafkaer = kafkaer
		err := initKafka(ctx, c, rsc)
		require.NoError(t, err)
		require.Equal(t, kafka, rsc.Kafka())
	})

	t.Run("with invalid kafka config", func(t *testing.T) {
		kafkaer := mock_bootstrap.NewKafkaer(t)
		kafkaer.On("NewKafkaManagement", l, &c.KafkaCluster).Once().Return(nil, errors.New("some errors"))
		rsc := NewResources().WithLogger(l)
		rsc.kafkaer = kafkaer
		err := initKafka(ctx, c, rsc)
		require.NoError(t, err)
		require.PanicsWithError(t, "some errors", func() { _ = rsc.Kafka() })
	})

	t.Run("without kafka config", func(t *testing.T) {
		type testConfig struct{}
		rsc := NewResources().WithLogger(l)
		err := initKafka(ctx, testConfig{}, rsc)
		require.NoError(t, err)
		require.Nil(t, rsc.kafkaMgmt)
		require.PanicsWithValue(t, "unable to init Kafka client: Kafka config is not provided", func() { _ = rsc.Kafka() })
	})
}
