package bootstrap

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConf struct {
	Common     configs.CommonConfig
	PostgresV2 configs.PostgresConfigV2
	NatsJS     configs.NatsJetStreamConfig
}

func TestExtract(t *testing.T) {
	t.Parallel()

	data := testConf{
		Common: configs.CommonConfig{
			Name: "svc",
		},
		PostgresV2: configs.PostgresConfigV2{
			Databases: map[string]configs.PostgresDatabaseConfig{
				"bob": {User: "postgres", DBName: "bob"},
			},
		},
		NatsJS: configs.NatsJetStreamConfig{
			Address: "nats://...",
		},
	}
	require.NotPanics(t, func() {
		v, err := extract[configs.CommonConfig](data, commonFieldName)
		require.NoError(t, err)
		assert.Equal(t, data.Common, *v)
	})
	require.NotPanics(t, func() {
		v, err := extract[configs.PostgresConfigV2](data, postgresV2FieldName)
		require.NoError(t, err)
		assert.Equal(t, data.PostgresV2, *v)
	})
	require.NotPanics(t, func() {
		v, err := extract[configs.NatsJetStreamConfig](data, natsjsFieldName)
		require.NoError(t, err)
		assert.Equal(t, data.NatsJS, *v)
	})
	require.NotPanics(t, func() {
		v, err := extract[configs.PostgresConfigV2](data, natsjsFieldName)
		require.EqualError(t, err, `expected field "NatsJS" to be of "configs.PostgresConfigV2" type, got "github.com/manabie-com/backend/internal/golibs/configs.NatsJetStreamConfig"`)
		assert.Nil(t, v)
	})
}
