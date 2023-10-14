package testdata

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/manabie-com/backend/internal/golibs/database"
	"go.uber.org/zap"
	"math"
	"math/rand"
	"net/url"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// Row is a convenience wrapper over Rows that is returned by QueryRow.
//
// Row is an interface instead of a struct to allow tests to mock QueryRow. However,
// adding a method to an interface is technically a breaking change. Because of this
// the Row interface is partially excluded from semantic version requirements.
// Methods will not be removed or changed, but new methods may be added.
type Row interface {
	pgx.Row
}

// Rows is the result set returned from *Conn.Query. Rows must be closed before
// the *Conn can be used again. Rows are closed by explicitly calling Close(),
// calling Next() until it returns false, or when a fatal error occurs.
//
// Once a Rows is closed the only methods that may be called are Close(), Err(), and CommandTag().
//
// Rows is an interface instead of a struct to allow tests to mock Query. However,
// adding a method to an interface is technically a breaking change. Because of this
// the Rows interface is partially excluded from semantic version requirements.
// Methods will not be removed or changed, but new methods may be added.
type Rows interface {
	pgx.Rows
}

// queryer is an interface for Query
type queryer interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
}

// execer is an interface for Exec
type execer interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type QueryExecer interface {
	queryer
	execer
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

// TxStarter is an interface to deal with transaction
type TxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// TxController is an interface to deal with transaction
type TxController interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Ext is a union interface which can bind, query, and exec
type Ext interface {
	QueryExecer
	TxStarter
}

type TxHandler = func(ctx context.Context, tx pgx.Tx) error

func ExecInTx(ctx context.Context, db Ext, txHandler TxHandler) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Begin")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = errors.Wrap(tx.Commit(ctx), "tx.Commit")
	}()
	err = txHandler(ctx, tx)
	return err
}

// ExecuteTx runs fn inside a transaction and retries it as needed.
// On non-retryable failures, the transaction is aborted and rolled back;
// On success, the transaction is committed.
func ExecInTxWithRetry(ctx context.Context, db Ext, fn TxHandler) error {
	maxRetries := 10
	n := 0
	for {
		if n >= maxRetries {
			return fmt.Errorf("max retries ExecInTxWithRetry")
		}
		err := ExecInTx(ctx, db, fn)
		if err == nil || !isErrRetryable(err) {
			return err
		}

		sleep := int(math.Pow(2, float64(n)))*100 + (rand.Intn(100-1) + 1)
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		n++
	}
}

func errorCause(err error) error {
	for err != nil {
		if c, ok := err.(interface{ Unwrap() error }); ok {
			err = c.Unwrap()
		} else {
			break
		}
	}

	return err
}

func isErrRetryable(err error) bool {
	switch e := errorCause(err).(type) {
	case *pgconn.PgError:
		return e.Code == pgerrcode.SerializationFailure
	default:
		return false
	}
}

// BatchResults is a wrapper around pgx.BatchResults.
type BatchResults interface {
	pgx.BatchResults
}

// Tx is a wrapper around pgx.Tx.
//
// Tx represents a database transaction.
//
// Tx is an interface instead of a struct to enable connection pools to be implemented without relying on internal pgx
// state, to support pseudo-nested transactions with savepoints, and to allow tests to mock transactions. However,
// adding a method to an interface is technically a breaking change. If new methods are added to Conn it may be
// desirable to add them to Tx as well. Because of this the Tx interface is partially excluded from semantic version
// requirements. Methods will not be removed or changed, but new methods may be added.
type Tx interface {
	pgx.Tx
}

func NewConnectionPool(ctx context.Context,
	logger *zap.Logger,
	connectionURI string,
	maxConnection int32,
	logLvl pgx.LogLevel,
) *pgxpool.Pool {
	connPoolConfig, err := pgxpool.ParseConfig(connectionURI)
	if err != nil {
		logger.Panic("cannot read PG_CONNECTION_URI", zap.Error(err))
	}

	connPoolConfig.ConnConfig.Logger = zapadapter.NewLogger(logger)
	connPoolConfig.ConnConfig.LogLevel = logLvl
	connPoolConfig.MaxConns = maxConnection
	//connPoolConfig.BeforeAcquire = setPostgres(logger)
	connPoolConfig.MaxConnIdleTime = 15 * time.Second
	logger.Info("NewConnectionPool max connection", zap.Int32("max connection", maxConnection))

	pool, err := pgxpool.ConnectConfig(ctx, connPoolConfig)
	if err != nil {
		if u, err2 := url.Parse(connectionURI); err2 != nil {
			logger.Panic("cannot create new connection to Postgres (failed to parse URI)", zap.Error(err))
		} else {
			logger.Panic(fmt.Sprintf("cannot create new connection to %q", u.Redacted()), zap.Error(err))
		}
	}

	return pool
}

type SimpleStruct struct {
	DB database.Ext
}

func (s *SimpleStruct) Query() (pgx.Rows, error) {
	return s.DB.Query(context.Background(), "")
}