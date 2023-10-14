package database

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type DBTrace struct {
	DB Ext
}

var _ Ext = (*DBTrace)(nil)

func (rcv *DBTrace) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	ctx, span := interceptors.StartSpan(ctx, "DBTrace.Query")
	defer span.End()

	return rcv.DB.Query(ctx, query, args...)
}

func (rcv *DBTrace) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	ctx, span := interceptors.StartSpan(ctx, "DBTrace.QueryRow")
	defer span.End()

	return rcv.DB.QueryRow(ctx, query, args...)
}

func (rcv *DBTrace) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	ctx, span := interceptors.StartSpan(ctx, "DBTrace.Exec")
	defer span.End()

	return rcv.DB.Exec(ctx, sql, args...)
}

func (rcv *DBTrace) Begin(ctx context.Context) (pgx.Tx, error) {
	ctx, span := interceptors.StartSpan(ctx, "DBTrace.Begin")
	defer span.End()

	tx, err := rcv.DB.Begin(ctx)
	return &TxTrace{
		Tx: tx,
	}, err
}

func (rcv *DBTrace) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	ctx, span := interceptors.StartSpan(ctx, "DBTrace.SendBatch")
	defer span.End()

	return rcv.DB.SendBatch(ctx, b)
}
