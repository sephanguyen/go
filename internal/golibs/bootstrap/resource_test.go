package bootstrap

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	mock_bootstrap "github.com/manabie-com/backend/mock/golibs/bootstrap"
	mock_kafka "github.com/manabie-com/backend/mock/golibs/kafka"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestResources_ServiceName(t *testing.T) {
	t.Parallel()
	rsc := NewResources().WithServiceName("abcd")
	assert.Equal(t, rsc.svcName, "abcd")
}

func TestResources_WithLogger(t *testing.T) {
	t.Parallel()

	c := &configs.CommonConfig{Environment: "local", Log: configs.LogConfig{ApplicationLevel: "info"}}
	l := logger.NewZapLogger("warn", false)

	// Calling WithLogger should override WithLoggerC, and vice versa
	t.Run("WithLogger", func(t *testing.T) {
		r := NewResources().WithLoggerC(c).WithLogger(l)
		assert.Nil(t, r.loggerConfig)
		assert.Same(t, r.Logger(), l)
	})

	t.Run("WithLoggerC", func(t *testing.T) {
		r := NewResources().WithLogger(l).WithLoggerC(c)
		assert.Nil(t, r.logger)
		assert.Equal(t, zap.InfoLevel, r.Logger().Level()) // check level matches
	})

	t.Run("missing config", func(t *testing.T) {
		r := NewResources()
		assert.Panics(t, func() { r.Logger() })
	})
}

func TestResources_WithDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := map[string]configs.PostgresDatabaseConfig{"eureka": {User: "postgres", DBName: "eureka"}}
	dbs := map[string]*database.DBTrace{"bob": {DB: &database.DBTrace{}}}
	l := zap.NewNop()

	t.Run("WithDatabase", func(t *testing.T) {
		r := NewResources().WithLogger(l).WithDatabaseC(ctx, c).WithDatabase(dbs)
		require.Nil(t, r.databaseCtx)
		require.Nil(t, r.databaseConfigs)
		require.Same(t, dbs["bob"], r.DBWith("bob"))
	})

	t.Run("WithDatabaseC", func(t *testing.T) {
		dbpool := &pgxpool.Pool{}
		cancelCalled := false
		dbcancel := func() error { cancelCalled = true; return nil }
		mock_databaser := mock_bootstrap.NewDatabaser(t)
		mock_databaser.On("ConnectV2", ctx, l, c["eureka"]).Once().Return(dbpool, dbcancel, nil)
		r := NewResources().WithLogger(l).WithDatabase(dbs).WithDatabaseC(ctx, c)
		r.databaser = mock_databaser

		require.Empty(t, r.databases)
		require.Equal(t, c, r.databaseConfigs)
		require.Same(t, ctx, r.databaseCtx)

		require.Same(t, dbpool, r.DBWith("eureka").DB)

		require.False(t, cancelCalled)
		require.NoError(t, r.Cleanup())
		require.True(t, cancelCalled, "dbcancel must be call after Resources.Cleanup()")
	})

	t.Run("missing config", func(t *testing.T) {
		r := NewResources().WithLogger(l)
		require.Panics(t, func() { _ = r.DBWith("bob") })
	})
}

func TestResources_WithElastic(t *testing.T) {
	t.Parallel()
	c := &configs.ElasticSearchConfig{}
	e := &elastic.SearchFactoryImpl{}
	l := zap.NewNop()

	t.Run("WithElastic", func(t *testing.T) {
		r := NewResources().WithElasticC(c).WithElastic(e)
		assert.Nil(t, r.elasticConfig)
		assert.Same(t, e, r.Elastic())
	})

	t.Run("WithElasticC", func(t *testing.T) {
		e2 := &elastic.SearchFactoryImpl{}
		elasticer := mock_bootstrap.NewElasticer(t)
		elasticer.On("Init", l, []string(nil), "", "", "", "").Return(e2, nil).Once()

		r := NewResources().WithLogger(l).WithElastic(e).WithElasticC(c)
		r.elasticer = elasticer
		assert.Nil(t, r.elastic)
		assert.Same(t, e2, r.Elastic())
		assert.NotSame(t, e, r.Elastic())
	})

	t.Run("missing config", func(t *testing.T) {
		r := NewResources().WithLogger(l)
		assert.Panics(t, func() { _ = r.Elastic() })
	})
}

