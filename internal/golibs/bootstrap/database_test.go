package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	mock_bootstrap "github.com/manabie-com/backend/mock/golibs/bootstrap"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := zap.NewNop()

	t.Run("invalid struct error", func(t *testing.T) {
		type missingPostgresFieldConfig struct{}

		rsc := NewResources().WithLogger(l)

		err := initDatabase(ctx, &missingPostgresFieldConfig{}, rsc)
		require.NoError(t, err)
		require.Empty(t, rsc.databases)
		require.Panics(t, func() { rsc.DB() })
		require.Panics(t, func() { rsc.DBWith("bob") })
	})

	type dbConfig2 struct {
		PostgresV2 configs.PostgresConfigV2
	}

	t.Run("connects successfully", func(t *testing.T) {
		c := dbConfig2{
			PostgresV2: configs.PostgresConfigV2{
				Databases: map[string]configs.PostgresDatabaseConfig{"bob": {CloudSQLInstance: "abcd"}},
			},
		}
		dbpool := &pgxpool.Pool{}
		dbcancel := func() error { return nil }
		mockDatabaser := mock_bootstrap.NewDatabaser(t)
		mockDatabaser.On("ConnectV2", ctx, l, c.PostgresV2.Databases["bob"]).Once().Return(dbpool, dbcancel, nil)
		rsc := NewResources().WithServiceName("bob").WithLogger(l)
		rsc.databaser = mockDatabaser

		err := initDatabase(ctx, c, rsc)
		require.NoError(t, err)
		require.Same(t, dbpool, rsc.DBWith("bob").DB)
		require.Same(t, dbpool, rsc.DB().DB) // can connect using default service name for db name
	})

	t.Run("errors when connecting to db", func(t *testing.T) {
		c := &dbConfig2{
			PostgresV2: configs.PostgresConfigV2{Databases: map[string]configs.PostgresDatabaseConfig{"bob": {CloudSQLInstance: "abcd"}}},
		}

		mockDatabaser := mock_bootstrap.NewDatabaser(t)
		mockDatabaser.On("ConnectV2", ctx, l, c.PostgresV2.Databases["bob"]).Once().Return(nil, nil, errors.New("connection timed out"))
		rsc := NewResources().WithServiceName("bob").WithLogger(l)
		rsc.databaser = mockDatabaser

		err := initDatabase(ctx, c, rsc)
		require.NoError(t, err)
		require.PanicsWithError(t, "failed to connect to database bob: connection timed out", func() { rsc.DB() })
		require.Panics(t, func() { rsc.DBWith("bob") })
	})
}
