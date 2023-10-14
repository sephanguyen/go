package database

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// defaultMaxConnIdleTime is the default value for MaxConnIdleTime.
var defaultMaxConnIdleTime = time.Minute * 5

// NewConnectionPoolDemo returns a pgxpool.Pool that can be used to query database.
// Caller must invoke the cleanup function (2nd returned value) to clean up all resources.
// The cleanup function also calls pgxpool.Pool.Close().
//
// If cfg.CSQLInstance is not empty (usually for non-local environments), that means we are
// connecting to a Cloud SQL instance. In that case, the ConnConfig.DialFunc will be replaced
// by a special dialer from https://github.com/GoogleCloudPlatform/cloud-sql-go-connector.
// Host and port values in that cases are unused.
//
// On the other hand, when cfg.CSQLInstance is empty (usually for local environment),
// nothing special happens to the dialer.
func NewPool(ctx context.Context, l *zap.Logger, cfg configs.PostgresDatabaseConfig) (*pgxpool.Pool, func() error, error) {
	l.Info("database configuration",
		zap.String("instance", cfg.CloudSQLInstance),
		zap.String("user", cfg.User),
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("dbname", cfg.DBName),
		zap.Int32("max_conns", cfg.MaxConns),
	)

	connstr, err := cfg.PGXConnectionString()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get db connection string: %s", err)
	}
	poolconfig, err := pgxpool.ParseConfig(connstr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse database.connection: %s", err)
	}

	// If ShardID is given, set it to context to set up BeforeAcquire.
	if cfg.ShardID != nil {
		ctx = context.WithValue(ctx, ConfigShardIDKey, *cfg.ShardID)
	}

	// Set the rest of the configs
	poolconfig.MaxConns = cfg.MaxConns
	poolconfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolconfig.ConnConfig.Logger = zapadapter.NewLogger(l)
	poolconfig.BeforeAcquire = setPostgres(ctx, l)

	// Set log level of pgx to DEBUG, so that the log level
	// is mainly determined only by the zap.Logger's log level.
	poolconfig.ConnConfig.LogLevel = pgx.LogLevelDebug

	// idle time must always be > 0
	if cfg.MaxConnIdleTime.Seconds() <= 0 {
		cfg.MaxConnIdleTime = defaultMaxConnIdleTime
	}

	dialerCleanup := func() error { return nil }
	if cfg.IsCloudSQL() {
		// Use a dialer from Cloud SQL Go Connector to connect to database
		connopts, err := cfg.DefaultCloudSQLConnOpts(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("cfg.DefaultCloudSQLConnOpts: %s", err)
		}
		d, err := cloudsqlconn.NewDialer(ctx, connopts...)
		if err != nil {
			return nil, nil, fmt.Errorf("cloudsqlconn.NewDialer: %s", err)
		}
		poolconfig.ConnConfig.DialFunc = func(dialCtx context.Context, _ string, instance string) (net.Conn, error) {
			return d.Dial(dialCtx, cfg.CloudSQLInstance)
		}
		dialerCleanup = func() error { return d.Close() }
	}

	// pgxpool.ConnectConfig, but with retries
	var pool *pgxpool.Pool
	err = try.Do(func(attempt int) (retry bool, err error) {
		pool, err = pgxpool.ConnectConfig(ctx, poolconfig)
		if err != nil {
			l.Warn("failed to connect to database, will retry",
				zap.String("instance", cfg.CloudSQLInstance),
				zap.String("user", cfg.User),
				zap.String("host", cfg.Host),
				zap.String("port", cfg.Port),
				zap.String("dbname", cfg.DBName),
				zap.Int32("max_conns", cfg.MaxConns),
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
			if attempt < cfg.RetryAttempts {
				time.Sleep(cfg.RetryWaitInterval)
				return true, err
			}
			return false, err
		}
		return false, nil
	})

	cleanup := func() error {
		l.Debug("closing database connection", zap.String("dbname", cfg.DBName))
		pool.Close()
		return dialerCleanup()
	}

	return pool, cleanup, err
}

type Config int

const (
	ConfigShardIDKey Config = iota
)

func setPostgres(ctx context.Context, logger *zap.Logger) func(context.Context, *pgx.Conn) bool {
	shardID := ctx.Value(ConfigShardIDKey)

	return func(ctx context.Context, conn *pgx.Conn) bool {
		claims := interceptors.JWTClaimsFromContext(ctx)
		ctx, span := interceptors.StartSpan(ctx, "setPostgres")
		defer span.End()
		var resourcePath, userGroup, userID string
		if claims != nil && claims.Manabie != nil {
			resourcePath = claims.Manabie.ResourcePath
			userGroup = claims.Manabie.UserGroup
			userID = claims.Manabie.UserID
		}
		var resourcePathCheck, userGroupCheck, userIDCheck string
		err := conn.QueryRow(ctx,
			"SELECT set_config('permission.resource_path', $1::text, false) as resource_path, set_config('permission.user_group', $2::text, false) as user_group, set_config('app.user_id', $3::text, false)",
			&resourcePath, &userGroup, &userID).Scan(&resourcePathCheck, &userGroupCheck, &userIDCheck)
		if err != nil {
			logger.Error("setPostgres",
				zap.String("resourcePath", resourcePath),
				zap.String("userGroup", userGroup),
				zap.String("userID", userID),
				zap.String("resourcePathCheck", resourcePathCheck),
				zap.String("userGroupCheck", userGroupCheck),
				zap.String("userIDCheck", userIDCheck),
				zap.Error(err))
			return false
		}

		if resourcePath != resourcePathCheck || userGroup != userGroupCheck || userID != userIDCheck {
			logger.Error("setPostgres",
				zap.String("resourcePath", resourcePath),
				zap.String("userGroup", userGroup),
				zap.String("userID", userID),
				zap.String("resourcePathCheck", resourcePathCheck),
				zap.String("userGroupCheck", userGroupCheck),
				zap.String("userIDCheck", userIDCheck))
			return false
		}

		switch shardID := shardID.(type) {
		case int:
			if shardID <= 0 {
				break
			}

			stmt := "SELECT set_config('database.shard_id', $1::text, false);"
			arg := strconv.Itoa(shardID)

			var queriedShardID string
			err := conn.QueryRow(ctx, stmt, &arg).Scan(&queriedShardID)
			if err != nil {
				logger.Error("setPostgres",
					zap.Int("shardID", shardID),
					zap.String("queriedShardId", queriedShardID),
					zap.Error(err))
				return false
			}
		default:
			break
		}

		return true
	}
}