func TestResources_WithNATS(t *testing.T) {
	t.Parallel()
	l := zap.NewNop()
	c := &configs.NatsJetStreamConfig{}

	t.Run("WithNATS", func(t *testing.T) {
		n := new(mock_nats.JetStreamManagement)
		r := NewResources().WithLogger(l).WithNATSC(c).WithNATS(n)
		require.Nil(t, r.natsjsConfig)
		require.Same(t, n, r.natsjs)
		require.Same(t, n, r.NATS())
	})

	t.Run("WithNATSC", func(t *testing.T) {
		n := new(mock_nats.JetStreamManagement)
		n.On("ConnectToJS").Once()
		natsjser := mock_bootstrap.NewNATSJetstreamer(t)
		natsjser.On("NewJetStreamManagement", l, c).Once().Return(n, nil)
		r := NewResources().WithLogger(l).WithNATS(n).WithNATSC(c)
		r.natsjser = natsjser
		require.Equal(t, c, r.natsjsConfig)
		require.Nil(t, r.natsjs)
		require.Same(t, n, r.NATS())
	})

	t.Run("missing config", func(t *testing.T) {
		r := NewResources().WithLogger(l)
		require.Panics(t, func() { _ = r.NATS() })
	})
}

func TestResources_WithKafka(t *testing.T) {
	t.Parallel()
	l := zap.NewNop()
	c := &configs.KafkaClusterConfig{}

	t.Run("WithKafka", func(t *testing.T) {
		k := new(mock_kafka.KafkaManagement)
		r := NewResources().WithLogger(l).WithKafkaC(c).WithKafka(k)
		require.Nil(t, r.kafkaConfig)
		require.Same(t, k, r.kafkaMgmt)
		require.Same(t, k, r.Kafka())
	})

	t.Run("WithKafkaC", func(t *testing.T) {
		k := new(mock_kafka.KafkaManagement)
		k.On("ConnectToKafka").Once()
		kafkaer := mock_bootstrap.NewKafkaer(t)
		kafkaer.On("NewKafkaManagement", l, c).Once().Return(k, nil)
		r := NewResources().WithLogger(l).WithKafka(k).WithKafkaC(c)
		r.kafkaer = kafkaer
		require.Equal(t, c, r.kafkaConfig)
		require.Nil(t, r.kafkaMgmt)
		require.Same(t, k, r.Kafka())
	})

	t.Run("missing config", func(t *testing.T) {
		r := NewResources().WithLogger(l)
		require.Panics(t, func() { _ = r.Kafka() })
	})
}

func TestResources_WithUnleash(t *testing.T) {
	t.Parallel()
	c := &configs.UnleashClientConfig{}
	l := zap.NewNop()

	t.Run("WithUnleash", func(t *testing.T) {
		u := new(mock_unleash_client.UnleashClientInstance)
		r := NewResources().WithUnleashC(c).WithUnleash(u)
		assert.Nil(t, r.unleashConfig)
		assert.Same(t, u, r.Unleash())
	})

	t.Run("WithUnleashC", func(t *testing.T) {
		u := new(mock_unleash_client.UnleashClientInstance)
		u.On("ConnectToUnleashClient").Return(nil).Once()
		unleashInitF = func(_, _, _ string, _ *zap.Logger) (unleashclient.ClientInstance, error) { return u, nil }
		defer func() { unleashInitF = unleashclient.NewUnleashClientInstance }()
		r := NewResources().WithLogger(l).WithUnleash(u).WithUnleashC(c)
		assert.Nil(t, r.unleash)
		assert.Same(t, u, r.Unleash())
	})

	t.Run("missing config", func(t *testing.T) {
		r := NewResources().WithLogger(l)
		assert.Panics(t, func() { _ = r.Unleash() })
	})
}

func Test_getAddress(t *testing.T) {
	t.Parallel()
	mockAddresses := map[string]configs.ListenerConfig{
		"a": {GRPC: ":8050", HTTP: ":8080"},
		"b": {GRPC: ":7050"},
		"c": {HTTP: ":6080", MigratedEnvironments: []string{"local", "stag", "uat", "prod"}},
		"d": {GRPC: ":5050", MigratedEnvironments: []string{"local", "stag"}},
	}
	t.Run("missing config", func(t *testing.T) {
		c := &configs.CommonConfig{Name: "b", Environment: "stag", ActualEnvironment: "stag", Organization: "manabie"}
		assert.Panics(t, func() { _ = getAddress(mockAddresses, c, "b", "HTTP", false) })
		assert.Panics(t, func() { _ = getAddress(mockAddresses, c, "c", "GRPC", false) })
	})

	t.Run("not migrated", func(t *testing.T) {
		c := &configs.CommonConfig{Name: "b", Environment: "stag", ActualEnvironment: "stag", Organization: "manabie"}
		assert.Equal(t, "a:8080", getAddress(mockAddresses, c, "a", "HTTP", false))
		assert.Equal(t, "a:8050", getAddress(mockAddresses, c, "a", "GRPC", false))
	})

	t.Run("force full address", func(t *testing.T) {
		c := &configs.CommonConfig{Name: "b", Environment: "stag", ActualEnvironment: "stag", Organization: "manabie"}
		assert.Equal(t, "a.stag-manabie-services.svc.cluster.local:8080", getAddress(mockAddresses, c, "a", "HTTP", true))
		assert.Equal(t, "a.stag-manabie-services.svc.cluster.local:8050", getAddress(mockAddresses, c, "a", "GRPC", true))
	})

	t.Run("migrated local", func(t *testing.T) {
		a := &configs.CommonConfig{Name: "a", Environment: "local", ActualEnvironment: "local", Organization: "manabie"}
		c := &configs.CommonConfig{Name: "c", Environment: "local", ActualEnvironment: "local", Organization: "manabie"}
		assert.Equal(t, "a.backend.svc.cluster.local:8080", getAddress(mockAddresses, c, "a", "HTTP", false))
		assert.Equal(t, "a.backend.svc.cluster.local:8050", getAddress(mockAddresses, c, "a", "GRPC", false))
		assert.Equal(t, "c.local-manabie-backend.svc.cluster.local:6080", getAddress(mockAddresses, a, "c", "HTTP", false))
	})

	t.Run("migrated staging", func(t *testing.T) {
		c := &configs.CommonConfig{Name: "c", Environment: "stag", ActualEnvironment: "stag", Organization: "manabie"}
		d := &configs.CommonConfig{Name: "d", Environment: "stag", ActualEnvironment: "stag", Organization: "manabie"}
		assert.Equal(t, "d:5050", getAddress(mockAddresses, c, "d", "GRPC", false))
		assert.Equal(t, "c:6080", getAddress(mockAddresses, d, "c", "HTTP", false))
	})

	t.Run("migrated uat", func(t *testing.T) {
		c := &configs.CommonConfig{Name: "c", Environment: "uat", ActualEnvironment: "uat", Organization: "manabie"}
		d := &configs.CommonConfig{Name: "d", Environment: "uat", ActualEnvironment: "uat", Organization: "manabie"}
		assert.Equal(t, "d.uat-manabie-services.svc.cluster.local:5050", getAddress(mockAddresses, c, "d", "GRPC", false))
		assert.Equal(t, "c.uat-manabie-backend.svc.cluster.local:6080", getAddress(mockAddresses, d, "c", "HTTP", false))
	})

	t.Run("service calls itself", func(t *testing.T) {
		d := &configs.CommonConfig{Name: "d", Environment: "prod", ActualEnvironment: "prod", Organization: "manabie"}
		assert.Equal(t, "d:5050", getAddress(mockAddresses, d, "d", "GRPC", false))
	})

	t.Run("both service not migrated yet", func(t *testing.T) {
		a := &configs.CommonConfig{Name: "a", Environment: "prod", ActualEnvironment: "prod", Organization: "manabie"}
		assert.Equal(t, "d:5050", getAddress(mockAddresses, a, "d", "GRPC", false))
	})

	t.Run("preproduction vs production", func(t *testing.T) {
		c := &configs.CommonConfig{Name: "c", Environment: "prod", ActualEnvironment: "prod", Organization: "manabie"}
		d := &configs.CommonConfig{Name: "d", Environment: "prod", ActualEnvironment: "prod", Organization: "manabie"}
		assert.Equal(t, "d.prod-manabie-services.svc.cluster.local:5050", getAddress(mockAddresses, c, "d", "GRPC", false))
		assert.Equal(t, "c.prod-manabie-backend.svc.cluster.local:6080", getAddress(mockAddresses, d, "c", "HTTP", false))

		c = &configs.CommonConfig{Name: "c", Environment: "prod", ActualEnvironment: "dorp", Organization: "manabie"}
		d = &configs.CommonConfig{Name: "d", Environment: "prod", ActualEnvironment: "dorp", Organization: "manabie"}
		assert.Equal(t, "d.dorp-manabie-services.svc.cluster.local:5050", getAddress(mockAddresses, c, "d", "GRPC", false))
		assert.Equal(t, "c.dorp-manabie-backend.svc.cluster.local:6080", getAddress(mockAddresses, d, "c", "HTTP", false))
	})
}
